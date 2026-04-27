package init

import (
	"mayfly-go/internal/milvus/api"
	"mayfly-go/internal/milvus/application"
	"mayfly-go/internal/milvus/infra/persistence"
	"mayfly-go/pkg/starter"
)

func init() {
	starter.AddInitIocFunc(func() {
		persistence.InitIoc()
		application.InitIoc()
		api.InitIoc()
	})
}
