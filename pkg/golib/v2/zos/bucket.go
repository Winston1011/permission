package zos

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"permission/pkg/golib/v2/utils"
	"permission/pkg/golib/v2/zlog"
)

const (
	// 单次上传文件大小使用tcos的默认值5G
	maxUnSliceFileSize int = 5 * 1024 * 1024 * 1024
	// 最小分块大小为1M
	minSliceFileSize int = 1 * 1024 * 1024
)

func (b Bucket) UploadLocalFile(ctx *gin.Context, localFilePath, fileName, fileType string, needSize bool) (dwUrl string, err error) {
	var fd *os.File
	if fd, err = os.Open(localFilePath); err != nil {
		zlog.Warnf(ctx, "open local file error: %s", err.Error())
		return dwUrl, err
	}
	defer fd.Close()

	content, err := ioutil.ReadAll(fd)
	if err != nil {
		return dwUrl, err
	}

	objectKey := fileName + "." + fileType
	if fileName == "" {
		// 默认规则生成存储到bos的object key
		name := utils.Md5(localFilePath)
		if needSize {
			if im, _, err := image.DecodeConfig(bytes.NewReader(content)); err == nil {
				name = fmt.Sprintf("%s_%d_%d", name, im.Width, im.Height)
			}
		}
		objectKey = fmt.Sprintf("%s%s.%s", b.Conf.FilePrefix, name, fileType)
	}

	if b.Conf.Directory != "" {
		objectKey = path.Join(b.Conf.Directory, objectKey)
	}

	if len(content) <= b.Conf.UploadChunk {
		// 使用简单上传
		return b.simpleUpload(ctx, objectKey, content)
	} else {
		// 分段上传
		return b.multiUpload(ctx, objectKey, content)
	}
}

func (b Bucket) UploadFileContent(ctx *gin.Context, content, fileName, fileType string) (dwUrl string, err error) {
	objectKey, err := b.genObjectKey4FileContentUpload(fileName, fileType, content, "")
	if err != nil {
		return dwUrl, err
	}

	data := utils.StringToBytes(content)

	if len(content) <= b.Conf.UploadChunk {
		// 使用简单上传
		return b.simpleUpload(ctx, objectKey, data)
	} else {
		// 分段上传
		return b.multiUpload(ctx, objectKey, data)
	}
}

func (b Bucket) simpleUpload(ctx *gin.Context, object string, content []byte) (dwUrl string, err error) {
	addr := b.methodPath(methodUpload, object)
	res, err := b.client.HttpPost(ctx, addr, bytesBody(content))
	if err := checkError(res, err); err != nil {
		return dwUrl, err
	}

	dwUrl = utils.BytesToString(res.Response)
	return dwUrl, nil
}

func (b Bucket) multiUpload(ctx *gin.Context, object string, content []byte) (dwUrl string, err error) {
	addr := b.methodPath(methodUpload, object)
	totalSize := len(content)

	blockSize := b.Conf.UploadChunk
	parts := (totalSize / blockSize) + 1

	type PartEtag struct {
		PartNumber int
		Etag       string
	}
	var opts []PartEtag
	var uploadID string
	for i := 1; i <= parts; i++ {
		pos := blockSize
		if len(content) < blockSize {
			pos = len(content)
		}
		buf := content[:pos]
		content = content[pos:]

		curAddr := addr + "&partNumber=" + strconv.Itoa(i) + "&uploadID=" + uploadID
		res, err := b.client.HttpPost(ctx, curAddr, bytesBody(buf))
		if err := checkError(res, err); err != nil {
			return dwUrl, err
		}

		uploadID = res.Header.Get("X-ZYB-UploadID")
		etag := res.Header.Get("X-ZYB-Etag")

		e := PartEtag{
			PartNumber: i,
			Etag:       etag,
		}
		opts = append(opts, e)
	}

	curAddr := addr + "&chunkFlag=end&uploadID=" + uploadID
	data, err := json.Marshal(opts)
	if err != nil {
		return dwUrl, err
	}

	res, err := b.client.HttpPost(ctx, curAddr, bytesBody(data))
	if err := checkError(res, err); err != nil {
		return dwUrl, err
	}

	return b.GetUrlByFileName(ctx, object, 30*time.Minute)
}

// 简单上传，上传<5G的内容
func (b Bucket) UploadContent(ctx *gin.Context, fileSize int, io io.Reader, objectKey string) (dwUrl string, err error) {
	if fileSize > maxUnSliceFileSize {
		return dwUrl, errors.New("upload file too large")
	}

	buf := &bytes.Buffer{}
	if _, err := buf.ReadFrom(io); err != nil {
		return dwUrl, err
	}

	addr := b.methodPath(methodUpload, objectKey)
	res, err := b.client.HttpPost(ctx, addr, bytesBody(buf.Bytes()))
	if err := checkError(res, err); err != nil {
		return dwUrl, err
	}

	dwUrl = utils.BytesToString(res.Response)

	return dwUrl, nil
}

func (b Bucket) Download2Local(ctx *gin.Context, srcFileName, dstFileName string) (err error) {
	if srcFileName == "" {
		return errors.New("fileName is empty")
	}

	objectKey := srcFileName
	if b.Conf.Directory != "" {
		objectKey = fmt.Sprintf("%s/%s", b.Conf.Directory, srcFileName)
	}

	addr := b.methodPath(methodDownload, objectKey)
	res, err := b.client.HttpPost(ctx, addr, bytesBody(nil))
	if err := checkError(res, err); err != nil {
		return err
	}

	// If file exist, overwrite it
	fd, err := os.OpenFile(dstFileName, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0660)
	if err != nil {
		return err
	}

	defer fd.Close()

	_, err = fd.Write(res.Response)
	if err != nil {
		return err
	}

	return nil
}

func (b Bucket) DownloadContent(ctx *gin.Context, srcFileName string) (content []byte, err error) {
	if srcFileName == "" {
		return content, errors.New("fileName is empty")
	}

	objectKey := srcFileName
	if b.Conf.Directory != "" {
		objectKey = fmt.Sprintf("%s/%s", b.Conf.Directory, srcFileName)
	}

	addr := b.methodPath(methodDownload, objectKey)
	res, err := b.client.HttpPost(ctx, addr, bytesBody(nil))
	if err := checkError(res, err); err != nil {
		return content, err
	}

	content = res.Response

	return content, nil
}

func (b Bucket) IsExist(ctx *gin.Context) (bool, error) {
	addr := uriExist + "?bucket=" + b.Conf.Bucket

	res, err := b.client.HttpPost(ctx, addr, bytesBody(nil))
	return checkExist(res, err)
}

func (b Bucket) GetObjectList(ctx *gin.Context, option ListObjectsOption) (objectList []Object, err error) {
	data, err := json.Marshal(option)
	if err != nil {
		return nil, err
	}

	addr := b.methodPath(methodListObject, "")
	res, err := b.client.HttpPost(ctx, addr, bytesBody(data))
	if err := checkError(res, err); err != nil {
		return nil, err
	}

	if err := json.Unmarshal(res.Response, &objectList); err != nil {
		return nil, err
	}

	return objectList, nil
}

func (b Bucket) GetUrlByFileName(ctx *gin.Context, name string, expired time.Duration) (string, error) {
	if name == "" {
		return "", errors.New("invalid name")
	}

	objectKey := name
	if b.Conf.Directory != "" {
		objectKey = b.Conf.Directory + "/" + name
	}

	addr := b.methodPath(methodURL, objectKey)
	res, err := b.client.HttpPost(ctx, addr, stringBody("expire="+strconv.Itoa(int(expired.Seconds()))))
	if err := checkError(res, err); err != nil {
		return "", err
	}

	body := utils.BytesToString(res.Response)
	return body, nil
}

func (b Bucket) GetUrlByFileNames(ctx *gin.Context, names []string, expired time.Duration) (resp map[string]string, err error) {
	var objectKeys string
	for _, name := range names {
		if name == "" {
			continue
		}

		objectKey := name
		if b.Conf.Directory != "" {
			objectKey = b.Conf.Directory + "/" + name
		}
		if objectKeys != "" {
			objectKeys += "," + objectKey
		} else {
			objectKeys += objectKey
		}
	}

	addr := b.methodPath(methodURL, objectKeys)
	res, err := b.client.HttpPost(ctx, addr, stringBody("batch=1&expire="+strconv.Itoa(int(expired.Seconds()))))
	if err := checkError(res, err); err != nil {
		return nil, err
	}

	resp = make(map[string]string, len(names))
	if err = json.Unmarshal(res.Response, &resp); err != nil {
		return resp, err
	}

	return resp, nil
}

func (b Bucket) FileIsExist(ctx *gin.Context, fileName string) (bool, error) {
	objectKey := path.Join(b.Conf.Directory, fileName)

	addr := b.methodPath(methodObjectExist, objectKey)
	res, err := b.client.HttpPost(ctx, addr, bytesBody(nil))
	return checkExist(res, err)
}

func (b Bucket) DeleteFile(ctx *gin.Context, fileName string) error {
	// fileName 支持批量删除，使用逗号分隔两个文件名
	fs := strings.Split(fileName, ",")
	var objectKey string
	for _, name := range fs {
		if name == "" {
			continue
		}

		if objectKey == "" {
			objectKey += path.Join(b.Conf.Directory, name)
		} else {
			objectKey += "," + path.Join(b.Conf.Directory, name)
		}
	}

	addr := b.methodPath(methodDelete, objectKey)
	res, err := b.client.HttpPost(ctx, addr, bytesBody(nil))
	if err := checkError(res, err); err != nil {
		return err
	}

	return nil
}

func (b Bucket) methodPath(method string, object string) string {
	if method == "" {
		return uriMethod
	}

	// query参数
	v := url.Values{}
	v.Add("bucket", b.Conf.Bucket)
	v.Add("method", method)
	if object != "" {
		v.Add("object", object)
	}

	// 固定请求地址
	addr := uriMethod + "?" + v.Encode()
	return addr
}

func (b Bucket) genObjectKey4FileContentUpload(fileName, fileType string, content string, version string) (objectKey string, err error) {
	if content == "" {
		return objectKey, errors.New("content is empty")
	}

	if fileType == "" && version == "" {
		fileType = "jpg"
	}

	objectKey = fileName
	if objectKey == "" {
		// 默认规则生成存储到cos的object key
		objectKey = b.Conf.FilePrefix + utils.Md5(content)
	}

	if fileType != "" {
		objectKey = objectKey + "." + fileType
	}

	if b.Conf.Directory != "" {
		objectKey = fmt.Sprintf("%s/%s", b.Conf.Directory, objectKey)
	}

	return objectKey, nil
}

func (b Bucket) GetTempKeys(ctx *gin.Context, expired time.Duration) (c Credentials, err error) {
	addr := b.methodPath(methodTempKey, "")
	res, err := b.client.HttpPost(ctx, addr, stringBody("expire="+strconv.Itoa(int(expired.Seconds()))))
	if err := checkError(res, err); err != nil {
		return c, err
	}

	if err := json.Unmarshal(res.Response, &c); err != nil {
		return c, err
	}

	return c, nil
}
