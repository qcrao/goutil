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
