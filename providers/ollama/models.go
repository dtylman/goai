package ollama

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/ollama/ollama/api"
)

// ListModels queries the local Ollama server for available models.
// It accepts an optional base URL; if empty, it uses OLLAMA_HOST or localhost.
func ListModels(ctx context.Context, baseURL string) ([]string, error) {
	c, err := newAPIClient(baseURL)
	if err != nil {
		return nil, err
	}
	resp, err := c.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("ollama: list models: %w", err)
	}
	models := make([]string, 0, len(resp.Models))
	for _, m := range resp.Models {
		models = append(models, m.Name)
	}
	return models, nil
}

func newAPIClient(baseURL string) (*api.Client, error) {
	if baseURL != "" {
		base, err := url.Parse(baseURL)
		if err != nil {
			return nil, fmt.Errorf("ollama: invalid base URL: %w", err)
		}
		return api.NewClient(base, http.DefaultClient), nil
	}
	if os.Getenv("OLLAMA_HOST") == "" {
		os.Setenv("OLLAMA_HOST", "http://localhost:11434")
	}
	c, err := api.ClientFromEnvironment()
	if err != nil {
		return nil, fmt.Errorf("ollama: %w", err)
	}
	return c, nil
}
