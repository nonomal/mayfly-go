package form

import (
	tagentity "mayfly-go/internal/tag/domain/entity"
	"mayfly-go/pkg/model"
)

// Milvus Milvus 实例表单
type Milvus struct {
	model.ExtraData

	Id                 uint64 `json:"id"`
	Code               string `json:"code" validate:"required"`
	Name               string `json:"name" validate:"required"`
	Host               string `json:"host" validate:"required"`
	Database           string `json:"database"`
	SshTunnelMachineId int    `json:"sshTunnelMachineId"`

	AuthCerts    []*tagentity.ResourceAuthCert `json:"authCerts" binding:"required"`
	TagCodePaths []string                      `json:"tagCodePaths" binding:"required"`
}
