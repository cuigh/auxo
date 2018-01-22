package log

import (
	"testing"

	"github.com/cuigh/auxo/test/assert"
)

func TestTextLayout_Parse(t *testing.T) {
	cases := []struct {
		layout   string
		segments int
	}{
		{"[{L}]{T}: {M}{N}", 7},
		{"[{level: short}]{time}: {msg}{newline}", 7},
	}

	for _, c := range cases {
		segments, err := TextLayout{}.Parse(c.layout)
		assert.NoError(t, err)
		assert.Equal(t, c.segments, len(segments))
	}
}

func TestJSONLayout_Parse(t *testing.T) {
	const layout = "{level->lvl: long},{time->t: 2016-01-02},{msg->msg},{file->f: full},{text->abc: test}"

	segments, err := JSONLayout{}.Parse(layout)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(segments))
}
