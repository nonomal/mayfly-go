package vo

import (
	"mayfly-go/pkg/utils/collx"
	"time"

	"github.com/cloudwego/eino/schema"
)

type ChatMsg struct {
	Type             string            `json:"type,omitempty"` // "text", "tool", "end"
	SessionId        string            `json:"sessionId"`      // 会话ID，用于前端过滤消息
	Time             time.Time         `json:"time"`
	TurnId           string            `json:"turnId"`
	Role             string            `json:"role"`
	Content          string            `json:"content"`
	ReasoningContent string            `json:"reasoningContent,omitempty"`
	ToolCalls        []schema.ToolCall `json:"toolCalls,omitempty"`
	ActionId         string            `json:"actionId,omitempty"`
	ToolCallId       string            `json:"toolCallId,omitempty"`
	Extra            collx.M           `json:"extra,omitempty"`
}
