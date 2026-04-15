package agent

import (
	"context"
	"errors"
	"mayfly-go/internal/ai/memory"
	"mayfly-go/internal/ai/session"

	"github.com/cloudwego/eino/adk"
)

var (
	// DefaultSessionStore 默认会话存储
	DefaultSessionStore session.Store
)

// GetDefaultContextManager 获取默认的上下文管理器实例
func GetDefaultContextManager() (*ContextManager, error) {
	if DefaultSessionStore != nil {
		return NewContextManager(session.NewManager(DefaultSessionStore)), nil
	}

	sessionStore, err := session.NewStoreJSONL("./sessions")
	if err != nil {
		return nil, err
	}
	DefaultSessionStore = sessionStore
	return NewContextManager(session.NewManager(sessionStore)), nil
}

type ContextManager struct {
	sessionManager *session.Manager // 会话管理器
	memoryManager  *memory.Manager  // 记忆管理器
}

// NewContextManager 创建并初始化 ContextManager 实例
func NewContextManager(sm *session.Manager) *ContextManager {
	return &ContextManager{
		sessionManager: sm,
	}
}

func (c *ContextManager) GetSessionKey(ctx context.Context) string {
	return session.GetSessionKey(ctx)
}

// BuildMessages 从上下文中构建消息列表，供Agent执行使用
func (c *ContextManager) BuildMessages(ctx context.Context) ([]adk.Message, error) {
	sessionKey := c.GetSessionKey(ctx)
	if sessionKey == "" {
		return nil, errors.New("session key is empty")
	}
	history, err := c.sessionManager.GetHistory(ctx, sessionKey)
	if err != nil {
		return nil, err
	}
	return history, nil
}

func (c *ContextManager) AppendMsgs(ctx context.Context, msgs ...adk.Message) error {
	return c.sessionManager.AppendMsgs(ctx, c.GetSessionKey(ctx), msgs...)
}

func (c *ContextManager) ClearHistory(ctx context.Context) error {
	return c.sessionManager.ClearHistory(ctx, c.GetSessionKey(ctx))
}
