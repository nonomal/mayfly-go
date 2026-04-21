package vo

import (
	"mayfly-go/pkg/utils/collx"
	"time"

	"github.com/cloudwego/eino/schema"
)

type ChatMsg struct {
	Type             string            `json:"type,omitempty"` // "text", "tool", "end"
	Time             time.Time         `json:"time"`
	TurnId           string            `json:"turnId"`
	Role             string            `json:"role"`
	Content          string            `json:"content"`
	ReasoningContent string            `json:"reasoningContent,omitempty"`
	ToolCalls        []schema.ToolCall `json:"toolCalls,omitempty"`
	ActionId         string            `json:"actionId,omitempty"`
	Extra            collx.M           `json:"extra,omitempty"`
}
