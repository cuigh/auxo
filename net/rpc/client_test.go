package rpc_test

import (
	"context"
	"testing"

	"github.com/cuigh/auxo/net/rpc"
	_ "github.com/cuigh/auxo/net/rpc/codec/json"
	"github.com/cuigh/auxo/test/assert"
)

func TestClient_Call(t *testing.T) {
	c := rpc.Dial("json", rpc.Address{URL: "127.0.0.1:9000"})

	var s string
	err := c.Call(context.Background(), "Test", "Hello", []interface{}{"auxo"}, &s)
	assert.NoError(t, err)
	assert.Equal(t, "Hello, auxo", s)

	err = c.Call(context.Background(), "Test", "Ping", nil, &s)
	assert.NoError(t, err)
	assert.Equal(t, "pong", s)
}

func BenchmarkClient_Call(b *testing.B) {
	c := rpc.Dial("json", rpc.Address{URL: "127.0.0.1:9000"})

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var s string
		err := c.Call(context.Background(), "Test", "Hello", []interface{}{"auxo"}, &s)
		if err != nil {
			b.Fatal(err)
		}
	}
}
