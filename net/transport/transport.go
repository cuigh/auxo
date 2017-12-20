package transport

import (
	"context"
	"crypto/tls"
	"net"
	"strings"

	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/util/cast"
)

var (
	dialers   = map[string]Dialer{}
	listeners = map[string]Listener{}

	//ErrUnknownSchema = errors.New("transport: unknown schema")
)

type Dialer interface {
	Dial(ctx context.Context, addr string, opts data.Map) (net.Conn, error)
}

func RegisterDialer(schema string, dialer Dialer) {
	dialers[schema] = dialer
}

func Dial(ctx context.Context, uri string, opts data.Map) (net.Conn, error) {
	schema, addr := parseAddr(uri)
	dialer := dialers[schema]
	if dialer == nil {
		dialer = &simpleDialer{network: schema}
	}
	return dialer.Dial(ctx, addr, opts)
}

type Listener interface {
	Listen(addr string, opts data.Map) (net.Listener, error)
}

func RegisterListener(schema string, listener Listener) {
	listeners[schema] = listener
}

func Listen(uri string, opts data.Map) (net.Listener, error) {
	schema, addr := parseAddr(uri)
	listener := listeners[schema]
	if listener == nil {
		listener = &simpleListener{network: schema}
	}
	return listener.Listen(addr, opts)
}

// simpleDialer dials to addr with net.Dialer. It returns a tls.Conn if TLS cert & key is configured.
type simpleDialer struct {
	// Known networks are "tcp", "tcp4" (IPv4-only), "tcp6" (IPv6-only), "udp", "udp4" (IPv4-only),
	// "udp6" (IPv6-only), "ip", "ip4" (IPv4-only), "ip6" (IPv6-only), "unix", "unixgram" and "unixpacket".
	network string
}

func (d *simpleDialer) Dial(ctx context.Context, addr string, opts data.Map) (net.Conn, error) {
	c, err := loadTLSConfig(opts)
	if err != nil {
		return nil, err
	}

	dialer := &net.Dialer{}
	if opt := opts.Get("timeout"); opt != "" {
		dialer.Timeout = cast.ToDuration(opt)
	}
	if opt := opts.Get("keep_alive"); opt != "" {
		dialer.KeepAlive = cast.ToDuration(opt)
	}
	conn, err := dialer.DialContext(ctx, d.network, addr)
	if err == nil && c != nil {
		conn = tls.Client(conn, c)
	}
	return conn, err
}

// simpleListener listens to addr with net.Listen. It returns a tls.Conn if TLS cert & key is configured.
type simpleListener struct {
	// The network must be "tcp", "tcp4", "tcp6", "unix" or "unixpacket".
	network string
}

func (l *simpleListener) Listen(addr string, opts data.Map) (net.Listener, error) {
	c, err := loadTLSConfig(opts)
	if err != nil {
		return nil, err
	}

	ln, err := net.Listen(l.network, addr)
	if err == nil && c != nil {
		ln = tls.NewListener(ln, c)
	}
	return ln, err
}

func parseAddr(uri string) (schema, addr string) {
	parts := strings.SplitN(uri, "://", 2)
	if len(parts) == 1 {
		schema, addr = "tcp", uri
	} else {
		schema, addr = parts[0], parts[1]
	}
	return
}

func loadTLSConfig(opts data.Map) (*tls.Config, error) {
	var c *tls.Config
	certFile := cast.ToString(opts.Get("tls_cert"))
	keyFile := cast.ToString(opts.Get("tls_key"))
	verify := cast.ToBool(opts.Get("tls_verify"), false)
	if certFile != "" && keyFile != "" {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, err
		}
		c = &tls.Config{
			Certificates:             []tls.Certificate{cert},
			PreferServerCipherSuites: true,
			InsecureSkipVerify:       !verify,
		}
	}
	return c, nil
}
