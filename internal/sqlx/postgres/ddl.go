package postgres

import (
	"fmt"
	"github.com/zzztttkkk/0.0/internal/sqlx"
	"reflect"
)

func (my *Driver) DDL(v any) string {
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Struct {
		panic(fmt.Errorf("`%+v` is not a struct", v))
	}

	smap := sqlx.DBReflectMapper.TypeMap(val.Type())
	var fields []string
	for _, info := range smap.Index {
		if info.Path != info.Name || info.Embedded {
			continue
		}
		if ddl := info.Options["ddl"]; len(ddl) > 0 {
			fields = append(fields, ddl)
			fmt.Println(ddl)
			continue
		}
		fmt.Println(info.Name, psqlType(info.Name, info.Zero.Type()))
	}
	return ""
}

func psqlType(name string, t reflect.Type) string {
	switch t.Kind() {
	case reflect.Int, reflect.Int64:
		return "bigint"
	case reflect.Int8:
		return fmt.Sprintf("numeric(3) CHECK (%s < 128 AND %s > -128)", name, name)
	case reflect.Int16:
		return "smallint"
	case reflect.Int32:
		return "integer"
	case reflect.Uint, reflect.Uint64:
		return fmt.Sprintf("numeric(19) CHECK (%s >= 0)", name)
	case reflect.Uint8:
		return fmt.Sprintf("numeric(3) CHECK (%s < 256 AND %s >= 0)", name, name)
	case reflect.Uint16:
		return fmt.Sprintf("numeric(5) CHECK (%s < 65536 AND %s >= 0)", name, name)
	case reflect.Uint32:
		return fmt.Sprintf("numeric(10) CHECK (%s < 4294967296 AND %s >= 0)", name, name)
	}
	return ""
}
