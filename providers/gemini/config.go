package gemini

// Config holds Gemini provider configuration.
type Config struct {
	// APIKey is the Google AI API key.
	APIKey string
	// Model is the default model to use (e.g., "gemini-2.0-flash", "gemini-1.5-pro").
	Model string
}
