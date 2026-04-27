package form

// Milvus Milvus 实例表单
type Milvus struct {
	Id                 uint64 `json:"id"`
	Code               string `json:"code" validate:"required"`
	Name               string `json:"name" validate:"required"`
	Host               string `json:"host" validate:"required"`
	Username           string `json:"username"`
	Password           string `json:"password"`
	Database           string `json:"database"`
	SshTunnelMachineId int    `json:"sshTunnelMachineId"`

	TagCodePaths []string `json:"tagCodePaths" binding:"required"`
}
