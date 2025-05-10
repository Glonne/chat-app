# 使用官方 Go 1.23 镜像作为构建阶段
FROM golang:1.23 AS builder

# 设置工作目录为 /app
WORKDIR /app

# 设置 Go 模块代理
RUN go env -w GOPROXY=https://goproxy.cn,direct

# 复制 go.mod 和 go.sum 文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制整个项目文件
COPY . .

# 构建应用，禁用 CGO 并指定输出文件
RUN CGO_ENABLED=0 GOOS=linux go build -o chat-app ./main.go

# 使用轻量级 Alpine 镜像作为运行阶段
FROM alpine:latest

# 设置工作目录为 /home/golone
WORKDIR /home/golone

# 创建 chat 文件夹并切换到该目录
RUN mkdir -p chat
WORKDIR /home/golone/chat

# 从构建阶段复制可执行文件
COPY --from=builder /app/chat-app .

# 复制静态文件
COPY --from=builder /app/static ./static

# 暴露端口
EXPOSE 8080

# 设置默认环境变量
ENV MYSQL_DSN="root:Bixilong5201!@tcp(localhost:3306)/chat_app?charset=utf8mb4&parseTime=True&loc=Local"
ENV REDIS_ADDR="localhost:6379"
ENV RABBITMQ_URL="amqp://guest:guest@localhost:5672/"

# 运行应用
CMD ["./chat-app"]