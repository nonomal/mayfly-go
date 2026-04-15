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
	"mayfly-go/pkg/utils/collx"
	"mayfly-go/pkg/utils/stringx"
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
			&middleware.ApprovalMiddleware{},
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
func (a *Agent) Run(ctx context.Context, messages []adk.Message, runOpts ...runOption) (string, error) {
	ctx = contextx.WithTraceId(ctx)

	runOptions := &runOptions{}
	for _, opt := range runOpts {
		opt(runOptions)
	}
	if runOptions.sessionKey != "" {
		ctx = session.WithSessionKey(ctx, runOptions.sessionKey)
	}

	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		EnableStreaming: true,
		Agent:           a.agent,
		CheckPointStore: NewInMemoryStore(),
	})

	contextMessages, err := a.contextManager.BuildMessages(ctx)
	if err != nil {
		logx.InfoContext(ctx, err.Error())
		contextMessages = []adk.Message{}
	}

	adkRunOptions := runOptions.adkRunOptions
	adkRunOptions = append(adkRunOptions,
		adk.WithCallbacks(logCallback),
		adk.WithCheckPointID(session.GetSessionKey(ctx)))
	iter := runner.Run(ctx, slices.Concat(contextMessages, messages), adkRunOptions...)
	var outputMessages []adk.Message
	messageId := stringx.RandUUID()
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}

		err = event.Err
		if err != nil {
			break
		}

		var msg adk.Message
		sr := event.Output.MessageOutput.MessageStream
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
						logx.Warnf("stream recv error: %v", err)
						break
					}
					chunkMessages = append(chunkMessages, chunk)
					if err := runOptions.CallOnChunk(ctx, chunk); err != nil {
						logx.Warnf("onStreaming callback error: %v", err)
						break
					}
				}
				if len(chunkMessages) > 0 {
					if message, err := schema.ConcatMessages(chunkMessages); err != nil {
						logx.Warnf("concat streamed messages error: %v", err)
					} else {
						msg = message
					}
				}
			}()
		} else {
			msg = event.Output.MessageOutput.Message
		}
		if msg == nil {
			break
		}

		if msg.ToolCallID != "" {
			if _, ok := adk.GetSessionValue(ctx, msg.ToolCallID); ok {
				m := collx.M(msg.Extra)
				msg.Extra = *m.Set("toolStatus", "error")
			}
		}
		SetMessageId(msg, messageId)
		outputMessages = append(outputMessages, msg)
		if err := runOptions.CallOnEvent(ctx, event, msg); err != nil {
			logx.Warnf("onEvent callback error: %v", err)
			break
		}

		LogEventAndMsg(ctx, event, msg)
	}

	if err != nil {
		if toolErr, ok := errors.AsType[*tools.ToolError](err); ok {
			// 工具调用失败，并且没有重试，则记录对应错误消息
			toolErrMsg := &schema.Message{
				Role:       schema.Tool,
				Content:    tools.GetToolErrorMsg(err),
				ToolCallID: toolErr.ToolCallId,
				ToolName:   toolErr.ToolName,
			}
			errMsg := &schema.Message{
				Role:    schema.Assistant,
				Content: err.Error(),
			}
			SetMessageId(toolErrMsg, messageId)
			SetMessageId(errMsg, messageId)
			runOptions.CallOnEvent(ctx, nil, toolErrMsg)

			outputMessages = append(outputMessages, toolErrMsg, errMsg)
		} else {
			logx.ErrorContext(ctx, err.Error())
			return "", err
		}
	}

	if len(outputMessages) > 0 {
		// 追加输出消息到上下文中，构造要保存的消息列表：先存用户消息，再存AI回复，保持对话顺序
		if err := a.contextManager.AppendMsgs(ctx, slices.Concat(messages, outputMessages)...); err != nil {
			logx.ErrorfContext(ctx, "agent append message error: %v", err)
		}
		return outputMessages[len(outputMessages)-1].Content, err
	}

	return "finished without output message", err
}
