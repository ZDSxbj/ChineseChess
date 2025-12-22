# 使用多阶段构建，减小镜像体积
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
# 使用国内代理可加速；若不需要可删除 GOPROXY 行
RUN go env -w GOPROXY=https://goproxy.cn,direct && go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o main .

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/main /app/main
# config.json 将通过 docker-compose 以只读方式挂载到 /app/config.json
EXPOSE 8080
ENV GIN_MODE=release
CMD ["/app/main"]