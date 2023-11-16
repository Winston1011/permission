package signal

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// 信号处理入口函数
func Handle(processed chan struct{}) {
	c := make(chan os.Signal, 1)
	hookAbleSignals := []os.Signal{
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	}

	signal.Notify(c, hookAbleSignals...)
	<-c

	log.Println("start to shutdown services gracefully")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	closed := make(chan struct{})
	go shutdown(ctx, closed)

	select {
	case <-ctx.Done():
		log.Println("shutdown timeout: ", ctx.Err())
	case <-closed:
		log.Println("all services has been gracefully closed")
	}

	// notify main process exit
	close(processed)
}

// shutdown 相关方法
var shutService []Shutdown

type Shutdown struct {
	service string
	f       func(ctx context.Context) error
}

func RegisterShutdown(service string, f func(ctx context.Context) error) {
	shutService = append(shutService, Shutdown{
		service: service,
		f:       f,
	})
}

func shutdown(ctx context.Context, closed chan struct{}) {
	var wg sync.WaitGroup
	for _, s := range shutService {
		wg.Add(1)
		ss := s
		go func() {
			if err := ss.f(ctx); err != nil {
				log.Printf("service[%s] shutdown failed, error: %s", ss.service, err.Error())
			}
			wg.Done()
		}()
	}
	wg.Wait()

	close(closed)
}
