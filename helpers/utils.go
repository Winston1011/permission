package helpers

import (
	md52 "crypto/md5"
	"crypto/sha1"
	"fmt"
	"hash/crc32"
	"io"
	"permission/pkg/golib/v2/base"
)

func NewError(err base.Error, str string) base.Error {
	return base.Error{
		ErrNo:  err.ErrNo,
		ErrMsg: fmt.Sprintf(err.ErrMsg, str),
	}
}

func Sha1(str string) string {
	h := sha1.New()
	h.Write([]byte(str))
	return fmt.Sprintf("%x", h.Sum(nil))
}

func Md5(value string) string {
	md5 := md52.New()
	_, err := io.WriteString(md5, value)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", md5.Sum(nil))
}

func CRC32(str string) uint32 {
	return crc32.ChecksumIEEE([]byte(str))
}

func Min(num ...int64) int64 {
	var min int64
	for _, val := range num {
		if min == 0 || val <= min {
			min = val
		}
	}
	return min
}

func Subtraction(a []int64, b []int64) []int64 {
	var c []int64
	temp := map[int64]struct{}{} // map[string]struct{}{}创建了一个key类型为String值类型为空struct的map，Equal -> make(map[string]struct{})
	for _, val := range b {
		if _, ok := temp[val]; !ok {
			temp[val] = struct{}{} // 空struct 不占内存空间
		}
	}
	for _, val := range a {
		if _, ok := temp[val]; !ok {
			c = append(c, val)
		}
	}
	return c
}
