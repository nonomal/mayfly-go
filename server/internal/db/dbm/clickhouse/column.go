package clickhouse

import (
	"mayfly-go/internal/db/dbm/dbi"
)

// ClickHouse data types
var (
	// Numeric types
	UInt8  = dbi.NewDbDataType("UInt8", dbi.DTInt8).WithCT(dbi.CTInt1)
	UInt16 = dbi.NewDbDataType("UInt16", dbi.DTInt16).WithCT(dbi.CTInt2)
	UInt32 = dbi.NewDbDataType("UInt32", dbi.DTInt32).WithCT(dbi.CTInt4)
	UInt64 = dbi.NewDbDataType("UInt64", dbi.DTInt64).WithCT(dbi.CTInt8)
	Int8   = dbi.NewDbDataType("Int8", dbi.DTInt8).WithCT(dbi.CTInt1)
	Int16  = dbi.NewDbDataType("Int16", dbi.DTInt16).WithCT(dbi.CTInt2)
	Int32  = dbi.NewDbDataType("Int32", dbi.DTInt32).WithCT(dbi.CTInt4)
	Int64  = dbi.NewDbDataType("Int64", dbi.DTInt64).WithCT(dbi.CTInt8)

	Float32 = dbi.NewDbDataType("Float32", dbi.DTNumeric).WithCT(dbi.CTNumeric)
	Float64 = dbi.NewDbDataType("Float64", dbi.DTNumeric).WithCT(dbi.CTNumeric)

	// String types
	String      = dbi.NewDbDataType("String", dbi.DTString).WithCT(dbi.CTVarchar)
	FixedString = dbi.NewDbDataType("FixedString", dbi.DTString).WithCT(dbi.CTChar)

	// Date and time types
	DateTime   = dbi.NewDbDataType("DateTime", dbi.DTDateTime).WithCT(dbi.CTDateTime)
	DateTime64 = dbi.NewDbDataType("DateTime64", dbi.DTDateTime).WithCT(dbi.CTDateTime)
	Date       = dbi.NewDbDataType("Date", dbi.DTDate).WithCT(dbi.CTDate)
	Date32     = dbi.NewDbDataType("Date32", dbi.DTDate).WithCT(dbi.CTDate)

	// Other types
	UUID = dbi.NewDbDataType("UUID", dbi.DTString).WithCT(dbi.CTVarchar)
	IPv4 = dbi.NewDbDataType("IPv4", dbi.DTString).WithCT(dbi.CTVarchar)
	IPv6 = dbi.NewDbDataType("IPv6", dbi.DTString).WithCT(dbi.CTVarchar)
	Bool = dbi.NewDbDataType("Bool", dbi.DTBool).WithCT(dbi.CTBool)

	// Decimal types
	Decimal    = dbi.NewDbDataType("Decimal", dbi.DTDecimal).WithCT(dbi.CTDecimal)
	Decimal32  = dbi.NewDbDataType("Decimal32", dbi.DTDecimal).WithCT(dbi.CTDecimal)
	Decimal64  = dbi.NewDbDataType("Decimal64", dbi.DTDecimal).WithCT(dbi.CTDecimal)
	Decimal128 = dbi.NewDbDataType("Decimal128", dbi.DTDecimal).WithCT(dbi.CTDecimal)
	Decimal256 = dbi.NewDbDataType("Decimal256", dbi.DTDecimal).WithCT(dbi.CTDecimal)

	// Enum types
	Enum8  = dbi.NewDbDataType("Enum8", dbi.DTString).WithCT(dbi.CTEnum)
	Enum16 = dbi.NewDbDataType("Enum16", dbi.DTString).WithCT(dbi.CTEnum)

	// Complex types
	Array                   = dbi.NewDbDataType("Array", dbi.DTString).WithCT(dbi.CTVarchar)
	Tuple                   = dbi.NewDbDataType("Tuple", dbi.DTString).WithCT(dbi.CTVarchar)
	Map                     = dbi.NewDbDataType("Map", dbi.DTString).WithCT(dbi.CTVarchar)
	Nested                  = dbi.NewDbDataType("Nested", dbi.DTString).WithCT(dbi.CTVarchar)
	AggregateFunction       = dbi.NewDbDataType("AggregateFunction", dbi.DTString).WithCT(dbi.CTVarchar)
	SimpleAggregateFunction = dbi.NewDbDataType("SimpleAggregateFunction", dbi.DTString).WithCT(dbi.CTVarchar)

	// Special types
	LowCardinality = dbi.NewDbDataType("LowCardinality", dbi.DTString).WithCT(dbi.CTVarchar)
	Nullable       = dbi.NewDbDataType("Nullable", dbi.DTString).WithCT(dbi.CTVarchar)
)

// Get all ClickHouse data types as a map for easy lookup
func GetAllClickHouseDataTypes() map[string]*dbi.DbDataType {
	return map[string]*dbi.DbDataType{
		"UInt8": UInt8, "UInt16": UInt16, "UInt32": UInt32, "UInt64": UInt64,
		"Int8": Int8, "Int16": Int16, "Int32": Int32, "Int64": Int64,
		"Float32": Float32, "Float64": Float64,
		"String": String, "FixedString": FixedString,
		"DateTime": DateTime, "DateTime64": DateTime64, "Date": Date, "Date32": Date32,
		"UUID": UUID, "IPv4": IPv4, "IPv6": IPv6, "Bool": Bool,
		"Decimal": Decimal, "Decimal32": Decimal32, "Decimal64": Decimal64,
		"Decimal128": Decimal128, "Decimal256": Decimal256,
		"Enum8": Enum8, "Enum16": Enum16,
		"Array": Array, "Tuple": Tuple, "Map": Map, "Nested": Nested,
		"AggregateFunction": AggregateFunction, "SimpleAggregateFunction": SimpleAggregateFunction,
		"LowCardinality": LowCardinality, "Nullable": Nullable,
	}
}
