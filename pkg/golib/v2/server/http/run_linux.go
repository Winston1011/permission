package http

import (
	"log"
	"net"

	"github.com/gin-gonic/gin"
)

func StartTCP(engine *gin.Engine, conf ServerConfig) (err error) {
	conf.check()

	listener, err := net.Listen("tcp", conf.Address)
	if err != nil {
		log.Fatal("listen unix error: " + err.Error())
	}

	defer listener.Close()
	return serve(engine, listener, conf)
}
