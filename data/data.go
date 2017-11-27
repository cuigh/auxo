package data

type Value interface {
	IsNil() bool
	Scan(i interface{}) error
	Bytes() ([]byte, error)
	Bool() (bool, error)
	String() (string, error)
	//Time() (time.Time, error)
	//Duration() (time.Duration, error)
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
