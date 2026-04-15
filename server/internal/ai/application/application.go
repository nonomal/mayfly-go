package application

import (
	"mayfly-go/internal/ai/agent"
	"mayfly-go/pkg/ioc"
)

func Init() {
	sessionAppImpl := new(sessionAppImpl)
	ioc.Register(sessionAppImpl)
	// 注册session存储
	agent.DefaultSessionStore = sessionAppImpl
}
