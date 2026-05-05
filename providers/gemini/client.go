package gemini

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/dtylman/goai/chat"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// Client implements chat.Client using the Google Gemini API.
type Client struct {
	client *genai.Client
	model  string
}

// New creates a new Gemini chat client.
func New(cfg Config) (*Client, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("gemini: API key is required")
	}
	if cfg.Model == "" {
		return nil, errors.New("gemini: model is required")
	}
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(cfg.APIKey))
	if err != nil {
		return nil, fmt.Errorf("gemini: %w", err)
	}
	return &Client{
		client: client,
		model:  cfg.Model,
	}, nil
}

// Close releases resources held by the client.
func (c *Client) Close() error {
	return c.client.Close()
}

// Chat sends a chat completion request to Gemini.
func (c *Client) Chat(ctx context.Context, req *chat.Request) (*chat.Response, error) {
	model := c.client.GenerativeModel(c.resolveModel(req.Model))

	if req.Schema != nil {
		model.ResponseMIMEType = "application/json"
		model.ResponseSchema = toGeminiSchema(req.Schema)
	}

	var userParts []genai.Part
	var systemInstruction string

	for _, msg := range req.Messages {
		if msg.Role == chat.RoleSystem {
			if systemInstruction != "" {
				systemInstruction += "\n"
			}
			systemInstruction += msg.Content
		} else {
			userParts = append(userParts, genai.Text(msg.Content))
		}
	}

	if systemInstruction != "" {
		model.SystemInstruction = genai.NewUserContent(genai.Text(systemInstruction))
	}

	resp, err := model.GenerateContent(ctx, userParts...)
	if err != nil {
		return nil, fmt.Errorf("gemini: %w", err)
	}

	content := extractContent(resp)
	return &chat.Response{Content: content}, nil
}

func (c *Client) resolveModel(override string) string {
	if override != "" {
		return override
	}
	return c.model
}

func extractContent(resp *genai.GenerateContentResponse) string {
	if resp == nil || len(resp.Candidates) == 0 {
		return ""
	}
	candidate := resp.Candidates[0]
	if candidate.Content == nil {
		return ""
	}
	var content string
	for _, part := range candidate.Content.Parts {
		if text, ok := part.(genai.Text); ok {
			content += string(text)
		}
	}
	return content
}

func toGeminiSchema(schema *chat.JSONSchema) *genai.Schema {
	if schema == nil {
		return nil
	}
	gs := &genai.Schema{
		Description: schema.Description,
		Required:    schema.Required,
	}

	switch schema.Type {
	case "object":
		gs.Type = genai.TypeObject
	case "array":
		gs.Type = genai.TypeArray
	case "string":
		gs.Type = genai.TypeString
	case "number":
		gs.Type = genai.TypeNumber
	case "integer":
		gs.Type = genai.TypeInteger
	case "boolean":
		gs.Type = genai.TypeBoolean
	}

	if schema.Items != nil {
		gs.Items = toGeminiSchema(schema.Items)
	}

	if len(schema.Properties) > 0 {
		gs.Properties = make(map[string]*genai.Schema, len(schema.Properties))
		for name, prop := range schema.Properties {
			gs.Properties[name] = toGeminiSchema(&prop)
		}
	}

	return gs
}

// toJSON is a helper for debugging (unused in production).
func toJSON(v any) string {
	data, _ := json.Marshal(v)
	return string(data)
}
