package redis

import (
	"bytes"
	"encoding"
	"encoding/gob"
	"time"

	"github.com/cuigh/auxo/cache"
	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/db/redis"
	"github.com/cuigh/auxo/errors"
	"github.com/cuigh/auxo/util/cast"
)

var ErrNilValue = errors.New("value is nil")

type Provider struct {
	client redis.Client
}

func NewProvider(opts data.Map) (*Provider, error) {
	db := cast.ToString(opts.Get("db"))
	if db == "" {
		db = "cache"
	}
	cmd, err := redis.Open(db)
	if err != nil {
		return nil, err
	}
	return &Provider{client: cmd}, nil
}

func (p *Provider) Exist(key string) (bool, error) {
	i, err := p.client.Exists(key).Result()
	return i == 1, err
}

func (p *Provider) Get(key string) (v data.Value, err error) {
	cmd := p.client.Get(key)
	err = cmd.Err()
	if err == nil {
		return (*value)(cmd), nil
	} else if err.Error() == "redis: nil" { // HACK: redis driver should expose Nil error
		return (*value)(nil), nil
	}
	//} else if err == redis.Nil {
	//	return (*value)(nil), nil
	//}
	return nil, err
}

func (p *Provider) Remove(key string) error {
	return p.client.Del(key).Err()
}

func (p *Provider) Set(key string, value interface{}, expiry time.Duration) error {
	v, err := p.encode(value)
	if err != nil {
		return err
	}
	return p.client.Set(key, v, expiry).Err()
}

func (p *Provider) encode(value interface{}) (r interface{}, err error) {
	switch v := value.(type) {
	case nil, bool, string, []byte, float32, float64, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		r = v
	case encoding.BinaryMarshaler:
		r = v
	default:
		buf := new(bytes.Buffer)
		encoder := gob.NewEncoder(buf)
		err = encoder.Encode(value)
		if err == nil {
			r = buf.Bytes()
		}
	}
	return
}

type value redis.StringCmd

func (v *value) IsNil() bool {
	if v != nil {
		b, _ := v.cmd().Bytes()
		return len(b) == 0
	}
	return true
}

func (v *value) Scan(i interface{}) error {
	if v == nil {
		return ErrNilValue
	}

	cmd := (*redis.StringCmd)(v)
	switch i.(type) {
	case nil, *bool, *string, *[]byte, *float32, *float64, *int, *int8, *int16, *int32, *int64, *uint, *uint8, *uint16, *uint32, *uint64:
		return cmd.Scan(i)
	case encoding.BinaryUnmarshaler:
		return cmd.Scan(i)
	default:
		b, err := cmd.Bytes()
		if err != nil {
			return err
		}
		encoder := gob.NewDecoder(bytes.NewBuffer(b))
		return encoder.Decode(i)
	}
}

func (v *value) Bytes() (b []byte, err error) {
	if v == nil {
		err = ErrNilValue
		return
	}

	return v.cmd().Bytes()
}

func (v *value) Bool() (b bool, err error) {
	if v == nil {
		err = ErrNilValue
		return
	}

	err = v.cmd().Scan(&b)
	return
}

func (v *value) Int() (i int, err error) {
	if v == nil {
		err = ErrNilValue
	} else {
		err = v.cmd().Scan(&i)
	}
	return
}

func (v *value) Int8() (i int8, err error) {
	if v == nil {
		err = ErrNilValue
	} else {
		err = v.cmd().Scan(&i)
	}
	return
}

func (v *value) Int16() (i int16, err error) {
	if v == nil {
		err = ErrNilValue
	} else {
		err = v.cmd().Scan(&i)
	}
	return
}

func (v *value) Int32() (i int32, err error) {
	if v == nil {
		err = ErrNilValue
	} else {
		err = v.cmd().Scan(&i)
	}
	return
}

func (v *value) Int64() (i int64, err error) {
	if v == nil {
		return 0, ErrNilValue
	}
	return v.cmd().Int64()
}

func (v *value) Uint() (i uint, err error) {
	if v == nil {
		err = ErrNilValue
	} else {
		err = v.cmd().Scan(&i)
	}
	return
}

func (v *value) Uint8() (i uint8, err error) {
	if v == nil {
		err = ErrNilValue
	} else {
		err = v.cmd().Scan(&i)
	}
	return
}

func (v *value) Uint16() (i uint16, err error) {
	if v == nil {
		err = ErrNilValue
	} else {
		err = v.cmd().Scan(&i)
	}
	return
}

func (v *value) Uint32() (i uint32, err error) {
	if v == nil {
		err = ErrNilValue
	} else {
		err = v.cmd().Scan(&i)
	}
	return
}

func (v *value) Uint64() (i uint64, err error) {
	if v == nil {
		return 0, ErrNilValue
	}
	return v.cmd().Uint64()
}

func (v *value) Float32() (f float32, err error) {
	if v == nil {
		err = ErrNilValue
	} else {
		err = v.cmd().Scan(&f)
	}
	return
}

func (v *value) Float64() (f float64, err error) {
	if v == nil {
		return 0, ErrNilValue
	}
	return v.cmd().Float64()
}

func (v *value) String() (s string, err error) {
	if v == nil {
		err = ErrNilValue
		return
	}

	return v.cmd().String(), nil
}

func (v *value) cmd() *redis.StringCmd {
	return (*redis.StringCmd)(v)
}

func init() {
	cache.Register("redis", func(opts data.Map) (cache.Provider, error) {
		return NewProvider(opts)
	})
}
