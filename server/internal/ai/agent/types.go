package agent

import (
	"mayfly-go/pkg/utils/collx"

	"github.com/cloudwego/eino/schema"
)

const (
	RoleInternal = schema.RoleType("internal") // 内部系统
)

type InternalMessageType string

const (
	InternalMessageTypeInterrupt InternalMessageType = "interrupt" // 中断类型
	InternalMessageTypeResume    InternalMessageType = "resume"    // 中断恢复
)

// InternalMessageExtra 内部消息内容
type InternalMessageExtra struct {
	Type    InternalMessageType `json:"type"`
	Content any                 `json:"content"`
}

func NewInternalMessageExtra(t InternalMessageType, content any) collx.M {
	return collx.Kvs("type", t, "content", content)
}
