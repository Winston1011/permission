package models

import "gorm.io/gorm"

const (
	// proxy 支持的hints
	HintsReadWrite = "/*#mode=READWRITE*/"
	HintsReadOnly  = "/*#mode=READONLY*/"
)

type NormalPage struct {
	No   int // 当前第几页
	Size int // 每页大小
}

type Option struct {
	IsNeedCnt  bool
	IsNeedList bool
}

// NormalPaginate 传统分页示例
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
