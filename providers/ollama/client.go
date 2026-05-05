package ollama

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/dtylman/goai/chat"
	"github.com/ollama/ollama/api"
)

// Client implements chat.Client using the Ollama API.
type Client struct {
	c     *api.Client
	model string
}

// New creates a new Ollama chat client.
func New(cfg Config) (*Client, error) {
	if cfg.Model == "" {
		return nil, errors.New("ollama: model is required")
	}

	var c *api.Client
	var err error

	if cfg.BaseURL != "" {
		base, parseErr := url.Parse(cfg.BaseURL)
		if parseErr != nil {
			return nil, fmt.Errorf("ollama: invalid base URL: %w", parseErr)
		}
		c = api.NewClient(base, http.DefaultClient)
	} else {
		if os.Getenv("OLLAMA_HOST") == "" {
			os.Setenv("OLLAMA_HOST", "http://localhost:11434")
		}
		c, err = api.ClientFromEnvironment()
		if err != nil {
			return nil, fmt.Errorf("ollama: %w", err)
		}
	}

	return &Client{c: c, model: cfg.Model}, nil
}

// Chat sends a chat completion request to Ollama.
func (c *Client) Chat(ctx context.Context, req *chat.Request) (*chat.Response, error) {
	ollamaReq := c.buildRequest(req)

	var result *api.ChatResponse
	respFunc := func(resp api.ChatResponse) error {
		if resp.Done {
			result = &resp
		}
		return nil
	}

	if err := c.c.Chat(ctx, ollamaReq, respFunc); err != nil {
		return nil, fmt.Errorf("ollama: %w", err)
	}
	if result == nil {
		return nil, errors.New("ollama: no response received")
	}

	return &chat.Response{
		Content: result.Message.Content,
	}, nil
}

func (c *Client) buildRequest(req *chat.Request) *api.ChatRequest {
	ollamaReq := &api.ChatRequest{
		Model:    c.resolveModel(req.Model),
		Messages: make([]api.Message, 0, len(req.Messages)),
		Stream:   new(bool),
	}

	if req.Schema != nil {
		ollamaReq.Format = json.RawMessage(req.Schema.RawMessage())
	}

	for _, msg := range req.Messages {
		ollamaReq.Messages = append(ollamaReq.Messages, api.Message{
			Role:    string(msg.Role),
			Content: msg.Content,
		})
	}

	return ollamaReq
}

func (c *Client) resolveModel(override string) string {
	if override != "" {
		return override
	}
	return c.model
}
