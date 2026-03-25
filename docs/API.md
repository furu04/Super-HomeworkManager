# Super Homework Manager API ドキュメント

## 概要

Super Homework Manager REST APIは、課題管理機能をプログラムから利用するためのAPIです。

- **ベースURL**: `/api/v1`
- **認証方式**: APIキー認証
- **レスポンス形式**: JSON

---

## 認証

すべてのAPIエンドポイントはAPIキー認証が必要です。

### APIキーの取得

1. 管理者アカウントでログイン
2. 管理画面 → APIキー管理へ移動
3. 新規APIキーを発行

### 認証ヘッダー

```
Authorization: Bearer hm_XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
```

### 認証エラー

| ステータスコード | レスポンス |
|------------------|------------|
| 401 Unauthorized | `{"error": "Authorization header required"}` |
| 401 Unauthorized | `{"error": "Invalid authorization format. Use: Bearer <api_key>"}` |
| 401 Unauthorized | `{"error": "Invalid API key"}` |

---

## エンドポイント一覧

| メソッド | パス | 説明 |
|----------|------|------|
| GET | `/api/v1/assignments` | 課題一覧取得（フィルタ・ページネーション対応） |
| GET | `/api/v1/assignments/pending` | 未完了課題一覧取得 |
| GET | `/api/v1/assignments/completed` | 完了済み課題一覧取得 |
| GET | `/api/v1/assignments/overdue` | 期限切れ課題一覧取得 |
| GET | `/api/v1/assignments/due-today` | 本日期限の課題一覧取得 |
| GET | `/api/v1/assignments/due-this-week` | 今週期限の課題一覧取得 |
| GET | `/api/v1/assignments/:id` | 課題詳細取得 |
| POST | `/api/v1/assignments` | 課題作成 |
| PUT | `/api/v1/assignments/:id` | 課題更新 |
| DELETE | `/api/v1/assignments/:id` | 課題削除 |
| PATCH | `/api/v1/assignments/:id/toggle` | 完了状態トグル |
| GET | `/api/v1/statistics` | 統計情報取得 |
| GET | `/api/v1/recurring` | 繰り返し設定一覧取得 |
| GET | `/api/v1/recurring/:id` | 繰り返し設定詳細取得 |
| PUT | `/api/v1/recurring/:id` | 繰り返し設定更新 |
| DELETE | `/api/v1/recurring/:id` | 繰り返し設定削除 |

---

## 課題一覧取得

```
GET /api/v1/assignments
```

### クエリパラメータ

| パラメータ | 型 | 説明 |
|------------|------|------|
| `filter` | string | フィルタ: `pending`, `completed`, `overdue`（省略時: 全件） |
| `page` | integer | ページ番号（デフォルト: `1`） |
| `page_size` | integer | 1ページあたりの件数（デフォルト: `20`、最大: `100`） |

### レスポンス

**200 OK**

```json
{
  "assignments": [
    {
      "id": 1,
      "user_id": 1,
      "title": "数学レポート",
      "description": "第5章の練習問題",
      "subject": "数学",
      "priority": "medium",
      "due_date": "2025-01-15T23:59:00+09:00",
      "is_completed": false,
      "created_at": "2025-01-10T10:00:00+09:00",
      "updated_at": "2025-01-10T10:00:00+09:00"
    }
  ],
  "count": 1,
  "total_count": 15,
  "total_pages": 1,
  "current_page": 1,
  "page_size": 20
}
```

### 例

```bash
# 全件取得
curl -H "Authorization: Bearer hm_xxx" http://localhost:8080/api/v1/assignments

# 未完了のみ（ページネーション付き）
curl -H "Authorization: Bearer hm_xxx" "http://localhost:8080/api/v1/assignments?filter=pending&page=1&page_size=10"

# 期限切れのみ
curl -H "Authorization: Bearer hm_xxx" "http://localhost:8080/api/v1/assignments?filter=overdue"
```

---

## 絞り込み済み課題一覧取得

専用エンドポイントでも同等の絞り込みができます（`pending` / `completed` / `overdue` はページネーション対応）。

```
GET /api/v1/assignments/pending
GET /api/v1/assignments/completed
GET /api/v1/assignments/overdue
GET /api/v1/assignments/due-today
GET /api/v1/assignments/due-this-week
```

### クエリパラメータ（pending / completed / overdue のみ）

| パラメータ | 型 | 説明 |
|------------|------|------|
| `page` | integer | ページ番号（デフォルト: `1`） |
| `page_size` | integer | 1ページあたりの件数（デフォルト: `20`、最大: `100`） |

### レスポンス

`due-today` / `due-this-week` は `count` のみ返します（ページネーションなし）。その他は `GET /api/v1/assignments` と同形式です。

### 例

```bash
curl -H "Authorization: Bearer hm_xxx" http://localhost:8080/api/v1/assignments/due-today
curl -H "Authorization: Bearer hm_xxx" http://localhost:8080/api/v1/assignments/due-this-week
```

---

## 課題詳細取得

```
GET /api/v1/assignments/:id
```

### パスパラメータ

| パラメータ | 型 | 説明 |
|------------|------|------|
| `id` | integer | 課題ID |

### レスポンス

**200 OK**

```json
{
  "id": 1,
  "user_id": 1,
  "title": "数学レポート",
  "description": "第5章の練習問題",
  "subject": "数学",
  "priority": "medium",
  "due_date": "2025-01-15T23:59:00+09:00",
  "is_completed": false,
  "created_at": "2025-01-10T10:00:00+09:00",
  "updated_at": "2025-01-10T10:00:00+09:00"
}
```

**404 Not Found**

```json
{
  "error": "Assignment not found"
}
```

### 例

```bash
curl -H "Authorization: Bearer hm_xxx" http://localhost:8080/api/v1/assignments/1
```

---

## 課題作成

```
POST /api/v1/assignments
```

### リクエストボディ

| フィールド | 型 | 必須 | 説明 |
|------------|------|------|------|
| `title` | string | ✅ | 課題タイトル |
| `description` | string | | 説明 |
| `subject` | string | | 教科・科目 |
| `priority` | string | | 重要度: `low`, `medium`, `high`（デフォルト: `medium`） |
| `due_date` | string | ✅ | 提出期限（RFC3339 または `YYYY-MM-DDTHH:MM` または `YYYY-MM-DD`） |
| `reminder_enabled` | boolean | | リマインダーを有効にするか（デフォルト: `false`） |
| `reminder_at` | string | | リマインダー設定時刻（形式は `due_date` と同じ） |
| `urgent_reminder_enabled` | boolean | | 督促リマインダーを有効にするか（デフォルト: `true`） |
| `recurrence` | object | | 繰り返し設定（下記参照） |

### Recurrence オブジェクト

| フィールド | 型 | 説明 |
|------------|------|------|
| `type` | string | 繰り返しタイプ: `daily`, `weekly`, `monthly`（空文字で繰り返しなし） |
| `interval` | integer | 繰り返し間隔（例: `1` = 毎週、`2` = 隔週） |
| `weekday` | integer | 週次の曜日（`0`=日, `1`=月, ..., `6`=土） |
| `day` | integer | 月次の日付（1-31） |
| `until` | object | 終了条件 |

#### Recurrence.Until オブジェクト

| フィールド | 型 | 説明 |
|------------|------|------|
| `type` | string | 終了タイプ: `never`, `count`, `date` |
| `count` | integer | 終了回数（`count` 指定時） |
| `date` | string | 終了日（`date` 指定時） |

### リクエスト例

```json
{
  "title": "英語エッセイ",
  "description": "テーマ自由、1000語以上",
  "subject": "英語",
  "priority": "high",
  "due_date": "2025-01-20T17:00"
}
```

### レスポンス

**201 Created**

```json
{
  "id": 2,
  "user_id": 1,
  "title": "英語エッセイ",
  "description": "テーマ自由、1000語以上",
  "subject": "英語",
  "priority": "high",
  "due_date": "2025-01-20T17:00:00+09:00",
  "is_completed": false,
  "created_at": "2025-01-10T11:00:00+09:00",
  "updated_at": "2025-01-10T11:00:00+09:00"
}
```

繰り返し設定を含む場合は `recurring_assignment` を返します:

```json
{
  "message": "Recurring assignment created",
  "recurring_assignment": { ... }
}
```

**400 Bad Request**

```json
{
  "error": "Invalid input: Key: 'title' Error:Field validation for 'title' failed on the 'required' tag"
}
```

### 例

```bash
curl -X POST \
  -H "Authorization: Bearer hm_xxx" \
  -H "Content-Type: application/json" \
  -d '{"title":"英語エッセイ","subject":"英語","due_date":"2025-01-20"}' \
  http://localhost:8080/api/v1/assignments
```

---

## 課題更新

```
PUT /api/v1/assignments/:id
```

### パスパラメータ

| パラメータ | 型 | 説明 |
|------------|------|------|
| `id` | integer | 課題ID |

### リクエストボディ

すべてのフィールドはオプションです。省略されたフィールドは既存の値を維持します。

| フィールド | 型 | 説明 |
|------------|------|------|
| `title` | string | 課題タイトル |
| `description` | string | 説明 |
| `subject` | string | 教科・科目 |
| `priority` | string | 重要度: `low`, `medium`, `high` |
| `due_date` | string | 提出期限 |
| `reminder_enabled` | boolean | リマインダー有効/無効 |
| `reminder_at` | string | リマインダー時刻 |
| `urgent_reminder_enabled` | boolean | 督促リマインダー有効/無効 |

### リクエスト例

```json
{
  "title": "英語エッセイ（修正版）",
  "due_date": "2025-01-25T17:00"
}
```

### レスポンス

**200 OK** — 更新後の課題オブジェクト

**404 Not Found**

```json
{
  "error": "Assignment not found"
}
```

### 例

```bash
curl -X PUT \
  -H "Authorization: Bearer hm_xxx" \
  -H "Content-Type: application/json" \
  -d '{"title":"更新されたタイトル"}' \
  http://localhost:8080/api/v1/assignments/2
```

---

## 課題削除

```
DELETE /api/v1/assignments/:id
```

### パスパラメータ

| パラメータ | 型 | 説明 |
|------------|------|------|
| `id` | integer | 課題ID |

### クエリパラメータ

| パラメータ | 型 | 説明 |
|------------|------|------|
| `delete_recurring` | boolean | `true` の場合、関連する繰り返し設定も削除する |

### レスポンス

**200 OK**

```json
{ "message": "Assignment deleted" }
```

繰り返し設定も削除した場合:

```json
{ "message": "Assignment and recurring settings deleted" }
```

**404 Not Found**

```json
{ "error": "Assignment not found" }
```

### 例

```bash
# 課題のみ削除
curl -X DELETE \
  -H "Authorization: Bearer hm_xxx" \
  http://localhost:8080/api/v1/assignments/2

# 課題と繰り返し設定をまとめて削除
curl -X DELETE \
  -H "Authorization: Bearer hm_xxx" \
  "http://localhost:8080/api/v1/assignments/2?delete_recurring=true"
```

---

## 完了状態トグル

課題の完了状態を切り替えます（未完了 ↔ 完了）。

```
PATCH /api/v1/assignments/:id/toggle
```

### パスパラメータ

| パラメータ | 型 | 説明 |
|------------|------|------|
| `id` | integer | 課題ID |

### レスポンス

**200 OK**

```json
{
  "id": 1,
  "user_id": 1,
  "title": "数学レポート",
  "description": "第5章の練習問題",
  "subject": "数学",
  "priority": "medium",
  "due_date": "2025-01-15T23:59:00+09:00",
  "is_completed": true,
  "completed_at": "2025-01-12T14:30:00+09:00",
  "created_at": "2025-01-10T10:00:00+09:00",
  "updated_at": "2025-01-12T14:30:00+09:00"
}
```

**404 Not Found**

```json
{ "error": "Assignment not found" }
```

### 例

```bash
curl -X PATCH \
  -H "Authorization: Bearer hm_xxx" \
  http://localhost:8080/api/v1/assignments/1/toggle
```

---

## 統計情報取得

ユーザーの課題統計を取得します。

```
GET /api/v1/statistics
```

### クエリパラメータ

| パラメータ | 型 | 説明 |
|------------|------|------|
| `subject` | string | 科目で絞り込み（省略時: 全科目） |
| `from` | string | 課題登録日の開始日（`YYYY-MM-DD`） |
| `to` | string | 課題登録日の終了日（`YYYY-MM-DD`） |
| `include_archived` | boolean | アーカイブ済み課題を含む（デフォルト: `false`） |

### レスポンス

**200 OK**

```json
{
  "total_assignments": 45,
  "completed_assignments": 30,
  "pending_assignments": 12,
  "overdue_assignments": 3,
  "on_time_completion_rate": 86.7,
  "filter": {
    "subject": null,
    "from": "2025-01-01",
    "to": "2025-12-31"
  },
  "subjects": [
    {
      "subject": "数学",
      "total": 15,
      "completed": 12,
      "pending": 2,
      "overdue": 1,
      "on_time_completion_rate": 91.7
    }
  ]
}
```

### 例

```bash
# 全体統計
curl -H "Authorization: Bearer hm_xxx" http://localhost:8080/api/v1/statistics

# 科目で絞り込み
curl -H "Authorization: Bearer hm_xxx" "http://localhost:8080/api/v1/statistics?subject=数学"

# 日付範囲で絞り込み
curl -H "Authorization: Bearer hm_xxx" "http://localhost:8080/api/v1/statistics?from=2025-01-01&to=2025-03-31"
```

---

## 繰り返し設定一覧取得

```
GET /api/v1/recurring
```

### レスポンス

**200 OK**

```json
{
  "recurring_assignments": [
    {
      "id": 1,
      "user_id": 1,
      "title": "週次ミーティング",
      "subject": "その他",
      "priority": "medium",
      "recurrence_type": "weekly",
      "recurrence_interval": 1,
      "recurrence_weekday": 1,
      "due_time": "23:59",
      "end_type": "never",
      "is_active": true,
      "created_at": "2025-01-01T00:00:00+09:00",
      "updated_at": "2025-01-01T00:00:00+09:00"
    }
  ],
  "count": 1
}
```

---

## 繰り返し設定詳細取得

```
GET /api/v1/recurring/:id
```

### パスパラメータ

| パラメータ | 型 | 説明 |
|------------|------|------|
| `id` | integer | 繰り返し設定ID |

### レスポンス

**200 OK** — 繰り返し設定オブジェクト（一覧と同形式）

**404 Not Found**

```json
{ "error": "Recurring assignment not found" }
```

---

## 繰り返し設定更新

```
PUT /api/v1/recurring/:id
```

### パスパラメータ

| パラメータ | 型 | 説明 |
|------------|------|------|
| `id` | integer | 繰り返し設定ID |

### リクエストボディ

すべてのフィールドはオプションです。省略されたフィールドは既存の値を維持します。

| フィールド | 型 | 説明 |
|------------|------|------|
| `title` | string | タイトル |
| `description` | string | 説明 |
| `subject` | string | 教科・科目 |
| `priority` | string | 重要度: `low`, `medium`, `high` |
| `recurrence_type` | string | 繰り返しタイプ: `daily`, `weekly`, `monthly` |
| `recurrence_interval` | integer | 繰り返し間隔 |
| `recurrence_weekday` | integer | 週次の曜日（0-6） |
| `recurrence_day` | integer | 月次の日付（1-31） |
| `due_time` | string | 締切時刻（`HH:MM`） |
| `end_type` | string | 終了タイプ: `never`, `count`, `date` |
| `end_count` | integer | 終了回数 |
| `end_date` | string | 終了日（`YYYY-MM-DD`） |
| `is_active` | boolean | `false` で停止、`true` で再開 |
| `reminder_enabled` | boolean | リマインダー有効/無効 |
| `reminder_offset` | integer | リマインダーのオフセット（分） |
| `urgent_reminder_enabled` | boolean | 督促リマインダー有効/無効 |
| `edit_behavior` | string | 編集範囲: `this_only`, `this_and_future`, `all`（デフォルト: `this_only`） |

### リクエスト例（一時停止）

```json
{
  "is_active": false
}
```

### レスポンス

**200 OK** — 更新後の繰り返し設定オブジェクト

**404 Not Found**

```json
{ "error": "Recurring assignment not found" }
```

### 例

```bash
# 一時停止
curl -X PUT \
  -H "Authorization: Bearer hm_xxx" \
  -H "Content-Type: application/json" \
  -d '{"is_active": false}' \
  http://localhost:8080/api/v1/recurring/1

# タイトルと締切時刻を変更
curl -X PUT \
  -H "Authorization: Bearer hm_xxx" \
  -H "Content-Type: application/json" \
  -d '{"title":"更新済みタスク","due_time":"22:00"}' \
  http://localhost:8080/api/v1/recurring/1
```

---

## 繰り返し設定削除

```
DELETE /api/v1/recurring/:id
```

### パスパラメータ

| パラメータ | 型 | 説明 |
|------------|------|------|
| `id` | integer | 繰り返し設定ID |

### レスポンス

**200 OK**

```json
{ "message": "Recurring assignment deleted" }
```

**404 Not Found**

```json
{ "error": "Recurring assignment not found or failed to delete" }
```

### 例

```bash
curl -X DELETE \
  -H "Authorization: Bearer hm_xxx" \
  http://localhost:8080/api/v1/recurring/1
```

---

## エラーレスポンス

すべてのエラーレスポンスは以下の形式で返されます：

```json
{
  "error": "エラーメッセージ"
}
```

### 共通エラーコード

| ステータスコード | 説明 |
|------------------|------|
| 400 Bad Request | リクエストの形式が不正 |
| 401 Unauthorized | 認証エラー |
| 404 Not Found | リソースが見つからない |
| 429 Too Many Requests | レート制限超過 |
| 500 Internal Server Error | サーバー内部エラー |

---

## 日付形式

APIは以下の日付形式を受け付けます（優先度順）：

1. **RFC3339**: `2025-01-15T23:59:00+09:00`
2. **日時形式**: `2025-01-15T23:59`
3. **日付のみ**: `2025-01-15`（時刻は `23:59` に設定）

レスポンスの日付はすべてRFC3339形式で返されます。

---

## Rate Limiting

アプリケーションレベルでのRate Limitingが実装されています。

- **制限単位**: IPアドレスごと
- **デフォルト制限**: 100リクエスト / 60秒
- **超過時レスポンス**: `429 Too Many Requests`

```json
{
  "error": "リクエスト数が制限を超えました。しばらくしてからお試しください。"
}
```

設定ファイル (`config.ini`) または環境変数で制限値を変更可能です。
