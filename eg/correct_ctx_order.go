package main

import (
	"context"
	"fmt"
	"sync"
	"syscall"

	"github.com/Code-Hex/sigctx"
)

// This is a simple case using `sigctx.WithSignals`
// and `sigctx.WithCancelSignals` together.
// It is very simple compared with mistake_ctx_order.go.
// So You should run `sigctx.WithCancelSignals` first and
// then `sigctx.WithSignals`.

func main() {
	// First, Create a context to cancel all goroutines.
	cancelctx := sigctx.WithCancelSignals(context.Background(), syscall.SIGHUP)
	// Second, Wrap the cancel signal context to notify all goroutines.
	signalctx := sigctx.WithSignals(cancelctx, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				// You do not have to think anything!!
				// Let's use the `signalctx` for everything.
				case <-sigctx.Recv(signalctx):
					signal, err := sigctx.Signal(signalctx)
					if err != nil {
						fmt.Println(err.Error())
					} else {
						fmt.Println(signal.String())
					}
				case <-signalctx.Done():
					fmt.Println(signalctx.Err())
					return
				}
			}
		}()
	}
	wg.Wait()
}
