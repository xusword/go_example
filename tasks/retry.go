package tasks

import (
	"fmt"
	"time"
)

// ErrorMaxRetryReached means maximum number of retry reached
var ErrorMaxRetryReached = fmt.Errorf("Maximum number of retry reached")

// ErrorTaskCancelled means task is cancelled with a token
var ErrorTaskCancelled = fmt.Errorf("Task is cancelled")

// RetryPolicy => given the number of iteration already tried,
// determine whether more retry is needed and for how long
type RetryPolicy func() (time.Duration, error)

// CancellationToken => Technically any value (include nil) can be sent
// to the channel but just don't. Use Cancel() instead because nil error
// sent to cancellation token will cause dependent to misbehave
type CancellationToken chan error

// NewCancellationToken makes things look more neat, you can use the raw type if you want
func NewCancellationToken() CancellationToken {
	return make(CancellationToken)
}

// Cancel function is also here just to make things look more neat
func (c CancellationToken) Cancel() {
	c <- ErrorTaskCancelled
}

// FixedDuration is the basic retry policy where you always
// wait for the same amount of time for a fix number of times
// the closure returned by this method is not meant for reuse
// by different invokations of RetryOperation
func FixedDuration(retryPeriod time.Duration, maxRetry int) RetryPolicy {
	count := 0
	return func() (time.Duration, error) {
		var err error
		count++
		if count >= maxRetry {
			err = ErrorMaxRetryReached
		}
		return retryPeriod, err
	}
}

// RetryOperation help you make your code look cleaner but I am not here
// to protect you from infinite loops
func RetryOperation(operation func() error, retryPolicy RetryPolicy, token CancellationToken) error {
	for {
		err := operation()
		if err == nil {
			return nil
		}
		duration, policyViolation := retryPolicy()
		if policyViolation != nil {
			return policyViolation
		}
		if err := Wait(duration, token); err != nil {
			return err
		}
	}
}

// Wait for a specific amount of time, but cancel any time
func Wait(duration time.Duration, token CancellationToken) error {
	select {
	case <-time.After(duration):
		return nil
	case err := <-token:
		return err
	}
}
