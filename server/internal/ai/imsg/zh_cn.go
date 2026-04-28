package imsg

import "mayfly-go/pkg/i18n"

var Zh_CN = map[i18n.MsgId]string{
	InfoIncomplete:           "信息不全，请完善",
	ParamCompletionTitle:     "参数补全",
	ApprovalTitle:            "高危操作审批",
	ApprovalDesc:             "该操作需要审批后才能执行",
	RejectReasonDefault:      "用户未提供具体原因",
	MissingRequiredParams:    "缺少必要参数",
	SqlExecApprovalReason:    "执行SQL是高危操作，请审批",
	ExecSqlToolDesc:          "ExecSql【数据库SQL执行】",
	ExecSqlToolInfo:          "【数据库】SQL执行 - 执行非查询类 SQL 语句（如 INSERT、UPDATE、DELETE 等）。适用于执行数据变更操作的场景。注意：仅限变更操作，禁止执行 SELECT、SHOW、DESC、EXPLAIN 等查询类 SQL。",
	DbQueryDataToolDesc:      "DbQueryData【数据库SQL查询】",
	DbQueryDataToolInfo:      "【数据库】SQL查询 - 执行只读类 SQL 语句（如 SELECT、SHOW、DESC、EXPLAIN 等）。适用于查询表数据、分析执行计划等场景。注意：仅限查询操作，禁止执行 INSERT、UPDATE、DELETE 等变更类 SQL。\n\n【重要】如果用户没有指定数据库ID(dbId)，请直接调用此工具并只提供SQL语句，系统会自动弹出资产选择界面让用户选择数据库。不要询问用户数据库ID！",
	DbQueryTableInfoToolDesc: "DbQueryTableInfo【查询数据库表信息】",
	DbQueryTableInfoToolInfo: "【数据库】表结构查询工具 - 获取指定数据表的 DDL 定义，包含字段名、数据类型、约束、索引等完整元数据。适用于编写 SQL 前了解表结构、排查数据问题时查看表定义等场景。\n\n【重要】如果用户没有指定数据库ID(dbId)，请直接调用此工具并只提供表名，系统会自动弹出资产选择界面让用户选择数据库。不要询问用户数据库ID！",
	DbInfoIncomplete:         "缺少数据库信息，请完善参数",
}
