package sqlx

import (
	"fmt"
	"github.com/zzztttkkk/0.0/internal/utils"
	"reflect"
	"strings"
)

type _Param struct {
	name  string
	begin int
	end   int
}

func scanParams(q []byte) []_Param {
	var lst []_Param
	begin := -1
	status := -1
	var quote byte
	var buf []byte
	for idx, r := range q {
		if quote != 0 {
			if r == quote {
				quote = 0
			}
			continue
		}

		if r == '"' || r == '\'' {
			quote = r
			continue
		}

		if status < 0 && r == '$' {
			status++
			begin = idx
			continue
		}

		switch status {
		case 0:
			if r == '{' {
				status++
			} else {
				status = -1
			}
			continue
		case 1:
			if r == '}' {
				status = -1
				lst = append(lst, _Param{name: string(buf), begin: begin, end: idx})
				buf = buf[:0]
			} else {
				buf = append(buf, r)
			}
		}
	}

	return lst
}

func ScanParams(txt string, driver Driver) (string, []string) {
	if !strings.Contains(txt, "${") {
		return txt, nil
	}

	q := utils.B(txt)
	lst := scanParams(q)
	if len(lst) < 1 {
		return txt, nil
	}

	var buf strings.Builder
	var keys []string
	cur := 0
	for idx, param := range lst {
		for cur < param.begin {
			buf.WriteByte(q[cur])
			cur++
		}
		buf.WriteString(driver.Placeholder(idx, param.name))
		keys = append(keys, param.name)
		cur = param.end + 1
	}

	for cur < len(q) {
		buf.WriteByte(q[cur])
		cur++
	}
	return buf.String(), keys
}

type Params map[string]interface{}
type ParamSlice []interface{}

func (p Params) Keys() []string {
	if len(p) < 1 {
		return nil
	}
	lst := make([]string, 0, len(p))
	for k := range p {
		lst = append(lst, k)
	}
	return lst
}

func (p Params) Values(keys []string) ([]interface{}, error) {
	if len(p) < 1 {
		return nil, nil
	}
	var lst []interface{}
	for _, k := range keys {
		v, ok := p[k]
		if !ok {
			return nil, fmt.Errorf("0.0/internal/sqlx: missing key `%s`", k)
		}
		lst = append(lst, v)
	}
	return lst, nil
}

var (
	paramsType      = reflect.TypeOf(Params{})
	paramSliceType  = reflect.TypeOf(ParamSlice{})
	sliceType       = reflect.TypeOf([]interface{}{})
	DBReflectMapper = utils.NewMapper("db")
)

func paramsToMap(params interface{}) (Params, error) {
	if params == nil {
		return nil, nil
	}

	t := reflect.TypeOf(params)
	if t == paramsType {
		return params.(Params), nil
	}
	if t == mapType {
		return params.(map[string]interface{}), nil
	}

	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("0.0/internal/sqlx: bad params value `%v`, should be a `Parmas` or `struct`", params)
	}

	var m Params
	var pv = reflect.ValueOf(params)
	for n, f := range DBReflectMapper.TypeMap(t).Names {
		fv := pv.FieldByIndex(f.Index)
		if m == nil {
			m = make(Params)
		}
		m[n] = fv.Interface()
	}
	return m, nil
}

func paramsToArgs(params interface{}, keys []string) ([]interface{}, error) {
	if len(keys) < 1 {
		return nil, nil
	}

	t := reflect.TypeOf(params)
	switch t {
	case paramsType:
		return params.(Params).Values(keys)
	case mapType:
		return Params(params.(map[string]interface{})).Values(keys)
	case sliceType:
		return params.([]interface{}), nil
	case paramSliceType:
		return params.(ParamSlice), nil
	}

	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("0.0/internal/sqlx: bad params value `%v`, should be a `Parmas` or `struct`", params)
	}

	var args []interface{}
	var m = DBReflectMapper.TypeMap(t).Names
	var pv = reflect.ValueOf(params)
	for _, k := range keys {
		f, ok := m[k]
		if !ok {
			return nil, fmt.Errorf("0.0/internal/sqlx: missing key `%s`", k)
		}
		fv := pv.FieldByIndex(f.Index)
		args = append(args, fv.Interface())
	}
	return args, nil
}
