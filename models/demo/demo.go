package demo

import (
	"time"

	"permission/components"
	"permission/helpers"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"permission/pkg/hints"
)

const (
	NotDel = 0
	Del    = 1
)

type Demo struct {
	ID         int       `gorm:"primary_key;column:id" json:"ID"`
	Name       string    `gorm:"column:name" json:"name"`
	Desc       string    `gorm:"column:desc" json:"desc"`
	CreateTime time.Time `gorm:"column:create_time;default:(-)" json:"createTime"`
	UpdateTime time.Time `gorm:"column:update_time;default:(-)" json:"updateTime"`
	DelFlag    int8      `gorm:"column:del_flag" json:"DelFlag"`
}

func (d Demo) TableName() string {
	return "demo"
}

/*
单个插入。

sql示例：
INSERT INTO `demo` (`name`,`desc`,`del_flag`) VALUES ('permission','this is permission desc',0)
*/

func (d *Demo) Insert(ctx *gin.Context, db *gorm.DB) (id int, rows int64, err error) {
	if db == nil {
		db = helpers.MysqlClientDemo
	}
	result := db.WithContext(ctx).Create(d)

	// 可以获得数据库插入时获得的自增ID的值, 需要在定义结构体时使用 gorm tag 指明是 primary key： gorm:"primary_key"
	id = d.ID

	err = result.Error
	rows = result.RowsAffected
	if err != nil || rows < 1 {
		err = components.ErrorDbInsert.WrapPrintf(result.Error, "template insert error, input: %+v", d)
		return id, rows, err
	}

	return id, rows, err
}

/*
批量插入。

sql示例：
INSERT INTO `demo` (`name`,`desc`,`del_flag`) VALUES
('demo_0','this is desc0',0),('demo_1','this is desc1',0),('demo_2','this is desc2',0)
*/
func DemosBatchInsert(ctx *gin.Context, db *gorm.DB, demos []Demo) (rows int64, err error) {
	if len(demos) == 0 {
		return rows, err
	}

	if db == nil {
		db = helpers.MysqlClientDemo
	}

	result := db.WithContext(ctx).Create(demos)

	err = result.Error
	rows = result.RowsAffected
	if err != nil || rows < 1 {
		err = components.ErrorDbInsert.WrapPrintf(result.Error, "demo insert error")
		return rows, err
	}

	return rows, err
}

/*
不存在则插入，存在则忽略。

sql示例：
INSERT INTO `demo` (`name`,`desc`,`del_flag`,`id`)
VALUES ('permission','this is TestUpsertIgnore desc',0,5)
ON DUPLICATE KEY UPDATE `id`=`id`
*/
func (d *Demo) UpsertIgnore(ctx *gin.Context) (rows int64, err error) {
	result := helpers.MysqlClientDemo.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoNothing: true,
	}).Create(&d)

	err = result.Error
	rows = result.RowsAffected
	if err != nil {
		err = components.ErrorDbInsert.WrapPrintf(result.Error, "demo insert error")
		return rows, err
	}
	return rows, nil
}

/*
不存在则插入，存在则更新指定字段。

sql示例：
INSERT INTO `demo` (`name`,`desc`,`del_flag`,`id`)
VALUES ('goweb-upsert','this is TestUpsert, should update value ',0,6)
ON DUPLICATE KEY UPDATE `name`=VALUES(`name`),`desc`=VALUES(`desc`)
*/
func (d *Demo) UpsertDemo(ctx *gin.Context) (rows int64, err error) {
	result := helpers.MysqlClientDemo.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "desc"}),
	}).Create(d)

	err = result.Error
	rows = result.RowsAffected
	if err != nil {
		err = components.ErrorDbInsert.WrapPrintf(result.Error, "demo insert error")
		return rows, err
	}
	return rows, nil
}

/*
根据指定字段更新字段。

sql示例：
UPDATE `demo` SET `desc`='this is TestUpdate',`status`=3 WHERE name = 'goweb-upsert' AND `status` = 1
*/
func UpdateByName(ctx *gin.Context, db *gorm.DB, name string, option, fields map[string]interface{}) (rows int64, err error) {
	if len(fields) == 0 {
		return 0, nil
	}

	if db == nil {
		db = helpers.MysqlClientDemo
	}

	result := db.
		Model(&Demo{}). // 注意这里，如果使用d而非一个空的Demo对象 且 d.ID != 0，那么where条件中会默认增加 where id = d.ID
		WithContext(ctx).
		Where("name = ?", name).
		Where(option).
		Updates(fields)

	rows, err = result.RowsAffected, result.Error
	if err != nil {
		msg := "demo UpdateByName error, input: " + name
		err := components.ErrorDbUpdate.WrapPrint(err, msg)
		return rows, err
	}
	return rows, nil
}

/*
检索单个对象。
GORM 提供了 First、Take、Last 方法，以便从数据库中检索单个对象。
当查询数据库时它添加了 LIMIT 1 条件，且没有找到记录时，它会返回 ErrRecordNotFound 错误!

如果上层不想处理 ErrRecordNotFound 的错误，可以改为返回slice：
var demos []Demo
result := helpers.MysqlClientDemo.Where("id = ?", id).Ctx(ctx).Take(&demos)
这样上层可以通过 len(demos) == 0 来判断是否查询到了记录。


// 获取第一条记录（主键升序）
db.First(&demo)
SELECT * FROM `demo` WHERE name = 'permission' ORDER BY `demo`.`id` LIMIT 1;

// 获取一条记录，没有指定排序字段
db.Take(&demo)
SELECT * FROM `demo` WHERE name = 'permission';

// 获取最后一条记录（主键降序）
db.Last(&demo)
SELECT * FROM `demo` WHERE name = 'permission' ORDER BY `demo`.`id` DESC LIMIT 1;

*/
func GetDemoByID(ctx *gin.Context, id int) (demo Demo, err error) {
	result := helpers.MysqlClientDemo.Where("id = ?", id).WithContext(ctx).Take(&demo)
	err = result.Error
	if err != nil {
		// 上游业务需要利用此 error 判断是否是 ErrRecordNotFound
		// 所以，该 error 不能作为msg被追加到其他error后，只能做为 cause error 返回
		return demo, err
	}

	return demo, nil
}

/*
检索全部对象。
Find() 指定返回多行记录，当返回的slice为空时，说明没有命中记录；

sql示例：
SELECT * FROM `demo` WHERE name in ('permission','demo_0')
*/

func GetDemoByName(ctx *gin.Context, names []string) (demo []Demo, err error) {
	result := helpers.MysqlClientDemo.Where("name in ?", names).WithContext(ctx).Find(&demo)
	err = result.Error
	if err != nil {
		return demo, components.ErrorDbSelect.WrapPrintf(err, "input name=%v", names)
	}

	return demo, nil
}

// 对于延迟特别敏感的业务，不能够接受延迟的， 可以开启事物或者配置注释来查询主库，
// 使用注释的话，在SQL中SELECT字段之后插入 /*#mode=READWRITE*/
// sql 示例：
// SELECT /*#mode=READWRITE*/ * FROM `demo` WHERE id = 2
func GetDemoByIDUsingHits(ctx *gin.Context, id int) (demo []Demo, err error) {
	result := helpers.MysqlClientDemo.WithContext(ctx).Where("id = ?", id).Clauses(hints.NewHint(HintsReadWrite)).Find(&demo)
	err = result.Error
	if err != nil {
		return demo, components.ErrorDbSelect.WrapPrintf(err, "input id=%d", id)
	}

	return demo, nil
}

func UpdateDemo(ctx *gin.Context, fields map[string]interface{}, options ...Option) (rows int64, err error) {
	if len(fields) == 0 {
		return 0, nil
	}

	db := getDB(ctx, Demo{}.TableName())
	for _, option := range options {
		db = option(db)
	}

	result := db.Updates(fields)

	rows, err = result.RowsAffected, result.Error
	if err != nil {
		err := components.ErrorDbUpdate.Wrap(err)
		return rows, err
	}
	return rows, nil
}

/*
关于 Demo表的各种查询
使用需要注意，尽量包含一个带索引的 option
*/
func Info(ctx *gin.Context, options ...Option) (ts []Demo, err error) {
	db := getDB(ctx, Demo{}.TableName())
	for _, option := range options {
		db = option(db)
	}
	err = db.Find(&ts).Error
	if err != nil {
		return nil, err
	}

	return ts, nil
}

// 传统分页示例
func GetNormalList(ctx *gin.Context, option *FilterOption, page *NormalPage) (demos []Demo, cnt int, err error) {
	if !option.IsNeedCnt && !option.IsNeedList {
		return demos, cnt, nil
	}
	db := helpers.MysqlClientDemo.WithContext(ctx).Model(&Demo{})
	if !option.CreateStartTime.IsZero() {
		db = db.Scopes(FilterCreateStartTime(option.CreateStartTime))
	}
	if !option.CreateEndTime.IsZero() {
		db = db.Scopes(FilterCreateEndTime(option.CreateEndTime))
	}

	db = db.Scopes(WithValidStatus)

	if option.IsNeedCnt {
		var c int64
		db = db.Count(&c)
		cnt = int(c)
	}

	if option.IsNeedList {
		db = db.Scopes(NormalPaginate(page)).Find(&demos)
	}

	if db.Error != nil {
		return demos, cnt, components.ErrorDbSelect.Wrap(db.Error)
	}

	return demos, cnt, nil
}

// 瀑布流分页示例
func GetFlowList(ctx *gin.Context, option *FilterOption, page *ScrollPage) (demos []Demo, err error) {
	db := helpers.MysqlClientDemo.WithContext(ctx)

	if !option.CreateStartTime.IsZero() {
		db = db.Scopes(FilterCreateStartTime(option.CreateStartTime))
	}
	if !option.CreateEndTime.IsZero() {
		db = db.Scopes(FilterCreateEndTime(option.CreateEndTime))
	}

	db = db.Scopes(WithValidStatus).Scopes(ScrollingPaginate(page)).Find(&demos)
	if db.Error != nil {
		return demos, components.ErrorDbSelect.Wrap(db.Error)
	}

	return demos, nil
}
