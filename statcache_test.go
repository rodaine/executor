package executor

import (
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func testCache(t *testing.T, cache statCache) {
	wg := &sync.WaitGroup{}
	wg.Add(runtime.NumCPU())

	for i := 1; i < runtime.NumCPU(); i++ {
		go func() {
			assert.NotNil(t, cache.get("foo"))
			wg.Done()
		}()
	}

	go func() {
		assert.NotNil(t, cache.get("bar"))
		wg.Done()
	}()

	wg.Wait()
}

func TestMutexCache(t *testing.T) {
	t.Parallel()
	cache := newMutexCache(stubSource{}).(*mutexCache)

	testCache(t, cache)

	assert.Len(t, cache.lookup, 2)
	assert.Contains(t, cache.lookup, "foo")
	assert.Contains(t, cache.lookup, "bar")
}

func TestSyncMapCache(t *testing.T) {
	t.Parallel()
	cache := newSyncMapCache(stubSource{}).(*syncMapCache)

	testCache(t, cache)

	ct := 0
	var seenFoo, seenBar bool
	cache.lookup.Range(func(k, v interface{}) bool {
		ct++
		seenFoo = seenFoo || k.(string) == "foo"
		seenBar = seenBar || k.(string) == "bar"
		return true
	})

	assert.Equal(t, 2, ct)
	assert.True(t, seenFoo)
	assert.True(t, seenBar)
}

type stubSource struct{}

func (stubSource) Timer(name string) Timer {
	return func(time.Duration) {}
}

func (stubSource) Counter(name string) Counter {
	return func(int) {}
}
