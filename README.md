# tenki

Open-Meteo を使って天気を表示する Go 製 CLI です。

場所名を Open-Meteo Geocoding API で解決し、Forecast API から現在天気・日別予報・時間別予報を取得します。
API key は不要です。

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

基本形は、天気を見たい場所名を 1 つ渡します。

```bash
tenki <location>
```

例:

```bash
tenki tokyo
tenki osaka
tenki kyoto
```

デフォルトでは、解決された場所名、現在天気、今日の最高/最低気温、今日の降水確率、明日以降のざっくり予報を表示します。

```text
Tokyo, Japan
Now: 22C, Partly cloudy
Today: 17C / 24C, rain 20%

Fri  Sunny         25C / 16C
Sat  Cloudy        23C / 18C
Sun  Light rain    21C / 17C
```

## Options

### Daily forecast

日別予報だけを表示します。

```bash
tenki tokyo --daily
tenki tokyo --days 3
```

`--days N` は `1` から `7` まで指定できます。`--days` を指定した場合は、自動的に daily 表示になります。

### Hourly forecast

時間別予報だけを表示します。

```bash
tenki tokyo --hourly
tenki tokyo --hours 24
```

`--hours N` は `1` から `24` まで指定できます。`--hours` を指定した場合は、自動的に hourly 表示になります。

### JSON output

```bash
tenki tokyo --json
tenki tokyo --daily --json
tenki tokyo --hourly --hours 24 --json
```

`--json` を付けると、整形済みの JSON だけを標準出力に出します。

### Version

```bash
tenki --version
```

## Flag rules

- `--daily` と `--hourly` は同時に指定できません。
- `--days` と `--hours` は同時に指定できません。
- `--daily` と `--hours`、`--hourly` と `--days` も同時に指定できません。
- `--days` は `1..7`、`--hours` は `1..24` を指定できます。
