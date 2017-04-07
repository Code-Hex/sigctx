package sigctx

import (
	"context"
	"os"
	"runtime"
	"sync"
	"syscall"
	"testing"
	"time"
)

var (
	bgctx = context.Background()
	cpu   = runtime.NumCPU()
	pid   = os.Getpid()
)

func init() {
	runtime.GOMAXPROCS(cpu)
}

func TestWithSignals(t *testing.T) {

	ctx, cancel := context.WithCancel(bgctx)
	sctx := WithSignals(ctx, syscall.SIGUSR1)
	counter := make(map[string]int, 2)
	cntchan := make(chan string)

	go func() {
		time.Sleep(time.Millisecond * 500)
		p, err := os.FindProcess(pid)
		if err != nil {
			t.Fatal("Failed to get pid")
		}
		p.Signal(syscall.SIGUSR1)
		p.Signal(syscall.SIGUSR2)
		cancel()
	}()

	go func() {
		for {
			select {
			case signame := <-cntchan:
				counter[signame]++
			case <-ctx.Done():
				return
			}
		}
	}()

	var wg sync.WaitGroup
	for i := 0; i < cpu; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case <-Recv(sctx):
				signal, err := Signal(sctx)
				if err != nil {
					t.Error("Failed to catch signal")
				}
				switch signal {
				case syscall.SIGUSR1, syscall.SIGUSR2:
					cntchan <- signal.String()
				default:
					t.Errorf("Got signal. but unexpected it...")
				}
			case <-time.After(3 * time.Second):
				t.Error("Timeout")
				return
			case <-ctx.Done():
				return
			}
		}()
	}
	wg.Wait()

	// check to send signals
	for k := range counter {
		if counter[k] != cpu {
			t.Errorf("count got %d, expected %d", counter[k], cpu)
		}
	}
}
