package demo

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"time"

	"permission/helpers"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"permission/pkg/golib/v2/env"
)

const (
	// proxy 支持的hints
	HintsReadWrite = "/*#mode=READWRITE*/"
	HintsReadOnly  = "/*#mode=READONLY*/"
)

type Option func(*gorm.DB) *gorm.DB

func getDB(ctx *gin.Context, tableName string) *gorm.DB {
	return helpers.MysqlClientDemo.Table(tableName).WithContext(ctx)
}

type FilterOption struct {
	CreateStartTime time.Time
	CreateEndTime   time.Time
	IsNeedCnt       bool
	IsNeedList      bool
}

// --------------------------- scopes ---------------------

/*
scopes 实现代码共享。
Scopes 使你可以复用通用的逻辑，共享的逻辑需要定义为 func(*gorm.DB) *gorm.DB 类型
*/

func WithID(id int) Option {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("`id` = ?", id)
	}
}

func WithName(name string) Option {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("`name` = ?", name)
	}
}

func WithNames(name []string) Option {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("`name` in ?", name)
	}
}

func WithValidStatus(db *gorm.DB) *gorm.DB {
	return db.Where("del_flag = ?", NotDel)
}

func FilterStatus(status []int8) Option {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("del_flag in (?)", status)
	}
}

func FilterCreateStartTime(start time.Time) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("create_time > ?", start)
	}
}
func FilterCreateEndTime(end time.Time) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("create_time < ?", end)
	}
}

// --------------------------- 分页相关 ---------------------
type NormalPage struct {
	No   int // 当前第几页
	Size int // 每页大小
}

// 传统分页示例
func NormalPaginate(page *NormalPage) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		pageNo := 1
		if page.No > 0 {
			pageNo = page.No
		}

		pageSize := page.Size
		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}

		offset := (pageNo - 1) * pageSize
		return db.Order("id asc").Offset(offset).Limit(pageSize)
	}
}

// 瀑布流分页示例
type ScrollPage struct {
	Start int // 当前页开始标示
	Size  int // 每页大小
}

func ScrollingPaginate(page *ScrollPage) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		start := -1
		if page.Start > 0 {
			start = page.Start
		}

		pageSize := page.Size
		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}

		return db.Where("id > ?", start).Order("id asc").Limit(pageSize)
	}
}

/*
自定义类型示例
*/

// 对外以秒级时间戳展示
type UnixTime struct {
	Time time.Time
}

// 记录json marshal 时用
func (t UnixTime) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%v\"", t.Time.Unix())), nil
}

func (t *UnixTime) String() (driver.Value, error) {
	return t.Time.Unix(), nil
}

// 写入数据库之前，对数据做类型转换
func (t UnixTime) Value() (driver.Value, error) {
	return t.Time, nil
}

// 将数据库中取出的数据，赋值给目标类型
func (t *UnixTime) Scan(v interface{}) error {
	switch vt := v.(type) {
	case time.Time:
		t.Time = vt
	default:
		return errors.New("format error")
	}

	return nil
}

/*
将用户手机号加密的示例:
入库的时候对手机号加密；从数据库取出后对手机号解密；
*/

type PhoneNumber string

func (t PhoneNumber) Value() (driver.Value, error) {
	return env.EncodeDBSensitiveField(string(t)), nil
}

func (t *PhoneNumber) Scan(v interface{}) error {
	switch vt := v.(type) {
	case []byte:
		// 从数据库里读出后解密
		*t = PhoneNumber(env.DecodeDBSensitiveField(string(vt)))
	case string:
		*t = PhoneNumber(env.DecodeDBSensitiveField(vt))
	default:
		return errors.New("format error")
	}
	return nil
}
