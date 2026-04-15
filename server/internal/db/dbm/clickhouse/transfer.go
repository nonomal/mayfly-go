package clickhouse

import (
	"mayfly-go/internal/db/dbm/dbi"
)

// ClickHousetransfer处理Clickhouse的数据传输操作
type ClickHouseTransfer struct {
	dc *dbi.DbConn
}

// newClickhousetransfer创建了一个新的ClickHousetransfer实例
func NewClickHouseTransfer(conn *dbi.DbConn) *ClickHouseTransfer {
	return &ClickHouseTransfer{dc: conn}
}

// getinsertsql生成插入SQL，用于将数据传输到Clickhouse
func (ct *ClickHouseTransfer) GetInsertSql(tableName string, columns []dbi.Column, values [][]any) []string {
	generator := NewClickHouseSqlGenerator(ct.dc.GetDialect())
	return generator.GenerateInsertSql(tableName, columns, values)
}

// getBatchInsertsql生成批处理插入SQL以提高性能
func (ct *ClickHouseTransfer) GetBatchInsertSql(tableName string, columns []dbi.Column, values [][]any) string {
	generator := NewClickHouseSqlGenerator(ct.dc.GetDialect())
	return generator.generateBatchInsertSql(tableName, columns, values)
}

// ProcessColumns processes columns for ClickHouse compatibility
func (ct *ClickHouseTransfer) ProcessColumns(columns []dbi.Column) []dbi.Column {
	processed := make([]dbi.Column, len(columns))

	for i, col := range columns {
		processed[i] = col

		// 将数据类型转换为Clickhouse兼容类型
		processed[i].DataType = ct.convertDataType(col.DataType)

		// 处理无效类型
		if col.Nullable {
			processed[i].DataType = "Nullable(" + processed[i].DataType + ")"
		}
	}

	return processed
}

// 将源数据库数据类型转换为clickhouse数据类型
func (ct *ClickHouseTransfer) convertDataType(sourceType string) string {
	// 这是一个简化的映射。实际上，这将更加全面
	// 并且可能需要根据源数据库类型进行调整

	switch sourceType {
	case "VARCHAR", "CHAR", "TEXT", "MEDIUMTEXT", "LONGTEXT":
		return "String"
	case "INT", "INTEGER", "MEDIUMINT":
		return "Int32"
	case "BIGINT":
		return "Int64"
	case "SMALLINT":
		return "Int16"
	case "TINYINT":
		return "Int8"
	case "FLOAT":
		return "Float32"
	case "DOUBLE", "DECIMAL", "NUMERIC":
		return "Float64"
	case "DATE":
		return "Date"
	case "DATETIME", "TIMESTAMP":
		return "DateTime"
	case "BOOLEAN", "BOOL":
		return "Bool"
	case "BLOB", "BINARY", "VARBINARY":
		return "String"
	default:
		// For unknown types, default to String
		return "String"
	}
}

// GetTableOptions返回数据传输的Clickhouse特定表选项
func (ct *ClickHouseTransfer) GetTableOptions() map[string]string {
	return map[string]string{
		"engine":   "MergeTree()",
		"order_by": "tuple()",
	}
}

// PreTransfer prepares the target ClickHouse database for data transfer
func (ct *ClickHouseTransfer) PreTransfer(tableName string) error {
	// In ClickHouse, we might want to drop the table if it exists before creating it
	// This is optional and depends on the transfer strategy

	// For now, we'll just return nil as no specific preparation is needed
	return nil
}

// PostTransfer performs any cleanup or optimization after data transfer
func (ct *ClickHouseTransfer) PostTransfer(tableName string) error {
	// ClickHouse might benefit from optimization after bulk inserts
	// For example, we might want to optimize the table

	_, err := ct.dc.Exec("OPTIMIZE TABLE " + ct.dc.GetDialect().Quoter().Quote(tableName) + " FINAL")
	return err
}

// GetDuplicateStrategySupport checks if ClickHouse supports duplicate handling strategies
func (ct *ClickHouseTransfer) GetDuplicateStrategySupport() bool {
	// ClickHouse has limited support for duplicate handling
	// It depends on the table engine being used
	return true
}

// GetBatchSize returns the recommended batch size for ClickHouse inserts
func (ct *ClickHouseTransfer) GetBatchSize() int {
	// ClickHouse benefits from larger batch sizes for better performance
	return 10000
}
