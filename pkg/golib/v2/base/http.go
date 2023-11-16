package base

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	json "github.com/json-iterator/go"
	"golang.org/x/net/http/httpguts"
	"permission/pkg/golib/v2/env"
	"permission/pkg/golib/v2/gomcpack/mcpack"
	"permission/pkg/golib/v2/utils"
	"permission/pkg/golib/v2/zlog"
)

const HttpHeaderService = "SERVICE"

const (
	EncodeJson    = "_json"
	EncodeForm    = "_form"
	EncodeMcPack  = "_mcPack"
	EncodeRaw     = "_raw"
	EncodeRawByte = "_raw_byte"
)

type HttpRequestOptions struct {
	// 通用请求体，可通过Encode来对body做编码
	RequestBody interface{}
	// deprecated , use RequestBody instead
	// 老的请求data，httpGet / httPost 仍支持
	Data map[string]string
	// deprecated , use RequestBody instead
	// 针对 httpPostJson 接口的特定body
	JsonBody interface{}

	// 请求头指定
	Headers map[string]string
	// cookie 设定
	Cookies map[string]string

	/*
		httpGet / httPost 默认 application/x-www-form-urlencoded
		httpPostJson 默认 application/json
	*/
	ContentType string
	// 针对 RequestBody 的编码
	Encode string

	// 接口请求级timeout。不管retry是多少，那么每次执行的总时间都是timeout。
	// 这个timeout与client.Timeout 没有直接关系，总执行超时时间取二者最小值。
	Timeout time.Duration

	// 重试策略，可不指定，默认使用`defaultRetryPolicy`(只有在`api.yaml`中指定retry>0 时生效)
	RetryPolicy RetryPolicy `json:"-"`
}

func (o *HttpRequestOptions) getData() ([]byte, error) {
	if len(o.Data) > 0 {
		// Data 只支持form编码方式
		v := url.Values{}
		for key, value := range o.Data {
			v.Add(key, value)
		}
		return utils.StringToBytes(v.Encode()), nil
	}

	if o.JsonBody != nil {
		reqBody, err := json.Marshal(o.JsonBody)
		return reqBody, err
	}

	if o.RequestBody == nil {
		return nil, nil
	}

	switch o.Encode {
	case EncodeJson:
		reqBody, err := json.Marshal(o.RequestBody)
		return reqBody, err
	case EncodeMcPack:
		reqBody, err := mcpack.Marshal(o.RequestBody)
		return reqBody, err
	case EncodeRaw:
		var err error
		encodeData, ok := o.RequestBody.(string)
		if !ok {
			err = errors.New("EncodeRaw need string type")
		}
		return utils.StringToBytes(encodeData), err
	case EncodeRawByte:
		var err error
		encodeData, ok := o.RequestBody.([]byte)
		if !ok {
			err = errors.New("EncodeRawByte need []byte type")
		}
		return encodeData, err
	case EncodeForm: // 由于历史原因，默认Form编码方式
		fallthrough
	default:
		encodeData, err := o.getFormRequestData()
		return utils.StringToBytes(encodeData), err
	}
}
func (o *HttpRequestOptions) getFormRequestData() (string, error) {
	v := url.Values{}

	if data, ok := o.RequestBody.(map[string]string); ok {
		for key, value := range data {
			v.Add(key, value)
		}
		return v.Encode(), nil
	}

	if data, ok := o.RequestBody.(map[string]interface{}); ok {
		for key, value := range data {
			var vStr string
			switch value.(type) {
			case string:
				vStr = value.(string)
			default:
				if tmp, err := json.Marshal(value); err != nil {
					return "", err
				} else {
					vStr = string(tmp)
				}
			}

			v.Add(key, vStr)
		}
		return v.Encode(), nil
	}

	return "", errors.New("unSupport RequestBody type")
}
func (o *HttpRequestOptions) GetContentType() (cType string) {
	if cType = o.ContentType; cType != "" {
		return cType
	}

	// 根据 encode 获得一个默认的类型
	switch o.Encode {
	case EncodeJson:
		cType = "application/json"
	case EncodeMcPack:
		fallthrough
	case EncodeForm: // 由于历史原因，默认Form编码方式
		fallthrough
	default:
		cType = "application/x-www-form-urlencoded"
	}
	return cType
}

const (
	_defaultPrintRequestLen  = 10240
	_defaultPrintResponseLen = 10240
)

type ApiClient struct {
	Service         string        `yaml:"service"`
	AppKey          string        `yaml:"appkey"`
	AppSecret       string        `yaml:"appsecret"`
	Domain          string        `yaml:"domain"`
	Timeout         time.Duration `yaml:"timeout"`
	ConnectTimeout  time.Duration `yaml:"connectTimeout"`
	Retry           int           `yaml:"retry"`
	HttpStat        bool          `yaml:"httpStat"`
	Host            string        `yaml:"host"`
	Proxy           string        `yaml:"proxy"`
	MaxIdleConns    int           `yaml:"maxIdleConns"`
	IdleConnTimeout time.Duration `yaml:"idleConnTimeout"`
	BasicAuth       struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	}

	// request body 最大长度展示，0表示采用默认的10240，-1表示不打印
	MaxReqBodyLen int `yaml:"maxReqBodyLen"`
	// response body 最大长度展示，0表示采用默认的10240，-1表示不打印。指定长度的时候需注意，返回的json可能被截断
	MaxRespBodyLen int `yaml:"maxRespBodyLen"`

	// 配置中设置了该值后当 err!=nil || httpCode >= retryHttpCode 时会重试（该策略优先级最低）
	RetryHttpCode int `yaml:"retryHttpCode"`

	// 重试策略，可不指定，默认使用`defaultRetryPolicy`(只有在`api.yaml`中指定retry>0 时生效)
	retryPolicy RetryPolicy  `json:"-"`
	HTTPClient  *http.Client `json:"-"`
	clientInit  sync.Once    `json:"-"`
}

func (client *ApiClient) SetRetryPolicy(retry RetryPolicy) {
	client.retryPolicy = retry
}

func (client *ApiClient) GetTransPort() *http.Transport {
	connectTimeout := 3 * time.Second
	if client.ConnectTimeout != 0 {
		connectTimeout = client.ConnectTimeout
	}

	// 兼容之前全局transport的逻辑，每个host最大空闲连接为拆开后的transport的最大空闲连接
	maxIdleConns := 100
	if client.MaxIdleConns != 0 {
		maxIdleConns = client.MaxIdleConns
	} else if globalTransport != nil && globalTransport.MaxIdleConnsPerHost != 0 {
		maxIdleConns = globalTransport.MaxIdleConnsPerHost
	}

	idleConnTimeout := 300 * time.Second
	if client.IdleConnTimeout != 0 {
		idleConnTimeout = client.IdleConnTimeout
	} else if globalTransport != nil && globalTransport.IdleConnTimeout != 0 {
		idleConnTimeout = globalTransport.IdleConnTimeout
	}

	trans := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: connectTimeout,
		}).DialContext,

		MaxIdleConns:        maxIdleConns,
		MaxIdleConnsPerHost: maxIdleConns,
		IdleConnTimeout:     idleConnTimeout,
	}

	if client.Proxy != "" {
		u, err := url.Parse(client.Proxy)
		if err != nil {
			zlog.Errorf(nil, "client proxy URL format error")
		}
		trans.Proxy = http.ProxyURL(u)
	} else {
		trans.Proxy = http.ProxyFromEnvironment
	}

	return trans
}

func (client *ApiClient) makeRequest(ctx *gin.Context, method, path string, opts HttpRequestOptions) (reqBody []byte, req *http.Request, err error) {
	urlData, err := opts.getData()
	if err != nil {
		zlog.WarnLogger(ctx, "http client make data error: "+err.Error(), zlog.String(zlog.TopicType, zlog.LogNameModule))
		return nil, nil, err
	}

	reqURL := client.Domain + path
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodDelete:
		if strings.Contains(reqURL, "?") {
			reqURL = reqURL + "&" + string(urlData)
		} else {
			reqURL = reqURL + "?" + string(urlData)
		}

	case http.MethodPost, http.MethodPut:
		reqBody = urlData
	}

	req, err = http.NewRequest(method, reqURL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, nil, err
	}

	if opts.Headers != nil {
		for k, v := range opts.Headers {
			req.Header.Set(k, v)
		}
	}

	if client.Host != "" {
		req.Host = client.Host
	} else if h := req.Header.Get("host"); h != "" {
		req.Host = h
	}

	for k, v := range opts.Cookies {
		req.AddCookie(&http.Cookie{
			Name:  k,
			Value: v,
		})
	}

	if client.BasicAuth.Username != "" {
		req.SetBasicAuth(client.BasicAuth.Username, client.BasicAuth.Password)
	}

	req.Header.Set("Content-Type", opts.GetContentType())

	req.Header.Set(zlog.TraceHeaderKey, zlog.GetRequestID(ctx))

	req.Header.Set(HttpHeaderService, env.AppName)
	req.Header.Set(zlog.LogIDHeaderKey, zlog.GetLogID(ctx))
	req.Header.Set(zlog.LogIDHeaderKeyLower, zlog.GetLogID(ctx))

	// 全链路压测标记，如果有且合法就透传
	callerURI, _ := utils.GetPressureFlag(ctx)
	if callerURI != "" && httpguts.ValidHeaderFieldValue(callerURI) {
		req.Header.Set(utils.HttpXBDCallerURI, callerURI)
	}

	// 通用header匹配，符合条件就透传
	for k, v := range utils.GetTransportHeader(ctx) {
		req.Header.Set(k, v)
	}
	return reqBody, req, nil
}

type ApiResult struct {
	HttpCode int
	Response []byte
	Header   http.Header
	Ctx      *gin.Context
}

func (client *ApiClient) HttpGet(ctx *gin.Context, path string, opts HttpRequestOptions) (*ApiResult, error) {
	return client.HttpDo(ctx, http.MethodGet, path, opts)
}

func (client *ApiClient) HttpPost(ctx *gin.Context, path string, opts HttpRequestOptions) (*ApiResult, error) {
	return client.HttpDo(ctx, http.MethodPost, path, opts)
}

// deprecated , use HttpPost instead
func (client *ApiClient) HttpPostJson(ctx *gin.Context, path string, opts HttpRequestOptions) (*ApiResult, error) {
	opts.ContentType = "application/json"
	return client.HttpDo(ctx, "POST", path, opts)
}

func (client *ApiClient) HttpDo(ctx *gin.Context, method, path string, opts HttpRequestOptions) (*ApiResult, error) {
	reqBody, req, rErr := client.makeRequest(ctx, method, path, opts)
	if rErr != nil {
		zlog.WarnLogger(ctx, "http client makeRequest error: "+rErr.Error(), zlog.String(zlog.TopicType, zlog.LogNameModule))
		return nil, rErr
	}

	start := time.Now()

	t := client.beforeHttpStat(ctx, req)
	resp, fields, err := client.do(ctx, req, &opts)
	client.afterHttpStat(ctx, req.URL.Scheme, t)

	res := ApiResult{Ctx: ctx}
	msg := "http request success"
	var bodySize int64
	if err != nil {
		msg = err.Error()
	} else if resp != nil {
		res.HttpCode = resp.StatusCode
		res.Header = resp.Header
		bodySize = resp.ContentLength

		switch resp.Header.Get("Content-Encoding") {
		case "gzip":
			gzipReader, err := gzip.NewReader(resp.Body)
			if err == nil {
				res.Response, err = io.ReadAll(gzipReader)
				_ = gzipReader.Close()
			}
		default:
			res.Response, err = io.ReadAll(resp.Body)
		}

		_ = resp.Body.Close()
	}

	reqData, respData := client.formatLogMsg(reqBody, res.Response)

	end := time.Now()

	zlog.DebugLogger(ctx, "http "+method+" request",
		zlog.String(zlog.TopicType, zlog.LogNameModule),
		zlog.String("url", req.URL.String()),
		zlog.ByteString("params", reqData),
		zlog.Int("responseCode", res.HttpCode),
		zlog.ByteString("responseBody", respData),
	)

	fields = append(fields,
		zlog.String("requestStartTime", utils.GetFormatRequestTime(start)),
		zlog.String("requestEndTime", utils.GetFormatRequestTime(end)),
		zlog.Float64("cost", utils.GetRequestCost(start, end)),
		zlog.Int64("bodySize", bodySize),
	)

	zlog.InfoLogger(ctx, msg, fields...)
	return &res, err
}

// 建议仅当流式读取response时，使用该方法。必须要记得调用 resp.Body.Close() !!!!!
func (client *ApiClient) HttpStream(ctx *gin.Context, method, path string, opts HttpRequestOptions) (*http.Response, error) {
	reqBody, req, rErr := client.makeRequest(ctx, method, path, opts)
	if rErr != nil {
		zlog.WarnLogger(ctx, "http client makeRequest error: "+rErr.Error(), zlog.String(zlog.TopicType, zlog.LogNameModule))
		return nil, rErr
	}

	start := time.Now()

	t := client.beforeHttpStat(ctx, req)
	resp, fields, err := client.do(ctx, req, &opts)
	client.afterHttpStat(ctx, req.URL.Scheme, t)

	msg := "http request success"
	if err != nil {
		msg = err.Error()
	}

	reqData, respData := client.formatLogMsg(reqBody, nil)

	end := time.Now()

	zlog.DebugLogger(ctx, "http "+method+" request",
		zlog.String(zlog.TopicType, zlog.LogNameModule),
		zlog.String("url", req.URL.String()),
		zlog.ByteString("params", reqData),
		zlog.ByteString("responseBody", respData),
	)

	fields = append(fields,
		zlog.String("requestStartTime", utils.GetFormatRequestTime(start)),
		zlog.String("requestEndTime", utils.GetFormatRequestTime(end)),
		zlog.Float64("cost", utils.GetRequestCost(start, end)),
	)
	zlog.InfoLogger(ctx, msg, fields...)
	return resp, err
}

func (client *ApiClient) GetRetryPolicy(opts *HttpRequestOptions) (retryPolicy RetryPolicy) {
	if opts != nil && opts.RetryPolicy != nil {
		// 接口维度超时策略
		retryPolicy = opts.RetryPolicy
	} else if client.retryPolicy != nil {
		// client维度超时策略(代码中指定的)
		retryPolicy = client.retryPolicy
	} else if client.RetryHttpCode > 0 {
		// 配置中指定的
		retryPolicy = func(resp *http.Response, err error) bool {
			return err != nil || resp == nil || resp.StatusCode >= client.RetryHttpCode
		}
	} else {
		// 默认超时策略
		retryPolicy = defaultRetryPolicy
	}
	return retryPolicy
}

func (client *ApiClient) do(ctx *gin.Context, req *http.Request, opts *HttpRequestOptions) (resp *http.Response, field []zlog.Field, err error) {
	// 一次请求的timeout 选择策略：
	// 1. 代码优先，如果代码中传了timeout，则会忽略配置中的timeout
	// 2. 如果代码中没指定，则会选择配置中的timeout
	// 3. 如果配置中也没有指定，则默认3s
	timeout := 3 * time.Second
	if opts != nil && opts.Timeout > 0 {
		timeout = opts.Timeout
	} else if client.Timeout > 0 {
		timeout = client.Timeout
	}

	fields := []zlog.Field{
		zlog.String(zlog.TopicType, zlog.LogNameModule),
		zlog.String("prot", "http"),
		zlog.String("service", client.Service),
		zlog.String("method", req.Method),
		zlog.String("domain", client.Domain),
		zlog.String("requestUri", req.URL.Path),
		zlog.Duration("timeout", timeout),
	}

	client.clientInit.Do(func() {
		if client.HTTPClient == nil {
			client.HTTPClient = &http.Client{
				Transport: client.GetTransPort(),
			}
		}
	})

	var (
		dataBuffer   *bytes.Reader
		maxAttempts  int
		attemptCount int
		doErr        error
		shouldRetry  bool
	)

	attemptCount, maxAttempts = 0, client.Retry

	// 策略选择优先级：option > client > default
	retryPolicy := client.GetRetryPolicy(opts)

	for {
		if req.GetBody != nil {
			bodyReadCloser, _ := req.GetBody()
			req.Body = bodyReadCloser
		} else if req.Body != nil {
			if dataBuffer == nil {
				data, err := io.ReadAll(req.Body)
				_ = req.Body.Close()
				if err != nil {
					return nil, fields, err
				}
				dataBuffer = bytes.NewReader(data)
				req.ContentLength = int64(dataBuffer.Len())
				req.Body = io.NopCloser(dataBuffer)
			}
			_, _ = dataBuffer.Seek(0, io.SeekStart)
		}

		attemptCount++

		c, _ := context.WithTimeout(context.Background(), timeout)
		req = req.WithContext(c)
		resp, doErr = client.HTTPClient.Do(req)

		shouldRetry = retryPolicy(resp, doErr)
		if !shouldRetry {
			break
		}

		msg := "hit retry policy attemptCount: " + strconv.Itoa(attemptCount)
		if doErr != nil {
			msg += ", error: " + doErr.Error()
		}
		zlog.WarnLogger(ctx, msg, fields...)

		if attemptCount > maxAttempts {
			break
		}

		// 符合retry条件...
		if doErr == nil {
			drainAndCloseBody(resp, 16384)
		}
	}

	if shouldRetry {
		msg := "hit retry policy"
		if doErr != nil {
			msg += ", error: " + doErr.Error()
		}
		err = fmt.Errorf("giving up after %d attempt(s): %s", attemptCount, msg)
	}

	httpCode := 0
	if resp != nil {
		httpCode = resp.StatusCode
	}
	fields = append(fields,
		zlog.String("retry", fmt.Sprintf("%d/%d", attemptCount-1, client.Retry)),
		zlog.Int("httpCode", httpCode),
		zlog.Int("ralCode", client.calRalCode(resp, err)),
	)

	return resp, fields, err
}

func (client *ApiClient) formatLogMsg(requestParam, responseData []byte) (req, resp []byte) {
	maxReqBodyLen := client.MaxReqBodyLen
	if maxReqBodyLen == 0 {
		maxReqBodyLen = _defaultPrintRequestLen
	}

	maxRespBodyLen := client.MaxRespBodyLen
	if maxRespBodyLen == 0 {
		maxRespBodyLen = _defaultPrintResponseLen
	}

	if maxReqBodyLen != -1 {
		req = requestParam
		if len(requestParam) > maxReqBodyLen {
			req = req[:maxReqBodyLen]
		}
	}

	if maxRespBodyLen != -1 {
		resp = responseData
		if len(responseData) > maxRespBodyLen {
			resp = resp[:maxRespBodyLen]
		}
	}

	return req, resp
}

// 本次请求正确性判断
func (client *ApiClient) calRalCode(resp *http.Response, err error) int {
	if err != nil || resp == nil || resp.StatusCode >= 400 || resp.StatusCode == 0 {
		return -1
	}
	return 0
}

type timeTrace struct {
	dnsStartTime,
	dnsDoneTime,
	connectStartTime,
	connectDoneTime,
	tlsHandshakeStartTime,
	tlsHandshakeDoneTime,
	getConnTime,
	gotConnTime,
	gotFirstRespTime,
	finishTime time.Time
}

func (client *ApiClient) beforeHttpStat(ctx *gin.Context, req *http.Request) *timeTrace {
	if client.HttpStat == false {
		return nil
	}

	var t = &timeTrace{}
	trace := &httptrace.ClientTrace{
		// before get a connection
		GetConn:  func(_ string) { t.getConnTime = time.Now() },
		DNSStart: func(_ httptrace.DNSStartInfo) { t.dnsStartTime = time.Now() },
		DNSDone:  func(_ httptrace.DNSDoneInfo) { t.dnsDoneTime = time.Now() },
		// before a new connection
		ConnectStart: func(_, _ string) { t.connectStartTime = time.Now() },
		// after a new connection
		ConnectDone: func(net, addr string, err error) { t.connectDoneTime = time.Now() },
		// after get a connection
		GotConn:              func(_ httptrace.GotConnInfo) { t.gotConnTime = time.Now() },
		GotFirstResponseByte: func() { t.gotFirstRespTime = time.Now() },
		TLSHandshakeStart:    func() { t.tlsHandshakeStartTime = time.Now() },
		TLSHandshakeDone:     func(_ tls.ConnectionState, _ error) { t.tlsHandshakeDoneTime = time.Now() },
	}
	*req = *req.WithContext(httptrace.WithClientTrace(context.Background(), trace))
	return t
}

func (client *ApiClient) afterHttpStat(ctx *gin.Context, scheme string, t *timeTrace) {
	if client.HttpStat == false {
		return
	}
	t.finishTime = time.Now() // after read body

	cost := func(d time.Duration) float64 {
		if d < 0 {
			return -1
		}
		return float64(d.Nanoseconds()/1e4) / 100.0
	}

	serverProcessDuration := t.gotFirstRespTime.Sub(t.gotConnTime)
	contentTransDuration := t.finishTime.Sub(t.gotFirstRespTime)
	if t.gotConnTime.IsZero() {
		// 没有拿到连接的情况
		serverProcessDuration = 0
		contentTransDuration = 0
	}

	switch scheme {
	case "https":
		f := []zlog.Field{
			zlog.String(zlog.TopicType, zlog.LogNameModule),
			zlog.Float64("dnsLookupCost", cost(t.dnsDoneTime.Sub(t.dnsStartTime))),                      // dns lookup
			zlog.Float64("tcpConnectCost", cost(t.connectDoneTime.Sub(t.connectStartTime))),             // tcp connection
			zlog.Float64("tlsHandshakeCost", cost(t.tlsHandshakeDoneTime.Sub(t.tlsHandshakeStartTime))), // tls handshake
			zlog.Float64("serverProcessCost", cost(serverProcessDuration)),                              // server processing
			zlog.Float64("contentTransferCost", cost(contentTransDuration)),                             // content transfer
			zlog.Float64("totalCost", cost(t.finishTime.Sub(t.getConnTime))),                            // total cost
		}
		zlog.InfoLogger(ctx, "time trace", f...)
	case "http":
		f := []zlog.Field{
			zlog.String(zlog.TopicType, zlog.LogNameModule),
			zlog.Float64("dnsLookupCost", cost(t.dnsDoneTime.Sub(t.dnsStartTime))),          // dns lookup
			zlog.Float64("tcpConnectCost", cost(t.connectDoneTime.Sub(t.connectStartTime))), // tcp connection
			zlog.Float64("serverProcessCost", cost(serverProcessDuration)),                  // server processing
			zlog.Float64("contentTransferCost", cost(contentTransDuration)),                 // content transfer
			zlog.Float64("totalCost", cost(t.finishTime.Sub(t.getConnTime))),                // total cost
		}
		zlog.InfoLogger(ctx, "time trace", f...)
	}
}

func drainAndCloseBody(resp *http.Response, maxBytes int64) {
	if resp != nil {
		_, _ = io.CopyN(io.Discard, resp.Body, maxBytes)
		_ = resp.Body.Close()
	}
}

// retry 原因
type RetryPolicy func(resp *http.Response, err error) bool

// 默认重试策略，仅当底层返回error时重试。不解析http包
var defaultRetryPolicy = func(resp *http.Response, err error) bool {
	return err != nil
}

type TransportOption struct {
	MaxIdleConns        int
	MaxIdleConnsPerHost int
	IdleConnTimeout     time.Duration
	CustomTransport     *http.Transport
}

var globalTransport *http.Transport

// deprecated : 初始化全局的transport, 兼容老的配置
// 建议直接在api.yaml中指定该服务的连接配置（最大空闲连接数目和最大空闲连接时间）
func InitHttp(opts *TransportOption) {
	if opts != nil {
		if opts.CustomTransport != nil {
			globalTransport = opts.CustomTransport
		} else {
			globalTransport = &http.Transport{
				MaxIdleConns:        opts.MaxIdleConns,
				MaxIdleConnsPerHost: opts.MaxIdleConnsPerHost,
				IdleConnTimeout:     opts.IdleConnTimeout,
			}
		}
	}
}

const pathProxy = "/api/mcRawRalRequest"

type NsHead struct {
	Id       uint16
	Version  uint16
	LogId    uint32 `json:"log_id"`
	Provider string `json:"provider"`
	MagicNum uint32
	Reserved uint32
	BodyLen  uint32 `json:"body_len"`
}

func (client *ApiClient) Nshead(ctx *gin.Context, requestBody interface{}, header *NsHead) (response []byte, err error) {
	formData := make(map[string]interface{})

	// 请求第三方nshead服务的数据
	data, err := mcpack.Marshal(requestBody)
	if err != nil {
		return nil, errors.New("json marshal error")
	}

	// header中必须传递body_len
	if header == nil {
		header = new(NsHead)
	}
	header.BodyLen = uint32(len(data))
	h, err := json.Marshal(header)
	if err != nil {
		return nil, errors.New("invalid header")
	}

	formData["header"] = utils.BytesToString(h)
	formData["serviceName"] = client.Service
	formData["body"] = utils.BytesToString(data)
	formData["method"] = "POST"

	opts := HttpRequestOptions{
		Encode:      EncodeForm,
		RequestBody: formData,
	}
	result, err := client.HttpPost(ctx, pathProxy, opts)
	if err != nil {
		return nil, err
	}
	if result.Response == nil {
		return nil, errors.New("get empty response")
	}

	// proxy 成功返回了数据
	resp := struct {
		ErrNo  int    `json:"errNo"`
		ErrMsg string `json:"errMsg"`
		Data   struct {
			Data string `json:"Data"`
		} `json:"data"`
	}{}
	if err = json.Unmarshal(result.Response, &resp); err != nil {
		return nil, err
	}

	content, err := base64.StdEncoding.DecodeString(resp.Data.Data)
	if err != nil {
		return nil, errors.New("response not base64, error is: " + err.Error())
	}

	return content, nil
}
