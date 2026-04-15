package vo

import (
	"time"

	"github.com/cloudwego/eino/schema"
)

type ChatMsg struct {
	Type             string            `json:"type,omitempty"` // "text", "tool", "end"
	Time             time.Time         `json:"time"`
	MessageId        string            `json:"messageId"`
	Role             string            `json:"role"`
	Content          string            `json:"content"`
	ReasoningContent string            `json:"reasoningContent,omitempty"`
	ToolCalls        []schema.ToolCall `json:"toolCalls,omitempty"`
	ToolCallId       string            `json:"toolCallId,omitempty"`
}
