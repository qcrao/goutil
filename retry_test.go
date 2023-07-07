package goutil

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestRetryTimeDuration(t *testing.T) {
	retries := []struct {
		name  string
		retry BackoffWait
	}{
		{name: "NoRetry", retry: NoRetry},
		{name: "DefaultRetry", retry: DefaultRetry},
		{name: "FastRetry", retry: FastRetry},
		{name: "UnlimitedRetry", retry: UnlimitedRetry},
	}

	for _, r := range retries {
		if r.name == "UnlimitedRetry" {
			r.retry.TotalRuns = 5
		}

		t.Run(r.name, func(t *testing.T) {
			fmt.Printf("Retry strategy: %s\n", r.name)
			for i := 0; r.retry.TotalRuns >= 2; i++ {
				fmt.Printf("ran %d times. next run wait duration: %v\n", i+1, r.retry.wait())
			}
			fmt.Println()
		})
	}
}

func TestRetryWithExponentialBackoff(t *testing.T) {
	tests := []struct {
		name        string
		backoff     BackoffWait
		fn          RetryableFunc
		wantErr     bool
		expectedErr error
	}{
		{
			name:        "retry succeeds",
			backoff:     DefaultRetry,
			fn:          func() (bool, error) { return true, nil },
			wantErr:     false,
			expectedErr: nil,
		},
		{
			name:        "retry exceeds limit",
			backoff:     NoRetry,
			fn:          func() (bool, error) { return false, nil },
			wantErr:     true,
			expectedErr: ErrTimeout,
		},
		{
			name:        "function returns error",
			backoff:     DefaultRetry,
			fn:          func() (bool, error) { return false, errors.New("custom error") },
			wantErr:     true,
			expectedErr: errors.New("custom error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RetryWithExponentialBackoff(tt.backoff, tt.fn)

			if (err != nil) != tt.wantErr {
				t.Errorf("RetryWithExponentialBackoff() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && err.Error() != tt.expectedErr.Error() {
				t.Errorf("RetryWithExponentialBackoff() got = %v, want %v", err, tt.expectedErr)
			}
		})
	}
}

func TestRetryWithExponentialBackoffUntilTimeout(t *testing.T) {
	timeoutCtx, _ := context.WithTimeout(context.Background(), 1*time.Millisecond)
	tests := []struct {
		name        string
		ctx         context.Context
		fn          RetryableFuncWithContext
		wantErr     bool
		expectedErr error
	}{
		{
			name:        "retry succeeds",
			ctx:         timeoutCtx,
			fn:          func(context.Context) (bool, error) { return true, nil },
			wantErr:     false,
			expectedErr: nil,
		},
		{
			name:        "retry exceeds limit",
			ctx:         context.Background(),
			fn:          func(context.Context) (bool, error) { return false, nil },
			wantErr:     true,
			expectedErr: ErrNotSetDeadline,
		},
		{
			name:        "return custom error",
			ctx:         timeoutCtx,
			fn:          func(context.Context) (bool, error) { return false, errors.New("custom error") },
			wantErr:     true,
			expectedErr: errors.New("custom error"),
		},
		{
			name:        "function returns error",
			ctx:         context.Background(),
			fn:          func(context.Context) (bool, error) { return false, errors.New("custom error") },
			wantErr:     true,
			expectedErr: ErrNotSetDeadline,
		},
		{
			name:        "context deadline exceeded",
			ctx:         timeoutCtx,
			fn:          func(context.Context) (bool, error) { time.Sleep(5 * time.Millisecond); return false, nil },
			wantErr:     true,
			expectedErr: ErrTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := RetryWithExponentialBackoffUntilTimeout(tt.ctx, tt.fn)

			if (err != nil) != tt.wantErr {
				t.Errorf("RetryWithExponentialBackoffUntilTimeout() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && err.Error() != tt.expectedErr.Error() {
				t.Errorf("RetryWithExponentialBackoffUntilTimeout() got = %v, want %v", err, tt.expectedErr)
			}
		})
	}
}

func TestBackoffWait(t *testing.T) {
	tests := []struct {
		name    string
		backoff BackoffWait
	}{
		{
			name:    "total runs less than 1",
			backoff: BackoffWait{TotalRuns: 0, BaseDuration: time.Second, Factor: 2.0, JitterFactor: 0.0},
		},
		{
			name:    "total runs less than 1 with jitter",
			backoff: BackoffWait{TotalRuns: 0, BaseDuration: time.Second, Factor: 2.0, JitterFactor: 0.1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.backoff.wait()
			if tt.name == "total runs less than 1" && got != tt.backoff.BaseDuration {
				t.Errorf("BackoffWait.wait() = %v, want %v", got, tt.backoff.BaseDuration)
			}
			if tt.name == "total runs less than 1 with jitter" && (got <= tt.backoff.BaseDuration || got > tt.backoff.BaseDuration+time.Duration(float64(tt.backoff.BaseDuration)*tt.backoff.JitterFactor)) {
				t.Errorf("BackoffWait.wait() with jitter = %v, want in range (%v, %v)", got, tt.backoff.BaseDuration, tt.backoff.BaseDuration+time.Duration(float64(tt.backoff.BaseDuration)*tt.backoff.JitterFactor))
			}
		})
	}
}

func TestAddJitter(t *testing.T) {
	type args struct {
		duration time.Duration
		factor   float64
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "BaseDuration 1s with 0.1 Jitter", args: args{duration: time.Second, factor: 0.1}},
		{name: "BaseDuration 2s with 0.2 Jitter", args: args{duration: 2 * time.Second, factor: 0.2}},
		{name: "BaseDuration 1s with 0.5 Jitter", args: args{duration: time.Second, factor: 0.5}},
		{name: "BaseDuration 1s with 0 Jitter", args: args{duration: time.Second, factor: 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := addJitter(tt.args.duration, tt.args.factor)
			if Float64Equal(tt.args.factor, 0.0) {
				tt.args.factor = 1.0
			}

			maxJitter := time.Duration(tt.args.factor * float64(tt.args.duration))
			minDuration := tt.args.duration - maxJitter
			maxDuration := tt.args.duration + maxJitter
			if got < minDuration || got > maxDuration {
				t.Error(got, minDuration, maxDuration, maxJitter)
				t.Errorf("addJitter() = %v, want in range (%v, %v)", got, minDuration, maxDuration)
			}
		})
	}
}

func TestRetryableFuncWithContext(t *testing.T) {
	tests := []struct {
		name       string
		fn         RetryableFunc
		wantErr    bool
		wantResult bool
	}{
		{
			name:       "function succeeds",
			fn:         func() (bool, error) { return true, nil },
			wantErr:    false,
			wantResult: true,
		},
		{
			name:       "function fails",
			fn:         func() (bool, error) { return false, errors.New("custom error") },
			wantErr:    true,
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fnWithContext := tt.fn.WithContext()

			done, err := fnWithContext(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("RetryableFuncWithContext() error = %v, wantErr %v", err, tt.wantErr)
			}

			if done != tt.wantResult {
				t.Errorf("RetryableFuncWithContext() result = %v, wantResult %v", done, tt.wantResult)
			}
		})
	}
}

func TestExponentialBackoffWithCtx(t *testing.T) {
	var ErrCustom = errors.New("custom error")

	tests := []struct {
		name    string
		backoff BackoffWait
		fn      RetryableFuncWithContext
		wantErr error
	}{
		{
			name:    "function returns true, no error",
			backoff: BackoffWait{TotalRuns: 2, BaseDuration: time.Millisecond, Factor: 1.5, JitterFactor: 0.5},
			fn:      func(context.Context) (bool, error) { return true, nil },
			wantErr: nil,
		},
		{
			name:    "function returns false, no error",
			backoff: BackoffWait{TotalRuns: 2, BaseDuration: time.Millisecond, Factor: 1.5, JitterFactor: 0.5},
			fn:      func(context.Context) (bool, error) { return false, nil },
			wantErr: ErrTimeout,
		},
		{
			name:    "function returns error",
			backoff: BackoffWait{TotalRuns: 2, BaseDuration: time.Millisecond, Factor: 1.5, JitterFactor: 0.5},
			fn:      func(context.Context) (bool, error) { return false, ErrCustom },
			wantErr: ErrCustom,
		},
		{
			name:    "context deadline exceeded",
			backoff: BackoffWait{TotalRuns: 2, BaseDuration: time.Second, Factor: 1.5, JitterFactor: 0.5},
			fn:      func(context.Context) (bool, error) { time.Sleep(1 * time.Millisecond); return false, nil },
			wantErr: ErrTimeout,
		},
		{
			name:    "exceed_total_runs",
			backoff: BackoffWait{TotalRuns: 1, BaseDuration: 1 * time.Millisecond, Factor: 1, JitterFactor: 0},
			fn:      func(context.Context) (bool, error) { return false, nil },
			wantErr: ErrTimeout,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			defer cancel()

			if err := exponentialBackoffWithCtx(ctx, tt.backoff, tt.fn); !errors.Is(err, tt.wantErr) {
				t.Errorf("exponentialBackoffWithCtx() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
