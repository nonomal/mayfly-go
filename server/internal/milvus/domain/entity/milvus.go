package entity

import (
	"mayfly-go/internal/milvus/mvm"
	"mayfly-go/pkg/model"
	"mayfly-go/pkg/utils/structx"
)

// Milvus Milvus 实例信息
type Milvus struct {
	model.Model

	Code               string  `json:"code" gorm:"size:32;comment:code"`                      // code
	Name               string  `json:"name" gorm:"not null;size:50;comment:名称"`               // 名称
	Host               string  `json:"host" gorm:"not null;size:255;comment:连接地址"`            // 连接地址，格式：host:port
	Username           *string `json:"username" gorm:"size:100;comment:用户名"`                  // 用户名
	Password           *string `json:"password" gorm:"size:100;comment:密码"`                   // 密码
	Database           string  `json:"database" gorm:"size:100;comment:数据库名;default:default"` // 数据库名，默认为 default
	SshTunnelMachineId int     `json:"sshTunnelMachineId" gorm:"comment:ssh 隧道的机器 id"`        // ssh 隧道机器 id
}

// TableName 表名
func (m *Milvus) TableName() string {
	return "t_milvus"
}

// ToMilvusInfo 转换为 milvusInfo 进行连接
func (m *Milvus) ToMilvusInfo() *mvm.MilvusInfo {
	milvusInfo := new(mvm.MilvusInfo)
	_ = structx.Copy(milvusInfo, m)
	return milvusInfo
}
