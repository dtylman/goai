package translate_test

import (
"context"
"strings"
"testing"

"github.com/dtylman/goai/chat"
"github.com/dtylman/goai/tasks/translate"
)

type mockClient struct {
lastMessages []chat.Message
response     string
}

func (m *mockClient) Chat(_ context.Context, req *chat.Request) (*chat.Response, error) {
m.lastMessages = req.Messages
return &chat.Response{Content: m.response}, nil
}

func TestTranslate_Basic(t *testing.T) {
mock := &mockClient{response: "שלום עולם"}
task := translate.New(chat.SingleClient(mock))

result, err := task.Translate(context.Background(), &translate.Request{
SourceLanguage: "en",
TargetLanguage: "he",
Text:           "Hello world",
})
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if result.Text != "שלום עולם" {
t.Errorf("unexpected result: %q", result.Text)
}
sys := mock.lastMessages[0].Content
if !strings.Contains(sys, "en") || !strings.Contains(sys, "he") {
t.Errorf("system prompt missing language info: %s", sys)
}
}

func TestTranslate_WithProjectContext(t *testing.T) {
mock := &mockClient{response: "translated"}
task := translate.New(chat.SingleClient(mock), translate.WithProjectContext(&translate.ProjectContext{
Title: "The Great Novel",
Genre: "fiction",
}))

_, err := task.Translate(context.Background(), &translate.Request{
SourceLanguage: "en",
TargetLanguage: "he",
Text:           "It was a dark and stormy night.",
})
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
sys := mock.lastMessages[0].Content
if !strings.Contains(sys, "The Great Novel") {
t.Errorf("system prompt missing project title: %s", sys)
}
if !strings.Contains(sys, "fiction") {
t.Errorf("system prompt missing genre: %s", sys)
}
}

func TestTranslate_WithWritingStyle(t *testing.T) {
mock := &mockClient{response: "translated"}
task := translate.New(chat.SingleClient(mock), translate.WithProjectContext(&translate.ProjectContext{
Title:        "The Great Novel",
WritingStyle: "third-person omniscient, dark and introspective",
}))

_, err := task.Translate(context.Background(), &translate.Request{
SourceLanguage: "en",
TargetLanguage: "he",
Text:           "It was a dark and stormy night.",
})
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
sys := mock.lastMessages[0].Content
if !strings.Contains(sys, "third-person omniscient, dark and introspective") {
t.Errorf("system prompt missing writing style: %s", sys)
}
}

func TestTranslate_WithGlossary(t *testing.T) {
mock := &mockClient{response: "translated"}
task := translate.New(chat.SingleClient(mock), translate.WithProjectContext(&translate.ProjectContext{
Title: "Harry Potter",
Glossary: map[string]string{
"butter-beer": "בירת חמאה",
},
}))

_, err := task.Translate(context.Background(), &translate.Request{
SourceLanguage: "en",
TargetLanguage: "he",
Text:           "He ordered a butter-beer.",
})
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
sys := mock.lastMessages[0].Content
if !strings.Contains(sys, "butter-beer") {
t.Errorf("system prompt missing glossary term: %s", sys)
}
if !strings.Contains(sys, "בירת חמאה") {
t.Errorf("system prompt missing glossary translation: %s", sys)
}
}

func TestFix_WithWritingStyleAndGlossary(t *testing.T) {
mock := &mockClient{response: "fixed"}
task := translate.New(chat.SingleClient(mock), translate.WithProjectContext(&translate.ProjectContext{
Title:        "The Book",
WritingStyle: "fast-paced with witty dialogue",
Glossary:     map[string]string{"wand": "שרביט"},
}))

_, err := task.Fix(context.Background(), &translate.Request{
SourceLanguage: "en",
TargetLanguage: "he",
Text:           "He waved his wand.",
}, "bad translation")
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
sys := mock.lastMessages[0].Content
if !strings.Contains(sys, "fast-paced with witty dialogue") {
t.Errorf("fix system prompt missing writing style: %s", sys)
}
if !strings.Contains(sys, "wand") || !strings.Contains(sys, "שרביט") {
t.Errorf("fix system prompt missing glossary: %s", sys)
}
}

func TestGlossaryFormatted(t *testing.T) {
pc := &translate.ProjectContext{
Glossary: map[string]string{"wand": "שרביט", "spell": "כישוף"},
}
formatted := pc.GlossaryFormatted()
if !strings.Contains(formatted, `"wand"`) || !strings.Contains(formatted, `"שרביט"`) {
t.Errorf("GlossaryFormatted missing entries: %s", formatted)
}

empty := &translate.ProjectContext{}
if empty.GlossaryFormatted() != "" {
t.Error("expected empty string for empty glossary")
}
}

func TestTranslate_WithPreviousContext(t *testing.T) {
mock := &mockClient{response: "translated"}
task := translate.New(chat.SingleClient(mock))

_, err := task.Translate(context.Background(), &translate.Request{
SourceLanguage: "en",
TargetLanguage: "he",
Text:           "She nodded.",
PreviousSource: []string{"He asked her a question."},
PreviousTarget: []string{"הוא שאל אותה שאלה."},
})
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
user := mock.lastMessages[1].Content
if !strings.Contains(user, "He asked her a question.") {
t.Errorf("user prompt missing previous source: %s", user)
}
if !strings.Contains(user, "הוא שאל אותה שאלה.") {
t.Errorf("user prompt missing previous target: %s", user)
}
}

func TestTranslate_WithAutoProofread(t *testing.T) {
countingMock := &countingClient{responses: []string{"initial translation", "proofread result"}}
task := translate.New(chat.SingleClient(countingMock), translate.WithAutoProofread(true))

result, err := task.Translate(context.Background(), &translate.Request{
SourceLanguage: "en",
TargetLanguage: "he",
Text:           "Hello",
})
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if countingMock.callCount != 2 {
t.Errorf("expected 2 calls (translate + proofread), got %d", countingMock.callCount)
}
if result.Text != "proofread result" {
t.Errorf("expected proofread result, got %q", result.Text)
}
}

func TestProofread(t *testing.T) {
mock := &mockClient{response: "improved translation"}
task := translate.New(chat.SingleClient(mock))

result, err := task.Proofread(context.Background(), &translate.Request{
SourceLanguage: "en",
TargetLanguage: "he",
Text:           "Hello world",
}, "שלום עולמ")
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if result.Text != "improved translation" {
t.Errorf("unexpected result: %q", result.Text)
}
user := mock.lastMessages[1].Content
if !strings.Contains(user, "שלום עולמ") {
t.Errorf("user prompt missing current translation: %s", user)
}
}

func TestFix(t *testing.T) {
mock := &mockClient{response: "fixed translation"}
task := translate.New(chat.SingleClient(mock))

result, err := task.Fix(context.Background(), &translate.Request{
SourceLanguage: "en",
TargetLanguage: "he",
Text:           "Hello world",
}, "bad translation")
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if result.Text != "fixed translation" {
t.Errorf("unexpected result: %q", result.Text)
}
}

func TestTranslate_LiteraryStyle(t *testing.T) {
mock := &mockClient{response: "literary result"}
task := translate.New(chat.SingleClient(mock), translate.WithStyle("literary"))

_, err := task.Translate(context.Background(), &translate.Request{
SourceLanguage: "en",
TargetLanguage: "he",
Text:           "The night was dark.",
})
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
sys := mock.lastMessages[0].Content
if !strings.Contains(sys, "spirit") {
t.Errorf("expected literary style system prompt, got: %s", sys)
}
}

func TestTranslate_StyleOverridePerRequest(t *testing.T) {
mock := &mockClient{response: "result"}
task := translate.New(chat.SingleClient(mock), translate.WithStyle("strict"))

_, err := task.Translate(context.Background(), &translate.Request{
SourceLanguage: "en",
TargetLanguage: "he",
Text:           "Test",
Style:          "literary",
})
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
sys := mock.lastMessages[0].Content
if !strings.Contains(sys, "spirit") {
t.Errorf("expected literary prompt from per-request override, got: %s", sys)
}
}

func TestTranslate_PromptOverride(t *testing.T) {
mock := &mockClient{response: "result"}
task := translate.New(chat.SingleClient(mock),
translate.WithSystemPrompt("translate", "Custom system: translate {{.Text}} to {{.TargetLang}}"),
)

_, err := task.Translate(context.Background(), &translate.Request{
SourceLanguage: "en",
TargetLanguage: "he",
Text:           "Hello",
})
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
sys := mock.lastMessages[0].Content
if sys != "Custom system: translate Hello to he" {
t.Errorf("expected custom system prompt, got: %s", sys)
}
}

func TestTranslate_MultiModelRouting(t *testing.T) {
translateMock := &mockClient{response: "translated by deepseek"}
proofreadMock := &mockClient{response: "proofread by gemini"}

router := chat.Map(map[string]chat.Client{
"translate": translateMock,
"proofread": proofreadMock,
}, translateMock)

task := translate.New(router, translate.WithAutoProofread(true))

result, err := task.Translate(context.Background(), &translate.Request{
SourceLanguage: "en",
TargetLanguage: "he",
Text:           "Hello",
})
if err != nil {
t.Fatalf("unexpected error: %v", err)
}
if result.Text != "proofread by gemini" {
t.Errorf("expected proofread result from gemini, got: %q", result.Text)
}
if translateMock.lastMessages == nil {
t.Error("translate client was not called")
}
if proofreadMock.lastMessages == nil {
t.Error("proofread client was not called")
}
}

type countingClient struct {
responses []string
callCount int
}

func (c *countingClient) Chat(_ context.Context, _ *chat.Request) (*chat.Response, error) {
idx := c.callCount
c.callCount++
if idx < len(c.responses) {
return &chat.Response{Content: c.responses[idx]}, nil
}
return &chat.Response{Content: ""}, nil
}
