package executor

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMetrics(t *testing.T) {
	t.Parallel()

	noop := ActionFunc(func(ctx context.Context) error { return nil })

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ss := new(fakeStatSource)
		ex := Metrics(Parallel{}, ss)

		err := ex.Execute(context.Background(), noop, Named("foo", "123", noop))

		assert.NoError(t, err)

		ss.testCounter(t, "all_actions.success", 2)
		ss.testCounter(t, "all_actions.error", 0)
		ss.testTimer(t, "all_actions", 2)

		ss.testCounter(t, "foo.success", 1)
		ss.testCounter(t, "foo.error", 0)
		ss.testTimer(t, "foo", 1)
	})

	t.Run("error", func(t *testing.T) {
		t.Parallel()

		ss := new(fakeStatSource)
		ex := Metrics(Sequential{}, ss)

		expected := errors.New("some error")
		errFn := Named("bar", "456", func(context.Context) error { return expected })

		err := ex.Execute(context.Background(), noop, errFn)
		assert.Equal(t, expected, err)

		ss.testCounter(t, "all_actions.success", 1)
		ss.testCounter(t, "all_actions.error", 1)
		ss.testTimer(t, "all_actions", 2)

		ss.testCounter(t, "bar.success", 0)
		ss.testCounter(t, "bar.error", 1)
		ss.testTimer(t, "bar", 1)
	})
}

type fakeStatSource struct {
	timers   sync.Map
	counters sync.Map
}

func (s *fakeStatSource) Timer(name string) Timer {
	c := make(chan time.Duration, 10)
	s.timers.Store(name, c)
	return func(d time.Duration) { c <- d }
}

func (s *fakeStatSource) Counter(name string) Counter {
	val := new(int64)
	s.counters.Store(name, val)
	return func(d int) { atomic.AddInt64(val, int64(d)) }
}

func (s *fakeStatSource) testCounter(t *testing.T, name string, expected int) bool {
	ct, ok := s.counters.Load(name)
	return assert.True(t, ok) &&
		assert.Equal(t, int64(expected), *(ct.(*int64)))
}

func (s *fakeStatSource) testTimer(t *testing.T, name string, expected int) bool {
	ch, ok := s.timers.Load(name)
	return assert.True(t, ok) &&
		assert.Len(t, ch.(chan time.Duration), expected)
}
