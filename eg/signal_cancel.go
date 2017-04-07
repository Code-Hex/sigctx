package main

import (
	"context"
	"fmt"
	"sync"
	"syscall"

	"github.com/Code-Hex/sigctx"
)

func main() {

	ctx := sigctx.WithCancelSignals(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGHUP,
	)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				// You can cancel by send signal.
				case <-ctx.Done():
					fmt.Println(ctx.Err())
					return
				}
			}
		}()
	}
	wg.Wait()
}
