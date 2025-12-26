# kakucho - 任意のCLIを拡張するためのCLI

## 概要

`kakucho` は任意のCLI（例: `hoge`）に対して、後からサブコマンドを追加可能にするツールです。
GitHub CLIの拡張機能（`gh extension`）のような仕組みを、任意のCLIに適用できます。

## 動作原理

### Shell Integration

ユーザーは `.bashrc` や `.zshrc` に以下を追加します：

```bash
eval "$(kakucho init bash)"  # bash の場合
eval "$(kakucho init zsh)"   # zsh の場合
```

これにより、設定ファイルで定義された各CLIに対してシェル関数が定義されます。
シェル関数は以下のロジックで動作します：

1. サブコマンドが拡張として登録されているか確認
2. 登録されていれば拡張コマンドを実行
3. 登録されていなければ元のCLIにそのまま渡す
4. `help` や `-h` の場合は、元のヘルプに拡張コマンドのヘルプを追記

## 設定ファイル

### 配置場所

XDG Base Directory仕様に準拠：
- `$XDG_CONFIG_HOME/kakucho/config.toml`
- デフォルト: `~/.config/kakucho/config.toml`

### 形式 (TOML)

```toml
# ターゲットCLI名をセクション名として定義
# キーはサブコマンド名、値は説明
# 実行ファイルは {target}-{name} の命名規則でPATHから検索される

[hoge]
foo = "Foo command for hoge"   # 実行ファイル: hoge-foo
bar = "Bar command for hoge"   # 実行ファイル: hoge-bar

[fuga]
baz = "Baz command for fuga"   # 実行ファイル: fuga-baz
```

### 拡張コマンドの実行ファイル

拡張コマンドの実行ファイルは `{target}-{name}` という命名規則で、PATH上に配置する必要があります。

例：
- `hoge foo` → `hoge-foo` を実行
- `git hello` → `git-hello` を実行

## コマンド

### `kakucho init <shell>`

指定されたシェル用の初期化スクリプトを出力します。

```bash
kakucho init bash
kakucho init zsh
```

サポートするシェル：
- bash
- zsh

### `kakucho ls [target]`

登録されている拡張コマンドの一覧を表示します。

```bash
# 全ての拡張を表示
kakucho ls

# 特定のターゲットの拡張のみ表示
kakucho ls hoge
```

### `kakucho add <target> <name> [--description <desc>]`

新しい拡張コマンドを追加します。
実行ファイルは `{target}-{name}` という名前でPATH上に存在する必要があります。

```bash
kakucho add hoge foo --description "Foo command"
# hoge-foo がPATH上に存在する必要がある
```

### `kakucho rm <target> <name>`

拡張コマンドを削除します。
ターゲットの拡張がすべて削除されると、ターゲット自体も設定から削除されます。

```bash
kakucho rm hoge foo
```

### `kakucho version`

バージョン情報を表示します。

## ヘルプ表示の拡張

### 動作仕様

`hoge help` または `hoge -h` が実行された場合：

1. 元のCLIのヘルプを取得（`hoge help` または `hoge -h` を実行）
2. 終了コードが0（エラーでない）場合のみ出力を使用
3. 元のヘルプの後に、kakuchoで追加した拡張コマンドの一覧を追記

### 拡張コマンドのヘルプ形式

```
EXTENSIONS (via kakucho):
  foo    Foo command for hoge
  bar    Bar command for hoge
```

### 個別拡張コマンドのヘルプ

`hoge foo -h` のように拡張コマンドに `-h` が渡された場合：

1. 拡張コマンドに `-h` を渡して実行
2. 終了コードが0の場合はその出力を表示
3. 終了コードが0でない場合は、設定ファイルの description を表示

## アーキテクチャ

```
┌─────────────────────────────────────────────────────────────────┐
│  Shell (.bashrc / .zshrc)                                       │
│  eval "$(kakucho init bash)"                                    │
│                                                                 │
│  ┌───────────────────────────────────────────────────────────┐  │
│  │  hoge() {                                                 │  │
│  │    command kakucho exec hoge "$@"                         │  │
│  │  }                                                        │  │
│  └───────────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│  kakucho exec hoge <args>                                       │
│                                                                 │
│  1. 設定ファイル読み込み                                        │
│  2. サブコマンドが拡張かどうか判定                              │
│  3. help/-h の場合は特別処理                                    │
│  4. 拡張なら拡張コマンドを実行、そうでなければ元CLIを実行       │
└─────────────────────────────────────────────────────────────────┘
                              │
              ┌───────────────┴───────────────┐
              ▼                               ▼
┌──────────────────────────┐    ┌──────────────────────────┐
│  拡張コマンド実行        │    │  元のCLI実行             │
│  hoge-foo args           │    │  /usr/bin/hoge args      │
│  (PATH上から検索)        │    │                          │
└──────────────────────────┘    └──────────────────────────┘
```

## 実装詳細

### 使用ライブラリ

- `github.com/spf13/cobra` - CLIフレームワーク
- `github.com/BurntSushi/toml` - TOML パーサー

### ディレクトリ構造

```
kakucho/
├── cmd/
│   └── kakucho/
│       └── main.go
├── internal/
│   ├── config/
│   │   └── config.go      # 設定ファイル読み込み
│   ├── shell/
│   │   └── init.go        # shell integration スクリプト生成
│   └── help/
│       └── help.go        # ヘルプ拡張ロジック
├── go.mod
├── go.sum
└── spec.md
```

## インストール

```bash
go install github.com/qawatake/kakucho/cmd/kakucho@latest
```

## セットアップ

1. インストール後、シェルの設定ファイルに追加：

```bash
# ~/.bashrc または ~/.zshrc
eval "$(kakucho init bash)"  # または zsh
```

2. 拡張コマンドを登録：

```bash
# hoge-foo という実行ファイルをPATH上に配置済みとする
kakucho add hoge foo --description "Foo command"
```

または、設定ファイルを直接編集：

```bash
mkdir -p ~/.config/kakucho
cat > ~/.config/kakucho/config.toml << 'EOF'
[hoge]
foo = "Foo command"
EOF
```

3. シェルを再起動またはsource

```bash
source ~/.bashrc
```

4. 拡張コマンドを使用

```bash
hoge foo  # hoge-foo が実行される
hoge help # 元のヘルプ + 拡張コマンド一覧
```
