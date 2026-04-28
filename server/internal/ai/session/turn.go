package session

import (
	"context"
	"mayfly-go/pkg/contextx"
	"mayfly-go/pkg/utils/collx"
	"sync"
)

// Turn 表示一次 Agent 调用中的上下文
// 用于在一次 Agent 调用过程中存储临时数据，在调用结束时自动清理
type Turn struct {
	TurnId string
	Values collx.SM[string, any] // 存储任意键值对，供中间件和工具使用
	mu     sync.Mutex            // 用于串行化该 Turn 下的消息更新操作
}

// NewTurn 创建一个新的 Turn
func NewTurn() *Turn {
	return &Turn{}
}

// Set 设置值到缓存
func (t *Turn) Set(key string, value any) {
	t.Values.Store(key, value)
}

// Get 从缓存获取值
func (t *Turn) Get(key string) (any, bool) {
	return t.Values.Load(key)
}

// context key 类型
const TurnCtxKey contextx.CtxKey = "turn"

// WithTurn 将 Turn 设置到 context 中
func WithTurn(ctx context.Context, turnId string) context.Context {
	return context.WithValue(ctx, TurnCtxKey, &Turn{
		TurnId: turnId,
	})
}

// GetTurn 从 context 中获取 Turn
func GetTurn(ctx context.Context) *Turn {
	if turn, ok := ctx.Value(TurnCtxKey).(*Turn); ok {
		return turn
	}
	return nil
}

// WithTurnLock 对当前 Context 中 Turn 的消息更新操作加锁执行。
// 利用 Turn 的生命周期（一次 Agent Run）自动管理锁，无需全局 Map。
func WithTurnLock(ctx context.Context, fn func()) {
	turn := GetTurn(ctx)
	if turn == nil {
		fn()
		return
	}
	turn.mu.Lock()
	defer turn.mu.Unlock()
	fn()
}
