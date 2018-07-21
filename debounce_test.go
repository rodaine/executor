package executor

import (
	"context"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDebounce(t *testing.T) {
	// cannot be parallel due to the following:
	prev := runtime.NumCPU()
	runtime.GOMAXPROCS(8)
	defer runtime.GOMAXPROCS(prev)

	exec := Debounce(Parallel{})

	var ct uint32

	addToCt := func(ctx context.Context) error {
		time.Sleep(time.Millisecond) // keep the action in-flight
		atomic.AddUint32(&ct, 1)
		return nil
	}

	noop := func(ctx context.Context) error { return nil }

	err := exec.Execute(context.Background(),
		Named("add", "+1", addToCt),      //  named, +1
		Named("add", "+1", addToCt),      //  will be debounced
		Named("add", "add one", addToCt), // different name, +1

		ActionFunc(addToCt), // unnamed, +1
		ActionFunc(noop),    // unrelated
	)

	assert.NoError(t, err)
	assert.Equal(t, uint32(3), ct)
}
