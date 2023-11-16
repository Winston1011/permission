package kms

import (
	"context"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"permission/pkg/golib/v2/base"
	"permission/pkg/golib/v2/env"
)

type Kms interface {
	Encrypt(ctx *gin.Context, plaintext string) (cipherText string, err error)
	Decrypt(ctx *gin.Context, cipherText string) (plaintext string, err error)
	IsEncrypt(ctx *gin.Context, text string) bool
}

type Option func(*base.ApiClient)

func Init(options ...Option) Kms {
	// 测试环境和本地开发环境相同
	if !env.IsDockerPlatform() || env.GetRunEnv() == env.RunEnvTest {
		return &kmsDev{
			key: []byte("dev"),
			// 符合KMS标准的一个测试前缀，
			prefix: []byte("ZYBKMS1000117880000005663616c6c63656e7465722d746573742e6f70656e706c6174666f726dBe7ds6e706c617466"),
		}
	}
	domain := os.Getenv("KMS_SOCK")
	if domain == "" {
		domain = "/usr/local/var/run/kms.sock"
	}
	client := &base.ApiClient{
		Service:         "kms",
		Domain:          "http://unix",
		Timeout:         50 * time.Millisecond,
		ConnectTimeout:  10 * time.Millisecond,
		Retry:           3,
		IdleConnTimeout: 0,
		HTTPClient: &http.Client{
			Transport: &http.Transport{
				DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
					return net.Dial("unix", domain)
				},
				MaxConnsPerHost:     300,
				MaxIdleConns:        300,
				MaxIdleConnsPerHost: 300,
				IdleConnTimeout:     10 * time.Minute,
			},
		},
	}
	for _, option := range options {
		option(client)
	}
	return &kms{client: client, errorPrefix: "ERROR:"}

}
