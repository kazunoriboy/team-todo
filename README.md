# Team Todo - Asana Clone

チーム向けタスク管理アプリケーション（Asanaクローン）

## 技術スタック

### バックエンド
- **言語**: Go 1.25
- **フレームワーク**: Echo v4
- **ORM**: ent (entgo.io)
- **ホットリロード**: cosmtrek/air

### フロントエンド
- **ランタイム**: Node.js 24
- **パッケージマネージャー**: pnpm
- **フレームワーク**: Next.js 15 (App Router)
- **スタイリング**: Tailwind CSS
- **言語**: TypeScript

### データベース
- **RDBMS**: PostgreSQL 16

## プロジェクト構成

```
team-todo/
├── backend/              # Go バックエンド
│   ├── Dockerfile
│   ├── .air.toml        # ホットリロード設定
│   ├── main.go
│   ├── go.mod
│   └── go.sum
├── frontend/            # Next.js フロントエンド
│   ├── Dockerfile
│   ├── app/             # App Router
│   │   ├── layout.tsx
│   │   ├── page.tsx
│   │   └── globals.css
│   ├── package.json
│   ├── tailwind.config.ts
│   └── tsconfig.json
├── docker-compose.yml   # Docker構成
├── .gitignore
└── README.md
```

## 起動方法

### 初回セットアップ

```bash
# フロントエンドの依存関係をインストール（初回のみ）
cd frontend
pnpm install
cd ..
```

### Docker Composeで起動

```bash
# 全サービスを起動
docker compose up --build

# バックグラウンドで起動
docker compose up -d --build
```

### アクセス

- **フロントエンド**: http://localhost:3000
- **バックエンドAPI**: http://localhost:8080
- **ヘルスチェック**: http://localhost:8080/health

### 開発時の操作

```bash
# ログを確認
docker compose logs -f

# 特定のサービスのログを確認
docker compose logs -f backend
docker compose logs -f frontend

# 停止
docker compose down

# データベースを含めて完全にリセット
docker compose down -v
```

## API エンドポイント

| メソッド | パス | 説明 |
|----------|------|------|
| GET | `/health` | ヘルスチェック |
| GET | `/api/v1/` | API ルート |

## 環境変数

### バックエンド

| 変数名 | デフォルト値 | 説明 |
|--------|-------------|------|
| DB_HOST | localhost | データベースホスト |
| DB_PORT | 5432 | データベースポート |
| DB_USER | postgres | データベースユーザー |
| DB_PASSWORD | postgres | データベースパスワード |
| DB_NAME | team_todo | データベース名 |
| PORT | 8080 | サーバーポート |

### フロントエンド

| 変数名 | デフォルト値 | 説明 |
|--------|-------------|------|
| NEXT_PUBLIC_API_URL | http://localhost:8080 | バックエンドAPIのURL |

## 今後の実装予定

- [ ] ent スキーマ定義（User, Project, Task）
- [ ] 認証機能（JWT）
- [ ] タスクCRUD API
- [ ] プロジェクト管理機能
- [ ] フロントエンド画面実装
- [ ] チームメンバー招待機能
