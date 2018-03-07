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

func parseCriteriaType(t string) int {
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
	panic("invalid criteria type: " + t)
}

type Criteria interface {
}

type ExprCriteria string

type OneColumnCriteria struct {
	Column
	Type  int
	Value interface{}
}

type TwoColumnCriteria struct {
	Left  Column
	Right Column
	Type  int
}

type CriteriaSet interface {
	Empty() bool
}

type JoinCriteriaSet interface {
	CriteriaSet
	Left() CriteriaSet
	Right() CriteriaSet
	Joiner() string
}

type NotCriteriaSet struct {
	Inner CriteriaSet
}

func Not(inner CriteriaSet) CriteriaSet {
	return &NotCriteriaSet{inner}
}

func (c *NotCriteriaSet) Empty() bool {
	return c == nil || c.Inner.Empty()
}

type AndCriteriaSet struct {
	left, right CriteriaSet
}

func And(left, right CriteriaSet) CriteriaSet {
	return &AndCriteriaSet{left: left, right: right}
}

func (c *AndCriteriaSet) Left() CriteriaSet {
	return c.left
}

func (c *AndCriteriaSet) Right() CriteriaSet {
	return c.right
}

func (c *AndCriteriaSet) Empty() bool {
	return c == nil || (c.left == nil && c.right.Empty())
}

func (c *AndCriteriaSet) Joiner() string {
	return "AND"
}

type OrCriteriaSet struct {
	left, right CriteriaSet
}

func Or(left, right CriteriaSet) CriteriaSet {
	return &OrCriteriaSet{left: left, right: right}
}

func (c *OrCriteriaSet) Left() CriteriaSet {
	return c.left
}

func (c *OrCriteriaSet) Right() CriteriaSet {
	return c.right
}

func (c *OrCriteriaSet) Empty() bool {
	return c == nil || (c.left == nil && c.right.Empty())
}

func (c *OrCriteriaSet) Joiner() string {
	return "OR"
}

type SimpleCriteriaSet struct {
	Items []Criteria
}

func Equal(col, val interface{}) *SimpleCriteriaSet {
	set := &SimpleCriteriaSet{}
	set.Equal(col, val)
	return set
}

func (c *SimpleCriteriaSet) Empty() bool {
	return c == nil || len(c.Items) == 0
}

// Wild add a custom expression criteria
//
//     f.Wild("a > (b + 100)")
func (c *SimpleCriteriaSet) Wild(expr string) *SimpleCriteriaSet {
	c.Items = append(c.Items, ExprCriteria(expr))
	return c
}

// Like add a LIKE criteria
//
//     c.Like("name", "%abc")
//     c.Like("name", "abc_")
func (c *SimpleCriteriaSet) Like(col interface{}, expr string) *SimpleCriteriaSet {
	return c.add(col, LK, expr)
}

func (c *SimpleCriteriaSet) LikeIf(when bool, col interface{}, expr string) *SimpleCriteriaSet {
	return c.addIf(when, col, LK, expr)
}

func (c *SimpleCriteriaSet) Equal(col, val interface{}) *SimpleCriteriaSet {
	return c.add(col, EQ, val)
}

func (c *SimpleCriteriaSet) EqualIf(when bool, col, val interface{}) *SimpleCriteriaSet {
	return c.addIf(when, col, EQ, val)
}

func (c *SimpleCriteriaSet) Equal2(left, right interface{}) *SimpleCriteriaSet {
	return c.add2(left, right, EQ)
}

func (c *SimpleCriteriaSet) NotEqual(col, val interface{}) *SimpleCriteriaSet {
	return c.add(col, NE, val)
}

func (c *SimpleCriteriaSet) NotEqualIf(when bool, col, val interface{}) *SimpleCriteriaSet {
	return c.addIf(when, col, NE, val)
}

func (c *SimpleCriteriaSet) NotEqual2(left, right interface{}) *SimpleCriteriaSet {
	return c.add2(left, right, NE)
}

func (c *SimpleCriteriaSet) In(col, val interface{}) *SimpleCriteriaSet {
	return c.add(col, IN, val)
}

func (c *SimpleCriteriaSet) InIf(when bool, col, val interface{}) *SimpleCriteriaSet {
	return c.addIf(when, col, IN, val)
}

func (c *SimpleCriteriaSet) NotIn(col, val interface{}) *SimpleCriteriaSet {
	return c.add(col, NIN, val)
}

func (c *SimpleCriteriaSet) NotInIf(when bool, col, val interface{}) *SimpleCriteriaSet {
	return c.addIf(when, col, NIN, val)
}

func (c *SimpleCriteriaSet) Less(col, val interface{}) *SimpleCriteriaSet {
	return c.add(col, LT, val)
}

func (c *SimpleCriteriaSet) LessIf(when bool, col, val interface{}) *SimpleCriteriaSet {
	return c.addIf(when, col, LT, val)
}

func (c *SimpleCriteriaSet) Less2(left, right interface{}) *SimpleCriteriaSet {
	return c.add2(left, right, LT)
}

func (c *SimpleCriteriaSet) LessOrEqual(col, val interface{}) *SimpleCriteriaSet {
	return c.add(col, LTE, val)
}

func (c *SimpleCriteriaSet) LessOrEqualIf(when bool, col, val interface{}) *SimpleCriteriaSet {
	return c.addIf(when, col, LTE, val)
}

func (c *SimpleCriteriaSet) LessOrEqual2(left, right interface{}) *SimpleCriteriaSet {
	return c.add2(left, right, LTE)
}

func (c *SimpleCriteriaSet) Greater(col, val interface{}) *SimpleCriteriaSet {
	return c.add(col, GT, val)
}

func (c *SimpleCriteriaSet) GreaterIf(when bool, col, val interface{}) *SimpleCriteriaSet {
	return c.addIf(when, col, GT, val)
}

func (c *SimpleCriteriaSet) Greater2(left, right interface{}) *SimpleCriteriaSet {
	return c.add2(left, right, GT)
}

func (c *SimpleCriteriaSet) GreaterOrEqual(col, val interface{}) *SimpleCriteriaSet {
	return c.add(col, GTE, val)
}

func (c *SimpleCriteriaSet) GreaterOrEqualIf(when bool, col, val interface{}) *SimpleCriteriaSet {
	return c.addIf(when, col, GTE, val)
}

func (c *SimpleCriteriaSet) GreaterOrEqual2(left, right interface{}) *SimpleCriteriaSet {
	return c.add2(left, right, GTE)
}

func (c *SimpleCriteriaSet) Between(col, start, end interface{}) *SimpleCriteriaSet {
	return c.add(c.toColumn(col), GTE, start).add(c, LTE, end)
}

func (c *SimpleCriteriaSet) Prefix(col interface{}, s string) *SimpleCriteriaSet {
	return c.Like(col, s+"%")
}

func (c *SimpleCriteriaSet) Suffix(col interface{}, s string) *SimpleCriteriaSet {
	return c.Like(col, "%"+s)
}

func (c *SimpleCriteriaSet) Contains(col interface{}, s string) *SimpleCriteriaSet {
	return c.Like(col, "%"+s+"%")
}

func (c *SimpleCriteriaSet) Date(col interface{}, date time.Time) *SimpleCriteriaSet {
	column := c.toColumn(col)
	start := times.Date(date)
	end := start.AddDate(0, 0, 1)
	return c.GreaterOrEqual(column, start).Less(c, end)
}

func (c *SimpleCriteriaSet) DateRange(col interface{}, start, end time.Time) *SimpleCriteriaSet {
	return c.GreaterOrEqual(c.toColumn(col), times.Date(start)).Less(c, times.Date(end).AddDate(0, 0, 1))
}

func (c *SimpleCriteriaSet) add(col interface{}, t int, val interface{}) *SimpleCriteriaSet {
	c.Items = append(c.Items, &OneColumnCriteria{
		Column: c.toColumn(col),
		Type:   t,
		Value:  val,
	})
	return c
}

func (c *SimpleCriteriaSet) addIf(when bool, col interface{}, t int, val interface{}) *SimpleCriteriaSet {
	if when {
		return c.add(col, t, val)
	}
	return c
}

func (c *SimpleCriteriaSet) add2(left, right interface{}, t int) *SimpleCriteriaSet {
	c.Items = append(c.Items, &TwoColumnCriteria{
		Left:  c.toColumn(left),
		Type:  t,
		Right: c.toColumn(right),
	})
	return c
}

func (c *SimpleCriteriaSet) toColumn(i interface{}) (col Column) {
	switch v := i.(type) {
	case string:
		col = SimpleColumn(v)
	case Column:
		col = v
	default:
		panic(errors.Format("not a valid Column: %#v", i))
	}
	return
}
