# Team Todo - チーム向けタスク管理アプリ

Asanaライクなチーム向けタスク管理アプリケーション

## 技術スタック

### バックエンド
- **言語**: Go 1.25
- **フレームワーク**: Echo v4
- **ORM**: ent (entgo.io)
- **認証**: JWT (JSON Web Token)
- **ホットリロード**: cosmtrek/air

### フロントエンド
- **ランタイム**: Node.js 24
- **パッケージマネージャー**: pnpm
- **フレームワーク**: Next.js 15 (App Router)
- **スタイリング**: Tailwind CSS
- **言語**: TypeScript
- **テーマ**: ダークモード (Asanaライク)

### データベース
- **RDBMS**: PostgreSQL 16

### メール送信
- **サービス**: Resend API

## プロジェクト構成

```
team-todo/
├── backend/
│   ├── ent/
│   │   └── schema/           # entスキーマ定義
│   │       ├── user.go
│   │       ├── organization.go
│   │       ├── project.go
│   │       ├── organization_member.go
│   │       ├── project_member.go
│   │       └── invite.go
│   ├── internal/
│   │   ├── auth/             # 認証関連
│   │   │   ├── jwt.go
│   │   │   ├── middleware.go
│   │   │   └── password.go
│   │   ├── handler/          # APIハンドラー
│   │   │   ├── auth.go
│   │   │   ├── organization.go
│   │   │   ├── project.go
│   │   │   └── context.go
│   │   └── service/          # サービス層
│   │       └── email.go
│   ├── main.go
│   ├── Dockerfile
│   └── entrypoint.sh
├── frontend/
│   ├── app/
│   │   ├── login/            # ログインページ
│   │   ├── signup/           # 登録ページ
│   │   ├── invite/[token]/   # 招待承認ページ
│   │   ├── org/
│   │   │   ├── new/          # 組織作成ページ
│   │   │   └── [slug]/       # 組織ダッシュボード
│   │   │       └── projects/
│   │   │           └── [project_id]/  # プロジェクトページ
│   │   ├── layout.tsx
│   │   ├── page.tsx
│   │   └── globals.css
│   ├── components/
│   │   ├── ui/               # 共通UIコンポーネント
│   │   │   ├── button.tsx
│   │   │   └── input.tsx
│   │   └── layout/           # レイアウトコンポーネント
│   │       ├── sidebar.tsx
│   │       └── create-project-modal.tsx
│   ├── lib/
│   │   ├── api.ts            # API クライアント
│   │   ├── auth.tsx          # 認証コンテキスト
│   │   └── utils.ts          # ユーティリティ
│   └── package.json
└── docker-compose.yml
```

## 実装済み機能

### Phase 1: ユーザー認証
- ✅ ユーザー登録 (メール、パスワード、表示名)
- ✅ ログイン/ログアウト
- ✅ JWT認証 (アクセストークン + リフレッシュトークン)
- ✅ パスワードハッシュ化 (bcrypt)

### Phase 2: 組織管理
- ✅ 組織作成 (名前、URLスラッグ)
- ✅ 組織一覧取得
- ✅ 組織詳細取得
- ✅ メンバー招待 (メール通知)
- ✅ 招待承認
- ✅ ロール管理 (owner, admin, member)

### Phase 3: プロジェクト管理
- ✅ プロジェクト作成 (公開/非公開)
- ✅ プロジェクト一覧取得
- ✅ プロジェクト詳細取得
- ✅ プロジェクトメンバー追加
- ✅ 権限管理 (edit, view)

### Phase 4: コンテキスト復元
- ✅ 最終アクセス組織/プロジェクトの保存
- ✅ ログイン時の自動リダイレクト
- ✅ URLベースの状態管理

## 起動方法

### 初回セットアップ

```bash
# フロントエンドの依存関係をインストール
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

## API エンドポイント

### 認証 (Public)
| メソッド | パス | 説明 |
|----------|------|------|
| POST | `/api/v1/auth/register` | ユーザー登録 |
| POST | `/api/v1/auth/login` | ログイン |
| POST | `/api/v1/auth/refresh` | トークンリフレッシュ |

### 招待 (Public)
| メソッド | パス | 説明 |
|----------|------|------|
| GET | `/api/v1/invites/:token` | 招待情報取得 |

### ユーザー (Protected)
| メソッド | パス | 説明 |
|----------|------|------|
| GET | `/api/v1/me` | 現在のユーザー取得 |
| PATCH | `/api/v1/me` | ユーザー情報更新 |

### コンテキスト (Protected)
| メソッド | パス | 説明 |
|----------|------|------|
| GET | `/api/v1/context` | 現在のコンテキスト取得 |
| PUT | `/api/v1/context` | コンテキスト更新 |

### 組織 (Protected)
| メソッド | パス | 説明 |
|----------|------|------|
| POST | `/api/v1/organizations` | 組織作成 |
| GET | `/api/v1/organizations` | 組織一覧 |
| GET | `/api/v1/organizations/:slug` | 組織詳細 |
| POST | `/api/v1/organizations/:slug/invites` | メンバー招待 |
| POST | `/api/v1/invites/:token/accept` | 招待承認 |

### プロジェクト (Protected)
| メソッド | パス | 説明 |
|----------|------|------|
| POST | `/api/v1/organizations/:slug/projects` | プロジェクト作成 |
| GET | `/api/v1/organizations/:slug/projects` | プロジェクト一覧 |
| GET | `/api/v1/organizations/:slug/projects/:id` | プロジェクト詳細 |
| POST | `/api/v1/organizations/:slug/projects/:id/members` | メンバー追加 |

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
| JWT_SECRET | (開発用デフォルト) | JWTシークレットキー |
| RESEND_API_KEY | re_test_key | Resend APIキー |
| EMAIL_FROM | noreply@example.com | 送信元メールアドレス |
| APP_URL | http://localhost:3000 | アプリケーションURL |

### フロントエンド
| 変数名 | デフォルト値 | 説明 |
|--------|-------------|------|
| NEXT_PUBLIC_API_URL | http://localhost:8080 | バックエンドAPIのURL |

## データベース構造

### ER図

```
Users
├── id (UUID, PK)
├── email (Unique)
├── password_hash
├── display_name
├── last_org_id (FK → Organizations)
└── last_project_id (FK → Projects)

Organizations
├── id (UUID, PK)
├── name
└── slug (Unique)

Projects
├── id (UUID, PK)
├── organization_id (FK → Organizations)
├── name
└── is_private

Organization_Members
├── user_id (FK → Users)
├── organization_id (FK → Organizations)
└── role (owner/admin/member)

Project_Members
├── user_id (FK → Users)
├── project_id (FK → Projects)
└── permission (edit/view)

Invites
├── id (UUID, PK)
├── token (Unique)
├── email
├── organization_id (FK → Organizations)
├── project_id (FK → Projects, Nullable)
├── role
├── created_by_id (FK → Users)
├── expires_at
└── used_at (Nullable)
```

## 今後の実装予定

- [ ] タスクCRUD API
- [ ] タスクの担当者割り当て
- [ ] タスクの期限・優先度設定
- [ ] ボードビュー (カンバン)
- [ ] タイムラインビュー
- [ ] 通知機能
- [ ] コメント機能
- [ ] ファイル添付
