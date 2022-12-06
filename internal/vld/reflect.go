package vld

import (
	"fmt"
	"github.com/zzztttkkk/0.0/internal/utils"
	"math"
	"mime/multipart"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	mapper = utils.NewMapper("vld")
	cache  = make(map[reflect.Type]*Rules)

	fileType  = reflect.TypeOf((*multipart.FileHeader)(nil)).Elem()
	timeType  = reflect.TypeOf((*time.Time)(nil)).Elem()
	vlderType = reflect.TypeOf((*Vlder)(nil)).Elem()

	regexps = make(map[string]*regexp.Regexp)
)

func stringToIntRange(v string) (*int64, *int64, bool) {
	if !strings.Contains(v, "-") {
		return nil, nil, false
	}

	parts := strings.Split(v, "-")
	if len(parts) != 2 {
		return nil, nil, false
	}

	var (
		minp *int64
		maxp *int64
	)

	sv := strings.TrimSpace(parts[0])
	if len(sv) > 0 {
		num, err := strconv.ParseInt(sv, 10, 64)
		if err != nil {
			return nil, nil, false
		}
		minp = new(int64)
		*minp = num
	}

	sv = strings.TrimSpace(parts[1])
	if len(sv) > 0 {
		num, err := strconv.ParseInt(sv, 10, 64)
		if err != nil {
			return nil, nil, false
		}
		maxp = new(int64)
		*maxp = num
	}

	if maxp == nil && minp == nil {
		return nil, nil, false
	}
	return minp, maxp, true
}

func stringToFloatRange(v string) (*float64, *float64, bool) {
	if !strings.Contains(v, "-") {
		return nil, nil, false
	}

	parts := strings.Split(v, "-")
	if len(parts) != 2 {
		return nil, nil, false
	}

	var (
		minp *float64
		maxp *float64
	)

	sv := strings.TrimSpace(parts[0])
	if len(sv) > 0 {
		num, err := strconv.ParseFloat(sv, 64)
		if err != nil {
			return nil, nil, false
		}
		minp = new(float64)
		*minp = num
	}

	sv = strings.TrimSpace(parts[1])
	if len(sv) > 0 {
		num, err := strconv.ParseFloat(sv, 64)
		if err != nil {
			return nil, nil, false
		}
		maxp = new(float64)
		*maxp = num
	}

	if maxp == nil && minp == nil {
		return nil, nil, false
	}
	return minp, maxp, true
}

func infoToRule(info *utils.FieldInfo, ft reflect.Type) (*Rule, error) {
	rule := &Rule{
		Name:   info.Name,
		Index:  info.Index,
		Gotype: info.Field.Type,
	}

	var err error

	defer func() {
		if err != nil {
			return
		}

		for k, v := range info.Options {
			k = strings.ToLower(strings.TrimSpace(k))
			switch k {
			case "numrange":
				{
					if rule.RuleType == RuleTypeInt {
						minp, maxp, ok := stringToIntRange(v)
						if !ok {
							err = fmt.Errorf("bad num range, %s", v)
							return
						}
						rule.MinInt = minp
						rule.MaxInt = maxp
					} else if rule.RuleType == RuleTypeDouble {
						minp, maxp, ok := stringToFloatRange(v)
						if !ok {
							err = fmt.Errorf("bad num range, %s", v)
							return
						}
						rule.MinDouble = minp
						rule.MaxDouble = maxp
					}
				}
			case "lenrange":
				{
					minp, maxp, ok := stringToIntRange(v)
					if !ok {
						err = fmt.Errorf("bad len range, %s", v)
						return
					}

					if minp != nil {
						rule.MinLen = new(int)
						*rule.MinLen = int(*minp)
					}
					if maxp != nil {
						rule.MaxLen = new(int)
						*rule.MaxLen = int(*maxp)
					}
				}
			case "nummax":
				{
					if rule.RuleType == RuleTypeInt {
						num, e := strconv.ParseInt(v, 10, 64)
						if e != nil {
							err = fmt.Errorf("bad nummax, %s", v)
							return
						}
						rule.MaxInt = new(int64)
						*rule.MaxInt = num
					} else if rule.RuleType == RuleTypeDouble {
						num, e := strconv.ParseFloat(v, 64)
						if e != nil {
							err = fmt.Errorf("bad nummax, %s", v)
							return
						}
						rule.MaxDouble = new(float64)
						*rule.MaxDouble = num
					}
				}
			case "nummin":
				{
					if rule.RuleType == RuleTypeInt {
						num, e := strconv.ParseInt(v, 10, 64)
						if e != nil {
							err = fmt.Errorf("bad nummin, %s", v)
							return
						}
						rule.MinInt = new(int64)
						*rule.MinInt = num
					} else if rule.RuleType == RuleTypeDouble {
						num, e := strconv.ParseFloat(v, 64)
						if e != nil {
							err = fmt.Errorf("bad nummin, %s", v)
							return
						}
						rule.MinDouble = new(float64)
						*rule.MinDouble = num
					}
				}
			case "lenmax":
				{
					num, e := strconv.ParseInt(v, 10, 64)
					if e != nil {
						err = fmt.Errorf("bad lenmax, %s", v)
						return
					}
					rule.MaxLen = new(int)
					*rule.MaxLen = int(num)
				}
			case "lenmin":
				{
					num, e := strconv.ParseInt(v, 10, 64)
					if e != nil {
						err = fmt.Errorf("bad lenmin, %s", v)
						return
					}
					rule.MinLen = new(int)
					*rule.MinLen = int(num)
				}
			case "optional":
				{
					rule.Optional = true
				}
			case "regexp":
				{
					ptr := regexps[v]
					if ptr == nil {
						err = fmt.Errorf(`unregister regexp name, %s`, v)
						return
					}
					rule.Regexp = ptr
				}
			case "timelayout":
				{
					now := time.Now()
					nt, e := time.Parse(v, now.Format(v))
					if e != nil || nt.UnixNano() != now.UnixNano() {
						err = fmt.Errorf(`bad time layout, %s`, v)
						return
					}
					rule.TimeLayout = v
				}
			case "timeunit":
				{
					v = strings.ToLower(strings.TrimSpace(v))
					switch v {
					case "", "s":
						rule.TimeUnit = "s"
					case "ms":
						rule.TimeUnit = "ms"
					default:
						err = fmt.Errorf(`bad time unit(""/"s"/"ms"), %s`, v)
						return
					}
				}
			}
		}
	}()

	if ft == nil {
		ft = info.Field.Type
	}

	if ft.Implements(vlderType) {
		rule.RuleType = RuleTypeVlder
		return rule, err
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
			eleRule, eleErr := infoToRule(info, ft.Elem())
			if eleErr != nil {
				err = eleErr
				return nil, err
			}

			*rule = *eleRule
			rule.IsSlice = true
			rule.Gotype = ft.Elem()
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
	return rule, err
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
		rules.Data = append(rules.Data, utils.Must(infoToRule(info, nil)))
	}
	return rules
}
