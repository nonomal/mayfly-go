package memory

import (
	"context"
	"time"
)

// MemoryItem 用户记忆结构
// 每个用户一条记录，使用Markdown格式存储所有记忆内容
type MemoryItem struct {
	// 用户ID（主键）
	UserID string `json:"userId"`
	// 记忆内容（Markdown格式）
	Memory string `json:"memory"`
	// 创建时间
	CreatedAt time.Time `json:"createdAt"`
	// 最后更新时间
	UpdatedAt time.Time `json:"updatedAt"`
}

// Store 记忆存储接口
type Store interface {
	// 短期记忆
	GetShortTerm(ctx context.Context, userId string) ([]*MemoryItem, error)
	SaveShortTerm(ctx context.Context, userId string, items []*MemoryItem) error

	// 长期记忆
	GetLongTerm(ctx context.Context, userId string) ([]*MemoryItem, error)
	SaveLongTerm(ctx context.Context, userId string, items []*MemoryItem) error

	// 向量检索
	Search(ctx context.Context, query string, limit int) ([]*MemoryItem, error)
}
