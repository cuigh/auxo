package reflects_test

import (
	"errors"
	"reflect"
	"testing"
	"unsafe"

	"github.com/cuigh/auxo/ext/reflects"
	"github.com/cuigh/auxo/test/assert"
)

func TestPointer(t *testing.T) {
	var i interface{} = 1
	ptr := reflects.Pointer(i)
	assert.Equal(t, 1, *(*int)(ptr))
}

func TestInterface(t *testing.T) {
	type TestObj struct {
		field1 string
	}

	struct_ := &TestObj{}
	field, _ := reflect.TypeOf(struct_).Elem().FieldByName("field1")
	field1Ptr := uintptr(unsafe.Pointer(struct_)) + field.Offset
	*((*string)(unsafe.Pointer(field1Ptr))) = "hello"
	t.Log(struct_)

	structInter := (interface{})(struct_)
	structPtr := reflects.Pointer(&structInter) //(*emptyInterface)(unsafe.Pointer(&structInter)).word
	field, _ = reflect.TypeOf(structInter).Elem().FieldByName("field1")
	field1Ptr = uintptr(structPtr) + field.Offset
	*((*string)(unsafe.Pointer(field1Ptr))) = "hello"
	t.Log(struct_)
}

func TestStruct(t *testing.T) {
	fields := []reflect.StructField{
		{Name: "A", Type: reflects.TypeInt},
		{Name: "B", Type: reflects.TypeDuration},
	}
	st := reflect.StructOf(fields)
	t.Log(st)
}

func BenchmarkNew1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		v := int(1)
		if v > 0 {
		}
	}
}

func BenchmarkNew2(b *testing.B) {
	t := reflect.TypeOf(1)
	for i := 0; i < b.N; i++ {
		v := reflect.New(t).Elem().Int()
		if v > 0 {
		}
	}
}

func TestProxy(t *testing.T) {
	type UserServiceClient struct {
		Find      func(id int32) bool
		FindAsync func(id int32) chan struct{}
	}
	c := UserServiceClient{}
	err := Proxy(&c.Find)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(c.Find(1))
	t.Log(c.Find(2))
}

func Proxy(fn interface{}) error {
	ft := reflect.TypeOf(fn)
	if ft.Kind() != reflect.Ptr {
		return errors.New("fn must be func pointer")
	}
	ft = ft.Elem()
	if ft.Kind() != reflect.Func {
		return errors.New("fn must be func pointer")
	}
	f := reflect.MakeFunc(ft, func(ins []reflect.Value) (outs []reflect.Value) {
		//fmt.Println("ins: ", ins)
		outs = make([]reflect.Value, 1)
		outs[0] = reflect.ValueOf(ins[0].Int() == 1)
		//fmt.Println("outs: ", outs)
		return
	})
	reflect.ValueOf(fn).Elem().Set(f)
	return nil
}

func TestConnect(t *testing.T) {
	type Hello struct {
		Hello func(n string)
	}

	d := &Hello{}
	s := hello(0)
	err := reflects.Connect(d, s)
	if err != nil {
		t.Fatal(err)
	}

	d.Hello("world!")
}

type hello int

func (h hello) Hello(n string) {
	//fmt.Println("Hello,", n)
}

func BenchmarkConnect1(b *testing.B) {
	type Hello struct {
		Hello func(n string)
	}

	d := &Hello{}
	err := reflects.Connect(d, hello(0))
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		d.Hello("world!")
	}
}

func BenchmarkConnect2(b *testing.B) {
	type Hello struct {
		Hello func(n string)
	}

	d := &Hello{}
	d.Hello = hello(0).Hello

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		d.Hello("world!")
	}
}

func TestSimpleValue(t *testing.T) {
	v := struct {
		Int    int
		IntPtr *int
	}{}
	cases := []struct {
		Value  reflect.Value
		Value1 int
		Value2 int
	}{
		{reflect.ValueOf(&v).Elem().FieldByName("Int"), 1, 2},
		{reflect.ValueOf(&v).Elem().FieldByName("IntPtr"), 1, 2},
	}

	for _, c := range cases {
		v := reflects.SimpleOf(c.Value)

		v.Set(c.Value1)
		assertEqual(t, c.Value1, c.Value.Interface())

		v.SetValue(reflect.ValueOf(c.Value2))
		assertEqual(t, c.Value2, c.Value.Interface())
	}
}

func TestIsEmpty(t *testing.T) {
	i, j := 0, 1
	var k *int
	cases := []struct {
		Value    interface{}
		Expected bool
	}{
		{true, false},
		{false, true},
		{"a", false},
		{"", true},
		{0, true},
		{1, false},
		{0.0, true},
		{1.0, false},
		{&i, true},
		{&j, false},
		{k, true},
	}

	for _, c := range cases {
		assert.Equal(t, c.Expected, reflects.IsEmpty(reflect.ValueOf(c.Value)))
	}
}
