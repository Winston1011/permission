package utils

import (
	"crypto/md5"
	"crypto/rc4"
	"encoding/base64"
	"encoding/hex"
	"math/rand"
	"reflect"
	"time"
	"unsafe"
)

func Base64Encode(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func Md5(plain string) string {
	h := md5.New()
	h.Write([]byte(plain))
	cipherStr := h.Sum(nil)
	return hex.EncodeToString(cipherStr)
}

/*
  获取随机数
  不传参：0-100
  传1个参数：0-指定参数
  传2个参数：第1个参数-第2个参数
*/

func RandNum(num ...int) int {
	var start, end int
	if len(num) == 0 {
		start = 0
		end = 100
	} else if len(num) == 1 {
		start = 0
		end = num[0]
	} else {
		start = num[0]
		end = num[1]
	}

	rRandNumUtils := rand.New(rand.NewSource(time.Now().UnixNano()))
	return rRandNumUtils.Intn(end-start+1) + start
}

// StringToBytes converts string to byte slice without a memory allocation.
func StringToBytes(s string) (b []byte) {
	sh := *(*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	bh.Data, bh.Len, bh.Cap = sh.Data, sh.Len, sh.Len
	return b
}

// BytesToString converts byte slice to string without a memory allocation.
func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func Rc4Encode(key, plainText string) (string, error) {
	src := StringToBytes(plainText)
	k := StringToBytes(key)

	return Rc4EncodeBytes(k, src)
}

func Rc4EncodeBytes(key, plainText []byte) (string, error) {
	c, err := rc4.NewCipher(key)
	if err != nil {
		return "", err
	}

	dst := make([]byte, len(plainText))
	c.XORKeyStream(dst, plainText)
	return BytesToString(dst), nil
}
