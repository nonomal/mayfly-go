package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"mayfly-go/internal/ai/imsg"
	"mayfly-go/internal/ai/session"
	"mayfly-go/pkg/i18n"
	"mayfly-go/pkg/utils/jsonx"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/compose"
)

// CompletionParamInfo 参数补全信息
type CompletionParamInfo struct {
	Param     string `json:"param"`     // 参数名
	Name      string `json:"name"`      // 参数描述
	Cacheable bool   `json:"cacheable"` // 是否缓存本次补全的参数值
}

// ParamCompletionInterruptInfo 参数完善中断信息
type ParamCompletionInterruptInfo struct {
	BaseInterruptInfo
	// 参数类型，如"db"、"machine"、"table"等
	ParamType     string                `json:"paramType"`
	MissingParams []CompletionParamInfo `json:"missingParams"` // 缺失参数列表
}

// InterruptOrResumeParamCompletion 中断或恢复参数完善
func InterruptOrResumeParamCompletion(ctx context.Context, toolDesc string, args any, reason string, paramType string, missingParams []CompletionParamInfo) error {
	isResume, err := ResumeParamCompletion(ctx, args)
	if !isResume {
		return InterruptParamCompletion(ctx, toolDesc, args, reason, paramType, missingParams)
	}
	if err == nil {
		return nil
	}
	return NewToolError(err, RecoverRetry)
}

// InterruptParamCompletion 中断参数完善
func InterruptParamCompletion(ctx context.Context, toolDesc string, args any, reason string, paramType string, missingParams []CompletionParamInfo) error {
	argsInJSON := jsonx.ToStr(args)
	// 创建中断信息（包含完整的MissingParams）
	interruptInfo := &ParamCompletionInterruptInfo{
		BaseInterruptInfo: BaseInterruptInfo{
			Type:        InterruptTypeParamCompletion,
			Title:       i18n.T(imsg.InfoIncomplete),
			Description: reason,
			Payload:     missingParams,
			ToolCallId:  compose.GetToolCallID(ctx),
			ToolInfo:    &ToolInfo{Name: toolDesc},
			Arguments:   argsInJSON,
		},
		ParamType:     paramType,
		MissingParams: missingParams,
	}

	return tool.StatefulInterrupt(ctx, interruptInfo, argsInJSON)
}

// ResumeParamCompletion 恢复参数完善
func ResumeParamCompletion(ctx context.Context, args any) (bool, error) {
	// 首先检查是否有参数补全过的中断消息，并从中提取参数值进行恢复
	messages, _ := session.DefaultSessionStore.GetMessage(ctx, &session.MessageQuery{MessageType: string(InterruptTypeParamCompletion), ToolCallId: compose.GetToolCallID(ctx)})
	if len(messages) > 0 {
		for _, msg := range messages {
			var resumeInfo ParamCompletionResume
			if err := msg.Extra.Unmarshal("resumeInfo", &resumeInfo); err != nil {
				continue
			}
			return true, handleParamCompletion(ctx, &resumeInfo, args)
		}
	}

	// 检查是否是从中断恢复
	wasInterrupted, _, _ := tool.GetInterruptState[string](ctx)
	if !wasInterrupted {
		return false, nil
	}

	// 直接使用 GetResumeContext 检查参数补全恢复
	isTarget, hasData, data := tool.GetResumeContext[*ParamCompletionResume](ctx)
	if !isTarget || !hasData {
		// 不是参数补全目标，继续执行
		return false, nil
	}

	if err := handleParamCompletion(ctx, data, args); err != nil {
		return false, err
	}

	// 修改参数调用的消息体，更新参数后
	msg := AppendResumeInfo(ctx, data.InterruptId, data)
	if msg == nil {
		return true, nil
	}

	// 对同一 TurnId 的 RMW 操作加锁，防止并发覆盖
	session.WithTurnLock(ctx, func() {
		toolCallMsgs, err := session.DefaultSessionStore.GetMessage(ctx, &session.MessageQuery{TurnId: data.TurnId, MessageType: "tool_call"})
		if err != nil || len(toolCallMsgs) == 0 {
			return
		}

		for _, toolCallMsg := range toolCallMsgs {
			for i := range toolCallMsg.ToolCalls {
				if toolCallMsg.ToolCalls[i].ID == msg.ToolCallId {
					toolCallMsg.ToolCalls[i].Function.Arguments = jsonx.ToStr(args)
					session.DefaultSessionStore.UpdateMessage(ctx, toolCallMsg)
					break
				}
			}
		}
	})

	return true, nil
}

func handleParamCompletion(ctx context.Context, data *ParamCompletionResume, args any) error {
	if data.Action != "complete" {
		return fmt.Errorf("param completion cancelled")
	}

	// 从 Payload 中获取 params 和 caches
	payload := data.Payload
	paramValues, ok := payload["params"].(map[string]any)
	if !ok {
		return fmt.Errorf("missing params in payload")
	}

	if err := json.Unmarshal([]byte(jsonx.ToStr(paramValues)), args); err != nil {
		return fmt.Errorf("unmarshal stored args failed: %w", err)
	}

	return nil
}
