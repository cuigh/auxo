package breaker

import (
	"sync"
)

// Summary represents statistics of Breaker.
type Summary interface {
	// Total is total attempts
	Total() uint32
	// Success is success count
	Success() uint32
	// Failure is failure count
	Failure() uint32
	// Intercept is failure count with ErrCircuitOpen
	Intercept() uint32
	// Reject is failure count with ErrMaxConcurrency
	Reject() uint32
	// Timeout is failure count with ErrTimeout
	Timeout() uint32
	// FallbackSuccess is success count by fallback
	FallbackSuccess() uint32
	// FallbackFailure is failure count by fallback
	FallbackFailure() uint32
}

type Counter interface {
	Summary
	// Failure is consecutive failure count
	ConsecutiveFailure() uint32
}

type summary struct {
	total           uint32
	success         uint32
	failure         uint32
	intercept       uint32
	reject          uint32
	timeout         uint32
	fallbackSuccess uint32
	fallbackFailure uint32
}

func (s summary) Total() uint32 {
	return s.total
}

func (s summary) Success() uint32 {
	return s.success
}

func (s summary) Failure() uint32 {
	return s.failure
}

func (s summary) Intercept() uint32 {
	return s.intercept
}

func (s summary) Reject() uint32 {
	return s.reject
}

func (s summary) Timeout() uint32 {
	return s.timeout
}

func (s summary) FallbackSuccess() uint32 {
	return s.fallbackSuccess
}

func (s summary) FallbackFailure() uint32 {
	return s.fallbackFailure
}

func (s summary) reset() {
	s.total = 0
	s.success = 0
	s.failure = 0
	s.reject = 0
	s.timeout = 0
	s.fallbackSuccess = 0
	s.fallbackFailure = 0
}

type counter struct {
	summary
	consecutiveFailure uint32
}

func (c counter) ConsecutiveFailure() uint32 {
	return c.consecutiveFailure
}

func (c counter) reset() {
	c.summary.reset()
	c.consecutiveFailure = 0
}

type statistics struct {
	sync.Mutex
	summary
	counter
}

func (s *statistics) reset(summary, counter bool) {
	s.Lock()
	if summary {
		s.summary.reset()
	}
	if counter {
		s.counter.reset()
	}
	s.Unlock()
}

func (s *statistics) update(final, middle error, fallback bool) {
	s.Lock()

	s.summary.total++
	s.counter.total++
	if final == nil {
		s.summary.success++
		s.counter.success++
		s.consecutiveFailure = 0
		if fallback {
			s.summary.fallbackSuccess++
			s.counter.fallbackSuccess++
		}
	} else {
		s.summary.failure++
		s.counter.failure++
		s.consecutiveFailure++
		if fallback {
			s.summary.fallbackFailure++
			s.counter.fallbackFailure++
		}
		if final == ErrCircuitOpen {
			s.summary.intercept++
			s.counter.intercept++
		} else if middle == ErrMaxConcurrency {
			s.summary.reject++
			s.counter.reject++
		} else if middle == ErrTimeout {
			s.summary.timeout++
			s.counter.timeout++
		}
	}
	s.Unlock()
}
