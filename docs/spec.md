# tenki-cli 最小仕様

更新日: 2026-05-21

## 1. 何を作るか

`tenki-cli` は Open-Meteo を使って天気を表示する、シンプルな Go 製 CLI とする。

まず作るのは次のような使い方でよい。

- `tenki tokyo`
- `tenki tokyo --hourly`
- `tenki tokyo --daily`
- `tenki tokyo --days 3`
- `tenki tokyo --hours 24`
- `tenki tokyo --json`

目的は「場所名を入れたら、すぐ天気が見える」こと。

## 2. v1 で入れる機能

### 2.1 デフォルト表示

```bash
tenki tokyo
```

このコマンドでは次を表示する。

- 解決された場所名
- 現在の天気
- 今日の最高気温 / 最低気温
- 今日の降水確率
- 明日以降 2-3 日のざっくり予報

出力イメージ:

```text
Tokyo, Japan
Now: 22C, Partly cloudy
Today: 17C / 24C, rain 20%

Fri  Sunny        25C / 16C
Sat  Cloudy       23C / 18C
Sun  Light rain   21C / 17C
```

### 2.2 daily 表示

```bash
tenki tokyo --daily
tenki tokyo --days 3
```

このコマンドでは、日別予報だけを表示する。

表示項目:

- 日付
- 最高気温
- 最低気温
- 天気
- 降水確率

出力イメージ:

```text
Tokyo, Japan

Fri  Sunny        25C / 16C  10%
Sat  Cloudy       23C / 18C  40%
Sun  Light rain   21C / 17C  70%
```

`--days` で日数を指定できるようにする。
まずは `1` から `7` まで対応すれば十分。

### 2.3 hourly 表示

```bash
tenki tokyo --hourly
tenki tokyo --hours 24
```

このコマンドでは、今から先の時間別予報を表示する。

表示項目:

- 時刻
- 気温
- 天気
- 降水確率

出力イメージ:

```text
Tokyo, Japan

00:00  21C  Cloudy       10%
01:00  20C  Cloudy       10%
02:00  19C  Light rain   40%
```

デフォルトは「次の 12 時間」でよい。
`--hours` で件数を変えられるようにする。
まずは `1` から `24` まで対応すれば十分。

### 2.4 JSON 表示

```bash
tenki tokyo --json
tenki tokyo --hourly --json
tenki tokyo --daily --json
```

`--json` を付けたら、人間向け整形ではなく JSON を返す。

JSON で返す内容:

- 解決された場所
- 現在天気
- daily データ
- hourly データ

v1 では Open-Meteo の生レスポンスをそのまま返さず、CLI 側で使いやすい形に少し整えた JSON を返す。

## 3. 入力

v1 では引数は場所名だけでよい。

受ける例:

- `tokyo`
- `osaka`
- `kyoto`

座標直指定や複数地点対応はやらない。

## 4. 内部の動き

### 4.1 場所解決

ユーザーが入れた地名は Geocoding API で解決する。

使う API:

- `https://geocoding-api.open-meteo.com/v1/search`

方針:

- 検索結果の先頭 1 件を使う
- どの場所を採用したかは必ず表示する
- 見つからなければエラー終了する

### 4.2 天気取得

場所が決まったら Forecast API を呼ぶ。

使う API:

- `https://api.open-meteo.com/v1/forecast`

使うデータは最低限だけにする。

デフォルト表示で使うもの:

- `current`
  - `temperature_2m`
  - `weather_code`
- `daily`
  - `weather_code`
  - `temperature_2m_max`
  - `temperature_2m_min`
  - `precipitation_probability_max`

hourly 表示で使うもの:

- `hourly`
  - `temperature_2m`
  - `weather_code`
  - `precipitation_probability`

timezone は `auto` を使う。

## 5. フラグ

v1 のフラグは最小限にするが、表示切り替えに必要なものは入れる。

- `--daily`
- `--hourly`
- `--days`
- `--hours`
- `--json`
- `--help`
- `--version`

フラグの扱い:

- `--daily` は daily 表示に切り替える
- `--hourly` は hourly 表示に切り替える
- `--days <n>` を指定したら daily 表示として扱う
- `--hours <n>` を指定したら hourly 表示として扱う
- `--daily` と `--hourly` の同時指定は v1 ではエラーにする

それ以外はまだ入れない。

入れないもの:

- `--unit`
- `--model`
- `--lang`
- `--country`
- `--no-cache`

## 6. 表示ルール

- 人間向け表示と JSON 表示だけ作る
- 人間向け表示は見やすいプレーンテキストでよい
- weather code は CLI 側で短い文字列に変換する
- `--json` 時は JSON だけを stdout に出す

例:

- `0` -> `Clear`
- `1` -> `Mainly clear`
- `2` -> `Partly cloudy`
- `3` -> `Overcast`
- `61` -> `Light rain`

## 7. エラー

最低限、次だけ分かればよい。

- 場所が見つからない
- API 呼び出しに失敗した
- 引数がない
- `--days` や `--hours` の値が不正

エラーは stderr に出して、非 0 で終了する。

## 8. 実装方針

実装はできるだけ軽くする。

- Go
- 標準ライブラリ中心
- HTTP は `net/http`
- JSON は `encoding/json`

CLI ライブラリは必須ではない。
最初は標準ライブラリでもよい。

構成も最低限でよい。

```text
cmd/tenki/
internal/geocode/
internal/forecast/
internal/output/
```

## 9. 実装順

この順で作る。

1. `tenki <location>` の引数処理
2. geocoding で場所解決
3. current + daily の取得
4. デフォルト表示
5. `--daily` と `--days` の追加
6. `--hourly` と `--hours` の追加
7. `--json` の追加

## 10. やらないこと

最初はやらない。

- キャッシュ
- 設定ファイル
- 環境変数
- 単位切り替え
- モデル指定
- 履歴天気
- 複数地点

まずは `tenki tokyo` が気持ちよく使えるところまでで十分。
