package postgres

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/zzztttkkk/0.0/internal/utils"
)

type AnyJSON struct {
	Val any
}

func (v *AnyJSON) Scan(src any) error {
	var ds []byte
	switch ns := src.(type) {
	case string:
		ds = utils.B(ns)
	case []byte:
		ds = ns
	default:
		return fmt.Errorf("bad value type")
	}
	return json.Unmarshal(ds, &v.Val)
}

func (v *AnyJSON) Value() (driver.Value, error) {
	return json.Marshal(v.Val)
}

var _ driver.Valuer = (*AnyJSON)(nil)
var _ sql.Scanner = (*AnyJSON)(nil)

var (
	ErrorUnexpectedJSONValue = errors.New("0.0/internal/sqlx/postgres: unexpected value")
	ErrorUnexpectedJSONKey   = errors.New("0.0/internal/sqlx/postgres: unexpected key")
)

func (v *AnyJSON) Peek(keys ...any) (any, error) {
	var current = v.Val

	for _, key := range keys {
		switch kv := key.(type) {
		case string:
			switch crv := current.(type) {
			case map[string]any:
				current = crv[kv]
			default:
				return nil, ErrorUnexpectedJSONValue
			}
		case int:
			switch crv := current.(type) {
			case []any:
				current = crv[kv]
			default:
				return nil, ErrorUnexpectedJSONValue
			}
		default:
			return nil, ErrorUnexpectedJSONKey
		}
	}
	return current, nil
}

func (v *AnyJSON) PeekMap(keys ...any) (map[string]any, error) {
	pv, e := v.Peek(keys...)
	if e != nil {
		return nil, e
	}
	switch rv := pv.(type) {
	case map[string]any:
		return rv, nil
	default:
		return nil, ErrorUnexpectedJSONValue
	}
}

func (v *AnyJSON) PeekArray(keys ...any) ([]any, error) {
	pv, e := v.Peek(keys...)
	if e != nil {
		return nil, e
	}
	switch rv := pv.(type) {
	case []any:
		return rv, nil
	default:
		return nil, ErrorUnexpectedJSONValue
	}
}

func (v *AnyJSON) PeekString(keys ...any) (string, error) {
	pv, e := v.Peek(keys...)
	if e != nil {
		return "", e
	}
	switch rv := pv.(type) {
	case string:
		return rv, nil
	default:
		return "", ErrorUnexpectedJSONValue
	}
}

func (v *AnyJSON) PeekNumber(keys ...any) (float64, error) {
	pv, e := v.Peek(keys...)
	if e != nil {
		return 0, e
	}
	switch rv := pv.(type) {
	case float64:
		return rv, nil
	default:
		return 0, ErrorUnexpectedJSONValue
	}
}

func (v *AnyJSON) PeekBool(keys ...any) (bool, error) {
	pv, e := v.Peek(keys...)
	if e != nil {
		return false, e
	}
	switch rv := pv.(type) {
	case bool:
		return rv, nil
	default:
		return false, ErrorUnexpectedJSONValue
	}
}
