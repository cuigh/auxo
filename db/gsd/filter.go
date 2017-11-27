package gsd

import (
	"strings"
	"time"

	"github.com/cuigh/auxo/errors"
	"github.com/cuigh/auxo/ext/times"
)

const (
	EQ = iota
	NE
	LT
	LTE
	GT
	GTE
	IN
	NIN
	LK
)

func parseFilterType(t string) int {
	switch strings.ToLower(t) {
	case "", "eq":
		return EQ
	case "ne":
		return NE
	case "lt":
		return LT
	case "lte":
		return LTE
	case "gt":
		return GT
	case "gte":
		return GTE
	case "in":
		return IN
	case "nin":
		return NIN
	case "lk":
		return LK
	}
	panic("invalid filter type: " + t)
}

type Filter interface {
}

type ExprFilter string

type OneColumnFilter struct {
	Column
	Type  int
	Value interface{}
}

type TwoColumnFilter struct {
	Left  Column
	Right Column
	Type  int
}

type Filters interface {
	Empty() bool
}

type JoinFilters interface {
	Filters
	Left() Filters
	Right() Filters
	Joiner() string
}

type NotFilters struct {
	Inner Filters
}

//func NotFilters(inner Filters) Filters {
//	return &NotFilters{left: left, right: right}
//}

func Not(inner Filters) Filters {
	return &NotFilters{inner}
}

func (f *NotFilters) Empty() bool {
	return f == nil || f.Inner.Empty()
}

type AndFilters struct {
	left, right Filters
}

//func NewAndFilters(left, right Filters) Filters {
//	return &AndFilters{left: left, right: right}
//}

func And(left, right Filters) Filters {
	return &AndFilters{left: left, right: right}
}

func (f *AndFilters) Left() Filters {
	return f.left
}

func (f *AndFilters) Right() Filters {
	return f.right
}

func (f *AndFilters) Empty() bool {
	return f == nil || (f.left == nil && f.right.Empty())
}

func (f *AndFilters) Joiner() string {
	return "AND"
}

type OrFilters struct {
	left, right Filters
}

//func NewOrFilters(left, right Filters) Filters {
//	return &OrFilters{left: left, right: right}
//}

func Or(left, right Filters) Filters {
	return &OrFilters{left: left, right: right}
}

func (f *OrFilters) Left() Filters {
	return f.left
}

func (f *OrFilters) Right() Filters {
	return f.right
}

func (f *OrFilters) Empty() bool {
	return f == nil || (f.left == nil && f.right.Empty())
}

func (f *OrFilters) Joiner() string {
	return "OR"
}

type SimpleFilters []Filter

func NewFilters() *SimpleFilters {
	return &SimpleFilters{}
}

func Equal(col, val interface{}) *SimpleFilters {
	return NewFilters().Equal(col, val)
}

func (f *SimpleFilters) Empty() bool {
	return f == nil || len(*f) == 0
}

// Wild add a custom expression filter
//
//     f.Wild("a > (b + 100)")
func (f *SimpleFilters) Wild(expr string) *SimpleFilters {
	*f = append(*f, ExprFilter(expr))
	return f
}

// Like add a LIKE filter
//
//     f.Like("name", "%abc")
//     f.Like("name", "abc_")
func (f *SimpleFilters) Like(col interface{}, expr string) *SimpleFilters {
	return f.add(col, LK, expr)
}

func (f *SimpleFilters) LikeIf(when bool, col interface{}, expr string) *SimpleFilters {
	return f.addIf(when, col, LK, expr)
}

func (f *SimpleFilters) Equal(col, val interface{}) *SimpleFilters {
	return f.add(col, EQ, val)
}

func (f *SimpleFilters) EqualIf(when bool, col, val interface{}) *SimpleFilters {
	return f.addIf(when, col, EQ, val)
}

func (f *SimpleFilters) Equal2(left, right interface{}) *SimpleFilters {
	return f.add2(left, right, EQ)
}

func (f *SimpleFilters) NotEqual(col, val interface{}) *SimpleFilters {
	return f.add(col, NE, val)
}

func (f *SimpleFilters) NotEqualIf(when bool, col, val interface{}) *SimpleFilters {
	return f.addIf(when, col, NE, val)
}

func (f *SimpleFilters) NotEqual2(left, right interface{}) *SimpleFilters {
	return f.add2(left, right, NE)
}

func (f *SimpleFilters) In(col, val interface{}) *SimpleFilters {
	return f.add(col, IN, val)
}

func (f *SimpleFilters) InIf(when bool, col, val interface{}) *SimpleFilters {
	return f.addIf(when, col, IN, val)
}

func (f *SimpleFilters) NotIn(col, val interface{}) *SimpleFilters {
	return f.add(col, NIN, val)
}

func (f *SimpleFilters) NotInIf(when bool, col, val interface{}) *SimpleFilters {
	return f.addIf(when, col, NIN, val)
}

func (f *SimpleFilters) Less(col, val interface{}) *SimpleFilters {
	return f.add(col, LT, val)
}

func (f *SimpleFilters) LessIf(when bool, col, val interface{}) *SimpleFilters {
	return f.addIf(when, col, LT, val)
}

func (f *SimpleFilters) Less2(left, right interface{}) *SimpleFilters {
	return f.add2(left, right, LT)
}

func (f *SimpleFilters) LessOrEqual(col, val interface{}) *SimpleFilters {
	return f.add(col, LTE, val)
}

func (f *SimpleFilters) LessOrEqualIf(when bool, col, val interface{}) *SimpleFilters {
	return f.addIf(when, col, LTE, val)
}

func (f *SimpleFilters) LessOrEqual2(left, right interface{}) *SimpleFilters {
	return f.add2(left, right, LTE)
}

func (f *SimpleFilters) Greater(col, val interface{}) *SimpleFilters {
	return f.add(col, GT, val)
}

func (f *SimpleFilters) GreaterIf(when bool, col, val interface{}) *SimpleFilters {
	return f.addIf(when, col, GT, val)
}

func (f *SimpleFilters) Greater2(left, right interface{}) *SimpleFilters {
	return f.add2(left, right, GT)
}

func (f *SimpleFilters) GreaterOrEqual(col, val interface{}) *SimpleFilters {
	return f.add(col, GTE, val)
}

func (f *SimpleFilters) GreaterOrEqualIf(when bool, col, val interface{}) *SimpleFilters {
	return f.addIf(when, col, GTE, val)
}

func (f *SimpleFilters) GreaterOrEqual2(left, right interface{}) *SimpleFilters {
	return f.add2(left, right, GTE)
}

func (f *SimpleFilters) Between(col, start, end interface{}) *SimpleFilters {
	c := f.toColumn(col)
	return f.add(c, GTE, start).add(c, LTE, end)
}

func (f *SimpleFilters) Prefix(col interface{}, s string) *SimpleFilters {
	return f.Like(col, s+"%")
}

func (f *SimpleFilters) Suffix(col interface{}, s string) *SimpleFilters {
	return f.Like(col, "%"+s)
}

func (f *SimpleFilters) Contains(col interface{}, s string) *SimpleFilters {
	return f.Like(col, "%"+s+"%")
}

func (f *SimpleFilters) Date(col interface{}, date time.Time) *SimpleFilters {
	c := f.toColumn(col)
	start := times.Date(date)
	end := start.AddDate(0, 0, 1)
	return f.GreaterOrEqual(c, start).Less(c, end)
}

func (f *SimpleFilters) DateRange(col interface{}, start, end time.Time) *SimpleFilters {
	c := f.toColumn(col)
	return f.GreaterOrEqual(c, times.Date(start)).Less(c, times.Date(end).AddDate(0, 0, 1))
}

func (f *SimpleFilters) add(col interface{}, t int, val interface{}) *SimpleFilters {
	*f = append(*f, &OneColumnFilter{
		Column: f.toColumn(col),
		Type:   EQ,
		Value:  val,
	})
	return f
}

func (f *SimpleFilters) addIf(when bool, col interface{}, t int, val interface{}) *SimpleFilters {
	if when {
		return f.add(col, t, val)
	}
	return f
}

func (f *SimpleFilters) add2(left, right interface{}, t int) *SimpleFilters {
	*f = append(*f, &TwoColumnFilter{
		Left:  f.toColumn(left),
		Type:  EQ,
		Right: f.toColumn(right),
	})
	return f
}

func (f *SimpleFilters) toColumn(i interface{}) (c Column) {
	switch v := i.(type) {
	case string:
		c = SimpleColumn(v)
	case Column:
		c = v
	default:
		panic(errors.Format("not a valid Column: %#v", i))
	}
	return
}
