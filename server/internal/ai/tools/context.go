package tools

import (
	"context"
	"mayfly-go/pkg/contextx"

	"github.com/spf13/cast"
)

const (
	SessionKey contextx.CtxKey = "sessionKey"
)

// GetSessionKey 从上下文中获取会话Key
func GetSessionKey(ctx context.Context) string {
	return cast.ToString(ctx.Value(SessionKey))
}

// WithSessionKey 设置会话Key到上下文中
func WithSessionKey(ctx context.Context, sessionKey string) context.Context {
	return context.WithValue(ctx, SessionKey, sessionKey)
}
