package sqlx

import (
	"database/sql"
	"fmt"
	"github.com/zzztttkkk/0.0/internal/utils"
	"reflect"
)

func and(l, r string) string {
	return fmt.Sprintf("((%s) AND (%s))", l, r)
}

func or(l, r string) string {
	return fmt.Sprintf("((%s) OR (%s))", l, r)
}

type IndexFieldOrderType bool

const (
	IndexFieldOrderDesc = IndexFieldOrderType(false)
	IndexFieldOrderAsc  = IndexFieldOrderType(true)
)

type IndexField struct {
	Name        string
	OrderType   IndexFieldOrderType
	SortInIndex int
	appendIdx   int
}

type FieldDefinition struct {
	Name     string
	SqlType  string
	Default  sql.NullString
	Check    sql.NullString
	Nullable bool
	Unique   bool
	Indexes  []IndexField
}

func (fd *FieldDefinition) AppendIndex(field IndexField) {
	fd.Indexes = append(fd.Indexes, field)
	fd.Indexes[len(fd.Indexes)-1].appendIdx = len(fd.Indexes) - 1
}

func (fd *FieldDefinition) CheckAnd(v string, args ...any) {
	if fd == nil {
		return
	}

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

func (db *DB) CreateTable(v any) {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		panic(fmt.Errorf("0.0/internal/sqlx: `%+v` is not a struct", v))
	}

	smap := DBReflectMapper.TypeMap(val.Type())
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
		if fd == nil {
			fd = db.driver.DDL(info)
		}
		fd.Name = info.Name
		fields = append(fields, fd)
	}

	fields = utils.SliceFilter(fields, func(v *FieldDefinition) bool {
		return v != nil
	})

	if len(fields) < 1 {
		panic(fmt.Errorf("0.0/internal/sqlx: `%+v` got empty filed definitions", v))
	}

	for _, field := range fields {
		fmt.Println(field.Name, field.SqlType)
	}
}
