package client

import (
	"context"
	"time"
)

type RetryOptions struct {
	Attempts    int
	Delay       time.Duration
	Backoff     func(attempt int) time.Duration
	ShouldRetry func(err error) bool
}

func WithRetry(ctx context.Context, opts RetryOptions, fn func(context.Context) error) error {
	var lastErr error

	for i := 0; i < opts.Attempts; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		err := fn(ctx)
		if err == nil {
			return nil
		}

		lastErr = err

		if opts.ShouldRetry != nil && !opts.ShouldRetry(err) {
			return err
		}

		if i < opts.Attempts-1 {
			time.Sleep(opts.Backoff(i))
		}
	}

	return lastErr
}
