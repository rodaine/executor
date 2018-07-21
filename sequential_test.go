package executor

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSequential(t *testing.T) {
	t.Parallel()

	seq := Sequential{}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		buf := new(bytes.Buffer)

		n := 10
		actions := make([]Action, n)
		for i := 0; i < n; i++ {
			x := i
			actions[i] = ActionFunc(func(ctx context.Context) error {
				fmt.Fprint(buf, x)
				return nil
			})
		}

		err := seq.Execute(context.Background(), actions...)
		assert.NoError(t, err)
		assert.Equal(t, "0123456789", buf.String())
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()

		ct := 0

		addToCt := ActionFunc(func(ctx context.Context) error {
			ct++
			return nil
		})

		actions := []Action{
			addToCt,
			ActionFunc(func(ctx context.Context) error { return errors.New("some error") }),
			addToCt,
		}

		err := seq.Execute(context.Background(), actions...)
		assert.Error(t, err)
		assert.Equal(t, 1, ct)
	})

	t.Run("cancelled", func(t *testing.T) {
		t.Parallel()

		ct := 0

		addToCt := ActionFunc(func(ctx context.Context) error {
			ct++
			return nil
		})

		actions := []Action{addToCt, addToCt, addToCt}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := seq.Execute(ctx, actions...)
		assert.Equal(t, context.Canceled, err)
		assert.Zero(t, ct)
	})
}
