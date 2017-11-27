package gsd

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"
	"strconv"
	"time"
	"unsafe"

	"github.com/cuigh/auxo/ext/reflects"
)

// Scanner valid types: int64, float64, bool, []byte, string, time.NullTime, nil - for NULL values

var timeFormats = []string{
	"2006-01-02 15:04:05.999999999",
	"2006-01-02T15:04:05.999999999",
	"2006-01-02 15:04:05",
	"2006-01-02T15:04:05",
	"2006-01-02 15:04",
	"2006-01-02T15:04",
	"2006-01-02",
	"2006-01-02 15:04:05-07:00",
}

// NullTime is a nullable NullTime value
type NullTime struct {
	Time  time.Time
	Valid bool // Valid is true if NullTime is not NULL
}

// Scan implements the Scanner interface.
func (n *NullTime) Scan(value interface{}) error {
	switch t := value.(type) {
	case nil:
		fmt.Println("nil value")
	case time.Time:
		n.Time, n.Valid = t, true
	case []byte:
		n.Valid = false
		for _, f := range timeFormats {
			var err error
			if n.Time, err = time.Parse(f, string(t)); err == nil {
				n.Valid = true
				break
			}
		}
	default:
		return fmt.Errorf("null: can't convert %T to time.NullTime", value)
	}
	return nil
}

// Value implements the driver Valuer interface.
func (n NullTime) Value() (driver.Value, error) {
	if n.Valid {
		return n.Time, nil
	}
	return nil, nil
}

type NullInt32 struct {
	Int32 int32
	Valid bool // Valid is true if Int32 is not NULL
}

// Scan implements the Scanner interface.
func (n *NullInt32) Scan(value interface{}) (err error) {
	if value == nil {
		n.Int32, n.Valid = 0, false
		return
	}

	switch v := value.(type) {
	case int64:
		n.Int32, n.Valid = int32(v), true
		return
	case []uint8:
		if i, e := strconv.ParseInt(string(v), 10, 32); e == nil {
			n.Int32, n.Valid = int32(i), true
		}
		return
	}

	n.Valid = false
	return fmt.Errorf("null: can't convert %T to int32", value)
}

// Value implements the driver Valuer interface.
func (n NullInt32) Value() (driver.Value, error) {
	if n.Valid {
		return n.Int32, nil
	}
	return nil, nil
}

/********** NullFloat32 **********/

type NullFloat32 struct {
	Float32 float32
	Valid   bool // Valid is true if Float32 is not NULL
}

// Scan implements the Scanner interface.
func (n *NullFloat32) Scan(value interface{}) (err error) {
	if value == nil {
		n.Float32, n.Valid = 0, false
		return
	}

	switch v := value.(type) {
	case float64:
		n.Float32, n.Valid = float32(v), true
		return
	case []uint8:
		if i, e := strconv.ParseFloat(string(v), 10); e == nil {
			n.Float32, n.Valid = float32(i), true
		}
		return
	}

	n.Valid = false
	return fmt.Errorf("null: can't convert %T to float32", value)
}

// Value implements the driver Valuer interface.
func (n NullFloat32) Value() (driver.Value, error) {
	if n.Valid {
		return n.Float32, nil
	}
	return nil, nil
}

func init() {
	reflects.RegisterFieldAccessor(typeNullBool, &reflects.FieldAccessor{
		Get: func(ptr uintptr, f *reflect.StructField) interface{} {
			return *((*sql.NullBool)(unsafe.Pointer(ptr + f.Offset)))
		},
		Set: func(ptr uintptr, f *reflect.StructField, v interface{}) {
			*((*sql.NullBool)(unsafe.Pointer(ptr + f.Offset))) = v.(sql.NullBool)
		},
		GetPointer: func(ptr uintptr, f *reflect.StructField) interface{} {
			return (*sql.NullBool)(unsafe.Pointer(ptr + f.Offset))
		},
	})
	reflects.RegisterFieldAccessor(typeNullString, &reflects.FieldAccessor{
		Get: func(ptr uintptr, f *reflect.StructField) interface{} {
			return *((*sql.NullString)(unsafe.Pointer(ptr + f.Offset)))
		},
		Set: func(ptr uintptr, f *reflect.StructField, v interface{}) {
			*((*sql.NullString)(unsafe.Pointer(ptr + f.Offset))) = v.(sql.NullString)
		},
		GetPointer: func(ptr uintptr, f *reflect.StructField) interface{} {
			return (*sql.NullString)(unsafe.Pointer(ptr + f.Offset))
		},
	})
	reflects.RegisterFieldAccessor(typeNullTime, &reflects.FieldAccessor{
		Get: func(ptr uintptr, f *reflect.StructField) interface{} {
			return *((*NullTime)(unsafe.Pointer(ptr + f.Offset)))
		},
		Set: func(ptr uintptr, f *reflect.StructField, v interface{}) {
			*((*NullTime)(unsafe.Pointer(ptr + f.Offset))) = v.(NullTime)
		},
		GetPointer: func(ptr uintptr, f *reflect.StructField) interface{} {
			return (*NullTime)(unsafe.Pointer(ptr + f.Offset))
		},
	})
	reflects.RegisterFieldAccessor(typeNullInt32, &reflects.FieldAccessor{
		Get: func(ptr uintptr, f *reflect.StructField) interface{} {
			return *((*NullInt32)(unsafe.Pointer(ptr + f.Offset)))
		},
		Set: func(ptr uintptr, f *reflect.StructField, v interface{}) {
			*((*NullInt32)(unsafe.Pointer(ptr + f.Offset))) = v.(NullInt32)
		},
		GetPointer: func(ptr uintptr, f *reflect.StructField) interface{} {
			return (*NullInt32)(unsafe.Pointer(ptr + f.Offset))
		},
	})
	reflects.RegisterFieldAccessor(typeNullInt64, &reflects.FieldAccessor{
		Get: func(ptr uintptr, f *reflect.StructField) interface{} {
			return *((*sql.NullInt64)(unsafe.Pointer(ptr + f.Offset)))
		},
		Set: func(ptr uintptr, f *reflect.StructField, v interface{}) {
			*((*sql.NullInt64)(unsafe.Pointer(ptr + f.Offset))) = v.(sql.NullInt64)
		},
		GetPointer: func(ptr uintptr, f *reflect.StructField) interface{} {
			return (*sql.NullInt64)(unsafe.Pointer(ptr + f.Offset))
		},
	})
	reflects.RegisterFieldAccessor(typeNullFloat32, &reflects.FieldAccessor{
		Get: func(ptr uintptr, f *reflect.StructField) interface{} {
			return *((*NullFloat32)(unsafe.Pointer(ptr + f.Offset)))
		},
		Set: func(ptr uintptr, f *reflect.StructField, v interface{}) {
			*((*NullFloat32)(unsafe.Pointer(ptr + f.Offset))) = v.(NullFloat32)
		},
		GetPointer: func(ptr uintptr, f *reflect.StructField) interface{} {
			return (*NullFloat32)(unsafe.Pointer(ptr + f.Offset))
		},
	})
	reflects.RegisterFieldAccessor(typeNullFloat64, &reflects.FieldAccessor{
		Get: func(ptr uintptr, f *reflect.StructField) interface{} {
			return *((*sql.NullFloat64)(unsafe.Pointer(ptr + f.Offset)))
		},
		Set: func(ptr uintptr, f *reflect.StructField, v interface{}) {
			*((*sql.NullFloat64)(unsafe.Pointer(ptr + f.Offset))) = v.(sql.NullFloat64)
		},
		GetPointer: func(ptr uintptr, f *reflect.StructField) interface{} {
			return (*sql.NullFloat64)(unsafe.Pointer(ptr + f.Offset))
		},
	})
}

type Value struct {
	bytes []byte
	err   error
}

func (v Value) String() (string, error) {
	if v.err != nil {
		return "", v.err
	}
	return string(v.bytes), nil
}

func (v Value) Int() (i int, err error) {
	if v.err != nil {
		return 0, v.err
	}
	return strconv.Atoi(string(v.bytes))
}

func (v Value) Int32() (int32, error) {
	if v.err != nil {
		return 0, v.err
	}
	i, err := strconv.Atoi(string(v.bytes))
	if err != nil {
		return 0, err
	}
	return int32(i), nil
}

func (v Value) Int64() (i int64, err error) {
	if v.err != nil {
		return 0, v.err
	}
	return strconv.ParseInt(string(v.bytes), 10, 32)
}
