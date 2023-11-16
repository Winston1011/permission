package zos

import (
	"encoding/json"
	"fmt"
	"regexp"
	"time"

	"github.com/gin-gonic/gin"
	"permission/pkg/golib/v2/zlog"
)

/*
   注意，图片相关的方法都没有拼接配置中的 file_prefix ，使用需要注意
*/
type ImageMeta struct {
	Format   string `json:"format"`
	Width    string `json:"width"`
	Height   string `json:"height"`
	Size     string `json:"size"`
	Md5      string `json:"md5"`
	PhotoRgb string `json:"photoRgb"`

	// 兼容php，回填数据部分
	Bucket    string `json:"bucket"`
	Pid       string `json:"pid"`
	SourceUrl string `json:"sourceUrl"`
}

//  由于历史原因，该方法没有返回error，该方法返回空的时候表示出错
func (b Bucket) GetImageUrlByPid(ctx *gin.Context, pid, fileType string) (url string) {
	if pid == "" {
		return url
	}

	// 做了pid的判断，php中的逻辑，以防止业务使用了该逻辑，这里不移除pid格式判断的逻辑了。
	// 通用获取URL的方法没有该逻辑，无特殊需求业务可直接调用 GetUrlByFileName
	if match, err := regexp.MatchString(b.Conf.FilePrefix, pid); err != nil || !match {
		zlog.Warnf(ctx, "Invalid pid!")
		return url
	}

	name := pid + "." + fileType
	url, _ = b.GetUrlByFileName(ctx, name, 30*time.Minute)
	return url
}

func (b Bucket) GetThumbnailUrlByPid(ctx *gin.Context, pid, thumbnail, fileType, outType string) (url string) {
	if pid == "" {
		return url
	}

	name := pid + "." + fileType
	imgUrl, _ := b.GetUrlByFileName(ctx, name, 30*time.Minute)

	url = fmt.Sprintf("%s&imageView2/2/%s", imgUrl, thumbnail)
	if outType != "" {
		url = fmt.Sprintf("%s/format/%s", url, outType)
	}

	return url
}

func (b Bucket) GetImageMeta(ctx *gin.Context, pid, fileType string) (m ImageMeta, err error) {
	objectKey := pid + "." + fileType

	addr := b.methodPath(methodImageMeta, objectKey)
	res, err := b.client.HttpPost(ctx, addr, bytesBody(nil))
	if err := checkError(res, err); err != nil {
		return m, err
	}

	if err = json.Unmarshal(res.Response, &m); err != nil {
		return m, err
	}

	m.Bucket = b.Conf.Bucket
	m.Pid = pid
	m.SourceUrl = b.GetImageUrlByPid(ctx, pid, fileType)

	return m, nil
}
