package vld

import (
	"errors"
	"github.com/zzztttkkk/0.0/internal/utils"
	"html"
	"mime/multipart"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type RuleType int

const (
	RuleTypeInt = RuleType(iota)
	RuleTypeDouble
	RuleTypeBool
	RuleTypeString
	RuleTypeFile
	RuleTypeTime
	RuleTypeVlder
)

type Rule struct {
	Name     string
	RuleType RuleType
	Gotype   reflect.Type
	IsSlice  bool
	Index    []int

	Optional bool

	MaxInt    *int64
	MinInt    *int64
	MaxDouble *float64
	MinDouble *float64

	MaxLen *int
	MinLen *int

	MaxRuneCount *int
	MinRuneCount *int
	NoTrim       bool
	NoEscape     bool
	Regexp       *regexp.Regexp

	TimeLayout string
	TimeUnit   string
}

func (rule *Rule) intOk(num int64, err *Error) bool {
	if rule.MinInt != nil && num < *rule.MinInt {
		err.Reason = ErrorReasonNumOutOfRange
		return false
	}

	if rule.MaxInt != nil && num > *rule.MaxInt {
		err.Reason = ErrorReasonNumOutOfRange
		return false
	}
	return true
}

func (rule *Rule) floatOk(num float64, err *Error) bool {
	if rule.MinDouble != nil && num < *rule.MinDouble {
		err.Reason = ErrorReasonNumOutOfRange
		return false
	}

	if rule.MaxDouble != nil && num > *rule.MaxDouble {
		err.Reason = ErrorReasonNumOutOfRange
		return false
	}
	return true
}

func (rule *Rule) string2Int(v string, ep *Error) (any, bool) {
	num, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		ep.Reason = ErrorReasonCanNotCastToNum
		return nil, false
	}

	if !rule.intOk(num, ep) {
		return nil, false
	}

	switch rule.Gotype.Kind() {
	case reflect.Int:
		return int(num), true
	case reflect.Int8:
		return int8(num), true
	case reflect.Int16:
		return int16(num), true
	case reflect.Int32:
		return int32(num), true
	case reflect.Int64:
		return num, true
	case reflect.Uint:
		return uint(num), true
	case reflect.Uint8:
		return uint8(num), true
	case reflect.Uint16:
		return uint16(num), true
	case reflect.Uint32:
		return uint32(num), true
	case reflect.Uint64:
		return uint64(num), true
	}
	return num, true
}

func (rule *Rule) string2Double(v string, ep *Error) (any, bool) {
	num, err := strconv.ParseFloat(v, 10)
	if err != nil {
		ep.Reason = ErrorReasonCanNotCastToNum
		return nil, false
	}

	if !rule.floatOk(num, ep) {
		return nil, false
	}

	switch rule.Gotype.Kind() {
	case reflect.Float32:
		return float32(num), true
	default:
		return num, true
	}
}

func (rule *Rule) string2Time(v string, ep *Error) (any, bool) {
	if len(rule.TimeLayout) > 0 {
		t, e := time.Parse(rule.TimeLayout, v)
		if e != nil {
			ep.Reason = ErrorReasonBadTimeValue
			return nil, false
		}
		return t, true
	}

	num, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		ep.Reason = ErrorReasonNumOutOfRange
		return nil, false
	}

	switch rule.TimeUnit {
	case "", "s":
		return time.Unix(num, 0), true
	default:
		return time.UnixMilli(num), true
	}
}

func (rule *Rule) stringOk(v string, ep *Error) bool {
	var runeCount = -1
	if rule.MaxRuneCount != nil {
		runeCount = utf8.RuneCount(utils.B(v))
		if runeCount > *rule.MaxRuneCount {
			ep.Reason = ErrorReasonLengthOutOfRange
			return false
		}
	}

	if rule.MinRuneCount != nil {
		if runeCount < 0 {
			runeCount = utf8.RuneCount(utils.B(v))
		}
		if runeCount < *rule.MinRuneCount {
			ep.Reason = ErrorReasonLengthOutOfRange
			return false
		}
	}

	if rule.Regexp != nil && !rule.Regexp.Match(utils.B(v)) {
		ep.Reason = ErrorReasonNotMatchRegexp
		return false
	}
	return true
}

func (rule *Rule) sliceLenOk(lenv int, ep *Error) bool {
	if rule.MinLen != nil && lenv < *rule.MinLen {
		ep.Reason = ErrorReasonLengthOutOfRange
		return false
	}
	if rule.MaxLen != nil && lenv > *rule.MaxLen {
		ep.Reason = ErrorReasonLengthOutOfRange
		return false
	}
	return true
}

func (rule *Rule) bindAndValidateSingleSimpleEle(raw string, ep *Error) (any, bool) {
	switch rule.RuleType {
	case RuleTypeString:
		{
			if !rule.NoTrim {
				raw = strings.TrimSpace(raw)
			}

			if !rule.NoEscape {
				raw = html.EscapeString(raw)
			}

			if rule.stringOk(raw, ep) {
				return raw, true
			}
			return nil, false
		}
	case RuleTypeInt:
		{
			return rule.string2Int(raw, ep)
		}
	case RuleTypeDouble:
		{
			return rule.string2Double(raw, ep)
		}
	case RuleTypeBool:
		{
			bol, err := strconv.ParseBool(raw)
			if err != nil {
				ep.Reason = ErrorReasonCanNotCastToBool
				return nil, false
			}
			return bol, true
		}
	case RuleTypeTime:
		{
			return rule.string2Time(raw, ep)
		}
	}
	panic(errors.New("unreachable error"))
}

func (rule *Rule) one(v string, ep *Error) (any, bool) {
	switch rule.RuleType {
	case RuleTypeVlder:
		{
			val := reflect.New(rule.Gotype)
			ok := val.Interface().(Vlder).FromString(v, ep)
			if !ok {
				return nil, false
			}
			if rule.Gotype.Kind() == reflect.Pointer {
				return val.Interface(), true
			}
			return val.Elem().Interface(), true
		}
	default:
		{
			return rule.bindAndValidateSingleSimpleEle(v, ep)
		}
	}
}

func getFiles(req *http.Request, name string) []*multipart.FileHeader {
	_ = req.ParseForm()
	if req.MultipartForm == nil || req.MultipartForm.File == nil {
		return nil
	}
	return req.MultipartForm.File[name]
}

func (rule *Rule) get(req *http.Request, ep *Error) (any, bool) {
	if rule.IsSlice {
		if rule.RuleType == RuleTypeFile {
			fhs := getFiles(req, rule.Name)
			if len(fhs) < 1 {
				if rule.Optional {
					return nil, true
				}
				ep.Reason = ErrorReasonMissRequired
				return nil, false
			}

			if !rule.sliceLenOk(len(fhs), ep) {
				return nil, false
			}

			nfhs := make([]*multipart.FileHeader, len(fhs), len(fhs))
			copy(nfhs, fhs)
			return nfhs, true
		} else {
			_ = req.ParseForm()
			svs := req.Form[rule.Name]
			if len(svs) < 1 {
				if rule.Optional {
					return nil, true
				}
				ep.Reason = ErrorReasonMissRequired
				return nil, false
			}

			if !rule.sliceLenOk(len(svs), ep) {
				return nil, false
			}

			sliceVal := reflect.MakeSlice(reflect.SliceOf(rule.Gotype), 0, len(svs))

			for _, sv := range svs {
				ele, ok := rule.one(sv, ep)
				if !ok {
					return nil, false
				}
				sliceVal = reflect.Append(sliceVal, reflect.ValueOf(ele))
			}
			return sliceVal.Interface(), true
		}
	}

	if rule.RuleType == RuleTypeFile {
		fhs := getFiles(req, rule.Name)
		if len(fhs) < 1 {
			if rule.Optional {
				return nil, true
			}
			ep.Reason = ErrorReasonMissRequired
			return nil, false
		}
		return fhs[0], true
	}

	sv := req.FormValue(rule.Name)
	if len(sv) < 1 {
		if rule.Optional {
			return nil, true
		}
		ep.Reason = ErrorReasonMissRequired
		return nil, false
	}
	return rule.one(sv, ep)
}

type Rules struct {
	Gotype reflect.Type
	Data   []*Rule
}

func (rules *Rules) BindAndValidate(req *http.Request) (any, error) {
	val := reflect.New(rules.Gotype).Elem()
	var err Error
	for _, rule := range rules.Data {
		v, ok := rule.get(req, &err)
		if !ok {
			err.PkgPath = rules.Gotype.PkgPath()
			err.TypeName = rules.Gotype.Name()
			err.Rule = rule
			return nil, &err
		}
		if v == nil {
			continue
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
