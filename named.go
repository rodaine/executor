package executor

// A NamedAction describes an Action that also has a unique identifier. This
// interface is used by various Executors such as Debounce and Metrics to
// categorize Actions.
type NamedAction interface {
	Action

	// Type returns the general category of this Action. Type is not expected to
	// be unique but shared between unique actions with the same behavior.
	Type() string

	// ID returns the unique name for this Action. Identical actions should
	// return the same ID value.
	ID() string
}

type namedAction struct {
	ActionFunc
	typ, id string
}

func (a namedAction) Type() string { return a.typ }

func (a namedAction) ID() string { return a.id }

// Named creates a NamedAction, with the specified Type and ID. This function is
// just a helper to simplify creating NamedActions.
func Named(typ, id string, a ActionFunc) NamedAction {
	return namedAction{
		ActionFunc: a,
		typ:        typ,
		id:         id,
	}
}

var _ NamedAction = namedAction{}
