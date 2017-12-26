# RPC

## Features

* Simple and lightweight
* High performance
* Support port multiplexing (Multiple protocols on the same port)
* Support multiple codecs: JSON, HTTP, Protocol Buffers...
* Load balance: Random, RoundRobin...
* OpenTracing support
* Service discovery and governance with etcd, consul... (TODO)
* Support authentication and authorization (TODO)
* Circuit breaker (TODO)
* Rate limit (TODO)

## Server

```go
type TestService struct {
}

func (TestService) Hello(ctx context.Context, name string) string {
	return "Hello, " + name
}

func main() {
	s := rpc.Listen(rpc.Address{URL: ":9000"})
	s.Match(json.Matcher, "json")
	//s.Match(jsoniter.Matcher, "json")
	//s.Match(proto.Matcher, "proto")
	//s.Use(filter.Test())
	s.RegisterService("Test", TestService{})
	s.RegisterFunc("Test", "Ping", func() string {
		return "pong"
	})
	log.Fatal(s.Serve())
}
```

## Client

```go
c := rpc.Dial("json", rpc.Address{URL: "127.0.0.1:9000"})

var s string
err := c.Call(context.Background(), "Test", "Hello", []interface{}{"auxo"}, &s)
```