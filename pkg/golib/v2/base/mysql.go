package base

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus/collectors"

	"github.com/gin-gonic/gin"
	driver "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	ormUtil "gorm.io/gorm/utils"
	"permission/pkg/golib/v2/env"
	"permission/pkg/golib/v2/utils"
	"permission/pkg/golib/v2/zlog"
)

const prefix = "@@mysql."

type MysqlConf struct {
	Service           string        `yaml:"service"`
	DataBase          string        `yaml:"database"`
	Addr              string        `yaml:"addr"`
	User              string        `yaml:"user"`
	Password          string        `yaml:"password"`
	Charset           string        `yaml:"charset"`
	InterpolateParams string        `yaml:"interpolateParams"`
	MaxIdleConns      int           `yaml:"maxidleconns"`
	MaxOpenConns      int           `yaml:"maxopenconns"`
	ConnMaxIdlTime    time.Duration `yaml:"maxIdleTime"`
	ConnMaxLifeTime   time.Duration `yaml:"connMaxLifeTime"`
	ConnTimeOut       time.Duration `yaml:"connTimeOut"`
	WriteTimeOut      time.Duration `yaml:"writeTimeOut"`
	ReadTimeOut       time.Duration `yaml:"readTimeOut"`

	// sql 字段最大长度
	MaxSqlLen int `yaml:"maxSqlLen"`

	// db metrics switch
	CollectMetrics bool `yaml:"collectMetrics"`
}

func (conf *MysqlConf) checkConf() {
	env.CommonSecretChange(prefix, *conf, conf)

	if conf.MaxIdleConns == 0 {
		conf.MaxIdleConns = 50
	}
	if conf.MaxOpenConns == 0 {
		conf.MaxOpenConns = 50
	}
	if conf.ConnMaxIdlTime == 0 {
		conf.ConnMaxIdlTime = 5 * time.Minute
	}
	if conf.ConnMaxLifeTime == 0 {
		conf.ConnMaxLifeTime = 30 * time.Minute
	}
	if conf.ConnTimeOut == 0 {
		conf.ConnTimeOut = 3 * time.Second
	}
	if conf.WriteTimeOut == 0 {
		conf.WriteTimeOut = 1200 * time.Millisecond
	}
	if conf.ReadTimeOut == 0 {
		conf.ReadTimeOut = 1200 * time.Millisecond
	}
	if conf.InterpolateParams == "" {
		// 使用string，方便后续修改默认值
		conf.InterpolateParams = "false"
	}
	if conf.MaxSqlLen == 0 {
		// 日志中sql字段长度：
		// 如果不指定使用默认2048；如果<0表示不展示sql语句；否则使用用户指定的长度，过长会被截断
		conf.MaxSqlLen = 2048
	}
}

func InitMysqlClient(conf MysqlConf) (client *gorm.DB, err error) {
	conf.checkConf()

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?timeout=%s&readTimeout=%s&writeTimeout=%s&parseTime=True&loc=Asia%%2FShanghai&interpolateParams=%s",
		conf.User,
		conf.Password,
		conf.Addr,
		conf.DataBase,
		conf.ConnTimeOut,
		conf.ReadTimeOut,
		conf.WriteTimeOut,
		conf.InterpolateParams,
	)

	if conf.Charset != "" {
		dsn = dsn + "&charset=" + conf.Charset
	}

	l := newLogger(&conf)
	_ = driver.SetLogger(l)

	c := &gorm.Config{
		SkipDefaultTransaction:                   true,
		NamingStrategy:                           nil,
		FullSaveAssociations:                     false,
		Logger:                                   l,
		NowFunc:                                  nil,
		DryRun:                                   false,
		PrepareStmt:                              false,
		DisableAutomaticPing:                     false,
		DisableForeignKeyConstraintWhenMigrating: false,
		AllowGlobalUpdate:                        false,
		ClauseBuilders:                           nil,
		ConnPool:                                 nil,
		Dialector:                                nil,
		Plugins:                                  nil,
	}

	client, err = gorm.Open(mysql.Open(dsn), c)
	if err != nil {
		return client, err
	}

	sqlDB, err := client.DB()
	if err != nil {
		return client, err
	}

	// 注册 mysql collector
	if RuntimeMetricsRegister != nil && conf.CollectMetrics {
		RuntimeMetricsRegister.MustRegister(collectors.NewDBStatsCollector(sqlDB, conf.Service))
	}

	// SetMaxIdleConns 设置空闲连接池中连接的最大数量
	sqlDB.SetMaxIdleConns(conf.MaxIdleConns)

	// SetMaxOpenConns 设置打开数据库连接的最大数量
	sqlDB.SetMaxOpenConns(conf.MaxOpenConns)

	// SetConnMaxLifetime 设置了连接可复用的最大时间
	sqlDB.SetConnMaxLifetime(conf.ConnMaxLifeTime)

	// only for go version >= 1.15 设置最大空闲连接时间
	sqlDB.SetConnMaxIdleTime(conf.ConnMaxIdlTime)

	return client, nil
}

type ormLogger struct {
	Service   string
	Addr      string
	Database  string
	MaxSqlLen int
	logger    *zlog.Logger
}

func newLogger(conf *MysqlConf) *ormLogger {
	s := conf.Service
	if conf.Service == "" {
		s = conf.DataBase
	}

	return &ormLogger{
		Service:   s,
		Addr:      conf.Addr,
		Database:  conf.DataBase,
		MaxSqlLen: conf.MaxSqlLen,
		logger:    zlog.ZapLogger.WithOptions(zlog.AddCallerSkip(2)),
	}
}

// go-sql-driver error log
func (l *ormLogger) Print(args ...interface{}) {
	l.logger.Error(fmt.Sprint(args...), l.commonFields(nil)...)
}

// LogMode log mode
func (l *ormLogger) LogMode(level logger.LogLevel) logger.Interface {
	return l
}

// Info print info
func (l ormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	m := fmt.Sprintf(msg, append([]interface{}{ormUtil.FileWithLineNum()}, data...)...)
	// 非trace日志改为debug级别输出
	l.logger.Debug(m, l.commonFields(ctx)...)
}

// Warn print warn messages
func (l ormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	m := fmt.Sprintf(msg, append([]interface{}{ormUtil.FileWithLineNum()}, data...)...)
	l.logger.Warn(m, l.commonFields(ctx)...)
}

// Error print error messages
func (l ormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	m := fmt.Sprintf(msg, append([]interface{}{ormUtil.FileWithLineNum()}, data...)...)
	l.logger.Error(m, l.commonFields(ctx)...)
}

// Trace print sql message
func (l ormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	end := time.Now()
	elapsed := end.Sub(begin)
	cost := float64(elapsed.Nanoseconds()/1e4) / 100.0

	// 请求是否成功
	msg := "mysql do success"
	ralCode := -0
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		// 没有找到记录不统计在请求错误中
		msg = err.Error()
		ralCode = -1
	}

	sql, rows := fc()
	if l.MaxSqlLen < 0 {
		sql = ""
	} else if len(sql) > l.MaxSqlLen {
		sql = sql[:l.MaxSqlLen]
	}

	fileLineNum := ormUtil.FileWithLineNum()
	fields := l.commonFields(ctx)
	fields = append(fields,
		zlog.Int64("affectedrow", rows),
		zlog.String("requestEndTime", utils.GetFormatRequestTime(end)),
		zlog.String("requestStartTime", utils.GetFormatRequestTime(begin)),
		zlog.String("fileLine", fileLineNum),
		zlog.Float64("cost", cost),
		zlog.Int("ralCode", ralCode),
		zlog.String("sql", sql),
	)

	l.logger.Info(msg, fields...)
}

func (l ormLogger) commonFields(ctx context.Context) []zlog.Field {
	var logID, requestID, uri string
	if c, ok := ctx.(*gin.Context); (ok && c != nil) || (!ok && !utils.IsNil(ctx)) {
		logID, _ = ctx.Value(zlog.ContextKeyLogID).(string)
		requestID, _ = ctx.Value(zlog.ContextKeyRequestID).(string)
		uri, _ = ctx.Value(zlog.ContextKeyUri).(string)
	}

	fields := []zlog.Field{
		zlog.String(zlog.TopicType, zlog.LogNameModule),
		zlog.String("logId", logID),
		zlog.String("requestId", requestID),
		zlog.String("uri", uri),
		zlog.String("prot", "mysql"),
		zlog.String("module", env.GetAppName()),
		zlog.String("service", l.Service),
		zlog.String("addr", l.Addr),
		zlog.String("db", l.Database),
		zlog.String("localIp", env.LocalIP),
	}
	return fields
}
