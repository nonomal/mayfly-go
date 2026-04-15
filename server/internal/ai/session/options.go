package session

// GetOptions 获取或创建会话的配置选项
type GetOptions struct {
	messageLimit int // 消息条数限制，0 表示加载全部
}

// GetOptions 选项函数类型
type GetOption func(*GetOptions)

// WithGetMessageLimit 设置加载的历史消息条数
// limit: 消息条数，0 表示加载全部历史消息
func WithGetMessageLimit(limit int) GetOption {
	return func(o *GetOptions) {
		o.messageLimit = limit
	}
}

// defaultGetOptions 返回默认配置
func defaultGetOptions() *GetOptions {
	return &GetOptions{
		messageLimit: 100000, // 默认加载全部
	}
}
