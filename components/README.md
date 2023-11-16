
## 错误的定义

业务中出现的error信息应该都有明确的errno/errmsg，所以建议对于错误码都在`components/error.go`中进行定义：

```go
// api调用相关错误
var ErrorApiGetUserCourseV1 = base.Error{
   ErrNo:  5102000,
   ErrMsg: "call getUserCourse error: %s",
}

// models层错误
var ErrorMdlUpdateGormDemoDescByName = base.Error{
   ErrNo:  5103000,
   ErrMsg: "UpdateGormDemoDescByName error: %s",
}
```
对error错误段建议，可以对号段进一步拆分，通过errno方便识别错误类型

|号段范围   |  含义 | 
|---|---|
|5000000-5999999	| 内部逻辑错误 |
|4000000-4999999    | 参数检查错误 |
|3000000-3999999	| 下游系统返回错误 |


## 非定义错误在发生处转化为定义错误
定义指的是在 `components/error.go` 中定义的错误。
当调用第三方sdk时，一般抛出一个非`base.Error` 类型的错误，该种类型没有明确的错误码，或者错误码的定义不符合规范。
这种情况下，我们应当在`components/error.go` 提前定义好错误码。
比如，在调用gorm时gorm返回使用 `errors.New()` 得到的错误，此时我们要把他转为 base.Error 类型的错误：

```go
func (demo *GormDemo) Insert(ctx *gin.Context) (err error) {
   dbErr := helpers.MysqlClientDemo.Ctx(ctx).Create(demo).Error
   if dbErr != nil {
      err := components.ErrorDbInsert.WrapPrintf(dbErr, "demo insert error, input: %+v", demo)
      return err
   }
 
   return nil
}
```
另外，虽然框架提供了 NewError()/NewBaseError() 方法，但是不建议使用，推荐在 `components/error.go` 中统一定义，避免error分散在各个文件中，不便查找。
更不推荐直接使用原生的 `errors.New()`  直接new一个新的非base.Error类型的错误。



## 终止类错误记录错误栈
终止类错误一般在业务最底层调用时发生，比如api层调用下游接口、model层调用db、或者调用其他第三方库redis/es/cos等。

### 为什么这样做
当业务逻辑很复杂的时候，可以通过堆栈信息来方便开发查找调用路径以便快速定位问题。

### 如何做

1. 首先在`components/error.go`中为其定一个具体的error
2. 然后利用`golib/base/error.go`中的`Wrap()`/`WrapPrint()`/`WrapPrintf()` 记录错误栈。

函数说明：
* Wrap() ：不修改原始错误，记录错误栈并将 Error.ErrMsg ( components.ErrorApiGetUserCourseV1.ErrMsg) 作为追加信息；
* WrapPrint /  WrapPrintf()：记录错误栈，使用 Error.ErrMsg 格式化原始错误信息，并将新的入参 message 作为追加信息。

示例：
```go
res, err := conf.Api.Demo.HttpGet(ctx, pathGetUserCourseV1, opt)
 if err != nil {
     // Wrap 示例
     err = components.ErrorApiGetUserCourseV1.Wrap(err)
     // WrapPrint 示例
     // err = components.ErrorApiGetUserCourseV1.WrapPrint(err, "httpGet error")
     return info, err
 }
```

**效果展示**
当HttpGet() 连接超时时，假设返回的原始错误信息是：`giving up after 1 attempt(s): Get http://127.0.0.1:8081/user/v1/getUserCourse?userID=tom: dial tcp 127.0.0.1:8081: connect: connection refused`
那么使用 Wrap() 封装后返回的错误如下，errNo使用用户定义的，errMsg使用原始错误信息：

```json
{
  "errNo": 5102000,
  "errMsg": "giving up after 1 attempt(s): Get http://127.0.0.1:8081/user/v1/getUserCourse?userID=tom: dial tcp 127.0.0.1:8081: connect: connection refused",
  "data": {}
}
```

使用 WrapPrint /  WrapPrintf() 后错展示的信息如下（利用components.ErrorApiGetUserCourseV1 对原始错误信息进行了格式化）：
```json
{
  "errNo": 5102000,
  "errMsg": "call getUserCourse error: giving up after 1 attempt(s): Get http://127.0.0.1:8081/user/v1/getUserCourse?userID=tom: dial tcp 127.0.0.1:8081: connect: connection refused",
  "data": {}
}
```


## 错误逐层传递时，层层追加日志
这里借助错误库 "github.com/pkg/errors"  的errors.WithMessage[f](err error, message string) 接口实现。
该接口会以传入的err 为核心，附加绑定message的错误说明信息。

附加绑定的message一般用于记录调用下游函数的入参。

### 为什么这样做
这样做的好处是可以实现把用户追加的日志自动记录错误日志。
之前我们遇到错误，为了方便排查问题，一般这样做：

```go
desc := "update des"
err := models.UpdateGormDemoDescByName(ctx, name, desc)
if err != nil {
    zlog.Warn(ctx, "[models.UpdateGormDemoDescByName] error, name=%s, desc=%s", name, desc)
    return err
}
```
如果忘记了打印debug，当 models.UpdateGormDemoDescByName() 发生错误时，将很难复现问题（这里是一个简单的示例，如果下游是类似创建订单这样逻辑比较多的接口，将会更难处理）。

使用 errors.WithMessagef() 追加错误信息后，无需特意在返回error的地方打印warn/error 日志，框架会自动打印出用户追加的错误信息。

当发生错误时，可以根据错误栈中记录的信息进行单元测试进而可以快速复现问题。


### 如何做
当底层向上抛出错误时，不要直接返回错误，使用 errors.WithMessage[f] 追加错误后返回：

```go
func Foo(ctx *gin.Context) error {
	userID := "6379"
	info, err := api.GetUserCourseV1(ctx, userID)
	if err != nil {
		return errors.WithMessagef(err, "userID: %s", userID)
	}

	// do something
	return nil
}
```

使用这种方式抛给上游的错误不用担心err信息变得不可读，使用 base.RenderJsonFail(ctx, err) 输出的错误只会包含cause 错误（即下层函数返回的error信息），
用户追加的message只是作为附加信息打印在日志中。

---

**日志展示**
下面给出了一个遵守以上规范，当发生错误时打印的日志示例：

```bash
{"level":"error","module":"errorstack","requestId":"3889369160","time":"2020-10-13 16:35:01"}
-------------------stack-start-------------------
db update error: Error 1146: Table 'demo.demo1' doesn't exist
demo UpdateGormDemoDescByName error, input: lee, update des
permission/pkg/golib/v2/base.Error.WrapPrint
    /Users/jiangshuai/Go/pkg/mod/permission/pkg/golib/v2@v2.0.2-0.20201013040232-2bc9e949b471/base/error.go:55
permission/models.UpdateGormDemoDescByName
    /Users/jiangshuai/GO_CODE/docker-frame/permission/models/gorm_demo.go:69
permission/data.Test3
     
....
 
    /Users/jiangshuai/Go/pkg/mod/permission/pkg/gin@v1.0.0/gin.go:361
net/http.serverHandler.ServeHTTP
    /usr/local/go/src/net/http/server.go:2802
net/http.(*conn).serve
    /usr/local/go/src/net/http/server.go:1890
runtime.goexit
    /usr/local/go/src/runtime/asm_amd64.s:1357
Test3 UpdateGormDemoDescByName, input: name=lee desc=update des
Test2, input: name=lee
Test, input: name=lee
-------------------stack-end-------------------
```

`errorstack` 一行json日志是错误栈开始的信息，这里可以根据`requestId` 查到。后续可以查找该行日志后的N行记录来查看详细信息:
`stack-start`和`stack-end`之间多处的错误信息含义：

* 第一行 是函数发生错误的原始错误信息
* 第二行 发生错误时，用户追加的错误说明
* 第三部分 是错误栈信息，可以详细的找到错误调用的异常信息
* 第四部分 如果多个层级调用，会按照调用栈的顺序依次打印用户层层追加的错误信息说明
