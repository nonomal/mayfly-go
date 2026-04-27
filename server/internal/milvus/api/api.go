// Package api 提供 Milvus 相关的 API 接口处理。
package api

import (
	"mayfly-go/pkg/ioc"
)

// InitIoc 初始化 IOC
func InitIoc() {
	ioc.Register(new(Milvus))
	ioc.Register(new(Collection))

}
