package redis

import (
	"time"

	"github.com/gin-gonic/gin"
	redigo "github.com/gomodule/redigo/redis"
	"permission/pkg/golib/v2/utils"
	"permission/pkg/golib/v2/zlog"
)

func (r *Redis) Lua(ctx *gin.Context, script string, keyCount int, keysAndArgs ...interface{}) (interface{}, error) {
	start := time.Now()

	lua := redigo.NewScript(keyCount, script)

	conn := r.pool.Get()
	if err := conn.Err(); err != nil {
		r.logger.Error("get connection error: "+err.Error(), r.commonFields(ctx)...)
		return nil, err
	}
	defer conn.Close()

	reply, err := lua.Do(conn, keysAndArgs...)

	ralCode := 0
	msg := "pipeline exec succ"
	if err != nil {
		ralCode = -1
		msg = "pipeline exec error: " + err.Error()
	}
	end := time.Now()

	fields := append(r.commonFields(ctx),
		zlog.String("remoteAddr", r.remoteAddr),
		zlog.String("service", r.service),
		zlog.String("requestStartTime", utils.GetFormatRequestTime(start)),
		zlog.String("requestEndTime", utils.GetFormatRequestTime(end)),
		zlog.Float64("cost", utils.GetRequestCost(start, end)),
		zlog.Int("ralCode", ralCode),
	)

	r.logger.Info(msg, fields...)

	return reply, err
}
