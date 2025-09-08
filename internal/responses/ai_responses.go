package responses

// BedrockRewriteResponse represents the response from the Bedrock API rewrite
// @Description Response from the Bedrock API rewrite
type BedrockRewriteResponse struct {
	Message string `json:"message"`
	Tokens  struct {
		Input  int `json:"input"`
		Output int `json:"output"`
	} `json:"tokens"`
}

type BedrockDetectFormFieldsResponse struct {
	Fields []Field `json:"fields"`
	Tokens struct {
		Input  int `json:"input"`
		Output int `json:"output"`
	} `json:"tokens"`
}

type Field struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"`
	Label      string          `json:"label"`
	Required   bool            `json:"required"`
	Confidence float64         `json:"confidence"`
	BBox       BoundingBox     `json:"bbox"`
	BBoxNorm   BoundingBoxNorm `json:"bboxNorm"`
}

type BoundingBox struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w"`
	H int `json:"h"`
}

type BoundingBoxNorm struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	W float64 `json:"w"`
	H float64 `json:"h"`
}
