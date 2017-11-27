package retry

import (
	"math"
	"math/rand"
	"time"
)

type Jitter func(d time.Duration) time.Duration

func Deviation(factor float64) Jitter {
	return func(d time.Duration) time.Duration {
		min := int64(math.Floor(float64(d) * (1 - factor)))
		max := int64(math.Ceil(float64(d) * (1 + factor)))
		return time.Duration(rand.Int63n(max-min) + min)
	}
}
