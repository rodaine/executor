package executor

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParallel(t *testing.T) {
	t.Parallel()

	exec := Parallel{}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		var ct uint32

		addToCt := ActionFunc(func(ctx context.Context) error {
			atomic.AddUint32(&ct, 1)
			return nil
		})

		n := 10
		actions := make([]Action, n)
		for i := 0; i < n; i++ {
			actions[i] = addToCt
		}

		err := exec.Execute(context.Background(), actions...)
		assert.NoError(t, err)
		assert.Equal(t, uint32(n), ct)
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()

		var ct uint32

		addToCt := ActionFunc(func(ctx context.Context) error {
			atomic.AddUint32(&ct, 1)
			return nil
		})

		waitForCancel := ActionFunc(func(ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		})

		retErr := ActionFunc(func(ctx context.Context) error {
			return errors.New("some error")
		})

		err := exec.Execute(context.Background(),
			addToCt,
			waitForCancel,
			retErr)

		assert.Error(t, err)
		assert.True(t, 0 == ct || 1 == ct)
	})

	t.Run("cancelled", func(t *testing.T) {
		t.Parallel()

		var ct uint32

		addToCt := ActionFunc(func(ctx context.Context) error {
			atomic.AddUint32(&ct, 1)
			return nil
		})

		n := 10
		actions := make([]Action, n)
		for i := 0; i < n; i++ {
			actions[i] = addToCt
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := exec.Execute(ctx, actions...)
		assert.Equal(t, context.Canceled, err)
		assert.Zero(t, ct)
	})
}
