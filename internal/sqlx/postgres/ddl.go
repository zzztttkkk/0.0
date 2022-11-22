package postgres

import (
	"database/sql"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/zzztttkkk/0.0/internal/sqlx"
	"github.com/zzztttkkk/0.0/internal/utils"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func (_ *Driver) DDL(info *utils.FieldInfo) *sqlx.FieldDefinition {
	var fd = &sqlx.FieldDefinition{}

	fd.SqlType = psqlType(info.Name, info.Field.Type, info.Options, fd)

	_, incr := info.Options["incr"]
	if incr {
		switch fd.SqlType {
		case "smallint":
			{
				fd.SqlType = "smallserial"
			}
		case "integer":
			{
				fd.SqlType = "serial"
			}
		case "bigint":
			{
				fd.SqlType = "bigserial"
			}
		}
	}
	return fd
}

func getLength(v string) (int, bool) {
	if v[0] == '~' {
		n, e := strconv.ParseUint(v[1:], 10, 32)
		if e != nil {
			panic(fmt.Errorf("bad field length: `%s`", v))
		}
		return int(n), false
	}
	n, e := strconv.ParseUint(v, 10, 32)
	if e != nil {
		panic(fmt.Errorf("bad field length: `%s`", v))
	}
	return int(n), true
}

var (
	_UuidType    = reflect.TypeOf((*pgtype.UUID)(nil)).Elem()
	_HStoreType  = reflect.TypeOf((*pgtype.Hstore)(nil)).Elem()
	_DateType    = reflect.TypeOf((*pgtype.Date)(nil)).Elem()
	_NullTypes   = make(map[reflect.Type]reflect.Type)
	_TimeType    = reflect.TypeOf((*time.Time)(nil)).Elem()
	_AnyJsonType = reflect.TypeOf((*AnyJSON)(nil)).Elem()
)

func init() {
	addToMap := func(a, b any) {
		_NullTypes[reflect.TypeOf(a)] = reflect.TypeOf(b)
	}
	addToMap(sql.NullString{}, "")
	addToMap(sql.NullBool{}, false)
	addToMap(sql.NullFloat64{}, float64(0))
	addToMap(sql.NullInt16{}, int16(0))
	addToMap(sql.NullInt32{}, int32(0))
	addToMap(sql.NullInt64{}, int64(0))
	addToMap(sql.NullTime{}, time.Now())
	addToMap(sql.NullByte{}, uint8(0))
}

func psqlType(name string, t reflect.Type, opts map[string]string, fd *sqlx.FieldDefinition) string {
	userType := strings.TrimSpace(opts["sqltype"])
	if len(userType) > 0 {
		return userType
	}

	switch t {
	case _HStoreType:
		return "hstore"
	case _UuidType:
		return "uuid"
	case _TimeType, _DateType:
		return "timestamp"
	case _AnyJsonType:
		return "jsonb"
	}

	switch t.Kind() {
	case reflect.Int, reflect.Int64:
		return "bigint"
	case reflect.Int8:
		fd.CheckAnd("%s < 128", name)
		fd.CheckAnd("%s > -128", name)
		return "numeric(3)"
	case reflect.Int16:
		return "smallint"
	case reflect.Int32:
		return "integer"
	case reflect.Uint, reflect.Uint64:
		fd.CheckAnd("%s >= 0", name)
		return "numeric(20)"
	case reflect.Uint8:
		fd.CheckAnd("%s < 256", name)
		fd.CheckAnd("%s >= 0", name)
		return "numeric(3)"
	case reflect.Uint16:
		fd.CheckAnd("%s < 65536", name)
		fd.CheckAnd("%s >= 0", name)
		return "numeric(5)"
	case reflect.Uint32:
		fd.CheckAnd("%s < 4294967296", name)
		fd.CheckAnd("%s >= 0", name)
		return "numeric(10)"
	case reflect.String:
		{
			lengthOV := opts["length"]
			if len(lengthOV) < 1 {
				return "text"
			}

			length, isFixed := getLength(lengthOV)
			if length < 1 || length > 6555 {
				return "text"
			}
			if isFixed {
				return fmt.Sprintf("char(%d)", length)
			}
			return fmt.Sprintf("varchar(%d)", length)
		}
	case reflect.Bool:
		return "boolean"
	case reflect.Float32:
		return "real"
	case reflect.Float64:
		return "double precision"
	case reflect.Slice:
		{
			// bytes
			if t.Elem().Kind() == reflect.Uint8 {
				return "bytea"
			}

			eleSqlType := psqlType(name, t.Elem(), opts, nil)
			return fmt.Sprintf("[]%s", eleSqlType)
		}
	case reflect.Struct:
		{
			realType := _NullTypes[t]
			if realType != nil {
				if fd != nil {
					fd.Nullable = true
				}
				return psqlType(name, realType, opts, fd)
			}
		}
	}
	return ""
}
