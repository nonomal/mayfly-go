package dbtool

import (
	"context"
	"fmt"

	"mayfly-go/internal/ai/imsg"
	"mayfly-go/internal/ai/tools"
	"mayfly-go/internal/db/application"
	"mayfly-go/internal/db/dbm/dbi"
	"mayfly-go/pkg/i18n"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type QueryDataParam struct {
	DbId   int64  `json:"dbId" jsonschema_description:"数据库ID。如果用户未明确提供，请传0，不要猜测！"`
	DbName string `json:"dbName" jsonschema_description:"数据库名称。如果用户未明确提供，请留空，不要猜测！"`
	SQL    string `json:"sql" jsonschema_description:"SQL语句" jsonschema:"required" `
}

type QueryDataOutput struct {
	Columns []*dbi.QueryColumn `json:"columns" jsonschema_description:"查询结果列信息"`
	Rows    []map[string]any   `json:"rows" jsonschema_description:"查询结果数据"`
}

func GetQueryData() (tool.InvokableTool, error) {
	return utils.InferTool("DbQueryData",
		i18n.T(imsg.DbQueryDataToolInfo),
		func(ctx context.Context, param *QueryDataParam) (*QueryDataOutput, error) {
			toolDesc := i18n.TC(ctx, imsg.DbQueryDataToolDesc)
			// 检查必要参数，触发参数完善
			if param.DbId == 0 || param.DbName == "" {
				if err := tools.InterruptOrResumeParamCompletion(ctx, toolDesc, param, i18n.TC(ctx, imsg.DbInfoIncomplete), "db", []tools.CompletionParamInfo{
					{Param: "dbId", Name: "数据库ID", Cacheable: true},
					{Param: "dbName", Name: "数据库名称", Cacheable: true},
				}); err != nil {
					return nil, err
				}
			}

			if param.SQL == "" {
				return nil, fmt.Errorf("sql parameter is required")
			}

			conn, err := application.GetDbApp().GetDbConn(ctx, uint64(param.DbId), param.DbName)
			if err != nil {
				return nil, tools.NewToolError(err, tools.RecoverRetry)
			}

			rows := make([]map[string]any, 0)
			columns, err := conn.WalkQueryRows(ctx, param.SQL, func(row map[string]any, columns []*dbi.QueryColumn) error {
				rows = append(rows, row)
				if len(rows) > 1000 {
					return dbi.NewStopWalkQueryError("The maximum number of query rows is exceeded: 1000")
				}
				return nil
			})
			if err != nil {
				return nil, tools.NewToolError(err, tools.RecoverRetry)
			}

			output := &QueryDataOutput{Columns: columns, Rows: rows}
			return output, nil
		},
	)
}
