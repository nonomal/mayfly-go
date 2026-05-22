package vo

import (
	tagentity "mayfly-go/internal/tag/domain/entity"
	"mayfly-go/pkg/model"
)

type Milvus struct {
	model.Model
	tagentity.AuthCerts // 授权凭证信息

	Code               string `json:"code"`
	Name               string `json:"name"`
	Host               string `json:"host"`
	Database           string `json:"database"`
	SshTunnelMachineId int    `json:"sshTunnelMachineId"`
}

func (m *Milvus) GetCode() string {
	return m.Code
}
