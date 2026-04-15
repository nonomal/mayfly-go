package contextx

import (
	"context"
	"mayfly-go/pkg/model"
	"mayfly-go/pkg/utils/stringx"
)

type CtxKey string

const (
	LoginAccountKey CtxKey = "loginAccount"
	TraceIdKey      CtxKey = "traceId"
)

func NewLoginAccount(la *model.LoginAccount) context.Context {
	return WithLoginAccount(context.Background(), la)
}

func WithLoginAccount(ctx context.Context, la *model.LoginAccount) context.Context {
	return context.WithValue(ctx, LoginAccountKey, la)
}

// GetLoginAccount 从context中获取登录账号信息，不存在返回nil
func GetLoginAccount(ctx context.Context) *model.LoginAccount {
	if la, ok := ctx.Value(LoginAccountKey).(*model.LoginAccount); ok {
		return la
	}
	return nil
}

/**   traceId   **/

// WithTraceId 生成traceId并放置于context中, 如果已存在则不覆盖
func WithTraceId(ctx context.Context) context.Context {
	if GetTraceId(ctx) != "" {
		return ctx
	}
	return context.WithValue(ctx, TraceIdKey, stringx.RandByChars(16, stringx.Nums+stringx.LowerChars))
}

// 从context中获取traceId
func GetTraceId(ctx context.Context) string {
	if val, ok := ctx.Value(TraceIdKey).(string); ok {
		return val
	}
	return ""
}
