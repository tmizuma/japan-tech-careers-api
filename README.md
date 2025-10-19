# Japan Tech Careers API

Go + AWS Lambda + SAM を使用したブログ自動化システムのバックエンド API

## 構成

- **言語**: Go 1.25
- **フレームワーク**: chi (ルーター)
- **Logger**: zap (構造化ログ)
- **デプロイ**: AWS SAM
- **実行環境**: AWS Lambda (コンテナイメージ)
- **CI/CD**: GitHub Actions (OIDC 認証)

## プロジェクト構造

```
apps/api-server/
├── cmd/
│   └── main.go              # エントリーポイント (Lambda/ローカル対応)
├── config/
│   └── config.go            # 環境変数ベースの設定管理
└── internal/
    ├── application/
    │   └── di.go            # 依存性注入
    ├── domain/
    │   ├── model/           # ドメインモデル
    │   │   └── job.go
    │   └── service/         # ビジネスロジック
    │       └── service.go
    ├── infra/
    │   ├── httpclient/      # 外部APIクライアント
    │   │   └── client.go
    │   └── router/          # ルーティング
    │       └── handler.go
    └── shared/
        └── logger/          # zapベースのロガー
            └── logger.go
```

## アーキテクチャ

### 依存関係フロー

```
main.go
  ↓
config.NewConfig() (環境変数読み込み)
  ↓
application.New(config)
  ↓
  ├── httpclient.New(config)
  ├── service.NewServiceImpl(httpClient)
  └── router.NewRouter(service)
```

### レイヤー構成

- **Config 層**: 環境変数の管理、デフォルト値の提供
- **Logger 層**: zap を使用した構造化ログ、trace_id 対応
- **HttpClient 層**: 外部 API 呼び出しの抽象化
- **Service 層**: ビジネスロジックの実装
- **Router/Handler 層**: HTTP リクエストの処理
- **DI 層**: 依存性注入による疎結合化

## ローカル開発

### 必要なツール

- Go 1.25+
- AWS SAM CLI
- Docker

### ローカルで実行

```bash
# 依存関係のインストール
go mod download

# ローカル起動
go run apps/api-server/cmd/main.go

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
          "token.actions.githubusercontent.com:sub": "repo:YOUR_ORG/japan-tech-careers-api:*"
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
git commit -m "Initial commit"
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

curl https://5lhcnptds4.execute-api.ap-northeast-1.amazonaws.com/jobs

# {"count":2,"jobs":[{"id":"1","title":"Senior Go Developer","company":"Tech Company A","location":"Tokyo, Japan","description":"Looking for an experienced Go developer"},...]}

```

## 環境変数

Lambda関数で使用される環境変数は `template.yaml` で定義されています:

- `ENVIRONMENT`: 実行環境 (dev, prod, local)
- `LOG_LEVEL`: ログレベル (info, debug, error)
- `API_ENDPOINT`: 外部APIのエンドポイント
- `API_TIMEOUT`: HTTPタイムアウト(秒)

ローカル開発時は、これらの環境変数が未設定の場合、デフォルト値が使用されます。
```
