package agent

import (
	"context"
	"errors"
	"io"
	"mayfly-go/internal/ai/agent/middleware"
	"mayfly-go/internal/ai/session"
	"mayfly-go/internal/ai/tools"
	"mayfly-go/pkg/contextx"
	"mayfly-go/pkg/logx"
	"slices"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/deep"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

// GetDefaultAgent 获取默认agent
func GetDefaultAgent(ctx context.Context, opts ...option) (*Agent, error) {
	return NewDeepAgent(ctx, opts...)
}

const (
	DefaultAgentId = "main"
)

func NewAgent(ctx context.Context, opts ...option) (*Agent, error) {
	return newAgent(ctx, func(ctx context.Context, a *Agent) (adk.Agent, error) {
		return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
			Name:          a.name,
			Description:   a.description,
			Model:         a.chatModel,
			MaxIterations: a.maxStep,
			ToolsConfig: adk.ToolsConfig{
				ToolsNodeConfig: compose.ToolsNodeConfig{
					Tools: a.tools.GetAll(),
				},
			},
			Handlers: a.middlewares,
		})
	}, opts...)
}

func NewDeepAgent(ctx context.Context, opts ...option) (*Agent, error) {
	return newAgent(ctx, func(ctx context.Context, cfg *Agent) (adk.Agent, error) {
		return deep.New(ctx, &deep.Config{
			Name:        cfg.name,
			Description: cfg.description,
			ChatModel:   cfg.chatModel,
			ToolsConfig: adk.ToolsConfig{
				ToolsNodeConfig: compose.ToolsNodeConfig{
					Tools: cfg.tools.GetAll(),
				},
			},
			MaxIteration: cfg.maxStep,
			Handlers:     cfg.middlewares,
		})
	}, opts...)
}

func newAgent(ctx context.Context, factory agentFactory, opts ...option) (*Agent, error) {
	agent := &Agent{
		id:          DefaultAgentId,
		name:        "OpsExpert",
		description: "an agent for general task",
		maxStep:     20,
		tools:       tools.DefaultRegistry,
		middlewares: []adk.ChatModelAgentMiddleware{
			&middleware.SafeToolMiddleware{},
		},
	}

	for _, opt := range opts {
		opt(agent)
	}

	if agent.chatModel == nil {
		chatModel, err := GetChatModel(ctx)
		if err != nil {
			return nil, err
		}
		agent.chatModel = chatModel
	}

	if agent.contextManager == nil {
		if ctxManager, err := GetDefaultContextManager(); err != nil {
			return nil, err
		} else {
			agent.contextManager = ctxManager
		}
	}

	adkAgent, err := factory(ctx, agent)
	if err != nil {
		return nil, err
	}
	agent.agent = adkAgent
	return agent, nil
}

// agentFactory 定义创建 adk.Agent 的回调函数签名
type agentFactory func(ctx context.Context, cfg *Agent) (adk.Agent, error)

type Agent struct {
	agent     adk.Agent
	chatModel model.ToolCallingChatModel // agent使用的chat model

	id          string
	name        string // agent名称
	description string // agent描述
	maxStep     int    // agent最大执行步数，防止死循环

	tools          *tools.Registry                // 可调用的工具注册中心
	middlewares    []adk.ChatModelAgentMiddleware // 中间件
	contextManager *ContextManager                // 上下文管理器
}

// Run 运行agent
func (a *Agent) Run(ctx context.Context, messages []adk.Message, runOpts ...RunOption) (string, error) {
	ctx = contextx.WithTraceId(ctx)

	runOptions := newRunOptions(ctx, runOpts...)
	if runOptions.sessionKey != "" {
		ctx = session.WithSessionKey(ctx, runOptions.sessionKey)
		ctx = session.WithTurn(ctx, runOptions.turnId)
	}
	if len(messages) > 0 {
		for _, inputMsg := range messages {
			SetTurnId(inputMsg, runOptions.turnId)
		}
	}

	checkPointStore := GetDefaultCheckPointStore()
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		EnableStreaming: true,
		Agent:           a.agent,
		CheckPointStore: checkPointStore,
	})

	adkRunOptions := runOptions.adkRunOptions
	adkRunOptions = append(adkRunOptions,
		adk.WithCallbacks(logCallback),
		adk.WithCheckPointID(runOptions.turnId))

	var outputMessages []adk.Message
	var events *adk.AsyncIterator[*adk.AgentEvent]
	var err error

	// 保证消息持久化：无论正常返回、错误返回还是 panic，都会尝试保存
	// 避免崩溃或提前 return 导致整轮消息丢失
	defer func() {
		saveMsgs := slices.Concat(messages, outputMessages)
		if len(saveMsgs) > 0 {
			if err := a.contextManager.AppendMsgs(ctx, saveMsgs...); err != nil {
				logx.ErrorfContext(ctx, "agent append message error: %v", err)
			}
		}
	}()

	if runOptions.resumeParams != nil {
		resumePrams := runOptions.resumeParams
		targets := map[string]any{}

		var resumeMsgs []adk.Message
		for _, v := range resumePrams {
			data, ok := v.(*tools.InterruptResume)
			if !ok {
				continue
			}
			interruptId := data.InterruptId

			// key -> interruptId  value -> InterruptResume
			targets[interruptId] = data.ToTarget()

			// 中断恢复消息
			internalResumeMessage := &schema.Message{
				Role: RoleInternal,
			}
			extra := NewInternalMessageExtra(InternalMessageTypeResume, data)
			internalResumeMessage.Extra = extra
			SetTurnId(internalResumeMessage, runOptions.turnId)
			SetActionId(internalResumeMessage, interruptId)
			resumeMsgs = append(resumeMsgs, internalResumeMessage)
		}

		events, err = runner.ResumeWithParams(ctx, runOptions.turnId, &adk.ResumeParams{
			Targets: targets,
		}, adkRunOptions...)
		if err != nil {
			return "", err
		}

		// 事件处理
		for _, resumeMsg := range resumeMsgs {
			runOptions.CallOnEvent(ctx, nil, resumeMsg)
		}
		checkPointStore.Delete(ctx, runOptions.turnId)
	} else {
		contextMessages, err := a.contextManager.BuildMessages(ctx)
		if err != nil {
			logx.ErrorContext(ctx, err.Error())
			contextMessages = []adk.Message{}
		}
		events = runner.Run(ctx, slices.Concat(contextMessages, messages), adkRunOptions...)
	}

	eventOutputMessages, err := a.handleEvents(ctx, events, runOptions)
	// 先合并已收到的消息，即使后续出错也不丢失部分输出
	outputMessages = append(outputMessages, eventOutputMessages...)

	if err != nil {
		if toolErr, ok := errors.AsType[*tools.ToolError](err); ok {
			// 工具调用失败，并且没有重试，则记录对应错误消息
			toolErrMsg := &schema.Message{
				Role:       schema.Tool,
				Content:    tools.GetToolErrorMsg(err),
				ToolName:   toolErr.ToolName,
				ToolCallID: toolErr.ToolCallId,
			}
			SetTurnId(toolErrMsg, runOptions.turnId)
			SetToolStatus(toolErrMsg, tools.ToolStatusError)

			errMsg := &schema.Message{
				Role:    schema.Assistant,
				Content: err.Error(),
			}
			SetTurnId(errMsg, runOptions.turnId)
			runOptions.CallOnEvent(ctx, nil, toolErrMsg)

			outputMessages = append(outputMessages, toolErrMsg, errMsg)
		} else {
			logx.ErrorContext(ctx, err.Error())
			return "", err // defer 会保存 outputMessages 中已有的消息
		}
	}

	if len(outputMessages) > 0 {
		return outputMessages[len(outputMessages)-1].Content, err
	}

	return "finished without output message", err
}

// handleEvents 处理事件
func (a *Agent) handleEvents(ctx context.Context, events *adk.AsyncIterator[*adk.AgentEvent], runOptions *runOptions) ([]adk.Message, error) {
	var outputMessages []adk.Message
	var err error

	for {
		event, ok := events.Next()
		if !ok {
			break
		}

		err = event.Err
		if err != nil {
			break
		}

		var msg adk.Message

		sr := getMessageStream(event)
		if sr != nil {
			// 使用匿名函数或直接在处理完后关闭
			func() {
				defer sr.Close()
				var chunkMessages []adk.Message
				for {
					chunk, err := sr.Recv()
					if errors.Is(err, io.EOF) {
						break
					}
					if err != nil {
						logx.WarnfContext(ctx, "stream recv error: %v", err)
						break
					}
					chunkMessages = append(chunkMessages, chunk)
					if err := runOptions.CallOnChunk(ctx, chunk); err != nil {
						logx.WarnfContext(ctx, "onStreaming callback error: %v", err)
						break
					}
				}
				if len(chunkMessages) > 0 {
					// 拼接chunk为完整的消息
					if message, err := schema.ConcatMessages(chunkMessages); err != nil {
						logx.WarnfContext(ctx, "concat streamed messages error: %v", err)
					} else {
						msg = message
					}
				}
			}()
		} else {
			msg = getMessage(event)
		}

		if event.Action != nil && event.Action.Interrupted != nil {
			interruptInfo := event.Action.Interrupted
			if interruptMessages, err := a.handleInterrupt(interruptInfo); err != nil {
				logx.ErrorfContext(ctx, "interrupt error: %v", err)
				continue
			} else {
				outputMessages = append(outputMessages, interruptMessages...)
				for _, msg := range interruptMessages {
					SetTurnId(msg, runOptions.turnId)
					if err := runOptions.CallOnEvent(ctx, event, msg); err != nil {
						logx.WarnfContext(ctx, "onEvent callback error: %v", err)
						break
					}
				}
			}
		}

		if msg == nil {
			continue
		}
		SetTurnId(msg, runOptions.turnId)

		outputMessages = append(outputMessages, msg)
		if err := runOptions.CallOnEvent(ctx, event, msg); err != nil {
			logx.WarnfContext(ctx, "onEvent callback error: %v", err)
			return outputMessages, err
		}

		LogEventAndMsg(ctx, event, msg)
	}

	return outputMessages, err
}

// handleIntererrupt 处理中断
func (a *Agent) handleInterrupt(interruptInfo *adk.InterruptInfo) ([]adk.Message, error) {
	outputMessages := []adk.Message{}
	for _, ic := range interruptInfo.InterruptContexts {
		if !ic.IsRootCause {
			continue
		}

		info, ok := ic.Info.(tools.InterruptMetadata)
		if !ok {
			continue
		}

		interruptType := info.GetType()
		toolCallId := info.GetToolCallId()

		extra := NewInternalMessageExtra(string(interruptType), info)
		internalMsg := &schema.Message{
			Role:       RoleInternal,
			Content:    info.GetDescription(),
			ToolName:   info.GetToolInfo().Name,
			ToolCallID: toolCallId,
			Extra:      extra,
		}
		SetActionId(internalMsg, ic.ID)
		SetToolStatus(internalMsg, tools.ToolStatusInterrupted)
		outputMessages = append(outputMessages, internalMsg)
	}

	return outputMessages, nil
}

func getMessageStream(event *adk.AgentEvent) adk.MessageStream {
	eo := event.Output
	if eo == nil {
		return nil
	}
	mo := eo.MessageOutput
	if mo == nil {
		return nil
	}
	return mo.MessageStream
}

func getMessage(event *adk.AgentEvent) *schema.Message {
	eo := event.Output
	if eo == nil {
		return nil
	}
	mo := eo.MessageOutput
	if mo == nil {
		return nil
	}
	return mo.Message
}
