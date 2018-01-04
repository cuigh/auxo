package log

import (
	"testing"

	"github.com/cuigh/auxo/test/assert"
)

func TestTextLayout_Parse(t *testing.T) {
	cases := []struct {
		layout string
		fields int
	}{
		{"[{L}]{T}: {M}{N}", 7},
		{"[{level: short}]{time}: {msg}{newline}", 7},
	}

	for _, c := range cases {
		fields, err := textLayout{}.Parse(c.layout)
		assert.NoError(t, err)
		assert.Equal(t, c.fields, len(fields))
	}
}

func TestJSONLayout_Parse(t *testing.T) {
	const layout = "{level->lvl: long},{time->t: 2016-01-02},{msg->msg},{file->f: full},{text->abc: test}"

	fields, err := jsonLayout{}.Parse(layout)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(fields))
}
