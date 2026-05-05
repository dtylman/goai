package ollama

// Config holds Ollama provider configuration.
type Config struct {
	// Model is the default model to use (e.g., "qwen2.5:1.5b", "llama3").
	Model string
	// BaseURL optionally overrides the Ollama server URL.
	// If empty, the client uses the OLLAMA_HOST environment variable or localhost.
	BaseURL string
}
