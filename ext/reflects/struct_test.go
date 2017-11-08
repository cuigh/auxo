package reflects_test

import (
	"testing"

	"github.com/cuigh/auxo/ext/reflects"
)

func TestStructTag_All(t *testing.T) {
	tag := reflects.StructTag(`json:"a" xml:"a"`)
	m := tag.All()
	t.Log(m)
}
