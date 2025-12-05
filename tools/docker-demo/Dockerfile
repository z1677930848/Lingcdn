# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod .
RUN go mod download
COPY . .
# 为了得到更小镜像，构建时关闭 CGO 并指定目标系统
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o lingcdn-demo main.go

FROM alpine:3.20
WORKDIR /app
RUN adduser -D demo
COPY --from=builder /src/lingcdn-demo /app/lingcdn-demo
ENV PORT=8080 \
    DEMO_MESSAGE="Lingcdn Docker 示例服务运行中"
EXPOSE 8080
USER demo
ENTRYPOINT ["./lingcdn-demo"]
