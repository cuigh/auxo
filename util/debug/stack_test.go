package debug_test

import (
	"testing"

	"github.com/cuigh/auxo/util/debug"
)

func TestStack(t *testing.T) {
	t.Log(string(debug.Stack()))
}

func TestStackSkip(t *testing.T) {
	t.Log(string(debug.StackSkip(1)))
}
