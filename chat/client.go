package chat

import (
	"context"
	"encoding/json"
	"fmt"
)

// Client is the interface that provider adapters implement.
// When req.Schema is non-nil, the provider should request JSON output
// from the model using whatever native mechanism it supports.
type Client interface {
	Chat(ctx context.Context, req *Request) (*Response, error)
}

// ChatInto generates a JSON schema from the target type, attaches it to
// the request, calls Chat, and decodes the response into target.
func ChatInto(ctx context.Context, c Client, req *Request, target any) (*Response, error) {
	schema, err := NewJSONSchema(target)
	if err != nil {
		return nil, fmt.Errorf("schema generation: %w", err)
	}
	req.Schema = schema

	resp, err := c.Chat(ctx, req)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(resp.Content), target); err != nil {
		return resp, fmt.Errorf("decode response: %w", err)
	}
	return resp, nil
}
