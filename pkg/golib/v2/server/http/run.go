package http

import (
	"log"
	"net"
	"net/http"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"permission/pkg/golib/v2/env"
	"permission/pkg/golib/v2/server/signal"
)

type ServerConfig struct {
	Address      string        `yaml:"address"`
	ReadTimeout  time.Duration `yaml:"readtimeout"`
	WriteTimeout time.Duration `yaml:"writetimeout"`
	CloseChan    chan struct{}
}

func (conf *ServerConfig) check() {
	if strings.Trim(conf.Address, " ") == "" {
		conf.Address = ":8080"
	}
}

func Start(engine *gin.Engine, conf ServerConfig) (err error) {
	if env.IsDockerPlatform() {
		// unix 监听方式
		return StartUnix(engine, conf)
	} else {
		// tcp 监听方式
		return StartTCP(engine, conf)
	}
}

func serve(engine *gin.Engine, listener net.Listener, conf ServerConfig) (err error) {
	appServer := &http.Server{
		Handler: engine,
	}

	appServer.RegisterOnShutdown(func() {
		conf.CloseChan <- struct{}{}
	})

	// 超时时间 (如果设置的太小，可能导致接口响应时间超过该值，进而导致504错误)
	if conf.ReadTimeout > 0 {
		appServer.ReadTimeout = conf.ReadTimeout
	}

	if conf.WriteTimeout > 0 {
		appServer.WriteTimeout = conf.WriteTimeout
	}

	processed := make(chan struct{})
	go signal.Handle(processed)

	log.Println(syscall.Getpid(), listener.Addr().String())

	signal.RegisterShutdown("httpServer", appServer.Shutdown)
	if err := appServer.Serve(listener); err != http.ErrServerClosed {
		log.Fatalf("server not gracefully shutdown, err :%v\n", err)
	}

	log.Println("waiting for the registered service to finish...")
	<-processed

	return nil
}
