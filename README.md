# tenki

Open-Meteo を使って天気を表示する Go 製 CLI です。

現時点では CLI の雛形だけを実装しています。天気取得処理はまだ未実装です。

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
