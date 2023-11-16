package antispam

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gin-gonic/gin"
	"permission/pkg/golib/v2/base"
	"permission/pkg/golib/v2/zlog"
)

func TestAppCheck(t *testing.T) {
	base.InitHttp(nil)
	zlog.InitLog(zlog.LogConfig{})
	InitAntiSpam()
	p := struct {
		AppId string
		Cuid  string
	}{
		AppId: "iot_app",
		Cuid:  "955C675F1FC148F19C9183E4118DABEB",
	}
	b, _ := json.Marshal(p)
	newR := httptest.NewRequest("POST", "http://pluto.epochdz.com/antispam-server/gettoken", bytes.NewBuffer(b))
	newR.PostForm = make(url.Values)
	newR.Form = make(url.Values)
	newR.PostForm["appId"] = []string{"iot_app"}
	newR.PostForm["cuid"] = []string{"955C675F1FC148F19C9183E4118DABEB"}
	//newR.PostForm["_t_"] = []string{"1635848405"}
	//newR.Form["_t_"] = []string{"1635848405"}
	newR.Form["appId"] = []string{"iot_app"}
	newR.Form["cuid"] = []string{"955C675F1FC148F19C9183E4118DABEB"}
	newR.Form["sign"] = []string{"095c49c91e7b9b3a190a74c43e616f6a"}
	ctx := &gin.Context{Request: newR}
	if err := SdkCheck(ctx); err != nil {
		t.Error("[TestAppCheck] fail", err)
		return
	}
	t.Log("success")
}
