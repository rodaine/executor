package executor

import (
	"context"
	"errors"
	"runtime"
)

// A CloseFunc is returned by Pool to release resources held by a Pool. The
// function should be called only once; subsequent calls may result in a
// panic.
type CloseFunc func()

type pool struct {
	done chan struct{}
	in   chan poolAction
}

// Pool creates an Executor Interface instance backed by a concurrent worker
// pool. Up to n Actions can be in-flight simultaneously; if n is less than or
// equal to zero, runtime.NumCPU is used. The returned CloseFunc must be called
// to release resources held by the pool.
func Pool(n int) (Interface, CloseFunc) {
	if n <= 0 {
		n = runtime.NumCPU()
	}

	p := pool{done: make(chan struct{}), in: make(chan poolAction, n)}

	for i := 0; i < n; i++ {
		go p.work(p.in, p.done)
	}

	return p, func() { close(p.done) }
}

// Execute enqueues all Actions on the worker pool, failing closed on the
// first error or if ctx is cancelled. This method blocks until all enqueued
// Actions have returned. In the event of an error, not all Actions may be
// executed.
func (p pool) Execute(ctx context.Context, actions ...Action) error {
	qty := len(actions)
	if qty == 0 {
		return nil
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	res := make(chan error, qty)

	var err error
	var queued uint64

enqueue:
	for _, action := range actions {
		pa := poolAction{ctx: ctx, act: action, res: res}
		select {
		case <-p.done: // pool is closed
			cancel()
			return errors.New("pool is closed")
		case <-ctx.Done(): // ctx is closed by caller
			err = ctx.Err()
			break enqueue
		case p.in <- pa: // enqueue action
			queued++
		}
	}

	for ; queued > 0; queued-- {
		if r := <-res; r != nil {
			if err == nil {
				err = r
				cancel()
			}
		}
	}

	return err
}

func (p pool) work(in <-chan poolAction, done <-chan struct{}) {
	for {
		select {
		case <-done:
			return
		case a := <-in:
			a.res <- a.act.Execute(a.ctx)
		}
	}
}

type poolAction struct {
	ctx context.Context
	act Action
	res chan<- error
}

var _ Interface = pool{}
