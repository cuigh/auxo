package times

import (
	"fmt"
	"testing"
	"time"

	"github.com/cuigh/auxo/test/assert"
)

func TestWheel(t *testing.T) {
	w := NewWheel(time.Millisecond*50, 5)
	i := 0
	expected := 3
	w.Add(func() bool {
		fmt.Println(time.Now())
		i++
		return i < expected
	})
	<-time.After(time.Second * 1)
	w.Stop()
	assert.Equal(t, expected, i)
}
