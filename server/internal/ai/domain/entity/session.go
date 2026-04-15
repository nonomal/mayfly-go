package entity

import (
	"mayfly-go/pkg/model"
)

// Session AI 会话元数据
type Session struct {
	model.Model
	model.ExtraData

	SessionKey string `gorm:"column:session_key;size:64;not null;comment:会话唯一标识" json:"sessionKey"`
	Title      string `gorm:"column:title;size:255;;comment:会话标题" json:"title"`
	Summary    string `gorm:"column:summary;size:3000;;comment:会话摘要" json:"summary"`

	MessageCount int `gorm:"column:message_count;default:0;comment:消息数量" json:"messageCount"`
	TokenCount   int `gorm:"column:token_count;default:0;comment:消耗Token总数" json:"tokenCount"`
}

func (s *Session) TableName() string {
	return "t_ai_session"
}
