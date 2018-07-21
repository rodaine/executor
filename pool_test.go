package executor

import (
	"context"
	"errors"
	"runtime"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPool(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		exec, done := Pool(0)
		defer done()

		var ct uint32

		addToCt := ActionFunc(func(ctx context.Context) error {
			atomic.AddUint32(&ct, 1)
			return nil
		})

		n := runtime.NumCPU() * 10
		actions := make([]Action, n)
		for i := 0; i < n; i++ {
			actions[i] = addToCt
		}

		err := exec.Execute(context.Background(), actions...)
		assert.NoError(t, err)
		assert.Equal(t, uint32(n), ct)
	})

	t.Run("empty actions", func(t *testing.T) {
		t.Parallel()

		exec, done := Pool(0)
		defer done()

		err := exec.Execute(context.Background())
		assert.NoError(t, err)
	})

	t.Run("action error", func(t *testing.T) {
		t.Parallel()

		exec, done := Pool(0)
		defer done()

		noopAct := ActionFunc(func(ctx context.Context) error { return nil })

		waitAct := ActionFunc(func(ctx context.Context) error {
			<-ctx.Done()
			return nil
		})

		errAct := ActionFunc(func(ctx context.Context) error { return errors.New("some error") })

		err := exec.Execute(context.Background(), noopAct, waitAct, errAct)
		assert.Error(t, err)
	})

	t.Run("context cancelled", func(t *testing.T) {
		t.Parallel()

		exec, done := Pool(0)
		defer done()

		var ct uint32

		addToCt := ActionFunc(func(ctx context.Context) error {
			atomic.AddUint32(&ct, 1)
			return nil
		})

		n := runtime.NumCPU() * 10
		actions := make([]Action, n)
		for i := 0; i < n; i++ {
			actions[i] = addToCt
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := exec.Execute(ctx, actions...)
		assert.Equal(t, context.Canceled, err)
	})

	t.Run("done pool", func(t *testing.T) {
		t.Parallel()

		exec, done := Pool(0)
		done()

		var ct uint32

		addToCt := ActionFunc(func(ctx context.Context) error {
			atomic.AddUint32(&ct, 1)
			return nil
		})

		n := runtime.NumCPU() * 10
		actions := make([]Action, n)
		for i := 0; i < n; i++ {
			actions[i] = addToCt
		}

		err := exec.Execute(context.Background(), actions...)
		assert.Error(t, err)
	})
}
