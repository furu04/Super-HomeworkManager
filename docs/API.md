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
| GET | `/api/v1/assignments` | 課題一覧取得 |
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
| `filter` | string | フィルタ: `pending`, `completed`, `overdue` (省略時: 全件) |

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
      "due_date": "2025-01-15T23:59:00+09:00",
      "is_completed": false,
      "created_at": "2025-01-10T10:00:00+09:00",
      "updated_at": "2025-01-10T10:00:00+09:00"
    }
  ],
  "count": 1
}
```

### 例

```bash
# 全件取得
curl -H "Authorization: Bearer hm_xxx" http://localhost:8080/api/v1/assignments

# 未完了のみ取得
curl -H "Authorization: Bearer hm_xxx" http://localhost:8080/api/v1/assignments?filter=pending

# 期限切れのみ取得
curl -H "Authorization: Bearer hm_xxx" http://localhost:8080/api/v1/assignments?filter=overdue
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
| `due_date` | string | ✅ | 提出期限（RFC3339 または `YYYY-MM-DDTHH:MM` または `YYYY-MM-DD`） |
| `reminder_enabled` | boolean | | リマインダーを有効にするか（省略時: false） |
| `reminder_at` | string | | リマインダー設定時刻（形式はdue_dateと同じ） |
| `urgent_reminder_enabled` | boolean | | 期限切れ時の督促リマインダーを有効にするか（省略時: true） |
| `recurrence` | object | | 繰り返し設定（以下参照） |

### Recurrence オブジェクト

| フィールド | 型 | 説明 |
|------------|------|------|
| `type` | string | 繰り返しタイプ (`daily`, `weekly`, `monthly`, または空文字で無効) |
| `interval` | integer | 間隔 (例: 1 = 毎週, 2 = 隔週) |
| `weekday` | integer | 週次の曜日 (0=日, 1=月, ..., 6=土) |
| `day` | integer | 月次の日付 (1-31) |
| `until` | object | 終了条件 |

#### Recurrence.Until オブジェクト

| フィールド | 型 | 説明 |
|------------|------|------|
| `type` | string | 終了タイプ (`never`, `count`, `date`) |
| `count` | integer | 回数指定時の終了回数 |
| `date` | string | 日付指定時の終了日 |

### リクエスト例

```json
{
  "title": "英語エッセイ",
  "description": "テーマ自由、1000語以上",
  "subject": "英語",
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
  "due_date": "2025-01-20T17:00:00+09:00",
  "is_completed": false,
  "created_at": "2025-01-10T11:00:00+09:00",
  "updated_at": "2025-01-10T11:00:00+09:00"
}
```

**400 Bad Request**

```json
{
  "error": "Invalid input: title and due_date are required"
}
```

### 例

```bash
curl -X POST \
  -H "Authorization: Bearer hm_xxx" \
  -H "Content-Type: application/json" \
  -d '{"title":"英語エッセイ","due_date":"2025-01-20"}' \
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

**200 OK**

```json
{
  "id": 2,
  "user_id": 1,
  "title": "英語エッセイ（修正版）",
  "description": "テーマ自由、1000語以上",
  "subject": "英語",
  "due_date": "2025-01-25T17:00:00+09:00",
  "is_completed": false,
  "created_at": "2025-01-10T11:00:00+09:00",
  "updated_at": "2025-01-10T12:00:00+09:00"
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
{
  "message": "Assignment deleted"
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
curl -X DELETE \
  -H "Authorization: Bearer hm_xxx" \
  http://localhost:8080/api/v1/assignments/2
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
  "due_date": "2025-01-15T23:59:00+09:00",
  "is_completed": true,
  "completed_at": "2025-01-12T14:30:00+09:00",
  "created_at": "2025-01-10T10:00:00+09:00",
  "updated_at": "2025-01-12T14:30:00+09:00"
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
curl -X PATCH \
  -H "Authorization: Bearer hm_xxx" \
  http://localhost:8080/api/v1/assignments/1/toggle
```

---

## 統計情報取得

ユーザーの課題統計を取得します。科目、日付範囲でフィルタリング可能です。

```
GET /api/v1/statistics
```

### クエリパラメータ

| パラメータ | 型 | 説明 |
|------------|------|------|
| `subject` | string | 科目で絞り込み（省略時: 全科目） |
| `from` | string | 課題登録日の開始日（YYYY-MM-DD形式） |
| `to` | string | 課題登録日の終了日（YYYY-MM-DD形式） |

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
    },
    {
      "subject": "英語",
      "total": 10,
      "completed": 8,
      "pending": 2,
      "overdue": 0,
      "on_time_completion_rate": 87.5
    }
  ]
}
```

### 科目別統計 (特定科目のみ)

```json
{
  "total_assignments": 15,
  "completed_assignments": 12,
  "pending_assignments": 2,
  "overdue_assignments": 1,
  "on_time_completion_rate": 91.7,
  "filter": {
    "subject": "数学",
    "from": null,
    "to": null
  }
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

# 科目と日付範囲の組み合わせ
curl -H "Authorization: Bearer hm_xxx" "http://localhost:8080/api/v1/statistics?subject=数学&from=2025-01-01&to=2025-03-31"
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
| 500 Internal Server Error | サーバー内部エラー |

---

## 日付形式

APIは以下の日付形式を受け付けます（優先度順）：

1. **RFC3339**: `2025-01-15T23:59:00+09:00`
2. **日時形式**: `2025-01-15T23:59`
3. **日付のみ**: `2025-01-15`（時刻は23:59に設定）

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
      "recurrence_type": "weekly",
      "interval": 1,
      "weekday": 1,
      "is_active": true
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

### レスポンス

**200 OK**

---

## 繰り返し設定更新

```
PUT /api/v1/recurring/:id
```

### リクエストボディ

各フィールドはオプション。省略時は更新なし。

| フィールド | 型 | 説明 |
|------------|------|------|
| `title` | string | タイトル |
| `is_active` | boolean | `false` で停止、`true` で再開 |
| `recurrence_type` | string | `daily`, `weekly`, `monthly` |
| ... | ... | その他の設定フィールド |

### リクエスト例（停止）

```json
{
  "is_active": false
}
```

---

## 繰り返し設定削除

```
DELETE /api/v1/recurring/:id
```

### レスポンス

**200 OK**
