package data

import "errors"

var (
	Nil         Value = nilValue{}
	ErrNilValue       = errors.New("nil value")
)

type Value interface {
	IsNil() bool
	Scan(i interface{}) error
	Bytes() ([]byte, error)
	Bool() (bool, error)
	String() (string, error)
	Int() (int, error)
	Int8() (int8, error)
	Int16() (int16, error)
	Int32() (int32, error)
	Int64() (int64, error)
	Uint() (uint, error)
	Uint8() (uint8, error)
	Uint16() (uint16, error)
	Uint32() (uint32, error)
	Uint64() (uint64, error)
	Float32() (float32, error)
	Float64() (float64, error)
	//Time() (time.Time, error)
	//Duration() (time.Duration, error)
}

type nilValue struct {
}

func (nilValue) IsNil() bool {
	return true
}

func (nilValue) Scan(i interface{}) error {
	return ErrNilValue
}

func (nilValue) Bytes() ([]byte, error) {
	return nil, ErrNilValue
}

func (nilValue) Bool() (bool, error) {
	return false, ErrNilValue
}

func (nilValue) String() (string, error) {
	return "", ErrNilValue
}

func (nilValue) Int() (int, error) {
	return 0, ErrNilValue
}

func (nilValue) Int8() (int8, error) {
	return 0, ErrNilValue
}

func (nilValue) Int16() (int16, error) {
	return 0, ErrNilValue
}

func (nilValue) Int32() (int32, error) {
	return 0, ErrNilValue
}

func (nilValue) Int64() (int64, error) {
	return 0, ErrNilValue
}

func (nilValue) Uint() (uint, error) {
	return 0, ErrNilValue
}

func (nilValue) Uint8() (uint8, error) {
	return 0, ErrNilValue
}

func (nilValue) Uint16() (uint16, error) {
	return 0, ErrNilValue
}

func (nilValue) Uint32() (uint32, error) {
	return 0, ErrNilValue
}

func (nilValue) Uint64() (uint64, error) {
	return 0, ErrNilValue
}

func (nilValue) Float32() (float32, error) {
	return 0, ErrNilValue
}

func (nilValue) Float64() (float64, error) {
	return 0, ErrNilValue
}

type Option struct {
	Name  string `json:"name" xml:"name,attr"`
	Value string `json:"value" xml:"value,attr"`
}

type Options []Option

func (opts Options) Get(name string) string {
	for _, opt := range opts {
		if opt.Name == name {
			return opt.Value
		}
	}
	return ""
}
