package retry

import (
	"math"
	"time"
)

// Backoff defines a function that control the back-off duration between call retries.
type Backoff func(attempt int) time.Duration

// WithJitter creates a new Backoff with Jitter attached.
func (b Backoff) WithJitter(jitter Jitter) Backoff {
	return func(attempt int) time.Duration {
		return jitter(b(attempt))
	}
}

// Fixed creates a Backoff with fixed interval between calls.
func Fixed(interval time.Duration) Backoff {
	return func(_ int) time.Duration {
		return interval
	}
}

// Scheduled creates a Backoff with scheduled periods of time between calls.
func Scheduled(intervals ...time.Duration) Backoff {
	l := len(intervals)
	return func(attempt int) time.Duration {
		if attempt < l {
			return intervals[attempt-1]
		}
		return 0
	}
}

// Linear creates a Backoff that linearly multiplies the interval by the attempt number for each attempt.
func Linear(initial, interval time.Duration) Backoff {
	return func(attempt int) time.Duration {
		return initial + interval*time.Duration(attempt)
	}
}

// Exponential creates a Backoff that multiplies the interval by
// an exponentially increasing factor for each attempt.
func Exponential(interval time.Duration, base float64) Backoff {
	return func(attempt int) time.Duration {
		return interval * time.Duration(math.Pow(base, float64(attempt)))
	}
}
