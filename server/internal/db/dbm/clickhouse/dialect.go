package clickhouse

import (
	"errors"
	"fmt"
	"mayfly-go/internal/db/dbm/dbi"
	"mayfly-go/internal/db/dbm/sqlparser"
	"mayfly-go/internal/db/dbm/sqlparser/pgsql"
	"strings"
)

type ClickHouseDialect struct {
	dc *dbi.DbConn
}

func (cd *ClickHouseDialect) Quoter() dbi.Quoter {
	return dbi.Quoter{
		Prefix:     '`',
		Suffix:     '`',
		IsReserved: dbi.AlwaysReserve,
	}
}

func (cd *ClickHouseDialect) GetDbProgram() (dbi.DbProgram, error) {
	return nil, errors.New("not support db program")
}

func (cd *ClickHouseDialect) GetDumpHelper() dbi.DumpHelper {
	return new(dbi.DefaultDumpHelper)
}

func (cd *ClickHouseDialect) GetSQLParser() sqlparser.SqlParser {
	return new(pgsql.PgsqlParser)
}

func (cd *ClickHouseDialect) CopyTable(copy *dbi.DbCopyTable) error {
	// ClickHouse doesn't support traditional table copying
	// This would need to be implemented with CREATE TABLE ... AS SELECT
	return errors.New("not implemented")
}

func (cd *ClickHouseDialect) GetSQLGenerator() dbi.SQLGenerator {
	return &ClickHouseSQLGenerator{dialect: cd}
}

// ClickHouseSQLGenerator implements the SQLGenerator interface for ClickHouse
type ClickHouseSQLGenerator struct {
	dialect *ClickHouseDialect
}

func (csg *ClickHouseSQLGenerator) GenTableDDL(table dbi.Table, columns []dbi.Column, dropBeforeCreate bool) []string {
	var sqls []string

	if dropBeforeCreate {
		sqls = append(sqls, fmt.Sprintf("DROP TABLE IF EXISTS %s", csg.dialect.Quoter().Quote(table.TableName)))
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("CREATE TABLE %s (\n", csg.dialect.Quoter().Quote(table.TableName)))

	for i, col := range columns {
		if i > 0 {
			sb.WriteString(",\n")
		}
		sb.WriteString(fmt.Sprintf("  %s %s", csg.dialect.Quoter().Quote(col.ColumnName), col.DataType))

		if col.ColumnComment != "" {
			sb.WriteString(fmt.Sprintf(" COMMENT '%s'", strings.ReplaceAll(col.ColumnComment, "'", "''")))
		}
	}

	sb.WriteString("\n) ENGINE = MergeTree() ORDER BY tuple()")

	if table.TableComment != "" {
		sb.WriteString(fmt.Sprintf(" COMMENT '%s'", strings.ReplaceAll(table.TableComment, "'", "''")))
	}

	sqls = append(sqls, sb.String())
	return sqls
}

func (csg *ClickHouseSQLGenerator) GenIndexDDL(table dbi.Table, indexs []dbi.Index) []string {
	// ClickHouse indexes are typically defined in the CREATE TABLE statement
	// This is a simplified implementation
	return []string{}
}

func (csg *ClickHouseSQLGenerator) GenInsert(tableName string, columns []dbi.Column, values [][]any, duplicateStrategy int) []string {
	if len(values) == 0 {
		return []string{}
	}

	quote := csg.dialect.Quoter().Quote

	// Build column list
	var columnNames []string
	var columnTypes []*dbi.DbDataType

	for _, column := range columns {
		columnNames = append(columnNames, quote(column.ColumnName))
		columnType := dbi.GetDbDataType(DbTypeClickHouse, column.DataType)
		columnTypes = append(columnTypes, columnType)
	}

	// Build values
	var valueRows []string
	for _, row := range values {
		var rowValues []string
		for i, value := range row {
			rowValues = append(rowValues, columnTypes[i].DataType.SQLValue(value))
		}
		valueRows = append(valueRows, fmt.Sprintf("(%s)", strings.Join(rowValues, ", ")))
	}

	// 处理Clickhouse的重复策略
	switch duplicateStrategy {
	case dbi.DuplicateStrategyNone:
		// 对于DuplicateStrategyNone，直接插入数据，不处理重复
		sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
			quote(tableName),
			strings.Join(columnNames, ", "),
			strings.Join(valueRows, ", "))

		return []string{sql}
	case dbi.DuplicateStrategyIgnore:
		// 对于DuplicateStrategyIgnore，使用INSERT IGNORE语法
		sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
			quote(tableName),
			strings.Join(columnNames, ", "),
			strings.Join(valueRows, ", "))

		return []string{sql}
	case dbi.DuplicateStrategyUpdate:
		// 对于DuplicateStrategyIgnore和DuplicateStrategyUpdate，先删除重复数据，再插入新数据
		keyColumn := columnNames[0]

		// 构建删除重复数据的SQL
		var deleteSqls []string
		var keyValues []string

		// 提取主键值
		for _, row := range values {
			if len(row) > 0 {
				// 获取主键列的值
				keyValue := columnTypes[0].DataType.SQLValue(row[0])
				keyValues = append(keyValues, keyValue)
			}
		}

		// 如果有主键值，构建删除语句
		if len(keyValues) > 0 {
			// 将主键值用逗号连接
			keyValueList := strings.Join(keyValues, ", ")
			deleteSql := fmt.Sprintf("ALTER TABLE %s DELETE WHERE %s IN (%s)",
				quote(tableName),
				keyColumn,
				keyValueList)
			deleteSqls = append(deleteSqls, deleteSql)
		}

		// 构建插入数据的SQL
		sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
			quote(tableName),
			strings.Join(columnNames, ", "),
			strings.Join(valueRows, ", "))

		// 返回删除和插入的SQL
		result := append(deleteSqls, sql)
		return result
	default:
		// 默认情况下，直接插入数据
		sql := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
			quote(tableName),
			strings.Join(columnNames, ", "),
			strings.Join(valueRows, ", "))

		return []string{sql}
	}
}
