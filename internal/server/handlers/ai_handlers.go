package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/textract"
	"github.com/aws/aws-sdk-go-v2/service/textract/types"
	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/pkg/token"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type AIHandler struct {
	server *s.Server
}

func NewAIHandler(server *s.Server) *AIHandler {
	return &AIHandler{server: server}
}

// getMarinaIDFromContext extracts marina ID from JWT token and user data
func (h *AIHandler) getMarinaIDFromContext(c echo.Context) (uuid.UUID, error) {
	userToken := c.Get("user").(*jwt.Token)
	if userToken == nil {
		return uuid.Nil, fmt.Errorf("authentication required")
	}

	claims := userToken.Claims.(*token.JwtCustomClaims)
	userID := claims.ID

	// Fetch the user's current marina_id from the database
	user, err := h.server.DB.Queries().GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to load user: %w", err)
	}

	return user.MarinaID, nil
}

// hasDocumentPaidPlan checks if the marina has a paid document plan
func (h *AIHandler) hasDocumentPaidPlan(ctx context.Context, marinaID uuid.UUID) (bool, error) {
	// Get marina's document plan
	documentPlan, err := h.server.DB.Queries().GetMarinaDocumentPlan(ctx, marinaID)
	if err != nil {
		return false, fmt.Errorf("failed to get document plan: %w", err)
	}

	// Allow prepaid plans
	if documentPlan.Name == "Prepaid" {
		return true, nil
	}

	// Block free plans and pay-as-you-go plans
	if documentPlan.MonthlyPrice == 0 || documentPlan.Name == "Pay as You Go" {
		return false, nil
	}

	// Allow paid monthly plans
	return documentPlan.MonthlyPrice > 0, nil
}

// hasAnyPaidPlan checks if the marina has any paid plan (storage, notes/messages, or document)
func (h *AIHandler) hasAnyPaidPlan(ctx context.Context, marinaID uuid.UUID) (bool, error) {
	// Get marina to access plan IDs
	marina, err := h.server.DB.Queries().GetMarinaByID(ctx, marinaID)
	if err != nil {
		return false, fmt.Errorf("failed to get marina: %w", err)
	}

	// Check document plan
	documentPlan, err := h.server.DB.Queries().GetDocumentPlanByID(ctx, marina.DocumentPlanID)
	if err == nil {
		// Allow prepaid or paid monthly plans (excluding pay-as-you-go)
		if documentPlan.Name == "Prepaid" || (documentPlan.MonthlyPrice > 0 && documentPlan.Name != "Pay as You Go") {
			return true, nil
		}
	}

	// Check storage plan
	storagePlan, err := h.server.DB.Queries().GetStoragePlanByID(ctx, marina.StoragePlanID)
	if err == nil {
		// Allow prepaid or paid monthly plans (excluding pay-as-you-go)
		if storagePlan.Name == "Prepaid" || (storagePlan.MonthlyPrice > 0 && storagePlan.Name != "Pay as You Go") {
			return true, nil
		}
	}

	// Check notes/messages plan
	notesMessagesPlan, err := h.server.DB.Queries().GetNotesMessagesPlanByID(ctx, marina.NotesMessagesPlanID)
	if err == nil {
		// Allow prepaid or paid monthly plans (excluding pay-as-you-go)
		if notesMessagesPlan.Name == "Prepaid" || (notesMessagesPlan.MonthlyPrice > 0 && notesMessagesPlan.Name != "Pay as You Go") {
			return true, nil
		}
	}

	return false, nil
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
	ctx := c.Request().Context()

	// Get marina ID from user context
	marinaID, err := h.getMarinaIDFromContext(c)
	if err != nil {
		logger.Zap.Errorw("Failed to get marina ID", "error", err)
		return responses.NewErrorResponse(http.StatusUnauthorized, "Authentication required").JSON(c)
	}

	// Check if marina has any paid plan
	hasPaidPlan, err := h.hasAnyPaidPlan(ctx, marinaID)
	if err != nil {
		logger.Zap.Errorw("Failed to check paid plan status", "error", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to verify subscription status").JSON(c)
	}
	if !hasPaidPlan {
		return responses.NewErrorResponse(http.StatusForbidden, "AI compose message is only available on paid plans. Please upgrade your plan to access this feature.").JSON(c)
	}

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

	// Increment AI compose message usage counter
	_, err = h.server.DB.Queries().IncrementMarinaAIComposeMessageUsage(ctx, db.IncrementMarinaAIComposeMessageUsageParams{
		ID:      marinaID,
		Column2: 1,
	})
	if err != nil {
		logger.Zap.Errorw("Failed to increment AI compose message usage", "error", err)
		// Don't fail the request if usage tracking fails
	}

	// Return response to client
	return c.JSON(http.StatusOK, bedrockResponse)
}

// DetectFormFieldsHandler detects form fields in an image using AWS Textract
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
	ctx := c.Request().Context()

	// Get marina ID from user context
	marinaID, err := h.getMarinaIDFromContext(c)
	if err != nil {
		logger.Zap.Errorw("Failed to get marina ID", "error", err)
		return responses.NewErrorResponse(http.StatusUnauthorized, "Authentication required").JSON(c)
	}

	// Check if marina has a paid document plan
	hasPaidDocumentPlan, err := h.hasDocumentPaidPlan(ctx, marinaID)
	if err != nil {
		logger.Zap.Errorw("Failed to check document plan status", "error", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to verify subscription status").JSON(c)
	}
	if !hasPaidDocumentPlan {
		return responses.NewErrorResponse(http.StatusForbidden, "AI form field detection is only available on paid document plans. Please upgrade your plan to access this feature.").JSON(c)
	}

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

	// Create AWS configuration with credentials
	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		awsRegion = "us-east-1"
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(awsRegion),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			os.Getenv("AWS_ACCESS_KEY_ID"),
			os.Getenv("AWS_SECRET_ACCESS_KEY"),
			"",
		)),
	)
	if err != nil {
		logger.Zap.Errorw("Failed to load AWS config", "error", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to configure AWS client").JSON(c)
	}

	// Create Textract client
	textractClient := textract.NewFromConfig(cfg)

	// Call AnalyzeDocument API
	input := &textract.AnalyzeDocumentInput{
		Document: &types.Document{
			Bytes: fileBytes,
		},
		FeatureTypes: []types.FeatureType{
			types.FeatureTypeForms,
		},
	}

	result, err := textractClient.AnalyzeDocument(context.TODO(), input)
	if err != nil {
		logger.Zap.Errorw("Failed to call Textract API", "error", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to analyze document: "+err.Error()).JSON(c)
	}

	// Build a map of block IDs to blocks for easy lookup
	blockMap := make(map[string]types.Block)
	for _, block := range result.Blocks {
		blockMap[*block.Id] = block
	}

	// Define page size constants (standard letter size)
	const pageWidthPx = 850
	const pageHeightPx = 1100

	// Extract form fields (KEY_VALUE_SET blocks)
	type BoundingBox struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
		W float64 `json:"w"`
		H float64 `json:"h"`
	}

	type FormField struct {
		ID         string      `json:"id"`
		Type       string      `json:"type"`
		Label      string      `json:"label"`
		Required   bool        `json:"required"`
		Confidence float64     `json:"confidence"`
		Bbox       BoundingBox `json:"bbox"`
		BboxNorm   BoundingBox `json:"bboxNorm"`
	}

	type PageSize struct {
		WidthPx  int `json:"widthPx"`
		HeightPx int `json:"heightPx"`
	}

	type Page struct {
		PageNumber int         `json:"pageNumber"`
		PageSize   PageSize    `json:"pageSize"`
		Fields     []FormField `json:"fields"`
	}

	type DetectFormFieldsResponse struct {
		Pages []Page `json:"pages"`
	}

	var fields []FormField
	fieldID := 1

	// First, identify SELECTION_ELEMENT blocks (checkboxes/radio buttons)
	selectionElements := make(map[string]types.Block)
	for _, block := range result.Blocks {
		if block.BlockType == types.BlockTypeSelectionElement {
			selectionElements[*block.Id] = block
		}
	}

	// Process KEY_VALUE_SET blocks with EntityType "KEY"
	for _, block := range result.Blocks {
		if block.BlockType == types.BlockTypeKeyValueSet && len(block.EntityTypes) > 0 && block.EntityTypes[0] == types.EntityTypeKey {
			// Get the label text from child blocks
			labelText := ""
			if block.Relationships != nil {
				for _, rel := range block.Relationships {
					if rel.Type == types.RelationshipTypeChild {
						for _, childID := range rel.Ids {
							if childBlock, ok := blockMap[childID]; ok {
								if childBlock.Text != nil && *childBlock.Text != "" {
									if labelText != "" {
										labelText += " "
									}
									labelText += *childBlock.Text
								}
							}
						}
					}
				}
			}

			// Find the VALUE block (the input field coordinates)
			var valueBlock *types.Block
			if block.Relationships != nil {
				for _, rel := range block.Relationships {
					if rel.Type == types.RelationshipTypeValue && len(rel.Ids) > 0 {
						if vBlock, ok := blockMap[rel.Ids[0]]; ok {
							valueBlock = &vBlock
							break
						}
					}
				}
			}

			// Skip if no VALUE block found
			if valueBlock == nil || valueBlock.Geometry == nil {
				continue
			}

			// Calculate bounding box from VALUE block
			var bbox, bboxNorm BoundingBox
			var bboxHeight float64

			// Prefer polygon coordinates if available, otherwise use BoundingBox
			if len(valueBlock.Geometry.Polygon) >= 4 {
				polygon := valueBlock.Geometry.Polygon
				minX, maxX := float64(polygon[0].X), float64(polygon[0].X)
				minY, maxY := float64(polygon[0].Y), float64(polygon[0].Y)

				for _, point := range polygon[1:] {
					x, y := float64(point.X), float64(point.Y)
					if x < minX {
						minX = x
					}
					if x > maxX {
						maxX = x
					}
					if y < minY {
						minY = y
					}
					if y > maxY {
						maxY = y
					}
				}

				bboxNorm = BoundingBox{X: minX, Y: minY, W: maxX - minX, H: maxY - minY}
				bbox = BoundingBox{
					X: minX * pageWidthPx,
					Y: minY * pageHeightPx,
					W: (maxX - minX) * pageWidthPx,
					H: (maxY - minY) * pageHeightPx,
				}
				bboxHeight = maxY - minY
			} else if valueBlock.Geometry.BoundingBox != nil {
				geomBox := valueBlock.Geometry.BoundingBox
				bboxNorm = BoundingBox{
					X: float64(geomBox.Left),
					Y: float64(geomBox.Top),
					W: float64(geomBox.Width),
					H: float64(geomBox.Height),
				}
				bbox = BoundingBox{
					X: float64(geomBox.Left) * pageWidthPx,
					Y: float64(geomBox.Top) * pageHeightPx,
					W: float64(geomBox.Width) * pageWidthPx,
					H: float64(geomBox.Height) * pageHeightPx,
				}
				bboxHeight = float64(geomBox.Height)
			} else {
				continue
			}

			// Check if VALUE block contains SELECTION_ELEMENT (checkbox/radio)
			fieldType := "text"
			if valueBlock.Relationships != nil {
				for _, rel := range valueBlock.Relationships {
					if rel.Type == types.RelationshipTypeChild {
						for _, childID := range rel.Ids {
							if _, ok := selectionElements[childID]; ok {
								// Found a selection element - determine if checkbox or radio
								fieldType = classifySelectionType(labelText)
								break
							}
						}
					}
					if fieldType == "checkbox" || fieldType == "radio" {
						break
					}
				}
			}

			// If not a selection element, classify using heuristics
			if fieldType == "text" {
				fieldType = classifyFieldType(labelText, bboxHeight)
			}

			// Create the field - use KEY block confidence
			var confidence float64
			if block.Confidence != nil {
				confidence = float64(*block.Confidence) / 100.0 // Convert to 0-1 scale
			}

			field := FormField{
				ID:         fmt.Sprintf("field_%d", fieldID),
				Type:       fieldType,
				Label:      labelText,
				Required:   false,
				Confidence: confidence,
				Bbox:       bbox,
				BboxNorm:   bboxNorm,
			}

			fields = append(fields, field)
			fieldID++
		}
	}

	// Build structured response
	response := DetectFormFieldsResponse{
		Pages: []Page{
			{
				PageNumber: 1,
				PageSize: PageSize{
					WidthPx:  pageWidthPx,
					HeightPx: pageHeightPx,
				},
				Fields: fields,
			},
		},
	}

	// Increment AI form detection usage counter
	_, err = h.server.DB.Queries().IncrementMarinaAIFormDetectionUsage(ctx, db.IncrementMarinaAIFormDetectionUsageParams{
		ID:      marinaID,
		Column2: 1,
	})
	if err != nil {
		logger.Zap.Errorw("Failed to increment AI form detection usage", "error", err)
		// Don't fail the request if usage tracking fails
	}

	// Return the response as JSON
	return c.JSON(http.StatusOK, map[string]interface{}{
		"pages": response.Pages,
		"tokens": map[string]int{
			"input":  0, // Textract doesn't provide token counts
			"output": 0,
		},
		"stopReason": "end_turn", // Textract always completes successfully
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

// classifySelectionType determines if a selection element is a checkbox or radio button
func classifySelectionType(label string) string {
	labelLower := strings.ToLower(label)

	// Radio buttons typically have labels like "Yes/No", "Male/Female", "Option 1/Option 2"
	// or appear in groups with similar naming patterns
	if strings.Contains(labelLower, "yes") || strings.Contains(labelLower, "no") ||
		strings.Contains(labelLower, "male") || strings.Contains(labelLower, "female") ||
		strings.Contains(labelLower, "option") || strings.Contains(labelLower, "choice") {
		return "radio"
	}

	// Default to checkbox for selection elements
	return "checkbox"
}

// classifyFieldType determines the field type based on the label text and bounding box dimensions
func classifyFieldType(label string, bboxHeight float64) string {
	labelLower := strings.ToLower(label)

	// Check for signature fields
	if strings.Contains(labelLower, "signature") || strings.Contains(labelLower, "sign") {
		return "signature"
	}

	// Check for date fields
	if strings.Contains(labelLower, "date") {
		return "date"
	}

	// Check for email fields
	if strings.Contains(labelLower, "email") || strings.Contains(labelLower, "e-mail") {
		return "email"
	}

	// Check for phone fields
	if strings.Contains(labelLower, "phone") || strings.Contains(labelLower, "tel") ||
		strings.Contains(labelLower, "mobile") || strings.Contains(labelLower, "cell") {
		return "phone"
	}

	// Check for number fields
	if strings.Contains(labelLower, "number") || strings.Contains(labelLower, "rate") ||
		strings.Contains(labelLower, "value") || strings.Contains(labelLower, "price") ||
		strings.Contains(labelLower, "amount") || strings.Contains(labelLower, "cost") {
		return "number"
	}

	// Check for initials
	if strings.Contains(labelLower, "initial") {
		return "initials"
	}

	// Check for textarea based on height (if field is tall, likely a textarea)
	if bboxHeight > 0.05 { // If field height is more than 5% of page height
		return "textarea"
	}

	// Default to text input
	return "text"
}
