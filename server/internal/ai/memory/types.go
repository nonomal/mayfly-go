package memory

import "github.com/cloudwego/eino/adk"

// ExtractMemoryReq 提取记忆的请求参数
type ExtractMemoryReq struct {
	UserId string        // 用户ID
	Msgs   []adk.Message // 消息列表
}
