package entity

import "mayfly-go/pkg/model"

// 消息类型
const (
	MsgTypeUser                     = "user"                       // 用户消息
	MsgTypeAssistant                = "assistant"                  // AI 回复（普通回复）
	MsgTypeToolCall                 = "tool_call"                  // 工具调用（assistant 带 tool_calls）
	MsgTypeToolResult               = "tool_result"                // 工具结果（role = tool）
	MsgTypeInternal                 = "internal"                   // 内部消息
	MsgTypeInterrupt                = "interrupt"                  // 中断消息（role=internal且extra.type=interrupt）
	MsgTypeInterruptApproval        = "interrupt_approval"         // 中断-审批
	MsgTypeInterruptParamCompletion = "interrupt_param_completion" // 中断-参数补全
)

type SessionMessage struct {
	model.CreateModel
	model.ExtraData

	SessionKey string `gorm:"column:session_key;size:64;not null;comment:会话唯一标识" json:"sessionKey"`
	TurnId     string `gorm:"column:turn_id;size:64;not null;comment:消息唯一标识" json:"turnId"`
	Role       string `gorm:"column:role;size:10;not null;comment:消息角色" json:"role"`
	MsgType    string `gorm:"column:msg_type;size:20;not null;comment:消息类型" json:"msgType"`
	Content    string `gorm:"column:content;type:text;comment:消息内容" json:"content"`
	ToolCalls  string `gorm:"column:tool_calls;type:text;comment:工具调用" json:"toolCalls"`
	// 若role = tool，表示为toolCallId，若role = internal且为中断类型，则表示为中断id等
	ActionId   string `gorm:"column:action_id;size:64;comment:动作id" json:"actionId"`
	ToolCallId string `gorm:"column:tool_call_id;size:64;comment:工具调用id" json:"toolCallId"`
}

func (s *SessionMessage) TableName() string {
	return "t_ai_session_message"
}
