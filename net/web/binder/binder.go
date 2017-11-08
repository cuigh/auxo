package binder

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"reflect"
	"strings"

	"github.com/cuigh/auxo/byte/size"
	"github.com/cuigh/auxo/errors"
	"github.com/cuigh/auxo/ext/texts"
	"github.com/cuigh/auxo/util/cast"
)

var unmarshalerType = reflect.TypeOf((*Unmarshaler)(nil)).Elem()

// Unmarshaler is the interface implemented by types that can unmarshal a param to themselves.
type Unmarshaler interface {
	// UnmarshalParam converts a string value to type.
	UnmarshalParam(param string) error
}

// Options represents options of Binder.
type Options struct {
	MaxMemory size.Size
}

// Binder is an implementation of the `web#Binder` interface.
type Binder struct {
	opts Options
}

// Context is the interface implemented by web framework which hold the request information.
//type Context interface {
//	Request() *http.Request
//	Param(name string) string
//}

// New creates a Binder instance.
func New(opts Options) *Binder {
	if opts.MaxMemory <= 0 {
		opts.MaxMemory = 10 * size.MB
	}
	return &Binder{opts: opts}
}

// Bind takes data out of the request and deserializes into a struct according
// to the Content-Type of the request.
func (b *Binder) Bind(r *http.Request, i interface{}) (err error) {
	if r.Method == http.MethodGet {
		err = b.unmarshal(i, r.URL.Query())
		return
	}

	if r.ContentLength == 0 {
		return errors.New("request body is empty")
	}

	ct := r.Header.Get("Content-Type")
	switch {
	case strings.HasPrefix(ct, "application/x-www-form-urlencoded"):
		err = r.ParseForm()
		if err != nil {
			return
		}
		return b.unmarshal(i, r.Form)
	case strings.HasPrefix(ct, "multipart/form-data"):
		err = r.ParseMultipartForm(int64(b.opts.MaxMemory))
		if err != nil {
			return
		}
		// todo: add support to files
		return b.unmarshal(i, r.Form)
	case strings.HasPrefix(ct, "application/json"):
		err = json.NewDecoder(r.Body).Decode(i)
	case strings.HasPrefix(ct, "application/xml"):
		err = xml.NewDecoder(r.Body).Decode(i)
	default:
		return errors.New("unsupported content type")
	}
	return
}

func (b *Binder) unmarshal(v interface{}, data map[string][]string) error {
	vt := reflect.TypeOf(v)
	if vt.Kind() != reflect.Ptr {
		panic("v must be a pointer of struct")
	}

	vv := reflect.ValueOf(v).Elem()
	if vv.Kind() != reflect.Struct {
		panic("v must be a pointer of struct")
	}

	return b.unmarshalValue(vv, data)
}

func (b *Binder) unmarshalValue(v reflect.Value, data map[string][]string) error {
	var (
		x   reflect.Value
		err error
	)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	t := v.Type()
	count := t.NumField()

	for i := 0; i < count; i++ {
		f := v.Field(i)
		sf := t.Field(i)
		name := b.getName(&sf)
		values := data[name]
		if values == nil {
			// TODO: handle struct field
			continue
		}
		value := values[0]

		if f.Kind() == reflect.Ptr {
			if f.IsNil() {
				f.Set(reflect.New(f.Type().Elem()))
			}
			f = f.Elem()
		}

		// check custom unmarshal interface
		if f.CanAddr() && f.Addr().Type().Implements(unmarshalerType) {
			err = f.Addr().Interface().(Unmarshaler).UnmarshalParam(value)
			if err != nil {
				return err
			}
			continue
		}

		if f.Kind() == reflect.Slice {
			x, err = cast.TryToSliceValue(values, f.Type().Elem())
		} else {
			x, err = cast.TryToValue(value, f.Type())
		}
		if err != nil {
			if f.Kind() == reflect.Struct {
				return b.unmarshalValue(f, data)
			}
			return errors.Format("can't decode param '%s' to type %v: %v", name, sf.Type, err)
		}
		f.Set(x)
	}
	return nil
}

func (b *Binder) getName(sf *reflect.StructField) (name string) {
	tag := sf.Tag.Get("bind")
	if tag != "" {
		// TODO: support complex tags > bind:"id,path=id,cookie=cid,header=hid"
		items := strings.Split(tag, ",")
		name = items[0]
	}
	if name == "" {
		name = texts.Rename(sf.Name, texts.Lower)
	}
	return
}
