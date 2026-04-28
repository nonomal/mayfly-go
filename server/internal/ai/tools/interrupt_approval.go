package tools

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"mayfly-go/internal/ai/imsg"
	"mayfly-go/internal/ai/session"
	"mayfly-go/pkg/i18n"
	"mayfly-go/pkg/utils/jsonx"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

func InterruptOrResumeApproval(ctx context.Context, toolDesc string, arguments any, reason string) error {
	isApprovalResume, err := ResumeApproval(ctx, toolDesc)
	if !isApprovalResume {
		return InterruptApproval(ctx, toolDesc, arguments, reason)
	}
	if err == nil {
		return nil
	}
	return NewToolError(err, RecoverRetry)
}

func InterruptApproval(ctx context.Context, toolDesc string, arguments any, reason string) error {
	argumentsInJSON := jsonx.ToStr(arguments)
	ai := &ApprovalInfo{
		BaseInterruptInfo: BaseInterruptInfo{
			Type:        InterruptTypeApproval,
			ToolCallId:  compose.GetToolCallID(ctx),
			Arguments:   argumentsInJSON,
			Description: reason,
			ToolInfo:    &ToolInfo{Name: toolDesc},
			Title:       i18n.T(imsg.ApprovalTitle),
		}}
	return tool.StatefulInterrupt(ctx, ai, argumentsInJSON)
}

func ResumeApproval(ctx context.Context, toolDesc string) (bool, error) {
	// 首先检查是否已审批过
	messages, _ := session.DefaultSessionStore.GetMessage(ctx, &session.MessageQuery{MessageType: string(InterruptTypeApproval), ToolCallId: compose.GetToolCallID(ctx)})
	if len(messages) > 0 {
		for _, msg := range messages {
			var resumeInfo ApprovalResume
			if err := msg.Extra.Unmarshal("resumeInfo", &resumeInfo); err != nil {
				continue
			}
			return true, handleApprovalResult(ctx, toolDesc, &resumeInfo)
		}
	}

	wasInterrupted, _, _ := tool.GetInterruptState[string](ctx)
	if !wasInterrupted {
		return false, nil
	}

	isResumeTarget, hasData, data := tool.GetResumeContext[*ApprovalResume](ctx)
	if !isResumeTarget || !hasData {
		return false, nil
	}

	AppendResumeInfo(ctx, data.InterruptId, data)
	return true, handleApprovalResult(ctx, toolDesc, data)
}

func handleApprovalResult(ctx context.Context, toolDesc string, data *ApprovalResume) error {
	if data.Action == "approve" {
		return nil
	}

	if data.Action == "reject" {
		reason := cmp.Or(data.Payload.GetStr("reason"), i18n.T(imsg.RejectReasonDefault))
		// 构建更清晰的拒绝消息
		msg := fmt.Sprintf(
			"[OPERATION_REJECTED] The tool '%s' was explicitly rejected by the user.\nReason: %s\nPlease do not retry this action automatically. Ask the user for further instructions if needed.",
			toolDesc,
			reason,
		)
		return errors.New(msg)
	}

	return fmt.Errorf("[OPERATION_CANCELLED] The tool '%s' execution was cancelled due to invalid action: %s", toolDesc, data.Action)
}
