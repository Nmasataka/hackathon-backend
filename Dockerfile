# ビルドステージ
FROM golang:1.23 AS build

WORKDIR /app

# 依存ファイルを先にコピー
COPY go.mod go.sum ./

# 依存関係の解決
RUN go mod tidy

# ソースコードをコピー
COPY . .

# バイナリをビルド
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /main-app

# 実行ステージ
FROM debian:bullseye-slim



WORKDIR /app

RUN apt-get update && apt-get install ca-certificates openssl

# ビルドしたバイナリをコピー
COPY --from=build /main-app .

# アプリケーションのポートを公開
EXPOSE 8000

# アプリケーションを起動
CMD ["./main-app"]