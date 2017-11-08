package i18n_test

import (
	"testing"

	"github.com/cuigh/auxo/test/assert"
	"github.com/cuigh/auxo/util/i18n"
)

func TestAll(t *testing.T) {
	m, err := i18n.New(".").All()
	assert.NoError(t, err)

	msg := m.Get("en").Format("welcome", "auxo")
	t.Log(msg)

	msg = m.Get("zh").Format("welcome", "auxo")
	t.Log(msg)

	msg = m.Find("zh-CN").Format("welcome", "auxo")
	t.Log(msg)
}

func TestCombine(t *testing.T) {
	m, err := i18n.New(".").All()
	assert.NoError(t, err)

	trans := i18n.Combine(m.Get("zh"), m.Get("en"))
	assert.NoError(t, err)

	msg := trans.Get("name")
	t.Log(msg)

	msg = trans.Format("welcome", "auxo")
	t.Log(msg)
}
