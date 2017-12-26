package trace

import (
	"context"
	"database/sql"

	"github.com/cuigh/auxo/apm/trace"
	"github.com/cuigh/auxo/db/gsd"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

const component = "gsd"

type Options struct {
	*trace.Tracer
}

func Trace(opts Options) gsd.Interceptor {
	tracer := opts.Tracer
	if tracer == nil {
		tracer = trace.GetTracer()
	}

	return func(e gsd.Executor) gsd.Executor {
		return &executor{
			Executor: e,
			Tracer:   tracer,
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

func (e *executor) startSpan(operation string, ctx context.Context, query string, args ...interface{}) opentracing.Span {
	span := e.Tracer.StartChildFromContext(ctx, operation)
	ext.Component.Set(span, component)
	ext.DBType.Set(span, "sql")
	ext.DBInstance.Set(span, e.Executor.Database())
	//ext.DBUser.Set(span, "") // todo:
	ext.DBStatement.Set(span, query) // todo: only set slow query?
	return span
}
