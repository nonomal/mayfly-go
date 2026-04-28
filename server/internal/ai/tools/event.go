package tools

import "mayfly-go/pkg/eventbus"

// ResumeEventBus 中断恢复事件总线（使用 any 类型以支持不同中断类型）
var ResumeEventBus eventbus.Bus[any] = eventbus.New[any]()

const (
	// EventTopicInterruptResume 中断恢复时触发
	// 事件值类型为 *InterruptResume
	EventTopicInterruptResume = "tools:interrupt-resume"

	// EventTopicApprovalResume 审批恢复时触发
	// 事件值类型为 *ApprovalResume
	EventTopicApprovalResume = "tools:approval-resume"

	// EventTopicParamCompletionResume 参数补全恢复时触发
	// 事件值类型为 *ParamCompletionResume
	EventTopicParamCompletionResume = "tools:param-completion-resume"
)
