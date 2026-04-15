package tools

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type ApprovalInfo struct {
	ToolName        string
	ArgumentsInJSON string
}

type ApprovalResult struct {
	Approved         bool
	DisapproveReason *string
}

func (ai *ApprovalInfo) String() string {
	return fmt.Sprintf("tool '%s' interrupted with arguments '%s', waiting for your approval, "+
		"please answer with Y/N",
		ai.ToolName, ai.ArgumentsInJSON)
}

func init() {
	schema.Register[*ApprovalInfo]()
}

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
		return "", tool.StatefulInterrupt(ctx, &ApprovalInfo{
			ToolName:        toolInfo.Name,
			ArgumentsInJSON: argumentsInJSON,
		}, argumentsInJSON)
	}

	isResumeTarget, hasData, data := tool.GetResumeContext[*ApprovalResult](ctx)
	if isResumeTarget && hasData {
		if data.Approved {
			return i.InvokableTool.InvokableRun(ctx, storedArguments, opts...)
		}

		if data.DisapproveReason != nil {
			return fmt.Sprintf("tool '%s' disapproved, reason: %s", toolInfo.Name, *data.DisapproveReason), nil
		}

		return fmt.Sprintf("tool '%s' disapproved", toolInfo.Name), nil
	}

	isResumeTarget, _, _ = tool.GetResumeContext[any](ctx)
	if !isResumeTarget {
		return "", tool.StatefulInterrupt(ctx, &ApprovalInfo{
			ToolName:        toolInfo.Name,
			ArgumentsInJSON: storedArguments,
		}, storedArguments)
	}

	return i.InvokableTool.InvokableRun(ctx, storedArguments, opts...)
}
