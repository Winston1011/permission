package zos

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strconv"
	"time"

	"permission/pkg/golib/v2/base"
	"permission/pkg/golib/v2/env"
)

const (
	methodUpload      = "upload"
	methodDownload    = "download"
	methodDelete      = "delete"
	methodListObject  = "listObject"
	methodImageMeta   = "imageMeta"
	methodURL         = "url"
	methodObjectExist = "objectExist"
	methodTempKey     = "tempKey"
)

const (
	uriMethod = "/zos-proxy/bucket/method"
	uriExist  = "/zos-proxy/bucket/exist"
)

const (
	// 资源不存在
	statusCodeNoSuchKey = 244
	// 成功
	statusCodeSuccess = 200
)

const sockAddr = "/usr/local/var/run/zos.sock"
const headerDevAppName = "X-ZYB-Appname" // 线下环境app透传appName作为对象的prefix前缀

type Bucket struct {
	Conf   Conf
	client *base.ApiClient
}

type Conf struct {
	BucketConf `yaml:"buckets"`
	ClientConf `yaml:"client"`
}

type BucketConf struct {
	Bucket      string `yaml:"bucket"`
	Directory   string `yaml:"directory"`
	FilePrefix  string `yaml:"filePrefix"`
	UploadChunk int    `yaml:"uploadChunk"`
}

type ClientConf struct {
	Domain         string        `yaml:"domain"`
	Retry          int           `yaml:"retry"`
	ConnectTimeout time.Duration `yaml:"connectTimeout"`
	Timeout        time.Duration `yaml:"timeout"`

	MaxIdleConns        int           `yaml:"maxIdleConns"`
	MaxIdleConnsPerHost int           `yaml:"maxIdleConnsPerHost"`
	MaxConnsPerHost     int           `yaml:"maxConnsPerHost"`
	IdleConnTimeout     time.Duration `yaml:"idleConnTimeout"`
}

type CustomerConfig struct {
	Buckets    []BucketConf `yaml:"buckets"`
	ClientConf `yaml:"client"`
}

type ListObjectsOption struct {
	Delimiter    string `json:"delimiter"`
	Marker       string `json:"marker"`
	MaxKeys      int    `json:"maxKeys"`
	Prefix       string `json:"prefix"`
	EncodingType string `json:"encodingType"`
}

type Object struct {
	Key          string `json:"key"`
	LastModified string `json:"lastModified"`
	ETag         string `json:"etag"`
	Size         int64  `json:"size"`
	StorageClass string `json:"storageClass"`
}

type Credentials struct {
	SessionToken string `json:"sessionToken"`
	TmpSecretID  string `json:"tmpSecretId"`
	TmpSecretKey string `json:"tmpSecretKey"`
}

func (conf *CustomerConfig) checkConf() {
	if conf.ClientConf.ConnectTimeout == 0 {
		conf.ClientConf.ConnectTimeout = 500 * time.Millisecond
	}
	if conf.ClientConf.Timeout == 0 {
		conf.ClientConf.Timeout = 10 * time.Second
	}

	if conf.ClientConf.MaxIdleConns == 0 {
		conf.ClientConf.MaxIdleConns = 50
	}
	if conf.ClientConf.MaxIdleConnsPerHost == 0 {
		conf.ClientConf.MaxIdleConnsPerHost = 50
	}

	if conf.ClientConf.IdleConnTimeout == 0 {
		conf.ClientConf.IdleConnTimeout = 10 * time.Minute
	}

}

func NewBucket(conf CustomerConfig) map[string]Bucket {
	conf.checkConf()

	var domain string
	var dialContext func(ctx context.Context, network, addr string) (net.Conn, error)
	if env.IsDockerPlatform() {
		domain = "http://zos"
		dialContext = func(_ context.Context, _, _ string) (net.Conn, error) {
			d := &net.Dialer{
				Timeout: conf.ClientConf.ConnectTimeout,
			}
			return d.Dial("unix", sockAddr)
		}
	} else {
		domain = "http://zos-proxy-base-cc.suanshubang.cc"
		if conf.ClientConf.Domain != "" {
			domain = conf.ClientConf.Domain
		}
		dialContext = (&net.Dialer{
			Timeout: conf.ClientConf.ConnectTimeout,
		}).DialContext
	}

	client := &base.ApiClient{
		Service:         "zos",
		Domain:          domain,
		Timeout:         conf.ClientConf.Timeout,
		ConnectTimeout:  conf.ClientConf.ConnectTimeout,
		Retry:           conf.ClientConf.Retry,
		IdleConnTimeout: conf.ClientConf.IdleConnTimeout,
		MaxReqBodyLen:   -1,
		MaxRespBodyLen:  -1,
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				DialContext:         dialContext,
				MaxConnsPerHost:     conf.ClientConf.MaxIdleConnsPerHost,
				MaxIdleConns:        conf.ClientConf.MaxIdleConns,
				MaxIdleConnsPerHost: conf.ClientConf.MaxIdleConnsPerHost,
				IdleConnTimeout:     conf.ClientConf.IdleConnTimeout,
			},
		},
	}

	bucketList := make(map[string]Bucket, 10)

	for _, bucketDetail := range conf.Buckets {
		v := bucketDetail
		if v.UploadChunk == 0 {
			v.UploadChunk = maxUnSliceFileSize // 默认分段上传大小为5G
		} else if v.UploadChunk < minSliceFileSize {
			v.UploadChunk = minSliceFileSize
		}

		b := Bucket{
			Conf: Conf{
				BucketConf: v,
				ClientConf: conf.ClientConf,
			},
			client: client,
		}

		bucketList[bucketDetail.Bucket] = b
	}

	return bucketList
}

func bytesBody(data []byte) base.HttpRequestOptions {
	opt := base.HttpRequestOptions{}
	if data != nil {
		opt.RequestBody = data
		opt.Encode = base.EncodeRawByte
	}

	if env.GetRunEnv() == env.RunEnvTest {
		opt.Headers = map[string]string{
			headerDevAppName: env.AppName,
			"Content-Length": strconv.Itoa(len(data)),
		}
	}

	return opt
}

func stringBody(data string) base.HttpRequestOptions {
	opt := base.HttpRequestOptions{}
	if data != "" {
		opt.RequestBody = data
		opt.Encode = base.EncodeRaw
	}

	if env.GetRunEnv() == env.RunEnvTest {
		opt.Headers = map[string]string{
			headerDevAppName: env.AppName,
			"Content-Length": strconv.Itoa(len(data)),
		}
	}

	return opt
}

func checkError(res *base.ApiResult, err error) error {
	if err != nil {
		return err
	}

	if res.HttpCode != statusCodeSuccess {
		return errors.New(string(res.Response))
	}
	return nil
}

func checkExist(res *base.ApiResult, err error) (bool, error) {
	if err != nil {
		return false, err
	}

	if res.HttpCode == statusCodeSuccess {
		return true, nil
	} else if res.HttpCode == statusCodeNoSuchKey {
		return false, nil
	}

	return false, errors.New(string(res.Response))
}
