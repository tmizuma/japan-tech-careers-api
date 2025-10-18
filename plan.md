# ブログ自動化システム - 実装方針書

## 📋 プロジェクト概要

ChatGPT とのやり取りを通じてブログ記事を作成し、多言語翻訳、GitHub への commit/push、AWS Amplify への自動デプロイを行うシステム。

---

## 🏗️ システムアーキテクチャ

```
ChatGPT (MCP Client)
    ↓ MCP Protocol
MCP Server
    ↓ HTTPS
Go REST API (AWS Lambda)
    ↓ GitHub API (repository_dispatch)
GitHub Actions (blog-repo)
    ↓ Git push
AWS Amplify (自動デプロイ)
```

### コンポーネント説明

1. **ChatGPT (MCP Client)**

   - ユーザーとの対話インターフェース
   - ブログテーマの選択、内容レビュー、修正指示

2. **MCP Server**

   - ChatGPT からのリクエストを受信
   - Go API を呼び出し

3. **Go REST API (AWS Lambda)**

   - Amazon Bedrock でブログ生成・翻訳
   - GitHub API 経由で GitHub Actions をトリガー
   - Slack 通知
   - AWS SAM でデプロイ

4. **GitHub Actions (blog-repo)**

   - Markdown ファイルの作成
   - Git commit & push
   - Amplify への自動デプロイトリガー

5. **AWS Amplify**
   - Next.js プロジェクトのビルド
   - 多言語サイトの公開

---

## 🔧 技術スタック

### Go REST API

- **フレームワーク**: Gin
- **AWS SDK**:
  - `github.com/aws/aws-sdk-go-v2/service/bedrockruntime` (ブログ生成・翻訳)
  - `github.com/aws/aws-lambda-go` (Lambda 統合)
  - `github.com/awslabs/aws-lambda-go-api-proxy/gin` (Gin の Lambda 対応)
- **GitHub SDK**: `github.com/google/go-github/v66/github`
- **デプロイ**: AWS SAM

### インフラ

- **Lambda**: コンテナイメージ (Go)
- **API Gateway**: HTTP API (SAM 自動作成)
- **Secrets Manager**: GitHub Token, Slack Token 等

### CI/CD

- **GitHub Actions**: OIDC 認証で AWS と連携（クレデンシャル不要）

---

## 📡 API 仕様

### エンドポイント

#### 1. POST /blog/generate

ブログ記事を生成

**Request:**

```json
{
  "theme": "AWSのコスト最適化",
  "length": 2500,
  "tone": "technical",
  "additional_instructions": "セキュリティの観点も追加"
}
```

**Response:**

```json
{
  "content": "# AWSのコスト最適化\n\n...",
  "title": "AWSのコスト最適化戦略",
  "summary": "..."
}
```

#### 2. POST /blog/translate

ブログ記事を翻訳

**Request:**

```json
{
  "content": "# ブログ内容...",
  "title": "タイトル",
  "target_language": "en"
}
```

**Response:**

```json
{
  "translated_content": "# Blog content...",
  "translated_title": "Title"
}
```

#### 3. POST /blog/deploy

ブログを GitHub にデプロイ（GitHub Actions トリガー）

**Request:**

```json
{
  "title": "ブログタイトル",
  "slug": "blog-slug",
  "content_ja": "日本語コンテンツ",
  "content_en": "English content",
  "content_ko": "한국어 콘텐츠",
  "content_zh": "中文内容",
  "content_vi": "Nội dung tiếng Việt",
  "commit_message": "Add new blog post"
}
```

**Response:**

```json
{
  "status": "triggered",
  "message": "GitHub Actions has been triggered"
}
```

#### 4. POST /slack/notify

Slack 通知

**Request:**

```json
{
  "message": "ブログが公開されました",
  "channel": "#blog-updates",
  "urls": {
    "ja": "https://blog.com/ja/...",
    "en": "https://blog.com/en/..."
  }
}
```

---

## 🔑 実装の重要ポイント

### 1. Go API の実装

**GitHub Actions トリガー実装:**

```go
import "github.com/google/go-github/v66/github"

func triggerGitHubActions(payload DeployRequest) error {
    ctx := context.Background()
    token := os.Getenv("GITHUB_TOKEN")

    client := github.NewClient(nil).WithAuthToken(token)

    _, _, err := client.Repositories.CreateDispatchEvent(
        ctx,
        "your-org",
        "blog-repo",
        github.DispatchRequestOptions{
            EventType: "deploy-blog",
            ClientPayload: &payload,
        },
    )

    return err
}
```

**Bedrock 呼び出し実装:**

```go
import (
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/bedrockruntime"
)

func generateBlog(theme string) (string, error) {
    cfg, _ := config.LoadDefaultConfig(context.TODO())
    client := bedrockruntime.NewFromConfig(cfg)

    prompt := fmt.Sprintf("以下のテーマでブログ記事を書いてください: %s", theme)

    input := &bedrockruntime.InvokeModelInput{
        ModelId: aws.String("anthropic.claude-3-5-sonnet-20241022-v2:0"),
        Body: []byte(fmt.Sprintf(`{
            "anthropic_version": "bedrock-2023-05-31",
            "max_tokens": 4000,
            "messages": [{"role": "user", "content": "%s"}]
        }`, prompt)),
    }

    result, _ := client.InvokeModel(context.TODO(), input)
    // レスポンスをパース
    return content, nil
}
```

### 2. AWS SAM 設定

**template.yaml:**

```yaml
AWSTemplateFormatVersion: "2010-09-09"
Transform: AWS::Serverless-2016-10-31

Resources:
  BlogApiFunction:
    Type: AWS::Serverless::Function
    Properties:
      PackageType: Image
      Timeout: 300
      MemorySize: 1024
      Environment:
        Variables:
          GITHUB_TOKEN: !Sub "{{resolve:secretsmanager:github-token:SecretString}}"
      Policies:
        - AWSSecretsManagerGetSecretValuePolicy:
            SecretArn: !Sub "arn:aws:secretsmanager:${AWS::Region}:${AWS::AccountId}:secret:github-token-*"
        - Statement:
            - Effect: Allow
              Action:
                - bedrock:InvokeModel
              Resource: "*"
      Events:
        ApiEvent:
          Type: HttpApi
          Properties:
            Path: /{proxy+}
            Method: ANY

Metadata:
  BlogApiFunction:
    Dockerfile: Dockerfile
    DockerContext: .
    DockerTag: latest

Outputs:
  ApiUrl:
    Description: "API Gateway endpoint URL"
    Value: !Sub "https://${ServerlessHttpApi}.execute-api.${AWS::Region}.amazonaws.com/"
```

**Dockerfile:**

```dockerfile
FROM golang:1.21 AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

FROM public.ecr.aws/lambda/provided:al2023
COPY --from=builder /app/main /main
ENTRYPOINT ["/main"]
```

### 3. Secrets Manager 設定

AWS CLI で事前にシークレット作成:

```bash
aws secretsmanager create-secret \
  --name github-token \
  --secret-string "ghp_xxxxxxxxxxxxx"
```

---

## 🚀 CI/CD 設定

### blog-api-repo の GitHub Actions

**.github/workflows/deploy.yml:**

```yaml
name: Deploy Go API to Lambda

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    permissions:
      id-token: write # OIDC認証用
      contents: read

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Setup SAM
        uses: aws-actions/setup-sam@v2

      - name: Configure AWS Credentials (OIDC)
        uses: aws-actions/configure-aws-credentials@v4
        with:
          role-to-assume: arn:aws:iam::ACCOUNT_ID:role/GitHubActionsRole
          aws-region: us-east-1

      - name: SAM Build
        run: sam build

      - name: SAM Deploy
        run: |
          sam deploy \
            --no-confirm-changeset \
            --no-fail-on-empty-changeset \
            --stack-name blog-api \
            --capabilities CAPABILITY_IAM \
            --resolve-s3
```

### blog-repo の GitHub Actions

**.github/workflows/deploy-blog.yml:**

```yaml
name: Deploy Blog Content

on:
  repository_dispatch:
    types: [deploy-blog]

jobs:
  deploy:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Create blog files
        run: |
          SLUG="${{ github.event.client_payload.slug }}"

          # 各言語のディレクトリにMarkdownファイル作成
          mkdir -p content/ja content/en content/ko content/zh content/vi

          echo "${{ github.event.client_payload.content_ja }}" > "content/ja/${SLUG}.md"
          echo "${{ github.event.client_payload.content_en }}" > "content/en/${SLUG}.md"
          echo "${{ github.event.client_payload.content_ko }}" > "content/ko/${SLUG}.md"
          echo "${{ github.event.client_payload.content_zh }}" > "content/zh/${SLUG}.md"
          echo "${{ github.event.client_payload.content_vi }}" > "content/vi/${SLUG}.md"

      - name: Commit and Push
        run: |
          git config user.name "Blog Bot"
          git config user.email "bot@company.com"
          git add content/
          git commit -m "${{ github.event.client_payload.commit_message }}"
          git push

      - name: Wait for Amplify Deploy
        run: |
          echo "Amplify will automatically detect the push and deploy"
          echo "Check Amplify console for deployment status"
```

### AWS 側の OIDC 設定（一度だけ実行）

```bash
# 1. OIDC プロバイダー作成
aws iam create-open-id-connect-provider \
  --url https://token.actions.githubusercontent.com \
  --client-id-list sts.amazonaws.com \
  --thumbprint-list 6938fd4d98bab03faadb97b34396831e3780aea1

# 2. IAMロール作成
cat > trust-policy.json <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Federated": "arn:aws:iam::ACCOUNT_ID:oidc-provider/token.actions.githubusercontent.com"
      },
      "Action": "sts:AssumeRoleWithWebIdentity",
      "Condition": {
        "StringEquals": {
          "token.actions.githubusercontent.com:aud": "sts.amazonaws.com"
        },
        "StringLike": {
          "token.actions.githubusercontent.com:sub": "repo:your-org/blog-api:*"
        }
      }
    }
  ]
}
EOF

aws iam create-role \
  --role-name GitHubActionsRole \
  --assume-role-policy-document file://trust-policy.json

# 3. 必要な権限をアタッチ
aws iam attach-role-policy \
  --role-name GitHubActionsRole \
  --policy-arn arn:aws:iam::aws:policy/AWSCloudFormationFullAccess

aws iam attach-role-policy \
  --role-name GitHubActionsRole \
  --policy-arn arn:aws:iam::aws:policy/AWSLambda_FullAccess

# その他必要な権限...
```

---

## 💰 コスト試算（月間）

**前提:** 週 1 回のブログ投稿、各処理 5 分

| サービス                 | 使用量                    | 月額コスト     |
| ------------------------ | ------------------------- | -------------- |
| Lambda (1GB, 5 分 ×4 回) | 20 分/月                  | $0.02          |
| Bedrock (Claude Sonnet)  | 生成 1 回+翻訳 4 回 ×4 週 | $5-8           |
| API Gateway (HTTPApi)    | 20 リクエスト             | $0.00          |
| Secrets Manager          | 1 シークレット            | $0.40          |
| Amplify Hosting          | Next.js                   | $0 (無料枠)    |
| **合計**                 |                           | **$5.42-8.42** |

---

## 🔄 ワークフロー全体

1. **ユーザー → ChatGPT**

   - 「今週のブログを書いて」

2. **ChatGPT → MCP Server → Go API**

   - `POST /blog/generate` → ブログ生成
   - `POST /blog/translate` (×4 言語) → 翻訳

3. **ユーザーレビュー（ChatGPT 経由）**

   - 内容確認・修正指示

4. **承認後 → Go API**

   - `POST /blog/deploy` → GitHub Actions トリガー

5. **GitHub Actions (blog-repo)**

   - Markdown ファイル作成
   - Git commit & push

6. **AWS Amplify**

   - Git の push を検知
   - 自動ビルド・デプロイ

7. **完了通知**
   - `POST /slack/notify` → Slack 通知

---

## 🎯 実装の優先順位

### Phase 1: 基本機能

1. Go API の基本実装（generate, translate, deploy）
2. SAM 設定とローカルテスト
3. GitHub Actions 設定（両リポジトリ）
4. Lambda デプロイ

### Phase 2: 統合

1. MCP Server 実装
2. ChatGPT との連携テスト
3. エンドツーエンドテスト

### Phase 3: 改善

1. Slack 通知機能
2. エラーハンドリング強化
3. ログ・監視設定

---

## 📝 補足事項

### ローカル開発

```bash
# Goアプリをローカルで起動
go run main.go

# ローカルでテスト
curl -X POST http://localhost:8080/blog/generate \
  -H "Content-Type: application/json" \
  -d '{"theme": "test"}'

# SAMでローカルテスト
sam local start-api
```

### デバッグ

- CloudWatch Logs で Lambda 実行ログを確認
- GitHub Actions のログで実行状況を確認
- Amplify コンソールでデプロイ状況を確認

### セキュリティ

- GitHub Token は Secrets Manager で管理
- IAM ロールは最小権限の原則
- OIDC 認証でクレデンシャル不要

---

## 🔗 参考リンク

- [AWS SAM Documentation](https://docs.aws.amazon.com/serverless-application-model/)
- [GitHub Actions OIDC](https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/configuring-openid-connect-in-amazon-web-services)
- [go-github SDK](https://github.com/google/go-github)
- [AWS SDK for Go v2](https://aws.github.io/aws-sdk-go-v2/docs/)
- [Gin Web Framework](https://gin-gonic.com/)
