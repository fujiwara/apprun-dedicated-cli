# apprun-dedicated-cli Implementation Plan

## Overview

AppRun Dedicated (専有型) の管理CLIツール。
apprun-cli (共用型CLI) と同様のコマンド体系・設計パターンを踏襲し、ecspresso の優れた機能を取り入れる。

SDK: [sacloud/apprun-dedicated-api-go](https://github.com/sacloud/apprun-dedicated-api-go)

## Architecture

### Dedicated vs Shared の違い

Shared (共用型) は Application 1リソースで完結するが、Dedicated (専有型) は階層的なリソース構造を持つ:

```
Cluster
├── Certificate
├── AutoScalingGroup (ASG)
│   ├── LoadBalancer
│   └── WorkerNode
└── Application
    └── Version (内部的に世代管理)
```

### ユーザー体験の設計方針

apprun-cli と同様に、ユーザーは **application の定義を書いて deploy する** だけ。
Version (アプリケーションバージョン) は deploy のたびに SDK 側で自動的に採番される内部概念であり、ユーザーが直接操作する必要はない。

- `provision` → 定義ファイルの cluster/asg/lb/certificate を作成・更新（初回 or インフラ変更時に明示的に実行）
- `deploy` → application 定義から新しい Version を作成し、アクティブに切り替え（**application のみ**）
- `rollback` → 前の Version に戻す
- `versions` → 過去の Version 一覧を確認（運用・デバッグ用途）
- `status` → 現在のアクティブ Version を含むアプリケーション状態を表示

> **安全性の設計**: `deploy` は application (Version) の操作のみ行い、インフラリソースには触れない。
> Cluster/ASG/LB/Certificate の作成・更新は `provision` コマンドで明示的に実行する。
> これにより、image tag を変えただけの deploy で ASG 設定が意図せず変更される事故を防ぐ。

### 設定ファイル構造

**config + 定義ファイル2つ** の構成。操作対象（setup / deploy）ごとにファイルを分ける。

**config.jsonnet / config.json** (ecspresso の ecspresso.yml に相当):
```jsonnet
{
  cluster: "my-cluster",                        // cluster name
  application: "my-app",                        // application name
  cluster_definition: "cluster.jsonnet",        // setup が操作
  application_definition: "application.jsonnet", // deploy が操作
}
```

**cluster.jsonnet** (インフラ定義 - `provision` が操作):
```jsonnet
{
  // クラスタ
  cluster: {
    ports: [80, 443],
    lets_encrypt_email: "admin@example.com",
  },

  // ASG
  asg: {
    name: "my-asg",
    zone: "tk1v",
    min_nodes: 1,
    max_nodes: 3,
    worker_service_class_path: "...",
    interfaces: [ ... ],
  },

  // ロードバランサ (optional)
  lb: {
    name: "my-lb",
    service_class_path: "...",
    interfaces: [ ... ],
  },

  // 証明書 (optional)
  certificate: {
    name: "my-cert",
    certificate_pem: "...",
    private_key_pem: "...",
  },
}
```

**application.jsonnet** (アプリケーション定義 - `deploy` が操作):
```jsonnet
{
  cpu: 2,
  memory: 4,
  scaling_mode: "...",
  image: {
    path: std.native('must_env')('IMAGE_PATH'),
    tag: std.native('must_env')('IMAGE_TAG'),
  },
  exposed_ports: [ ... ],
  environment_variables: [ ... ],
}
```

設定ファイルは JSON / Jsonnet 対応 (`fujiwara/jsonnet-armed` でテンプレート関数を提供)。

> **設計意図**: ファイルがコマンドの操作範囲と一致する。
> - `cluster.jsonnet` → `provision` が読む。初回構築後はほぼ触らない。
> - `application.jsonnet` → `deploy` が読む。日常的に編集する。
> - `diff` はそれぞれの定義ファイルに対応する差分を表示する。

### テンプレート関数

apprun-cli / ecspresso と同様:
- `env(name, default)` - 環境変数
- `must_env(name)` - 必須環境変数
- `tfstate(resource_path)` - Terraform state 参照
- `secret(id)` - さくらクラウド Secret Manager

## Commands

### Phase 1: Core (MVP)

apprun-cli と同等の基本機能。

| Command | Description | Notes |
|---------|-------------|-------|
| `version` | CLIバージョン表示 | 既存 |
| `init` | 既存リソースから定義ファイル生成 | ecspresso の init と同様。`--cluster`, `--application` で対象指定 |
| `provision` | インフラのセットアップ | cluster/asg/lb/certificate の作成・更新。初回 or インフラ変更時に実行 |
| `deploy` | アプリケーションのデプロイ | application の新 Version 作成 + アクティブ切替のみ。インフラには触れない |
| `diff` | ローカル定義とリモートの差分表示 | apprun-cli と同様に jsondiff 使用 |
| `render` | 定義ファイルのレンダリング結果表示 | テンプレート展開後の JSON を stdout へ |
| `status` | アプリケーション/クラスタのステータス表示 | |
| `delete` | リソース削除 | `delete application` 等 |

### Phase 2: Node Management & Operations

運用時に個別操作が必要な場面向け。

| Command | Description | Notes |
|---------|-------------|-------|
| `node list` | ワーカーノード一覧 | |
| `node drain` | ワーカーノードのドレイン | |

### Phase 3: Advanced Features (ecspresso inspired)

| Command | Description | Origin |
|---------|-------------|--------|
| `rollback` | 前 Version へのロールバック | ecspresso |
| `verify` | リソース依存関係の検証 | ecspresso |
| `wait` | デプロイ完了待ち | ecspresso |
| `versions` | Version 一覧・詳細・削除 (運用向け) | apprun-cli |
| `containers` | コンテナ配置情報表示 | Dedicated SDK 固有 |

## Implementation Steps

### Step 1: Project Foundation

- [ ] go.mod に依存追加 (`apprun-dedicated-api-go`, `kong`, `jsonnet-armed`, etc.)
- [ ] CLI 構造体定義 (`cli.go`) - kong ベース
- [ ] クライアント初期化 (認証: `SAKURA_ACCESS_TOKEN` / `SAKURA_ACCESS_TOKEN_SECRET`)
- [ ] config ファイル読み込み (`config.go`)
- [ ] リソース定義ファイル読み込み (`loader.go`) - jsonnet-armed 統合

### Step 2: Status

- [ ] `status` コマンド - アプリケーション/クラスタの詳細表示

### Step 3: Init & Render

- [ ] `init` コマンド - 既存リソースから定義ファイル生成 (JSON/Jsonnet)
- [ ] `render` コマンド - テンプレートレンダリング結果表示

### Step 4: Setup, Deploy & Diff

- [ ] `provision` コマンド - cluster/asg/lb/certificate の作成・更新
- [ ] `deploy` コマンド - application の Version 作成 + アクティブ切替
- [ ] `diff` コマンド - ローカル vs リモート差分表示
- [ ] `delete` コマンド - リソース削除

### Step 5: Node Management

- [ ] `node list` / `node drain` コマンド

### Step 6: Advanced Features

- [ ] `rollback` コマンド
- [ ] `verify` コマンド
- [ ] `wait` コマンド
- [ ] `versions` コマンド
- [ ] `containers` コマンド

## Technical Decisions

### CLI Framework
**kong** (`alecthomas/kong`) - apprun-cli と同じ

### Config/Template
**jsonnet-armed** (`fujiwara/jsonnet-armed`) - apprun-cli と同じ
env, must_env, tfstate, secret 等のネイティブ関数を提供

### Diff
**jsondiff** (`aereal/jsondiff`) + **gojq** (`itchyny/gojq`) - apprun-cli と同じ

### Output
- 一覧: テーブル形式 (デフォルト) / JSON (`--json`)
- 詳細: JSON
- Diff: unified diff (color)

### Logging
**sloghandler** (`fujiwara/sloghandler`) を使用
- デフォルト: text 形式
- `--log-format json` で JSON 形式に切替
- `--debug` フラグで詳細ログ (DEBUG レベル)

### Error Handling
- SDK の error を適切にラップ

### Testing
- 各コマンドの単体テスト
- 定義ファイル読み込みのテスト (testdata/)
- SDK モック or integration test

## File Structure (Planned)

```
apprun-dedicated-cli/
├── cmd/apprun-dedicated-cli/
│   └── main.go              # エントリーポイント (既存)
├── cli.go                   # CLI 構造体・コマンドルーティング
├── config.go                # config ファイル読み込み
├── definition.go            # 定義ファイル読み込み・型定義 (jsonnet-armed)
├── client.go                # SDK クライアント初期化
├── provision.go                 # setup コマンド (cluster/asg/lb/certificate の作成・更新)
├── deploy.go                # deploy コマンド (application の Version 作成 + アクティブ切替)
├── diff.go                  # diff コマンド
├── init.go                  # init コマンド
├── render.go                # render コマンド
├── list.go                  # list コマンド
├── status.go                # status コマンド
├── delete.go                # delete コマンド
├── versions.go              # versions コマンド (Version 一覧・詳細・削除)
├── rollback.go              # rollback コマンド
├── node.go                  # node list / drain コマンド
├── verify.go                # verify コマンド
├── wait.go                  # wait コマンド
├── containers.go            # containers コマンド
├── version.go               # CLIバージョン情報 (既存)
├── testdata/                # テスト用定義ファイル
├── go.mod
└── go.sum
```
