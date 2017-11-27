package retry_test

import (
	"testing"
	"time"

	"github.com/cuigh/auxo/test/assert"
	"github.com/cuigh/auxo/util/retry"
)

func TestDo(t *testing.T) {
	cases := []retry.Backoff{
		retry.Fixed(time.Millisecond * 100),
		retry.Scheduled(time.Millisecond*100, time.Millisecond*300, time.Millisecond*500),
		retry.Linear(0, time.Millisecond*100),
		retry.Exponential(time.Millisecond*100, 2),
	}
	for _, c := range cases {
		err := retry.Do(3, c, func() error {
			return nil
		})
		assert.NoError(t, err)
	}
}

func TestWithJitter(t *testing.T) {
	err := retry.Do(3, retry.Fixed(time.Second).WithJitter(retry.Deviation(0.5)), func() error {
		return nil
	})
	assert.NoError(t, err)
}
