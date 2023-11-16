package nmq

// golang nmq边车依赖
import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	json "github.com/json-iterator/go"
	"permission/pkg/golib/v2/base"
	"permission/pkg/golib/v2/gomcpack/mcpack"
	"permission/pkg/golib/v2/utils"
	"permission/pkg/golib/v2/zlog"
)

const sockPath = "/usr/local/var/run/odpproxy.sock"
const reqPath = "/api/mcRalRequest"
const serviceName = "nmq-producer"

const (
	packKeyProduct   = "_product"
	packKeyTopic     = "_topic"
	packKeyCmd       = "_cmd"
	packKeyCallerURI = "_caller_uri"
	PackKeyErrorNo   = "_error_no"
	PackKeyErrorMsg  = "_error_msg"
	PackKeyTransID   = "_transid"
)

type sidecarResponse struct {
	ErrNo      int    `json:"errNo"`
	ErrMsg     string `json:"errMsg"`
	ErrUserMsg string `json:"-"`
	Data       struct {
		Data struct {
			ErrNo   int    `json:"_error_no"`
			ErrMsg  string `json:"_error_msg"`
			TransID int    `json:"_transid"`
		}
	} `json:"data"`
}

var client *base.ApiClient
var once = &sync.Once{}

func SendCmd(ctx *gin.Context, cmd int64, topic string, product string, data map[string]interface{}) (resp NmqResponse, err error) {
	once.Do(initProducer)

	var logID uint32
	if zlog.GetLogID(ctx) != "" {
		var logIDParsed uint64
		logIDParsed, err := strconv.ParseUint(zlog.GetLogID(ctx), 10, 32)
		if err != nil {
			logIDParsed = 0
		}
		logID = uint32(logIDParsed)
	}

	var patchCallerURI string
	if value, ok := data[packKeyCallerURI]; ok {
		patchCallerURI, _ = value.(string)
	}
	callerURI, _ := utils.GetPressureFlag(ctx)
	if callerURI == "" {
		// patch: 一个兜底，避免业务不合理的使用了ctx，导致从ctx中获取callerURI失败情况
		if patchCallerURI != "" {
			callerURI = patchCallerURI
		} else {
			callerURI = "\x00"
		}
	}

	data[packKeyProduct] = product
	data[packKeyTopic] = topic
	data[packKeyCmd] = strconv.FormatInt(cmd, 10)
	data[packKeyCallerURI] = callerURI
	//data[packKeyLogID] = logID
	body, err := mcpack.Marshal(data)
	if err != nil {
		return resp, fmt.Errorf("msg encode fail, err: %s", err.Error())
	}

	header := make(map[string]uint32)
	header["log_id"] = logID
	header["body_len"] = uint32(len(body))
	headerStr, _ := json.Marshal(header)

	formData := make(map[string]interface{})
	formData["serviceName"] = serviceName
	formData["method"] = "POST"
	formData["path"] = reqPath
	formData["header"] = string(headerStr)
	formData["body"] = string(body)

	opt := base.HttpRequestOptions{
		RequestBody: formData,
		Encode:      base.EncodeForm,
	}
	ret, err := client.HttpPost(ctx, reqPath, opt)
	if err != nil {
		resp.ErrNo = -1
		resp.ErrStr = err.Error()
		return resp, err
	}

	sidecarResp := sidecarResponse{}
	err = json.Unmarshal(ret.Response, &sidecarResp)
	if err != nil {
		resp.ErrNo = -1
		resp.ErrStr = "response json decode error, err: " + err.Error()
		return resp, err
	}

	resp.ErrNo = sidecarResp.Data.Data.ErrNo
	resp.ErrStr = sidecarResp.Data.Data.ErrMsg
	resp.TransID = uint64(sidecarResp.Data.Data.TransID)
	if resp.ErrNo != 0 {
		return resp, errors.New(resp.ErrStr)
	}
	return resp, nil
}

func initProducer() {
	if client != nil && client.Service != "" {
		return
	}

	client = &base.ApiClient{
		Service:         serviceName,
		Domain:          "http://unix",
		Timeout:         5 * time.Second,
		ConnectTimeout:  100 * time.Millisecond,
		Retry:           1,
		IdleConnTimeout: 0,
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
					return net.Dial("unix", sockPath)
				},
				MaxConnsPerHost:     100,
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     0,
			},
		},
	}

	return
}
