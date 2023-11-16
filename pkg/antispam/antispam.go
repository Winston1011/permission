package antispam

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"permission/pkg/antispam/logger"
	"permission/pkg/antispam/utils"
	"permission/pkg/golib/v2/base"
)

const (
	OSIos          = "ios"
	OSAndroid      = "android"
	OSSdk          = "sdk"
	serverURI      = "/antispam-server/gettoken"
	zYBClusterType = "ZYB_CLUSTER_TYPE"
)

type Config struct {
	Common struct {
		TimeControl struct {
			Switch  int `yaml:"switch"`
			MaxTime int `yaml:"maxTime"`
		} `yaml:"timeControl"`
	} `yaml:"common"`

	Keystore  map[string]string `yaml:"keystore"`
	Logger    logger.Interface
	ServerAPI *base.ApiClient
}

type Request struct {
	OS         string // 设备操作系统
	CuID       string // 设备cuid
	AppID      string
	path       string
	sign       string
	clientTime string
	SignParam  map[string]string
}

var (
	config  *Config
	zLogger logger.Interface
)

type Option func(c *Config)

func InitAntiSpam(options ...Option) {
	domain := "http://antispam-server-svc.inf:8080"
	if os.Getenv(zYBClusterType) == "" {
		domain = "http://pluto-base-e.suanshubang.cc"
	}

	var conf = &Config{
		ServerAPI: &base.ApiClient{
			Service: "antispam-server",
			Domain:  domain,
		},
		Logger:   logger.Default,
		Keystore: map[string]string{"key3": "8&%d*"},
	}
	for _, opt := range options {
		opt(conf)
	}
	zLogger = conf.Logger
	config = conf
}

func initRequestParam(ctx *gin.Context) (r *Request, err error) {
	err = ctx.Request.ParseForm()
	if err != nil {
		zLogger.Logger(ctx, logger.Error, "parse params err: "+err.Error())
		return nil, ErrorParseParams
	}

	params := make(map[string]string)
	scuID := ""
	for k, vs := range ctx.Request.Form {
		if len(vs) <= 0 {
			continue
		}

		if k == "__scuid" {
			// saf传递签名的cuid，替换签名参数中的cuid
			scuID = vs[0]
		} else if k == "sign" {
			// 不存储
		} else {
			params[k] = vs[0]
		}
	}
	if scuID != "" {
		// 覆盖签名参数中的cuid
		params["cuid"] = scuID
	}

	r = &Request{
		AppID:     getAppID(ctx),
		path:      ctx.Request.URL.Path,
		SignParam: params,
	}

	cuID := ctx.Request.Form.Get("cuid")
	if cuID == "" {
		zLogger.Logger(ctx, logger.Warn, "[getRandomKey] cuid is empty!")
		return nil, ErrorCuIDEmpty
	}
	r.CuID = cuID

	_os := ctx.Request.Form.Get("os")
	if _os == "" {
		_os = OSAndroid
	}
	r.OS = _os

	r.clientTime = ctx.Request.Form.Get("_t_")

	sign := ctx.Request.Form.Get("sign")
	if sign == "" {
		zLogger.Logger(ctx, logger.Warn, "params lack sign key word!")
		return nil, ErrorAntiSpamLackSign
	}

	r.sign = sign

	return r, nil
}

func getAppID(ctx *gin.Context) (appID string) {
	appID = ctx.Request.Form.Get("appId")
	if appID == "" {
		// 兼容不规范的业务获取id方式，可根据自己业务线传递appID方式去掉兼容的逻辑
		appID = ctx.Request.Form.Get("appid")
	}

	if appID != "" {
		return appID
	}

	return "homework"
}

func AppCheck(ctx *gin.Context) (err error) {
	r, err := initRequestParam(ctx)
	if err != nil {
		return err
	}

	// 签名校验
	randomKey, err := r.getRandomKey(ctx)
	// 出现未知错误，保持放行
	if err == ErrorIgnore {
		return nil
	}
	if err != nil {
		return err
	}

	if randomKey == "" {
		zLogger.Logger(ctx, logger.Warn, "[AppCheck] empty getRandomKey!")
		return ErrorEmptyToken
	}

	if err = r.signVerify(ctx, randomKey); err != nil {
		return err
	}

	if err = r.timeControl(ctx); err != nil {
		return err
	}

	return nil
}

func SdkCheck(ctx *gin.Context) (err error) {
	r, err := initRequestParam(ctx)
	if err != nil {
		return err
	}

	// 签名校验
	randomKey, err := r.getRandomKey(ctx)
	if err == ErrorIgnore {
		return nil
	}
	if err != nil {
		return err
	}

	if randomKey == "" {
		// 返回特定错误码错误
		zLogger.Logger(ctx, logger.Warn, "[SdkCheck] empty getRandomKey!")
		return ErrorEmptyToken
	}

	if err = r.signVerify(ctx, randomKey); err != nil {
		return err
	}

	if err = r.timeControl(ctx); err != nil {
		return err
	}

	return nil
}

func (r *Request) getRandomKey(ctx *gin.Context) (randomKey string, err error) {
	expireTime := 0
	randomKey, expireTime, err = r.getRandomToken(ctx)
	if err != nil {
		zLogger.Logger(ctx, logger.Warn, "GetRandomToken error: "+err.Error())
		return randomKey, err
	}

	now := int(time.Now().Unix())
	if expireTime-now < 10 {
		// 过期时间接近临界时间，直接判断过期
		zLogger.Logger(ctx, logger.Warn, "expire time is "+strconv.Itoa(expireTime)+" is closing to now: "+strconv.Itoa(now))
		return randomKey, ErrorTokenNearlyExpired
	}
	return randomKey, nil
}

// 计算并校验签名
func (r *Request) signVerify(ctx *gin.Context, randomKey string) error {
	// kSort
	var keys []string
	for k := range r.SignParam {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 拼接
	var signStr string
	for _, k := range keys {
		signStr += k + "=" + r.SignParam[k]
	}
	zLogger.Logger(ctx, logger.Debug, "signStr is: "+signStr)

	if r.OS == OSAndroid {
		prefix := config.Keystore["key3"]
		signStr = fmt.Sprintf("%s[%s]@%s", prefix, utils.Md5(randomKey), utils.Base64Encode(signStr))
	} else {
		str := "K$L@aPb$O^Ic%U*Y`T=f+R~d954e1215aef11a512c1585a0fcd5648ff189f1e?Q0\"9{8<7@6#5(4%3&2+1"

		key := utils.Md5(str + r.CuID + randomKey)
		randomKey, err := utils.Rc4Encode(key, randomKey)
		if err != nil {
			zLogger.Logger(ctx, logger.Error, "token rc4 error: "+err.Error())
			return ErrorTokenRc4
		}

		randomKey = utils.Base64Encode(randomKey)
		signStr = fmt.Sprintf("[%s]@%s", randomKey, utils.Base64Encode(signStr))
	}

	signHash := utils.Md5(signStr)

	if signHash != r.sign {
		zLogger.Logger(ctx, logger.Warn, "user input sign is: "+r.sign+" signHash is: "+signHash)
		return ErrorAntiSpamSignErr
	}

	return nil
}

func (r *Request) timeControl(ctx *gin.Context) error {
	if config.Common.TimeControl.Switch == 0 {
		return nil
	}

	t, err := strconv.Atoi(r.clientTime)
	if err != nil {
		zLogger.Logger(ctx, logger.Warn, "client time format error! _t_: "+r.clientTime)
		return ErrorClientTimeFormat
	}

	now := int(time.Now().Unix())
	maxTime := config.Common.TimeControl.MaxTime
	if now-t > maxTime {
		zLogger.Logger(ctx, logger.Warn, "currentTime is "+strconv.Itoa(now)+", clientTime is "+strconv.Itoa(t)+", diff greater than configMaxTime: "+strconv.Itoa(maxTime))
		return ErrorClientTimeTooOld
	}
	return nil
}

// 通过设备cuid获取设备注册反作弊信息
func (r *Request) getRandomToken(ctx *gin.Context) (randomKey string, expire int, err error) {
	requestBody := map[string]string{
		"appId": r.AppID,
		"cuid":  r.CuID,
	}
	opt := base.HttpRequestOptions{
		RequestBody: requestBody,
	}
	resp, err := config.ServerAPI.HttpPost(ctx, serverURI, opt)
	if err != nil {
		zLogger.Logger(ctx, logger.Error, "get token err: "+err.Error())
		return randomKey, expire, ErrorIgnore
	}

	info := struct {
		ErrNo  int    `json:"errNo"`
		ErrStr string `json:"errstr"`
		Data   struct {
			RandomKey  string `json:"randomKey"`
			ExpireTime int    `json:"expireTime"`
		} `json:"data"`
	}{}
	err = json.Unmarshal(resp.Response, &info)
	if err != nil {
		zLogger.Logger(ctx, logger.Error, "get token json unmarshal err: "+err.Error())
		return randomKey, expire, ErrorIgnore
	}
	if info.ErrNo != 0 {
		zLogger.Logger(ctx, logger.Error, "get token from antispam-server, err message is: "+info.ErrStr+"error no is"+strconv.Itoa(info.ErrNo))
		return randomKey, expire, ErrorGetToken
	}
	return info.Data.RandomKey, info.Data.ExpireTime, nil
}
