package sqlx

import (
	"context"
	"fmt"
	"github.com/zzztttkkk/0.0/internal/utils"
	"reflect"
	"strings"
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
}

type IndexInfo struct {
	Fields []*IndexField
	Unique bool
}

type FieldDefinition struct {
	Name       string
	SqlType    string
	PrimaryKey bool
	Default    string
	Check      string
	Nullable   bool
	Unique     bool
	Indexes    []IndexField
}

func (fd *FieldDefinition) AppendIndex(field IndexField) {
	fd.Indexes = append(fd.Indexes, field)
}

func (fd *FieldDefinition) CheckAnd(v string, args ...any) {
	if fd == nil {
		return
	}

	v = fmt.Sprintf(v, args...)
	if len(fd.Check) < 1 {
		fd.Check = v
	} else {
		fd.Check = and(fd.Check, v)
	}
}

func (fd *FieldDefinition) CheckOr(v string, args ...any) {
	v = fmt.Sprintf(v, args)

	if len(fd.Check) < 1 {
		fd.Check = v
	} else {
		fd.Check = or(fd.Check, v)
	}
}

func (db *DB) TableName(val reflect.Value) string {
	var tablename string
	if tableNameFn := val.MethodByName("TableName"); tableNameFn.IsValid() {
		if fn, _ := tableNameFn.Interface().(func() string); fn != nil {
			tablename = fn()
		}
	}
	if len(tablename) < 1 {
		tablename = strings.ToLower(val.Type().Name())
	}
	return tablename
}

func (db *DB) DropTable(ctx context.Context, name string) error {
	_, err := db.Execute(ctx, fmt.Sprintf("%s TABLE %s", "DROP", name), nil)
	return err
}

func parseIndex(v string) []*IndexField {
	if len(v) < 1 {
		return nil
	}

	var fs []*IndexField

	for _, fi := range strings.Split(v, "|") {
		fi = strings.TrimSpace(fi)
		parts := strings.Split(fi, ";")

		fv := &IndexField{}

		if len(parts) == 1 {
			fv.Name = parts[0]
		} else if len(parts) == 2 {
			fv.Name = parts[0]
			switch strings.ToLower(parts[1]) {
			case "desc":
				fv.OrderType = IndexFieldOrderDesc
			case "asc":
				fv.OrderType = IndexFieldOrderAsc
			default:
				fv.OrderType = IndexFieldOrderDesc
			}
		}
		fs = append(fs, fv)
	}
	return fs
}

func (db *DB) CreateTable(ctx context.Context, v any) error {
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

		if _, ok := info.Options["nullable"]; ok {
			fd.Nullable = true
		}

		if _, ok := info.Options["unique"]; ok {
			fd.Unique = true
		}

		if _, ok := info.Options["primary"]; ok {
			fd.PrimaryKey = true
		}

		if dv, ok := info.Options["default"]; ok {
			fd.Default = dv
		}

		parseIndex(info.Options["index"])

		fd.Name = info.Name
		fields = append(fields, fd)
	}

	var primaryKeys []*FieldDefinition

	fields = utils.SliceMap(
		utils.SliceFilter(fields, func(v *FieldDefinition) bool { return v != nil }),

		func(_ int, v *FieldDefinition) *FieldDefinition {
			if v.PrimaryKey {
				primaryKeys = append(primaryKeys, v)
			}
			return v
		},
	)

	if len(fields) < 1 {
		panic(fmt.Errorf("0.0/internal/sqlx: `%+v` got empty filed definitions", v))
	}

	if len(primaryKeys) < 1 {
		panic(fmt.Errorf("0.0/internal/sqlx: `%+v` got empty primary keys", v))
	}

	tablename := db.TableName(val)

	var sb strings.Builder
	sb.WriteString("CREATE TABLE IF NOT EXISTS ")
	sb.WriteString(tablename)
	sb.WriteString(" (\r\n")

	for _, field := range fields {
		sb.WriteRune('\t')
		sb.WriteString(field.Name)
		sb.WriteRune(' ')
		sb.WriteString(field.SqlType)

		if field.Unique {
			sb.WriteString(" UNIQUE")
		}

		if !field.Nullable {
			sb.WriteString(" NOT NULL")
		}

		if len(field.Check) > 0 {
			sb.WriteString(" CHECK (")
			sb.WriteString(field.Check)
			sb.WriteRune(')')
		}

		if len(field.Default) > 0 {
			sb.WriteString(" DEFAULT ")
			sb.WriteString(field.Default)
		}

		sb.WriteString(",\r\n")
	}

	sb.WriteString("\tprimary key (")
	sb.WriteString(strings.Join(utils.SliceMap(primaryKeys, func(_ int, fd *FieldDefinition) string { return fd.Name }), ","))
	sb.WriteString(")\r\n)\r\n")

	ddl := sb.String()
	if db.logger != nil {
		db.logger.Printf(ddl)
	}
	_, err := db.Execute(ctx, ddl, nil)
	return err
}
