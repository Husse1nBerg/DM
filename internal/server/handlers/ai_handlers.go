package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
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

	// Get image format from content type
	format := getImageFormat(contentType)

	// Prepare Bedrock API request
	bedrockRequest := map[string]interface{}{
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"text": "Detect form fields in the PDF image and return a valid JSON object. Follow these strict rules:\n\n1. Schema:\n{\n  \"pages\": [{\n    \"pageNumber\": integer,\n    \"pageSize\": {\n      \"widthPx\": integer,\n      \"heightPx\": integer\n    },\n    \"fields\": [{\n      \"id\": string,\n      \"type\": string (one of: text|textarea|checkbox|radio|signature|date|initials|email|phone|number),\n      \"label\": string or null,\n      \"required\": boolean,\n      \"confidence\": number between 0-1,\n      \"bbox\": {\n        \"x\": integer,\n        \"y\": integer,\n        \"w\": integer,\n        \"h\": integer\n      },\n      \"bboxNorm\": {\n        \"x\": number between 0-1,\n        \"y\": number between 0-1,\n        \"w\": number between 0-1,\n        \"h\": number between 0-1\n      }\n    }]\n  }]\n}\n\n2. Rules:\n- Use double quotes for all JSON keys and string values\n- All numbers must be valid JSON numbers (no spaces, valid decimals)\n- Coordinates origin is top-left\n- Normalized coordinates must be calculated by dividing by width/height\n- Set required=true only for fields with asterisk or explicit required indicator\n- Use radio type only for mutually exclusive option groups\n- Exclude decorative elements\n- Return only the JSON object, no markdown or commentary",
					},
					{
						"image": map[string]interface{}{
							"format": format,
							"source": map[string]interface{}{
								"bytes": base64.StdEncoding.EncodeToString(fileBytes),
							},
						},
					},
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
	var apiResponse responses.BedrockDetectFormFieldsResponse

	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		logger.Zap.Errorw("Failed to decode response from Bedrock API", "error", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to decode response").JSON(c)
	}

	// Safely extract the message text
	if len(apiResponse.Output.Message.Content) == 0 {
		logger.Zap.Error("No content in Bedrock response")
		return responses.NewErrorResponse(http.StatusInternalServerError, "No content in response").JSON(c)
	}

	// Return the parsed JSON
	return c.JSON(http.StatusOK, apiResponse)
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
