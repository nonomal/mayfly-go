package dbtool

import (
	"context"

	"mayfly-go/internal/ai/tools"
	"mayfly-go/internal/db/application"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

type SqlExecParam struct {
	DbId   uint64 `json:"dbId" jsonschema_description:"数据库ID"`
	DbName string `json:"dbName" jsonschema_description:"数据库名称"`
	SQL    string `json:"sql" jsonschema_description:"SQL语句"`
}

type SqlExecOutput struct {
	Effected int64 `json:"effected" jsonschema_description:"影响的行数"`
}

func GetSqlExec() (tool.InvokableTool, error) {
	tool, err := utils.InferTool("ExecSql",
		"【数据库】SQL执行 - 执行非查询类 SQL 语句（如 INSERT、UPDATE、DELETE 等）。适用于执行数据变更操作的场景。注意：仅限变更操作，禁止执行 SELECT、SHOW、DESC、EXPLAIN 等查询类 SQL。",
		func(ctx context.Context, param *SqlExecParam) (*SqlExecOutput, error) {
			conn, err := application.GetDbApp().GetDbConn(ctx, param.DbId, param.DbName)
			if err != nil {
				return nil, tools.NewToolError(err, tools.RecoverRetry)
			}

			res, err := conn.ExecContext(ctx, param.SQL)
			if err != nil {
				return nil, tools.NewToolError(err, tools.RecoverRetry)
			}

			return &SqlExecOutput{Effected: res}, nil
		},
	)

	if err != nil {
		return nil, err
	}

	return tools.InvokableApprovableTool{
		InvokableTool: tool,
	}, nil
}
