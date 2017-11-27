package gsd

/********** Common **********/

type ResultClause interface {
	Submit() error
	Result() (Result, error)
}

/********** Select Clauses **********/

type SelectClause interface {
	From(table interface{}) FromClause
}

type FromClause interface {
	LimitClause
	SelectResultClause
	Join(table interface{}, on Filters) JoinClause
	LeftJoin(t interface{}, on Filters) JoinClause
	RightJoin(t interface{}, on Filters) JoinClause
	FullJoin(t interface{}, on Filters) JoinClause
	Where(f Filters) WhereClause
}

type JoinClause interface {
	FromClause
}

type WhereClause interface {
	LimitClause
	//One(i interface{}) error
	SelectResultClause
	GroupBy(cols *Columns) GroupByClause
	OrderBy(orders ...*Order) OrderByClause
}

type GroupByClause interface {
	LimitClause
	OrderBy(orders ...*Order) OrderByClause
	Having(f Filters) HavingClause
	SelectResultClause
}

type HavingClause interface {
	LimitClause
	SelectResultClause
	OrderBy(orders ...*Order) OrderByClause
}

type OrderByClause interface {
	LimitClause
	SelectResultClause
}

type LimitClause interface {
	//Lock(Exclusive/Shared) SelectResultClause
	Limit(skip, take int) SelectResultClause
	Page(index, size int) SelectResultClause
}

type SelectResultClause interface {
	Value() Value
	//Int() (int, error)
	Scan(dst ...interface{}) error
	Fill(i interface{}) error
	//One(i interface{}) error
	//All(i interface{}) error
	List(i interface{}, total *int) error
	Reader() (Reader, error)
	For(fn func(r Reader) error) error
}

/********** Count Clauses **********/

type CountClause interface {
	Join(table interface{}, on Filters) CountClause
	LeftJoin(t interface{}, on Filters) CountClause
	RightJoin(t interface{}, on Filters) CountClause
	FullJoin(t interface{}, on Filters) CountClause
	Where(f Filters) CountWhereClause
	CountResultClause
}

type CountWhereClause interface {
	GroupBy(cols *Columns) CountGroupByClause
	CountResultClause
}

type CountGroupByClause interface {
	Having(f Filters) CountResultClause
	CountResultClause
}

type CountResultClause interface {
	Value() (int, error)
	Scan(dst interface{}) error
}

/********** Update Clauses **********/

type UpdateClause interface {
	Set(col string, val interface{}) SetClause
	Inc(col string, val interface{}) SetClause
	Dec(col string, val interface{}) SetClause
	Expr(col string, val string) SetClause
}

type SetClause interface {
	UpdateClause
	ResultClause
	Where(f Filters) ResultClause
}

/********** Delete Clauses **********/

type DeleteClause interface {
	Where(f Filters) ResultClause
}

/********** Insert Clauses **********/

type InsertClause interface {
	Columns(cols ...string) InsertColumnsClause
}

type InsertColumnsClause interface {
	Values(values ...interface{}) InsertValuesClause
}

type InsertValuesClause interface {
	Values(values ...interface{}) InsertValuesClause
	InsertResultClause
}

//type CreateClause interface {
//	Result() CreateResult
//}

type InsertResultClause interface {
	Submit() error
	Result() (InsertResult, error)
}

/********** Execute Clauses **********/

type ExecuteClause interface {
	Result() (ExecuteResult, error)
	Value() Value
	Scan(dst ...interface{}) error
	Fill(i interface{}) error
	//One(i interface{}) error
	//All(i interface{}) error
	Reader() (Reader, error)
	For(fn func(r Reader) error) error
}

/********** Call Clauses **********/

//type CallClause interface {
//	Result() (InsertResult, error)
//	Value() Value
//	Scan(dst ...interface{}) error
//	One(i interface{}) error
//	All(i interface{}) error
//	Set() *Set
//}
