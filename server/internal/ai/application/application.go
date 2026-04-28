package application

import (
	"mayfly-go/internal/ai/session"
	"mayfly-go/pkg/ioc"
)

func Init() {
	sessionAppImpl := new(sessionAppImpl)
	ioc.Register(sessionAppImpl)
	// 注册session存储
	session.DefaultSessionStore = sessionAppImpl
}
