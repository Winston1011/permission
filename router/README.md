# 接口规范 - 路由

## 协议
内部接口建议统一使用 http 协议，如果涉及到接口中有加密信息，建议在服务测自己实现加解密逻辑。

## 方法
建议仅对外提供GET/POST 两种方法。当不确定用GET/POST时，可参考以下建议：

GET: 多适用于查询操作
POST:  当对资源的操作类型为修改/删除/新增时，常用场景为表单提交操作

## URI 
### URI命名规范
uri应能够清晰简要的表示出接口的含义，组成： 小写字母、数字、 横线
使用一致的资源命名约定和URI格式来实现最小的模糊性和最大的可读性和可维护性。可以遵循以下的设计提示去保持一致性：
* 不要在末尾使用正斜杠，作为URI的路径的最后一个字符，正斜杠没有意义而且可能会然人迷惑，最好是把它删除

```bash
http://api.example.com/device-management/managed-devices/
http://api.example.com/device-management/managed-devices    /*This is much better version*/
```

* 使用连接符号
为了使URIs让人更容易读和理解，在长路径中使用连字号（-）去增加名字的可读性
* 不要使用下划线符
在使用连字号`-`作为分隔符的情形下，使用下划线`_`字符也是很有可能的。但是这取决于应用的字体，在某些浏览器或屏幕中，
下划线`_`字符可能会被部分掩盖或完全隐藏。为了避免混淆，使用连字符`-`代替下划线符号`_`

```bash
http://api.example.com/inventory-management/managed-entities/{id}/install-script-location  //More readable
http://api.example.com/inventory_management/managed_entities/{id}/install_script_location  //More error prone
```

* 使用小写
如果可以，在URI路径中小写是被推荐的，[RFC 3986](https://www.ietf.org/rfc/rfc3986.txt) 规范中定义了URIs除了scheme和host部分其他是大小写敏感的。
```bash
http://api.example.org/my-folder/my-doc  //1
HTTP://API.EXAMPLE.ORG/my-folder/my-doc  //2
http://api.example.org/My-Folder/my-doc  //3
```
在上面，1和2是相同的，但是和3是不同的，因为3中My-Folder使用了首部大写

* 不要增加文件后缀
文件扩展看起来很不好并且没有任何优势，移除可以减少URIs的长度。没有任何理由去用它。
如果你想强调使用了文件扩展名的API的媒体类型，你应当用媒体类型，通过“Content-Type”请求头来交互，
去说明如何处理请求内容

```bash
http://api.example.com/device-management/managed-devices.xml  /*Do not use it*/
http://api.example.com/device-management/managed-devices    /*This is correct URI*/
```

* 使用查询组合来过滤URI集合
很多时候，你会遇到这些请求，你需要通过资源的某些属性去排序，过滤或者限制资源集合。
为了实现这个，不要生成一个新的api，可以直接在资源集合API中通过传入参数作为查询过滤、分页、排序等。
```bash
http://api.example.com/device-management/managed-devices
http://api.example.com/device-management/managed-devices?region=USA
http://api.example.com/device-management/managed-devices?region=USA&brand=XYZ
http://api.example.com/device-management/managed-devices?region=USA&brand=XYZ&sort=installation-date&size=10
```

### URI组成
 * 第一层级：模块名
 * 第二层级：逻辑功能模块，如果模块较为复杂，可以分多个层级
 * 最后一级：对接口功能的描述


## http状态码
默认返回 200

## 接口返回
返回内容应有两部分组成：
* error 层：需要有错误码和至少一个错误信息
* data 层 ：里面包含返回数据的完整信息
go服务中推荐直接使用框架封装好的方法渲染输出结果：
```bash
base.RenderJsonAbort(ctx, components.ErrorSystemError)
base.RenderJsonFail(ctx, err)
base.RenderJsonSucc(ctx, gin.H{})
```