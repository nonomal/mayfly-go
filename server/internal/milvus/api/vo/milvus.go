package vo

import (
	tagentity "mayfly-go/internal/tag/domain/entity"
	"mayfly-go/pkg/model"
)

type Milvus struct {
	model.Model
	tagentity.ResourceTags

	Code     string `json:"code"`
	Name     string `json:"name"`
	Host     string `json:"host"`
	Username string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database" gorm:"size:100;comment:数据库名;default:default"` // 数据库名，默认为 default

	SshTunnelMachineId int `json:"sshTunnelMachineId"` // ssh隧道机器id
}

func (m *Milvus) GetCode() string {
	return m.Code
}
