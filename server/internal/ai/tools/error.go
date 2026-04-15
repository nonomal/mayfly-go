package tools

import (
	"errors"
	"fmt"
)

var (
	ErrToolFailed = errors.New("tool failed")
)

type RecoverStrategy string

const (
	RecoverNone  RecoverStrategy = "none"
	RecoverRetry RecoverStrategy = "retry" // 重试，将错误消息返回给llm
)

type ToolError struct {
	Msg      string          // 错误信息
	Strategy RecoverStrategy // 恢复策略

	ToolName   string // 工具名
	ToolCallId string // tool call id

	err error // 原始错误
}

func (e *ToolError) Error() string {
	return e.Msg
}

func (e *ToolError) Unwrap() error {
	return e.err
}

func (e *ToolError) WithToolName(toolName string) *ToolError {
	e.ToolName = toolName
	return e
}

func (e *ToolError) WithToolCallId(toolCallId string) *ToolError {
	e.ToolCallId = toolCallId
	return e
}

func NewToolError(err error, strategy RecoverStrategy) *ToolError {
	if err == nil {
		err = ErrToolFailed
	}

	return &ToolError{
		Msg:      err.Error(),
		Strategy: strategy,
		err:      errors.Join(err, ErrToolFailed), // 带上ErrToolFailed，方便errors.Is判断
	}
}

// GetToolErrorMsg 获取工具错误信息
func GetToolErrorMsg(err error) string {
	return fmt.Sprintf("[tool error] %v", err)
}
