package user

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"permission/helpers"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"permission/pkg/golib/v2/base"
	"permission/pkg/golib/v2/env"
)

func TestGetUserInfo(t *testing.T) {
	testCase := "/permission/api/user/getlist"

	// 构造请求
	form := url.Values{
		"userIDList": []string{"11,22,33,44,55"},
	}
	req := httptest.NewRequest(
		http.MethodPost,
		testCase,
		ioutil.NopCloser(strings.NewReader(form.Encode())),
	)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// 搭建路由
	router := gin.New()
	router.POST(testCase, GetUserInfoList)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 结果验证
	assert.Equal(t, http.StatusOK, w.Code)

	res := base.DefaultRender{}
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Errorf("testCase: %s get response body error: %s", testCase, err.Error())
	}
	assert.Equal(t, res.ErrNo, 0)

	t.Logf("testCase: %s , result is: %v", testCase, res.Data)
}

func TestMain(m *testing.M) {
	env.SetRootPath("../../../")

	helpers.PreInit()
	helpers.InitResource(nil)
	m.Run()
	os.Exit(0)
}
