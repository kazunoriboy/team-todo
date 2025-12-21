FROM golang:1.25-alpine

WORKDIR /app

# 必要なツールのインストール
RUN apk add --no-cache git

# air（ホットリロード）のインストール
RUN go install github.com/air-verse/air@latest

# 依存関係のダウンロード
COPY go.mod ./
RUN go mod tidy

COPY . .

# エントリーポイントスクリプトに実行権限を付与
RUN chmod +x entrypoint.sh

ENTRYPOINT ["./entrypoint.sh"]

