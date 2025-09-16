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

	bedrockRequest := map[string]interface{}{
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"text": "Detect form fields in image. Return JSON with pages array. Each page has: pageNumber(int), pageSize(widthPx,heightPx), fields array. Each field has: id(str), type(text|textarea|checkbox|radio|signature|date|initials|email|phone|number), label(str|null), required(bool,true if *), confidence(0-1), bbox(x,y,w,h ints), bboxNorm(x,y,w,h 0-1). Rules: double quotes, valid numbers, origin top-left, bboxNorm=bbox/pageSize",
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

	bedrockReq, err := http.NewRequest("POST", "https://bedrock-runtime.us-east-1.amazonaws.com/model/amazon.nova-pro-v1:0/invoke", bytes.NewBuffer(reqBody))
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
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to decode response: "+err.Error()).JSON(c)
	}

	// Safely extract the message text
	if len(apiResponse.Output.Message.Content) == 0 {
		logger.Zap.Errorw("No content in Bedrock response", "response", apiResponse)
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
