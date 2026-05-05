package deepseek

import (
	"context"
	"errors"

	"github.com/cohesion-org/deepseek-go"
	"github.com/dtylman/goai/chat"
)

// Client implements chat.Client using the DeepSeek API.
type Client struct {
	ds    *deepseek.Client
	model string
}

// New creates a new DeepSeek chat client.
func New(cfg Config) (*Client, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("deepseek: API key is required")
	}
	model := cfg.Model
	if model == "" {
		model = deepseek.DeepSeekChat
	}
	return &Client{
		ds:    deepseek.NewClient(cfg.APIKey),
		model: model,
	}, nil
}

// Chat sends a chat completion request to DeepSeek.
func (c *Client) Chat(ctx context.Context, req *chat.Request) (*chat.Response, error) {
	dsReq := c.buildRequest(req)

	resp, err := c.ds.CreateChatCompletion(ctx, dsReq)
	if err != nil {
		return nil, err
	}
	if len(resp.Choices) == 0 {
		return nil, errors.New("deepseek: no choices returned")
	}

	return &chat.Response{
		Content: resp.Choices[0].Message.Content,
	}, nil
}

func (c *Client) buildRequest(req *chat.Request) *deepseek.ChatCompletionRequest {
	dsReq := &deepseek.ChatCompletionRequest{
		Model:    c.resolveModel(req.Model),
		Messages: make([]deepseek.ChatCompletionMessage, 0, len(req.Messages)+1),
	}

	if req.Schema != nil {
		dsReq.JSONMode = true
		dsReq.ResponseFormat = &deepseek.ResponseFormat{Type: "json_object"}
		dsReq.Messages = append(dsReq.Messages, deepseek.ChatCompletionMessage{
			Role:    deepseek.ChatMessageRoleSystem,
			Content: "Respond with a JSON object matching this schema: " + string(req.Schema.RawMessage()),
		})
	}

	for _, msg := range req.Messages {
		dsReq.Messages = append(dsReq.Messages, deepseek.ChatCompletionMessage{
			Role:    toDeepSeekRole(msg.Role),
			Content: msg.Content,
		})
	}

	return dsReq
}

func (c *Client) resolveModel(override string) string {
	if override != "" {
		return override
	}
	return c.model
}

func toDeepSeekRole(role chat.Role) string {
	switch role {
	case chat.RoleUser:
		return deepseek.ChatMessageRoleUser
	case chat.RoleAssistant:
		return deepseek.ChatMessageRoleAssistant
	case chat.RoleSystem:
		return deepseek.ChatMessageRoleSystem
	case chat.RoleTool:
		return deepseek.ChatMessageRoleTool
	default:
		return string(role)
	}
}
