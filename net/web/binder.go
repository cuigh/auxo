package web

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"

	"github.com/cuigh/auxo/errors"
	"github.com/cuigh/auxo/ext/texts"
	"github.com/cuigh/auxo/util/cast"
)

var unmarshalerType = reflect.TypeOf((*Unmarshaler)(nil)).Elem()

// Unmarshaler is the interface implemented by types that can unmarshal a param to themselves.
type Unmarshaler interface {
	// Unmarshal converts a param value to type.
	Unmarshal(param interface{}) error
}

// binder is an implementation of the `web#Binder` interface.
type binder struct {
}

// Bind takes data out of the request and deserializes into a struct according
// to the Content-Type of the request.
func (b *binder) Bind(ctx Context, i interface{}) (err error) {
	if ctx.Request().Method == http.MethodGet {
		err = b.unmarshal(ctx, i)
		return
	}

	if ctx.Request().ContentLength == 0 {
		return errors.New("request body is empty")
	}

	ct := ctx.ContentType()
	switch {
	case strings.HasPrefix(ct, "application/x-www-form-urlencoded"):
		return b.unmarshal(ctx, i)
	case strings.HasPrefix(ct, "multipart/form-data"):
		return b.unmarshal(ctx, i)
	case strings.HasPrefix(ct, "application/json"):
		return json.NewDecoder(ctx.Request().Body).Decode(i)
	case strings.HasPrefix(ct, "application/xml"):
		return xml.NewDecoder(ctx.Request().Body).Decode(i)
	default:
		return errors.New("unsupported content type")
	}
}

func (b *binder) unmarshal(ctx Context, v interface{}) error {
	vt := reflect.TypeOf(v)
	if vt.Kind() != reflect.Ptr {
		panic("v must be a pointer of struct")
	}

	vv := reflect.ValueOf(v).Elem()
	if vv.Kind() != reflect.Struct {
		panic("v must be a pointer of struct")
	}

	return b.unmarshalValue(ctx, vv)
}

func (b *binder) unmarshalValue(ctx Context, v reflect.Value) error {
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
		value := b.getValue(ctx, &sf)
		// TODO: handle struct field
		if value == nil {
			continue
		}

		if f.Kind() == reflect.Ptr {
			if f.IsNil() {
				f.Set(reflect.New(f.Type().Elem()))
			}
			f = f.Elem()
		}

		// check custom unmarshal interface
		if f.CanAddr() && f.Addr().Type().Implements(unmarshalerType) {
			err = f.Addr().Interface().(Unmarshaler).Unmarshal(value)
			if err != nil {
				return err
			}
			continue
		}

		if f.Kind() == reflect.Slice {
			x, err = cast.TryToSliceValue(value, f.Type().Elem())
		} else {
			x, err = cast.TryToValue(value, f.Type())
		}

		if err == nil {
			f.Set(x)
		} else {
			if f.Kind() != reflect.Struct {
				return errors.Format("can't decode param '%v' to type %v: %v", value, sf.Type, err)
			}

			err = b.unmarshalValue(ctx, f)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *binder) getValue(ctx Context, sf *reflect.StructField) interface{} {
	// bind:"name,name=query,name=file.name,name=path,name=form,name=cookie,name=header"
	tag := sf.Tag.Get("bind")
	if tag == "" {
		return b.getFormValue(ctx, sf, texts.Rename(sf.Name, texts.Lower))
	}

	names := strings.Split(tag, ",")
	for _, item := range names {
		pair := strings.SplitN(item, "=", 2)
		name := pair[0]

		// bind query/form value as default
		if len(pair) == 1 {
			if value := b.getFormValue(ctx, sf, name); value != nil {
				return value
			}
			continue
		}

		// For files, value may bind with 'Filename' or 'Size' field.
		if strings.HasPrefix(pair[1], "file") {
			if value := b.getFileValue(ctx, name, pair[1]); value != nil {
				return value
			}
			continue
		}

		switch pair[1] {
		case "path":
			if value := ctx.P(name); value != "" {
				return value
			}
		case "query":
			if value := b.getMapValue(ctx, sf, ctx.QueryValues(), name); value != "" {
				return value
			}
		case "cookie":
			if cookie, err := ctx.Cookie(name); err == nil {
				return cookie.Value
			}
		case "header":
			if value := b.getMapValue(ctx, sf, ctx.Request().Header, name); value != nil {
				return value
			}
		case "form":
			fallthrough
		default:
			if value := b.getFormValue(ctx, sf, name); value != nil {
				return value
			}
		}
	}
	return nil
}

func (b *binder) getFormValue(ctx Context, sf *reflect.StructField, name string) interface{} {
	if values, err := ctx.FormValues(); err == nil {
		if value := b.getMapValue(ctx, sf, values, name); value != nil {
			return value
		}
	}
	return nil
}

func (b *binder) getMapValue(ctx Context, sf *reflect.StructField, m map[string][]string, name string) interface{} {
	if values := m[name]; len(values) > 0 {
		kind := sf.Type.Kind()
		if kind == reflect.Ptr {
			kind = sf.Type.Elem().Kind()
		}
		if kind != reflect.Slice && kind != reflect.Array {
			return values[0]
		}
	}
	return nil
}

func (b *binder) getFileValue(ctx Context, name string, bindType string) interface{} {
	if f, fh, err := ctx.File(name); err == nil {
		defer f.Close()

		pair := strings.SplitN(bindType, ".", 2)
		if len(pair) == 2 {
			switch pair[1] {
			case "name":
				return fh.Filename
			case "size":
				return fh.Size
			}
		}

		if b, err := ioutil.ReadAll(f); err == nil {
			return b
		}
	}
	return nil
}
