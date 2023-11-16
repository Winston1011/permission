package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"permission/components"
	"permission/conf"
	"permission/pkg/golib/v2/base"
)

const (
	pathGetUserInfoByOpenId = "/passport/user/getInfoByOpenId"
)

type UserInfo struct {
	UserName string `mapstructure:"user_name"`
	UserId   int64  `mapstructure:"open_id"`
}

// GetUserInfoByUserId post json 请求示例
func GetUserInfoByUserId(ctx *gin.Context, appId, userId int64) (userInfo UserInfo, err error) {
	opt := base.HttpRequestOptions{
		Headers: map[string]string{
			"AppID": fmt.Sprintf("%d", appId),
		},
		RequestBody: map[string]int64{
			"open_id": userId,
		},
		Encode: base.EncodeJson,
	}
	res, err := conf.API.Passport.HttpPost(ctx, pathGetUserInfoByOpenId, opt)
	if err != nil {
		return userInfo, components.ErrorApiGetUserInfoV1.WrapPrintf(err, "opt=%+v", opt)
	}
	if _, err := decodeResponse(ctx, res, &userInfo); err != nil {
		return userInfo, components.ErrorApiGetUserInfoV1.WrapPrintf(err, "res=%+v", res)
	}
	return userInfo, nil
}
