package entity

import (
	"mayfly-go/pkg/model"
)

// MilvusQuery Milvus 查询条件
type MilvusQuery struct {
	model.Model
	model.PageParam

	Code               string `json:"code" query:"code" form:"code"`          // code
	Name               string `json:"name" query:"name" form:"name"`          // 名称
	Keyword            string `json:"keyword" query:"keyword" form:"keyword"` // 关键字
	SshTunnelMachineId uint64 // ssh隧道机器id
	TagPath            string `json:"tagPath" query:"tagPath" form:"tagPath"` // 标签路径

	Codes []string `json:"codes" query:"codes" form:"codes"` // code 列表
}
