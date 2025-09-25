package requests

// BedrockRewriteRequest is used to validate Bedrock API rewrite request body
type BedrockRewriteRequest struct {
	Draft             string `json:"draft" validate:"required"`
	MarinaName        string `json:"marinaName" validate:"required"`
	UserName          string `json:"userName" validate:"required"`
	Tone              string `json:"tone" validate:"required"`
	ExtraInstructions string `json:"extraInstructions"`
}

type BedrockDetectFormFieldsRequest struct {
	Document   string `json:"document" validate:"required"`
	PageNumber int    `json:"pageNumber" validate:"required"`
}
