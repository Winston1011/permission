package data_test

import (
	"testing"
	"time"

	"permission/helpers"

	"permission/pkg/golib/v2/zos"
)

/*
	上传本地文件到cos
	如果 fileName 不为空, 则文件名为：fileName.fileType
	如果 fileName 为空，系统默认生成 fileName = prefix_md5(localFile)(_w_h).fileType ,
	其中 prefix_md5(localFile)_w_h 为后面用到的pid
*/
func TestUploadFile(t *testing.T) {
	f := "/Users/jiangshuai/Desktop/img/timeout.png"
	u, err := helpers.Bucket.UploadLocalFile(ctx, f, "target_name", "jpg", true)
	if err != nil {
		t.Error("[UploadFileContent] error: ", err.Error())
		return
	}
	t.Log("[UploadLocalFile] upload success, url : ", u)
}

// 获取对象的访问地址
func TestGetImgUrl(t *testing.T) {
	pid, fileType := "target_name", "jpg"

	// 这个方法更通用，传入的是对象的objectKey，expired 表示访问链接的过期时间
	u, err := helpers.Bucket.GetUrlByFileName(ctx, pid+"."+fileType, 30*time.Minute)
	if err != nil {
		t.Error("[GetUrlByFileName] error: ", err.Error())
		return
	}
	t.Log("[GetUrlByFileName] url: ", u)

	// 批量查询对象的地址
	objectURLs, err := helpers.Bucket.GetUrlByFileNames(ctx, []string{pid + "." + fileType, "test_content.txt"}, 30*time.Minute)
	if err != nil {
		t.Error("[GetUrlByFileName] error: ", err.Error())
		return
	}

	t.Logf("[GetUrlByFileNames] url: %+v", objectURLs)

	// 这个方法是迁移自php，pid 是上传时不指定 objectKey时，lib默认生成的
	imgUrl := helpers.Bucket.GetImageUrlByPid(ctx, pid, fileType)
	t.Log("[GetImageUrlByPid] image url: ", imgUrl)

	// 在上述方法的基础上，指定了图像的宽高比例，表示获取图像的缩略图
	thumbnailURL := helpers.Bucket.GetThumbnailUrlByPid(ctx, pid, "w/50/h/50", fileType, "png")
	t.Log("[GetThumbnailUrlByPid] thumbnail image url: ", thumbnailURL)
}

// 下载文件到本地，srcFileName 为文件在cos上存储的fileName
func TestDownload(t *testing.T) {
	fileName := "target_name.jpg"
	dst := "./download.png"
	err := helpers.Bucket.Download2Local(ctx, fileName, dst)
	if err != nil {
		t.Error("[Download2Local] error: ", err.Error())
		return
	}
	t.Log("[Download2Local] success")
}

// 返回图片的meta信息（宽／高）, pid 为上传时系统默认生成的图片编号
func TestGetImageMeta(t *testing.T) {
	pid := "target_name"
	m, err := helpers.Bucket.GetImageMeta(ctx, pid, "jpg")
	if err != nil {
		t.Error("[GetImageMeta] error: ", err.Error())
		return
	}

	t.Logf("[GetImageMeta] error: %+v", m)
	return
}

/*
	上传内容到cos
	如果 fileName 不为空, 则文件名为：fileName.fileType
	如果 fileName 为空，系统默认生成 fileName = prefix_md5(content).fileType ,
*/

func TestUploadContent(t *testing.T) {
	content := "this is test content to upload"
	u, err := helpers.Bucket.UploadFileContent(ctx, content, "jiangshuai02/test_content", "txt")
	if err != nil {
		t.Error("[UploadFileContent] error: ", err.Error())
		return
	}

	t.Log("[Download2Local] upload success, url : ", u)
}

// 下载文件内容
func TestDownloadContent(t *testing.T) {
	fileName := "jiangshuai02/test_content.txt"
	content, err := helpers.Bucket.DownloadContent(ctx, fileName)
	if err != nil {
		t.Error("[Download2Local] error: ", err.Error())
		return
	}

	t.Log("[Download2Local] content is: ", string(content))
}

// 按照一定规则筛选bucket下的对象
func TestGetObjectList(t *testing.T) {
	opt := zos.ListObjectsOption{
		Prefix:  "jiangshuai02",
		MaxKeys: 10,
	}
	contents, err := helpers.Bucket.GetObjectList(ctx, opt)
	if err != nil {
		t.Error("[GetObjectList] error: ", err.Error())
		return
	}

	t.Logf("[GetObjectList] is: %+v", contents)
}

// 判断对象是否存在
func TestFileIsExist(t *testing.T) {
	fileName := "target_names.jpg"
	isExist, err := helpers.Bucket.FileIsExist(ctx, fileName)
	if err != nil {
		t.Error("[FileIsExist] error: ", err.Error())
		return
	}

	t.Logf("[FileIsExist] is: %+v", isExist)
}

// 删除对象，支持批量删除，逗号分隔，批量个数建议<10个
func TestDeleteFile(t *testing.T) {
	fileName := "target_name.jpg,jiangshuai02/test_content.txt"
	err := helpers.Bucket.DeleteFile(ctx, fileName)
	if err != nil {
		t.Error("[DeleteFile] error: ", err.Error())
		return
	}

	t.Log("[DeleteFile] success")
}

func TestGetTempKeys(t *testing.T) {
	c, err := helpers.Bucket.GetTempKeys(ctx, 30*time.Minute)
	if err != nil {
		t.Error("[GetTempKeys] error: ", err.Error())
		return
	}

	t.Logf("[GetTempKeys] success: %+v", c)
}
