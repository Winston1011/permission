package http

import (
	"github.com/gin-gonic/gin"
)

func StartTCP(engine *gin.Engine, conf ServerConfig) error {
	conf.check()
	return engine.Run(conf.Address)
}
