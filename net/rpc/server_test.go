package rpc_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/cuigh/auxo/net/rpc"
	_ "github.com/cuigh/auxo/net/rpc/codec/json"
	"github.com/cuigh/auxo/test/assert"
)

func TestServer_Match(t *testing.T) {
	s := rpc.Listen(rpc.Address{URL: ":9000"})
	s.Match(rpc.Any, "json")
	//s.Use(filter.Test())
	hs := HelloService{
		Ping: func() string {
			return "pong"
		},
	}
	s.RegisterService("Test", hs)
	s.RegisterFunc("Test", "Echo", func(ctx context.Context, s string) string { return s })
	err := s.Serve()
	assert.NoError(t, err)
}

type HelloService struct {
	Ping func() string
}

func (HelloService) Hello(ctx context.Context, name string) string {
	return "Hello, " + name
}

func TestService(t *testing.T) {
	svc := HelloService{}

	var (
		sv = reflect.ValueOf(svc)
		st = reflect.TypeOf(svc)
	)
	mv := sv.MethodByName("Hello")
	mt, _ := st.MethodByName("Hello")
	t.Log(mv.Type().NumIn(), mt.Type.NumIn())
}
