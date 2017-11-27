package gsd

type Table interface {
	Name() string
	Alias() string
	Prefix() string
}

func NewTable(name string, alias ...string) Table {
	if len(alias) == 0 {
		return simpleTable(name)
	}
	return newAliasTable(name, alias[0])
}

func toTable(table interface{}) Table {
	switch t := table.(type) {
	case Table:
		return t
	case string:
		return simpleTable(t)
	default:
		panic("gsd: table must be Table or string")
	}
}

type simpleTable string

func (t simpleTable) Name() string {
	return string(t)
}

func (t simpleTable) Alias() string {
	return ""
}

func (t simpleTable) Prefix() string {
	return string(t)
}

type aliasTable struct {
	name   string
	alias  string
	prefix string
}

func (t *aliasTable) Name() string {
	return t.name
}

func (t *aliasTable) Alias() string {
	return t.alias
}

func (t *aliasTable) Prefix() string {
	return t.prefix
}

func newAliasTable(name, alias string) Table {
	t := &aliasTable{
		name:  name,
		alias: alias,
	}
	if alias == "" {
		t.prefix = name
	} else {
		t.prefix = alias
	}
	return t
}
