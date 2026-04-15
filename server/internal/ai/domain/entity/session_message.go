package entity

import "mayfly-go/pkg/model"

type SessionMessage struct {
	model.CreateModel
	model.ExtraData

	SessionKey string `gorm:"column:session_key;size:64;not null;comment:会话唯一标识" json:"sessionKey"`
	MessageId  string `gorm:"column:message_id;size:64;not null;comment:消息唯一标识" json:"messageId"`
	Role       string `gorm:"column:role;size:10;not null;comment:消息角色" json:"role"`
	Content    string `gorm:"column:content;type:text;comment:消息内容" json:"content"`
	ToolCalls  string `gorm:"column:tool_calls;type:text;comment:工具调用" json:"toolCalls"`
	ToolCallId string `gorm:"column:tool_call_id;size:64;comment:工具调用id" json:"toolCallId"`
}

func (s *SessionMessage) TableName() string {
	return "t_ai_session_message"
}
