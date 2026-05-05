package deepseek

// Config holds DeepSeek provider configuration.
type Config struct {
	// APIKey is the DeepSeek API key.
	APIKey string
	// Model is the default model to use (e.g., "deepseek-reasoner", "deepseek-chat").
	Model string
}
