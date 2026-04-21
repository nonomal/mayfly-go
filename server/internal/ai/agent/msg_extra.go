package agent

import (
	"mayfly-go/pkg/utils/collx"

	"github.com/cloudwego/eino/adk"
)

// GetTurnId 获取turn id
func GetTurnId(msg adk.Message) string {
	return collx.M(msg.Extra).GetStr("turnId")
}

// SetTurnId 设置tern id
func SetTurnId(msg adk.Message, turnId string) {
	SetMessageExtra(msg, "turnId", turnId)
}

func SetActionId(msg adk.Message, actionId string) {
	SetMessageExtra(msg, "actionId", actionId)
}

func GetActionId(msg adk.Message) string {
	return collx.M(msg.Extra).GetStr("actionId")
}

func SetToolStatus(msg adk.Message, status string) {
	SetMessageExtra(msg, "toolStatus", status)
}

// SetMessageExtra 设置message extra
func SetMessageExtra(msg adk.Message, key string, value any) {
	m := collx.M(msg.Extra)
	msg.Extra = *m.Set(key, value)
}
