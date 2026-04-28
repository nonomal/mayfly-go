package agent

import (
	"context"
	"errors"
	"fmt"
	"mayfly-go/internal/ai/memory"
	"mayfly-go/internal/ai/session"
	"mayfly-go/pkg/contextx"
	"mayfly-go/pkg/eventbus"
	"mayfly-go/pkg/logx"
	"sync"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
)

// GetDefaultContextManager 获取默认的上下文管理器实例
func GetDefaultContextManager() (*ContextManager, error) {
	var sessionStore session.Store
	var err error

	if session.DefaultSessionStore != nil {
		sessionStore = session.DefaultSessionStore
	} else {
		sessionStore, err = session.NewStoreJSONL("./sessions")
		if err != nil {
			return nil, fmt.Errorf("create session store: %w", err)
		}
		session.DefaultSessionStore = sessionStore
	}

	chatModel, err := GetChatModel(context.Background())
	if err != nil {
		return nil, fmt.Errorf("get chat model: %w", err)
	}

	// 创建会话管理器并配置摘要功能
	sessionManager := session.NewManager(sessionStore)
	summaryConfig := session.DefaultSummaryConfig()

	// 如果提供了 ChatModel，注入到摘要器
	if chatModel != nil && summaryConfig.Summarizer != nil {
		if llmSummarizer, ok := summaryConfig.Summarizer.(*session.LLMSummarizer); ok {
			llmSummarizer.WithChatModel(chatModel)
			logx.Info("ChatModel injected into LLM Summarizer")
		}
	}
	sessionManager.WithSummaryConfig(summaryConfig)

	config := &ContextManagerConfig{
		SessionManager: sessionManager,
		ChatModel:      chatModel,
	}

	ctxManager, err := NewContextManager(config)
	if err != nil {
		return nil, err
	}

	logx.Info("Default ContextManager created, please inject ChatModel using WithChatModel() if LLM features are needed")
	return ctxManager, nil
}

// ContextManagerConfig ContextManager 配置
type ContextManagerConfig struct {
	SessionManager *session.Manager           // 会话管理器（必需）
	MemoryManager  *memory.Manager            // 记忆管理器（可选）
	ChatModel      model.ToolCallingChatModel // ChatModel 实例
}

// DefaultContextManagerConfig 返回默认配置
func DefaultContextManagerConfig() *ContextManagerConfig {
	return &ContextManagerConfig{}
}

type ContextManager struct {
	sessionManager *session.Manager           // 会话管理器
	memoryManager  *memory.Manager            // 统一记忆管理器（包含短期和长期）
	chatModel      model.ToolCallingChatModel // ChatModel 实例，用于 LLM 调用
	mu             sync.RWMutex               // 读写锁，保护并发访问
}

// NewContextManager 创建并初始化 ContextManager 实例
func NewContextManager(config *ContextManagerConfig) (*ContextManager, error) {
	if config == nil {
		config = DefaultContextManagerConfig()
	}

	// 验证必需参数
	if config.SessionManager == nil {
		return nil, fmt.Errorf("SessionManager is required")
	}

	// ChatModel 必须由外部注入（可选）
	chatModel := config.ChatModel

	// 自动初始化 MemoryManager（如果提供了 Store）
	memoryManager := config.MemoryManager

	cm := &ContextManager{
		sessionManager: config.SessionManager,
		memoryManager:  memoryManager,
		chatModel:      chatModel,
	}

	// 如果配置了 MemoryManager 和 ChatModel，自动配置 LLM Extractor
	if memoryManager != nil && chatModel != nil {
		extractor := memory.NewLLMExtractor()
		extractor.WithConfig(&memory.LLMExtractorConfig{
			Enabled:       true,
			MinConfidence: 0.7,
			ChatModel:     chatModel,
		})
		memoryManager.WithExtractor(extractor)
		logx.Info("LLM memory extractor auto-configured")
	} else if memoryManager != nil && chatModel == nil {
		logx.Warn("ChatModel not provided, memory extraction will be disabled")
	}

	// 注册记忆提取事件订阅：通过事件总线解耦 session 摘要和 memory 提取
	// 避免 session 包直接依赖 memory 包，后续可灵活替换或移除记忆模块
	if memoryManager != nil {
		registerMemoryExtraction(cm)
	}

	return cm, nil
}

// WithChatModel 手动覆盖 ChatModel 实例（可选）
func (c *ContextManager) WithChatModel(chatModel model.ToolCallingChatModel) *ContextManager {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.chatModel = chatModel
	return c
}

// GetSessionKey 获取会话Key
func (c *ContextManager) GetSessionKey(ctx context.Context) string {
	return session.GetSessionKey(ctx)
}

// BuildMessages 从上下文中构建消息列表，供Agent执行使用
// 在获取上下文时，会自动检查并应用摘要和短期记忆（如果存在）
func (c *ContextManager) BuildMessages(ctx context.Context) ([]adk.Message, error) {
	sessionKey := c.GetSessionKey(ctx)
	if sessionKey == "" {
		return nil, errors.New("session key is empty")
	}

	// 获取历史消息
	history, err := c.sessionManager.GetHistory(ctx, sessionKey)
	if err != nil {
		return nil, err
	}

	// 构建系统消息列表
	var systemMessages []adk.Message

	// 注入记忆
	if c.memoryManager != nil {
		la := contextx.GetLoginAccount(ctx)
		if la != nil {
			memoryMsg := c.memoryManager.BuildMemoryMessage(ctx, fmt.Sprintf("%d", la.Id))
			if memoryMsg != nil {
				systemMessages = append(systemMessages, memoryMsg)
			}
		}
	}

	// 将系统消息插入到历史消息前面
	if len(systemMessages) > 0 {
		history = append(systemMessages, history...)
		logx.InfofContext(ctx, "injected %d system messages into context", len(systemMessages))
	}

	return history, nil
}

// AppendMsgs 追加消息到会话，并在达到阈值时触发自动摘要
func (c *ContextManager) AppendMsgs(ctx context.Context, msgs ...adk.Message) error {
	if len(msgs) == 0 {
		return nil
	}

	sessionKey := c.GetSessionKey(ctx)
	if sessionKey == "" {
		return errors.New("session key is empty")
	}

	// 先追加消息
	if err := c.sessionManager.AppendMsgs(ctx, sessionKey, msgs...); err != nil {
		return err
	}

	// 异步提取并保存记忆
	// if c.memoryManager != nil {
	// 	gox.Go(func() {
	// 		la := contextx.GetLoginAccount(ctx)
	// 		if la == nil {
	// 			logx.WarnfContext(ctx, "no login account found in context, skipping memory extraction")
	// 			return
	// 		}

	// 		if err := c.memoryManager.ExtractAndSave(ctx, &memory.ExtractMemoryReq{
	// 			UserId: fmt.Sprintf("%d", la.Id),
	// 			Msgs:   msgs,
	// 		}); err != nil {
	// 			logx.ErrorfContext(ctx, "auto extract memories error: %v", err)
	// 		}
	// 	})
	// }

	return nil
}

// ClearHistory 清空会话历史
func (c *ContextManager) ClearHistory(ctx context.Context) error {
	return c.sessionManager.ClearHistory(ctx, c.GetSessionKey(ctx))
}

// GetSessionMeta 获取会话元数据
func (c *ContextManager) GetSessionMeta(ctx context.Context) (*session.SessionMeta, error) {
	sessionKey := c.GetSessionKey(ctx)
	if sessionKey == "" {
		return nil, errors.New("session key is empty")
	}
	return c.sessionManager.GetMeta(ctx, sessionKey)
}

// registerMemoryExtraction 注册会话摘要完成事件订阅器
// 当 session.Manager 完成自动摘要后，通过事件总线异步触发长期记忆提取
// 使用固定 subId 避免重复注册，后注册的实例会覆盖前者
func registerMemoryExtraction(cm *ContextManager) {
	session.EventBus.SubscribeAsync(session.EventTopicSummarized, "AgentMemoryExtractor", func(ctx context.Context, event *eventbus.Event[any]) error {
		evt, ok := event.Val.(*session.SummarizedEvent)
		if !ok {
			return nil
		}
		if cm.memoryManager == nil || evt.UserId == "" {
			return nil
		}

		// 获取当前会话历史消息用于记忆提取
		history, err := cm.sessionManager.GetHistory(ctx, evt.SessionKey)
		if err != nil {
			logx.WarnfContext(ctx, "get history for memory extraction failed: %v", err)
			return nil // 不阻塞事件总线
		}

		if err := cm.memoryManager.ExtractAndSave(ctx, &memory.ExtractMemoryReq{
			UserId: evt.UserId,
			Msgs:   history,
		}); err != nil {
			logx.ErrorfContext(ctx, "auto extract memories error: %v", err)
		}
		return nil
	}, false)
}
