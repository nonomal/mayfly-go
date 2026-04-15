package dbtool

import (
	"context"

	"mayfly-go/internal/ai/tools"
	"mayfly-go/internal/db/application"
	"mayfly-go/internal/db/dbm/dbi"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type QueryDataParam struct {
	DbId   uint64 `json:"dbId" jsonschema_description:"数据库ID"`
	DbName string `json:"dbName" jsonschema_description:"数据库名称"`
	SQL    string `json:"sql" jsonschema_description:"SQL语句"`
}

type QueryDataOutput struct {
	Columns []*dbi.QueryColumn `json:"columns" jsonschema_description:"查询结果列信息"`
	Rows    []map[string]any   `json:"rows" jsonschema_description:"查询结果数据"`
}

func GetQueryData() (tool.InvokableTool, error) {
	return utils.InferTool("DbQueryData",
		"【数据库】SQL查询 - 执行只读类 SQL 语句（如 SELECT、SHOW、DESC、EXPLAIN 等）。适用于查询表数据、分析执行计划等场景。注意：仅限查询操作，禁止执行 INSERT、UPDATE、DELETE 等变更类 SQL。",
		func(ctx context.Context, param *QueryDataParam) (*QueryDataOutput, error) {
			conn, err := application.GetDbApp().GetDbConn(ctx, param.DbId, param.DbName)
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
