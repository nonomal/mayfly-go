package session

import (
	"context"
	"fmt"
	"mayfly-go/pkg/utils/collx"
	"mayfly-go/pkg/utils/stringx"
	"time"

	"github.com/cloudwego/eino/adk"
)

// Manager 会话管理器
// 负责会话缓存管理和生命周期管理，底层存储委托给 Store 实现
type Manager struct {
	store Store // 底层存储
}

// NewManager 创建会话管理器
// store: 底层存储实现 (如 JSONLStore, MemoryStore 等)
func NewManager(store Store) *Manager {
	return &Manager{
		store: store,
	}
}

// GetHistory 获取会话历史消息
func (m *Manager) GetHistory(ctx context.Context, key string, opts ...GetOption) ([]adk.Message, error) {
	// 应用选项配置
	options := defaultGetOptions()
	for _, opt := range opts {
		opt(options)
	}
	return m.store.GetHistory(ctx, key, options.messageLimit)
}

// AppendMsgs 追加消息到会话
func (m *Manager) AppendMsgs(ctx context.Context, key string, msgs ...adk.Message) error {
	if key == "" || len(msgs) == 0 {
		return nil
	}

	// 追加消息到底层存储（Store 只负责存储，不更新元数据）
	if err := m.store.AppendMsgs(ctx, key, msgs...); err != nil {
		return err
	}

	meta, err := m.store.GetMeta(ctx, key)
	if err != nil {
		return err
	}
	// 如果元数据不存在，创建新会话
	if meta == nil {
		// 元数据不存在，创建新的
		meta = &SessionMeta{
			Key:       key,
			CreatedAt: time.Now(),
		}
		meta.Extra.Set("title", stringx.Truncate(msgs[0].Content, 50, 30, "..."))
	}

	// 计算新增消息的Token数量
	totalTokens := collx.ArrayReduce(msgs, 0, func(totalToken int, msg adk.Message) int {
		responseMeta := msg.ResponseMeta
		if responseMeta != nil && responseMeta.Usage != nil {
			return totalToken + responseMeta.Usage.TotalTokens
		}
		return totalToken
	})

	// 保存元数据
	meta.Count += len(msgs)
	meta.TokenCount += totalTokens
	meta.UpdatedAt = time.Now()
	return m.store.SaveMeta(ctx, meta)
}

// Delete 删除会话
// 同时删除历史消息、元数据和缓存
func (m *Manager) Delete(ctx context.Context, key string) error {
	// 先清空历史消息（可选，确保数据一致性）
	if err := m.store.ClearHistory(ctx, key); err != nil {
		return fmt.Errorf("manager: clear history: %w", err)
	}

	// 再删除元数据
	return m.store.DeleteMeta(ctx, key)
}

// List 列出所有会话
// 从 Store 加载最新的会话元数据列表
func (m *Manager) List(ctx context.Context) ([]*SessionMeta, error) {
	return m.store.ListMetas(ctx)
}

// ClearHistory 清空会话历史消息（保留元数据）
func (m *Manager) ClearHistory(ctx context.Context, key string) error {
	// 调用 Store 清空历史
	return m.store.ClearHistory(ctx, key)
}
