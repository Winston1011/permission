package api_test

import (
	"net/http/httptest"
	"os"
	"testing"

	"permission/api"
	"permission/helpers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"permission/pkg/golib/v2/env"
)

// post form
func TestGetUserInfoList(t *testing.T) {
	count, infoList, err := api.GetUserList(ctx)
	if err != nil {
		t.Errorf("[TestGetUserInfoList] error = %s", err.Error())
		return
	}
	t.Logf("[TestGetUserInfoList] total num = %+v", count)

	for _, info := range infoList {
		t.Logf("[TestGetUserInfoList] user:  %+v", info)
	}

	t.Log("success!")
}

// post json
func TestBuyCourse(t *testing.T) {
	orderID, err := api.BuyCourse(ctx, "7788", "3306")
	if err != nil {
		t.Errorf("[BuyCourse] error = %s", err.Error())
		return
	}
	t.Logf("[BuyCourse] orderID = %s", orderID)

	assert.Equal(t, "7788", orderID)
}

// post raw data
func TestBuyCourseRaw(t *testing.T) {
	orderID, err := api.BuyCourseRaw(ctx, "7788", "3306")
	if err != nil {
		t.Errorf("[TestBuyCourseRaw] error = %s", err.Error())
		return
	}
	t.Logf("[TestBuyCourseRaw] orderID = %s", orderID)

	assert.Equal(t, "7788", orderID)
}

// get
func TestGetCourseInfo(t *testing.T) {
	info, err := api.GetCourseInfo(ctx, "6379")
	if err != nil {
		t.Errorf("[TestGetCourseInfo] error = %s", err.Error())
		return
	}
	t.Logf("[TestGetCourseInfo] data = %+v", info)

	assert.Equal(t, "6379", info["courseID"])
}

// post mcPack
func TestGetUserCourseV2(t *testing.T) {
	info, err := api.GetUserCourse(ctx, "6379")
	if err != nil {
		t.Errorf("[GetCourseInfo] error = %s", err.Error())
		return
	}
	t.Logf("[GetCourseInfo] data = %+v", info)

	assert.Equal(t, "6379", info["userID"])
}

var ctx *gin.Context

func TestMain(m *testing.M) {
	w := httptest.NewRecorder()
	ctx, _ = gin.CreateTestContext(w)

	env.SetRootPath("../")
	helpers.PreInit()
	m.Run()
	os.Exit(0)
}
