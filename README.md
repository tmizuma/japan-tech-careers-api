# Japan Tech Careers API

Go + AWS Lambda + SAM を使用したブログ自動化システムのバックエンド API

## 構成

- **言語**: Go 1.25
- **フレームワーク**: chi (ルーター)
- **デプロイ**: AWS SAM
- **実行環境**: AWS Lambda (コンテナイメージ)
- **CI/CD**: GitHub Actions (OIDC 認証)

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
go run main.go

# テスト
curl http://localhost:8080/
# {"message":"hello"}
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

- `AWS_ROLE_ARN`: `arn:aws:iam::YOUR_ACCOUNT_ID:role/GitHubActionsRole`
- `AWS_REGION`: `us-east-1` (または任意のリージョン)

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

- `GET /` - ヘルスチェック (`{"message":"hello"}`)

## 次のステップ

plan.md を参照して、以下の機能を実装:

1. `POST /blog/generate` - ブログ生成 (Bedrock)
2. `POST /blog/translate` - 翻訳 (Bedrock)
3. `POST /blog/deploy` - GitHub Actions トリガー
4. `POST /slack/notify` - Slack 通知
