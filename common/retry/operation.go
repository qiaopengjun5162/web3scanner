// Package retry provides utilities for retrying operations with customizable strategies.
package retry

import (
	"context"
	"fmt"
	"time"
)

type ErrFailedPermanently struct {
	attempts int
	LastErr  error
}

func (e *ErrFailedPermanently) Error() string {
	return fmt.Sprintf("operation failed permanently after %d attempts: %v", e.attempts, e.LastErr)
}

func (e *ErrFailedPermanently) Unwrap() error {
	return e.LastErr
}

type pair[T, U any] struct {
	a T
	b U
}

// Do2 retries an operation that returns two values and an error, using the given strategy.
// It continues until the operation succeeds or the maximum number of attempts is reached.
//
// Parameters:
//   - ctx: A context.Context for cancellation and timeout control.
//   - maxAttempts: The maximum number of times to attempt the operation.
//   - strategy: The retry strategy to use between attempts.
//   - op: The operation function to be retried. It should return two values of types T and U, and an error.
//
// Returns:
//   - T: The first return value of the operation if successful.
//   - U: The second return value of the operation if successful.
//   - error: An error if the operation failed permanently, or nil if successful.
func Do2[T, U any](ctx context.Context, maxAttempts int, strategy Strategy, op func() (T, U, error)) (T, U, error) {
	f := func() (pair[T, U], error) {
		a, b, err := op()
		return pair[T, U]{a, b}, err
	}
	res, err := Do(ctx, maxAttempts, strategy, f)
	return res.a, res.b, err
}

// Do retry an operation that returns a value and an error, using the given strategy.
// It continues until the operation succeeds or the maximum number of attempts is reached.
//
// Parameters:
//   - ctx: A context.Context for cancellation and timeout control.
//   - maxAttempts: The maximum number of times to attempt the operation.
//   - strategy: The retry strategy to use between attempts.
//   - op: The operation function to be retried. It should return a value of type T and an error.
//
// Returns:
//   - T: The return value of the operation if successful.
//   - error: An error if the operation failed permanently, or nil if successful.
func Do[T any](ctx context.Context, maxAttempts int, strategy Strategy, op func() (T, error)) (T, error) {
	var empty, ret T
	var err error
	if maxAttempts < 1 {
		return empty, fmt.Errorf("need at least 1 attempt to run op, but have %d max attempts", maxAttempts)
	}

	for i := 0; i < maxAttempts; i++ {
		if ctx.Err() != nil {
			return empty, ctx.Err()
		}
		ret, err = op()
		if err == nil {
			return ret, nil
		}
		if i != maxAttempts-1 {
			time.Sleep(strategy.Duration(i))
		}
	}
	return empty, &ErrFailedPermanently{
		attempts: maxAttempts,
		LastErr:  err,
	}
}
