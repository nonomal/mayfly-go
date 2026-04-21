package session

import (
	"context"
	"fmt"
	"mayfly-go/pkg/gox"
	"mayfly-go/pkg/logx"
	"mayfly-go/pkg/utils/collx"
	"mayfly-go/pkg/utils/stringx"
	"time"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
)

// Manager 会话管理器
// 负责会话缓存管理和生命周期管理，底层存储委托给 Store 实现
type Manager struct {
	store         Store          // 底层存储
	summaryConfig *SummaryConfig // 摘要配置（可选）
}

// NewManager 创建会话管理器
// store: 底层存储实现 (如 JSONLStore, MemoryStore 等)
func NewManager(store Store) *Manager {
	return &Manager{
		store: store,
	}
}

// WithSummaryConfig 设置摘要配置（选项模式）
func (m *Manager) WithSummaryConfig(config *SummaryConfig) *Manager {
	if config != nil {
		m.summaryConfig = config
	}
	return m
}

// GetHistory 获取会话历史消息（Manager 层处理 Skip 优化和摘要组装）
func (m *Manager) GetHistory(ctx context.Context, key string, opts ...GetOption) ([]adk.Message, error) {
	// 应用选项配置
	options := defaultGetOptions()
	for _, opt := range opts {
		opt(options)
	}

	// 获取元数据，检查是否有摘要和 Skip
	meta, err := m.GetMeta(ctx, key)
	if err != nil {
		logx.WarnfContext(ctx, "get meta error in GetHistory: %v", err)
		// 如果获取元数据失败，直接返回 Store 的历史消息
		return m.store.GetHistory(ctx, key, options.messageLimit)
	}

	// 计算 Store 层实际需要读取的消息数量
	// 目标：只读取未摘要的消息，避免读取已摘要的历史记录
	var storeLimit int
	if meta != nil && meta.Skip > 0 && meta.Count > meta.Skip {
		// 未摘要的消息数 = 总消息数 - Skip
		unsummarizedCount := meta.Count - meta.Skip
		storeLimit = unsummarizedCount
		logx.DebugfContext(ctx, "count=%d, skip=%d, unsummarizedCount=%d, messageLimit=%d",
			meta.Count, meta.Skip, unsummarizedCount, options.messageLimit)
	} else {
		// 没有 Skip 或 Count <= Skip，直接使用用户的 limit
		storeLimit = options.messageLimit
	}

	// 从 Store 获取历史消息（返回最后 storeLimit 条，即所有未摘要的消息）
	messages, err := m.store.GetHistory(ctx, key, storeLimit)
	if err != nil {
		return nil, err
	}

	// 从所有未摘要的消息中，取最新的 messageLimit 条
	if options.messageLimit > 0 && len(messages) > options.messageLimit {
		messages = messages[len(messages)-options.messageLimit:]
		logx.DebugfContext(ctx, "trimmed to latest %d messages from %d unsummarized messages",
			options.messageLimit, len(messages)+options.messageLimit)
	}

	// 如果存在摘要，将其作为系统消息前置
	if meta != nil && meta.Summary != "" {
		summaryMsg := &schema.Message{
			Role:    schema.System,
			Content: fmt.Sprintf("[之前的对话摘要]\n%s\n\n[以下是新的对话内容]", meta.Summary),
		}
		messages = append([]adk.Message{summaryMsg}, messages...)
		logx.DebugfContext(ctx, "prepended summary message, total %d messages", len(messages))
	}

	return messages, nil
}

// AppendMsgs 追加消息到会话
func (m *Manager) AppendMsgs(ctx context.Context, key string, msgs ...adk.Message) error {
	if key == "" || len(msgs) == 0 {
		return nil
	}

	// 追加消息到底层存储（Store 只负责存储，不更新元数据）
	if err := m.store.AppendMsgs(ctx, key, msgs...); err != nil {
		return err
	}

	meta, err := m.store.GetMeta(ctx, key)
	if err != nil {
		return err
	}
	// 如果元数据不存在，创建新会话
	if meta == nil {
		// 元数据不存在，创建新的
		meta = &SessionMeta{
			Key:       key,
			CreatedAt: time.Now(),
		}
		meta.Extra.Set("title", stringx.Truncate(msgs[0].Content, 50, 30, "..."))
	}

	// 计算新增消息的Token数量
	totalTokens := collx.ArrayReduce(msgs, 0, func(totalToken int, msg adk.Message) int {
		responseMeta := msg.ResponseMeta
		if responseMeta != nil && responseMeta.Usage != nil {
			return totalToken + responseMeta.Usage.TotalTokens
		}
		return totalToken
	})

	// 保存元数据
	meta.Count += len(msgs)
	meta.TokenCount += totalTokens
	meta.UpdatedAt = time.Now()
	if err := m.store.SaveMeta(ctx, meta); err != nil {
		return err
	}

	// 异步检查并执行摘要（如果配置了摘要功能）
	if m.summaryConfig != nil && m.summaryConfig.Enabled {
		gox.Go(func() {
			if err := m.CheckAndSummarize(ctx, key); err != nil {
				logx.ErrorfContext(ctx, "auto summarize error: %v", err)
			}
		})
	}

	return nil
}

// Delete 删除会话
// 同时删除历史消息、元数据和缓存
func (m *Manager) Delete(ctx context.Context, key string) error {
	// 先清空历史消息（可选，确保数据一致性）
	if err := m.store.ClearHistory(ctx, key); err != nil {
		return fmt.Errorf("manager: clear history: %w", err)
	}

	// 再删除元数据
	return m.store.DeleteMeta(ctx, key)
}

// List 列出所有会话
// 从 Store 加载最新的会话元数据列表
func (m *Manager) List(ctx context.Context) ([]*SessionMeta, error) {
	return m.store.ListMetas(ctx)
}

// ClearHistory 清空会话历史消息（保留元数据）
func (m *Manager) ClearHistory(ctx context.Context, key string) error {
	// 调用 Store 清空历史
	return m.store.ClearHistory(ctx, key)
}

// GetMeta 获取会话元数据
func (m *Manager) GetMeta(ctx context.Context, key string) (*SessionMeta, error) {
	return m.store.GetMeta(ctx, key)
}

// SaveMeta 保存会话元数据
func (m *Manager) SaveMeta(ctx context.Context, meta *SessionMeta) error {
	return m.store.SaveMeta(ctx, meta)
}

// CheckAndSummarize 检查并执行自动摘要（使用 Manager 内部配置）
func (m *Manager) CheckAndSummarize(ctx context.Context, sessionKey string) error {
	// 使用 Manager 内部的配置
	if m.summaryConfig == nil || !m.summaryConfig.Enabled {
		return nil
	}

	config := m.summaryConfig

	// 获取会话元数据
	meta, err := m.GetMeta(ctx, sessionKey)
	if err != nil {
		return fmt.Errorf("get session meta: %w", err)
	}

	if meta == nil {
		return nil
	}

	// 计算未摘要的消息数
	unsummarizedCount := meta.Count - meta.Skip

	// 获取保留消息数配置
	keepCount := config.KeepRecentCount
	if keepCount <= 0 {
		keepCount = DefaultSummaryKeepCount
	}

	// 检查是否达到阈值（消息数量或 token 数量）
	needSummarize := false
	reason := ""

	// 条件1：未摘要消息数 >= (阈值 + 保留数)，确保摘要有足够的压缩价值
	// 例如：threshold=5, keepCount=3，则需要至少 8 条消息才触发摘要
	minTriggerCount := config.MessageThreshold + keepCount
	if unsummarizedCount >= minTriggerCount {
		needSummarize = true
		reason = fmt.Sprintf("unsummarized count %d >= minTriggerCount (%d + %d)",
			unsummarizedCount, config.MessageThreshold, keepCount)
	} else if meta.TokenCount >= config.TokenThreshold && unsummarizedCount > keepCount {
		// 条件2：Token 数 >= 阈值，且消息数 > 保留数（Token 触发时放宽条件）
		needSummarize = true
		reason = fmt.Sprintf("token count %d >= threshold %d and unsummarized count %d > keepCount %d",
			meta.TokenCount, config.TokenThreshold, unsummarizedCount, keepCount)
	}

	if !needSummarize {
		logx.DebugfContext(ctx, "skip summarize: unsummarized=%d, minTriggerCount=%d (threshold=%d + keepCount=%d), tokens=%d",
			unsummarizedCount, minTriggerCount, config.MessageThreshold, keepCount, meta.TokenCount)
		return nil
	}

	logx.InfofContext(ctx, "trigger auto summarize, reason: %s", reason)

	// 执行摘要
	return m.summarizeSession(ctx, sessionKey, meta, config)
}

// summarizeSession 对会话进行摘要处理（增量摘要）
func (m *Manager) summarizeSession(ctx context.Context, sessionKey string, meta *SessionMeta, config *SummaryConfig) error {
	keepCount := config.KeepRecentCount
	if keepCount <= 0 {
		keepCount = DefaultSummaryKeepCount
	}

	// 计算需要读取的消息数：未摘要的消息总数
	// 直接调用 store.GetHistory，避免 Manager.GetHistory 内部重复读取 meta
	unsummarizedCount := meta.Count - meta.Skip
	if unsummarizedCount <= 0 {
		logx.WarnfContext(ctx, "no unsummarized messages found, skip summarization")
		return nil
	}

	// 从 Store 层精确读取未摘要的消息（仅原始消息，不含摘要）
	rawMessages, err := m.store.GetHistory(ctx, sessionKey, unsummarizedCount)
	if err != nil {
		return fmt.Errorf("get history: %w", err)
	}

	if len(rawMessages) == 0 {
		logx.WarnfContext(ctx, "no unsummarized messages found after store query, skip summarization")
		return nil
	}

	logx.InfofContext(ctx, "summarizing %d unsummarized messages (skip=%d, count=%d, keep=%d)",
		len(rawMessages), meta.Skip, meta.Count, keepCount)

	// 获取摘要器
	summarizer := config.Summarizer
	if summarizer == nil {
		logx.WarnfContext(ctx, "summarizer is not configured")
		return nil
	}

	// 组装完整的摘要输入：旧摘要（如有）+ 未摘要的原始消息
	var fullContext []adk.Message

	// 如果存在旧摘要，作为 System 消息前置
	if meta.Summary != "" {
		summaryMsg := &schema.Message{
			Role:    schema.System,
			Content: fmt.Sprintf("[之前的对话摘要]\n%s\n\n[以下是新的对话内容]", meta.Summary),
		}
		fullContext = append(fullContext, summaryMsg)
		logx.DebugfContext(ctx, "prepended old summary to context")
	}

	// 追加所有未摘要的原始消息
	fullContext = append(fullContext, rawMessages...)

	logx.DebugfContext(ctx, "full context for summarization: %d messages (1 system + %d raw)",
		len(fullContext), len(rawMessages))

	// 裁剪需要摘要的消息：保留最后 keepCount 条，摘要前面的部分
	// 注意：触发条件已保证 len(fullContext) > keepCount
	var messagesToSummarize []adk.Message
	if len(fullContext) > keepCount {
		// 只取前面的部分进行摘要（包含旧摘要 System 消息）
		messagesToSummarize = fullContext[:len(fullContext)-keepCount]
		logx.DebugfContext(ctx, "will summarize %d messages, keeping last %d messages",
			len(messagesToSummarize), keepCount)
	} else {
		// 理论上不会走到这里（触发条件已过滤），但保留作为防御性编程
		logx.WarnfContext(ctx, "unexpected: full context count (%d) <= keepCount (%d), skip summarization",
			len(fullContext), keepCount)
		return nil
	}

	// 使用裁剪后的消息生成新摘要
	summaryText, err := summarizer.GenerateSummary(ctx, messagesToSummarize)
	if err != nil {
		logx.ErrorfContext(ctx, "generate summary error: %v", err)
		// 摘要生成失败不影响主流程，仅记录日志
		return nil
	}

	// 计算新的 Skip 值：跳过已被摘要的原始消息（不包括 System 摘要消息）
	newSkipCount := meta.Skip + len(rawMessages) - keepCount
	if newSkipCount < 0 {
		newSkipCount = 0
	}

	// 更新元数据：设置新摘要和 Skip 偏移量
	meta.Summary = summaryText
	meta.Skip = newSkipCount

	// 重新计算保留消息的 token 数（只计算最近 keepCount 条消息）
	meta.TokenCount = 0
	if len(fullContext) >= keepCount {
		// 只计算保留的最近消息的 token
		for _, msg := range fullContext[len(fullContext)-keepCount:] {
			if msg.ResponseMeta != nil && msg.ResponseMeta.Usage != nil {
				meta.TokenCount += msg.ResponseMeta.Usage.TotalTokens
			}
		}
	} else {
		// 如果历史消息少于 keepCount，计算所有消息的 token
		for _, msg := range fullContext {
			if msg.ResponseMeta != nil && msg.ResponseMeta.Usage != nil {
				meta.TokenCount += msg.ResponseMeta.Usage.TotalTokens
			}
		}
	}

	meta.UpdatedAt = time.Now()

	// 重新读取最新的元数据，避免并发覆盖
	latestMeta, err := m.GetMeta(ctx, sessionKey)
	if err != nil {
		logx.WarnfContext(ctx, "get latest meta error: %v, using current meta", err)
		latestMeta = meta // 降级：使用当前 meta
	} else if latestMeta != nil {
		// 合并：保留最新的 Count 和 TokenCount，但使用新计算的 Summary 和 Skip
		latestMeta.Summary = meta.Summary
		latestMeta.Skip = meta.Skip
		latestMeta.TokenCount = meta.TokenCount // 使用重新计算的 token 数
		latestMeta.UpdatedAt = meta.UpdatedAt
		meta = latestMeta
	}

	if err := m.SaveMeta(ctx, meta); err != nil {
		return fmt.Errorf("save meta: %w", err)
	}

	logx.InfofContext(ctx, "summarize completed, summary length: %d, skipped messages: %d, kept messages: %d",
		len(summaryText), newSkipCount, keepCount)

	return nil
}
