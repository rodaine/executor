package executor

import (
	"context"

	"golang.org/x/sync/singleflight"
)

// Debounce wraps an Executor Interface, preventing duplicate NamedActions
// from running concurrently, even from concurrent calls to Execute.
func Debounce(e Interface) Interface {
	return debouncer{
		ex: e,
		sf: new(singleflight.Group),
	}
}

type debouncer struct {
	ex Interface
	sf *singleflight.Group
}

func (d debouncer) Execute(ctx context.Context, actions ...Action) error {
	wrapped := make([]Action, len(actions))

	for i, a := range actions {
		if na, ok := a.(NamedAction); ok {
			wrapped[i] = debouncedAction{
				NamedAction: na,
				sf:          d.sf,
			}
		} else {
			wrapped[i] = actions[i]
		}
	}

	return d.ex.Execute(ctx, wrapped...)
}

type debouncedAction struct {
	NamedAction
	sf *singleflight.Group
}

func (da debouncedAction) Execute(ctx context.Context) error {
	// all actions only return an error, so we don't care if the value is shared
	// or not.
	_, err, _ := da.sf.Do(da.Type()+da.ID(), func() (interface{}, error) {
		return nil, da.NamedAction.Execute(ctx)
	})

	return err
}

var _ Interface = debouncer{}
