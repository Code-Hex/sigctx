package sigctx

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"time"
)

var Canceled = errors.New("signal recieved")

type signalCtx struct {
	context.Context

	recv chan struct{} // closed by coming signals.

	mu      sync.Mutex
	err     error
	errOnce sync.Once
	signal  os.Signal
	sigchan chan os.Signal
}

func WithSignals(parent context.Context, sig ...os.Signal) context.Context {
	sigctx := newSignalCtx(parent)
	signal.Notify(sigctx.sigchan, sig...)
	go func() {
		for {
			select {
			case <-sigctx.Done():
				return
			case sigctx.signal = <-sigctx.sigchan:
				if sigctx.recv != nil {
					sigctx.mu.Lock()
					close(sigctx.recv) // Notify that signals has been recieved
					sigctx.recv = make(chan struct{})
					sigctx.mu.Unlock()
				}
			}
		}
	}()
	return sigctx
}

func newSignalCtx(parent context.Context) *signalCtx {
	return &signalCtx{
		Context: parent,
		sigchan: make(chan os.Signal, 1),
		recv:    make(chan struct{}),
	}
}

func WithCancelSignals(parent context.Context, sig ...os.Signal) context.Context {
	cctx, cancel := context.WithCancel(parent)

	sigctx := newSignalCtx(cctx)
	signal.Notify(sigctx.sigchan, sig...)
	go func() {
		for {
			select {
			case <-sigctx.Done():
				return
			case sigctx.signal = <-sigctx.sigchan:
				sigctx.errOnce.Do(func() {
					cancel()
					sigctx.err = Canceled
				})
			}
		}
	}()
	return sigctx
}

func Signal(ctx context.Context) (os.Signal, error) {
	if sigctx, ok := ctx.(*signalCtx); ok {
		return sigctx.signal, nil
	}
	return nil, errors.New("Context is not sigctx type")
}

func Recv(ctx context.Context) <-chan struct{} {
	if sigctx, ok := ctx.(*signalCtx); ok {
		sigctx.mu.Lock()
		r := sigctx.recv
		sigctx.mu.Unlock()

		return r
	}
	return nil
}

func (sigctx *signalCtx) Deadline() (deadline time.Time, ok bool) {
	return sigctx.Context.Deadline()
}

func (sigctx *signalCtx) Done() <-chan struct{} {
	return sigctx.Context.Done()
}

func (sigctx *signalCtx) Err() error {
	if sigctx.err != nil {
		return sigctx.err
	}
	return sigctx.Context.Err()
}

func (sigctx *signalCtx) Value(key interface{}) interface{} {
	return sigctx.Context.Value(key)
}
