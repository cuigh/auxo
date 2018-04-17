package retry

import "time"

func Do(count int, backoff Backoff, fn func() error) error {
	var err error
	for i := 0; i < count; i++ {
		if i > 0 && backoff != nil {
			if d := backoff(i); d > 0 {
				time.Sleep(d)
			}
		}

		err = fn()
		if err == nil {
			break
		}
	}
	return err
}
