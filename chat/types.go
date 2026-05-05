package chat

// Role represents the role of a message sender in a conversation.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// Message represents a single message in a conversation.
type Message struct {
	Role    Role
	Content string
}

// Request represents a chat completion request.
type Request struct {
	// Model optionally overrides the provider's default model.
	Model string
	// Messages is the conversation history.
	Messages []Message
	// Schema is set automatically by ChatInto. When non-nil, providers
	// should request structured JSON output from the model.
	Schema *JSONSchema
}

// Response represents a chat completion response.
type Response struct {
	// Content is the generated text from the model.
	Content string
}
