package binder_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/cuigh/auxo/net/web/binder"
	"github.com/cuigh/auxo/test/assert"
)

const (
	HeaderContentType   = "Content-Type"
	MIMEApplicationJSON = "application/json"
	MIMEApplicationXML  = "application/xml"
	MIMEApplicationForm = "application/x-www-form-urlencoded"
	//MIMEMultipartForm   = "multipart/form-data"
)

type User struct {
	ID   int32  `json:"id" xml:"id"`
	Name string `json:"name" xml:"name" query:"name" form:"name" bind:"name,path=name,query=name,form=name,cookie=name,header=name"`
}

//type Context struct {
//	req *http.Request
//}
//
//func (c *Context) Request() *http.Request {
//	return c.req
//}
//
//func (c *Context) Param(name string) string {
//	return ""
//}

func TestBindQuery(t *testing.T) {
	const content = `id=1&name=test`

	r := httptest.NewRequest(http.MethodGet, "/?"+content, nil)
	bindUser(t, r)
}

func TestBindForm(t *testing.T) {
	const content = `id=1&name=test`

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(content))
	r.Header.Set(HeaderContentType, MIMEApplicationForm)
	bindUser(t, r)
}

func TestBindJSON(t *testing.T) {
	const content = `{"id":1,"name":"test"}`

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(content))
	r.Header.Set(HeaderContentType, MIMEApplicationJSON)
	bindUser(t, r)
}

func TestBindXML(t *testing.T) {
	const content = `<user><id>1</id><name>test</name></user>`

	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(content))
	r.Header.Set(HeaderContentType, MIMEApplicationXML)
	bindUser(t, r)
}

func TestBindPtrValue(t *testing.T) {
	v := &struct {
		ID *int
	}{}
	r := httptest.NewRequest(http.MethodGet, "/?id=1", nil)
	bind(t, r, v)
	assert.NotNil(t, v.ID)
	assert.Equal(t, *v.ID, 1)
}

func TestBindTimeValue(t *testing.T) {
	const s = "2011-11-11T11:11:11Z"
	v := &struct {
		Time time.Time
	}{}
	r := httptest.NewRequest(http.MethodGet, "/?time="+s, nil)
	bind(t, r, v)
	assert.NotEmpty(t, v.Time)
}

func TestBindDurationValue(t *testing.T) {
	const d = 15 * time.Minute
	v := &struct {
		Time time.Duration
	}{}
	r := httptest.NewRequest(http.MethodGet, "/?time="+d.String(), nil)
	bind(t, r, v)
	assert.Equal(t, d, v.Time)
}

func TestBindSliceValue(t *testing.T) {
	v := &struct {
		Names []string
	}{}
	r := httptest.NewRequest(http.MethodGet, "/?names=x&names=y", nil)
	bind(t, r, v)
	assert.Equal(t, 2, len(v.Names))
	assert.Equal(t, "x", v.Names[0])
	assert.Equal(t, "y", v.Names[1])
}

// TODO:
//func TestBindStructValue(t *testing.T) {
//	const name = "foobar"
//	v := &struct {
//		Book struct {
//			Name string
//		}
//	}{}
//	r := httptest.NewRequest(http.MethodGet, "/?book.name="+name, nil)
//	bind(t, r, v)
//	assert.Equal(t, name, v.Book.Name)
//}

func bindUser(t *testing.T, r *http.Request) {
	t.Helper()

	b := binder.New(binder.Options{})
	user := &User{}
	err := b.Bind(r, user)
	assert.NoError(t, err)
	assert.Equal(t, int32(1), user.ID)
	assert.Equal(t, "test", user.Name)
}

func bind(t *testing.T, r *http.Request, v interface{}) {
	t.Helper()

	b := binder.New(binder.Options{})
	err := b.Bind(r, v)
	assert.NoError(t, err)
}
