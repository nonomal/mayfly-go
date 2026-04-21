package form

type ChatMsgType string

const (
	ChatMsgTypeText            ChatMsgType = "text"            // 文本消息
	ChatMsgTypeInterruptResume ChatMsgType = "interruptResume" // 恢复中断
)

type ChatMsg struct {
	SessionId string      `json:"sessionId"`
	Type      ChatMsgType `json:"type"`
	Content   string      `json:"content"`
}

type InterruptResume struct {
	InterruptId string `json:"interruptId"`
	Action      string `json:"action"`
	Payload     any    `json:"payload"`
}
