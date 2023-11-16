package kms

import (
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"permission/pkg/golib/v2/base"
	"permission/pkg/golib/v2/utils"
)

type kms struct {
	client      *base.ApiClient
	errorPrefix string
}

func (k kms) Encrypt(ctx *gin.Context, plaintext string) (cipherText string, err error) {
	requestBody := map[string]string{
		"text": plaintext,
	}
	opt := base.HttpRequestOptions{
		RequestBody: requestBody,
	}

	result, err := k.client.HttpPost(ctx, "/encrypt", opt)
	if err != nil {
		return plaintext, err
	}

	response := string(result.Response)
	if strings.HasPrefix(response, k.errorPrefix) {
		return plaintext, errors.New(response)
	}

	return string(result.Response), nil
}

func (k kms) Decrypt(ctx *gin.Context, cipherText string) (plaintext string, err error) {
	requestBody := map[string]string{
		"text": cipherText,
	}
	opt := base.HttpRequestOptions{
		RequestBody: requestBody,
	}
	result, err := k.client.HttpPost(ctx, "/decrypt", opt)
	if err != nil {
		return cipherText, err
	}
	response := string(result.Response)
	if strings.HasPrefix(response, k.errorPrefix) {
		return cipherText, errors.New(response)
	}
	return string(result.Response), nil
}

func (k kms) IsEncrypt(_ *gin.Context, text string) bool {
	if len(text) == 0 {
		return false
	}
	cipher, err := base64.RawStdEncoding.DecodeString(text)
	if err != nil {
		return false
	}
	if len(cipher) < 24 {
		return false
	}
	if string(cipher[:6]) != "ZYBKMS" {
		return false
	}
	appNameLen, err := strconv.ParseInt(string(cipher)[15:23], 10, 64)

	if err != nil {
		return false
	}
	if len(cipher) < int(23+appNameLen) {
		return false
	}
	_, err = hex.DecodeString(string(cipher)[23 : 23+appNameLen])
	if err != nil {
		return false
	}

	return true
}

type kmsDev struct {
	key    []byte
	prefix []byte
}

func (k kmsDev) Encrypt(_ *gin.Context, plaintext string) (cipherText string, err error) {
	if len(plaintext) == 0 {
		return plaintext, errors.New("text is empty")
	}
	// 模拟和KMS一样的长度
	content, err := utils.Rc4EncodeBytes(k.key, []byte(plaintext))
	if err != nil {
		return plaintext, err
	}
	return base64.StdEncoding.EncodeToString(append(k.prefix, content...)), nil
}

func (k kmsDev) Decrypt(_ *gin.Context, cipherText string) (plaintext string, err error) {
	if len(cipherText) == 0 {
		return cipherText, errors.New("text is empty")
	}
	cipher, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return cipherText, err
	}
	if len(cipherText) <= len(k.prefix) {
		return cipherText, errors.New("this cipher text illegal")
	}

	content, err := utils.Rc4DecodeBytes(k.key, string(cipher[len(k.prefix):]))
	if err != nil {
		return cipherText, err
	}
	return string(content), nil
}

func (k kmsDev) IsEncrypt(_ *gin.Context, text string) bool {
	if len(text) == 0 {
		return false
	}
	cipher, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return false
	}
	if len(cipher) < len(k.prefix) {
		return false
	}
	if string(cipher[:len(k.prefix)]) != string(k.prefix) {
		return false
	}

	return true
}
