package dbtool

import (
	"mayfly-go/internal/ai/tools"
	"mayfly-go/pkg/logx"
)

func Init() {
	if queryTableTool, err := GetQueryTableInfo(); err != nil {
		logx.Errorf("agent tool - 获取QueryTableInfo工具失败: %v", err)
	} else {
		tools.DefaultRegistry.Register(queryTableTool)
	}

	if queryDataTool, err := GetQueryData(); err != nil {
		logx.Errorf("agent tool - 获取QueryData工具失败: %v", err)
	} else {
		tools.DefaultRegistry.Register(queryDataTool)
	}

	if sqlExecTool, err := GetSqlExec(); err != nil {
		logx.Errorf("agent tool - 获取ExecSql工具失败: %v", err)
	} else {
		tools.DefaultRegistry.Register(sqlExecTool)
	}
}
