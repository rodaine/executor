package executor

import (
	"sync"
)

// A StatSet is the cached value.
type statSet struct {
	// Latency measures how long an Action takes
	Latency Timer
	// Success is incremented when an Action does not return an error
	Success Counter
	// Error is incremented when an Action results in an error
	Error Counter
}

// newStatSet creates a statSet from the given src with the provided name.
func newStatSet(src StatSource, name string) *statSet {
	return &statSet{
		Latency: src.Timer(name),
		Success: src.Counter(name + ".success"),
		Error:   src.Counter(name + ".error"),
	}
}

// Cache describes a read-through cache to obtain
type statCache interface {
	// get returns a shared statSet for the given name, either from the cache or
	// a provided StatSource.
	get(name string) *statSet
}

// mutexCache implements statCache, backed by a map and sync.RWMutex
type mutexCache struct {
	src    StatSource
	mtx    sync.RWMutex
	lookup map[string]*statSet
}

func newMutexCache(src StatSource) statCache {
	return &mutexCache{
		src:    src,
		lookup: make(map[string]*statSet),
	}
}

func (mc *mutexCache) get(name string) *statSet {
	// take a read lock to see if the set already exists
	mc.mtx.RLock()
	set, ok := mc.lookup[name]
	mc.mtx.RUnlock()

	if ok { // the set exists, return it
		return set
	}

	// need to take a write lock to update the map
	mc.mtx.Lock()
	// While waiting for the write lock, another goroutine may have created the
	// set. Here, we check again after obtaining the lock before making a new one
	if set, ok = mc.lookup[name]; !ok {
		set = newStatSet(mc.src, name)
		mc.lookup[name] = set
	}
	mc.mtx.Unlock()

	return set
}

type syncMapCache struct {
	src    StatSource
	lookup sync.Map
}

func newSyncMapCache(src StatSource) statCache {
	return &syncMapCache{src: src}
}

func (smc *syncMapCache) get(name string) *statSet {
	val, _ := smc.lookup.Load(name)
	if set, ok := val.(*statSet); ok {
		return set
	}

	// create a new statSet, but don't store it if one was added since the last
	// load. This is not ideal since we can't atomically create the set and
	// write it.
	set, _ := smc.lookup.LoadOrStore(name, newStatSet(smc.src, name))
	return set.(*statSet)
}

var (
	_ statCache = (*mutexCache)(nil)
	_ statCache = (*syncMapCache)(nil)
)
