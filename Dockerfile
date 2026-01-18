# -----------------------------
# 1️⃣ Builder Stage
# -----------------------------
FROM registry.cn-hangzhou.aliyuncs.com/devflow/golang:1.25.6 AS builder

WORKDIR /app

ENV GOPROXY=https://goproxy.cn,direct

# 安装构建依赖
#RUN apk add --no-cache git curl build-base

# 安装 swag（自动生成 swagger 文档）
RUN go install github.com/swaggo/swag/cmd/swag@latest

# 提前复制 go.mod / go.sum，提高缓存命中率
COPY go.mod go.sum ./
RUN go mod download

# 复制整个项目
COPY . .

# 生成 swagger docs（docs/）
RUN swag init -g cmd/main.go --parseDependency

# 编译 DevFlow 主程序
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o devflow cmd/main.go

# -----------------------------
# 2️⃣ Runtime Stage
# -----------------------------
FROM alpine:3.19

WORKDIR /app

# 复制二进制和 swagger 文档
COPY --from=builder /app/devflow .
COPY --from=builder /app/docs ./docs

# 创建非 root 用户
RUN adduser -D devuser
USER devuser

EXPOSE 8080

ENTRYPOINT ["./devflow"]