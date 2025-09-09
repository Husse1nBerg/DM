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
