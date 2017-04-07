package main

import (
	"context"
	"fmt"
	"sync"
	"syscall"

	"github.com/Code-Hex/sigctx"
)

func main() {

	ctx := sigctx.WithSignals(
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
				// You can notify signal received.
				case <-sigctx.Recv(ctx):
					signal, err := sigctx.Signal(ctx)
					if err != nil {
						fmt.Println(err.Error())
						return
					}
					switch signal {
					case syscall.SIGINT:
						fmt.Println(signal.String())
						return
					default:
						fmt.Println(signal.String())
					}
				case <-ctx.Done():
					fmt.Println(ctx.Err())
					return
				}
			}
		}()
	}
	wg.Wait()
}
