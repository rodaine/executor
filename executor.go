package executor

import "context"

// Interface describes an executor that performs a set of Actions. It is up to
// the implementing type to define the concurrency and open/closed failure
// behavior of the actions.
type Interface interface {
	// Execute performs all provided actions by calling their Execute method.
	// This method should make a best-effort to cancel outstanding actions if the
	// provided ctx is cancelled.
	Execute(ctx context.Context, actions ...Action) error
}

// An Action performs a single arbitrary task.
type Action interface {
	// Execute performs the work of an Action. This method should make a best
	// effort to be cancelled if the provided ctx is cancelled.
	Execute(ctx context.Context) error
}

// ActionFunc permits using a standalone function as an Action.
type ActionFunc func(context.Context) error

// Execute satisfies the Action interface, delegating the call to the
// underlying function.
func (fn ActionFunc) Execute(ctx context.Context) error { return fn(ctx) }

var _ Action = ActionFunc(nil)
