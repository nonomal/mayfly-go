package form

import (
	"github.com/milvus-io/milvus/client/v2/entity"
)

type CreateDatabaseForm struct {
	Name       string            `json:"name" validate:"required"`
	Properties map[string]string `json:"properties" validate:"required"`
}

// CreateCollectionForm 创建 Collection 表单
type CreateCollectionForm struct {
	Name        string                       `json:"name" validate:"required"`
	Description string                       `json:"description"`
	ShardsNum   int32                        `json:"shardsNum"`
	Fields      []*CreateCollectionFieldForm `json:"fields" validate:"required"`
}

func (c *CreateCollectionForm) ToSchema() *entity.Schema {
	schema := entity.NewSchema().
		WithName(c.Name).
		WithDescription(c.Description)

	for _, f := range c.Fields {
		field := entity.NewField()
		field.TypeParams = f.TypeParams
		field.IndexParams = f.IndexParams

		field.WithName(f.Name).
			WithIsPrimaryKey(f.IsPrimaryKey).
			WithIsAutoID(f.AutoID).
			WithDescription(f.Description).
			WithDataType(f.DataType).
			WithDim(f.Dim).
			WithElementType(f.ElementType).
			WithIsDynamic(f.IsDynamic).
			WithIsPartitionKey(f.IsPartitionKey).
			WithIsClusteringKey(f.IsClusteringKey).
			WithMaxLength(f.MaxLength).
			WithMaxCapacity(f.MaxCapacity)

		schema.WithField(field)
	}

	return schema
}

// CreateCollectionFieldForm 创建 Collection 字段表单
type CreateCollectionFieldForm struct {
	Name            string                 `json:"name" validate:"required"`
	DataType        entity.FieldType       `json:"dataType" validate:"required"`
	IsPrimaryKey    bool                   `json:"isPrimaryKey"`
	AutoID          bool                   `json:"autoID"` // 主键是否由 Milvus 自动生成，仅支持 INT64。
	Description     string                 `json:"description"`
	Dim             int64                  `json:"dim"`
	ElementType     entity.FieldType       `json:"elementType"`
	IsDynamic       bool                   `json:"isDynamic"`
	TypeParams      map[string]string      `json:"typeParams"`
	IndexParams     map[string]string      `json:"indexParams"`
	IsPartitionKey  bool                   `json:"isPartitionKey"`
	IsClusteringKey bool                   `json:"isClusteringKey"`
	Params          map[string]interface{} `json:"params"`

	MaxLength   int64 `json:"maxLength"`
	MaxCapacity int64 `json:"maxCapacity"`
}

// RenameCollectionForm 重命名 Collection 表单
type RenameCollectionForm struct {
	NewName string `json:"newName" validate:"required"`
}

// CreatePartitionForm 创建分区表单
type CreatePartitionForm struct {
	Name string `json:"name" validate:"required"`
}

// LoadPartitionsForm 加载分区表单
type LoadPartitionsForm struct {
	Names []string `json:"names" validate:"required"`
	Async bool     `json:"async"`
}

// ReleasePartitionsForm 释放分区表单
type ReleasePartitionsForm struct {
	Names []string `json:"names" validate:"required"`
}

// CreateIndexForm 创建索引表单
type CreateIndexForm struct {
	IndexName  string                 `json:"indexName"`
	IndexType  string                 `json:"indexType" validate:"required"`
	MetricType string                 `json:"metricType" validate:"required"`
	Params     map[string]interface{} `json:"params"`
	Async      bool                   `json:"async"`
}

// InsertForm 插入数据表单
type InsertForm struct {
	PartitionName string                 `json:"partitionName"`
	Data          map[string]interface{} `json:"data" validate:"required"`
}

// GenerateMockDataForm 生成样本数据表单
type GenerateMockDataForm struct {
	Count         int    `json:"count" validate:"required,min=1,max=10000"`
	PartitionName string `json:"partitionName"`
}

// DeleteForm 删除数据表单
type DeleteForm struct {
	Expr string `json:"expr" validate:"required"`
}

// QueryForm 查询表单
type QueryForm struct {
	Expr             string         `json:"expr"`              // 过滤表达式，可选
	OutputFields     []string       `json:"outputFields"`      // 输出字段
	ConsistencyLevel int32          `json:"consistency_level"` // 0-4 默认 0,  Strong Session  Bounded Eventually Customized
	PartitionNames   []string       `json:"partitionNames"`    // 分区名列表
	TemplateParams   map[string]any `json:"templateParams"`    // 模板参数
	Page             int            `json:"page"`              // 页码，从 1 开始，默认 1
	PageSize         int            `json:"pageSize"`          // 每页条数，默认 20，最大 100
}

// SearchForm 搜索表单
type SearchForm struct {
	Vectors          []float32 `json:"vectors" validate:"required"`
	ConsistencyLevel int32     `json:"consistency_level"`
	OutputFields     []string  `json:"outputFields"`
}

// FlushForm 刷新表单
type FlushForm struct {
	CollectionNames []string `json:"collectionNames" validate:"required"`
}

// CreateUserForm 创建用户表单
type CreateUserForm struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// UpdatePasswordForm 更新密码表单
type UpdatePasswordForm struct {
	OldPassword string `json:"oldPassword" validate:"required"`
	NewPassword string `json:"newPassword" validate:"required"`
}
type RoleToUserForm struct {
	RoleName string `json:"roleName" validate:"required"`
}

// UpdateRoleForm 创建角色表单
type UpdateRoleForm struct {
	RoleName   string                  `json:"roleName" validate:"required"`
	Privileges map[string]DBPrivileges `json:"privileges"`
}

type DBPrivileges struct {
	Collections map[string]map[string]bool `json:"collections"`
}

// CreateResourceGroupForm 创建资源组表单
type CreateResourceGroupForm struct {
	Name string `json:"name" validate:"required"`
}

// TransferNodeForm 转移节点表单
type TransferNodeForm struct {
	Source  string `json:"source" validate:"required"`
	Target  string `json:"target" validate:"required"`
	NumNode int32  `json:"numNode" validate:"required"`
}

// TransferReplicaForm 转移副本表单
type TransferReplicaForm struct {
	CollectionName string `json:"collectionName" validate:"required"`
	Source         string `json:"source" validate:"required"`
	Target         string `json:"target" validate:"required"`
	NumReplica     int64  `json:"numReplica" validate:"required"`
}

// AlterCollectionFieldForm 修改 Collection 字段表单
type AlterCollectionFieldForm struct {
	FieldName  string            `json:"fieldName" validate:"required"` // 要修改的字段名
	Properties map[string]string `json:"properties" validate:"required"`
}

// AlterCollectionForm 修改 Collection 表单
type AlterCollectionForm struct {
	NewName          string                    `json:"newName"`           // 新名称（重命名）
	Description      string                    `json:"description"`       // 描述
	ConsistencyLevel int32                     `json:"consistency_level"` // 一致性级别
	MmapEnabled      *bool                     `json:"mmapEnabled"`       // MMap 开关
	Properties       map[string]interface{}    `json:"properties"`        // 其他属性
	Fields           []*AddCollectionFieldForm `json:"fields"`            // 新增字段列表
}

// AddCollectionFieldForm 添加 Collection 字段表单（前端格式，使用 snake_case）
type AddCollectionFieldForm struct {
	Name            string            `json:"name" validate:"required"`
	DataType        entity.FieldType  `json:"data_type" validate:"required"`
	IsPrimaryKey    bool              `json:"is_primary_key"`
	AutoID          bool              `json:"auto_id"`
	Description     string            `json:"description"`
	Dim             int64             `json:"dim"`
	ElementType     entity.FieldType  `json:"element_type"`
	IsDynamic       bool              `json:"is_dynamic"`
	IsPartitionKey  bool              `json:"is_partition_key"`
	IsClusteringKey bool              `json:"is_clustering_key"`
	MaxLength       int64             `json:"max_length"`
	MaxCapacity     int64             `json:"max_capacity"`
	Nullable        bool              `json:"nullable"` // 新增字段必须为 nullable
	TypeParams      map[string]string `json:"type_params"`
	IndexParams     map[string]string `json:"index_params"`
}

// ToEntityField 转换为 entity.Field
func (f *AddCollectionFieldForm) ToEntityField() *entity.Field {
	field := entity.NewField()
	field.TypeParams = f.TypeParams
	field.IndexParams = f.IndexParams

	field.WithName(f.Name).
		WithIsPrimaryKey(f.IsPrimaryKey).
		WithIsAutoID(f.AutoID).
		WithDescription(f.Description).
		WithDataType(f.DataType).
		WithDim(f.Dim).
		WithElementType(f.ElementType).
		WithIsDynamic(f.IsDynamic).
		WithIsPartitionKey(f.IsPartitionKey).
		WithIsClusteringKey(f.IsClusteringKey).
		WithMaxLength(f.MaxLength).
		WithMaxCapacity(f.MaxCapacity).
		WithNullable(f.Nullable)

	return field
}

// AddCollectionFieldRequest 添加 Collection 字段请求（单独接口使用，嵌套 Field）
type AddCollectionFieldRequest struct {
	Field *CreateCollectionFieldForm `json:"field" validate:"required"` // 要添加的字段信息
}

// DropCollectionFieldForm 删除 Collection 字段表单
type DropCollectionFieldForm struct {
	FieldName string `json:"fieldName" validate:"required"` // 要删除的字段名
}

type LoadPartitionForm struct {
	PartitionNames []string `json:"partitionNames" validate:"required"`
}
type ReleasePartitionForm struct {
	PartitionNames []string `json:"partitionNames" validate:"required"`
}

type UpdateDatabasePropertiesForm struct {
}

// CreateAliasForm 创建别名表单
type CreateAliasForm struct {
	Alias string `json:"alias" validate:"required"`
}

// SavePrivilegeGroupForm 保存权限组表单
type SavePrivilegeGroupForm struct {
	GroupName  string   `json:"groupName" validate:"required"`
	Privileges []string `json:"privileges"`
}
