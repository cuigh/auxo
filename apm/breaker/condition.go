package breaker

// Condition is a monitor for circuit, it return true if should open circuit.
type Condition func(c Counter) bool

func ErrorRate(rate float32, min uint32) Condition {
	return func(c Counter) bool {
		return c.Total() > 0 && c.Total() >= min && float32(c.Failure()) >= float32(c.Total())*rate
	}
}

func ErrorCount(count uint32) Condition {
	return func(c Counter) bool {
		return c.Failure() >= count
	}
}

func ConsecutiveErrorCount(count uint32) Condition {
	return func(c Counter) bool {
		return c.ConsecutiveFailure() >= count
	}
}
