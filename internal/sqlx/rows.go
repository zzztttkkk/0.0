package sqlx

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/zzztttkkk/0.0/internal/utils"
	"reflect"
)

type Rows struct {
	*sql.Rows
}

type DirectDist []interface{}

var (
	directDistType        = reflect.TypeOf(DirectDist{})
	interfaceType         = reflect.TypeOf((*interface{})(nil)).Elem()
	ErrUnexpectedDistType = errors.New("0.0/internal/sqlx: unexpected dist type")
	ErrEmptySlice         = errors.New(
		"0.0/internal/sqlx: you should know how many rows will be fetched, please use a pre-allocated slice",
	)
	mapType = reflect.TypeOf(map[string]interface{}{})
)

func isStrAnyMapType(t reflect.Type) bool {
	if t == mapType {
		return true
	}
	return t.Kind() == reflect.Map && t.Key().Kind() == reflect.String && t.Elem().Kind() == reflect.Interface
}

func (rows *Rows) doScan(dist interface{}, columns []string, temp *[]interface{}) error {
	dt := reflect.TypeOf(dist)
	if dt == directDistType {
		return rows.Rows.Scan(dist.(DirectDist)...)
	}

	v := reflect.ValueOf(dist).Elem()
	t := v.Type()

	if columns == nil {
		var err error
		columns, err = rows.Columns()
		if err != nil {
			return err
		}
	}

	if temp == nil {
		var dist = make([]interface{}, 0, len(columns))
		temp = &dist
	}

	if isStrAnyMapType(t) {
		return rows.scanToMap(&v, columns, temp)
	}

	if t.Kind() == reflect.Struct {
		if t == BuiltinTimeType || t.Implements(SqlScannerInterfaceType) {
			return rows.Rows.Scan(dist)
		}
		return rows.scanToStruct(&v, columns, temp)
	}
	return rows.Rows.Scan(dist)
}

// Scan
// @param `dist` must be a point
func (rows *Rows) Scan(dist interface{}) error { return rows.doScan(dist, nil, nil) }

func (rows *Rows) scanToMap(v *reflect.Value, columns []string, temp *[]interface{}) error {
	var m map[string]interface{}
	if v.IsNil() {
		m = map[string]interface{}{}
		v.Set(reflect.ValueOf(m))
	} else {
		m = v.Interface().(map[string]interface{})
	}

	defer func() { *temp = (*temp)[:0] }()

	for i := 0; i < len(columns); i++ {
		*temp = append(*temp, reflect.New(interfaceType).Interface())
	}

	if err := rows.Rows.Scan(*temp...); err != nil {
		return err
	}
	for idx, c := range columns {
		m[c] = reflect.ValueOf((*temp)[idx]).Elem().Interface()
	}
	return nil
}

func (rows *Rows) scanToStruct(v *reflect.Value, columns []string, temp *[]interface{}) error {
	defer func() { *temp = (*temp)[:0] }()

	vm := DBReflectMapper.FieldMap(*v)
	for _, c := range columns {
		v, ok := vm[c]
		if !ok {
			return fmt.Errorf("0.0/internal/sqlx: missing column `%s`", c)
		}
		*temp = append(*temp, v.Addr().Interface())
	}
	return rows.Rows.Scan(*temp...)
}

func (rows *Rows) selectToPointerSlice(sliceV reflect.Value, et reflect.Type) (*reflect.Value, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	var temp = make([]interface{}, 0, len(columns))

	et = et.Elem()
	for rows.Next() {
		elePtrV := reflect.New(et)
		if err := rows.doScan(elePtrV.Interface(), columns, &temp); err != nil {
			return nil, err
		}
		sliceV = reflect.Append(sliceV, elePtrV)
	}
	return &sliceV, rows.Err()
}

func (rows *Rows) selectToValueSlice(sliceV reflect.Value, et reflect.Type) (*reflect.Value, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	if sliceV.Cap() < 1 {
		return nil, ErrEmptySlice
	}

	var temp = make([]interface{}, 0, len(columns))

	for rows.Next() {
		l := sliceV.Len() + 1
		doAppend := true
		var eleV reflect.Value
		var elePtr interface{}
		if sliceV.CanSet() && l < sliceV.Cap() {
			doAppend = false
			sliceV.SetLen(l)
			eleV = sliceV.Index(l - 1)
			elePtr = eleV.Addr().Interface()
		} else {
			elePtrV := reflect.New(et)
			elePtr = elePtrV.Interface()
			eleV = elePtrV.Elem()
		}
		if err := rows.doScan(elePtr, columns, &temp); err != nil {
			return nil, err
		}
		if doAppend {
			sliceV = reflect.Append(sliceV, eleV)
		}
	}
	return &sliceV, rows.Err()
}

type JoinedEmbedDistGetter func(raw interface{}, idx int) interface{}

func (rows *Rows) ScanJoined(dist interface{}, get JoinedEmbedDistGetter) error {
	v := reflect.ValueOf(dist).Elem()
	t := v.Type()
	if t.Kind() != reflect.Struct {
		return ErrUnexpectedDistType
	}

	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	distIdx := 0
	var currentDist interface{}
	var cdV reflect.Value
	var cdT *utils.StructMap
	var cdFieldsRemain int

	var ptrs []interface{}
	var lowIdx int
	for idx, c := range columns {
		if lowIdx > 0 && idx < lowIdx {
			continue
		}
		lowIdx = -1

		if cdT == nil {
			currentDist = get(dist, distIdx)
			if currentDist == nil {
				return ErrUnexpectedDistType
			}

			cdV = reflect.ValueOf(currentDist)
			if cdV.Kind() == reflect.Slice {
				ptrs = append(ptrs, currentDist.([]interface{})...)
				distIdx++
				lowIdx = idx + cdV.Len()
				continue
			}

			cdV = cdV.Elem()
			cdT = DBReflectMapper.TypeMap(reflect.TypeOf(currentDist).Elem())
			cdFieldsRemain = len(cdT.Names)
			if cdFieldsRemain < 1 {
				cdT = nil
				distIdx++
				continue
			}
		}

		fi, ok := cdT.Names[c]
		if !ok {
			return fmt.Errorf("0.0/internal/sqlx: bad column name `%s`", c)
		}
		f := cdV.FieldByIndex(fi.Index)
		ptrs = append(ptrs, f.Addr().Interface())
		cdFieldsRemain--
		if cdFieldsRemain < 1 {
			cdT = nil
			distIdx++
			continue
		}
	}
	return rows.Rows.Scan(ptrs...)
}

func (rows *Rows) fetchOne(dist interface{}) error {
	columns, err := rows.Columns()
	if err != nil {
		return err
	}
	for rows.Next() {
		return rows.doScan(dist, columns, nil)
	}
	return sql.ErrNoRows
}

func (rows *Rows) fetchMany(slicePtr interface{}) error {
	var err error
	sliceV := reflect.ValueOf(slicePtr).Elem()
	eleT := sliceV.Type().Elem()
	var vPtr *reflect.Value
	if eleT.Kind() == reflect.Ptr {
		vPtr, err = rows.selectToPointerSlice(sliceV, eleT)
	} else {
		vPtr, err = rows.selectToValueSlice(sliceV, eleT)
	}
	if err != nil {
		return err
	}
	sliceV.Set(*vPtr)
	return nil
}

func (rows *Rows) getJoined(dist interface{}, joinedGet JoinedEmbedDistGetter) error {
	for rows.Next() {
		err := rows.ScanJoined(dist, joinedGet)
		if err != nil {
			return err
		}
		return nil
	}
	return sql.ErrNoRows
}

func (rows *Rows) selectJoined(slicePtr interface{}, joinedGet JoinedEmbedDistGetter) error {
	var err error
	sliceV := reflect.ValueOf(slicePtr).Elem()
	eleT := sliceV.Type().Elem()
	isPtrSlice := eleT.Kind() == reflect.Ptr
	if !isPtrSlice && sliceV.Cap() < 1 {
		return ErrEmptySlice
	}

	for rows.Next() {
		var elePtrV reflect.Value
		doAppend := true
		if isPtrSlice {
			elePtrV = reflect.New(eleT.Elem())
		} else {
			l := sliceV.Len() + 1
			if sliceV.CanSet() && l < sliceV.Cap() {
				doAppend = false
				sliceV.SetLen(l)
				elePtrV = sliceV.Index(l - 1).Addr()
			} else {
				elePtrV = reflect.New(eleT)
			}
		}

		err = rows.ScanJoined(elePtrV.Interface(), joinedGet)
		if err != nil {
			return err
		}

		if doAppend {
			if isPtrSlice {
				sliceV = reflect.Append(sliceV, elePtrV)
			} else {
				sliceV = reflect.Append(sliceV, elePtrV.Elem())
			}
		}
	}
	reflect.ValueOf(slicePtr).Elem().Set(sliceV)
	return rows.Err()
}
