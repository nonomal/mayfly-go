package entity

import "mayfly-go/pkg/model"

type SessionMessage struct {
	model.CreateModel
	model.ExtraData

	SessionKey string `gorm:"column:session_key;size:64;not null;comment:会话唯一标识" json:"sessionKey"`
	TurnId     string `gorm:"column:turn_id;size:64;not null;comment:消息唯一标识" json:"turnId"`
	Role       string `gorm:"column:role;size:10;not null;comment:消息角色" json:"role"`
	Content    string `gorm:"column:content;type:text;comment:消息内容" json:"content"`
	ToolCalls  string `gorm:"column:tool_calls;type:text;comment:工具调用" json:"toolCalls"`
	// 若role = tool，表示为toolCallId，若role = internal且为中断类型，则表示为中断id等
	ActionId string `gorm:"column:action_id;size:64;comment:动作id" json:"actionId"`
}

func (s *SessionMessage) TableName() string {
	return "t_ai_session_message"
}
