# tenki

Open-Meteo を使って天気を表示する Go 製 CLI です。

場所名を Open-Meteo Geocoding API で解決し、Forecast API から現在天気・日別予報・時間別予報を取得します。

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
