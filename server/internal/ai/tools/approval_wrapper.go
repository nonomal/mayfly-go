package tools

import (
	"context"
	"fmt"
	"mayfly-go/pkg/utils/jsonx"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

type ApprovalInfo struct {
	BaseInterruptInfo
}

var _ InterruptMetadata = (*ApprovalInfo)(nil)

func NewArrpovalInfo(ctx context.Context, toolInfo *schema.ToolInfo, arguments string) *ApprovalInfo {
	ti := &ToolInfo{
		Name: toolInfo.Name,
		Desc: toolInfo.Desc,
	}
	toolJsonSchema, err := toolInfo.ParamsOneOf.ToJSONSchema()
	if err != nil {
		ti.JsonSchema = jsonx.ToStr(toolJsonSchema)
	}

	ai := &ApprovalInfo{
		BaseInterruptInfo: BaseInterruptInfo{
			Type:        TypeApproval,
			ToolInfo:    ti,
			ToolCallId:  compose.GetToolCallID(ctx),
			Arguments:   arguments,
			Description: "该操作需要审批后才能执行",
			Title:       "高危操作审批",
		}}
	return ai
}

func init() {
	schema.Register[*ApprovalInfo]()
}

// InvokableApprovableTool 是一个包装器工具，用于将普通的 InvokableTool 转换为需要审批的工具。
type InvokableApprovableTool struct {
	tool.InvokableTool
}

func (i InvokableApprovableTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return i.InvokableTool.Info(ctx)
}

func (i InvokableApprovableTool) InvokableRun(ctx context.Context, argumentsInJSON string,
	opts ...tool.Option,
) (string, error) {
	toolInfo, err := i.Info(ctx)
	if err != nil {
		return "", err
	}

	wasInterrupted, _, storedArguments := tool.GetInterruptState[string](ctx)
	if !wasInterrupted {
		return "", tool.StatefulInterrupt(ctx, NewArrpovalInfo(ctx, toolInfo, argumentsInJSON), argumentsInJSON)
	}

	isResumeTarget, hasData, data := tool.GetResumeContext[*InterruptResume](ctx)
	if isResumeTarget && hasData {
		if data.Action == "approve" {
			return i.InvokableTool.InvokableRun(ctx, storedArguments, opts...)
		}

		if data.Action == "reject" {
			reason := "用户未提供具体原因"
			if data.Payload != nil {
				// 尝试将 Payload 转换为字符串，如果是 map 或 struct 可以格式化得更好
				if r, ok := data.Payload.(string); ok && r != "" {
					reason = r
				} else {
					// 如果是复杂结构，序列化为 JSON 字符串以便 LLM 理解
					reason = fmt.Sprintf("用户反馈: %v", jsonx.ToStr(data.Payload))
				}
			}

			// 构建更清晰的拒绝消息
			msg := fmt.Sprintf(
				"[OPERATION_REJECTED] The tool '%s' was explicitly rejected by the user.\nReason: %s\nPlease do not retry this action automatically. Ask the user for further instructions if needed.",
				toolInfo.Name,
				reason,
			)
			return msg, nil
		}

		return fmt.Sprintf("[OPERATION_CANCELLED] The tool '%s' execution was cancelled due to invalid action: %s", toolInfo.Name, data.Action), nil
	}

	isResumeTarget, _, _ = tool.GetResumeContext[any](ctx)
	if !isResumeTarget {
		return "", tool.StatefulInterrupt(ctx, NewArrpovalInfo(ctx, toolInfo, storedArguments), storedArguments)
	}

	return i.InvokableTool.InvokableRun(ctx, storedArguments, opts...)
}
