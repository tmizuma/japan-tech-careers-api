# Japan Tech Careers API - 開発コンテキスト

このドキュメントは、別の Claude Code インスタンスに引き継ぐための開発コンテキストを記述します。

## プロジェクト概要

Go + AWS Lambda + SAM を使用した API サーバーのプロジェクトです。Clean Architecture に基づいた設計で、interface による依存性注入と gomock を使った自動モック生成を採用しています。

## アーキテクチャ

### レイヤー構成

```
Router (handler.go)
  ↓
Controller (controller.go)
  ↓
Service (service.go)
  ↓
HttpClient (client.go)
```

### 依存関係フロー

```
main.go
  ↓
config.NewConfig() (環境変数読み込み)
  ↓
application.New(config) - DI
  ↓
  ├── httpclient.New(config)
  ├── service.NewServiceImpl(httpClient)
  ├── controller.NewController(service)
  └── router.NewRouter(controller)
```

### ディレクトリ構造

```
apps/api-server/
├── cmd/
│   └── main.go                      # エントリーポイント
├── config/
│   ├── config.go                    # 環境変数ベースの設定
│   └── config_test.go
└── internal/
    ├── application/
    │   └── di.go                    # 依存性注入
    ├── domain/
    │   ├── model/
    │   │   └── job.go               # ドメインモデル
    │   └── service/
    │       ├── service.go           # ビジネスロジック interface + 実装
    │       ├── service_test.go
    │       └── mock/                # 自動生成されるモック
    │           └── mock_service.go
    ├── infra/
    │   ├── controller/
    │   │   ├── controller.go        # Controller層 interface + 実装
    │   │   ├── controller_test.go
    │   │   └── mock/                # 自動生成されるモック
    │   │       └── mock_controller.go
    │   ├── httpclient/
    │   │   ├── client.go            # 外部API呼び出し interface + 実装
    │   │   └── mock/                # 自動生成されるモック
    │   │       └── mock_client.go
    │   └── router/
    │       ├── handler.go           # HTTPハンドラ
    │       └── handler_test.go
    └── shared/
        └── logger/
            └── logger.go            # zapベースのロガー
```

## 重要な設計判断

### 1. Controller 層の追加

当初は `Router → Service → HttpClient` でしたが、以下の理由で Controller 層を追加しました：

- **責務の分離**: ルーティングとビジネスロジックの調整を分離
- **テスト容易性**: Handler 層は Controller のモックで、Controller 層は Service のモックでテスト可能
- **拡張性**: 将来的に複数の Service を協調させる場合、Controller で調整可能

### 2. gomock による自動モック生成

手動でモックを書くのは保守コストが高いため、gomock を採用：

- **//go:generate ディレクティブ**: 各 interface ファイルの先頭に記述
- **make generate**: すべてのモックを自動生成
- **テストコード**: gomock の API を使ってモックの振る舞いを設定

### 3. Interface First 設計

すべての依存関係を interface で定義：

```go
// HttpClient interface
type HttpClient interface {
    GetJobs(ctx context.Context) ([]model.Job, error)
}

// Service interface
type Service interface {
    FetchJobs(ctx context.Context) ([]model.Job, error)
}

// Controller interface
type Controller interface {
    GetJobs(ctx context.Context) ([]model.Job, error)
}
```

これにより：

- ✅ モック作成が容易
- ✅ 実装の差し替えが容易
- ✅ テストが独立して実行可能
- ✅ 依存関係が明示的

## 開発フロー

### 1. モックの生成

```bash
make generate
```

### 2. テストの実行

```bash
# 通常のテスト
make test

# カバレッジ付き
make test-coverage
```

### 3. ローカルでの実行

```bash
make run
# または
go run apps/api-server/cmd/main.go
```

## 環境変数

Lambda 環境では`template.yaml`で定義、ローカルではデフォルト値を使用：

- `ENVIRONMENT`: 実行環境 (dev, prod, local) - デフォルト: "local"
- `LOG_LEVEL`: ログレベル (info, debug, error) - デフォルト: "info"
- `API_ENDPOINT`: 外部 API エンドポイント - デフォルト: "https://api.example.com"
- `API_TIMEOUT`: HTTP タイムアウト(秒) - デフォルト: 30

## 実装済みの機能

### エンドポイント

1. **GET /** - ヘルスチェック
2. **GET /jobs** - Job 一覧取得（現在はダミーデータ）

### テスト

- **Config 層**: 環境変数の読み込み、デフォルト値のテスト
- **HttpClient 層**: gomock でモック化（Service 層のテストで使用）
- **Service 層**: HttpClient をモックしてテスト
- **Controller 層**: Service をモックしてテスト
- **Handler 層**: Controller をモックしてテスト

## テスト戦略

### ユニットテスト

各層は独立してテスト可能：

```go
// Controller層のテスト例
ctrl := gomock.NewController(t)
defer ctrl.Finish()

mockService := mock_service.NewMockService(ctrl)
mockService.EXPECT().FetchJobs(gomock.Any()).Return(expectedJobs, nil)

controller := NewController(mockService)
jobs, err := controller.GetJobs(ctx)
// Assert...
```

### 統合テスト

実装予定：

- Lambda 統合テスト（SAM local）
- E2E テスト

## トラブルシューティング

### Go 1.25 のダウンロードエラー

現在の環境では Go 1.25 がインストールされていない場合があります。
`go mod tidy`をスキップするか、Go バージョンを変更してください。

### モックが見つからないエラー

```bash
make generate
```

を実行してモックを生成してください。

## 参考ドキュメント

- [TESTING.md](apps/api-server/TESTING.md) - テスト戦略と gomock の詳細
- [README.md](README.md) - プロジェクト全体の説明
- [template.yaml](template.yaml) - SAM 設定

## 連絡事項

- **テスト実行**: まだ実行していないため、初回実行時にエラーが出る可能性があります
- **モック生成**: `make generate`を実行してから`make test`を実行してください
- **依存関係**: `go mod tidy`はスキップされているため、必要に応じて実行してください

---

**最終更新**: 2025-10-19
**実装者**: Claude Code
