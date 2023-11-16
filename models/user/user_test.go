package user_test

import (
	"encoding/json"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"permission/helpers"
	"permission/models/demo"
	"permission/models/user"

	"github.com/gin-gonic/gin"
	"permission/pkg/golib/v2/env"
)

func TestSplitTable(t *testing.T) {
	u10 := &user.User{
		UserID:    9,
		UserName:  "go",
		UserPhone: "13112345678",
		Age:       1,
		CreateTime: demo.UnixTime{
			Time: time.Now(),
		},
	}
	err := u10.Insert(ctx)
	if err != nil {
		t.Error("insert error!")
		return
	}

	u11 := &user.User{
		UserID:   11,
		UserName: "js",
		Age:      1,
		CreateTime: demo.UnixTime{
			Time: time.Now(),
		},
	}
	err = u11.Insert(ctx)
	if err != nil {
		t.Error("insert error!")
		return
	}
}

func TestTableSplitGet(t *testing.T) {
	list, err := user.GetUserByUserIDList(ctx, []int{10, 11})
	if err != nil {
		t.Error("GetUserByUserID error: ", err.Error())
		return
	}
	for _, u := range list {
		t.Logf("userID=[%d], userName=[%s]", u.UserID, u.UserName)

		rest, err := json.Marshal(list[0])
		if err != nil {
			t.Error("[GetUserByUserIDList] err: ", err.Error())
		}
		t.Log(string(rest))
	}
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
