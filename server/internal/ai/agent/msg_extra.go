package agent

import (
	"mayfly-go/pkg/utils/collx"

	"github.com/cloudwego/eino/adk"
)

// GetMessageId 获取消息id
func GetMessageId(msg adk.Message) string {
	return collx.M(msg.Extra).GetStr("messageId")
}

// SetMessageId 设置消息id
func SetMessageId(msg adk.Message, messageId string) {
	m := collx.M(msg.Extra)
	msg.Extra = *m.Set("messageId", messageId)
}
