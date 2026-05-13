# goai Design

## Purpose

`goai` is a unified Go package for consuming AI chat models through a consistent API, while still allowing provider-specific implementations and higher-level task workflows.

The library is primarily intended for personal use across multiple projects, but it should remain clean, reusable, and open-source friendly.

## Scope

### In scope

- Chat-only support
- Provider-specific chat clients behind a shared chat abstraction
- Structured output as a first-class feature
- Embedded default prompt templates with caller overrides
- Task-focused packages built on top of the chat core
- Explicit provider config structs
- Official SDKs preferred for provider implementations

### Out of scope (for now)

- Embeddings
- Image generation
- Audio transcription
- Tool calling
- Streaming responses

## Design Principles

1. Chat first

The core API should optimize for request/response chat interactions. Future capabilities should not distort the chat design.

2. Shared surface, separate adapters

Common request and response types should be shared, while provider-specific behavior stays in provider packages.

3. Structured output is normal

Producing typed JSON responses should be a primary workflow, not an afterthought.

4. Tasks are first-class

The library should contain reusable higher-level packages for concrete workflows such as translation.

5. Defaults inside, overrides outside

The library should ship embedded prompt templates and reasonable defaults, but callers must be able to replace them without forking the library.

6. Avoid fake uniformity

Providers differ. The public API should normalize what is truly shared without hiding real provider differences when they matter.

## Layered Architecture

The library should have three layers.

### 1. Core chat layer

This layer defines the public request and response model used by all providers and tasks.

Responsibilities:

- Message and role types
- Chat request and response types
- Structured output support
- JSON schema generation for Go structs
- Shared interfaces and common errors

### 2. Provider adapters

This layer translates the shared chat model into specific provider SDK calls.

Responsibilities:

- Provider client construction
- Request translation
- Response translation
- Structured output integration using provider-native features when possible
- Provider-specific options where needed

### 3. Task packages

This layer contains reusable workflows such as translation and other concrete AI tasks.

Responsibilities:

- Task-specific input and output types
- Prompt selection
- Multi-model orchestration and routing
- Task orchestration across steps and data segments
- Validation and post-processing specific to the task

## Proposed Package Layout

```text
goai/
  design.md
  README.md
  go.mod
  chat/
    types.go
    client.go
    schema.go
    errors.go
  prompts/
    prompts.go
    store.go
  providers/
    deepseek/
      client.go
      config.go
    gemini/
      client.go
      config.go
    ollama/
      client.go
      config.go
  tasks/
    translate/
      task.go
      types.go
      prompts.go
      options.go
      embedded/
        system.tmpl
        user.tmpl
```

Notes:

- `chat` is the stable low-level package.
- `providers` contains transport adapters only.
- `tasks` contains higher-level workflows.
- `prompts` provides prompt rendering infrastructure and override mechanics only.
- Prompt template files live inside their owning task package.

## Core Chat API

The main capability in v1 is chat.

### Core types

```go
package chat

type Role string

const (
    RoleSystem    Role = "system"
    RoleUser      Role = "user"
    RoleAssistant Role = "assistant"
    RoleTool      Role = "tool"
)

type Message struct {
    Role    Role
    Content string
}

type Request struct {
    Model    string
    Messages []Message
    // Schema is set automatically by ChatInto. Providers use it to
    // request structured JSON output in their native format.
    Schema *JSONSchema
}

type Response struct {
    Content string
}

// Client is the interface that provider adapters implement.
// Providers only need to implement Chat. When Schema is set on the
// request, the provider should request JSON output from the model.
type Client interface {
    Chat(ctx context.Context, req *Request) (*Response, error)
}

// ChatInto is a convenience function that generates a JSON schema from
// the target type, attaches it to the request, calls Chat, and decodes
// the response into target.
func ChatInto(ctx context.Context, c Client, req *Request, target any) (*Response, error)
```

### Why `ChatInto` is a free function

`ChatInto` is a package-level function, not part of the `Client` interface.

This means providers only implement one method (`Chat`), and structured output is handled uniformly across all providers:

1. `ChatInto` generates a `JSONSchema` from the target type.
2. It sets `req.Schema` so the provider can use native JSON mode.
3. It calls `c.Chat(ctx, req)`.
4. It unmarshals `resp.Content` into the target.

Benefits:

- providers have minimal implementation surface
- structured output logic is consistent and tested once
- plain text remains the default path (just call `Chat` directly)
- the schema reaches the provider naturally via the request struct

## Structured Output

Structured output should be part of the core contract.

### Required behavior

When `ChatInto` is called with a target struct:

1. `ChatInto` generates a `JSONSchema` from the Go type and sets `req.Schema`.
2. The provider sees `req.Schema != nil` and requests JSON output using whatever native mechanism it supports.
3. `ChatInto` decodes `resp.Content` into the target value.
4. The raw text content is still returned in `Response`.

### Design constraints

- Schema generation must support nested structs, arrays, maps, pointers, and tags.
- Providers may differ in how strongly they enforce schemas.
- The public contract should promise best-effort typed decoding, not provider-perfect determinism.

### Error handling

Structured output failures should be distinguishable from transport failures.

Suggested error categories:

- invalid request
- provider transport error
- unsupported provider feature
- schema generation error
- decode error

## Provider Adapters

Each provider gets its own package and config struct.

### Provider package responsibilities

- Own the provider-specific config type
- Own the SDK dependency
- Implement `chat.Client`
- Translate shared requests to provider requests
- Decode provider responses back into shared responses

### Example config shapes

```go
package deepseek

type Config struct {
    APIKey string
    Model  string
}
```

```go
package ollama

type Config struct {
    Model   string
    BaseURL string
}
```

### Construction style

Use explicit constructors in each provider package.

Examples:

```go
client, err := deepseek.New(config)
client, err := gemini.New(config)
client, err := ollama.New(config)
```

The library should not depend on a string-based global factory as the primary API.

A registry or config-driven factory can be added later if needed, but explicit construction should remain the default.

## Prompt System

Prompt templates are part of the library and should be embedded by default.

### Goals

- Ship task-ready defaults inside the repo
- Allow callers to override prompts without copying task logic
- Keep prompt lookup stable and predictable

### Prompt model

Prompts should be addressed by stable IDs, not scattered raw strings.

Example prompt IDs:

- `translate/system/default`
- `translate/user/default`

### Resolution order

Prompt lookup should use this order:

1. caller override
2. task-provided override
3. embedded library default

### Template format

Use Go templates for prompt rendering.

Requirements:

- strict handling of missing template keys
- small parameter maps or typed task render inputs
- embedded default prompt files via `embed`

### Separation of concerns

Prompt templates are data owned by the task that uses them.
The `prompts` package provides only the rendering and override mechanics.

This keeps templates co-located with the logic that depends on them, while still allowing callers to replace them without forking the task.

## Task Packages

Task packages are a deliberate part of this library.

They should sit above the generic chat layer and below application code.

### Translation package

`tasks/translate` should be the first concrete workflow package.

It should own:

- translation input types
- translation result types
- prompt selection and rendering
- task-specific options
- post-processing and validation where necessary

### Translation design direction

The translation package should be generic enough for normal translation workflows, while still supporting richer project-specific metadata through options.

This means:

- default prompts should handle normal translation well
- book or domain metadata should be optional input, not a hard dependency
- callers should be able to override prompts or options for project-specific behavior

### Context-aware translation

Based on existing usage (saatool), the translator should support context-rich requests:

1. **Project context** — metadata about the work being translated (title, author, genre, synopsis, characters, terminology). This feeds into the system prompt so the model understands the domain and style.

2. **Previous translations** — when translating paragraph N, the request should include the N-k previously translated paragraphs as context. This keeps terminology, voice, and references consistent across the document.

3. **Context window budget** — the number of previous paragraphs to include should be configurable per task invocation. The caller knows the model's token limits and can tune accordingly.

4. **Translation styles** — the same task should support multiple prompt flavors (strict, literary, academic, etc.) selectable at construction or per-request.

### Translation workflow steps

A full translation cycle may include multiple distinct steps, each potentially using a different model:

1. **Translate** — produce initial translation of the target paragraph
2. **Proofread** — review translation for grammar and naturalness
3. **Fix** — re-translate a paragraph flagged as poor quality

Each step has its own prompt template.

### Translation input model

The request should carry enough context for the model to do its best work:

```go
package translate

type Request struct {
    // SourceLanguage is the language of the input text.
    SourceLanguage string
    // TargetLanguage is the language of the output text.
    TargetLanguage string
    // Text is the paragraph to translate.
    Text string
    // PreviousSource contains preceding source paragraphs for context.
    PreviousSource []string
    // PreviousTarget contains the corresponding previous translations.
    PreviousTarget []string
    // Style selects the prompt flavor (e.g., "strict", "literary").
    Style string
}
```

### Project context as an option

Project-level metadata (book details, characters, etc.) should be supplied via task options, not baked into the request struct:

```go
type ProjectContext struct {
    Title      string
    Author     string
    Genre      string
    Synopsis   string
    Characters []Character
}

func WithProjectContext(ctx *ProjectContext) Option
func WithContextWindow(paragraphs int) Option
func WithStyle(style string) Option
```

This keeps the base request simple for ad-hoc translation while allowing rich context for document-scale workflows.

### Suggested translation API direction

```go
package translate

type Task struct {
    client  chat.Client
    options Options
}

func New(client chat.Client, opts ...Option) *Task

func (t *Task) Translate(ctx context.Context, req *Request) (*Result, error)
```

Tasks accept a single `chat.Client`. If different steps need different models, callers can create separate task instances or compose at a higher level.

### Data-driven routing

For cases where different data segments should use different models (e.g., simple vs. complex paragraphs), the task can define a classifier function:

```go
// Classifier decides which role to use for a given input segment.
type Classifier func(segment string) string
```

This is task-specific and optional. The default should be a no-op that sends everything to the same role. Callers can supply a classifier when they need cost/quality optimization per segment.

## Request Options

Provider-neutral request options should stay minimal.

Candidate fields that may be added if needed:

- temperature
- max output tokens
- provider-specific metadata hooks

These should only be introduced when at least two providers can support them coherently.

The base request should remain small.

## Error Model

Errors should be meaningful and easy to inspect.

Suggested approach:

- exported sentinel errors for common categories where useful
- wrapped underlying provider errors
- enough context to understand whether the failure occurred during request construction, provider execution, or output decoding

Examples of useful categories:

- invalid prompt template
- invalid chat request
- provider request failure
- structured decode failure

## Testing Strategy

The first testing focus should be local and deterministic.

Priority order:

1. JSON schema generation
2. prompt rendering and override resolution
3. task orchestration with mocked clients
4. provider request translation where the SDK allows easy unit tests

Avoid making network-dependent integration tests a requirement for basic correctness.

## Example Usage

### Plain chat

```go
client, err := deepseek.New(deepseek.Config{
    APIKey: os.Getenv("DEEPSEEK_API_KEY"),
    Model:  "deepseek-reasoner",
})

resp, err := client.Chat(ctx, &chat.Request{
    Messages: []chat.Message{
        {Role: chat.RoleSystem, Content: "You are concise."},
        {Role: chat.RoleUser, Content: "Summarize this paragraph."},
    },
})
```

### Structured output

```go
type Summary struct {
    Title   string   `json:"title"`
    Points  []string `json:"points"`
    Verdict string   `json:"verdict"`
}

var summary Summary

resp, err := chat.ChatInto(ctx, client, &chat.Request{
    Messages: []chat.Message{
        {Role: chat.RoleUser, Content: "Summarize this lecture as JSON."},
    },
}, &summary)
```

### Task usage

```go
translator := translate.New(client)

result, err := translator.Translate(ctx, &translate.Request{
    SourceLanguage: "en",
    TargetLanguage: "he",
    Text:           "The archive preserves several manuscript traditions.",
})
```

## Implementation Plan

### Phase 1

- Initialize the Go module
- Implement `chat` core types and schema generation
- Add tests for schema generation

### Phase 2

- Implement provider packages for DeepSeek, Gemini, and Ollama
- Normalize plain-text and structured-output behavior

### Phase 3

- Implement prompt store and embedded defaults
- Add override support

### Phase 4

- Implement `tasks/translate`
- Port useful prompt defaults from earlier projects
- Add tests using mocked chat clients

## Open Questions

These can stay unresolved until implementation begins.

1. Should `chat.Response` include provider metadata such as model name or token usage?
2. Should structured decode failures return partial raw content in a richer error type?
3. Should prompt overrides be passed as a store, per-task option, or both?
4. Should there be a small internal retry helper for transient provider failures?

## Summary

The intended shape of `goai` is:

- a small shared chat core
- explicit provider adapters using official SDKs
- embedded prompt defaults with caller overrides
- first-class task packages such as translation

This keeps the library useful for current personal workflows while still making it reusable and understandable as a public package.