# tenki

Open-Meteo を使って天気を表示する Go 製 CLI です。

場所名を Open-Meteo Geocoding API で解決し、Forecast API から現在天気・日別予報・時間別予報を取得します。

## Install

Go 1.26 以降でインストールできます。

```bash
go install github.com/ktsu2i/tenki/cmd/tenki@latest
```

特定バージョンを指定する場合:

```bash
go install github.com/ktsu2i/tenki/cmd/tenki@v0.1.0
```

インストール後、`$(go env GOPATH)/bin` に PATH が通っていれば `tenki` を実行できます。

## Usage

```bash
go run ./cmd/tenki tokyo
go run ./cmd/tenki tokyo --daily --days 3
go run ./cmd/tenki tokyo --hourly --hours 24
go run ./cmd/tenki tokyo --json
go run ./cmd/tenki --version
```

## Development

```bash
go test ./...
```

## Release

`go install` 配布に GitHub Actions は必須ではありません。GitHub に push して SemVer tag を作れば、利用者はその tag を指定してインストールできます。

`gh release create` を使うと、remote に tag がない場合は GitHub 側で tag も作成されます。

```bash
gh release create v0.1.0 --generate-notes
```

tag の作成元 branch や commit を明示する場合:

```bash
gh release create v0.1.0 --target main --generate-notes
```

tag を作る前に確認すること:

```bash
go test ./...
go run ./cmd/tenki --version
```
