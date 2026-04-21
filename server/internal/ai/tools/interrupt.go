package tools

import "github.com/cloudwego/eino/schema"

// InterruptType 定义中断类型
type InterruptType string

const (
	TypeApproval InterruptType = "APPROVAL" // 人工审批
)

type ToolInfo struct {
	Name       string `json:"name"`
	Desc       string `json:"desc"`
	JsonSchema string `json:"jsonSchema"`
}

// InterruptMetadata 定义中断元数据接口
// 任何需要触发人机交互中断的信息都应实现此接口
type InterruptMetadata interface {
	// GetType 返回中断类型，前端根据此类型渲染不同组件
	GetType() InterruptType

	// GetTitle 返回标题
	GetTitle() string

	// GetDescription 返回详细描述
	GetDescription() string

	// GetPayload 返回中断业务负载数据
	GetPayload() any

	// GetToolInfo 获取工具信息
	GetToolInfo() *ToolInfo

	// GetToolCallId 获取工具调用ID
	GetToolCallId() string

	// GetArgument 获取工具参数
	GetArgument() string
}

// BaseInterruptInfo 基础中断信息结构体，实现 InterruptMetadata 接口
// 具体中断类型应嵌入此结构体
type BaseInterruptInfo struct {
	Type        InterruptType `json:"type"`              // 中断类型
	Title       string        `json:"title"`             // 中断标题
	Description string        `json:"description"`       // 中断描述
	Payload     any           `json:"payload,omitempty"` // 中断负载数据
	ToolCallId  string        `json:"toolCallId"`
	ToolInfo    *ToolInfo     `json:"toolInfo"`
	Arguments   string        `json:"arguments"` //原始参数字符串
}

var _ InterruptMetadata = (*BaseInterruptInfo)(nil)

// --- 实现 InterruptMetadata 接口 ---

func (b *BaseInterruptInfo) GetType() InterruptType {
	return b.Type
}

func (b *BaseInterruptInfo) GetTitle() string {
	return b.Title
}

func (b *BaseInterruptInfo) GetDescription() string {
	return b.Description
}

func (b *BaseInterruptInfo) GetPayload() any {
	return b.Payload
}

func (b *BaseInterruptInfo) GetToolInfo() *ToolInfo {
	return b.ToolInfo
}

func (b *BaseInterruptInfo) GetArgument() string {
	return b.Arguments
}

func (b *BaseInterruptInfo) GetToolCallId() string {
	return b.ToolCallId
}

type InterruptResume struct {
	TurnId      string `json:"turnId" binding:"required"`
	InterruptId string `json:"interruptId" binding:"required"` // 中断id
	Action      string `json:"action" binding:"required"`      // 操作
	Payload     any    `json:"payload"`                        // 操作参数
}

func init() {
	schema.Register[*InterruptResume]()
}
