package main

import (
	"context"
	"fmt"
	"sync"
	"syscall"

	"github.com/Code-Hex/sigctx"
)

// This is a complicated case using `sigctx.WithSignals`
// and `sigctx.WithCancelSignals` together.
// If you mistake the order of using `sigctx.WithSignals` and
// `sigctx.WithCancelSignals`, You will suffer from using context.

func main() {
	// First, Create a context to notify all goroutines.
	signalctx := sigctx.WithSignals(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	// Second, Wrap the signal context to cancel all goroutines.
	cancelctx := sigctx.WithCancelSignals(signalctx, syscall.SIGHUP)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				// Since it is the context to receive,
				// You should pass the signal context as an argument.
				case <-sigctx.Recv(signalctx):
					// `sigctx.Signal(signalctx)` is also the same reason.
					signal, err := sigctx.Signal(signalctx)
					if err != nil {
						fmt.Println(err.Error())
					} else {
						fmt.Println(signal.String())
					}
				// `cancelctx.Done()` Because of the context to cancel by signal.
				case <-cancelctx.Done():
					fmt.Println(cancelctx.Err())
					return
				}
			}
		}()
	}
	wg.Wait()
}
