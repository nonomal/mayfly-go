package session

import "context"

var (
	// DefaultSessionStore 默认会话存储
	DefaultSessionStore Store
)

type MessageQuery struct {
	ActionId    string
	TurnId      string
	ToolCallId  string
	MessageType string
}

// Store 会话相关信息存储
type Store interface {
	// ============= 消息操作  =============

	// AppendMsgs 追加消息到会话末尾
	AppendMsgs(ctx context.Context, sessionKey string, msgs ...*Message) error
	// GetHistory 获取会话历史消息（按时间正序排列）
	GetHistory(ctx context.Context, sessionKey string, limit int) ([]*Message, error)
	// ClearHistory 清空会话历史消息
	ClearHistory(ctx context.Context, sessionKey string) error

	// GetMessage 根据查询条件获取单条消息
	GetMessage(ctx context.Context, query *MessageQuery) ([]*Message, error)
	// UpdateMessage 更新单条消息
	UpdateMessage(ctx context.Context, msg *Message) error

	// ============= 元数据操作  =============

	// ListMetas 列出所有会话元信息
	ListMetas(ctx context.Context) ([]*SessionMeta, error)
	// GetMeta 获取会话元信息
	GetMeta(ctx context.Context, sessionKey string) (*SessionMeta, error)
	// SaveMeta 保存会话元信息
	SaveMeta(ctx context.Context, meta *SessionMeta) error
	// DeleteMeta 删除会话元信息
	DeleteMeta(ctx context.Context, sessionKey string) error
}
