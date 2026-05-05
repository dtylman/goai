package goai

import (
	"fmt"

	"github.com/dtylman/goai/chat"
	"github.com/dtylman/goai/providers/deepseek"
	"github.com/dtylman/goai/providers/gemini"
	"github.com/dtylman/goai/providers/ollama"
)

// NewClient creates a chat.Client from a vendor name, model, and API key.
// Supported vendors: "deepseek", "gemini", "ollama".
func NewClient(vendor, model, apiKey string) (chat.Client, error) {
	switch vendor {
	case "deepseek":
		return deepseek.New(deepseek.Config{
			APIKey: apiKey,
			Model:  model,
		})
	case "gemini":
		return gemini.New(gemini.Config{
			APIKey: apiKey,
			Model:  model,
		})
	case "ollama":
		return ollama.New(ollama.Config{
			Model: model,
		})
	default:
		return nil, fmt.Errorf("unsupported vendor: %q", vendor)
	}
}
