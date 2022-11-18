package postgres

import (
	"database/sql"
	"fmt"
	"github.com/zzztttkkk/0.0/internal/sqlx"
	"reflect"
	"strconv"
)

type FieldDefinition struct {
	SqlType  string
	Default  sql.NullString
	Check    sql.NullString
	Nullable bool
	Unique   bool
}

func (fd *FieldDefinition) CheckAnd(v string, args ...any) {
	v = fmt.Sprintf(v, args)
	fd.Check.Valid = true
	if len(fd.Check.String) < 1 {
		fd.Check.String = v
	} else {
		fd.Check.String = and(fd.Check.String, v)
	}
}

func (fd *FieldDefinition) CheckOr(v string, args ...any) {
	v = fmt.Sprintf(v, args)

	fd.Check.Valid = true
	if len(fd.Check.String) < 1 {
		fd.Check.String = v
	} else {
		fd.Check.String = or(fd.Check.String, v)
	}
}

func and(l, r string) string {
	return fmt.Sprintf("((%s) AND (%s))", l, r)
}

func or(l, r string) string {
	return fmt.Sprintf("((%s) OR (%s))", l, r)
}

func (my *Driver) DDL(v any) string {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Struct {
		panic(fmt.Errorf("`%+v` is not a struct", v))
	}

	smap := sqlx.DBReflectMapper.TypeMap(val.Type())
	var fields []*FieldDefinition
	for _, info := range smap.Index {
		if info.Path != info.Name || info.Embedded {
			continue
		}
		var fd *FieldDefinition
		fv := val.MethodByName(fmt.Sprintf("DDL%s", info.Field.Name))
		if fv.IsValid() && fv.Type().Out(0) == reflect.TypeOf(fd) {
			temp := fv.Call(nil)
			if len(temp) == 1 {
				fd = temp[0].Interface().(*FieldDefinition)
			}
		}

		if fd != nil {
			fields = append(fields, fd)
			continue
		}

		fd = &FieldDefinition{}
		fd.SqlType = psqlType(info.Name, info.Zero.Type(), info.Options, fd)
		_, nullable := info.Options["nullable"]
		fd.Nullable = nullable
		_, unique := info.Options["unique"]
		fd.Unique = unique

		fields = append(fields, fd)
	}
	return ""
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

func psqlType(name string, t reflect.Type, opts map[string]string, fd *FieldDefinition) string {
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
