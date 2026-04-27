package vo

// CollectionVO Collection 信息
type CollectionVO struct {
	Name             string      `json:"name"`
	Description      string      `json:"description"`
	Schema           *SchemaVO   `json:"schema"`
	Partitions       []string    `json:"partitions"`
	Indexes          []IndexInfo `json:"indexes"`
	Loaded           bool        `json:"loaded"`
	EntityCount      int64       `json:"entityCount"`
	CreatedTime      string      `json:"createdTime"`
	ConsistencyLevel string      `json:"consistencyLevel"`
	ShardsNum        int32       `json:"shardsNum"`
}

// SchemaVO Schema 信息
type SchemaVO struct {
	Fields      []*FieldVO `json:"fields"`
	Description string     `json:"description"`
	AutoID      bool       `json:"autoID"`
}

// FieldVO 字段信息
type FieldVO struct {
	Name         string `json:"name"`
	DataType     string `json:"dataType"`
	IsPrimaryKey bool   `json:"isPrimaryKey"`
	AutoID       bool   `json:"autoID"`
	Description  string `json:"description"`
	ElementType  string `json:"elementType"`
	Dim          int64  `json:"dim,omitempty"`
}

// IndexInfo 索引信息
type IndexInfo struct {
	FieldName  string `json:"fieldName"`
	IndexName  string `json:"indexName"`
	IndexType  string `json:"indexType"`
	MetricType string `json:"metricType"`
	Params     string `json:"params"`
}

// PartitionVO 分区信息
type PartitionVO struct {
	Id         int64  `json:"id"`
	Name       string `json:"name"`
	CreateTime string `json:"createTime"`
	RowCount   int64  `json:"rowCount"`
}

// DatabaseVO 数据库信息
type DatabaseVO struct {
	Name       string            `json:"name"`
	Properties map[string]string `json:"properties"`
}

// UserVO 用户信息
type UserVO struct {
	Name  string   `json:"name"`
	Roles []string `json:"roles"`
}

// RoleVO 角色信息
type RoleVO struct {
	Name       string          `json:"name"`
	Privileges []PrivilegeInfo `json:"privileges"`
}

// PrivilegeInfo 权限信息
type PrivilegeInfo struct {
	ObjectType string `json:"objectType"`
	ObjectName string `json:"objectName"`
	Privilege  string `json:"privilege"`
}

// ResourceGroupVO 资源组信息
type ResourceGroupVO struct {
	Name                 string           `json:"name"`
	Capacity             int32            `json:"capacity"`
	AvailableNodesNumber int32            `json:"availableNodesNumber"`
	LoadedReplica        map[string]int32 `json:"loadedReplica"`
}

// ServerInfoVO 服务器信息
type ServerInfoVO struct {
	Version      string `json:"version"`
	DeployMode   string `json:"deployMode"`
	BuildVersion string `json:"buildVersion"`
	BuildTime    string `json:"buildTime"`
	GitCommit    string `json:"gitCommit"`
	GoVersion    string `json:"goVersion"`
}

// HealthStatusVO 健康状态
type HealthStatusVO struct {
	IsHealthy   bool                     `json:"isHealthy"`
	Reasons     []string                 `json:"reasons"`
	QuotaStates []map[string]interface{} `json:"quotaStates"`
}

// SearchParam 搜索参数
type SearchParam struct {
	CollectionName string                 `json:"collectionName" validate:"required"`
	PartitionNames []string               `json:"partitionNames"`
	Expr           string                 `json:"expr"`
	OutputFields   []string               `json:"outputFields"`
	Vectors        [][]float32            `json:"vectors" validate:"required"`
	VectorField    string                 `json:"vectorField" validate:"required"`
	MetricType     string                 `json:"metricType" validate:"required"`
	TopK           int                    `json:"topK" validate:"required"`
	IndexParams    map[string]interface{} `json:"indexParams"`
}

// QueryParam 查询参数
type QueryParam struct {
	CollectionName string   `json:"collectionName" validate:"required"`
	PartitionNames []string `json:"partitionNames"`
	Expr           string   `json:"expr" validate:"required"`
	OutputFields   []string `json:"outputFields"`
}

// InsertParam 插入参数
type InsertParam struct {
	CollectionName string                 `json:"collectionName" validate:"required"`
	PartitionName  string                 `json:"partitionName"`
	Data           map[string]interface{} `json:"data" validate:"required"`
}

// CreateCollectionParam 创建 Collection 参数
type CreateCollectionParam struct {
	Name        string              `json:"name" validate:"required"`
	Description string              `json:"description"`
	ShardsNum   int32               `json:"shardsNum"`
	Fields      []*CreateFieldParam `json:"fields" validate:"required"`
}

// CreateFieldParam 创建字段参数
type CreateFieldParam struct {
	Name         string                 `json:"name" validate:"required"`
	DataType     string                 `json:"dataType" validate:"required"`
	IsPrimaryKey bool                   `json:"isPrimaryKey"`
	AutoID       bool                   `json:"autoID"`
	Description  string                 `json:"description"`
	Dim          int64                  `json:"dim"`
	ElementType  string                 `json:"elementType"`
	Params       map[string]interface{} `json:"params"`
}

// CreateIndexParam 创建索引参数
type CreateIndexParam struct {
	CollectionName string                 `json:"collectionName" validate:"required"`
	FieldName      string                 `json:"fieldName" validate:"required"`
	IndexName      string                 `json:"indexName"`
	IndexType      string                 `json:"indexType" validate:"required"`
	MetricType     string                 `json:"metricType" validate:"required"`
	Params         map[string]interface{} `json:"params"`
}
