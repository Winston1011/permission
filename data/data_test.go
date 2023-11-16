package data_test

import (
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"permission/helpers"
	"permission/models/demo"
	"permission/models/user"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"permission/pkg/golib/v2/env"
	"permission/pkg/golib/v2/zlog"
)

func TestKms_Encrypt(t *testing.T) {
	phoneNum := "13512340010"
	encryptPhoneNum, err := helpers.Kms.Encrypt(ctx, phoneNum)
	if err != nil {
		zlog.Warnf(ctx, "[kms_Encrypt] error is %+v", err)
		return
	}

	t.Log("[TestKms_Encrypt] after encrypt: ", encryptPhoneNum)

	plainText, err := helpers.Kms.Decrypt(ctx, encryptPhoneNum)
	if err != nil || plainText == "" {
		zlog.Warnf(ctx, "[kms_Decrypt_fail]  cipherKms: %s, error(%+v) ", plainText, err)
		return
	}

	t.Log("[TestKms_Decrypt] after decrypt: ", plainText)

	assert.Equal(t, phoneNum, plainText)
}

// 更多 mysql 示例查看 models/*_test.go
func TestMysql_Query(t *testing.T) {
	// 查询
	names := []string{"permission", "demo_0"}
	demos, err := demo.GetDemoByName(ctx, names)
	if err != nil {
		t.Error("[TestMysql_Query] error: ", err.Error())
	}
	zlog.Debugf(ctx, "we have got %d num", len(demos))

	// 分表的查询示例
	userIDList := []int{1, 2, 3, 4, 5, 6}
	list, err := user.GetUserByUserIDList(ctx, userIDList)
	if err != nil {
		t.Error("[TestMysql_Query] error: ", err.Error())
		return
	}
	t.Logf("user list: %+v", list)
}

func TestGCache(t *testing.T) {
	helpers.Cache1.Set("test", "testValue", time.Second*2)
	helpers.Cache2.Set("curCourseID", "7788", time.Second*3)

	time.Sleep(2 * time.Second)

	v, e := helpers.Cache1.Get("test")
	assert.Equal(t, e, false)

	v, e = helpers.Cache2.Get("curCourseID")
	assert.Equal(t, e, true)
	assert.Equal(t, v, "7788")
}

/*
	TestMain 示例:
	一个包内的所有*_test.go 里的 Testxxx(t * testing.T) 都会调用TestMain
	可以为一个包内的所有test函数初始化环境变量
*/

var ctx *gin.Context

func TestMain(m *testing.M) {
	w := httptest.NewRecorder()
	ctx, _ = gin.CreateTestContext(w)

	// 这里最好用相对路径，便于协作开发。且文件确定好后，相对于conf的路径一般是固定的。
	// 这里指的是项目路径，比如 /home/homework/goweb
	env.SetRootPath("../")

	// 初始化全局变量。
	// 注意如果test case中用到job，需要使用 gin.New() 生成 engine 传下去
	helpers.PreInit()
	helpers.InitResource(nil)
	m.Run()
	os.Exit(0)
}
