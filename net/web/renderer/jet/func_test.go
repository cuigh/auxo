package jet

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/CloudyKit/jet"
	"github.com/cuigh/auxo/test/assert"
)

func TestEqual(t *testing.T) {
	cases := []struct {
		X        interface{}
		Y        interface{}
		Expected bool
	}{
		{1, 1, true},
		{1, 0, false},
		{1, "1", true},
		//{1, nil, false},
		//{"", nil, true},
	}

	set := jet.NewHTMLSet()
	set.AddGlobalFunc("eq", equal)

	tpl, err := set.LoadTemplate("test", `{{ if eq(.X, .Y) }}true{{ else }}false{{ end }}`)
	assert.NoError(t, err)

	for _, c := range cases {
		buf := &bytes.Buffer{}
		err = tpl.Execute(buf, nil, c)
		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprint(c.Expected), buf.String())
	}
}
