package gsd

type ExecuteInfo struct {
	query string
	args  []interface{}
}

type executeContext struct {
	sql  string
	args []interface{}
	db   *database
	executor
}

func (c *executeContext) Reset() {
	c.sql = ""
	c.args = nil
}

func (c *executeContext) Value() (v Value) {
	v.err = c.Scan(&v.bytes)
	return
}

func (c *executeContext) Result() (ExecuteResult, error) {
	return c.exec(c.sql, c.args...)
}

func (c *executeContext) Scan(dst ...interface{}) error {
	return c.queryRow(c.sql, c.args...).Scan(dst...)
}

func (c *executeContext) Fill(i interface{}) error {
	r, err := c.Reader()
	if err != nil {
		return err
	}
	defer r.Close()

	if r.Next() {
		return r.Fill(i)
	}
	return ErrNoRows
}

func (c *executeContext) Reader() (Reader, error) {
	rows, err := c.query(c.sql, c.args...)
	return (*reader)(rows), err
}

func (c *executeContext) For(fn func(r Reader) error) error {
	r, err := c.Reader()
	if err != nil {
		return err
	}
	defer r.Close()
	return fn(r)
}
