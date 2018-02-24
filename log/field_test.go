package log

import (
	"bytes"
	"testing"
	"time"

	"github.com/cuigh/auxo/test/assert"
)

func TestStringField(t *testing.T) {
	const value = "test"

	buf := new(bytes.Buffer)
	f := newStringField("", value)
	err := f.Write(buf, nil)
	assert.NoError(t, err)
	assert.Equal(t, value, buf.String())
}

func TestLevelField(t *testing.T) {
	cases := [...]struct {
		Level Level
		Short string
		Long  string
	}{
		{LevelDebug, levelShortNames[0], levelLongNames[0]},
		{LevelInfo, levelShortNames[1], levelLongNames[1]},
		{LevelWarn, levelShortNames[2], levelLongNames[2]},
		{LevelError, levelShortNames[3], levelLongNames[3]},
		{LevelPanic, levelShortNames[4], levelLongNames[4]},
		{LevelFatal, levelShortNames[5], levelLongNames[5]},
		{LevelOff, levelShortNames[6], levelLongNames[6]},
	}

	buf := new(bytes.Buffer)

	// short name
	f := newLevelField("level")
	for i, c := range cases {
		e := &entry{
			lvl: c.Level,
		}
		err := f.Write(buf, e)
		assert.NoError(t, err)
		assert.Equal(t, levelShortNames[i], buf.String())
		buf.Reset()
	}

	// long name
	f = newLevelField("level", "L")
	for i, c := range cases {
		e := &entry{
			lvl: c.Level,
		}
		err := f.Write(buf, e)
		assert.NoError(t, err)
		assert.Equal(t, levelLongNames[i], buf.String())
		buf.Reset()
	}
}

func TestTimeField(t *testing.T) {
	const layout = "2006-01-02 15:04:05.000"
	buf := new(bytes.Buffer)
	e := &entry{
		time: time.Now(),
	}
	f := newTimeField("time", layout)
	err := f.Write(buf, e)
	assert.NoError(t, err)
	assert.Equal(t, e.time.Format(layout), buf.String())
}

func TestMessageField(t *testing.T) {
	var value = "test"
	buf := new(bytes.Buffer)
	e := &entry{
		msg: value,
	}
	f := newMessageField("msg")
	err := f.Write(buf, e)
	assert.NoError(t, err)
	assert.Equal(t, value, buf.String())
}

func TestFileField(t *testing.T) {
	var value = "field_test.go:"

	buf := new(bytes.Buffer)
	f := newFileField("file", "S")
	f.skip = 2
	err := f.Write(buf, nil)
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), value)

	buf.Reset()
	f = newFileField("file", "F")
	f.skip = 2
	err = f.Write(buf, nil)
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), value)
}
