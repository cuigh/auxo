package router

import (
	"net/http"
	"unsafe"

	"io"

	"fmt"

	"github.com/cuigh/auxo/errors"
)

// Tree represents a route tree based on Trie data structure.
type Tree struct {
	opts     Options
	root     *node
	maxParam int
}

// Options represents route tree options.
type Options struct {
	// todo:
	//IgnoreCase bool
	//DecodeParam bool
}

type Route interface {
	//Method() string
	Path() string
	Handler() unsafe.Pointer
	Params() []string
}

// New creates a route tree.
func New(opts Options) *Tree {
	return &Tree{
		root: newNode(kindStatic, "/", nil, nil, nil),
		opts: opts,
	}
}

// MaxParam returns max parameters of all routes.
func (t *Tree) MaxParam() int {
	return t.maxParam
}

// Add register a route with specific methods to the tree.
func (t *Tree) Add(path string, handler unsafe.Pointer, methods ...string) error {
	if path[0] != '/' {
		return errors.New("path must start with '/'")
	}

	if path == "/" {
		return t.root.setHandler(methods, handler)
	}

	n, err := t.root.add(methods, path[1:], handler)
	if err != nil {
		return err
	}

	if l := len(n.params); l > t.maxParam {
		t.maxParam = l
	}
	return nil
}

// Find tries to find a matched route in the tree.
func (t *Tree) Find(method, path string, paramValues []string) (r Route, tsr bool) {
	path = path[1:]
	var route *route
	if path == "" {
		route = t.root.getRoute(method)
		tsr = route == nil && t.root.routes != nil
	} else {
		route, tsr = t.root.find(method, path, paramValues, 0)
	}
	if route != nil {
		r = route
	}
	return
}

// Walk traverses all route nodes of tree.
func (t *Tree) Walk(fn func(r Route, method string)) {
	methods := [...]string{
		http.MethodConnect,
		http.MethodDelete,
		http.MethodGet,
		http.MethodHead,
		http.MethodOptions,
		http.MethodPatch,
		http.MethodPost,
		http.MethodPut,
		http.MethodTrace,
	}
	t.root.walk(false, func(n *node) {
		for _, m := range methods {
			if r := n.getRoute(m); r != nil && r.handler != nil {
				fn(r, m)
			}
		}
	})
}

// Print prints all routes.
func (t *Tree) Print(w io.Writer) {
	allMethods := [...]string{
		http.MethodConnect,
		http.MethodDelete,
		http.MethodGet,
		http.MethodHead,
		http.MethodOptions,
		http.MethodPatch,
		http.MethodPost,
		http.MethodPut,
		http.MethodTrace,
	}
	t.root.walk(false, func(n *node) {
		var methods []string
		for _, m := range allMethods {
			if r := n.getRoute(m); r != nil && r.handler != nil {
				methods = append(methods, m)
			}
		}

		if len(methods) > 0 {
			fmt.Fprintln(w, n.path, methods)
		} else {
			fmt.Fprintln(w, n.path)
		}
	})
}
