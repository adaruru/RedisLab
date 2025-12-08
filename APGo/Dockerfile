# 建置階段
FROM golang:1.24-alpine AS builder

WORKDIR /app

# 複製 go mod 檔案
COPY go.mod go.sum ./
RUN go mod download

# 複製原始碼
COPY . .

# 建置應用程式
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/apgo ./cmd/main.go

# 執行階段
FROM alpine:latest

WORKDIR /app

# 安裝 ca-certificates（用於 HTTPS）
RUN apk --no-cache add ca-certificates tzdata

# 從建置階段複製執行檔
COPY --from=builder /app/bin/apgo .

# 複製設定檔
COPY --from=builder /app/config*.yaml ./

# 設定時區
ENV TZ=Asia/Taipei

# 暴露端口
EXPOSE 8080

# 執行應用程式
CMD ["./apgo"]
