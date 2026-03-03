package llm

import (
	"context"
	"time"
)

type Message struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	Parts      []Part     `json:"parts,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
}

const (
	PartTypeText        = "text"
	PartTypeImageURL    = "image_url"
	PartTypeImageBase64 = "image_base64"
)

type Part struct {
	Type       string `json:"type"`
	Text       string `json:"text,omitempty"`
	URL        string `json:"url,omitempty"`
	DataBase64 string `json:"data_base64,omitempty"`
	MIMEType   string `json:"mime_type,omitempty"`
}

type Tool struct {
	Name           string `json:"name"`
	Description    string `json:"description,omitempty"`
	ParametersJSON string `json:"parameters_json,omitempty"`
}

type ToolCall struct {
	ID               string         `json:"id,omitempty"`
	Type             string         `json:"type,omitempty"`
	Name             string         `json:"name"`
	Arguments        map[string]any `json:"arguments,omitempty"`
	RawArguments     string         `json:"raw_arguments,omitempty"`
	ThoughtSignature string         `json:"thought_signature,omitempty"`
}

type Usage struct {
	InputTokens  int
	OutputTokens int
	TotalTokens  int
	Cost         float64 // USD
}

type StreamToolCallDelta struct {
	Index     int
	ID        string
	Name      string
	ArgsChunk string
}

type StreamEvent struct {
	Delta         string
	ToolCallDelta *StreamToolCallDelta
	Usage         *Usage
	Done          bool
}

type StreamHandler func(event StreamEvent) error

type Result struct {
	Text      string
	Parts     []Part
	JSON      any
	ToolCalls []ToolCall
	Usage     Usage
	Duration  time.Duration
}

type Request struct {
	Model      string
	Messages   []Message
	Tools      []Tool
	ForceJSON  bool
	Parameters map[string]any
	OnStream   StreamHandler
}

type Client interface {
	Chat(ctx context.Context, req Request) (Result, error)
}
