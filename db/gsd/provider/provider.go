package provider

import (
	"reflect"
	"strconv"

	"github.com/cuigh/auxo/db/gsd"
	"github.com/cuigh/auxo/errors"
	"github.com/cuigh/auxo/ext/reflects"
)

const (
	Comma = ','
	Dot   = '.'
)

var insertValueClauses = [...]string{
	"(?)",
	"(?,?)",
	"(?,?,?)",
	"(?,?,?,?)",
	"(?,?,?,?,?)",
	"(?,?,?,?,?,?)",
	"(?,?,?,?,?,?,?)",
	"(?,?,?,?,?,?,?,?)",
	"(?,?,?,?,?,?,?,?,?)",
	"(?,?,?,?,?,?,?,?,?,?)", // 10
	"(?,?,?,?,?,?,?,?,?,?,?)",
	"(?,?,?,?,?,?,?,?,?,?,?,?)",
	"(?,?,?,?,?,?,?,?,?,?,?,?,?)",
	"(?,?,?,?,?,?,?,?,?,?,?,?,?,?)",
	"(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",
	"(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",
	"(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",
	"(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",
	"(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)",
	"(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", // 20
}

type Provider struct {
	Quote func(b *gsd.Builder, s string)
	Limit func(b *gsd.Builder, skip, take int)
	Call  func(b *gsd.Builder, sp string, args ...interface{}) error
}

func (p *Provider) BuildCall(b *gsd.Builder, info *gsd.CallInfo) (err error) {
	return p.Call(b, info.SP, info.Args)
}

func (p *Provider) BuildInsert(b *gsd.Builder, info *gsd.InsertInfo) (err error) {
	cols := len(info.Columns)
	b.WriteString("INSERT INTO ")
	p.Quote(b, info.Table)
	b.WriteByte('(')
	p.Quote(b, info.Columns[0])
	for i := 1; i < cols; i++ {
		b.WriteByte(Comma)
		p.Quote(b, info.Columns[i])
	}
	b.WriteString(") VALUES ")

	rows := len(info.Values) / cols
	for i := 0; i < rows; i++ {
		if i > 0 {
			b.WriteByte(Comma)
		}
		if cols > 20 {
			// todo: optimize
			b.Write([]byte{'(', '?'})
			for j := 1; j < cols; j++ {
				b.Write([]byte{',', '?'})
			}
			b.WriteByte(')')
		} else {
			b.WriteString(insertValueClauses[cols-1])
		}
	}
	b.Args = info.Values
	return
}

func (p *Provider) BuildDelete(b *gsd.Builder, info *gsd.DeleteInfo) (err error) {
	b.WriteString("DELETE FROM ")
	p.Quote(b, info.Table)
	if info.Where == nil || info.Where.Empty() {
		return errors.New("delete action must have where clause")
	}
	b.WriteString(" WHERE ")
	return p.buildCriteriaSet(b, info.Where)
}

func (p *Provider) BuildUpdate(b *gsd.Builder, info *gsd.UpdateInfo) (err error) {
	b.WriteString("UPDATE ")
	p.Quote(b, info.Table)
	b.WriteString(" SET ")
	for i, col := range info.Columns {
		if i > 0 {
			b.WriteByte(Comma)
		}
		p.Quote(b, col)
		b.WriteByte('=')
		switch v := info.Values[i].(type) {
		case gsd.IncValue:
			p.Quote(b, col)
			b.Write([]byte{'+', '?'})
			b.Args = append(b.Args, v.Value)
		case gsd.DecValue:
			p.Quote(b, col)
			b.Write([]byte{'-', '?'})
			b.Args = append(b.Args, v.Value)
		case gsd.ExprValue:
			b.WriteString(v.Value)
		default:
			b.WriteByte('?')
			b.Args = append(b.Args, v)
		}
	}
	if info.Where != nil && !info.Where.Empty() {
		b.WriteString(" WHERE ")
		err = p.buildCriteriaSet(b, info.Where)
	}
	return
}

func (p *Provider) BuildSelect(b *gsd.Builder, info *gsd.SelectInfo) (err error) {
	// SELECT
	b.WriteString("SELECT ")
	if info.Distinct {
		b.WriteString("DISTINCT ")
	}
	for i, col := range info.Columns {
		if i > 0 {
			b.WriteByte(',')
		}
		if col.Table() == nil {
			if _, ok := col.(*gsd.ExprColumn); ok {
				b.WriteString(col.Name())
			} else {
				p.Quote(b, col.Name())
			}
		} else {
			p.Quote(b, col.Table().Prefix())
			b.WriteByte(Dot)
			p.Quote(b, col.Name())
		}
		if col.Alias() != "" {
			b.WriteString(" AS ", col.Alias())
		}
	}

	// FROM
	b.WriteString(" FROM ")
	p.Quote(b, info.Table.Name())
	if info.Table.Alias() != "" {
		b.WriteString(" AS ", info.Table.Alias())
	}

	// JOIN
	if err = p.buildJoin(b, info); err != nil {
		return
	}

	// WHERE
	if info.Where != nil && !info.Where.Empty() {
		b.WriteString(" WHERE ")
		if err = p.buildCriteriaSet(b, info.Where); err != nil {
			return
		}
	}

	// GROUP BY
	if err = p.buildGroupBy(b, info); err != nil {
		return
	}

	// ORDER BY
	if len(info.Orders) > 0 {
		b.WriteString(" ORDER BY ")
		for i, order := range info.Orders {
			if i > 0 {
				b.WriteByte(Comma)
			}
			for j, col := range order.Columns {
				if j > 0 {
					b.WriteByte(Comma)
				}
				p.buildColumn(b, col)
			}
			if order.Type == gsd.DESC {
				b.WriteString(" DESC")
			}
		}
	}

	// LIMIT
	if info.Skip != 0 || info.Take != 0 {
		p.Limit(b, info.Skip, info.Take)
	}

	// LOCK

	// TOTAL COUNT
	if info.Count {
		err = p.buildCount(b, info)
	}
	return
}

func (p *Provider) buildCount(b *gsd.Builder, info *gsd.SelectInfo) (err error) {
	// SELECT FROM
	b.WriteString(";SELECT COUNT(0) FROM ")
	p.Quote(b, info.Table.Name())
	if info.Table.Alias() != "" {
		b.WriteString(" AS ", info.Table.Alias())
	}

	// JOIN
	if err = p.buildJoin(b, info); err != nil {
		return
	}

	// WHERE
	if info.Where != nil && !info.Where.Empty() {
		b.WriteString(" WHERE ")
		if err = p.buildCriteriaSet(b, info.Where); err != nil {
			return err
		}
	}

	// GROUP BY
	return p.buildGroupBy(b, info)
}

func (p *Provider) buildJoin(b *gsd.Builder, info *gsd.SelectInfo) (err error) {
	for _, join := range info.Joins {
		b.WriteString(" ", join.Type, " ")
		p.Quote(b, join.Table.Name())
		if join.Table.Alias() != "" {
			b.WriteString(" AS ", join.Table.Alias())
		}
		b.WriteString(" ON ")
		if err = p.buildCriteriaSet(b, join.On); err != nil {
			return
		}
	}
	return
}

func (p *Provider) buildGroupBy(b *gsd.Builder, info *gsd.SelectInfo) (err error) {
	if len(info.Groups) > 0 {
		b.WriteString(" GROUP BY ")
		for i, col := range info.Groups {
			if i > 0 {
				b.WriteByte(',')
			}
			p.buildColumn(b, col)
		}
		if info.Having != nil {
			b.WriteString(" HAVING ")
			if err = p.buildCriteriaSet(b, info.Having); err != nil {
				return
			}
		}
	}
	return
}

func (p *Provider) buildColumn(b *gsd.Builder, col gsd.Column) {
	if col.Table() == nil {
		p.Quote(b, col.Name())
	} else {
		p.Quote(b, col.Table().Prefix())
		b.WriteByte(Dot)
		p.Quote(b, col.Name())
	}
}

func (p *Provider) buildCriteriaSet(b *gsd.Builder, cs gsd.CriteriaSet) (err error) {
	switch fs := cs.(type) {
	case *gsd.SimpleCriteriaSet:
		for i, c := range fs.Items {
			if i > 0 {
				b.WriteString(" AND ")
			}
			switch f := c.(type) {
			case *gsd.OneColumnCriteria:
				if err = p.buildOneColumnCriteria(b, f); err != nil {
					return
				}
			case *gsd.TwoColumnCriteria:
				if err = p.buildTwoColumnCriteria(b, f); err != nil {
					panic(err)
				}
			case gsd.ExprCriteria:
				b.WriteString(string(f))
			}
		}
	case *gsd.NotCriteriaSet:
		b.WriteString("NOT(")
		if err = p.buildCriteriaSet(b, fs.Inner); err != nil {
			return
		}
		b.WriteByte(')')
	case gsd.JoinCriteriaSet:
		if fs.Left().Empty() {
			if err = p.buildCriteriaSet(b, fs.Right()); err != nil {
				return
			}
		} else if fs.Right().Empty() {
			if err = p.buildCriteriaSet(b, fs.Left()); err != nil {
				return
			}
		} else {
			b.WriteByte('(')
			if err = p.buildCriteriaSet(b, fs.Left()); err != nil {
				return
			}
			b.WriteString(") ", fs.Joiner(), " (")
			if err = p.buildCriteriaSet(b, fs.Right()); err != nil {
				return
			}
			b.WriteByte(')')
		}
	}
	return
}

func (p *Provider) buildOneColumnCriteria(b *gsd.Builder, c *gsd.OneColumnCriteria) (err error) {
	if c.Table() != nil {
		p.Quote(b, c.Table().Prefix())
		b.WriteByte(Dot)
	}

	switch c.Type {
	case gsd.NE:
		if c.Value == nil {
			p.Quote(b, c.Name())
			b.WriteString(" IS NOT NULL")
			return
		}

		p.Quote(b, c.Name())
		b.WriteString("<>?")
	case gsd.LT:
		p.Quote(b, c.Name())
		b.WriteString("<?")
	case gsd.LTE:
		p.Quote(b, c.Name())
		b.WriteString("<=?")
	case gsd.GT:
		p.Quote(b, c.Name())
		b.WriteString(">?")
	case gsd.GTE:
		p.Quote(b, c.Name())
		b.WriteString(">=?")
	case gsd.IN:
		p.Quote(b, c.Name())
		b.WriteString(" IN(")
		if err = p.buildInValues(b, c.Value); err != nil {
			return
		}
		b.WriteByte(')')
		return
	case gsd.NIN:
		p.Quote(b, c.Name())
		b.WriteString(" NOT IN(")
		if err = p.buildInValues(b, c.Value); err != nil {
			return
		}
		b.WriteByte(')')
		return
	case gsd.LK:
		p.Quote(b, c.Name())
		b.WriteString(" LIKE '", c.Value.(string), "'")
		return
	default:
		if c.Value == nil {
			p.Quote(b, c.Name())
			b.WriteString(" IS NULL")
			return
		}

		p.Quote(b, c.Name())
		b.WriteString("=?")
	}
	b.Args = append(b.Args, c.Value)
	return
}

func (p *Provider) buildInValues(b *gsd.Builder, i interface{}) (err error) {
	t := reflect.TypeOf(i)
	if t.Kind() != reflect.Slice && t.Kind() != reflect.Array {
		return errors.Format("argument must be slice or array, got: %v", t)
	}

	si := reflects.NewSliceInfo(t)
	ptr := reflects.SlicePtr(i)
	length := reflects.SliceLen(i)
	for i := 0; i < length; i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		switch t.Elem().Kind() {
		case reflect.String:
			b.WriteString("'", si.GetString(ptr, i), "'")
		case reflect.Int:
			b.WriteString("'", strconv.Itoa(si.GetInt(ptr, i)), "'")
		case reflect.Int32:
			b.WriteString("'", strconv.Itoa(int(si.GetInt32(ptr, i))), "'")
		default:
			return errors.Format("not supported type: %v", t)
		}
	}
	return
}

func (p *Provider) buildTwoColumnCriteria(b *gsd.Builder, c *gsd.TwoColumnCriteria) (err error) {
	if c.Left.Table() != nil {
		p.Quote(b, c.Left.Table().Prefix())
		b.WriteByte(Dot)
	}
	p.Quote(b, c.Left.Name())

	switch c.Type {
	case gsd.EQ:
		b.WriteByte('=')
	case gsd.NE:
		b.Write([]byte{'<', '>'})
	case gsd.LT:
		b.WriteByte('<')
	case gsd.LTE:
		b.Write([]byte{'<', '='})
	case gsd.GT:
		b.WriteByte('>')
	case gsd.GTE:
		b.Write([]byte{'>', '='})
	default:
		return errors.Format("not supported criteria type: %v", c.Type)
	}

	if c.Right.Table() != nil {
		p.Quote(b, c.Right.Table().Prefix())
		b.WriteByte(Dot)
	}
	p.Quote(b, c.Right.Name())
	return
}
