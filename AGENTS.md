# AGENTS.md

このリポジトリは Open-Meteo を使って天気を表示する Go 製 CLI `tenki` の実装です。

## 現在の構成

```text
.
├── cmd/tenki/          # CLI バイナリの entrypoint
├── internal/cli/       # Kong による引数・フラグ定義と実行入口
├── internal/geocode/   # Open-Meteo Geocoding API による場所解決
├── internal/forecast/  # Open-Meteo Forecast API による天気取得
├── internal/output/    # プレーンテキスト / JSON 出力整形
├── docs/spec.md        # v1 の最小仕様
├── go.mod
├── go.sum
└── README.md
```

## 実装方針

- Go は `go.mod` の `go 1.26` を前提にする。
- CLI パーサーは `github.com/alecthomas/kong` を使う。
- CLI 定義の構造体名は `CLI` とし、receiver は `c` を使う。
- `cmd/tenki/main.go` は薄く保ち、実処理は `internal` 配下に置く。
- プロセス終了コードを扱う入口は `cli.Main`、テストしやすい入口は `cli.Run` にする。
- ユーザー向けエラーは `stderr` に出し、非 0 で終了する。
- 標準出力には通常出力または JSON だけを出す。`--json` 時は余計な文言を混ぜない。

## パッケージ分割

仕様書の v1 実装では、次の分割を基本にする。

```text
internal/cli/       # フラグ解釈と上位の orchestration
internal/geocode/   # Open-Meteo Geocoding API による場所解決
internal/forecast/  # Open-Meteo Forecast API による天気取得
internal/output/    # プレーンテキスト / JSON 出力整形
```

`internal/cli` はフラグ解釈と上位の orchestration に寄せ、API レスポンスの詳細や表示整形を抱え込ませない。

## 現在の CLI

Open-Meteo の Geocoding API と Forecast API を呼び、解決された場所名、現在天気、日別予報、時間別予報を表示する。

対応済みの入力:

```bash
go run ./cmd/tenki tokyo
go run ./cmd/tenki tokyo --daily --days 3
go run ./cmd/tenki tokyo --hourly --hours 24
go run ./cmd/tenki tokyo --json
go run ./cmd/tenki --version
```

出力:

- デフォルト表示では、解決された場所名、現在天気、今日の最高/最低気温、今日の降水確率、明日以降のざっくり予報を表示する。
- `--daily` / `--days` では日別予報を表示する。
- `--hourly` / `--hours` では時間別予報を表示する。
- `--json` では CLI 側で整えた JSON だけを stdout に出す。

フラグの基本ルール:

- `--daily` と `--hourly` の同時指定はエラー。
- `--days` 指定時は daily 表示として扱う。
- `--hours` 指定時は hourly 表示として扱う。
- `--days` は `1..7`、`--hours` は `1..24` を許可する。

## 開発コマンド

```bash
go test ./...
go run ./cmd/tenki --help
go run ./cmd/tenki tokyo
```

サンドボックス環境で Go build cache に書けない場合は、次のように workspace 外の許可された一時ディレクトリを使う。

```bash
GOCACHE=/private/tmp/tenki-gocache go test ./...
GOCACHE=/private/tmp/tenki-gocache go run ./cmd/tenki tokyo
```

サンドボックス環境では外部ネットワークへの DNS 解決が失敗することがある。その場合でも unit test は HTTP client fake で検証し、実 API の疎通確認だけ承認付きで実行する。

## コーディング規約

- 変更後は `gofmt` を実行する。
- テストは `go test ./...` を基本にする。
- 文字列処理だけで済ませず、HTTP/JSON は標準ライブラリの `net/http` と `encoding/json` を使う。
- 新しい外部依存は、明確に必要な場合だけ追加する。現時点の外部依存は Kong のみ。
- 仕様判断に迷ったら [docs/spec.md](docs/spec.md) を優先する。
