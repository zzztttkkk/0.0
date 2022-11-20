package postgres

import (
	"fmt"
	"github.com/zzztttkkk/0.0/internal/sqlx"
	"github.com/zzztttkkk/0.0/internal/utils"
	"reflect"
	"strconv"
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

func psqlType(name string, t reflect.Type, opts map[string]string, fd *sqlx.FieldDefinition) string {
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
		return "numeric(19)"
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
	case reflect.Slice:
		{
			// bytes
			if t.Elem().Kind() == reflect.Uint8 {

			}
		}
	case reflect.Bool:
		return "boolean"
	case reflect.Float32:
		return "real"
	case reflect.Float64:
		return "double precision"
	}
	return ""
}
