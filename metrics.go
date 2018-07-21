package executor

import (
	"context"
	"time"
)

// StatSource creates metrics with the given name. The returned metrics must be
// concurrency-safe.
type StatSource interface {
	Timer(name string) Timer
	Counter(name string) Counter
}

// Timer emits the duration of a particular event. The duration value is
// typically used to measure latencies and create histograms thereof.
type Timer func(duration time.Duration)

// Counter emits any number of events happening at a given time. For example,
// Counters are often used to measure RPS.
type Counter func(delta int)

type metrics struct {
	ex Interface
	statCache
}

// Metrics decorates the passed in executor and emits stats for all Actions
// executed, capturing success/failure counters as well as a latency timer for
// the each Action. If a NamedAction is passed in, per Action Type stats are
// emitted as well.
func Metrics(e Interface, src StatSource) Interface {
	return &metrics{
		ex:        e,
		statCache: newMutexCache(src),
	}
}

func (m *metrics) Execute(ctx context.Context, actions ...Action) error {
	wrapped := make([]Action, len(actions))
	global := m.get("all_actions")

	for i, a := range actions {
		if na, ok := a.(NamedAction); ok {
			wrapped[i] = namedStatAction{
				NamedAction: na,
				global:      global,
				stats:       m.get(na.Type()),
			}
		} else {
			wrapped[i] = statAction{
				Action: a,
				global: global,
			}
		}
	}

	return m.ex.Execute(ctx, wrapped...)
}

type namedStatAction struct {
	NamedAction
	global *statSet
	stats  *statSet
}

func (a namedStatAction) Execute(ctx context.Context) error {
	return captureMetrics(ctx, a.NamedAction, a.global, a.stats)
}

type statAction struct {
	Action
	global *statSet
}

func (a statAction) Execute(ctx context.Context) error {
	return captureMetrics(ctx, a.Action, a.global, nil)
}

func captureMetrics(ctx context.Context, a Action, global, stats *statSet) error {
	// execute the action, timing its latency
	start := time.Now()
	err := a.Execute(ctx)
	lat := time.Since(start)

	// create our counter values for error/success
	var errored, succeeded int
	if err != nil {
		errored = 1
	} else {
		succeeded = 1
	}

	// emit the global stats
	global.Latency(lat)
	global.Success(succeeded)
	global.Error(errored)

	// if there are name-scoped stats, emit those, too
	if stats != nil {
		stats.Latency(lat)
		stats.Success(succeeded)
		stats.Error(errored)
	}

	return err
}
