package executor

import (
	"context"
	"math"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestControlFlow(t *testing.T) {
	t.Parallel()

	parallel := Parallel{}

	t.Run("exceed max", func(t *testing.T) {
		t.Parallel()

		exec := ControlFlow(parallel, int64(runtime.NumCPU()), 2)

		noopAct := ActionFunc(func(ctx context.Context) error { return nil })

		err := exec.Execute(context.Background(), noopAct, noopAct, noopAct)
		assert.Error(t, err)
	})

	t.Run("deadline on calls", func(t *testing.T) {
		t.Parallel()

		exec := ControlFlow(parallel, 1, math.MaxInt64)

		go exec.Execute(context.Background(), ActionFunc(func(ctx context.Context) error {
			time.Sleep(time.Second)
			return nil
		}))

		time.Sleep(time.Millisecond)

		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()

		err := exec.Execute(ctx, nil)
		assert.Equal(t, context.DeadlineExceeded, err)
	})

	t.Run("deadline on actions", func(t *testing.T) {
		t.Parallel()

		exec := ControlFlow(parallel, math.MaxInt64, 1)

		go exec.Execute(context.Background(), ActionFunc(func(ctx context.Context) error {
			time.Sleep(time.Second)
			return nil
		}))

		time.Sleep(time.Millisecond)

		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
		defer cancel()

		err := exec.Execute(ctx, ActionFunc(func(ctx context.Context) error { return nil }))
		assert.Equal(t, context.DeadlineExceeded, err)
	})
}
