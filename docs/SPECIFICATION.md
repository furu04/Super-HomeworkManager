# Super Homework Manager 仕様書

## 1. 概要

Super Homework Managerは、学生の課題管理を支援するWebアプリケーションです。Go言語とGinフレームワークを使用して構築されており、課題の登録・管理、期限追跡、完了状況の管理機能を提供します。

### 1.1 技術スタック

| 項目 | 技術 |
|------|------|
| 言語 | Go |
| Webフレームワーク | Gin |
| データベース | SQLite (GORM with Pure Go driver - glebarez/sqlite) |
| セッション管理 | gin-contrib/sessions (Cookie store) |
| テンプレートエンジン | Go html/template |
| コンテナ | Docker対応 |

### 1.2 ディレクトリ構成

```
homework-manager/
├── cmd/server/           # アプリケーションエントリポイント
├── internal/
│   ├── config/           # 設定読み込み
│   ├── database/         # データベース接続・マイグレーション
│   ├── handler/          # HTTPハンドラ
│   ├── middleware/       # ミドルウェア
│   ├── models/           # データモデル
│   ├── repository/       # データアクセス層
│   └── service/          # ビジネスロジック
├── web/
│   ├── static/           # 静的ファイル (CSS, JS)
│   └── templates/        # HTMLテンプレート
├── Dockerfile
└── docker-compose.yml
```

---

## 2. データモデル

### 2.1 User（ユーザー）

ユーザー情報を管理するモデル。

| フィールド | 型 | 説明 | 制約 |
|------------|------|------|------|
| ID | uint | ユーザーID | Primary Key |
| Email | string | メールアドレス | Unique, Not Null |
| PasswordHash | string | パスワードハッシュ | Not Null |
| Name | string | 表示名 | Not Null |
| Role | string | 権限 (`user` or `admin`) | Default: `user` |
| CreatedAt | time.Time | 作成日時 | 自動設定 |
| UpdatedAt | time.Time | 更新日時 | 自動更新 |
| DeletedAt | gorm.DeletedAt | 論理削除日時 | ソフトデリート |

### 2.2 Assignment（課題）

課題情報を管理するモデル。

| フィールド | 型 | 説明 | 制約 |
|------------|------|------|------|
| ID | uint | 課題ID | Primary Key |
| UserID | uint | 所有ユーザーID | Not Null, Index |
| Title | string | 課題タイトル | Not Null |
| Description | string | 説明 | - |
| Subject | string | 教科・科目 | - |
| Priority | string | 重要度 (`low`, `medium`, `high`) | Default: `medium` |
| DueDate | time.Time | 提出期限 | Not Null |
| IsCompleted | bool | 完了フラグ | Default: false |
| IsArchived | bool | アーカイブフラグ | Default: false |
| CompletedAt | *time.Time | 完了日時 | Nullable |
| ReminderEnabled | bool | 1回リマインダー有効 | Default: false |
| ReminderAt | *time.Time | リマインダー通知日時 | Nullable |
| ReminderSent | bool | リマインダー送信済み | Default: false |
| UrgentReminderEnabled | bool | 督促通知有効 | Default: true |
| LastUrgentReminderSent | *time.Time | 最終督促通知日時 | Nullable |
| CreatedAt | time.Time | 作成日時 | 自動設定 |
| UpdatedAt | time.Time | 更新日時 | 自動更新 |
| DeletedAt | gorm.DeletedAt | 論理削除日時 | ソフトデリート |


### 2.3 RecurringAssignment（繰り返し課題）

繰り返し課題の設定を管理するモデル。

| フィールド | 型 | 説明 | 制約 |
|------------|------|------|------|
| ID | uint | 設定ID | Primary Key |
| UserID | uint | 所有ユーザーID | Not Null, Index |
| Title | string | 課題タイトル | Not Null |
| Description | string | 説明 | - |
| Subject | string | 教科・科目 | - |
| Priority | string | 重要度 | Default: `medium` |
| RecurrenceType | string | 繰り返しタイプ (`daily`, `weekly`, `monthly`) | Not Null |
| RecurrenceInterval | int | 繰り返し間隔 | Default: 1 |
| RecurrenceWeekday | *int | 曜日 (0-6, 日-土) | Nullable |
| RecurrenceDay | *int | 日 (1-31) | Nullable |
| DueTime | string | 締切時刻 (HH:MM) | Not Null |
| EndType | string | 終了条件 (`never`, `count`, `date`) | Default: `never` |
| EndCount | *int | 終了回数 | Nullable |
| EndDate | *time.Time | 終了日 | Nullable |
| IsActive | bool | 有効フラグ | Default: true |
| CreatedAt | time.Time | 作成日時 | 自動設定 |
| UpdatedAt | time.Time | 更新日時 | 自動更新 |
| DeletedAt | gorm.DeletedAt | 論理削除日時 | ソフトデリート |

### 2.4 UserNotificationSettings（通知設定）

ユーザーの通知設定を管理するモデル。

| フィールド | 型 | 説明 | 制約 |
|------------|------|------|------|
| ID | uint | 設定ID | Primary Key |
| UserID | uint | ユーザーID | Unique, Not Null |
| TelegramEnabled | bool | Telegram通知 | Default: false |
| TelegramChatID | string | Telegram Chat ID | - |
| LineEnabled | bool | LINE通知 | Default: false |
| LineNotifyToken | string | LINE Notifyトークン | - |
| CreatedAt | time.Time | 作成日時 | 自動設定 |
| UpdatedAt | time.Time | 更新日時 | 自動更新 |
| DeletedAt | gorm.DeletedAt | 論理削除日時 | ソフトデリート |

### 2.5 APIKey（APIキー）

REST API認証用のAPIキーを管理するモデル。

| フィールド | 型 | 説明 | 制約 |
|------------|------|------|------|
| ID | uint | APIキーID | Primary Key |
| UserID | uint | 所有ユーザーID | Not Null, Index |
| Name | string | キー名 | Not Null |
| KeyHash | string | キーハッシュ | Unique, Not Null |
| LastUsed | *time.Time | 最終使用日時 | Nullable |
| CreatedAt | time.Time | 作成日時 | 自動設定 |
| DeletedAt | gorm.DeletedAt | 論理削除日時 | ソフトデリート |

---

## 3. 認証・認可

### 3.1 Web認証

- **セッションベース認証**: Cookie storeを使用
- **セッション有効期限**: 7日間
- **パスワード要件**: 8文字以上
- **パスワードハッシュ**: bcryptを使用
- **CSRF対策**: 全フォームでのトークン検証

### 3.2 API認証

- **APIキー認証**: `Authorization: Bearer <API_KEY>` ヘッダーで認証
- **キー形式**: `hm_` プレフィックス + 32文字のランダム文字列
- **ハッシュ保存**: SHA-256でハッシュ化して保存

### 3.3 ユーザーロール

| ロール | 権限 |
|--------|------|
| `user` | 自分の課題のCRUD操作、プロフィール管理 |
| `admin` | 全ユーザー管理、APIキー管理、ユーザー権限の変更 |
※ 最初に登録されたユーザーには自動的に `admin` 権限が付与されます。2人目以降は `user` として登録されます。

---

## 4. 機能一覧

### 4.1 認証機能

| 機能 | 説明 |
|------|------|
| 新規登録 | メールアドレス、パスワード、名前で登録 |
| ログイン | メールアドレスとパスワードでログイン |
| ログアウト | セッションをクリアしてログアウト |

### 4.2 課題管理機能

| 機能 | 説明 |
|------|------|
| ダッシュボード | 課題の統計情報、本日期限の課題、期限切れ課題、今週期限の課題を表示。各統計カードをクリックすると対応するフィルタで課題一覧に遷移 |
| 課題一覧 | フィルタ付き（未完了/今日が期限/今週が期限/完了済み/期限切れ）で課題を一覧表示 |
| 課題登録 | タイトル、説明、教科、重要度、提出期限、通知設定を入力して新規登録 |
| 課題編集 | 既存の課題情報を編集 |
| 課題削除 | 課題を論理削除（繰り返し課題に関連する場合、繰り返し設定ごと削除するか選択可能） |
| 完了トグル | 課題の完了/未完了状態を切り替え |
| 統計 | 科目別の完了率、期限内完了率等を表示 |

### 4.3 繰り返し課題機能

周期的に発生する課題を自動生成する機能。

| 機能 | 説明 |
|------|------|
| 繰り返し作成 | 課題登録時に繰り返し条件（毎日/毎週/毎月）を設定して作成 |
| 自動生成 | 未完了の課題がなくなったタイミングで、設定に基づき次回の課題を自動生成 |
| 繰り返し一覧 | 登録されている繰り返し設定を一覧表示 (`/recurring`) |
| 繰り返し編集 | 繰り返し設定の内容（タイトル、条件、時刻など）を編集 |
| 停止・再開 | 繰り返し設定を一時停止、または停止中の設定を再開 |
| 繰り返し削除 | 繰り返し設定を完全に削除 |

### 4.4 通知機能

#### 4.4.1 1回リマインダー

指定した日時に1回だけ通知を送信する機能。

| 項目 | 説明 |
|------|------|
| 設定 | 課題登録・編集画面で通知日時を指定 |
| 送信 | 指定日時にTelegram/LINEで通知 |

#### 4.4.2 督促通知

課題を完了するまで繰り返し通知を送信する機能。デフォルトで有効。

| 項目 | 説明 |
|------|------|
| 開始タイミング | 期限の **3時間前** |
| 重要度「大」 | **10分**ごとに通知 |
| 重要度「中」 | **30分**ごとに通知 |
| 重要度「小」 | **60分**ごとに通知 |
| 停止条件 | 課題の完了ボタンを押すまで継続 |

#### 4.4.3 通知チャンネル

| チャンネル | 設定方法 |
|------------|----------|
| Telegram | config.iniでBot Token設定、プロフィールでChat ID入力 |
| LINE Notify | プロフィールでアクセストークン入力 |

### 4.5 プロフィール機能

| 機能 | 説明 |
|------|------|
| プロフィール表示 | ユーザー情報を表示 |
| プロフィール更新 | 表示名を変更 |
| パスワード変更 | 現在のパスワードを確認後、新しいパスワードに変更 |
| 通知設定 | Telegram/LINE通知の有効化とトークン設定 |

### 4.6 管理者機能

| 機能 | 説明 |
|------|------|
| ユーザー一覧 | 全ユーザーを一覧表示 |
| ユーザー削除 | ユーザーを論理削除（自分自身は削除不可） |
| 権限変更 | ユーザーのロールを変更（自分自身は変更不可） |
| APIキー一覧 | 全APIキーを一覧表示 |
| APIキー発行 | 新規APIキーを発行（発行時のみ平文表示） |
| APIキー削除 | APIキーを削除 |

---

## 5. 設定

### 5.1 設定ファイル (config.ini)

アプリケーションは `config.ini` ファイルから設定を読み込みます。`config.ini.example` をコピーして使用してください。

```ini
[server]
port = 8080
debug = true

[database]
driver = sqlite
path = homework.db

[session]
secret = your-secure-secret-key

[auth]
allow_registration = true

[security]
https = false
csrf_secret = your-secure-csrf-secret
rate_limit_enabled = true
rate_limit_requests = 100
rate_limit_window = 60

[notification]
 telegram_bot_token = your-telegram-bot-token
```

### 5.2 設定項目

| セクション | キー | 説明 | デフォルト値 |
|------------|------|------|--------------|
| `server` | `port` | サーバーポート | `8080` |
| `server` | `debug` | デバッグモード | `true` |
| `database` | `driver` | DBドライバー (sqlite, mysql, postgres) | `sqlite` |
| `database` | `path` | SQLiteファイルパス | `homework.db` |
| `session` | `secret` | セッション暗号化キー | (必須) |
| `auth` | `allow_registration` | 新規登録許可 | `true` |
| `security` | `https` | HTTPS設定(Secure Cookie) | `false` |
| `security` | `csrf_secret` | CSRFトークン秘密鍵 | (必須) |
| `security` | `rate_limit_enabled` | レート制限有効化 | `true` |
| `security` | `rate_limit_requests` | 期間あたりの最大リクエスト数 | `100` |
| `security` | `rate_limit_window` | 期間（秒） | `60` |
| `notification` | `telegram_bot_token` | Telegram Bot Token | - |

### 5.3 環境変数

環境変数が設定されている場合、config.iniの設定より優先されます。

| 変数名 | 説明 |
|--------|------|
| `PORT` | サーバーポート |
| `DATABASE_DRIVER` | データベースドライバー |
| `DATABASE_PATH` | SQLiteデータベースファイルパス |
| `SESSION_SECRET` | セッション暗号化キー |
| `CSRF_SECRET` | CSRFトークン秘密鍵 |
| `GIN_MODE` | `release` でリリースモード（debug=false） |
| `ALLOW_REGISTRATION` | 新規登録許可 (true/false) |
| `HTTPS` | HTTPSモード (true/false) |
| `TRUSTED_PROXIES` | 信頼するプロキシのリスト |
| `TELEGRAM_BOT_TOKEN` | Telegram Bot Token |

### 5.4 設定の優先順位

1. 環境変数（最優先）
2. config.ini
3. デフォルト値

---

## 6. セキュリティ

### 6.1 実装済みセキュリティ機能

- **パスワードハッシュ化**: bcryptによるソルト付きハッシュ
- **セッションセキュリティ**: HttpOnly Cookie
- **入力バリデーション**: 各ハンドラで基本的な入力検証
- **CSFR対策**: Double Submit Cookieパターンまたは同期トークンによるCSRF保護
- **レート制限**: IPベースのリクエスト制限によるDoS対策
- **論理削除**: データの完全削除を防ぐソフトデリート
- **権限チェック**: ミドルウェアによるロールベースアクセス制御
- **Secure Cookie**: HTTPS設定時のSecure属性付与

### 6.2 推奨される本番環境設定

- `SESSION_SECRET` と `CSRF_SECRET` を強力なランダム文字列に変更
- HTTPSを有効化し、`HTTPS=true` を設定
- `GIN_MODE=release` を設定
- 必要に応じて `TRUSTED_PROXIES` を設定

---

## 7. ライセンス

AGPLv3 (GNU Affero General Public License v3)
