package persistence

import "mayfly-go/pkg/ioc"

func InitIoc() {
	ioc.RegisterByType[*sessionRepoImpl]()
	ioc.RegisterByType[*sessionMessageRepoImpl]()
}
