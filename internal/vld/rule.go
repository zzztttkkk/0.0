package vld

import (
	"fmt"
	"html"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type RuleType int

const (
	RuleTypeInt = RuleType(iota)
	RuleTypeDouble
	RuleTypeBool
	RuleTypeString
	RuleTypeFile
	RuleTypeSlice
	RuleTypeTime
	RuleTypeVlder
)

const (
	defaultMemSize = 32 << 20
)

type Rule struct {
	Name      string
	RuleType  RuleType
	Gotype    reflect.Type
	Ele       *Rules
	SimpleEle *Rule
	Index     []int

	Optional bool

	MaxInt    *int64
	MinInt    *int64
	MaxDouble *float64
	MinDouble *float64

	MaxLen *int
	MinLen *int

	NoTrim   bool
	NoEscape bool
	Regexp   *regexp.Regexp

	TimeLayout string
	TimeUnit   string
}

func (rule *Rule) string2Int(v string) (any, error) {
	num, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("not a int")
	}

	if rule.MinInt != nil && num < *rule.MinInt {
		return nil, fmt.Errorf("out of range")
	}

	if rule.MaxInt != nil && num > *rule.MaxInt {
		return nil, fmt.Errorf("out of range")
	}

	switch rule.Gotype.Kind() {
	case reflect.Int:
		return int(num), nil
	case reflect.Int8:
		return int8(num), nil
	case reflect.Int16:
		return int16(num), nil
	case reflect.Int32:
		return int32(num), nil
	case reflect.Int64:
		return num, nil
	case reflect.Uint:
		return uint(num), nil
	case reflect.Uint8:
		return uint8(num), nil
	case reflect.Uint16:
		return uint16(num), nil
	case reflect.Uint32:
		return uint32(num), nil
	case reflect.Uint64:
		return uint64(num), nil
	}
	return num, nil
}

func (rule *Rule) string2Double(v string) (any, error) {
	num, err := strconv.ParseFloat(v, 10)
	if err != nil {
		return nil, fmt.Errorf("not a float")
	}

	if rule.MinDouble != nil && num < *rule.MinDouble {
		return nil, fmt.Errorf("out of range")
	}

	if rule.MaxDouble != nil && num > *rule.MaxDouble {
		return nil, fmt.Errorf("out of range")
	}

	switch rule.Gotype.Kind() {
	case reflect.Float32:
		return float32(num), nil
	default:
		return num, nil
	}
}

func (rule *Rule) string2Time(v string) (any, error) {
	if len(rule.TimeLayout) > 0 {
		return time.Parse(rule.TimeLayout, v)
	}

	num, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("can not cast to time")
	}

	switch rule.TimeUnit {
	case "", "s":
		return time.Unix(num, 0), nil
	default:
		return time.UnixMilli(num), nil
	}
}

func (rule *Rule) checkString(v string) (any, error) {
	if !rule.NoTrim {
		v = strings.TrimSpace(v)
	}

	if !rule.NoEscape {
		v = html.EscapeString(v)
	}

	l := len(v)
	if rule.MaxLen != nil && l > *rule.MaxLen {
		return nil, fmt.Errorf("length out of range")
	}
	if rule.MinLen != nil && l < *rule.MinLen {
		return nil, fmt.Errorf("length out of range")
	}

	if rule.Regexp != nil && !rule.Regexp.MatchString(v) {
		return nil, fmt.Errorf("not match regexp")
	}
	return v, nil
}

func (rule *Rule) singleSimpleEle(raw string) (any, error) {
	switch rule.RuleType {
	case RuleTypeString:
		{
			return rule.checkString(raw)
		}
	case RuleTypeInt:
		{
			return rule.string2Int(raw)
		}
	case RuleTypeDouble:
		{
			return rule.string2Double(raw)
		}
	case RuleTypeBool:
		{
			bol, err := strconv.ParseBool(raw)
			if err != nil {
				return nil, fmt.Errorf("not a bool")
			}
			return bol, nil
		}
	case RuleTypeTime:
		{
			return rule.string2Time(raw)
		}
	}
	return nil, fmt.Errorf("")
}

func (rule *Rule) get(req *http.Request) (any, error) {
	if rule.RuleType == RuleTypeSlice {
		return nil, nil
	}

	switch rule.RuleType {
	case RuleTypeVlder:
		{
			val := reflect.New(rule.Gotype)
			err := val.Interface().(Vlder).FromString(rule.Name)
			if err != nil {
				return nil, err
			}
			if rule.Gotype.Kind() == reflect.Pointer {
				return val.Interface(), nil
			}
			return val.Elem().Interface(), nil
		}
	case RuleTypeFile:
		{
			_ = req.ParseMultipartForm(defaultMemSize)
			fhs, ok := req.MultipartForm.File[rule.Name]
			if !ok || len(fhs) < 1 {
				if rule.Optional {
					return nil, nil
				}
				return nil, fmt.Errorf("miss required")
			}
			return fhs[0], nil
		}
	default:
		{
			raw := req.FormValue(rule.Name)
			if len(raw) < 1 {
				if rule.Optional {
					return nil, nil
				}
				return nil, fmt.Errorf("miss required")
			}

		}
	}
	return nil, nil
}

type Rules struct {
	Gotype reflect.Type
	Data   []*Rule
}

func (rules *Rules) BinFromRequest(req *http.Request) (any, error) {
	val := reflect.New(rules.Gotype).Elem()

	for _, rule := range rules.Data {
		v, err := rule.get(req)
		if err != nil {
			return nil, err
		}
		vv := reflect.ValueOf(v)
		if !vv.IsValid() {
			continue
		}
		val.FieldByIndex(rule.Index).Set(vv)
	}
	return val.Interface(), nil
}

func (rules *Rules) Validate(v any) error {
	return nil
}