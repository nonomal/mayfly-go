package tools

const (
	ToolStatusSuccess = "success"
	ToolStatusError   = "error"
	// 工具执行中间状态（中断）
	ToolStatusInterrupted = "interrupted" // 已中断，等待用户交互
)
