package repository

import (
	"mayfly-go/internal/milvus/domain/entity"
	"mayfly-go/pkg/base"
	"mayfly-go/pkg/model"
)

type Milvus interface {
	base.Repo[*entity.Milvus]

	// 分页获取列表
	GetList(condition *entity.MilvusQuery, orderBy ...string) (*model.PageResult[*entity.Milvus], error)
}
