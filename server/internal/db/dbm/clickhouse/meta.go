package clickhouse

import (
	"context"
	"database/sql"
	"fmt"
	"mayfly-go/internal/db/dbm/dbi"
	"mayfly-go/pkg/utils/collx"

	_ "github.com/ClickHouse/clickhouse-go/v2"
)

func init() {
	meta := new(Meta)
	dbi.Register(DbTypeClickHouse, meta)
}

const (
	DbTypeClickHouse dbi.DbType = "clickhouse"
)

type Meta struct {
}

func (cm *Meta) GetSqlDb(ctx context.Context, d *dbi.DbInfo) (*sql.DB, error) {
	// ClickHouse connection string
	// Format: clickhouse://username:password@host:port/database?param1=value1&param2=value2
	dbName := d.GetDatabase()
	if dbName == "" {
		dbName = "default"
	}

	// Build connection string
	dsn := fmt.Sprintf("clickhouse://%s:%s@%s:%d/%s", d.Username, d.Password, d.Host, d.Port, dbName)
	if d.Params != "" {
		dsn = fmt.Sprintf("%s?%s", dsn, d.Params)
	} else {
		// Add default parameters for better connection handling
		dsn = fmt.Sprintf("%s?dial_timeout=10s&read_timeout=30s", dsn)
	}

	const driverName = "clickhouse"
	return sql.Open(driverName, dsn)
}

func (cm *Meta) GetDialect(conn *dbi.DbConn) dbi.Dialect {
	return &ClickHouseDialect{dc: conn}
}

func (cm *Meta) GetMetadata(conn *dbi.DbConn) dbi.Metadata {
	return &ClickHouseMetadata{dc: conn}
}

func (cm *Meta) GetDbDataTypes() []*dbi.DbDataType {
	return collx.AsArray(
		UInt8, UInt16, UInt32, UInt64, Int8, Int16, Int32, Int64,
		Float32, Float64,
		String, FixedString,
		DateTime, DateTime64, Date, Date32,
		UUID,
		IPv4, IPv6,
		Bool,
		Decimal, Decimal32, Decimal64, Decimal128, Decimal256,
		Enum8, Enum16,
		Array, Tuple, Map, Nested, AggregateFunction, SimpleAggregateFunction,
		LowCardinality, Nullable,
	)
}

func (cm *Meta) GetCommonTypeConverter() dbi.CommonTypeConverter {
	return &commonTypeConverter{}
}

// Common type converter for ClickHouse
type commonTypeConverter struct{}

func (c *commonTypeConverter) Varchar(column *dbi.Column) *dbi.DbDataType {
	return String
}

func (c *commonTypeConverter) Char(column *dbi.Column) *dbi.DbDataType {
	return FixedString
}

func (c *commonTypeConverter) Text(column *dbi.Column) *dbi.DbDataType {
	return String
}

func (c *commonTypeConverter) Int1(column *dbi.Column) *dbi.DbDataType {
	return Int8
}

func (c *commonTypeConverter) Int2(column *dbi.Column) *dbi.DbDataType {
	return Int16
}

func (c *commonTypeConverter) Int4(column *dbi.Column) *dbi.DbDataType {
	return Int32
}

func (c *commonTypeConverter) Int8(column *dbi.Column) *dbi.DbDataType {
	return Int64
}

func (c *commonTypeConverter) Decimal(column *dbi.Column) *dbi.DbDataType {
	return Decimal
}

func (c *commonTypeConverter) UnsignedInt8(column *dbi.Column) *dbi.DbDataType {
	return UInt8
}

func (c *commonTypeConverter) UnsignedInt4(column *dbi.Column) *dbi.DbDataType {
	return UInt32
}

func (c *commonTypeConverter) UnsignedInt2(column *dbi.Column) *dbi.DbDataType {
	return UInt16
}

func (c *commonTypeConverter) UnsignedInt1(column *dbi.Column) *dbi.DbDataType {
	return UInt8
}

func (c *commonTypeConverter) Date(column *dbi.Column) *dbi.DbDataType {
	return Date
}

func (c *commonTypeConverter) Time(column *dbi.Column) *dbi.DbDataType {
	return DateTime
}

func (c *commonTypeConverter) Datetime(column *dbi.Column) *dbi.DbDataType {
	return DateTime
}

func (c *commonTypeConverter) Timestamp(column *dbi.Column) *dbi.DbDataType {
	return DateTime
}

func (c *commonTypeConverter) Binary(column *dbi.Column) *dbi.DbDataType {
	return String
}

func (c *commonTypeConverter) Mediumtext(column *dbi.Column) *dbi.DbDataType {
	return String
}

func (c *commonTypeConverter) Longtext(column *dbi.Column) *dbi.DbDataType {
	return String
}

func (c *commonTypeConverter) Bit(column *dbi.Column) *dbi.DbDataType {
	return UInt8
}

func (c *commonTypeConverter) Bool(column *dbi.Column) *dbi.DbDataType {
	return Bool
}

func (c *commonTypeConverter) Numeric(column *dbi.Column) *dbi.DbDataType {
	return Float64
}

func (c *commonTypeConverter) Enum(column *dbi.Column) *dbi.DbDataType {
	return Enum8
}

func (c *commonTypeConverter) JSON(column *dbi.Column) *dbi.DbDataType {
	return String
}

func (c *commonTypeConverter) Blob(column *dbi.Column) *dbi.DbDataType {
	return String
}

func (c *commonTypeConverter) Mediumblob(column *dbi.Column) *dbi.DbDataType {
	return String
}

func (c *commonTypeConverter) Longblob(column *dbi.Column) *dbi.DbDataType {
	return String
}

func (c *commonTypeConverter) Varbinary(column *dbi.Column) *dbi.DbDataType {
	return String
}
