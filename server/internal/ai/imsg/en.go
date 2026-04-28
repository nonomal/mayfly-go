package imsg

import "mayfly-go/pkg/i18n"

var En = map[i18n.MsgId]string{
	InfoIncomplete:           "Information incomplete, please complete",
	ParamCompletionTitle:     "Parameter Completion",
	ApprovalTitle:            "High-risk Operation Approval",
	ApprovalDesc:             "This operation requires approval before execution",
	RejectReasonDefault:      "User did not provide a specific reason",
	MissingRequiredParams:    "Missing required parameters",
	SqlExecApprovalReason:    "Executing SQL is a high-risk operation, please approve",
	ExecSqlToolDesc:          "ExecSql【Database SQL Execution】",
	ExecSqlToolInfo:          "[Database] SQL Execution - Execute non-query SQL statements (such as INSERT, UPDATE, DELETE, etc.). Applicable to data modification scenarios. Note: modification only, prohibit executing SELECT, SHOW, DESC, EXPLAIN and other query SQL.",
	DbQueryDataToolDesc:      "DbQueryData【Database SQL Query】",
	DbQueryDataToolInfo:      "[Database] SQL Query - Execute read-only SQL statements (such as SELECT, SHOW, DESC, EXPLAIN, etc.). Applicable to querying table data, analyzing execution plans, etc. Note: query only, prohibit executing INSERT, UPDATE, DELETE and other modification SQL.\n\n[Important] If the user does not specify a database ID (dbId), please call this tool directly and only provide the SQL statement. The system will automatically pop up the asset selection interface for the user to choose the database. Do not ask the user for the database ID!",
	DbQueryTableInfoToolDesc: "DbQueryTableInfo【Query Database Table Info】",
	DbQueryTableInfoToolInfo: "[Database] Table Structure Query Tool - Get the DDL definition of a specified table, including complete metadata such as field names, data types, constraints, and indexes. Applicable to understanding table structure before writing SQL, or checking table definitions when troubleshooting data issues.\n\n[Important] If the user does not specify a database ID (dbId), please call this tool directly and only provide the table name. The system will automatically pop up the asset selection interface for the user to choose the database. Do not ask the user for the database ID!",
	DbInfoIncomplete:         "Missing database information, please complete the parameters",
}
