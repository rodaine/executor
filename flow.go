package executor

import (
	"context"
	"fmt"

	"golang.org/x/sync/semaphore"
)

type flow struct {
	maxActions int64
	actions    *semaphore.Weighted
	calls      *semaphore.Weighted
	ex         Interface
}

// ControlFlow decorates an Executor, limiting it to a maximum concurrent
// number of calls and actions.
func ControlFlow(e Interface, maxCalls, maxActions int64) Interface {
	return flow{
		ex:         e,
		maxActions: maxActions,
		calls:      semaphore.NewWeighted(maxCalls),
		actions:    semaphore.NewWeighted(maxActions),
	}
}

// Execute attempts to acquire the semaphores for the concurrent calls and
// actions before delegating to the decorated Executor. If Execute is called
// with more actions than maxActions, an error is returned.
func (f flow) Execute(ctx context.Context, actions ...Action) error {
	qty := int64(len(actions))

	if qty > f.maxActions {
		return fmt.Errorf("maximum %d actions allowed", f.maxActions)
	}

	if err := f.calls.Acquire(ctx, 1); err != nil {
		return err
	}
	defer f.calls.Release(1)

	if err := f.actions.Acquire(ctx, qty); err != nil {
		return err
	}
	defer f.actions.Release(qty)

	return f.ex.Execute(ctx, actions...)
}

var _ Interface = flow{}
