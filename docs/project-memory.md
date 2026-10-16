# 项目记忆与数据导入流程

最后更新：2026-10-16。此文档用于新会话接手；运行时版本和数量应查询确认，不能把这里的快照当成数据库实时状态。

## 当前约定

用户已认可当前 MVP 页面和十条玩法样例。现在停止扩量，等预计 2026-10-22 的大版本更新后，由用户发起更新任务，再整理可获取的全量新版本数据。这里的日期不是补丁号，执行时核实实际国服版本与上游资源版本。

目标内容为英雄介绍、多种趣味玩法、各玩法的基础思路、海克斯搭配和装备选择。胜率、选取率、强度排名暂不接入。优先复用已有资料并整理改写，不自行生成一套攻略。缺失内容保持待补充，不补造推荐或装备顺序。

用户已明确接受“全量”以实际能获取并核对的全部资料为范围，不要求每位英雄、每种玩法或所有栏目都有内容。导入前需要逐条润色文案，避免生硬、机械的描述；必要时调整句式、拆分长句和精简重复内容，使其符合玩家阅读习惯。

2026-10-16 用户确认玩法展示应一一对应：每条玩法独立卡片，进入后查看这条玩法自己的操作、海克斯理由与装备。英雄详情取消原来的三个切换栏目，避免跨栏目寻找对应内容。列表保留摘要和核心海克斯，详细内容集中在本条玩法详情；英雄通用基础另列，不混入某条流派。此次只调整前端展示，不扩充或重新导入数据。

玩法手册移除“全部 / 通用玩法 / 趣味玩法 / 基础思路”分类切换，只展示玩法卡片与搜索。搜索支持英雄名、常见别称、玩法及关联海克斯名。`basic` 英雄通用基础仅在英雄详情另列入口（全局搜索仍可查到），不占玩法手册及首页玩法精选；数据仍保留原分类。只有多页时显示分页控件。

数据库位于用户已有阿里云 ECS，连接配置已在项目 `.env` 中；不要在记忆文件中保存凭据。客户端通过本项目 API 读取数据库，不在每次浏览时请求 Data Dragon 数据。图片暂时保留外部地址，不迁移到 ECS。

## 已有数据与文件

2026-10-15 已完成的快照：

- `16.20` 基础图鉴：173 位英雄、223 个海克斯、194 件模式装备、173 组英雄技能。
- `gameplay_guide_previews`：六位英雄共十条未审核样例，`reviewed: false`。塔姆四条（含通用基础）、盖伦两条，其余为卡特琳娜、韦鲁斯、卡尔玛、索拉卡各一条。
- `gameplay_guides`：尚无已审核玩法，不要把预览表称为正式攻略库。
- 十条样例已完成自然文案改写并重新入库，来源和关联仍保留。

主要表：`champion_patch_profiles`、`augment_patch_profiles`、`item_patch_profiles`、`champion_patch_skills` 存放版本图鉴；`data_sources` 保存来源；`gameplay_guide_previews` 和 `gameplay_guides` 分别存放预览与审核玩法。

资料入口：

- [README](../README.md)：静态图鉴同步、版本发布和服务运行。
- [样例来源与导入命令](guide-previews/README.md)：十条玩法的来源 URL、说明、预览开关。
- [玩法字段与显示规则](gameplay-content.md)：JSON 结构、校验与单条玩法关联。
- [候选数据源调查](reusable-gameplay-sources.md) 和 [来源评估](gameplay-source-assessment.md)：历史调查，更新时重新核查。
- `docs/guide-previews/*.json`：目前六份可重复导入的样例文件。
- `data/16.20/`：固定版本图鉴、原始快照和 manifest。部分数据文件被 Git 忽略，新检出环境可能不存在。

## 采集方式：区分图鉴与玩法

### 基础图鉴

已实现批量工具：`cmd/importer` 的 `fetch` / `sync`。Data Dragon 提供简中英雄、技能等资料；CommunityDragon 提供 KIWI 海克斯及模式装备所需数据。实现位于 `internal/staticdata/`。

固定实际资源版本采集，不使用未固定版本的 `latest` 覆盖旧快照。数据存入 `data/<patch>/` 后再导入自己的数据库。`sync` 包含基础图鉴抓取和导入，不包含玩法攻略采集。

### 玩法与联动

目前采用海斗小助手公开网页已有内容，读取英雄页的玩法/机制提示/装备候选，再用海克斯页的英雄适配资料交叉核对。例如：

- `https://lolhaidou.cn/hero/tahmkench.html`
- `https://lolhaidou.cn/hero/karma.html`
- `https://lolhaidou.cn/augment/etherealweapon.html`
- `https://lolhaidou.cn/augment/firstaidkit.html`

原页面快照在 `data/16.20/guide-previews/raw/`，补充来源 URL、哈希与保存时间见同目录上级的 `synergy-sources.json`。不要依赖 `/tmp` 中旧会话脚本；它们不是项目内可复用的采集工具。

**尚未实现全量玩法自动采集器，也没有确认现成的完整攻略接口。** 后续全量任务需先检查新版本页面和实际可用字段，再编写或完善项目内采集/转换脚本。不要把公开机制提示说成完整攻略，也不要把图鉴 `sync` 说成全量玩法同步。

现有公开内容未确认开放转载许可，样例保留原来源信息和未审核状态；语言改写不会自动改变许可或审核状态。已有调查还列出 Wiki 等机制资料候选，不能把代码仓库许可直接视为其上游数据许可。

## 转换规则

1. 保存原始网页/接口响应、来源 URL、采集时间、来源版本和哈希，便于后续追溯和比对。
2. 用目标版本图鉴匹配英雄、海克斯、装备 ID；中文名用于核对，不能跨版本盲目复用 ID。
3. 每种趣味玩法单独一条 `category: fun`，使用稳定 `id` 和 `playstyle_name`。
4. 该玩法的操作思路放 `basic_sections`；海克斯放 `core_augment_ids`、`alternative_augment_ids` 与带理由的 `recommendations`；装备放本条 `stages`。英雄通用基础另存 `category: basic`。
5. 装备候选池与购买顺序分开表达。没有装备资料时，预览允许 `stages: []`；不要借用别的玩法填空。
6. 正文直接解释搭配和操作，不写“来源明确提到”“公开资料认为”等过程描述。来源在 `source`/`references`，未确认的机制集中写“版本提示”。不把改写当成实测。
7. 海克斯详情由玩法记录反向关联英雄实例和玩法，无须另编一套独立推荐。相同英雄的多条玩法保留各自理由；展示优先级不等于统计排名。
8. 文案润色属于导入流程：逐条检查摘要、搭配理由、操作说明和装备说明，用简洁自然的语言解释“为什么适合、如何配合”。保留原有技能名称、数值、触发条件与限制，不为了好读增加未经支持的结论，也不把候选写成必选。来源描述与审核备注不混入攻略正文。

完整字段规则以 `internal/importer/guide.go`、`docs/gameplay-content.md` 为准。

## 新版本执行顺序

以下为后续任务的操作流程，本次只写文档，不执行采集或数据库变更。

1. 读取本文件与来源文档，运行 `go run ./cmd/importer status` 确认当前状态，核实国服实际补丁及 Data Dragon 对应版本；不得直接沿用示例版本。
2. 在新的版本快照同步基础图鉴。设置真实版本后执行：

   ```bash
   HEXIS_DDRAGON_VERSION='替换为已核实的资源版本'
   go run ./cmd/importer sync --version "$HEXIS_DDRAGON_VERSION"
   go run ./cmd/importer status
   ```

   此命令不自动发布，也不采集玩法；已发布版本的基础图鉴不可用 `sync` 覆盖。核对下载的 manifest、数量、KIWI 模式、描述与新增/移除 ID。
3. 检查玩法来源是否跟进新版本，采集所有实际可获取的内容，并按上述规则转换为新版 JSON。入库前逐条润色并对照原始资料确认语义未改变，再抽查页面实际阅读效果。旧玩法保留在旧版本，重新核对后才能写入新版本。列出缺失/过期内容，不能为了“全量”把未知内容补齐。
4. 先用小批量预览检验转换和显示，再批量导入预览表。命令模板：

   ```bash
   HEXIS_PATCH='替换为目标补丁号'
   HEXIS_GUIDE_FILE='替换为整理后的JSON路径'
   HEXIS_SOURCE_NAME='替换为该英雄或来源批次的独立名称'
   HEXIS_SOURCE_URL='替换为实际来源URL'
   go run ./cmd/importer import --patch "$HEXIS_PATCH" --kind guide-previews \
     --file "$HEXIS_GUIDE_FILE" --source "$HEXIS_SOURCE_NAME" \
     --source-url "$HEXIS_SOURCE_URL" \
     --source-license "填写实际许可状态" --source-attribution "填写署名、版本与采集日期"
   ```

   `data_sources` 按名称更新，英雄来源应使用各自名称，避免后导入覆盖前一英雄的来源 URL。具体旧版本示例见 `guide-previews/README.md`。导入器校验同版本引用，单次导入失败回滚；分文件导入并非整个批次的原子事务。相同补丁与玩法 ID 重复导入更新该条，不自动删除来源已移除的条目，需单独核对差异。
5. 预览要求 `reviewed: false`。正式 `--kind guides` 要求已核对的 `reviewed: true` 且装备阶段非空；不得自动改状态绕过审核，缺资料的条目继续留在预览。
6. 验证数量、来源、重复 ID、关联、图标和分页；检查每张玩法卡片进入自身详情、操作与搭配均来自本条记录，以及海克斯详情能反查英雄和玩法。相关检查：

   ```bash
   go test ./internal/importer ./internal/staticdata
   node scripts/test-mini-core.cjs
   node scripts/test-mini-relations.cjs
   git diff --check
   ```

7. 静态版本准备好且核对通过后，按任务范围执行 `go run ./cmd/importer publish --patch "$HEXIS_PATCH"` 切换图鉴默认版本；此命令不会把预览玩法转成审核玩法。新版本若尚无可发布玩法，应如实保留缺失状态。

## 本地展示与后续维护

- 微信项目为 `miniprogram/`，API 地址配置在 `miniprogram/app.js`，当前为正式接口 `https://hexis.xuninfo.top`。2026-10-16 用户确认后端和 HTTPS 已部署，健康检查返回正常；项目已开启域名校验。上传前需在小程序后台添加该 request 合法域名，按 `miniprogram/README.md` 完成体验版验证和审核发布。代码配置完成不代表已在微信后台上传或发布。
- 开发预览后端需 `GUIDE_PREVIEW_ENABLED=1`，API 为 `/api/v1/guides?preview=1`；正常查询始终读审核表。浏览器使用 `/mini/?guides=preview`，微信开发版自动启用预览，体验版/正式版不启用。
- 客户端共用 `miniprogram/lib/core.js`。修改共用前端后运行 `node scripts/build-mini-preview.cjs`；Go 内嵌预览资源需重新构建并重启本地服务。
- 再次交接时更新本文件中的实际版本、数量、来源、脚本位置、完成状态和待办；不记录临时进程 PID，也不承诺后台自动更新。当前没有设置定时采集任务。
