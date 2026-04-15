package clickhouse

import (
	"fmt"
	"mayfly-go/internal/db/dbm/dbi"
	"strings"
)

type ClickHouseSqlGenerator struct {
	Dialect dbi.Dialect
}

func NewClickHouseSqlGenerator(dialect dbi.Dialect) *ClickHouseSqlGenerator {
	return &ClickHouseSqlGenerator{Dialect: dialect}
}

// GenerateInsertSql generates INSERT SQL for ClickHouse
func (sg *ClickHouseSqlGenerator) GenerateInsertSql(tableName string, columns []dbi.Column, values [][]any) []string {
	if len(values) == 0 {
		return []string{}
	}

	// For ClickHouse, we can use batch inserts for better performance
	return []string{sg.generateBatchInsertSql(tableName, columns, values)}
}

func (sg *ClickHouseSqlGenerator) generateBatchInsertSql(tableName string, columns []dbi.Column, values [][]any) string {
	var sb strings.Builder

	// Build column list
	quoter := sg.Dialect.Quoter()
	sb.WriteString(fmt.Sprintf("INSERT INTO %s (", quoter.Quote(tableName)))
	for i, col := range columns {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(quoter.Quote(col.ColumnName))
	}
	sb.WriteString(") VALUES ")

	// Build values
	for i, row := range values {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString("(")
		for j, val := range row {
			if j > 0 {
				sb.WriteString(", ")
			}
			// Handle different data types
			sb.WriteString(sg.formatValue(val, columns[j].DataType))
		}
		sb.WriteString(")")
	}

	return sb.String()
}

// formatValue formats a value for SQL insertion based on its type
func (sg *ClickHouseSqlGenerator) formatValue(value any, dataType string) string {
	if value == nil {
		return "NULL"
	}

	switch v := value.(type) {
	case string:
		// Escape single quotes
		escaped := strings.ReplaceAll(v, "'", "''")
		return fmt.Sprintf("'%s'", escaped)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%v", v)
	case float32, float64:
		return fmt.Sprintf("%v", v)
	case bool:
		if v {
			return "1"
		}
		return "0"
	default:
		// For other types, convert to string and escape
		str := fmt.Sprintf("%v", v)
		escaped := strings.ReplaceAll(str, "'", "''")
		return fmt.Sprintf("'%s'", escaped)
	}
}

// GenerateUpdateSql generates UPDATE SQL for ClickHouse
// Note: ClickHouse uses ALTER TABLE ... UPDATE syntax
func (sg *ClickHouseSqlGenerator) GenerateUpdateSql(tableName string, columns []dbi.Column, values []any, condition string) string {
	var sb strings.Builder
	quoter := sg.Dialect.Quoter()
	sb.WriteString(fmt.Sprintf("ALTER TABLE %s UPDATE ", quoter.Quote(tableName)))

	for i, col := range columns {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%s = %s", quoter.Quote(col.ColumnName), sg.formatValue(values[i], col.DataType)))
	}

	if condition != "" {
		sb.WriteString(fmt.Sprintf(" WHERE %s", condition))
	}

	return sb.String()
}

// GenerateDeleteSql generates DELETE SQL for ClickHouse
// Note: ClickHouse uses ALTER TABLE ... DELETE syntax
func (sg *ClickHouseSqlGenerator) GenerateDeleteSql(tableName string, condition string) string {
	if condition == "" {
		// To delete all rows, we need a condition that matches all rows
		condition = "1 = 1"
	}
	quoter := sg.Dialect.Quoter()
	return fmt.Sprintf("ALTER TABLE %s DELETE WHERE %s", quoter.Quote(tableName), condition)
}

// GenerateCreateTableSql generates CREATE TABLE SQL for ClickHouse
func (sg *ClickHouseSqlGenerator) GenerateCreateTableSql(tableName string, columns []dbi.Column, indexes []dbi.Index) string {
	var sb strings.Builder
	quoter := sg.Dialect.Quoter()
	sb.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", quoter.Quote(tableName)))

	for i, col := range columns {
		if i > 0 {
			sb.WriteString(",\n")
		}
		sb.WriteString(fmt.Sprintf("  %s %s", quoter.Quote(col.ColumnName), col.DataType))

		if col.ColumnComment != "" {
			sb.WriteString(fmt.Sprintf(" COMMENT '%s'", strings.ReplaceAll(col.ColumnComment, "'", "''")))
		}
	}

	sb.WriteString("\n) ENGINE = MergeTree() ORDER BY tuple()")

	if len(columns) > 0 {
		// Add primary key if any column is marked as primary key
		var primaryKeys []string
		for _, col := range columns {
			if col.IsPrimaryKey {
				primaryKeys = append(primaryKeys, quoter.Quote(col.ColumnName))
			}
		}
		if len(primaryKeys) > 0 {
			sb.WriteString(fmt.Sprintf(" PRIMARY KEY (%s)", strings.Join(primaryKeys, ", ")))
		}
	}

	return sb.String()
}

// GenerateDropTableSql generates DROP TABLE SQL for ClickHouse
func (sg *ClickHouseSqlGenerator) GenerateDropTableSql(tableName string) string {
	quoter := sg.Dialect.Quoter()
	return fmt.Sprintf("DROP TABLE IF EXISTS %s", quoter.Quote(tableName))
}

// GenerateCreateDatabaseSql generates CREATE DATABASE SQL for ClickHouse
func (sg *ClickHouseSqlGenerator) GenerateCreateDatabaseSql(databaseName string) string {
	quoter := sg.Dialect.Quoter()
	return fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", quoter.Quote(databaseName))
}

// GenerateDropDatabaseSql generates DROP DATABASE SQL for ClickHouse
func (sg *ClickHouseSqlGenerator) GenerateDropDatabaseSql(databaseName string) string {
	quoter := sg.Dialect.Quoter()
	return fmt.Sprintf("DROP DATABASE IF EXISTS %s", quoter.Quote(databaseName))
}
