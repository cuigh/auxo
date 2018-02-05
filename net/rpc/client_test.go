package rpc_test

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/cuigh/auxo/data/guid"
	"github.com/cuigh/auxo/net/rpc"
	_ "github.com/cuigh/auxo/net/rpc/codec/http"
	_ "github.com/cuigh/auxo/net/rpc/codec/json"
	"github.com/cuigh/auxo/net/transport"
	"github.com/cuigh/auxo/test/assert"
)

func TestClient_Call(t *testing.T) {
	c, err := rpc.Dial("json", transport.Address{URL: "127.0.0.1:9000"})
	assert.NoError(t, err)

	var s string
	err = c.Call(context.Background(), "Test", "Hello", []interface{}{"auxo"}, &s)
	assert.NoError(t, err)
	assert.Equal(t, "Hello, auxo", s)

	err = c.Call(context.Background(), "Test", "Ping", nil, &s)
	assert.NoError(t, err)
	assert.Equal(t, "pong", s)
}

func BenchmarkClient_Call(b *testing.B) {
	c, err := rpc.Dial("json", transport.Address{URL: "127.0.0.1:9000"})
	assert.Error(b, err)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var s string
		err := c.Call(context.Background(), "Test", "Hello", []interface{}{"auxo"}, &s)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestHTTP(t *testing.T) {
	for i := 0; i < 10; i++ {
		callHTTP()
	}
}

func BenchmarkHTTP(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		callHTTP()
	}
}

func callHTTP() error {
	args := []interface{}{"auxo"}
	b, err := json.Marshal(args)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:9000/Test.Hello", bytes.NewBuffer(b))
	if err != nil {
		return err
	}
	req.Header.Set("ID", hex.EncodeToString(guid.New().Slice()))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	//fmt.Println(string(b))
	return nil
}
