package vld

import (
	"github.com/zzztttkkk/0.0/internal/utils"
	"math"
	"mime/multipart"
	"reflect"
)

var (
	mapper = utils.NewMapper("vld")
	cache  = make(map[reflect.Type]*Rules)

	fileType = reflect.TypeOf((*multipart.FileHeader)(nil)).Elem()
)

func infoToRule(info *utils.FieldInfo, ft reflect.Type) *Rule {
	rule := &Rule{
		Name:  info.Name,
		Index: info.Index,
	}

	if ft == nil {
		ft = info.Field.Type
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
			rule.Ele = infoToRule(info, ft.Elem())
		}
	case reflect.Struct:
		{
			if ft == fileType {
				rule.RuleType = RuleTypeFile
			}
		}
	}

	return rule
}

func GetRules(t reflect.Type) {
	tm := mapper.TypeMap(t)
	rules := &Rules{}
	cache[t] = rules
	for _, info := range tm.Index {
		rules.Data = append(rules.Data, infoToRule(info, nil))
	}
}
