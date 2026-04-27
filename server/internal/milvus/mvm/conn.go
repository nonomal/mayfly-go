package mvm

import (
	"context"
	"fmt"
	"mayfly-go/internal/milvus/api/form"
	"mayfly-go/internal/milvus/api/vo"
	"mayfly-go/pkg/biz"
	"mayfly-go/pkg/errorx"
	"mayfly-go/pkg/utils/timex"
	"slices"
	"strconv"
	"time"

	"github.com/milvus-io/milvus-proto/go-api/v2/commonpb"
	"github.com/milvus-io/milvus-proto/go-api/v2/milvuspb"
	"github.com/milvus-io/milvus/client/v2/column"
	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/index"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/samber/lo"
)

// MilvusConn Milvus 客户端实现
type MilvusConn struct {
	Id      uint64
	Service milvuspb.MilvusServiceClient
	cli     *milvusclient.Client // 使用新的 client/v2
	info    *MilvusInfo
	DbName  string
	Version string // 版本信息
}

// Close 关闭连接
func (mc *MilvusConn) Close() error {
	if mc.cli != nil {
		_ = mc.cli.Close(context.Background())
		mc.cli = nil
		mc.Service = nil
	}
	return nil
}

func (mc *MilvusConn) Ping() error {
	mc.CheckConnect()

	_, err := mc.GetVersion()
	if err != nil {
		return fmt.Errorf("milvus 连接已断开：%s", err.Error())
	}

	return nil
}

// CheckConnect 是否已连接
func (mc *MilvusConn) CheckConnect() {
	if mc.cli == nil || mc.Service == nil {
		panic("未连接到 Milvus")
	}
}

// GetClient 获取底层客户端
func (mc *MilvusConn) GetClient() *milvusclient.Client {
	return mc.cli
}

// GetVersion 获取版本信息
func (mc *MilvusConn) GetVersion() (string, error) {
	mc.CheckConnect()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	res, err := mc.cli.GetServerVersion(ctx, milvusclient.NewGetServerVersionOption())
	if err != nil {
		return "", err
	}
	return res, nil
}

// CheckHealth 检查健康状态
func (mc *MilvusConn) CheckHealth() (*vo.HealthStatusVO, error) {
	mc.CheckConnect()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 使用 MilvusServiceClient 调用 CheckHealth
	resp, err := mc.Service.CheckHealth(ctx, &milvuspb.CheckHealthRequest{})
	if err != nil {
		return nil, err
	}

	healthStatus := &vo.HealthStatusVO{
		IsHealthy: resp.IsHealthy,
		Reasons:   resp.Reasons,
	}

	// 处理配额状态
	if len(resp.QuotaStates) > 0 {
		quotaStates := make([]map[string]interface{}, 0, len(resp.QuotaStates))
		for _, qs := range resp.QuotaStates {
			quotaStates = append(quotaStates, map[string]interface{}{
				"state": qs.String(),
			})
		}
		healthStatus.QuotaStates = quotaStates
	}

	return healthStatus, nil
}

// ============ 数据库操作 ============

// ListDatabases 列出所有数据库
func (mc *MilvusConn) ListDatabases() ([]map[string]any, error) {
	mc.CheckConnect()
	ctx := context.Background()

	res, err := mc.Service.ListDatabases(ctx, &milvuspb.ListDatabasesRequest{})
	if err != nil {
		return nil, err
	}
	err = handleRespStatus(res.Status, err)
	if err != nil {
		return nil, err
	}

	var dbDetails []map[string]any
	for i, name := range res.DbNames {
		// 描述数据库
		dbDesc, err := mc.cli.DescribeDatabase(ctx, milvusclient.NewDescribeDatabaseOption(name))
		if err != nil {
			return nil, err
		}

		m := make(map[string]any)
		m["id"] = res.DbIds[i]
		m["properties"] = dbDesc.Properties
		m["name"] = name

		timestamp := int64(res.CreatedTimestamp[i])
		// 转换为 time.Time
		t := time.Unix(0, timestamp)
		m["create_time"] = t.Format(time.DateTime)
		dbDetails = append(dbDetails, m)
	}

	// 按 name 升序排序
	slices.SortFunc(dbDetails, func(a, b map[string]any) int {
		nameA := a["name"].(string)
		nameB := b["name"].(string)
		if nameA < nameB {
			return -1
		} else if nameA > nameB {
			return 1
		}
		return 0
	})

	return dbDetails, nil
}

// CreateDatabase 创建数据库
func (mc *MilvusConn) CreateDatabase(param *form.CreateDatabaseForm) error {
	mc.CheckConnect()

	if len(param.Properties) == 0 {
		param.Properties = map[string]string{
			"timezone": "Asia/Shanghai",
		}
	}

	ctx := context.Background()

	option := milvusclient.NewCreateDatabaseOption(param.Name)
	option.Properties = param.Properties
	err := mc.cli.CreateDatabase(ctx, option)
	return err
}

// DropDatabase 删除数据库
func (mc *MilvusConn) DropDatabase(dbName string) error {
	mc.CheckConnect()
	option := milvusclient.NewDropDatabaseOption(dbName)
	return mc.cli.DropDatabase(context.Background(), option)
}

func (mc *MilvusConn) DescribeDatabase(dbName string) (*entity.Database, error) {
	mc.CheckConnect()

	return mc.cli.DescribeDatabase(context.Background(), milvusclient.NewDescribeDatabaseOption(dbName))
}

func (mc *MilvusConn) AlterDatabase(param *form.CreateDatabaseForm) error {
	mc.CheckConnect()
	option := milvusclient.NewAlterDatabasePropertiesOption(param.Name)
	for k, v := range param.Properties {
		option.WithProperty(k, v)
	}
	err := mc.cli.AlterDatabaseProperties(context.Background(), option)
	return err
}

// ============ Collection 操作 ============

// ListCollections 列出所有集合
func (mc *MilvusConn) ListCollections() ([]map[string]any, error) {
	mc.CheckConnect()

	ctx := context.Background()

	res, err := mc.Service.ShowCollections(ctx, &milvuspb.ShowCollectionsRequest{DbName: mc.DbName, Type: milvuspb.ShowType_All})
	if err != nil {
		return nil, err
	}
	err = handleRespStatus(res.Status, err)
	if err != nil {
		return nil, err
	}

	colls := make([]map[string]any, len(res.CollectionIds))
	for i, id := range res.CollectionIds {
		timestamp := int64(res.CreatedUtcTimestamps[i])
		// 转换为 time.Time
		t := time.UnixMilli(timestamp)
		collectionName := res.CollectionNames[i]

		// 使用推荐的 GetLoadState API 获取加载状态
		loaded := false
		loadedPercentage := int64(0)
		loadState, err := mc.cli.GetLoadState(ctx, milvusclient.NewGetLoadStateOption(collectionName))
		if err == nil {
			// LoadStateLoaded = 已加载, LoadStateLoading = 加载中, LoadStateNotLoad = 未加载
			switch loadState.State {
			case entity.LoadStateLoaded:
				loaded = true
				loadedPercentage = 100
			case entity.LoadStateLoading:
				loaded = false
				loadedPercentage = loadState.Progress
			default:
				loaded = false
				loadedPercentage = 0
			}
		}

		m := map[string]any{
			"name":             collectionName,
			"id":               strconv.FormatInt(id, 10),
			"created_time":     t.Format(time.DateTime),
			"Loaded":           loaded,
			"LoadedPercentage": loadedPercentage,
		}
		colls[i] = m
	}
	// 按 name 升序排序
	slices.SortFunc(colls, func(a, b map[string]any) int {
		nameA := a["name"].(string)
		nameB := b["name"].(string)
		if nameA < nameB {
			return -1
		} else if nameA > nameB {
			return 1
		}
		return 0
	})

	return colls, nil
}

func handleRespStatus(status *commonpb.Status, err error) error {
	if err != nil {
		return err
	}
	if status.GetReason() != "" {
		return errorx.NewBiz(status.GetReason())
	}
	return nil
}

// CreateCollection 创建集合
func (mc *MilvusConn) CreateCollection(schema *entity.Schema, shardsNum int32) error {
	mc.CheckConnect()

	ctx := context.Background()
	option := milvusclient.NewCreateCollectionOption(schema.CollectionName, schema)
	option.WithShardNum(shardsNum)
	return mc.cli.CreateCollection(ctx, option)
}

// DropCollection 删除集合
func (mc *MilvusConn) DropCollection(collectionName string) error {
	mc.CheckConnect()

	return mc.cli.DropCollection(context.Background(), milvusclient.NewDropCollectionOption(collectionName))
}

// DescribeCollection 描述集合
func (mc *MilvusConn) DescribeCollection(collectionName string) (*entity.Collection, error) {
	mc.CheckConnect()

	coll, err := mc.cli.DescribeCollection(context.Background(), milvusclient.NewDescribeCollectionOption(collectionName))
	if err != nil {
		return nil, err
	}

	// 为每个字段查询索引信息
	if coll != nil && coll.Schema != nil && len(coll.Schema.Fields) > 0 {
		for _, field := range coll.Schema.Fields {
			if field == nil {
				continue
			}

			// 查询字段的索引
			indexes, err := mc.DescribeIndex(collectionName, field.Name)
			if err == nil {
				field.IndexParams = indexes.Params()
			}
		}
	}

	return coll, nil
}

// GetCollectionStatistics 获取集合统计信息
func (mc *MilvusConn) GetCollectionStatistics(collectionName string) (map[string]string, error) {
	mc.CheckConnect()
	return mc.cli.GetCollectionStats(context.Background(), milvusclient.NewGetCollectionStatsOption(collectionName))
}

// LoadCollection 加载集合
func (mc *MilvusConn) LoadCollection(collectionName string) error {
	mc.CheckConnect()
	option := milvusclient.NewLoadCollectionOption(collectionName)
	option.WithRefresh(true)
	task, err := mc.cli.LoadCollection(context.Background(), option)
	if err != nil {
		return err
	}
	// sync wait collection to be loaded
	err = task.Await(context.Background())
	return err
}

// ReleaseCollection 释放集合
func (mc *MilvusConn) ReleaseCollection(collectionName string) error {
	mc.CheckConnect()

	return mc.cli.ReleaseCollection(context.Background(), milvusclient.NewReleaseCollectionOption(collectionName))
}

// HasCollection 检查集合是否存在
func (mc *MilvusConn) HasCollection(collectionName string) (bool, error) {
	mc.CheckConnect()

	return mc.cli.HasCollection(context.Background(), milvusclient.NewHasCollectionOption(collectionName))
}

// RenameCollection 重命名集合
func (mc *MilvusConn) RenameCollection(oldName, newName string) error {
	mc.CheckConnect()

	return mc.cli.RenameCollection(context.Background(), milvusclient.NewRenameCollectionOption(oldName, newName))
}

// AddCollectionField 添加 Collection 字段
func (mc *MilvusConn) AddCollectionField(collectionName string, field *entity.Field) error {
	mc.CheckConnect()

	return mc.cli.AddCollectionField(context.Background(), milvusclient.NewAddCollectionFieldOption(collectionName, field))
}

// DropCollectionField 删除 Collection 字段
func (mc *MilvusConn) DropCollectionField(collectionName string, fieldName string) error {
	mc.CheckConnect()
	// 注意：Milvus 暂不支持直接删除字段
	return fmt.Errorf("Milvus 暂不支持直接删除 collection 中的字段")
}

// AlterCollectionProperty 修改 Collection 字段属性
func (mc *MilvusConn) AlterCollectionProperty(collectionName string, param *form.AlterCollectionFieldForm) error {
	mc.CheckConnect()

	option := milvusclient.NewAlterCollectionFieldPropertiesOption(collectionName, param.FieldName)
	for k, v := range param.Properties {
		option.WithProperty(k, v)
	}

	return mc.cli.AlterCollectionFieldProperty(context.Background(), option)
}

// AlterCollection 修改 Collection 属性
func (mc *MilvusConn) AlterCollection(collectionName string, param *form.AlterCollectionForm) error {
	mc.CheckConnect()
	ctx := context.Background()

	// 1. 重命名 Collection（如果提供了新名称）
	if param.NewName != "" && param.NewName != collectionName {
		if err := mc.cli.RenameCollection(ctx, milvusclient.NewRenameCollectionOption(collectionName, param.NewName)); err != nil {
			return fmt.Errorf("重命名 Collection 失败: %w", err)
		}
		// 重命名后，后续操作使用新名称
		collectionName = param.NewName
	}

	// 2. 添加新字段（如果有）
	if len(param.Fields) > 0 {
		for _, f := range param.Fields {
			field := f.ToEntityField()
			if err := mc.cli.AddCollectionField(ctx, milvusclient.NewAddCollectionFieldOption(collectionName, field)); err != nil {
				return fmt.Errorf("添加字段 %s 失败: %w", f.Name, err)
			}
		}
	}

	// 3. 修改 Collection 属性
	option := milvusclient.NewAlterCollectionPropertiesOption(collectionName)

	// 设置描述
	if param.Description != "" {
		option.WithProperty("collection.description", param.Description)
	}

	// 设置其他自定义属性
	for k, v := range param.Properties {
		valueStr := fmt.Sprintf("%v", v)
		option.WithProperty(k, valueStr)
	}

	// 如果有属性需要修改，执行修改操作
	if param.Description != "" || len(param.Properties) > 0 {
		if err := mc.cli.AlterCollectionProperties(ctx, option); err != nil {
			return fmt.Errorf("修改 Collection 属性失败: %w", err)
		}
	}

	return nil
}

// ============ 数据操作 ============

// Insert 插入数据
func (mc *MilvusConn) Insert(collectionName string, columns ...column.Column) (int64, error) {
	mc.CheckConnect()

	result, err := mc.cli.Insert(context.Background(), milvusclient.NewColumnBasedInsertOption(collectionName).WithColumns(columns...))
	if err != nil {
		return -1, err
	}

	// 返回 ID 列
	return result.InsertCount, nil
}

// Flush 刷新数据
func (mc *MilvusConn) Flush(collectionName string) error {
	mc.CheckConnect()
	task, err := mc.cli.Flush(context.Background(), milvusclient.NewFlushOption(collectionName))
	if err != nil {
		return err
	}

	return task.Await(context.Background())
}

// Delete 删除数据
func (mc *MilvusConn) Delete(collectionName string, expr string) (int64, error) {
	mc.CheckConnect()

	result, err := mc.cli.Delete(context.Background(), milvusclient.NewDeleteOption(collectionName).WithExpr(expr))
	if err != nil {
		return 0, err
	}

	return result.DeleteCount, err
}

// QueryResult 查询结果（行式存储，便于前端解析）
type QueryResult struct {
	Count    int                      `json:"count"`    // 当前页记录数
	Fields   []string                 `json:"fields"`   // 字段列表
	Data     []map[string]interface{} `json:"data"`     // 行式数据
	Page     int                      `json:"page"`     // 当前页码
	PageSize int                      `json:"pageSize"` // 每页条数
	Total    int                      `json:"total"`    // 总记录数
}

// Query 查询数据
func (mc *MilvusConn) Query(collectionName string, f *form.QueryForm) (*QueryResult, error) {
	mc.CheckConnect()

	// 处理分页参数
	page := f.Page
	if page < 1 {
		page = 1
	}
	pageSize := f.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	ctx := context.Background()

	// 第一页时查询总记录数
	totalCount := 0
	// 使用 count(*) 查询总数
	countOption := milvusclient.NewQueryOption(collectionName).
		WithOutputFields("count(*)").
		WithFilter(f.Expr)

	if f.ConsistencyLevel > 0 && f.ConsistencyLevel <= 4 {
		countOption.WithConsistencyLevel(entity.ConsistencyLevel(f.ConsistencyLevel))
	}
	if len(f.PartitionNames) > 0 {
		countOption.WithPartitions(f.PartitionNames...)
	}

	countResult, err := mc.cli.Query(ctx, countOption)
	if err != nil {
		return nil, fmt.Errorf("查询总记录数失败: %w", err)
	}

	// 解析 count(*) 结果
	if countResult.ResultCount > 0 && len(countResult.Fields) > 0 {
		for _, col := range countResult.Fields {
			if col.Name() == "count(*)" && col.Len() > 0 {
				if val, err := col.Get(0); err == nil {
					if count, ok := val.(int64); ok {
						totalCount = int(count)
					}
				}
			}
		}
	}

	// 使用 offset 和 limit 查询数据
	option := milvusclient.NewQueryOption(collectionName).
		WithOutputFields(f.OutputFields...).
		WithFilter(f.Expr).
		WithLimit(pageSize).
		WithOffset((page - 1) * pageSize)

	if f.ConsistencyLevel > 0 && f.ConsistencyLevel <= 4 {
		option.WithConsistencyLevel(entity.ConsistencyLevel(f.ConsistencyLevel))
	}
	if len(f.PartitionNames) > 0 {
		option.WithPartitions(f.PartitionNames...)
	}
	if len(f.TemplateParams) > 0 {
		for k, v := range f.TemplateParams {
			option.WithTemplateParam(k, v)
		}
	}

	result, err := mc.cli.Query(ctx, option)
	if err != nil {
		return nil, err
	}

	// 将列式数据转换为行式数据
	queryResult := &QueryResult{
		Page:     page,
		PageSize: pageSize,
		Count:    result.ResultCount,
		Total:    totalCount,
		Fields:   make([]string, 0, len(result.Fields)),
		Data:     make([]map[string]interface{}, 0, result.ResultCount),
	}

	// 提取字段名称
	for _, col := range result.Fields {
		queryResult.Fields = append(queryResult.Fields, col.Name())
	}

	// 转换为行式数据
	if result.ResultCount > 0 && len(result.Fields) > 0 {
		for i := 0; i < result.ResultCount; i++ {
			row := make(map[string]interface{})
			for _, col := range result.Fields {
				// 获取第 i 行的值
				value, err := col.Get(i)
				if err != nil {
					return nil, fmt.Errorf("获取字段 %s 的第 %d 行数据失败: %w", col.Name(), i, err)
				}
				row[col.Name()] = value
			}
			queryResult.Data = append(queryResult.Data, row)
		}
	}

	return queryResult, nil
}

// Search 向量搜索
func (mc *MilvusConn) Search(collectionName string, f *form.SearchForm) (any, error) {

	mc.CheckConnect()

	option := milvusclient.NewSearchOption(
		collectionName, // collectionName
		25,             // limit
		[]entity.Vector{entity.FloatVector(f.Vectors)},
	)

	if f.ConsistencyLevel > 0 && f.ConsistencyLevel <= 4 {
		option.WithConsistencyLevel(entity.ConsistencyLevel(f.ConsistencyLevel))
	}
	if len(f.OutputFields) > 0 {
		option.WithOutputFields(f.OutputFields...)
	}

	result, err := mc.cli.Search(context.Background(), option)
	if err != nil {
		return nil, err
	}

	return result[0], nil
}

// ============ 分区操作 ============

// CreatePartition 创建分区
func (mc *MilvusConn) CreatePartition(collectionName string, partitionName string) error {
	mc.CheckConnect()

	option := milvusclient.NewCreatePartitionOption(collectionName, partitionName)
	return mc.cli.CreatePartition(context.Background(), option)
}

// DropPartition 删除分区
func (mc *MilvusConn) DropPartition(collectionName string, partitionName string) error {
	mc.CheckConnect()

	option := milvusclient.NewDropPartitionOption(collectionName, partitionName)
	return mc.cli.DropPartition(context.Background(), option)
}

// HasPartition 检查分区是否存在
func (mc *MilvusConn) HasPartition(collectionName string, partitionName string) (bool, error) {
	mc.CheckConnect()
	option := milvusclient.NewHasPartitionOption(collectionName, partitionName)
	return mc.cli.HasPartition(context.Background(), option)
}

// ShowPartitions 显示分区列表
func (mc *MilvusConn) ShowPartitions(collectionName string) ([]vo.PartitionVO, error) {
	mc.CheckConnect()

	option := milvusclient.NewListPartitionOption(collectionName)
	resp, err := mc.Service.ShowPartitions(context.Background(), option.Request())

	var partitions []vo.PartitionVO
	if err != nil {
		return nil, err
	}
	for i, n := range resp.GetPartitionNames() {
		// 时间戳转换为日期格式
		createTime := time.UnixMilli(int64(resp.GetCreatedUtcTimestamps()[i]))
		partitions = append(partitions, vo.PartitionVO{
			Name:       n,
			Id:         resp.GetPartitionIDs()[i],
			CreateTime: timex.DefaultFormat(createTime),
		})
	}
	return partitions, nil
}

// LoadPartitions 加载分区
func (mc *MilvusConn) LoadPartitions(collectionName string, partitionNames []string) error {
	mc.CheckConnect()

	option := milvusclient.NewLoadPartitionsOption(collectionName, partitionNames...)
	option.WithRefresh(true)

	task, err := mc.cli.LoadPartitions(context.Background(), option)
	if err != nil {
		return err
	}

	return task.Await(context.Background())
}

// ReleasePartitions 释放分区
func (mc *MilvusConn) ReleasePartitions(collectionName string, partitionNames []string) error {
	mc.CheckConnect()
	option := milvusclient.NewReleasePartitionsOptions(collectionName, partitionNames...)
	return mc.cli.ReleasePartitions(context.Background(), option)
}

// ============ 索引操作 ============

// CreateIndex 创建索引
func (mc *MilvusConn) CreateIndex(collectionName string, fieldName string, idx index.Index) error {
	mc.CheckConnect()
	option := milvusclient.NewCreateIndexOption(collectionName, fieldName, idx)

	task, err := mc.cli.CreateIndex(context.Background(), option)
	if err != nil {
		return err
	}

	err = task.Await(context.Background())
	return err
}

// DescribeIndex 描述索引
func (mc *MilvusConn) DescribeIndex(collectionName string, indexName string) (index.Index, error) {
	mc.CheckConnect()

	option := milvusclient.NewDescribeIndexOption(collectionName, indexName)
	indexes, err := mc.cli.DescribeIndex(context.Background(), option)
	if err != nil {
		return nil, err
	}

	return indexes.Index, nil
}

// DropIndex 删除索引
func (mc *MilvusConn) DropIndex(collectionName string, fieldName string) error {
	mc.CheckConnect()
	return mc.cli.DropIndex(context.Background(), milvusclient.NewDropIndexOption(collectionName, fieldName))
}

// ============ 用户和权限操作 ============

// ListUsers 列出所有用户
func (mc *MilvusConn) ListUsers() ([]*vo.UserVO, error) {
	mc.CheckConnect()

	userNames, err := mc.cli.ListUsers(context.Background(), milvusclient.NewListUserOption())
	biz.ErrIsNil(err)

	var users []*vo.UserVO

	for _, user := range userNames {
		describeUser, err := mc.cli.DescribeUser(context.Background(), milvusclient.NewDescribeUserOption(user))
		biz.ErrIsNil(err)
		users = append(users, &vo.UserVO{
			Name:  user,
			Roles: describeUser.Roles,
		})
	}
	return users, nil
}

// CreateUser 创建用户
func (mc *MilvusConn) CreateUser(username string, password string) error {
	mc.CheckConnect()

	option := milvusclient.NewCreateUserOption(username, password)
	return mc.cli.CreateUser(context.Background(), option)
}

// DeleteUser 删除用户
func (mc *MilvusConn) DeleteUser(username string) error {
	mc.CheckConnect()

	option := milvusclient.NewDropUserOption(username)
	return mc.cli.DropUser(context.Background(), option)
}

// UpdatePassword 更新密码
func (mc *MilvusConn) UpdatePassword(username string, oldPassword string, newPassword string) error {
	mc.CheckConnect()

	option := milvusclient.NewUpdatePasswordOption(username, oldPassword, newPassword)
	return mc.cli.UpdatePassword(context.Background(), option)
}

// ListRoles 列出所有角色
func (mc *MilvusConn) ListRoles() ([]string, error) {
	mc.CheckConnect()

	roles, err := mc.cli.ListRoles(context.Background(), milvusclient.NewListRoleOption())
	if err != nil {
		return nil, err
	}
	return roles, nil
}

// UpdateRole 编辑角色
func (mc *MilvusConn) UpdateRole(f *form.UpdateRoleForm) error {
	mc.CheckConnect()
	// 判断是否存在此角色
	roles, err := mc.cli.ListRoles(context.Background(), milvusclient.NewListRoleOption())
	if err != nil {
		return err
	}
	// 不存在则创建角色
	if !slices.Contains(roles, f.RoleName) {
		err := mc.cli.CreateRole(context.Background(), milvusclient.NewCreateRoleOption(f.RoleName))
		if err != nil {
			return err
		}
	}

	// 编辑权限
	for dbName, dbPrivileges := range f.Privileges {
		for collectionName, collectionPrivileges := range dbPrivileges.Collections {
			for privilegeName, isGranted := range collectionPrivileges {
				var err error
				if isGranted {
					option := milvusclient.NewGrantPrivilegeV2Option(f.RoleName, privilegeName, collectionName)
					option.WithDbName(dbName)
					err = mc.cli.GrantPrivilegeV2(context.Background(), option)
				} else {
					option := milvusclient.NewRevokePrivilegeV2Option(f.RoleName, privilegeName, collectionName)
					option.WithDbName(dbName)
					err = mc.cli.RevokePrivilegeV2(context.Background(), option)
				}
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// DropRole 删除角色
func (mc *MilvusConn) DropRole(roleName string) error {
	mc.CheckConnect()
	// 删除角色所有权限
	//

	return mc.cli.DropRole(context.Background(), milvusclient.NewDropRoleOption(roleName))
}

// GrantRoleToUser 添加用户到角色
func (mc *MilvusConn) GrantRoleToUser(username string, roleName string) error {
	mc.CheckConnect()
	return mc.cli.GrantRole(context.Background(), milvusclient.NewGrantRoleOption(username, roleName))
}

// RevokeRoleFromUser 从角色移除用户
func (mc *MilvusConn) RevokeRoleFromUser(username string, roleName string) error {
	mc.CheckConnect()
	return mc.cli.RevokeRole(context.Background(), milvusclient.NewRevokeRoleOption(username, roleName))
}

// SelectRole 查询角色
func (mc *MilvusConn) SelectRole(roleName string) (*entity.Role, error) {
	mc.CheckConnect()

	option := milvusclient.NewDescribeRoleOption(roleName)
	option.WithDbName("*")

	role := &entity.Role{
		RoleName: roleName,
	}
	if roleName == "admin" {
		return role, nil
	}

	g1, err := mc.Service.SelectGrant(context.Background(), option.Request())
	biz.ErrIsNil(err)

	role.Privileges = lo.Map(g1.Entities, func(g *milvuspb.GrantEntity, _ int) entity.GrantItem {
		return entity.GrantItem{
			Object:     g.GetObject().GetName(),
			ObjectName: g.GetObjectName(),
			RoleName:   g.GetRole().GetName(),
			Grantor:    g.GetGrantor().GetUser().GetName(),
			Privilege:  g.GetGrantor().GetPrivilege().GetName(),
			DbName:     g.GetDbName(),
		}
	})

	return role, nil
}
func (mc *MilvusConn) GetPrivilegeGroup() ([]*entity.PrivilegeGroup, error) {
	return mc.cli.ListPrivilegeGroups(context.Background(), milvusclient.NewListPrivilegeGroupsOption())
}

// CreatePrivilegeGroup 创建权限组并添加初始权限
func (mc *MilvusConn) CreatePrivilegeGroup(name string, privileges []string) error {
	mc.CheckConnect()
	err := mc.cli.CreatePrivilegeGroup(context.Background(), milvusclient.NewCreatePrivilegeGroupOption(name))
	if err != nil {
		return err
	}
	if len(privileges) > 0 {
		return mc.cli.AddPrivilegesToGroup(context.Background(), milvusclient.NewAddPrivilegesToGroupOption(name, privileges...))
	}
	return nil
}

// DropPrivilegeGroup 删除权限组
func (mc *MilvusConn) DropPrivilegeGroup(name string) error {
	mc.CheckConnect()
	return mc.cli.DropPrivilegeGroup(context.Background(), milvusclient.NewDropPrivilegeGroupOption(name))
}

// UpdatePrivilegeGroup 更新权限组权限（增量计算）
func (mc *MilvusConn) UpdatePrivilegeGroup(name string, newPrivileges []string) error {
	mc.CheckConnect()
	// 获取当前权限组信息
	groups, err := mc.cli.ListPrivilegeGroups(context.Background(), milvusclient.NewListPrivilegeGroupsOption())
	if err != nil {
		return err
	}

	// 找到目标权限组的当前权限
	var currentPrivileges []string
	for _, g := range groups {
		if g.GroupName == name {
			currentPrivileges = g.Privileges
			break
		}
	}

	// 计算需要添加和移除的权限
	currentSet := make(map[string]bool)
	for _, p := range currentPrivileges {
		currentSet[p] = true
	}
	newSet := make(map[string]bool)
	for _, p := range newPrivileges {
		newSet[p] = true
	}

	var toAdd, toRemove []string
	for _, p := range newPrivileges {
		if !currentSet[p] {
			toAdd = append(toAdd, p)
		}
	}
	for _, p := range currentPrivileges {
		if !newSet[p] {
			toRemove = append(toRemove, p)
		}
	}

	if len(toAdd) > 0 {
		if err := mc.cli.AddPrivilegesToGroup(context.Background(), milvusclient.NewAddPrivilegesToGroupOption(name, toAdd...)); err != nil {
			return err
		}
	}
	if len(toRemove) > 0 {
		if err := mc.cli.RemovePrivilegesFromGroup(context.Background(), milvusclient.NewRemovePrivilegesFromGroupOption(name, toRemove...)); err != nil {
			return err
		}
	}
	return nil
}

// SelectUser 查询用户
func (mc *MilvusConn) SelectUser(username string) (*entity.User, error) {
	mc.CheckConnect()

	option := milvusclient.NewDescribeUserOption(username)
	return mc.cli.DescribeUser(context.Background(), option)
}

// ============ 资源组操作 ============

// ListResourceGroups 列出资源组（按名称升序排序）
func (mc *MilvusConn) ListResourceGroups() ([]string, error) {
	mc.CheckConnect()
	ctx := context.Background()
	option := milvusclient.NewListResourceGroupsOption()
	groups, err := mc.cli.ListResourceGroups(ctx, option)
	if err != nil {
		return nil, err
	}
	// 按名称升序排序
	slices.Sort(groups)
	return groups, nil
}

// CreateResourceGroup 创建资源组
func (mc *MilvusConn) CreateResourceGroup(name string) error {
	mc.CheckConnect()
	option := milvusclient.NewCreateResourceGroupOption(name)
	return mc.cli.CreateResourceGroup(context.Background(), option)
}

// DropResourceGroup 删除资源组
func (mc *MilvusConn) DropResourceGroup(name string) error {
	mc.CheckConnect()
	option := milvusclient.NewDropResourceGroupOption(name)
	return mc.cli.DropResourceGroup(context.Background(), option)
}

// DescribeResourceGroup 描述资源组
func (mc *MilvusConn) DescribeResourceGroup(name string) (*entity.ResourceGroup, error) {
	mc.CheckConnect()

	ctx := context.Background()
	option := milvusclient.NewDescribeResourceGroupOption(name)
	return mc.cli.DescribeResourceGroup(ctx, option)
}

// TransferReplica 转移副本
func (mc *MilvusConn) TransferReplica(collectionName string, source string, target string, numReplica int64) error {
	mc.CheckConnect()

	ctx := context.Background()
	return mc.cli.TransferReplica(ctx, milvusclient.NewTransferReplicaOption(collectionName, source, target, numReplica))
}

// ============ 别名操作 ============

// CreateAlias 创建别名
func (mc *MilvusConn) CreateAlias(collectionName string, alias string) error {
	mc.CheckConnect()

	return mc.cli.CreateAlias(context.Background(), milvusclient.NewCreateAliasOption(collectionName, alias))
}

// DropAlias 删除别名
func (mc *MilvusConn) DropAlias(alias string) error {
	mc.CheckConnect()

	return mc.cli.DropAlias(context.Background(), milvusclient.NewDropAliasOption(alias))
}

// AlterAlias 修改别名
func (mc *MilvusConn) AlterAlias(collectionName string, alias string) error {
	mc.CheckConnect()

	return mc.cli.AlterAlias(context.Background(), milvusclient.NewAlterAliasOption(collectionName, alias))
}

// ListAliases 列出集合的别名
func (mc *MilvusConn) ListAliases(collectionName string) ([]string, error) {
	mc.CheckConnect()

	return mc.cli.ListAliases(context.Background(), milvusclient.NewListAliasesOption(collectionName))
}
