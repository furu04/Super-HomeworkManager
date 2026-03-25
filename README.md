<div align="center">

# Super Homework Manager

シンプルで高機能な課題管理アプリケーション

[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![License](https://img.shields.io/badge/License-AGPLv3-blue.svg)](LICENSE.md)
[![Docker](https://img.shields.io/badge/Docker-Compose-2496ED?logo=docker&logoColor=white)](#docker-での実行)

</div>

---

## 概要

学生の課題管理を効率化するために設計されたWebアプリケーションです。
繰り返し課題の自動生成やダッシュボードによる期限管理など、日々の課題管理をサポートします。

## スクリーンショット

| ダッシュボード | 課題一覧 | API |
|:---:|:---:|:---:|
| ![ダッシュボード](./docs/images/dashboard.png) | ![課題一覧](./docs/images/list.png) | ![API](./docs/images/api.png) |

## 特徴

| 機能 | 説明 |
|---|---|
| **課題管理** | 課題の登録・編集・削除・完了状況の管理 |
| **繰り返し課題** | 日次・週次・月次の繰り返し課題を自動生成 |
| **ダッシュボード** | 期限切れ・本日期限・今週期限の課題をひと目で確認 |
| **REST API** | 外部連携用のAPIキー認証付きRESTful API |
| **セキュリティ** | CSRF対策 / レート制限 / セキュアなセッション管理 / 2FA対応 |
| **ポータビリティ** | Pure Go SQLiteドライバー使用でCGO不要 |

## クイックスタート

### 前提条件

- **Docker Desktop** または Docker / Docker Compose
- （ローカルで直接ビルドする場合のみ）**Go 1.24 以上**

### 最も簡単な起動方法

初めて使う方には、**Dockerを使用した起動**をおすすめします。

1. このリポジトリをダウンロード（または `git clone`）し、フォルダを開きます。
2. フォルダ内にある `config.ini.docker.example` というファイルをコピーし、**名前を `config.ini` に変更**します。
   （※ **必須**: この作業を忘れると起動エラーになります）
3. ターミナル（またはコマンドプロンプトやPowerShell）でこのフォルダを開き、以下のコマンドを実行します：
   ```bash
   docker-compose up -d --build
   ```
4. ブラウザを開き、**http://localhost** にアクセスしてください。

> **注意**: 本番環境（外部公開するサーバー上）で使用する場合は、`Caddyfile` の `:80` を実際のドメインに変更し、`config.ini` 内の `[session] secret` 等の安全な文字列への変更を必ず行ってください。

### ローカルビルド（開発者向け）

開発目的で直接実行する場合の手順です。

```bash
# 1. リポジトリのクローン
git clone <repository-url>
cd Homework-Manager

# 2. 依存関係のダウンロード
go mod download

# 3. ビルド
go build -o homework-manager cmd/server/main.go

# 4. 設定ファイルの準備
cp config.ini.example config.ini

# 5. 実行
./homework-manager
```

> **Windows (PowerShell)** の場合:
> `Copy-Item config.ini.example config.ini` → `.\homework-manager.exe`

ブラウザで **http://localhost:8080** にアクセスしてください。

## 更新方法

```bash
git pull
go build -o homework-manager cmd/server/main.go
# アプリケーションを再起動
```

## ドキュメント

| ドキュメント | 内容 |
|---|---|
| [仕様書](docs/SPECIFICATION.md) | 機能詳細・データモデル・設定項目 |
| [APIドキュメント](docs/API.md) | エンドポイント・リクエスト/レスポンス形式 |

## TODO

- [ ] 取り組み目安時間の登録
- [ ] SNS連携

## ライセンス

[AGPLv3 (GNU Affero General Public License v3)](LICENSE.md)
