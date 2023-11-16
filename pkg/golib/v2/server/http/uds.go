package http

import (
	"log"
	"net"
	"os"

	"github.com/gin-gonic/gin"
)

const (
	socketPath = "/usr/local/var/run/"
	socketName = "go.sock"
	socketEnv = "SOCK_NAME"
)

func StartUnix(engine *gin.Engine, conf ServerConfig) (err error) {
	if _, err = os.Stat(socketPath); os.IsNotExist(err) {
		err = os.MkdirAll(socketPath, os.ModePerm)
		if err != nil {
			log.Fatal("mkdir " + socketPath + "error: " + err.Error())
			return err
		}
	}

	// socket name
	socketName := socketName
	if s := os.Getenv(socketEnv); s != "" {
		socketName = s
	}

	fd := socketPath + socketName
	if _, err = os.Stat(fd); err == nil {
		_ = os.Remove(fd)
	}

	listener, err := net.Listen("unix", fd)
	if err != nil {
		log.Fatal("listen unix error: " + err.Error())
		return err
	}
	defer listener.Close()
	// defer os.Remove(fd)

	if err = os.Chmod(fd, 0666); err != nil {
		log.Fatal("unix socket chmod error: " + err.Error())
		return err
	}

	return serve(engine, listener, conf)
}
