package sigctx

import (
	"context"
	"fmt"
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
	_ctx, cancel := context.WithCancel(bgctx)
	go func() {
		time.Sleep(time.Millisecond * 100)
		p, err := os.FindProcess(pid)
		if err != nil {
			panic(fmt.Sprint("Failed to get pid"))
		}
		if err := p.Signal(syscall.SIGUSR1); err != nil {
			t.Error(err.Error())
		}
		time.Sleep(time.Millisecond * 50)
		if err := p.Signal(syscall.SIGUSR2); err != nil {
			t.Error(err.Error())
		}
		time.Sleep(time.Millisecond * 50)
		cancel()
	}()

	ctx := WithSignals(_ctx, syscall.SIGUSR1, syscall.SIGUSR2)
	counter := make(map[string]int)
	signamechan := make(chan string)

	go func() {
		for {
			select {
			case signame := <-signamechan:
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
			for {
				select {
				case <-Recv(ctx):
					signal, err := Signal(ctx)
					if err != nil {
						t.Error("Failed to catch signal")
					}
					switch signal {
					case syscall.SIGUSR1, syscall.SIGUSR2:
						signamechan <- signal.String()
					default:
						t.Errorf("Got signal. but it is unexpected")
					}
				case <-time.After(3 * time.Second):
					t.Error("Timeout")
					return
				case <-ctx.Done():
					return
				}
			}
		}()
	}
	wg.Wait()

	// check to send signals
	if len := len(counter); len != 2 {
		t.Errorf("Could not receive two signals: %d", len)
	}

	for k := range counter {
		if counter[k] != cpu {
			t.Errorf("Counter[%s] failed: got %d, expected %d", k, counter[k], cpu)
		}
	}
}

func TestWithCancelSignals(t *testing.T) {
	_ctx, cancel := context.WithCancel(bgctx)
	ctx := WithCancelSignals(_ctx, syscall.SIGUSR1)

	var once sync.Once
	defer once.Do(func() { cancel() })

	go func() {
		time.Sleep(time.Millisecond * 200)
		p, err := os.FindProcess(pid)
		if err != nil {
			panic(fmt.Sprint("Failed to get pid"))
		}
		p.Signal(syscall.SIGUSR1)
		time.Sleep(time.Second * 3)
		once.Do(func() { cancel() })
	}()

	var wg sync.WaitGroup
	for i := 0; i < cpu; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case <-time.After(3 * time.Second):
				t.Error("Timeout")
				return
			case <-ctx.Done():
				// check "signal received" message
				if ctx.Err() != Canceled {
					t.Errorf("got %s, expected %s", ctx.Err().Error(), Canceled.Error())
				}
				return
			}
		}()
	}
	wg.Wait()
}

type key string

func TestValue(t *testing.T) {
	key := key("hoge")
	value := "CodeHex"
	vctx := context.WithValue(bgctx, key, value)

	sctx := WithCancelSignals(vctx, syscall.SIGUSR1)

	if v, ok := sctx.Value(key).(string); ok {
		if v != value {
			t.Errorf("got %s, expected %s", v, value)
		}
	} else {
		t.Errorf("Failed to get string value")
	}

	ctx := WithSignals(sctx, syscall.SIGUSR2)

	if v, ok := ctx.Value(key).(string); ok {
		if v != value {
			t.Errorf("got %s, expected %s", v, value)
		}
	} else {
		t.Errorf("Failed to get string value")
	}
}

func TestDeadline(t *testing.T) {
	d := time.Now().Add(50 * time.Millisecond)
	ctx, cancel := context.WithDeadline(bgctx, d)
	var once sync.Once
	defer once.Do(func() { cancel() })

	sctx := WithSignals(ctx, syscall.SIGUSR1)

	select {
	case <-time.After(1 * time.Second):
		t.Error("Timeout")
		return
	case <-sctx.Done():
		if sctx.Err() != context.DeadlineExceeded {
			t.Errorf("got %s, expected %s",
				sctx.Err().Error(),
				context.DeadlineExceeded,
			)
		}
	}

}
