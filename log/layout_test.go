package log

import (
	"testing"

	"github.com/cuigh/auxo/test/assert"
)

func TestJsonLayout_Parse(t *testing.T) {
	const layout = "{level->lvl: a=b},{time->t:2016-01-02},{msg->msg},{file->f: s},{text->abc: test}"

	fields, err := jsonLayout{}.Parse(layout)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(fields))
}
