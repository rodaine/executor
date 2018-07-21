package executor

import "context"

// Sequential implements the Executor Interface, performing each Action in series.
type Sequential struct{}

// Execute performs each action in order, exiting on the first error or if the
// context is cancelled/deadlined.
func (Sequential) Execute(ctx context.Context, actions ...Action) error {
	for _, a := range actions {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := a.Execute(ctx); err != nil {
				return err
			}
		}
	}

	return nil
}

var _ Interface = Sequential{}
