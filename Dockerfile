# 第一阶段，使用Go 1.18的alpine镜像作为基础镜像
FROM golang:1.18-alpine AS builder

# 定义应用名称并设置环境变量
ARG APP_NAME
ENV APP_NAME=$APP_NAME

# 添加go timezone插件
RUN apk update && apk add tzdata

# 设置环境
ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct

# 设置工作目录并拷贝依赖文件
WORKDIR $GOPATH/${APP_NAME}/
COPY go.mod $GOPATH/${APP_NAME}/
COPY go.sum $GOPATH/${APP_NAME}/

# 下载依赖并拷贝代码
RUN go mod download
COPY . $GOPATH/${APP_NAME}/

# 编译应用并设置版本信息
RUN go build -ldflags "-X 'main.goVersion=$(go version)' -X 'main.buildTime=$(date +%s)'" -o /usr/local/bin/permission main.go

# 第二阶段，使用最新版的alpine镜像作为基础镜像
FROM alpine:latest

# 定义应用名称并设置环境变量
ARG APP_NAME
ENV APP_NAME $APP_NAME

# 设置工作目录并拷贝应用及配置文件
WORKDIR /usr/local/bin/
COPY --from=builder /usr/local/bin/permission /usr/local/bin/
COPY ./conf /usr/local/bin/conf

# 添加timezone
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
ENV TZ=Asia/Shanghai

# 暴露端口
EXPOSE 8080

# 启动应用
CMD ["/usr/local/bin/permission"]