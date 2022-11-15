package sqlx

import (
	"database/sql"
	"reflect"
	"time"
)

var (
	BuiltinSqlTypes         = map[reflect.Type]reflect.Type{}
	BuiltinTimeType         = reflect.TypeOf(time.Time{})
	BuiltinI64Type          = reflect.TypeOf(int64(0))
	BuiltinStringType       = reflect.TypeOf("")
	SqlScannerInterfaceType = reflect.TypeOf((*sql.Scanner)(nil)).Elem()
)

func init() {
	BuiltinSqlTypes[reflect.TypeOf((*sql.NullString)(nil)).Elem()] = BuiltinStringType
	BuiltinSqlTypes[reflect.TypeOf((*sql.NullByte)(nil)).Elem()] = reflect.TypeOf(uint8(0))
	BuiltinSqlTypes[reflect.TypeOf((*sql.NullBool)(nil)).Elem()] = reflect.TypeOf(false)
	BuiltinSqlTypes[reflect.TypeOf((*sql.NullFloat64)(nil)).Elem()] = reflect.TypeOf(float64(0))
	BuiltinSqlTypes[reflect.TypeOf((*sql.NullInt16)(nil)).Elem()] = reflect.TypeOf(int16(0))
	BuiltinSqlTypes[reflect.TypeOf((*sql.NullInt32)(nil)).Elem()] = reflect.TypeOf(int32(0))
	BuiltinSqlTypes[reflect.TypeOf((*sql.NullInt64)(nil)).Elem()] = BuiltinI64Type
	BuiltinSqlTypes[reflect.TypeOf((*sql.NullTime)(nil)).Elem()] = BuiltinTimeType
}
