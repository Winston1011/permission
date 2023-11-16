package components

import "permission/pkg/golib/v2/base"

// 7000000-7499999 permission逻辑错误
var ErrorPermissionParamsInvalid = base.Error{
	ErrNo:  7000,
	ErrMsg: "permission param invalid",
}

// 7500000-7999999 group权限组逻辑错误
var ErrorGroupParamsInvalid = base.Error{
	ErrNo:  7500,
	ErrMsg: "group param invalid",
}

// 8000000-8499999 policy校验规则逻辑错误
var ErrorPolicyParamsInvalid = base.Error{
	ErrNo:  8000,
	ErrMsg: "policy param invalid",
}

// 8500000-8999999 usergroup用户权限组逻辑错误
var ErrorUserGroupParamsInvalid = base.Error{
	ErrNo:  8500,
	ErrMsg: "user group param invalid",
}

// 9000000-9299999 node节点资源逻辑错误
var ErrorNodeParamsInvalid = base.Error{
	ErrNo:  9000,
	ErrMsg: "node param invalid",
}

// 9300000-9499999 menu节点资源逻辑错误
var ErrorMenuParamsInvalid = base.Error{
	ErrNo:  9300,
	ErrMsg: "menu param invalid",
}

// 9500000-9999999 用户未登录逻辑错误
var ErrorParamUserNotLogin = base.Error{
	ErrNo:  9500,
	ErrMsg: "param user not login",
}

// 6000000-6999999 oauth逻辑错误
var ErrorOauthParamsInvalid = base.Error{
	ErrNo:  6000,
	ErrMsg: "oauth param invalid",
}
var ErrorAppNotExist = base.Error{
	ErrNo:  6001,
	ErrMsg: "app not exist",
}
var ErrorAppSecretInvalid = base.Error{
	ErrNo:  6002,
	ErrMsg: "app secret invalid",
}
var ErrorTokenInvalid = base.Error{
	ErrNo:  6003,
	ErrMsg: "access token invalid",
}
var ErrorTokenOverdue = base.Error{
	ErrNo:  6004,
	ErrMsg: "access token overdue",
}
var ErrorNoAccess = base.Error{
	ErrNo:  6005,
	ErrMsg: "no access",
}
var ErrorRefreshTokenInvalid = base.Error{
	ErrNo:  6006,
	ErrMsg: "refresh token invalid",
}

// 4000000-4999999 参数检查错误
var ErrorParamInvalid = base.Error{
	ErrNo:  4000,
	ErrMsg: "param invalid",
}

// 5000000-5999999 内部逻辑错误
var ErrorSystemError = base.Error{
	ErrNo:  5000,
	ErrMsg: "system internal error",
}
var ErrorUserNotExist = base.Error{
	ErrNo:  5001,
	ErrMsg: "user not exist",
}

// 3000000-3999999 下游系统返回错误
// api调用相关错误
var ErrorApiGetUserInfoV1 = base.Error{
	ErrNo:  3000,
	ErrMsg: "call GetUserInfoV2 error: %s",
}
var ErrorApiGetUserInfoV2 = base.Error{
	ErrNo:  3001,
	ErrMsg: "call GetUserInfoV2 error: %s",
}
var ErrorApiGetUserCourseV1 = base.Error{
	ErrNo:  3002,
	ErrMsg: "call getUserCourse error: %s",
}
var ErrorApiGetUserCourseV2 = base.Error{
	ErrNo:  3003,
	ErrMsg: "call getUserCourse error: %s",
}
var ErrorApiGetUserInfo = base.Error{
	ErrNo:  3004,
	ErrMsg: "call getUserInfo from passport error: %s",
}

// model层错误
var ErrorDbError = base.Error{
	ErrNo:  3100,
	ErrMsg: "db error",
}

var ErrorDbInsert = base.Error{
	ErrNo:  3101,
	ErrMsg: "db insert error: %s",
}
var ErrorDbUpdate = base.Error{
	ErrNo:  3102,
	ErrMsg: "db update error: %s",
}
var ErrorDbSelect = base.Error{
	ErrNo:  3103,
	ErrMsg: "db get error: %s",
}
var ErrorDbDelete = base.Error{
	ErrNo:  3104,
	ErrMsg: "db delete error: %s",
}
var ErrorDbUpsert = base.Error{
	ErrNo:  3105,
	ErrMsg: "db upsert error: %s",
}

// 第三方sdk错误

// redis
var ErrorRedisGet = base.Error{
	ErrNo:  3201,
	ErrMsg: "redis get error: %s",
}
var ErrorRedisSet = base.Error{
	ErrNo:  3202,
	ErrMsg: "redis set error: %s",
}

// hbase
var ErrorHbaseGetTableName = base.Error{
	ErrNo:  3301,
	ErrMsg: "get hbase table name error",
}
var ErrorHbaseQuery = base.Error{
	ErrNo:  3302,
	ErrMsg: "hbase query error",
}

// kafka
var ErrorKafkaPub = base.Error{
	ErrNo:  3401,
	ErrMsg: "kafka pub error",
}

// nmq
var ErrorNmqPub = base.Error{
	ErrNo:  3402,
	ErrMsg: "nmq pub error",
}

// es
var ErrorEsPing = base.Error{
	ErrNo:  3501,
	ErrMsg: "es ping error",
}
var ErrorEsGetVersion = base.Error{
	ErrNo:  3502,
	ErrMsg: "es getVersion error",
}
var ErrorEsInsert = base.Error{
	ErrNo:  3503,
	ErrMsg: "es insert error",
}
var ErrorEsQuery = base.Error{
	ErrNo:  3504,
	ErrMsg: "es query error",
}
var ErrorEsUpdate = base.Error{
	ErrNo:  3505,
	ErrMsg: "es update error",
}
var ErrorEsDel = base.Error{
	ErrNo:  3506,
	ErrMsg: "es del error",
}

// cos
var ErrorCosUpload = base.Error{
	ErrNo:  3600,
	ErrMsg: "cos upload error: %s",
}
var ErrorCosDownload = base.Error{
	ErrNo:  3602,
	ErrMsg: "cos download error: %s",
}
var ErrorCosGetData = base.Error{
	ErrNo:  3603,
	ErrMsg: "cos getMetaData error: %s",
}
