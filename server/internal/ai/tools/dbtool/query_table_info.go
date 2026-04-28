package dbtool

import (
	"context"

	"mayfly-go/internal/ai/imsg"
	"mayfly-go/internal/ai/tools"
	"mayfly-go/internal/db/application"
	"mayfly-go/pkg/i18n"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type QueryTableInfoParam struct {
	DbId      int64  `json:"dbId" jsonschema_description:"数据库ID。如果用户未明确提供，请传0，不要猜测！"`
	DbName    string `json:"dbName" jsonschema_description:"数据库名称。如果用户未明确提供，请留空，不要猜测！"`
	TableName string `json:"tableName" jsonschema_description:"表名" jsonschema:"required" `
}

type QueryTableInfoOutput struct {
	DDL string `json:"ddl" jsonschema_description:"表DDL"`
}

func GetQueryTableInfo() (tool.InvokableTool, error) {
	return utils.InferTool("DbQueryTableInfo",
		i18n.T(imsg.DbQueryTableInfoToolInfo),
		func(ctx context.Context, param *QueryTableInfoParam) (*QueryTableInfoOutput, error) {
			toolDesc := i18n.TC(ctx, imsg.DbQueryTableInfoToolDesc)
			// 检查必要参数，触发参数完善
			if param.DbId == 0 || param.DbName == "" {
				if err := tools.InterruptOrResumeParamCompletion(ctx, toolDesc, param, i18n.TC(ctx, imsg.DbInfoIncomplete), "db", []tools.CompletionParamInfo{
					{Param: "dbId", Name: "数据库ID", Cacheable: true},
					{Param: "dbName", Name: "数据库名称", Cacheable: true},
				}); err != nil {
					return nil, err
				}
			}

			conn, err := application.GetDbApp().GetDbConn(ctx, uint64(param.DbId), param.DbName)
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
