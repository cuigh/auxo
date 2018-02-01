# RPC

## Features

* Simple and lightweight
* High performance
* Support port multiplexing (Multiple protocols on the same port)
* Support multiple codecs: JSON, HTTP, Protocol Buffers...
* Load balance: Random, RoundRobin...
* OpenTracing support
* Service discovery and governance with etcd, consul...
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
	s := rpc.Listen(transport.Address{URL: ":9000"})
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
c := rpc.Dial("json", transport.Address{URL: "127.0.0.1:9000"})

var s string
err := c.Call(context.Background(), "Test", "Hello", []interface{}{"auxo"}, &s)
```

## Code generating

If you use **Protocol Buffers** codec, you can generate contract codes from [Protocol Buffers](https://developers.google.com/protocol-buffers/docs/proto3) service definition files with [protoc-gen-auxo](https://github.com/cuigh/protoc-gen-auxo).

### Install tools

Use `go get` to install the code generator:

```bash
go install github.com/cuigh/protoc-gen-auxo
```

You will also need:

* [protoc](https://github.com/golang/protobuf), the protobuf compiler. You need version 3+.
* [github.com/golang/protobuf/protoc-gen-go](https://github.com/golang/protobuf/), the Go protobuf generator plugin. Get this with `go get`.

### Usage

```bash
protoc --go_out=. --auxo_out=. hello.proto
```

The name of generated code file is end with `.auxo.go`. The file include service interfaces and client proxies.