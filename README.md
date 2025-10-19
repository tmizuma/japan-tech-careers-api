# Japan Tech Careers API

Go + AWS Lambda + SAM を使用したブログ自動化システムのバックエンド API

## 構成

- **言語**: Go 1.25+
- **フレームワーク**: chi (ルーター)
- **テスト**: gomock (モック生成)
- **Logger**: zap (構造化ログ)
- **デプロイ**: AWS SAM
- **実行環境**: AWS Lambda (コンテナイメージ)
- **CI/CD**: GitHub Actions (OIDC 認証)

## アーキテクチャ

このプロジェクトは Clean Architecture に基づいて設計されており、各レイヤーが interface を通じて疎結合に接続されています。

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

各層の責務：

- **Router/Handler 層**: HTTP リクエストの処理、レスポンスの生成
- **Controller 層**: ビジネスロジックの調整、複数の Service の協調
- **Service 層**: ビジネスロジックの実装
- **HttpClient 層**: 外部 API 呼び出しの抽象化
- **Config 層**: 環境変数の管理、デフォルト値の提供
- **Logger 層**: zap を使用した構造化ログ、trace_id 対応

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

### Interface First 設計

すべての依存関係を interface で定義することで、テスト容易性と拡張性を実現しています：

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

- ✅ モック作成が容易（gomock で自動生成）
- ✅ 実装の差し替えが容易
- ✅ テストが独立して実行可能
- ✅ 依存関係が明示的

## プロジェクト構造

```
apps/api-server/
├── cmd/
│   └── main.go                      # エントリーポイント (Lambda/ローカル対応)
├── config/
│   ├── config.go                    # 環境変数ベースの設定管理
│   └── config_test.go
└── internal/
    ├── application/
    │   └── di.go                    # 依存性注入
    ├── domain/
    │   ├── model/                   # ドメインモデル
    │   │   └── job.go
    │   └── service/                 # ビジネスロジック
    │       ├── service.go           # interface + 実装
    │       ├── service_test.go
    │       └── mock/                # 自動生成されるモック
    │           └── mock_service.go
    ├── infra/
    │   ├── controller/              # Controller層
    │   │   ├── controller.go        # interface + 実装
    │   │   ├── controller_test.go
    │   │   └── mock/                # 自動生成されるモック
    │   │       └── mock_controller.go
    │   ├── httpclient/              # 外部APIクライアント
    │   │   ├── client.go            # interface + 実装
    │   │   └── mock/                # 自動生成されるモック
    │   │       └── mock_client.go
    │   └── router/                  # ルーティング
    │       ├── handler.go
    │       └── handler_test.go
    └── shared/
        └── logger/                  # zapベースのロガー
            └── logger.go
```

## ローカル開発

### 必要なツール

- Go 1.25+
- AWS SAM CLI
- Docker

### セットアップ

```bash
# リポジトリのクローン
git clone https://github.com/tmizuma/japan-tech-careers-api.git
cd japan-tech-careers-api

# 依存関係のインストール
go mod download

# モックの生成
make generate
```

### ローカルで実行

```bash
# ローカル起動
go run apps/api-server/cmd/main.go

# または
make run
```

```bash
# ヘルスチェック
curl http://localhost:8080/
# {"message":"Japan Tech Careers API is running","status":"healthy"}

# Job一覧取得
curl http://localhost:8080/jobs
# {"count":2,"jobs":[...]}
```

### SAM でローカルテスト

**注意**: Apple Silicon マシンでは`sam build`がエミュレーション（QEMU）を使用するため、非常に時間がかかります。ローカル開発では`go run main.go`の使用を推奨します。

```bash
# ビルド
sam build

# ローカル起動
sam local start-api

# テスト
curl http://localhost:3000/
```

## テスト戦略

このプロジェクトでは、gomock を使用した自動モック生成により、保守性の高いテストコードを実現しています。

### テストの実行

```bash
# すべてのテストを実行
make test

# カバレッジ付きでテストを実行
make test-coverage

# カバレッジレポートを開く
open apps/api-server/coverage.html
```

### gomock による自動モック生成

gomock は、Go の interface から自動的にモックコードを生成するツールです。

#### モックの生成

```bash
# すべてのモックを生成
make generate
```

各 interface ファイルには`//go:generate`ディレクティブが含まれています：

```go
//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -destination=mock/mock_$GOFILE -package=mock

package httpclient

type HttpClient interface {
    GetJobs(ctx context.Context) ([]model.Job, error)
}
```

#### テストの書き方

```go
func TestServiceImpl_FetchJobs(t *testing.T) {
    // Step 1: コントローラを作成
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    // Step 2: モックを作成
    mockClient := mock_httpclient.NewMockHttpClient(ctrl)

    // Step 3: モックの振る舞いを設定
    expectedJobs := []model.Job{
        {ID: "1", Title: "Test Job"},
    }
    mockClient.EXPECT().GetJobs(gomock.Any()).Return(expectedJobs, nil)

    // Step 4: テスト対象を初期化
    svc := NewServiceImpl(mockClient)

    // Step 5: テストを実行
    jobs, err := svc.FetchJobs(context.Background())

    // Step 6: 検証
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
    if len(jobs) != 1 {
        t.Errorf("Expected 1 job, got %d", len(jobs))
    }
}
```

#### モックの期待値設定

```go
// 任意の引数を受け入れる
mockClient.EXPECT().GetJobs(gomock.Any()).Return(jobs, nil)

// 特定の引数を期待する
mockClient.EXPECT().GetJobs(context.Background()).Return(jobs, nil)

// 複数回の呼び出しを期待する
mockClient.EXPECT().GetJobs(gomock.Any()).Return(jobs, nil).Times(2)

// エラーを返す
mockClient.EXPECT().GetJobs(gomock.Any()).Return(nil, errors.New("network error"))
```

### 新しい interface を追加する場合

1. **interface を定義**

```go
//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -destination=mock/mock_$GOFILE -package=mock

package mypackage

type MyInterface interface {
    DoSomething(ctx context.Context) error
}
```

2. **モックを生成**

```bash
make generate
```

3. **テストで使用**

```go
import (
    mock_mypackage "github.com/tmizuma/japan-tech-careers-api/apps/api-server/internal/mypackage/mock"
    "go.uber.org/mock/gomock"
)

func TestMyFunction(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockObj := mock_mypackage.NewMockMyInterface(ctrl)
    mockObj.EXPECT().DoSomething(gomock.Any()).Return(nil)

    // テストコード...
}
```

### ベストプラクティス

1. **テストのたびに ctrl.Finish()を呼ぶ**

   ```go
   ctrl := gomock.NewController(t)
   defer ctrl.Finish()  // 必ず呼ぶ
   ```

2. **モックの期待値は明示的に設定する**

   ```go
   // Good: 期待値を明示
   mockClient.EXPECT().GetJobs(gomock.Any()).Return(jobs, nil)
   ```

3. **interface を変更したら必ず再生成**
   ```bash
   make generate
   ```

## ビルド

```bash
# バイナリをビルド
make build

# 生成されたバイナリ
./bin/api-server
```

## AWS セットアップ

### 1. OIDC プロバイダーの作成（初回のみ）

```bash
aws iam create-open-id-connect-provider \
  --url https://token.actions.githubusercontent.com \
  --client-id-list sts.amazonaws.com \
  --thumbprint-list 6938fd4d98bab03faadb97b34396831e3780aea1
```

### 2. IAM ロールの作成

`trust-policy.json` を作成:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::YOUR_ACCOUNT_ID:oidc-provider/token.actions.githubusercontent.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "token.actions.githubusercontent.com:aud": "sts.amazonaws.com"
        },
        "StringLike": {
          "token.actions.githubusercontent.com:sub": "repo:tmizuma/japan-tech-careers-api:*"
        }
      }
    }
  ]
}
```

IAM ロールを作成:

```bash
aws iam create-role \
  --role-name GitHubActionsRole \
  --assume-role-policy-document file://trust-policy.json
```

### 3. 必要な権限をアタッチ

```bash
# CloudFormation
aws iam attach-role-policy \
  --role-name GitHubActionsRole \
  --policy-arn arn:aws:iam::aws:policy/AWSCloudFormationFullAccess

# Lambda
aws iam attach-role-policy \
  --role-name GitHubActionsRole \
  --policy-arn arn:aws:iam::aws:policy/AWSLambda_FullAccess

# IAM (Lambda用ロール作成のため)
aws iam attach-role-policy \
  --role-name GitHubActionsRole \
  --policy-arn arn:aws:iam::aws:policy/IAMFullAccess

# S3 (SAMのアーティファクト保存用)
aws iam attach-role-policy \
  --role-name GitHubActionsRole \
  --policy-arn arn:aws:iam::aws:policy/AmazonS3FullAccess

# ECR (コンテナイメージ用)
aws iam attach-role-policy \
  --role-name GitHubActionsRole \
  --policy-arn arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryFullAccess

# API Gateway
aws iam attach-role-policy \
  --role-name GitHubActionsRole \
  --policy-arn arn:aws:iam::aws:policy/AmazonAPIGatewayAdministrator
```

### 4. GitHub Secrets の設定

リポジトリの Settings > Secrets and variables > Actions で以下を追加:

- `AWS_ROLE_ARN`: `arn:aws:iam::904233098356:role/GitHubActionsRole`
- `AWS_REGION`: `ap-northeast-1` (または任意のリージョン)

## デプロイ

### 手動デプロイ

```bash
sam build
sam deploy --guided
```

### 自動デプロイ (GitHub Actions)

`main` ブランチに push すると自動的にデプロイされます:

```bash
git add .
git commit -m "Update API"
git push origin main
```

## エンドポイント

デプロイ後、以下のエンドポイントが利用可能:

### `GET /`

ヘルスチェックエンドポイント

```bash
curl https://5lhcnptds4.execute-api.ap-northeast-1.amazonaws.com/
# {"message":"Japan Tech Careers API is running","status":"healthy"}
```

### `GET /jobs`

Job 一覧を取得（現在はダミーデータを返却）

```bash
curl https://5lhcnptds4.execute-api.ap-northeast-1.amazonaws.com/jobs
# {"count":2,"jobs":[{"id":"1","title":"Senior Go Developer","company":"Tech Company A","location":"Tokyo, Japan","description":"Looking for an experienced Go developer"},...]}}
```

## 環境変数

Lambda 関数で使用される環境変数は `template.yaml` で定義されています:

- `ENVIRONMENT`: 実行環境 (dev, prod, local) - デフォルト: "local"
- `LOG_LEVEL`: ログレベル (info, debug, error) - デフォルト: "info"
- `API_ENDPOINT`: 外部 API のエンドポイント - デフォルト: "https://api.example.com"
- `API_TIMEOUT`: HTTP タイムアウト(秒) - デフォルト: 30

ローカル開発時は、これらの環境変数が未設定の場合、デフォルト値が使用されます。

## 開発フロー

### 1. モックの生成

新しい interface を追加したり、既存の interface を変更した場合：

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

### 3. ビルドと実行

```bash
# ビルド
make build

# ローカル実行
make run
```

## トラブルシューティング

### モックが見つからないエラー

```bash
make generate
```

を実行してモックを生成してください。

### go mod tidy / go mod vendor のエラー

テストファイルがまだ存在しない mock を import している場合、以下の手順で解決します：

1. テストファイルを一時的にリネーム
2. `go mod tidy`を実行
3. `make generate`で mock を生成
4. テストファイルを元に戻す
5. 再度`go mod tidy`を実行

または、プロジェクトの Makefile に定義されているコマンドを使用：

```bash
make generate
make test
```

## ライセンス

MIT

## 関連ドキュメント

- [CLAUDE.md](CLAUDE.md) - 開発コンテキストと引き継ぎ情報
- [template.yaml](template.yaml) - SAM 設定
