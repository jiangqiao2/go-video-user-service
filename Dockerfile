FROM golang:1.24-alpine AS builder
WORKDIR /app

# 安装证书和时区数据，避免拉取模块时网络/证书问题
RUN apk add --no-cache ca-certificates tzdata
# 使用国内 Go 模块代理
ENV GOPROXY=https://goproxy.cn,direct

# 先复制 go.mod / go.sum 并下载依赖，利用构建缓存
COPY go.mod go.sum ./
RUN go mod download

# 复制业务代码
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/user-service ./main.go

FROM alpine:3.19
WORKDIR /app

RUN apk add --no-cache tzdata && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo 'Asia/Shanghai' > /etc/timezone

COPY --from=builder /bin/user-service /usr/local/bin/user-service
# 从构建阶段复制配置
COPY --from=builder /app/configs ./configs

ARG CONFIG_PATH=/app/configs/config.dev.yaml
ENV CONFIG_PATH=${CONFIG_PATH}
EXPOSE 8081
ENTRYPOINT ["user-service"]
