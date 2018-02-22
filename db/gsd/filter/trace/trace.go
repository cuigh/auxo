package trace

import (
	"context"
	"database/sql"

	"github.com/cuigh/auxo/apm/trace"
	"github.com/cuigh/auxo/db/gsd"
	"github.com/opentracing/opentracing-go/ext"
)

const component = "gsd"

func Trace() gsd.Filter {
	return func(e gsd.Executor) gsd.Executor {
		return &executor{
			Executor: e,
			Tracer:   trace.GetTracer(),
		}
	}
}

type executor struct {
	*trace.Tracer
	gsd.Executor
}

func (e *executor) Exec(ctx context.Context, query string, args ...interface{}) (r sql.Result, err error) {
	span := e.startSpan("exec", ctx, query, args...)
	defer func() {
		ext.Error.Set(span, err != nil)
		span.Finish()
	}()

	r, err = e.Executor.Exec(ctx, query, args...)
	return
}

func (e *executor) QueryRow(ctx context.Context, query string, args ...interface{}) *sql.Row {
	span := e.startSpan("query_row", ctx, query, args...)
	defer span.Finish()

	return e.Executor.QueryRow(ctx, query, args...)
}

func (e *executor) QueryRows(ctx context.Context, query string, args ...interface{}) (r *sql.Rows, err error) {
	span := e.startSpan("query_rows", ctx, query, args...)
	defer func() {
		ext.Error.Set(span, err != nil)
		span.Finish()
	}()

	r, err = e.Executor.QueryRows(ctx, query, args...)
	return
}

func (e *executor) startSpan(operation string, ctx context.Context, query string, args ...interface{}) trace.Span {
	span := e.Tracer.StartChildFromContext(ctx, operation)
	ext.Component.Set(span, component)
	ext.DBType.Set(span, "sql")
	ext.DBInstance.Set(span, e.Executor.Database())
	//ext.DBUser.Set(span, "") // todo:
	ext.DBStatement.Set(span, query) // todo: only set slow query?
	return span
}
