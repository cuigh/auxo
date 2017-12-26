package router

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/cuigh/auxo/errors"
)

type nodeKind uint8

const (
	kindStatic nodeKind = iota
	kindParam
	kindAny
)

type route struct {
	m       *routeMap
	handler interface{}
}

func newRoute(m *routeMap, handler interface{}) *route {
	return &route{
		m:       m,
		handler: handler,
	}
}

func (r *route) Handler() interface{} {
	return r.handler
}

func (r *route) Params() []string {
	return r.m.parent.params
}

func (r *route) Path() string {
	return r.m.parent.path
}

func (r *route) URL(params ...interface{}) string {
	l := len(params)
	if l != len(r.Params()) {
		panic("Wrong parameters for route: " + r.Path())
	}

	var u string
	i := l - 1
	for n := r.m.parent; n != nil; n = n.parent {
		if n.kind == kindStatic {
			u = n.text + u
		} else {
			u = fmt.Sprint(params[i]) + u
			i--
		}
	}
	return u
}

type routeMap struct {
	parent *node

	get     *route
	post    *route
	put     *route
	delete  *route
	head    *route
	options *route
	connect *route
	patch   *route
	trace   *route
}

func newRouteMap(parent *node) *routeMap {
	m := &routeMap{
		parent: parent,
	}
	m.get = newRoute(m, nil)
	m.post = newRoute(m, nil)
	m.put = newRoute(m, nil)
	m.delete = newRoute(m, nil)
	m.head = newRoute(m, nil)
	m.options = newRoute(m, nil)
	m.connect = newRoute(m, nil)
	m.patch = newRoute(m, nil)
	m.trace = newRoute(m, nil)
	return m
}

func (m *routeMap) set(method string, handler interface{}) {
	switch method {
	case http.MethodGet:
		m.get.handler = handler
	case http.MethodPost:
		m.post.handler = handler
	case http.MethodPut:
		m.put.handler = handler
	case http.MethodDelete:
		m.delete.handler = handler
	case http.MethodPatch:
		m.patch.handler = handler
	case http.MethodOptions:
		m.options.handler = handler
	case http.MethodHead:
		m.head.handler = handler
	case http.MethodConnect:
		m.connect.handler = handler
	case http.MethodTrace:
		m.trace.handler = handler
	}
}

func (m *routeMap) find(method string) *route {
	switch method[0] {
	case 'G':
		return m.get
	case 'P':
		switch method[1] {
		case 'O':
			return m.post
		case 'U':
			return m.put
		case 'A':
			return m.patch
		}
	case 'D':
		return m.delete
	case 'O':
		return m.options
	case 'H':
		return m.head
	case 'C':
		return m.connect
	case 'T':
		return m.trace
	}
	return nil
}

type node struct {
	kind   nodeKind
	text   string
	path   string
	params []string
	routes *routeMap
	parent *node
	children
}

func newNode(kind nodeKind, text string, parent *node) *node {
	n := &node{
		kind:   kind,
		text:   text,
		parent: parent,
	}
	if parent == nil {
		n.path = text
	} else {
		n.path = parent.path + text
		n.params = parent.params
	}
	if kind != kindStatic {
		n.params = append(n.params, text[1:])
	}
	return n
}

func (n *node) find(method, path string, paramValues []string, paramIndex int) (r *route, tsr bool) {
	var tsr1, tsr2 bool

	// 1. search static nodes
	for _, c := range n.static {
		if c.text[0] != path[0] {
			continue
		}

		i, ln, lp := 0, len(c.text), len(path)
		for ; i < ln && i < lp; i++ {
			if c.text[i] != path[i] {
				goto WILD
			}
		}

		if lp > ln {
			r, tsr1 = c.find(method, path[len(c.text):], paramValues, paramIndex)
		} else if lp == ln {
			r = c.getRoute(method)
		} else {
			tsr1 = ln == lp+1 && c.text[i] == '/' && c.routes != nil
		}
		break
	}
	if r != nil && r.handler != nil {
		return r, false
	}

WILD:
	// 2. check param node
	if n.param != nil {
		i := strings.IndexByte(path, '/')
		if i > 0 {
			paramValues[paramIndex] = path[:i]
			r, tsr2 = n.param.find(method, path[i:], paramValues, paramIndex+1)
		} else {
			paramValues[paramIndex] = path
			r = n.param.getRoute(method)
			if r == nil || r.handler == nil {
				tsr2 = n.param.getStatic("/") != nil
			}
		}

		if r != nil && r.handler != nil {
			return r, false
		}
	}

	// 3. check any node
	if n.any != nil {
		paramValues[paramIndex] = path
		return n.any.getRoute(method), false
	}

	return r, tsr1 || tsr2 || (path == "/" && n.routes != nil)
}

func (n *node) getRoute(method string) *route {
	if n.routes == nil {
		return nil
	}
	return n.routes.find(method)
}

func (n *node) setHandler(method string, handler interface{}) (*route, error) {
	if n.routes == nil {
		n.routes = newRouteMap(n)
	}

	r := n.getRoute(method)
	if r.handler != nil {
		return nil, errors.Format("route conflict: %s > %s", method, n.path)
	}

	r.handler = handler
	return r, nil
}

func (n *node) correctPath() {
	n.path = n.text
	for p := n.parent; p != nil; p = p.parent {
		n.path = p.text + n.path
	}
}

func (n *node) add(path string) (*node, error) {
	var (
		err         error
		c           = n
		start, i, l = 0, 0, len(path)
	)
	for ; i < l; i++ {
		if path[i] == ':' {
			c = c.addSegment(path[start:i])
			start = i

			for ; i < l && path[i] != '/'; i++ {
			}
			if i == l {
				return c.children.setParam(c, path[start:i])
			}
			c, err = c.children.setParam(c, path[start:i])
			if err != nil {
				return nil, err
			}
			start = i
		} else if path[i] == '*' {
			c = c.addSegment(path[start:i])
			return c.setAny(c, path[i:])
		}
	}

	return c.addSegment(path[start:]), nil
}

func (n *node) addSegment(path string) *node {
	if path == "" {
		return n
	}

	for _, c := range n.children.static {
		i, ln, lp := 0, len(c.text), len(path)
		for ; i < ln && i < lp && c.text[i] == path[i]; i++ {
		}
		if i == 0 {
			continue
		}

		if i == ln {
			if ln == lp {
				return c
			}
			return c.addSegment(path[i:])
		} else if i == lp {
			c.split(i)
			return c
		} else {
			c.split(i)
			return c.children.addStatic(c, path[i:])
		}
	}

	return n.children.addStatic(n, path)
}

func (n *node) split(i int) *node {
	c := *n
	c.parent = n
	c.text = n.text[i:]
	c.children.setParent(&c)
	if c.routes != nil {
		c.routes.parent = &c
	}

	n.text = n.text[:i]
	n.children = children{static: []*node{&c}}
	n.routes = nil

	n.correctPath()
	c.correctPath()
	return &c
}

func (n *node) walk(all bool, fn func(n *node)) {
	if all || n.routes != nil {
		fn(n)
	}

	for _, t := range n.static {
		t.walk(all, fn)
	}
	if n.param != nil {
		n.param.walk(all, fn)
	}
	if n.any != nil {
		n.any.walk(all, fn)
	}
}

type children struct {
	static []*node
	param  *node
	any    *node
}

func (c *children) setParent(p *node) {
	for _, n := range c.static {
		n.parent = p
	}
	if c.param != nil {
		c.param.parent = p
	}
	if c.any != nil {
		c.any.parent = p
	}
}

func (c *children) getStatic(text string) *node {
	for _, n := range c.static {
		if n.text == text {
			return n
		}
	}
	return nil
}

func (c *children) setParam(parent *node, text string) (*node, error) {
	if c.param != nil {
		if c.param.text != text {
			return nil, errors.Format("route conflict: %s <=> %s", c.param.path, c.param.parent.path+text)
		}
		return c.param, nil
	}

	c.param = newNode(kindParam, text, parent)
	return c.param, nil
}

func (c *children) setAny(parent *node, text string) (*node, error) {
	if c.any != nil {
		if c.any.text != text {
			return nil, errors.Format("route conflict: %s <=> %s", c.any.path, c.any.parent.path+text)
		}
		return c.any, nil
	}

	c.any = newNode(kindAny, text, parent)
	return c.any, nil
}

func (c *children) addStatic(parent *node, text string) (n *node) {
	n = newNode(kindStatic, text, parent)
	c.static = append(c.static, n)
	return
}
