package hbase

import (
	"errors"
	"fmt"
	"time"

	"git.apache.org/thrift.git/lib/go/thrift"
	"github.com/gin-gonic/gin"
	secret "permission/pkg/golib/v2/env"
	"permission/pkg/golib/v2/pool/connpool"
	"permission/pkg/golib/v2/utils"
	"permission/pkg/golib/v2/zlog"
)

const prefix = "@@hbase."

// hbase conf
type HBaseConf struct {
	Service     string        `yaml:"service"`
	Addr        string        `yaml:"addr"`
	Timeout     time.Duration `yaml:"timeout"`
	MaxIdle     int           `yaml:"maxIdle"`
	MaxActive   int           `yaml:"maxActive"`
	IdleTimeout time.Duration `yaml:"idleTimeout"`
	WaitTimeOut time.Duration `yaml:"waitTimeout"`
}

func (conf *HBaseConf) checkConf() {
	secret.CommonSecretChange(prefix, *conf, conf)

	if conf.IdleTimeout == 0 {
		conf.IdleTimeout = 15 * time.Second
	}
	if conf.MaxActive == 0 {
		conf.MaxActive = 10
	}
	if conf.MaxIdle == 0 {
		conf.MaxIdle = 5
	}
}

func initPool(conf HBaseConf) (p connpool.Pool, err error) {
	fields := []zlog.Field{
		zlog.String(zlog.TopicType, zlog.LogNameModule),
		zlog.String("prot", "hbase"),
	}

	createConn := func() (interface{}, error) {
		trans, err := thrift.NewTSocket(conf.Addr)
		if err != nil {
			zlog.ErrorLogger(nil, "init Hbase pool error: "+err.Error(), fields...)
			return nil, err
		}

		if err := trans.SetTimeout(conf.Timeout); err != nil {
			zlog.WarnLogger(nil, "hbase set timeout error: "+err.Error(), fields...)
		}

		if err = trans.Open(); err != nil {
			zlog.WarnLogger(nil, "HBase open error: "+err.Error(), fields...)
			return nil, errors.New("HBase open error: " + err.Error())
		}

		f := thrift.NewTBinaryProtocolFactoryDefault()
		std := thrift.NewTStandardClient(f.GetProtocol(trans), f.GetProtocol(trans))
		c := NewHbaseClient(std)

		h := &poolClient{
			Client:  c,
			Trans:   trans,
			Service: conf.Service,
		}
		zlog.DebugLogger(nil, "open a new connection at: "+time.Now().String(), fields...)
		return h, nil
	}

	closeConn := func(client interface{}) error {
		if client != nil {
			zlog.DebugLogger(nil, "close connection at: "+time.Now().String(), fields...)
			return client.(*poolClient).Trans.Close()
		}
		return nil
	}
	poolConfig := &connpool.Config{
		Factory:     createConn,
		Close:       closeConn,
		Ping:        nil,
		InitialCap:  conf.MaxIdle,     // 资源池初始连接数
		MaxIdle:     conf.MaxIdle,     // 最大空闲连接数
		MaxCap:      conf.MaxActive,   // 最大并发连接数
		IdleTimeout: conf.IdleTimeout, // 连接最大空闲时间
		WaitTimeOut: conf.WaitTimeOut, // 获取连接最大等待时间
	}

	return connpool.NewChannelPool(poolConfig)
}

type poolClient struct {
	Client  *HbaseClient
	Trans   thrift.TTransport
	Service string
}

func (h *poolClient) Do(ctx *gin.Context, efunc func(c *HbaseClient) error) (err error) {
	if h.Client == nil {
		return errors.New("get client error")
	}
	fields := []zlog.Field{
		zlog.String(zlog.TopicType, zlog.LogNameModule),
		zlog.String("prot", "hbase"),
		zlog.String("service", h.Service),
	}

	start := time.Now()
	err = efunc(h.Client)
	ralCode := 0
	msg := "hbase exec success"
	if err != nil {
		ralCode = -1
		msg = fmt.Sprintf("hbase exec error: %s", err.Error())
		zlog.ErrorLogger(ctx, msg, fields...)
	}

	end := time.Now()

	fields = append(fields,
		zlog.String("requestStartTime", utils.GetFormatRequestTime(start)),
		zlog.String("requestEndTime", utils.GetFormatRequestTime(end)),
		zlog.Float64("cost", utils.GetRequestCost(start, end)), // 执行时间 单位:毫秒
		zlog.Int("ralCode", ralCode),
	)

	zlog.InfoLogger(ctx, msg, fields...)
	return err
}

// 使用连接池的hbase client
type HBaseClient struct {
	Service string
	pool    connpool.Pool
}

func NewHBaseClient(conf HBaseConf) (c *HBaseClient, err error) {
	// 为当前实例打开一个连接池
	conf.checkConf()
	t, err := initPool(conf)
	if err != nil {
		return nil, err
	}
	c = &HBaseClient{
		Service: conf.Service,
		pool:    t,
	}

	return c, nil
}

func (c *HBaseClient) Exec(ctx *gin.Context, efunc func(c *HbaseClient) error) (err error) {
	// get a conn from pool
	thriftClient, err := c.GetConn(ctx)
	if err != nil {
		return err
	}

	// execute
	err = thriftClient.Do(ctx, efunc)

	// put back to pool
	_ = c.Release(thriftClient)

	return err
}

func (c *HBaseClient) GetConn(ctx *gin.Context) (client *poolClient, err error) {
	var conn interface{}
	conn, err = c.pool.Get(ctx)

	fields := []zlog.Field{
		zlog.String(zlog.TopicType, zlog.LogNameModule),
		zlog.String("prot", "hbase"),
	}

	if err != nil {
		zlog.ErrorLogger(ctx, "get connection error: "+err.Error(), fields...)
		return nil, err
	}
	thriftClient, ok := conn.(*poolClient)
	if !ok || thriftClient == nil {
		// 无效的连接，直接关闭
		_ = c.pool.Close(thriftClient)
		zlog.ErrorLogger(ctx, "get a broken connection", fields...)
		return nil, errors.New("broken connection")
	}

	return thriftClient, nil
}

func (c *HBaseClient) Release(conn *poolClient) (err error) {
	ret := c.pool.Put(conn)
	return ret
}

// close pool when app shutdown
func (c *HBaseClient) Close() error {
	c.pool.Release()
	return nil
}

// pool conn stats
func (c *HBaseClient) Stats() (inUseCount, idleCount, activeCount int) {
	stats := c.pool.Stats()
	idleCount = stats.IdleCount
	activeCount = stats.ActiveCount
	inUseCount = activeCount - idleCount
	return inUseCount, idleCount, activeCount
}
