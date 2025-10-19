# テスト戦略とモック生成

このドキュメントでは、プロジェクトのテスト戦略とgomockを使用したモック生成について説明します。

## gomockとは

gomockは、Goのinterfaceから自動的にモックコードを生成するツールです。手動でモックを書く必要がなく、interfaceが変更された場合も再生成するだけで済みます。

## モックの生成

### 1. モックの自動生成

```bash
# すべてのモックを生成
make generate
```

これにより、以下のファイルに定義された`//go:generate`ディレクティブが実行されます：

- `internal/infra/httpclient/client.go` → `internal/infra/httpclient/mock/mock_client.go`
- `internal/domain/service/service.go` → `internal/domain/service/mock/mock_service.go`

### 2. モックの生成場所

生成されたモックは各パッケージの`mock/`ディレクトリに配置されます：

```
internal/
├── domain/
│   └── service/
│       ├── service.go
│       ├── service_test.go
│       └── mock/
│           └── mock_service.go  (自動生成)
└── infra/
    └── httpclient/
        ├── client.go
        ├── mock/
        │   └── mock_client.go  (自動生成)
```

## テストの実行

### 基本的なテスト実行

```bash
# すべてのテストを実行
make test

# または
cd apps/api-server && go test ./... -v
```

### カバレッジ付きテスト

```bash
# カバレッジレポートを生成
make test-coverage

# 生成されたHTMLレポートをブラウザで開く
open apps/api-server/coverage.html
```

## gomockの使い方

### 1. モックの作成

gomockを使ったテストは以下のパターンで書きます：

```go
func TestServiceImpl_FetchJobs_Success(t *testing.T) {
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

### 2. モックの期待値設定

#### 任意の引数を受け入れる

```go
mockClient.EXPECT().GetJobs(gomock.Any()).Return(jobs, nil)
```

#### 特定の引数を期待する

```go
mockClient.EXPECT().GetJobs(context.Background()).Return(jobs, nil)
```

#### 複数回の呼び出しを期待する

```go
mockClient.EXPECT().GetJobs(gomock.Any()).Return(jobs, nil).Times(2)
```

#### 呼び出し順序を指定する

```go
gomock.InOrder(
    mockClient.EXPECT().GetJobs(gomock.Any()).Return(jobs1, nil),
    mockClient.EXPECT().GetJobs(gomock.Any()).Return(jobs2, nil),
)
```

### 3. エラーケースのテスト

```go
func TestServiceImpl_FetchJobs_Error(t *testing.T) {
    ctrl := gomock.NewController(t)
    defer ctrl.Finish()

    mockClient := mock_httpclient.NewMockHttpClient(ctrl)

    // エラーを返すように設定
    expectedError := errors.New("network error")
    mockClient.EXPECT().GetJobs(gomock.Any()).Return(nil, expectedError)

    svc := NewServiceImpl(mockClient)
    jobs, err := svc.FetchJobs(context.Background())

    // エラーが返されることを検証
    if err == nil {
        t.Error("Expected error, got nil")
    }
    if jobs != nil {
        t.Errorf("Expected nil jobs, got %v", jobs)
    }
}
```

## 新しいinterfaceを追加する場合

1. **interfaceを定義**

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
    mock_mypackage "path/to/mypackage/mock"
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

## 手動モックとの比較

### 手動モック（非推奨）

```go
// 手動で書く必要がある
type MockHttpClient struct {
    GetJobsFunc func(ctx context.Context) ([]model.Job, error)
}

func (m *MockHttpClient) GetJobs(ctx context.Context) ([]model.Job, error) {
    if m.GetJobsFunc != nil {
        return m.GetJobsFunc(ctx)
    }
    return []model.Job{}, nil
}

// テストのたびに関数を書く必要がある
mockClient := &MockHttpClient{
    GetJobsFunc: func(ctx context.Context) ([]model.Job, error) {
        return []model.Job{{ID: "1"}}, nil
    },
}
```

**問題点:**
- interfaceが変更されるたびに手動で更新が必要
- 呼び出し回数の検証が面倒
- 大量のボイラープレートコード

### gomock（推奨）

```go
// 自動生成される
mockClient := mock_httpclient.NewMockHttpClient(ctrl)

// シンプルな期待値設定
mockClient.EXPECT().GetJobs(gomock.Any()).Return([]model.Job{{ID: "1"}}, nil)

// 呼び出し回数の検証も簡単
mockClient.EXPECT().GetJobs(gomock.Any()).Return(jobs, nil).Times(2)
```

**利点:**
- interfaceが変更されても`make generate`で自動更新
- 呼び出し回数、引数、順序の検証が簡単
- コード量が大幅に削減
- タイプセーフ

## ベストプラクティス

1. **テストのたびにctrl.Finish()を呼ぶ**
   ```go
   ctrl := gomock.NewController(t)
   defer ctrl.Finish()  // 必ず呼ぶ
   ```

2. **モックの期待値は明示的に設定する**
   ```go
   // Good: 期待値を明示
   mockClient.EXPECT().GetJobs(gomock.Any()).Return(jobs, nil)

   // Bad: 期待値なしでメソッドを呼ぶとテストが失敗する
   ```

3. **interfaceを変更したら必ず再生成**
   ```bash
   make generate
   ```

4. **生成されたモックファイルはgitにコミットしない**
   - `.gitignore`に`**/mock/`を追加推奨
   - CIで`make generate`を実行する

## まとめ

gomockを使うことで：
- ✅ モックコードを手動で書く必要がない
- ✅ interfaceの変更に強い
- ✅ テストコードが簡潔になる
- ✅ 保守コストが大幅に削減される
- ✅ 大規模プロジェクトでもスケールする
