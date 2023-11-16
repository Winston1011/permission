package components

import "permission/pkg/golib/v2/base"

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
var ErrorAPIGetUserInfoV1 = base.Error{
	ErrNo:  3000,
	ErrMsg: "call GetUserInfoV2 error: %s",
}
var ErrorAPIGetUserInfoV2 = base.Error{
	ErrNo:  3001,
	ErrMsg: "call GetUserInfoV2 error: %s",
}
var ErrorAPIGetUserCourseV1 = base.Error{
	ErrNo:  3002,
	ErrMsg: "call getUserCourse error: %s",
}
var ErrorAPIGetUserCourseV2 = base.Error{
	ErrNo:  3003,
	ErrMsg: "call getUserCourse error: %s",
}

// model层错误
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
