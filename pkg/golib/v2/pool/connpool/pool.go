package connpool

import "github.com/gin-gonic/gin"

// PoolStats contains pool statistics.
type Stats struct {
	// ActiveCount is the number of connections in the pool. The count includes
	// idle connections and connections in use.
	ActiveCount int
	// IdleCount is the number of idle connections in the pool.
	IdleCount int
}

type Pool interface {
	Get(ctx *gin.Context) (interface{}, error)

	Put(interface{}) error

	Close(interface{}) error

	Release()

	Len() int

	Stats() Stats
}
