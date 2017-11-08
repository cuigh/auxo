package times

import (
	"testing"
	"time"

	"github.com/cuigh/auxo/test/assert"
)

func TestFormat(t *testing.T) {
	type TestCase struct {
		Layout1 string
		Layout2 string
	}

	dt := time.Now()
	cases := []TestCase{
		{
			Layout1: "yyyy-MM-ddTHH:mm:ss.fffffffffzzz",
			Layout2: "2006-01-02T15:04:05.000000000-07:00",
		},
	}

	for _, c := range cases {
		assert.Equal(t, dt.Format(c.Layout2), Format(dt, c.Layout1))
	}
}

func TestToday(t *testing.T) {
	today := Today()
	now := time.Now()
	assert.Equal(t, time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local), today)
}

func BenchmarkFormat(b *testing.B) {
	layout := "yyyy-MM-ddTHH:mm:ss.fffffffff"
	dt := time.Now()
	for i := 0; i < b.N; i++ {
		Format(dt, layout)
	}
}

func BenchmarkFormatStd(b *testing.B) {
	layout := "2006-01-02T15:04:05.000000000"
	dt := time.Now()
	for i := 0; i < b.N; i++ {
		dt.Format(layout)
	}
}
