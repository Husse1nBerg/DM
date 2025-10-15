package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/labstack/echo/v4"
)

type AIHandler struct {
	server *s.Server
}

func NewAIHandler(server *s.Server) *AIHandler {
	return &AIHandler{server: server}
}

// RewriteHandler handles rewriting customer-facing messages using Bedrock API
// @Summary Rewrite customer-facing message
// @Description Rewrite a draft message to be more professional using Bedrock API
// @Tags AI
// @Accept json
// @Produce json
// @Param rewrite body requests.BedrockRewriteRequest true "Rewrite request"
// @Success 200 {object} responses.BedrockRewriteResponse "Rewritten message"
// @Failure 400 {object} responses.Error "Bad request"
// @Failure 500 {object} responses.Error "Server error"
// @Router /ai/compose-message [post]
func (h *AIHandler) RewriteHandler(c echo.Context) error {
	logger := h.server.Logger

	// Parse and validate request
	req := new(requests.BedrockRewriteRequest)
	if err := c.Bind(req); err != nil {
		logger.Zap.Errorw("Failed to bind request", "error", err)
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}
	if err := c.Validate(req); err != nil {
		logger.Zap.Errorw("Failed to validate request", "error", err)
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Prepare Bedrock API request
	bedrockRequest := map[string]interface{}{
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{"text": req.Draft},
					{"text": "The marina name is " + req.MarinaName + " and the staff sender is " + req.UserName + ". Use a " + req.Tone + " tone."},
					{"text": req.ExtraInstructions},
					{"text": "You rewrite customer-facing messages for a marina. Be concise, clear, and polite. Focus on the main message without additional formalities. Insert the marina name and the staff sender name in the signature if not present. Preserve facts; do not invent details. Output text only."},
				},
			},
		},
	}

	// Prepare HTTP request to Bedrock API
	client := &http.Client{}
	reqBody, err := json.Marshal(bedrockRequest)
	if err != nil {
		logger.Zap.Errorw("Failed to marshal request body", "error", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to prepare request").JSON(c)
	}

	bedrockReq, err := http.NewRequest("POST", "https://bedrock-runtime.us-east-1.amazonaws.com/model/amazon.nova-lite-v1:0/invoke", bytes.NewBuffer(reqBody))
	if err != nil {
		logger.Zap.Errorw("Failed to create new HTTP request", "error", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to create request").JSON(c)
	}

	bedrockReq.Header.Set("Authorization", "Bearer "+os.Getenv("AWS_BEARER_TOKEN_BEDROCK"))
	bedrockReq.Header.Set("Content-Type", "application/json")
	bedrockReq.Header.Set("Accept", "application/json")

	// Send request to Bedrock API
	resp, err := client.Do(bedrockReq)
	if err != nil {
		logger.Zap.Errorw("Failed to send request to Bedrock API", "error", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to send request").JSON(c)
	}
	defer resp.Body.Close()

	// Parse response from Bedrock API
	var apiResponse struct {
		Output struct {
			Message struct {
				Content []struct {
					Text string `json:"text"`
				} `json:"content"`
				Role string `json:"role"`
			} `json:"message"`
		} `json:"output"`
		StopReason string `json:"stopReason"`
		Usage      struct {
			InputTokens  int `json:"inputTokens"`
			OutputTokens int `json:"outputTokens"`
			TotalTokens  int `json:"totalTokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		logger.Zap.Errorw("Failed to decode response from Bedrock API", "error", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to decode response").JSON(c)
	}

	// Map the parsed response to your response structure
	bedrockResponse := responses.BedrockRewriteResponse{
		Message: apiResponse.Output.Message.Content[0].Text,
		Tokens: struct {
			Input  int `json:"input"`
			Output int `json:"output"`
		}{
			Input:  apiResponse.Usage.InputTokens,
			Output: apiResponse.Usage.OutputTokens,
		},
	}

	// Return response to client
	return c.JSON(http.StatusOK, bedrockResponse)
}

// DetectFormFieldsHandler detects form fields in an image
//
// @Summary Detect form fields
// @Description Detect form fields in an image (supports jpg, jpeg, png) and return structured JSON
// @Tags AI
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "Image file (jpg, jpeg, png)"
// @Success 200 {object} responses.BedrockDetectFormFieldsResponse "Detected form fields"
// @Failure 400 {object} responses.Error "Bad request - invalid file type or missing file"
// @Failure 500 {object} responses.Error "Server error"
// @Router /ai/detect-form-fields [post]
func (h *AIHandler) DetectFormFieldsHandler(c echo.Context) error {
	logger := h.server.Logger

	// Get file from form
	file, header, err := c.Request().FormFile("file")
	if err != nil {
		logger.Zap.Error("Error getting file", err)
		return responses.NewErrorResponse(http.StatusBadRequest, "File is required").JSON(c)
	}
	defer file.Close()

	// Validate file type
	contentType := header.Header.Get("Content-Type")

	if !isValidImageType(contentType) {
		logger.Zap.Error("Invalid file type", "contentType", contentType, "filename", header.Filename)
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid file type. Supported types: jpg, jpeg, png").JSON(c)
	}

	// Read file content into a byte slice
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		logger.Zap.Error("Failed to read file content", "error", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to read file content").JSON(c)
	}

	// Llama3 Vision uses a different format: prompt + image array
	// Format the prompt in Llama3's instruction template with image token
	userPrompt := `You are looking at a form document image. Your task is to identify every form field and determine its EXACT position by carefully observing where it appears in the image.

CRITICAL: Look at the ACTUAL image and measure where each field is positioned. Do NOT use template/example coordinates.

Step 1: Determine the image dimensions
- Measure or estimate the image width and height in pixels
- Common sizes: Letter (850x1100), A4 (595x842)

Step 2: Locate each field by observing the image
For each form field you see, determine its position:
- Measure from the TOP-LEFT corner (0,0) to the field
- If a field appears in the MIDDLE of the page, y should be around 500-600 (for 1100px height)
- If a field appears NEAR THE BOTTOM, y should be 700-900+ (for 1100px height)
- If a field appears at the TOP, y should be 50-200

IMPORTANT POSITION GUIDELINES:
- Top third of page: y = 0 to 366px (normalized: 0 to 0.33)
- Middle third: y = 367 to 733px (normalized: 0.33 to 0.67)
- Bottom third: y = 734 to 1100px (normalized: 0.67 to 1.0)

For signatures and dates which typically appear NEAR THE BOTTOM of forms:
- These should have y values of 700+ pixels (normalized 0.64+)

Step 3: Calculate normalized coordinates
- bboxNorm.x = bbox.x / pageSize.widthPx
- bboxNorm.y = bbox.y / pageSize.heightPx
- bboxNorm.w = bbox.w / pageSize.widthPx
- bboxNorm.h = bbox.h / pageSize.heightPx

Output format - RESPOND WITH ONLY THE YAML BELOW, NO OTHER TEXT:

pages:
  - pageNumber: 1
    pageSize: {widthPx: 850, heightPx: 1100}
    fields:
      - id: FIELD_ID
        type: text|signature|date|etc
        label: "Actual label from image"
        required: false
        confidence: 0.9
        bbox: {x: <actual_x>, y: <actual_y>, w: <actual_w>, h: <actual_h>}
        bboxNorm: {x: <calculated>, y: <calculated>, w: <calculated>, h: <calculated>}

CRITICAL:
- Do NOT add "Answer:" or any prefix
- Do NOT use markdown code blocks
- Do NOT add explanations
- Start your response directly with "pages:"
- Return ONLY the raw YAML data`

	// Include <|image|> token to indicate where the image should be placed
	formattedPrompt := fmt.Sprintf(`<|begin_of_text|><|start_header_id|>user<|end_header_id|>

<|image|>
%s
<|eot_id|>
<|start_header_id|>assistant<|end_header_id|>
`, userPrompt)

	bedrockRequest := map[string]interface{}{
		"prompt": formattedPrompt,
		"images": []string{
			base64.StdEncoding.EncodeToString(fileBytes),
		},
		"max_gen_len": 2048,
		"temperature": 0.2,
		"top_p":       0.9,
	}

	// Prepare HTTP request to Bedrock API
	client := &http.Client{}
	reqBody, err := json.Marshal(bedrockRequest)
	if err != nil {
		logger.Zap.Errorw("Failed to marshal request body", "error", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to prepare request").JSON(c)
	}

	bedrockReq, err := http.NewRequest("POST", "https://bedrock-runtime.us-east-1.amazonaws.com/model/us.meta.llama3-2-90b-instruct-v1:0/invoke", bytes.NewBuffer(reqBody))
	if err != nil {
		logger.Zap.Errorw("Failed to create new HTTP request", "error", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to create request").JSON(c)
	}

	bedrockReq.Header.Set("Authorization", "Bearer "+os.Getenv("AWS_BEARER_TOKEN_BEDROCK"))
	bedrockReq.Header.Set("Content-Type", "application/json")
	bedrockReq.Header.Set("Accept", "application/json")

	// Send request to Bedrock API
	resp, err := client.Do(bedrockReq)
	if err != nil {
		logger.Zap.Errorw("Failed to send request to Bedrock API", "error", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to send request").JSON(c)
	}
	defer resp.Body.Close()

	// Read response body for debugging
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Zap.Errorw("Failed to read response body", "error", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to read response").JSON(c)
	}

	// Log the raw response for debugging
	logger.Zap.Infow("Bedrock API Response", "status", resp.StatusCode, "body", string(bodyBytes))

	// Parse response from Bedrock API - Llama3 format
	var apiResponse struct {
		Generation           string `json:"generation"`
		PromptTokenCount     int    `json:"prompt_token_count"`
		GenerationTokenCount int    `json:"generation_token_count"`
		StopReason           string `json:"stop_reason"`
	}

	if err := json.Unmarshal(bodyBytes, &apiResponse); err != nil {
		logger.Zap.Errorw("Failed to decode response from Bedrock API", "error", err, "body", string(bodyBytes))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to decode response: "+err.Error()).JSON(c)
	}

	// Check if we got a generation
	if apiResponse.Generation == "" {
		logger.Zap.Errorw("No generation in Bedrock response", "response", string(bodyBytes))
		return responses.NewErrorResponse(http.StatusInternalServerError, "No content in response").JSON(c)
	}

	// Return the response with the generated YAML
	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": apiResponse.Generation,
		"tokens": map[string]int{
			"input":  apiResponse.PromptTokenCount,
			"output": apiResponse.GenerationTokenCount,
		},
		"stopReason": apiResponse.StopReason,
	})
}

// isValidImageType checks if the content type is a supported image format
func isValidImageType(contentType string) bool {
	validTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
	}
	return validTypes[contentType]
}

// getImageFormat returns the format string for Bedrock API based on content type
func getImageFormat(contentType string) string {
	switch contentType {
	case "image/jpeg", "image/jpg":
		return "jpeg"
	case "image/png":
		return "png"
	default:
		return "jpeg" // default to jpeg if unknown
	}
}
