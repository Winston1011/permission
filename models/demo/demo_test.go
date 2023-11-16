package demo_test

import (
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/pkg/errors"

	"permission/helpers"
	"permission/models/demo"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"permission/pkg/golib/v2/env"
)

/*
	如若测试，可将 sql/init.sql 导入数据库，然后修改 conf/resource.yaml 中 mysql的addr/user/password 相关信息。
*/

// 单个插入
func TestInsert(t *testing.T) {
	name := "permission"
	desc := "this is permission desc"
	d := &demo.Demo{
		Name: name,
		Desc: desc,
	}
	id, rows, err := d.Insert(ctx, nil)
	if err != nil {
		t.Errorf("err want to got nil, got error(%s)", err.Error())
		return
	}

	if rows != 1 {
		t.Errorf("affect row want to got %d, got error(%d)", 1, rows)
	}

	t.Log("insert success, id: ", id)
}

// 批量插入
func TestBatchInsert(t *testing.T) {
	var demos []demo.Demo
	cnt := 3
	for i := 0; i < cnt; i++ {
		idx := strconv.Itoa(i)
		d := demo.Demo{
			Name:    "demo_" + idx,
			Desc:    "this is desc" + idx,
			DelFlag: demo.NotDel,
		}
		demos = append(demos, d)
	}

	rows, err := demo.DemosBatchInsert(ctx, nil, demos)
	if err != nil {
		t.Errorf("err want to got nil, got error(%s)", err.Error())
		return
	}

	if rows != int64(cnt) {
		t.Errorf("affect row want to got %d, got error(%d)", cnt, rows)
	}

	t.Log("batch insert success")
}

// key存在则忽略，key 不存在则插入
func TestUpsertIgnore(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	id := 5
	name := "permission"
	desc := "this is TestUpsertIgnore desc"
	d := &demo.Demo{
		ID:      id,
		Name:    name,
		Desc:    desc,
		DelFlag: demo.NotDel,
	}
	rows, err := d.UpsertIgnore(ctx)
	if err != nil {
		t.Errorf("err want to got nil, got error(%s)", err.Error())
		return
	}

	if rows < 1 {
		t.Logf("row(name= %s )has alreay exit, ignore", name)
	} else if rows > 1 {
		t.Errorf("affect row want to got 0 or 1, got %d", rows)
	}

	t.Log("upsert success")
}

// key 不存在则插入，存在则更新字段
func TestUpsert(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	id := 6
	name := "goweb-upsert"
	desc := "this is TestUpsert, should update value "
	d := &demo.Demo{
		ID:   id,
		Name: name,
		Desc: desc,
	}
	rows, err := d.UpsertDemo(ctx)
	if err != nil {
		t.Errorf("err want to got nil, got error(%s)", err.Error())
		return
	}

	switch {
	case rows < 1:
		t.Logf("row(name= %s )has alreay exit, ignore", name)
	case rows > 1:
		t.Errorf("affect row want to got 0 or 1, got %d", rows)
	default:
		t.Log("upsert success")
	}
}

// 根据key更新字段
func TestUpdate(t *testing.T) {
	name := "goweb-upsert"                                                               // 更新的主条件，一般是索引
	option := map[string]interface{}{"del_flag": demo.NotDel}                            // option 可作为选传的where条件,可不指定
	fields := map[string]interface{}{"desc": "this is TestUpdate", "del_flag": demo.Del} // fields 作为被更新的字段

	rows, err := demo.UpdateByName(ctx, nil, name, option, fields)
	if err != nil {
		t.Errorf("err want to got nil, got error(%s)", err.Error())
		return
	}
	t.Logf("UpdateDescByName affect row got %d", rows)
}

/*
	开启事务一定要注意提交/回滚事务，否则可能会导致db连接泄露的问题，进而引起服务的阻塞问题。
*/
func TestDemo_Transaction(t *testing.T) {
	// 开始事务
	db := helpers.MysqlClientDemo
	tx := db.Begin()

	// 这里默认做了Rollback() 能够有效的减少业务忘记 commit 的情况
	// 在 Rollback() 开始之初，会判断事务是否已结束，如果已结束则返回一个 ErrTxDone 的错误（因此一般情况也无需处理 defer 这里的Rollback返回的错误）
	defer tx.Rollback()

	name := "permission"
	desc := "this is permission desc"
	d := &demo.Demo{
		Name: name,
		Desc: desc,
	}
	if _, _, err := d.Insert(ctx, db); err != nil {
		t.Error("db insert error: ", err.Error())
		return
	}

	if _, err := demo.UpdateByName(ctx, db, "goweb-upsert", nil, map[string]interface{}{"desc": "this is TestUpdate", "del_flag": demo.Del}); err != nil {
		t.Error("db insert error: ", err.Error())
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		t.Error("commit error: ", err.Error())
		return
	}

	t.Log("success for test transaction")
}

// 查询一条记录
func TestSelectOne(t *testing.T) {
	ctx.Set("logID", "222222")
	d, err := demo.GetDemoByID(ctx, 2)
	if err != nil && errors.Cause(err) != gorm.ErrRecordNotFound {
		t.Errorf("err want to got nil, got error(%s)", err.Error())
		return
	}

	if errors.Cause(err) == gorm.ErrRecordNotFound {
		t.Log("not found")
		return
	}

	t.Logf("success: %+v", d)
}

// 检索全部对象
func TestSelectAll(t *testing.T) {
	names := []string{"permission", "demo_0"}
	demos, err := demo.GetDemoByName(ctx, names)
	if err != nil {
		t.Errorf("err want to got nil, got error(%s)", err.Error())
		return
	}

	t.Logf("success: %+v", demos)
}

// hits
func TestHits(t *testing.T) {
	demos, err := demo.GetDemoByIDUsingHits(ctx, 2)
	if err != nil {
		t.Errorf("err want to got nil, got error(%s)", err.Error())
		return
	}

	// 返回包含一条记录的slice，这样上游方便通过len来判断是否检索到了记录。
	if len(demos) == 0 {
		t.Log("not found")
		return
	}

	t.Logf("success: %+v", demos[0])
}

/*
查询一条记录，可以指定多个 option ， 以下示例执行的查询：
SELECT * FROM `demo` WHERE `id` = 2 AND `name` = 'permission'
*/
func TestDemoInfo(t *testing.T) {
	ts, err := demo.Info(ctx, demo.WithID(2), demo.WithName("permission"))
	if err != nil {
		t.Errorf("err want to got nil, got error(%s)", err.Error())
		return
	}

	t.Logf("success: %+v", ts)
}

/*
通过option方式更新
UPDATE `demo` SET `desc`='[TestUpdateDemo]' WHERE `name` in ('demo_0','demo_1') AND del_flag = 0
*/
func TestUpdateDemo(t *testing.T) {
	fields := map[string]interface{}{"desc": "[TestUpdateDemo]"} // fields 作为被更新的字段

	rows, err := demo.UpdateDemo(ctx, fields, demo.WithNames([]string{"demo_0", "demo_1"}), demo.WithValidStatus)
	if err != nil {
		t.Errorf("err want to got nil, got error(%s)", err.Error())
		return
	}
	t.Logf("UpdateDemo affect row got %d", rows)
}

// 传统分页示例
func TestGetNormalList(t *testing.T) {
	s, err1 := time.Parse("2006-01-02", "2020-10-10")
	e, err2 := time.Parse("2006-01-02", "2023-12-12")
	if err1 != nil || err2 != nil {
		t.Error("get time error")
		return
	}

	o := &demo.FilterOption{
		CreateStartTime: s,
		CreateEndTime:   e,
		IsNeedCnt:       true,
	}

	// 计算总行数，以便确定总页数
	pageSize := 2 // 每页展示的行数
	_, cnt, err := demo.GetNormalList(ctx, o, nil)
	if err != nil {
		t.Error("GetDemoList cnt got error: ", err.Error())
		return
	}
	totalPageSize := cnt/pageSize + 1 // 总页数

	o.IsNeedCnt = false
	o.IsNeedList = true
	for i := 1; i <= totalPageSize; i++ {
		p := &demo.NormalPage{
			No:   i,
			Size: pageSize,
		}
		demos, _, err := demo.GetNormalList(ctx, o, p)
		if err != nil {
			t.Errorf("err want to got nil, got error(%s)", err.Error())
			return
		}
		t.Logf("page %d : %+v", i, demos)
	}

	t.Log("success, record num: ", cnt)
}

// 瀑布流分页示例
func TestGetFlowList(t *testing.T) {
	s, err1 := time.Parse("2006-01-02", "2020-10-10")
	e, err2 := time.Parse("2006-01-02", "2023-12-12")
	if err1 != nil || err2 != nil {
		t.Error("get time error")
		return
	}

	o := &demo.FilterOption{
		CreateStartTime: s,
		CreateEndTime:   e,
	}
	p := &demo.ScrollPage{
		Start: -1,
		Size:  2,
	}

	total := 0
	for {
		demos, err := demo.GetFlowList(ctx, o, p)
		if err != nil {
			t.Errorf("err want to got nil, got error(%s)", err.Error())
			return
		}

		sz := len(demos)
		total += sz
		if sz == 0 {
			// 最后一页已为空，表示后续无元素了
			break
		}

		for _, d := range demos {
			t.Logf("start: %d , demo: %+v", p.Start, d)
		}

		// 更新start
		p.Start = demos[sz-1].ID
	}

	t.Log("success, record num : ", total)
}

var ctx *gin.Context

func TestMain(m *testing.M) {
	env.SetRootPath("../../")

	helpers.PreInit()
	helpers.InitMysql()

	w := httptest.NewRecorder()
	ctx, _ = gin.CreateTestContext(w)

	m.Run()

	os.Exit(0)
}
