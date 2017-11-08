package web

import (
	"bufio"
	"net"
	"net/http"

	"github.com/cuigh/auxo/util/cast"
)

// ResponseWriter wraps an http.ResponseWriter and implements its interface to be used
// by an HTTP handler to construct an HTTP response.
type ResponseWriter interface {
	http.ResponseWriter
	http.Flusher
	http.Hijacker
	http.CloseNotifier
	http.Pusher
	WriteString(s string) (n int, err error)
	Committed() bool
	Size() int
	Status() int
}

type responseWriter struct {
	http.ResponseWriter
	status    int
	size      int
	committed bool
	server    *Server
}

func newResponse(w http.ResponseWriter, s *Server) (r *responseWriter) {
	return &responseWriter{
		ResponseWriter: w,
		server:         s,
		status:         http.StatusOK,
	}
}

// WriteHeader sends an HTTP response header with status code. If WriteHeader is
// not called explicitly, the first call to Write will trigger an implicit
// WriteHeader(http.StatusOK). Thus explicit calls to WriteHeader are mainly
// used to send error codes.
func (r *responseWriter) WriteHeader(code int) {
	if r.committed {
		r.server.Logger.Warn("header already committed")
		return
	}
	r.status = code
	r.Header().Set(HeaderServer, r.server.cfg.Name)
	r.ResponseWriter.WriteHeader(code)
	r.committed = true
}

func (r *responseWriter) CommitHeader() {
	if !r.committed {
		r.Header().Set(HeaderServer, r.server.cfg.Name)
		r.ResponseWriter.WriteHeader(r.status)
		r.committed = true
	}
}

// Write writes the data to the connection as part of an HTTP reply.
func (r *responseWriter) Write(b []byte) (n int, err error) {
	r.CommitHeader()
	n, err = r.ResponseWriter.Write(b)
	r.size += n
	return
}

func (r *responseWriter) WriteString(s string) (n int, err error) {
	return r.Write(cast.StringToBytes(s))
}

// Size returns the data size already send to response.
func (r *responseWriter) Size() int {
	return r.size
}

// Status returns the status code send to response.
func (r *responseWriter) Status() int {
	return r.status
}

// Committed returns if the header was send to response.
func (r *responseWriter) Committed() bool {
	return r.committed
}

// Flush implements the http.Flusher interface to allow an HTTP handler to flush
// buffered data to the client.
func (r *responseWriter) Flush() {
	if f, ok := r.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Hijack implements the http.Hijacker interface to allow an HTTP handler to
// take over the connection.
func (r *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := r.ResponseWriter.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

// CloseNotify implements the http.CloseNotifier interface to allow detecting
// when the underlying connection has gone away.
// This mechanism can be used to cancel long operations on the server if the
// client has disconnected before the response is ready.
func (r *responseWriter) CloseNotify() <-chan bool {
	return r.ResponseWriter.(http.CloseNotifier).CloseNotify()
}

func (r *responseWriter) Push(target string, opts *http.PushOptions) error {
	if pusher, ok := r.ResponseWriter.(http.Pusher); ok {
		return pusher.Push(target, opts)
	}
	return http.ErrNotSupported
}

func (r *responseWriter) reset(w http.ResponseWriter) {
	r.ResponseWriter = w
	r.size = 0
	r.status = http.StatusOK
	r.committed = false
}
