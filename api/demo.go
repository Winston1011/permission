package api

import (
	"encoding/json"

	"permission/components"
	"permission/conf"

	"github.com/gin-gonic/gin"
	"permission/pkg/golib/v2/base"
	"permission/pkg/golib/v2/zlog"
)

const (
	pathGetUserCourse = "/permission/api/user/get-user-course"
	pathGetUserInfo   = "/permission/api/user/getlist"

	pathBuyCourse     = "/permission/api/course/buy"
	pathGetCourseInfo = "/permission/api/course/getinfo"
)

type UserInfo struct {
	UserID int64 `mapstructure:"userID"`

	// 首字母要大写，否则后面用到的 mapstructure 解析后不会给该字段赋值
	NickName string `mapstructure:"userName" json:"userName"`

	// 使用 mapstructure tag 重命名，把业务返回的Email解析到 UserEmail上
	UserEmail string `mapstructure:"Email"`

	// 错误示例，多个tag之间不能用逗号分隔
	// UserAge string `json:"age",mapstructure:"age"`

	// 把age解析到UserAge，注意多个tag之间用空格分隔
	UserAge uint8 `mapstructure:"age" json:"age"`
}

// post form 请求示例
func GetUserList(ctx *gin.Context) (count int, info []UserInfo, err error) {
	test := map[string]string{
		"aa": "11",
	}
	testStr, err := json.Marshal(test)
	if err != nil {
		return count, info, components.ErrorAPIGetUserInfoV2.WrapPrintf(err, "data=%+v", test)
	}

	opt := base.HttpRequestOptions{
		// 最初 EncodeForm 只支持 map[string]string 类型，这种方式下，需要用户对非string类型主动jsonEncode
		// 比如下述test，下游拿到的test参数就是一个json字符串
		RequestBody: map[string]string{
			"userIDList": "1010,10",
			"test":       string(testStr),
		},

		// 考虑到很多业务传递的map元素个数较大，类型不统一，所以新增支持 map[string]interface{} 类型
		// 注意，这种情况下，底层会对value做jsonEncode
		// 如果传value本身是一个json字符串(比如下述test1)，那么下游接收到的是转义后的json字符串: "{\"aa\":\"11\"}"，需要使用 strconv.Unquote 处理一下转义字符
		// 如果传的value是一个map(比如下述test2)，那么下游收到的是json字符串: {"aa":"11"}
		// RequestBody: map[string]interface{}{
		//	"username": username,
		//	"test1":    string(testStr),
		//	"test2":    test,
		// },

		// EncodeForm 类型不支持string类型
		// RequestBody: "test",

		// 暂不支持该种类型, 可直接改为 map[string]interface 类型，form拼接后的参数形如：username=1&test=2
		// RequestBody: map[string]int {
		//	"username": 1,
		//	"test":     2,
		// },
		Encode: base.EncodeForm,
		Headers: map[string]string{
			"host": "1.1.1.1",
		},
		Cookies: map[string]string{
			"uname":  "xx",
			"ZYBUSS": "b612",
		},
	}
	res, err := conf.API.Demo.HttpPost(ctx, pathGetUserInfo, opt)
	if err != nil {
		return count, info, components.ErrorAPIGetUserInfoV2.WrapPrintf(err, "opt=%+v", opt)
	}

	resp := struct {
		ErrNo  int    `json:"errNo"`
		ErrMsg string `json:"errMsg"`
		Data   struct {
			Count int        `json:"count"`
			List  []UserInfo `json:"list"`
		} `json:"data"`
	}{}

	if err = json.Unmarshal(res.Response, &resp); err != nil {
		zlog.Errorf(ctx, "http response decode err, err: %s", err.Error())
		return count, info, components.ErrorAPIGetUserInfoV2.WrapPrintf(err, "res=%+v", res)
	}

	if resp.ErrNo != 0 {
		return count, info, components.ErrorAPIGetUserInfoV2.WrapPrintf(err, "res=%+v", res)
	}

	return resp.Data.Count, resp.Data.List, nil
}

// post json 请求示例
func BuyCourse(ctx *gin.Context, userID, courseID string) (orderID string, err error) {
	opt := base.HttpRequestOptions{
		RequestBody: map[string]string{
			"userID":   userID,
			"courseID": courseID,
			"password": "556677889900",
		},
		Encode: base.EncodeJson,
	}
	res, err := conf.API.Demo.HttpPost(ctx, pathBuyCourse, opt)
	if err != nil {
		return orderID, components.ErrorAPIGetUserInfoV1.WrapPrintf(err, "opt=%+v", opt)
	}

	resp := struct {
		UserID  string `mapstructure:"userID" json:"userID"`
		OrderID string `mapstructure:"orderID" json:"orderID"`
	}{}
	if _, err := decodeResponse(ctx, res, &resp); err != nil {
		return orderID, components.ErrorAPIGetUserInfoV1.WrapPrintf(err, "res=%+v", res)
	}

	orderID = resp.OrderID
	return orderID, nil
}

// post raw data 请求示例
func BuyCourseRaw(ctx *gin.Context, userID, courseID string) (orderID string, err error) {
	data := map[string]string{
		"userID":   userID,
		"courseID": courseID,
	}
	b, err := json.Marshal(data)
	if err != nil {
		return orderID, components.ErrorAPIGetUserInfoV2.WrapPrintf(err, "data=%+v", data)
	}

	opt := base.HttpRequestOptions{
		// 注意，发送raw data，需要指定encode为 EncodeRaw，同时保证data为string类型
		RequestBody: string(b),
		Encode:      base.EncodeRaw,
	}
	res, err := conf.API.Demo.HttpPost(ctx, pathBuyCourse, opt)
	if err != nil {
		return orderID, components.ErrorAPIGetUserInfoV1.WrapPrintf(err, "opt=%+v", opt)
	}

	resp := struct {
		UserID  string `mapstructure:"userID" json:"userID"`
		OrderID string `mapstructure:"orderID" json:"orderID"`
	}{}
	if _, err := decodeResponse(ctx, res, &resp); err != nil {
		return orderID, components.ErrorAPIGetUserInfoV1.WrapPrintf(err, "res=%+v", res)
	}

	return resp.OrderID, nil
}

// get 请求示例
func GetCourseInfo(ctx *gin.Context, courseID string) (info map[string]interface{}, err error) {
	opt := base.HttpRequestOptions{
		// 仍然兼容之前使用 Data 发送数据
		RequestBody: map[string]string{
			"courseID": courseID,
		},
	}
	res, err := conf.API.Demo.HttpGet(ctx, pathGetCourseInfo, opt)
	if err != nil {
		return info, components.ErrorAPIGetUserCourseV1.WrapPrintf(err, "RequestBody=%+v", opt.RequestBody)
	}

	if _, err := decodeResponse(ctx, res, &info); err != nil {
		return info, components.ErrorAPIGetUserCourseV1.WrapPrintf(err, "res=%+v", res)
	}

	return info, nil
}

// post mcPack 请求示例
func GetUserCourse(ctx *gin.Context, userID string) (info map[string]interface{}, err error) {
	opt := base.HttpRequestOptions{
		RequestBody: map[string]string{
			"userID":   userID,
			"username": "张三",
		},
		Encode: base.EncodeMcPack,
	}
	res, err := conf.API.Demo.HttpPost(ctx, pathGetUserCourse, opt)
	if err != nil {
		return info, components.ErrorAPIGetUserCourseV2.WrapPrintf(err, "RequestBody=%+v", opt.RequestBody)
	}

	if _, err := decodeResponse(ctx, res, &info); err != nil {
		return info, components.ErrorAPIGetUserCourseV2.WrapPrintf(err, "res=%+v", res)
	}

	return info, nil
}
