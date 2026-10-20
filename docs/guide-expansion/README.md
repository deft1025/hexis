# 玩法资料扩充与版本更新

2026-10-19 的整理针对 `16.20`，基础英雄、海克斯、装备、技能图鉴已有快照，本次只扩充玩法与关联。数据库实时结果以 `go run ./cmd/importer status` 为准。

## 文件与命令

- `cmd/guidefetch`：读取海斗小助手公开 sitemap，顺序保存公开英雄页；每页间隔 300 毫秒，保存 URL、页面日期、采集时间和 SHA-256。提取公开构建、装备候选、基础提示、正向机制与避坑。不会进入小程序或获取隐藏内容，不把胜率清单转成推荐。
- `docs/guide-expansion/16.20-curation.json`：逐条整理的玩家文案、每个核心海克斯的理由、需放入版本提示的异常交互及暂缓项目。
- `16.20-source-hashes.json`：上述文案对应的原文哈希。原文变化时转换器拒绝继续，需要重新对照、整理，不允许只改哈希绕过复核。
- `scripts/prepare-guide-expansion.cjs`：匹配目标图鉴、合并旧基础记录、生成逐英雄 JSON、`batch.json` 和 `coverage.json`。未知 ID 不模糊匹配，负向交互不生成推荐，已有同路线不重复补入。
- `cmd/importer backup-guides`：只读导出目标版本的正式/预览玩法及其来源；输出权限 0600，拒绝覆盖已有备份。不是整库备份，也不包含连接信息。恢复指定记录时从备份 `rows[].content` 提取 JSON，并使用原来源元数据走原有导入命令，不清空表。
- `cmd/importer import-guide-batch`：默认逐文件执行完整校验后回滚；`--apply` 先校验全批，再逐文件事务写入预览。正式提交不是整批原子事务，中断时报告已提交进度；修复后同 ID 可重复运行。默认试导入可能消耗自增序号，但不保留事务内记录。

本次运行文件位于忽略目录 `data/16.20/`：

```text
guide-sources/sitemap.xml
guide-sources/raw/*.html
guide-sources/pages.json
backups/guides-before-expansion-20261019.json
guide-expansion/*.json
guide-expansion/batch.json
guide-expansion/coverage.json
```

本次可复现命令：

```bash
go run ./cmd/guidefetch --out data/16.20/guide-sources
node scripts/prepare-guide-expansion.cjs 16.20 data/16.20/guide-sources docs/guide-expansion/16.20-curation.json data/16.20/backups/guides-before-expansion-20261019.json data/16.20/guide-expansion
go run ./cmd/importer import-guide-batch --file data/16.20/guide-expansion/batch.json
go run ./cmd/importer import-guide-batch --file data/16.20/guide-expansion/batch.json --apply
go run ./cmd/importer status
node scripts/test-guide-expansion.cjs
```

采集器复用指定目录中的已有 sitemap 和原文，适合中断续跑和离线重新解析；**更新采集必须使用全新目录**，否则不会重新获取现有文件。新检出没有 `data/` 时，先获取原始快照与数据库备份，不能假定本地忽略数据存在。

## 本次覆盖与缺口

173 个公开英雄页：122 条构建方案、415 条机制/避坑提示。新增构建记录 116 条、基础记录 158 条，更新旧基础记录 2 条；保留其余旧记录 8 条。合计预览 284 条，覆盖 163 位英雄，其中 124 条 `fun`、160 条 `basic`。基础记录不是额外流派，不在玩法手册充数。

已实际提交并查询确认，关联 99 个海克斯。写入后备份为 `data/16.20/backups/guides-after-expansion-20261019.json`，284 条内容与原记录加导入批次逐条一致。Go、原有客户端测试及扩充数据测试通过，本地 API 返回预览三页 100/100/84、无重复 ID 且来源齐全；正式接口仍为空，海克斯统计保持 `null`。未进行微信开发者工具或浏览器截图验收。

4 条已有同路线（盖伦、卡尔玛、塔姆、索拉卡）未重复创建。塞纳“活力再生电刀流”、巴德“轨道炮”因核心海克斯无法精确映射暂缓。“任务：钢化你心”经名称核对对应 `ARAM_Quest_SteelYourHeart`。

2026-10-19 用户确认“活力再生”当前版本已不存在，要求玩法排除当前图鉴中没有的海克斯。整理器不再把无法映射的机制放进版本提示、摘要或参考链接；`coverage.json` 的 `unresolved_links` 仅作内部排除记录。`curation.basics` 的显式 `null` 表示该条原始基础提示经检查后不采用；`excluded_augment_names` 阻止这些名称从其他人工文案重新混入当前批次。新版本需重新核实。

本次清理只更新金克丝、皇子、死歌、梅尔、小丑的 5 条预览记录，总数仍为 284。写入前备份：`data/16.20/backups/guides-before-retired-cleanup-20261019.json`；差异批次：`data/16.20/retired-cleanup/batch.json`；写入后备份：`data/16.20/backups/guides-after-retired-cleanup-20261019.json`。已全批试导入后提交，并逐条比对写入后备份：5 条与差异文件一致，其余 279 条及全部来源元数据保持不变，正文、摘要、版本提示与参考链接均无上述三个名称和对应链接。原始快照与历史备份保留。

上述首批未覆盖希维尔、塔里克、奥莉安娜、蕾欧娜、瑟庄妮、璐璐、千珏、娜美、巴德、薇古丝。后续核查发现其中 9 位有公开装备清单，但没有玩法正文；修复方式见下节。巴德的核心海克斯仍待核对。网页隐藏的购买顺序、替换条件及完整教学未采集，不自动补写。

**全部保持 `reviewed: false`。** 关联图鉴 ID 正确不代表国服机制或许可已核实，正式 `gameplay_guides` 仍为空。页面许可尚未确认，不能通过改写或直接翻转审核标记将这批内容自动发布。当前线上预览开关关闭，正常小程序仍查询正式表。需先核对版本、模式、来源及装备再按正式流程录入；缺装备的基础记录继续保留预览。

## 装备清单提取修复

2026-10-20 复查 10 月 19 日原始快照：173 个英雄页中，172 个有公开核心装备。原采集器只有构建包含 `desc-text` 或 `tips-text` 时才保留构建，导致其中 50 位英雄的装备没有转换入库；不能将这些缺失解释为来源没有装备。

`cmd/guidefetch` 现在将有装备、无玩法正文的卡片提取到独立的 `equipment_builds`，保留该卡片的流派名称、海克斯和装备边界；仅有标题的隐藏构建、其他英雄链接及统计推荐不会变成装备清单。原有 `builds` 仍仅保存有正文的构建，避免装备清单被包装成完整玩法。

`16.20-equipment-curation.json` 逐条整理这 50 个流派名称，删除原名中的夸张伤害与无限循环承诺；显式核对“任务：沃格勒特的巫师帽”对应当前图鉴 `ARAM_Quest_WoogletsWitchcap`，显示采用图鉴名。`scripts/prepare-equipment-repair.cjs` 校验原文哈希及装备/海克斯 ID，仅给同英雄 `basic` 记录追加带流派名的装备候选；没有基础记录时建立仅含装备参考和版本提示的基础记录。不生成操作机制或推荐理由，不把装备套给其他独立玩法，也不改审核状态。

此补全批次已提交更新 41 条、增加 9 条；写入后预览总数为 293（124 条趣味玩法、169 条基础），覆盖 172 位英雄。173 条记录有装备（122 条趣味玩法、51 条基础），不代表每位英雄都有完整出装。正式表仍为 0。写入前后备份已逐条比对：50 条与批次一致，其他 243 条内容和全部已有来源元数据不变。旧正文、摘要、推荐及原装备保留；相同装备阶段重复执行不追加，有人工修改冲突时停止复核。

从项目根目录执行，先备份再生成。以下路径对应本次运行；再次执行备份需使用新文件名，不能覆盖历史备份：

```bash
go run ./cmd/importer backup-guides --patch 16.20 --out data/16.20/backups/guides-before-equipment-repair-20261020.json
go run ./cmd/guidefetch --out data/16.20/guide-sources
node scripts/prepare-equipment-repair.cjs 16.20 data/16.20/guide-sources docs/guide-expansion/16.20-equipment-curation.json data/16.20/backups/guides-before-equipment-repair-20261020.json data/16.20/equipment-repair
node scripts/test-equipment-repair.cjs
go run ./cmd/importer import-guide-batch --file data/16.20/equipment-repair/batch.json
go run ./cmd/importer import-guide-batch --file data/16.20/equipment-repair/batch.json --apply
go run ./cmd/importer backup-guides --patch 16.20 --out data/16.20/backups/guides-after-equipment-repair-20261020.json
```

采集命令本次仅重新解析已保存网页，没有请求新版本或覆盖原始 HTML。`test-equipment-repair.cjs` 依赖上述本地批次及写入前备份，检查流派装备、旧正文/过期海克斯清理结果保留、重复执行、人工修改冲突及英雄页可读取装备。原 `test-guide-expansion.cjs` 仍只校验 10 月 19 日的历史扩充批次，不代表最新数据库数量。

后续版本需重新整理 `equipment_builds` 的流派名与图鉴映射，并在基础玩法整理后处理装备参考。不要用 10 月 19 日旧备份重导覆盖本次补全，实际同版本修改以最新数据库备份为基础。

## 下次版本更新

通常无需改表或改页面，现有版本化图鉴与玩法表可继续使用。主要工作是核对新增海克斯、改动机制及受影响玩法；来源结构或游戏规则变动时才需要修改采集器。

1. 用户发起后确认真实国服补丁与资源版本，先查 `status`。预计更新日期不是补丁号，不能预设 `16.21.1` 已可用。
2. 新补丁使用新的图鉴快照运行 `sync`，保留旧版本。对比新旧 `augments.json` 的 ID、名称、描述，分别列出新增、变更与移除；当前没有自动生成这份差异或受影响玩法报告的命令，需要脚本/人工核对。
3. 在全新的 `data/<新补丁>/guide-sources/` 采集网页，检查来源日期是否真正跟进补丁；网页更新时间不能当作模式版本。当前采集器只覆盖公开英雄页，海克斯详情的交叉核对仍需另行进行。
4. 先备份新补丁的玩法；即使目前为空，也需要建立目标版本并记录。旧玩法留在旧版本，不能自动复制成新版已验证内容。
5. 对照原文制作新的整理文件与哈希清单。新增海克斯单独整理理由；变更海克斯重新检查所有引用它的玩法及版本提示。不要直接改旧整理文件的 `patch` 继续使用。
6. 转换器当前含本批固定的路径别名、旧样例 ID 合并及去重规则，**不是通用的新版本自动迁移器**。新版本需检查并调整这些规则；缺少原样例 ID 时会停止，不得为通过校验伪造记录。
7. 先试导入、看覆盖差异和本地预览，再用 `--apply` 提交。当前同补丁、同玩法 ID 更新；没有出现在本批的旧记录仍保留。来源删除、玩法下架需另列清单处理，不能全表删除重灌。来源名称包含英雄与补丁，避免更新新来源覆盖旧来源 URL。
8. 最后查询数量、关联、来源和分页，审核通过后按任务范围发布。静态 `publish` 不会把预览转成正式玩法。

本次验证命令：

```bash
go test ./cmd/guidefetch ./cmd/importer ./internal/importer ./internal/staticdata
node scripts/test-mini-core.cjs
node scripts/test-mini-relations.cjs
node scripts/test-guide-expansion.cjs
git diff --check
```

`test-guide-expansion.cjs` 验证本次本地批次和备份，覆盖引用、旧数据保留、重复路线、负向关联、超过 100 条的分页与每条详情；它不是生产数据库校验命令。实际入库后另导出备份与本地批次比对。
