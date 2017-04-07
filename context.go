// MIT License
//
// Copyright (c) 2016 K
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// Package sigctx were developed to easily bind signals and context.Context.
// For example, When receive a signal, you can notify received signals to all
// goroutines using context. Or when you receive a signal, you can invoke a
// context cancel.
package sigctx

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"time"
)

// Canceled is the error returned by Context.Err when the context is canceled.
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

// WithSignals returns a copy of parent which the value can notify signals by
// Recv(ctx).
//
// When the process receive signals passed to the argument, it sends the
// information of the received signal to all the goroutines.
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

// newSignalCtx returns an initialized signalCtx.
func newSignalCtx(parent context.Context) *signalCtx {
	return &signalCtx{
		Context: parent,
		sigchan: make(chan os.Signal, 1),
	}
}

// WithCancelSignals returns a copy of parent which the value can context cancel by
// signals.
//
// Context cancellation is executed after receiving the signal.
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
					sigctx.err = Canceled
					cancel()
				})
			}
		}
	}()
	return sigctx
}

// Signal returns os.Signal and error.
//
// If the context passed to the argument is a signalCtx context, it returns
// the received os.Signal.
// Otherwise it returns nil and error message.
func Signal(ctx context.Context) (os.Signal, error) {
	if sigctx, ok := ctx.(*signalCtx); ok {
		sigctx.mu.Lock()
		defer sigctx.mu.Unlock()
		return sigctx.signal, nil
	}
	return nil, errors.New("Context is not sigctx type")
}

// Recv send channel when a process receives a signal.
//
// You can use Recv(ctx) only if you pass a signalCtx context as an argument.
// If you want to use Recv(ctx) please wrap context.Context so that the
// signalCtx context is the last.
func Recv(ctx context.Context) <-chan struct{} {
	if sigctx, ok := ctx.(*signalCtx); ok {
		sigctx.mu.Lock()
		if sigctx.recv == nil {
			sigctx.recv = make(chan struct{})
		}
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
		sigctx.mu.Lock()
		defer sigctx.mu.Unlock()
		return sigctx.err
	}
	return sigctx.Context.Err()
}

func (sigctx *signalCtx) Value(key interface{}) interface{} {
	return sigctx.Context.Value(key)
}
