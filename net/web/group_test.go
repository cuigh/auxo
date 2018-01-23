package web

import (
	"net/http"
	"testing"

	"github.com/cuigh/auxo/test/assert"
)

func TestGroup(t *testing.T) {
	cases := []struct {
		Route string
	}{
		{"/"},
	}

	s := Default()
	g := s.Group("/group")
	h := func(Context) error { return nil }

	for _, c := range cases {
		g.Connect(c.Route, h)
		g.Delete("/", h)
		g.Get("/", h)
		g.Head("/", h)
		g.Options("/", h)
		g.Patch("/", h)
		g.Post("/", h)
		g.Put("/", h)
		g.Trace("/", h)
		//g.Any("/", h)
		//g.Match([]string{http.MethodGet, http.MethodPost}, "/", h)
		g.Static("/static", "static")
		g.File("/favicon.ico", "favicon.ico")
	}

	ctx := s.AcquireContext(nil, nil)
	for _, c := range cases {
		r, tsr := s.router.Find(http.MethodGet, "/group"+c.Route, ctx.PathValues())
		assert.NotNil(t, r)
		assert.False(t, tsr)
	}
}
