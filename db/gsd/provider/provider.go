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
	b.Append("INSERT INTO ")
	p.Quote(b, info.Table)
	b.AppendByte('(')
	p.Quote(b, info.Columns[0])
	for i := 1; i < cols; i++ {
		b.AppendByte(Comma)
		p.Quote(b, info.Columns[i])
	}
	b.Append(") VALUES ")

	rows := len(info.Values) / cols
	for i := 0; i < rows; i++ {
		if i > 0 {
			b.AppendByte(Comma)
		}
		if cols > 20 {
			// todo: optimize
			b.AppendBytes('(', '?')
			for j := 1; j < cols; j++ {
				b.AppendBytes(',', '?')
			}
			b.AppendByte(')')
		} else {
			b.Append(insertValueClauses[cols-1])
		}
	}
	b.Args = info.Values
	return
}

func (p *Provider) BuildDelete(b *gsd.Builder, info *gsd.DeleteInfo) (err error) {
	b.Append("DELETE FROM ")
	p.Quote(b, info.Table)
	if info.Where == nil || info.Where.Empty() {
		return errors.New("delete action must have filters")
	}
	b.Append(" WHERE ")
	p.buildFilters(b, info.Where)
	return
}

func (p *Provider) BuildUpdate(b *gsd.Builder, info *gsd.UpdateInfo) (err error) {
	b.Append("UPDATE ")
	p.Quote(b, info.Table)
	b.Append(" SET ")
	for i, col := range info.Columns {
		if i > 0 {
			b.AppendByte(Comma)
		}
		p.Quote(b, col)
		b.AppendByte('=')
		switch v := info.Values[i].(type) {
		case gsd.IncValue:
			p.Quote(b, col)
			b.AppendBytes('+', '?')
			b.Args = append(b.Args, v.Value)
		case gsd.DecValue:
			p.Quote(b, col)
			b.AppendBytes('-', '?')
			b.Args = append(b.Args, v.Value)
		case gsd.ExprValue:
			b.Append(string(v))
		default:
			b.AppendByte('?')
			b.Args = append(b.Args, v)
		}
	}
	if info.Where != nil && !info.Where.Empty() {
		b.Append(" WHERE ")
		p.buildFilters(b, info.Where)
	}
	return
}

func (p *Provider) BuildSelect(b *gsd.Builder, info *gsd.SelectInfo) (err error) {
	// SELECT
	b.Append("SELECT ")
	if info.Distinct {
		b.Append("DISTINCT ")
	}
	for i, col := range info.Columns {
		if i > 0 {
			b.AppendByte(',')
		}
		if col.Table() == nil {
			if _, ok := col.(*gsd.ExprColumn); ok {
				b.Append(col.Name())
			} else {
				p.Quote(b, col.Name())
			}
		} else {
			p.Quote(b, col.Table().Prefix())
			b.AppendByte(Dot)
			p.Quote(b, col.Name())
		}
		if col.Alias() != "" {
			b.Append(" AS ", col.Alias())
		}
	}

	// FROM
	b.Append(" FROM ")
	p.Quote(b, info.Table.Name())
	if info.Table.Alias() != "" {
		b.Append(" AS ", info.Table.Alias())
	}

	// JOIN
	p.buildJoin(b, info)

	// WHERE
	if info.Where != nil && !info.Where.Empty() {
		b.Append(" WHERE ")
		p.buildFilters(b, info.Where)
	}

	// GROUP BY
	p.buildGroupBy(b, info)

	// ORDER BY
	if len(info.Orders) > 0 {
		b.Append(" ORDER BY ")
		for i, order := range info.Orders {
			if i > 0 {
				b.AppendByte(Comma)
			}
			for j, col := range order.Columns {
				if j > 0 {
					b.AppendByte(Comma)
				}
				p.buildColumn(b, col)
			}
			if order.Type == gsd.DESC {
				b.Append(" DESC")
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
		p.buildCount(b, info)
	}
	return
}

func (p *Provider) buildCount(b *gsd.Builder, info *gsd.SelectInfo) {
	// SELECT FROM
	b.Append(";SELECT COUNT(0) FROM ")
	p.Quote(b, info.Table.Name())
	if info.Table.Alias() != "" {
		b.Append(" AS ", info.Table.Alias())
	}

	// JOIN
	p.buildJoin(b, info)

	// WHERE
	if info.Where != nil && !info.Where.Empty() {
		b.Append(" WHERE ")
		p.buildFilters(b, info.Where)
	}

	// GROUP BY
	p.buildGroupBy(b, info)
}

func (p *Provider) buildJoin(b *gsd.Builder, info *gsd.SelectInfo) {
	for _, join := range info.Joins {
		b.Append(" ", join.Type, " ")
		p.Quote(b, join.Table.Name())
		if join.Table.Alias() != "" {
			b.Append(" AS ", join.Table.Alias())
		}
		b.Append(" ON ")
		p.buildFilters(b, join.On)
	}
}

func (p *Provider) buildGroupBy(b *gsd.Builder, info *gsd.SelectInfo) {
	if len(info.Groups) > 0 {
		b.Append(" GROUP BY ")
		for i, col := range info.Groups {
			if i > 0 {
				b.AppendByte(',')
			}
			p.buildColumn(b, col)
		}
		if info.Having != nil {
			b.Append(" HAVING ")
			p.buildFilters(b, info.Having)
		}
	}
}

func (p *Provider) buildColumn(b *gsd.Builder, col gsd.Column) {
	if col.Table() == nil {
		p.Quote(b, col.Name())
	} else {
		p.Quote(b, col.Table().Prefix())
		b.AppendByte(Dot)
		p.Quote(b, col.Name())
	}
}

func (p *Provider) buildFilters(b *gsd.Builder, filters gsd.Filters) {
	switch fs := filters.(type) {
	case *gsd.SimpleFilters:
		for i, filter := range *fs {
			if i > 0 {
				b.Append(" AND ")
			}
			switch f := filter.(type) {
			case *gsd.OneColumnFilter:
				p.buildOneColumnFilter(b, f)
			case *gsd.TwoColumnFilter:
				p.buildTwoColumnFilter(b, f)
			case gsd.ExprFilter:
				b.Append(string(f))
			}
		}
	case *gsd.NotFilters:
		b.Append("NOT(")
		p.buildFilters(b, fs.Inner)
		b.AppendByte(')')
	case gsd.JoinFilters:
		if fs.Left().Empty() {
			p.buildFilters(b, fs.Right())
		} else if fs.Right().Empty() {
			p.buildFilters(b, fs.Left())
		} else {
			b.AppendByte('(')
			p.buildFilters(b, fs.Left())
			b.Append(") ", fs.Joiner(), " (")
			p.buildFilters(b, fs.Right())
			b.AppendByte(')')
		}
	}
}

func (p *Provider) buildOneColumnFilter(b *gsd.Builder, f *gsd.OneColumnFilter) (err error) {
	if f.Table() != nil {
		p.Quote(b, f.Table().Prefix())
		b.AppendByte(Dot)
	}

	switch f.Type {
	case gsd.NE:
		if f.Value == nil {
			p.Quote(b, f.Name())
			b.Append(" IS NOT NULL")
			return
		} else {
			p.Quote(b, f.Name())
			b.Append("<>?")
		}
	case gsd.LT:
		p.Quote(b, f.Name())
		b.Append("<?")
	case gsd.LTE:
		p.Quote(b, f.Name())
		b.Append("<=?")
	case gsd.GT:
		p.Quote(b, f.Name())
		b.Append(">?")
	case gsd.GTE:
		p.Quote(b, f.Name())
		b.Append(">=?")
	case gsd.IN:
		p.Quote(b, f.Name())
		b.Append(" IN(")
		p.buildInValues(b, f.Value)
		b.AppendByte(')')
		return
	case gsd.NIN:
		p.Quote(b, f.Name())
		b.Append(" NOT IN(")
		p.buildInValues(b, f.Value)
		b.AppendByte(')')
		return
	case gsd.LK:
		p.Quote(b, f.Name())
		b.Append(" LIKE '", f.Value.(string), "'")
	default:
		if f.Value == nil {
			p.Quote(b, f.Name())
			b.Append(" IS NULL")
			return
		} else {
			p.Quote(b, f.Name())
			b.Append("=?")
		}
	}
	b.Args = append(b.Args, f.Value)
	return
}

func (p *Provider) buildInValues(b *gsd.Builder, i interface{}) (err error) {
	t := reflect.TypeOf(i)
	si := reflects.NewSliceInfo(t)
	ptr := reflects.SlicePtr(i)
	length := reflects.SliceLen(i)
	for i := 0; i < length; i++ {
		if i > 0 {
			b.AppendByte(',')
		}
		switch t.Kind() {
		case reflect.String:
			b.Append("'", si.GetString(ptr, i), "'")
		case reflect.Int:
			b.Append("'", strconv.Itoa(si.GetInt(ptr, i)), "'")
		case reflect.Int32:
			b.Append("'", strconv.Itoa(int(si.GetInt32(ptr, i))), "'")
		default:
			return errors.Format("not supported type: %v", t)
		}
	}
	return
}

func (p *Provider) buildTwoColumnFilter(b *gsd.Builder, f *gsd.TwoColumnFilter) (err error) {
	if f.Left.Table() != nil {
		p.Quote(b, f.Left.Table().Prefix())
		b.AppendByte(Dot)
	}
	p.Quote(b, f.Left.Name())

	switch f.Type {
	case gsd.EQ:
		b.AppendByte('=')
	case gsd.NE:
		b.AppendBytes('<', '>')
	case gsd.LT:
		b.AppendByte('<')
	case gsd.LTE:
		b.AppendBytes('<', '=')
	case gsd.GT:
		b.AppendByte('>')
	case gsd.GTE:
		b.AppendBytes('>', '=')
	default:
		return errors.Format("not supported filter type: %v", f.Type)
	}

	if f.Right.Table() != nil {
		p.Quote(b, f.Right.Table().Prefix())
		b.AppendByte(Dot)
	}
	p.Quote(b, f.Right.Name())
	return
}
