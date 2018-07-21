package executor

import (
	"context"

	"golang.org/x/sync/errgroup"
)

// Parallel is a concurrent implementation of the Executor Interface.
type Parallel struct{}

// Execute performs all provided actions concurrently, failing closed on the
// first error or if ctx is cancelled.
func (p Parallel) Execute(ctx context.Context, actions ...Action) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	grp, ctx := errgroup.WithContext(ctx)

	for _, a := range actions {
		grp.Go(parallelFunc(ctx, a))
	}

	return grp.Wait()
}

// parallelFunc binds the Context and Action to the proper function signature for an
// errgroup.Group.
func parallelFunc(ctx context.Context, a Action) func() error {
	return func() error { return a.Execute(ctx) }
}

var _ Interface = Parallel{}
