package vld

import (
	"fmt"
	"github.com/zzztttkkk/0.0/internal/utils"
	"math"
	"mime/multipart"
	"reflect"
	"time"
)

var (
	mapper = utils.NewMapper("vld")
	cache  = make(map[reflect.Type]*Rules)

	fileType  = reflect.TypeOf((*multipart.FileHeader)(nil)).Elem()
	timeType  = reflect.TypeOf((*time.Time)(nil)).Elem()
	vlderType = reflect.TypeOf((*Vlder)(nil)).Elem()
)

func infoToRule(info *utils.FieldInfo, ft reflect.Type) *Rule {
	rule := &Rule{
		Name:   info.Name,
		Index:  info.Index,
		Gotype: info.Field.Type,
	}

	if ft == nil {
		ft = info.Field.Type
	}

	if ft.Implements(vlderType) {
		rule.RuleType = RuleTypeVlder
		return rule
	}

	switch ft.Kind() {
	case reflect.Int, reflect.Int64:
		{
			rule.RuleType = RuleTypeInt
		}
	case reflect.Int8:
		{
			rule.RuleType = RuleTypeInt
			rule.MinInt = new(int64)
			*rule.MinInt = math.MinInt8
			rule.MaxInt = new(int64)
			*rule.MaxInt = math.MaxInt8
		}
	case reflect.Int16:
		{
			rule.RuleType = RuleTypeInt
			rule.MinInt = new(int64)
			*rule.MinInt = math.MinInt16
			rule.MaxInt = new(int64)
			*rule.MaxInt = math.MaxInt16
		}
	case reflect.Int32:
		{
			rule.RuleType = RuleTypeInt
			rule.MinInt = new(int64)
			*rule.MinInt = math.MinInt32
			rule.MaxInt = new(int64)
			*rule.MaxInt = math.MaxInt32
		}
	case reflect.Uint, reflect.Uint64:
		{
			rule.RuleType = RuleTypeInt
			rule.MinInt = new(int64)
			*rule.MinInt = 0
		}
	case reflect.Uint8:
		{
			rule.RuleType = RuleTypeInt
			rule.MinInt = new(int64)
			*rule.MinInt = 0
			rule.MaxInt = new(int64)
			*rule.MaxInt = math.MaxUint8
		}
	case reflect.Uint16:
		{
			rule.RuleType = RuleTypeInt
			rule.MinInt = new(int64)
			*rule.MinInt = 0
			rule.MaxInt = new(int64)
			*rule.MaxInt = math.MaxUint16
		}
	case reflect.Uint32:
		{
			rule.RuleType = RuleTypeInt
			rule.MinInt = new(int64)
			*rule.MinInt = 0
			rule.MaxInt = new(int64)
			*rule.MaxInt = math.MaxUint32
		}
	case reflect.Float32, reflect.Float64:
		{
			rule.RuleType = RuleTypeDouble
		}
	case reflect.String:
		{
			rule.RuleType = RuleTypeString
		}
	case reflect.Bool:
		{
			rule.RuleType = RuleTypeBool
		}
	case reflect.Slice:
		{
			rule.RuleType = RuleTypeSlice
			switch ft.Elem().Kind() {
			case reflect.Struct:
				{
					rule.Ele = GetRules(ft.Elem())
				}
			default:
				{
					rule.SimpleEle = infoToRule(info, ft.Elem())
				}
			}
		}
	case reflect.Struct:
		{
			switch ft {
			case timeType:
				{
					rule.RuleType = RuleTypeTime
				}
			default:
				{
					panic(fmt.Errorf("a2"))
				}
			}
		}
	case reflect.Pointer:
		{
			ft = ft.Elem()
			if ft == fileType {
				rule.RuleType = RuleTypeFile
			}
			panic(fmt.Errorf("a1"))
		}
	default:
		{
			panic(fmt.Errorf("unexpect type"))
		}
	}
	return rule
}

func GetRules(t reflect.Type) *Rules {
	if c, ok := cache[t]; ok {
		return c
	}

	if t.Kind() != reflect.Struct {
		panic(fmt.Errorf("`%s` is not a struct type", t))
	}

	tm := mapper.TypeMap(t)
	rules := &Rules{
		Gotype: t,
	}
	cache[t] = rules
	for _, info := range tm.Index {
		rules.Data = append(rules.Data, infoToRule(info, nil))
	}
	return rules
}
