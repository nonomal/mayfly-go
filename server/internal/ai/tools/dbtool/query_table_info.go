package dbtool

import (
	"context"

	"mayfly-go/internal/ai/tools"
	"mayfly-go/internal/db/application"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type QueryTableInfoParam struct {
	DbId      uint64 `json:"dbId" jsonschema_description:"数据库ID"`
	DbName    string `json:"dbName" jsonschema_description:"数据库名称"`
	TableName string `json:"tableName" jsonschema_description:"表名"`
}

type QueryTableInfoOutput struct {
	DDL string `json:"ddl" jsonschema_description:"表DDL"`
}

func GetQueryTableInfo() (tool.InvokableTool, error) {
	return utils.InferTool("DbQueryTableInfo",
		"【数据库】表结构查询工具 - 获取指定数据表的 DDL 定义，包含字段名、数据类型、约束、索引等完整元数据。适用于编写 SQL 前了解表结构、排查数据问题时查看表定义等场景。",
		func(ctx context.Context, param *QueryTableInfoParam) (*QueryTableInfoOutput, error) {
			conn, err := application.GetDbApp().GetDbConn(ctx, param.DbId, param.DbName)
			if err != nil {
				return nil, tools.NewToolError(err, tools.RecoverRetry)
			}

			ddl, err := conn.GetMetadata().GetTableDDL(param.TableName, false)
			if err != nil {
				return nil, tools.NewToolError(err, tools.RecoverRetry)
			}
			output := &QueryTableInfoOutput{DDL: ddl}
			return output, nil
		},
	)
}
