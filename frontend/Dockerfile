FROM node:24-alpine

# pnpmのインストール
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable

WORKDIR /app

# 依存関係のコピーとインストール
COPY package.json pnpm-lock.yaml* ./
RUN pnpm install || true

COPY . .

# Next.jsの開発サーバーを起動
CMD ["pnpm", "dev"]


