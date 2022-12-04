package vld

type RuleType int

const (
	RuleTypeInt = RuleType(iota)
	RuleTypeDouble
	RuleTypeBool
	RuleTypeString
	RuleTypeFile
	RuleTypeSlice
)

type Rule struct {
	Name      string
	RuleType  RuleType
	Optional  bool
	Index     []int
	IsSlice   bool
	IsPtr     bool
	MaxInt    *int64
	MinInt    *int64
	MaxDouble *float64
	MinDouble *float64
	Ele       *Rule
}

type Rules struct {
	Data []*Rule
}
