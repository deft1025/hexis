# 海克斯大乱斗（Hexis）

面向海克斯大乱斗数据查询小程序的 Go + MySQL 后端。提供按版本查询英雄、强化符文、装备和双英雄组合的接口。

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

## 真实静态资料

导入工具读取 `.env` 中的 MySQL 配置，支持阿里云 ECS 上的数据库。现有数据库首次接入装备时，执行增量迁移创建装备资料表，不影响已有英雄和符文表。再检查现有版本，下载固定版本的中文英雄、海克斯大乱斗强化符文和装备：

```bash
go run ./cmd/importer migrate
go run ./cmd/importer status
go run ./cmd/importer fetch --version 16.20.1
```

下载结果默认保存到 `data/16.20/`，也可通过 `--out` 指定目录。`champions.json`、`augments.json`、`items.json` 和 `skills.json` 可直接导入；`raw/` 保留原始资料，`manifest.json` 记录来源 URL、文件 SHA-256、下载时间、版本、模式、数量和包含可变数值的效果说明。下载命令不连接数据库，不写入统计数据。

英雄来自 Riot Data Dragon 的 `zh_CN` 和 `en_US` 资料；强化符文来自 CommunityDragon 的 `KIWI` 模式清单、同版本模式资源和简体中文文本。只导入清单内的符文，不混入 `CHERRY`（斗魂竞技场）或 `KIWI_JADE` 变体的独有资料。效果说明会去除客户端标签和图标标记，解析可确认的常量、属性公式、等级成长、近战/远程差异及部分任务条件。对局动态技能目标保持通用表达，未知计算继续保留提示。2026-10-19 已修复当前版本的 51 条数据库描述，剩余 6 条未完全解析；固定快照重算、备份和受控更新见 [海克斯描述修复](docs/augment-descriptions.md)。

装备以 CommunityDragon 地图配置中 `Maps/Shipping/Map12/Modes/KIWI` 引用的装备清单为范围，从同版本 Data Dragon 中文资料提取名称、图标、效果、分类、总价、卖价、合成配方和静态属性。排除隐藏及内部条目，保留可见的变形装备；不单凭地图编号筛选，避免混入其他模式。`16.20` 共 194 件装备。

以下示例使用静态资源版本 `16.20.1`，数据库版本为 `16.20`；不代表已经核实国服上线时间或国服专属调整。实际使用时应与国服当前游戏版本核对，同时替换命令中的版本和文件目录：

```bash
go run ./cmd/importer import --patch 16.20 --kind champions --file ./data/16.20/champions.json --source "Riot Data Dragon 16.20.1" --source-url "https://ddragon.leagueoflegends.com/cdn/16.20.1/data/zh_CN/champion.json" --source-attribution "Riot Games"
go run ./cmd/importer import --patch 16.20 --kind augments --file ./data/16.20/augments.json --source "CommunityDragon KIWI 16.20" --source-url "https://raw.communitydragon.org/16.20/game/maps/modespecificdata/kiwi.bin.json" --source-attribution "Riot Games assets via CommunityDragon; KIWI mode; zh_CN"
go run ./cmd/importer import --patch 16.20 --kind items --file ./data/16.20/items.json --source "Riot Data Dragon + CommunityDragon KIWI items 16.20.1" --source-url "https://raw.communitydragon.org/16.20/game/data/maps/shipping/map12/map12.bin.json" --source-attribution "Riot Games assets; zh_CN; KIWI"
go run ./cmd/importer import --patch 16.20 --kind skills --file ./data/16.20/skills.json --source "Riot Data Dragon skills 16.20.1" --source-url "https://ddragon.leagueoflegends.com/cdn/16.20.1/data/zh_CN/champion.json" --source-attribution "Riot Games"
go run ./cmd/importer status
go run ./cmd/importer publish --patch 16.20
```

先检查数量、来源和效果说明，再发布。已发布图鉴版本不能覆盖；仅允许为完全没有装备或技能资料的已发布版本一次性补入对应资料，后续仍不可覆盖。审核后的编辑玩法可以独立维护，详见 `docs/gameplay-content.md`。只导入静态资料时 API 中的 `stats` 为 `null`，页面不显示虚构的胜率和选取率。

英雄技能保存在 `champion_patch_skills`，浏览时只查询 MySQL。现有版本没有技能时可执行 `go run ./cmd/importer backfill-skills --version 16.20.1` 一次补齐。图鉴和技能图片继续使用数据库中保存的固定版本 Data Dragon 或 CommunityDragon 地址，由客户端直接加载；图片托管及 CDN 优化留待后续。

### 后续版本更新

改版后需要获取新版资料，但无需重新建表、手工录入或修改页面。`sync` 自动执行增量迁移，下载并导入英雄、海克斯、装备和技能资料，默认保持未发布状态。发布前检查所有英雄技能已入库；不下载图片文件。以下 `16.21.1` 仅作示例，需替换为与国服版本核对过且数据源已提供的版本：

```bash
go run ./cmd/importer sync --version 16.21.1
go run ./cmd/importer status
go run ./cmd/importer publish --patch 16.21
```

确认版本和来源后也可用 `sync --version 16.21.1 --publish` 一次完成。页面默认读取最新已发布版本，旧版本快照保留。同步失败时不会发布，可修复后重试；已发布版本不能同步覆盖。目前没有设置定时任务，需要每次更新时执行命令。

统计范围固定为**国服海克斯大乱斗**。后续统计必须核实地区为国服、模式为海克斯大乱斗、版本和样本量一致，不使用国际服、普通大乱斗、斗魂竞技场或云顶统计替代。静态资料公开下载不等于获得国服真实对局统计。英雄专属符文和出装仍需增加对应表与数据来源。

## 手动导入

玩法资料的公开采集、文案整理、旧记录备份和批次导入见 [玩法扩充与版本更新](docs/guide-expansion/README.md)。2026-10-19 已补入 284 条未审核预览，正式玩法仍为空；更新前先读 [项目记忆](docs/project-memory.md)，不要重导已完整的当前基础图鉴或清空旧表。

配置好 `.env`，准备 JSON 数组，然后执行：

```bash
go run ./cmd/importer import --patch 15.21 --kind champions --file ./champions.json
go run ./cmd/importer import --patch 15.21 --kind augments --file ./augments.json
go run ./cmd/importer import --patch 15.21 --kind items --file ./items.json
go run ./cmd/importer import --patch 15.21 --kind combos --file ./combos.json
go run ./cmd/importer publish --patch 15.21
```

`champions.json` 和 `augments.json` 的每条记录必须包含字符串 `id` 和 `name`。可选字段包括 `name_en`、`icon_url`、`description`、`rarity`、`sample_count`、`pick_rate`、`win_rate` 和 `tier`。统计比例为 0 到 1，填写统计字段时必须提供正整数 `sample_count`。只导入静态资料也可以发布，但需额外导入该版本完整的 `skills.json`。

`items.json` 使用下载工具生成的格式，包含 `id`、`name`、`mode: "KIWI"`、图标、效果、分类、价格、购买状态、配方和静态属性，不接受装备对局统计。

`combos.json` 的记录格式为 `{"champion_ids":["Aatrox","Ahri"],"sample_count":100,"pick_rate":0.05}`。组合涉及的英雄必须先导入同一版本。导入来源可以通过 `--source`、`--source-url`、`--source-license` 和 `--source-attribution` 标注；导入记录会保存文件哈希，未发布版本重复导入使用 Upsert。发布前需导入至少一个英雄、一个强化符文和一件装备，补齐英雄技能，并人工审核来源和数据。

## 查询接口

- `GET /healthz`：进程健康检查
- `GET /api/v1/patches/latest`：最新已发布版本
- `GET /api/v1/champions`、`GET /api/v1/champions/{id}`：英雄列表与详情
- `GET /api/v1/champions/{id}/skills`：查询数据库中的版本化技能资料
- `GET /api/v1/augments`：强化符文列表
- `GET /api/v1/items`：海克斯大乱斗装备列表，`attributes` 为静态属性，`stats` 暂为 `null`
- `GET /api/v1/combos`：双英雄组合列表

查询接口默认使用最新已发布版本，可通过 `?patch=15.21` 指定历史版本。列表接口支持 `limit`（1–100，默认 100）和 `offset`（默认 0）。返回的 `stats` 为 `null` 表示该版本只有静态资料；来源信息随资料和统计一起返回。

## 网页

微信原生小程序已加入 `miniprogram/`，以英雄与海克斯联动、趣味玩法和装备选择为核心。浏览器预览访问 `http://127.0.0.1:8182/mini/`，使用真实版本图鉴，不显示胜率和选取率。开发者工具运行方式见 `miniprogram/README.md`，玩法关联录入见 `docs/gameplay-content.md`。首次接入运行 `go run ./cmd/importer migrate` 创建玩法表；尚未导入审核玩法时显示对应空状态。

启动服务后访问 `http://127.0.0.1:8182/`。首页提供英雄、强化符文、装备和英雄组合入口。四类列表页面提供上下分页，默认每页 20 条，可切换为 50 或 100 条；搜索、筛选和排序作用于同版本的完整资料。当前页码保存在 URL 中，刷新后保留页码。强化符文图标使用深色底与稀有度边框，效果说明完整换行显示。英雄详情页按强化符文和出装路线组织，目前因缺少关联数据而显示待接入说明。没有已发布版本或统计数据时，页面会显示相应提示。

装备页通过 `/api/v1/items` 读取 MySQL 中的版本资料，展示名称、图标、多种分类、总价和完整效果，支持名称、效果或 ID 搜索及分类筛选、价格排序；尚无英雄专属出装或装备表现统计。图鉴和技能图标使用固定版本的 Data Dragon 或 CommunityDragon 地址，加载失败时显示占位符。正常浏览只从我们的 API 查询文字资料，不向外部源获取资料 JSON。

对局内选英雄后的页面规划见 `docs/product-roadmap.md`。当前尚无英雄与符文、英雄与装备的关联数据，页面不会把全局表现误写成专属推荐。

## 验证

```bash
curl http://127.0.0.1:8182/healthz
```
