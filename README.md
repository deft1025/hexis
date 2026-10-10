# 海克斯大乱斗（Hexis）

面向海克斯大乱斗数据查询小程序的 Go + MySQL 后端。提供按版本查询英雄、强化符文和双英雄组合的接口。

## 配置

配置通过 `.env` 文件管理，从 `.env.example` 复制一份：

```bash
cp .env.example .env
# 编辑 .env，填入数据库密码等
```

环境变量说明：

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `DB_HOST` | `127.0.0.1` | MySQL 主机 |
| `DB_PORT` | `3306` | MySQL 端口 |
| `DB_USER` | `root` | MySQL 用户 |
| `DB_PASSWORD` | （空） | MySQL 密码 |
| `DB_NAME` | `hexis` | 数据库名 |
| `DB_ENABLED` | `1` | 是否连接数据库（`0` 可启动无数据模式） |
| `HTTP_ADDR` | `:8182` | 服务监听地址 |
| `GIN_MODE` | `debug` | Gin 模式（debug/release） |

## 启动

### Docker Compose（推荐）

```bash
docker-compose up -d
```

会自动拉起 MySQL（并执行 `sql/` 下的初始化脚本）和 API 服务。

### 本地直接运行

```bash
go mod tidy
go run ./cmd/api
```

如需手动初始化数据库：

```bash
mysql -uroot -p < sql/001_init.sql
mysql -uroot -p < sql/002_queries.sql
```

## 手动导入

配置好 `.env`，准备 JSON 数组，然后执行：

仓库内提供了一套用于联调的轻量种子数据，位于 `data/`。它只包含少量资料，不包含真实统计数据；正式上线前应替换为已授权的数据。

```bash
go run ./cmd/importer import --patch 15.21 --kind champions --file ./champions.json
go run ./cmd/importer import --patch 15.21 --kind augments --file ./augments.json
go run ./cmd/importer import --patch 15.21 --kind combos --file ./combos.json
go run ./cmd/importer publish --patch 15.21
```

使用仓库内种子数据时，可直接执行：

```bash
go run ./cmd/importer import --patch 15.21 --kind champions --file ./data/champions.json --source "Riot Data Dragon" --source-url "https://developer.riotgames.com/docs/lol#data-dragon"
go run ./cmd/importer import --patch 15.21 --kind augments --file ./data/augments.json --source "MVP seed data" --source-attribution "Replace before production"
go run ./cmd/importer publish --patch 15.21
```

`champions.json` 和 `augments.json` 的每条记录必须包含字符串 `id` 和 `name`。可选字段包括 `name_en`、`icon_url`、`description`、`rarity`、`sample_count`、`pick_rate`、`win_rate` 和 `tier`。统计比例为 0 到 1，填写统计字段时必须提供正整数 `sample_count`。只导入静态资料也可以发布。

`combos.json` 的记录格式为 `{"champion_ids":["Aatrox","Ahri"],"sample_count":100,"pick_rate":0.05}`。组合涉及的英雄必须先导入同一版本。导入来源可以通过 `--source`、`--source-url`、`--source-license` 和 `--source-attribution` 标注；导入记录会保存文件哈希，重复导入使用 Upsert。发布前需导入至少一个英雄和一个强化符文，并人工审核来源和数据。

## 查询接口

- `GET /healthz`：进程健康检查
- `GET /api/v1/patches/latest`：最新已发布版本
- `GET /api/v1/champions`、`GET /api/v1/champions/{id}`：英雄列表与详情
- `GET /api/v1/augments`：强化符文列表
- `GET /api/v1/combos`：双英雄组合列表

查询接口默认使用最新已发布版本，可通过 `?patch=15.21` 指定历史版本。列表接口支持 `limit`（1–100，默认 100）和 `offset`（默认 0）。返回的 `stats` 为 `null` 表示该版本只有静态资料；来源信息随资料和统计一起返回。

## 验证

```bash
curl http://127.0.0.1:8182/healthz
```
