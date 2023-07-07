package goutil

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"time"
)

var ErrTimeout = errors.New("timed out waiting for the condition")
var ErrNotSetDeadline = errors.New("context doesn't set deadline")

var (
	DefaultRetry   = BackoffWait{TotalRuns: 2, BaseDuration: time.Second, Factor: 2.0, JitterFactor: 0.1}
	FastRetry      = BackoffWait{TotalRuns: 2, BaseDuration: 50 * time.Millisecond, Factor: 2.0, JitterFactor: 0.5}
	UnlimitedRetry = BackoffWait{TotalRuns: math.MaxInt32, BaseDuration: time.Second, Factor: 2.0, JitterFactor: 0.5}
	NoRetry        = BackoffWait{TotalRuns: 1}
)

// BackoffWait encapsulates parameters that control the behavior of backoff mechanism.
// TotalRuns denotes the maximum number of times the function is executed,
// BaseDuration is the initial waiting time before function execution,
// Factor is the multiplier for exponential growth of waiting time,
// JitterFactor is the factor for random increase to the waiting time.
type BackoffWait struct {
	TotalRuns    int           `json:"total_runs"`
	BaseDuration time.Duration `json:"base_duration"`
	Factor       float64       `json:"factor"`
	JitterFactor float64       `json:"jitter_factor"`
}

// wait returns a time duration to wait before next function execution.
// It updates the BackoffWait's BaseDuration and TotalRuns after each call.
func (b *BackoffWait) wait() time.Duration {
	if b.TotalRuns < 1 {
		if b.JitterFactor > 0 {
			return addJitter(b.BaseDuration, b.JitterFactor)
		}

		return b.BaseDuration
	}

	b.TotalRuns--

	duration := b.BaseDuration

	if b.Factor != 0 {
		b.BaseDuration = time.Duration(float64(b.BaseDuration) * b.Factor)
	}

	if b.JitterFactor > 0 {
		duration = addJitter(duration, b.JitterFactor)
	}

	return duration
}

// addJitter adds random jitter to the base duration.
func addJitter(base time.Duration, jitterFactor float64) time.Duration {
	if jitterFactor <= 0.0 {
		jitterFactor = 1.0
	}

	return base + time.Duration(rand.Float64()*jitterFactor*float64(base))
}

// RetryableFunc is a function type that can be retried until it succeeds or meets a certain condition.
type RetryableFunc func() (done bool, err error)

// RetryableFuncWithContext is a RetryableFunc that includes context, which can be used for cancellation or passing request-scoped data.
type RetryableFuncWithContext func(context.Context) (done bool, err error)

func (cf RetryableFunc) WithContext() RetryableFuncWithContext {
	return func(context.Context) (done bool, err error) {
		return cf()
	}
}

func exponentialBackoff(backoff BackoffWait, fn RetryableFunc) error {
	for backoff.TotalRuns > 0 {
		if done, err := fn(); err != nil || done {
			return err
		}

		if backoff.TotalRuns == 1 {
			break
		}

		time.Sleep(backoff.wait())
	}

	return ErrTimeout
}

func exponentialBackoffWithCtx(ctx context.Context, backoff BackoffWait, fnWithContext RetryableFuncWithContext) error {
	for backoff.TotalRuns > 0 {
		select {
		case <-ctx.Done():
			return ErrTimeout
		default:
			done, err := fnWithContext(ctx)
			backoff.TotalRuns--

			if err != nil || done {
				return err
			}

			if backoff.TotalRuns <= 0 {
				break
			}

			time.Sleep(backoff.wait())
		}

		if backoff.TotalRuns <= 0 {
			break
		}
	}

	return ErrTimeout
}

// RetryWithExponentialBackoff tries a function with exponential backoff, and return the error from the function or timeout.
func RetryWithExponentialBackoff(backoff BackoffWait, fn RetryableFunc) (err error) {
	waitErr := exponentialBackoff(backoff, func() (bool, error) {
		var done bool
		done, err = fn()
		return done, nil // swallow err in the process
	})

	if waitErr != nil && err == nil {
		return waitErr
	}

	return err
}

// RetryWithExponentialBackoffUntilTimeout tries a function with exponential backoff until it succeeds or the context is cancelled (timeout).
func RetryWithExponentialBackoffUntilTimeout(ctx context.Context, f RetryableFuncWithContext) (err error) {
	if _, ok := ctx.Deadline(); !ok {
		return ErrNotSetDeadline
	}

	backOff := BackoffWait{TotalRuns: math.MaxInt32, BaseDuration: time.Second, Factor: 1.5, JitterFactor: 0.5}
	waitErr := exponentialBackoffWithCtx(ctx, backOff, func(ctx context.Context) (bool, error) {
		var done bool
		done, err = f(ctx)
		return done, nil // swallow err in the process
	})

	if waitErr != nil && err == nil {
		return waitErr
	}

	return err
}
