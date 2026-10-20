# 公开玩法样例

2026-10-19 已在原有十条样例基础上整理扩充批次；本文件和六份 JSON 仅描述原始十条样例，不能再作为预览表的全量清单。新批次、旧数据处理、备份和后续版本更新见 [玩法扩充与版本更新](../guide-expansion/README.md)。盖伦和塔姆基础记录已有补充，不要不加核对地用这里的旧 JSON 覆盖数据库最新内容。

当前共十条，涉及六位英雄：塔姆三条流派样例与一条通用基础，盖伦一条流派样例与一条通用基础，另补卡特琳娜、韦鲁斯、卡尔玛、索拉卡各一条。由海斗小助手公开网页的玩法提示、机制提示和装备候选整理，重写描述并保留来源；未获取小程序内的完整攻略、购买顺序或付费内容。没有导入胜率、选取率或排名。

正文按玩家阅读习惯描述玩法、搭配和操作，采集与整理过程保留在本文档和来源信息中。未实测的机制集中写在各条“版本提示”，`reviewed` 仍为 `false`，不因文案润色改变审核状态。缺失的装备或顺序直接标注待补充。

塔姆的三个流派为“虚幻武器特效流”“霸王血铠转法强”“急救用具恢复流”。后两条是把公开的特殊机制提示分别整理为联动样例，并非原站提供的完整流派攻略。急救用具没有配套装备资料，因此 `stages` 为空；不借用其他路线的清单。公开页面另列“连拨击锤 AP 流”“坦克流”，但没有详细内容，本次没有补写。

每条流派自己的操作思路保存在 `basic_sections`，海克斯和装备仍保存在本条的 `recommendations`、`stages`。塔姆通用基础只保留被动、控制和体型提示，原先混在其中的两组联动已移入各自流派。

| 英雄 | 来源 URL | 页面更新时间 | 本次整理日期 |
| --- | --- | --- | --- |
| 塔姆 | https://lolhaidou.cn/hero/tahmkench.html | 2026-09-24 | 2026-10-14 |
| 盖伦 | https://lolhaidou.cn/hero/garen.html | 2026-09-24 | 2026-10-14 |
| 卡特琳娜 | https://lolhaidou.cn/hero/katarina.html | 原网页为准 | 2026-10-15 |
| 韦鲁斯 | https://lolhaidou.cn/hero/varus.html | 原网页为准 | 2026-10-15 |
| 卡尔玛 | https://lolhaidou.cn/hero/karma.html | 原网页为准 | 2026-10-15 |
| 索拉卡 | https://lolhaidou.cn/hero/soraka.html | 原网页为准 | 2026-10-15 |

页面未确认开放转载许可。改写不等于获得授权，也不等于机制已核实。这十条仅作为内容结构和显示效果的样例，`reviewed: false`；版本 `16.20` 表示关联的图鉴快照，不能据此认定来源的玩法已经过 16.20 或国服验证。

海克斯详情的典型实例覆盖：虚幻武器（塔姆、卡特琳娜、韦鲁斯），急救用具（塔姆、卡尔玛、索拉卡），旋转至胜（盖伦、卡特琳娜），以及其余已引用海克斯，共十种。英雄按核心/备选关系和名称展示，不代表强度排名。每个英雄可关联多条玩法，分别保留理由；默认显示三位，可展开。没有关联资料的海克斯保持待补充，列表提供“有联动”筛选，不根据图鉴标签生成假推荐。

卡尔玛和索拉卡的急救用具候选关系由公开页面 https://lolhaidou.cn/augment/firstaidkit.html 支持；英雄页提供各自的护盾/治疗路线。补充来源保存在玩法 JSON 的 `references` 字段，详情可查看或复制来源链接。虚幻武器的公开适配页 https://lolhaidou.cn/augment/etherealweapon.html 也用于交叉核对；不导入原站排名、胜率或固定伤害倍数。

原始页面本地快照保存在忽略目录 `data/16.20/guide-previews/raw/`。SHA-256：

- `tahmkench.html`：`0f3a66f5b6844cf66fac977d2899ae3fa935214c3847e9cca23fcc67a7076f02`
- `garen.html`：`10cd8a2f157fd6e263d510cffcd5c098b8b690d65729d20bf572ac002a10c09d`

2026-10-15 补充的四个英雄页与两个海克斯页快照同样保存到 `raw/`，URL、SHA-256 与保存时间记录在 `data/16.20/guide-previews/synergy-sources.json`。

执行 `go run ./cmd/importer migrate` 创建独立样例表 `gameplay_guide_previews`，然后分别导入，来源名称按英雄区分，避免覆盖来源链接：

```bash
go run ./cmd/importer import --patch 16.20 --kind guide-previews --file docs/guide-previews/tahmkench.json --source "海斗小助手公开样例：塔姆" --source-url "https://lolhaidou.cn/hero/tahmkench.html" --source-license "未确认开放转载许可" --source-attribution "海斗小助手；公开内容整理改写；页面更新 2026-09-24；未核对国服当前版本"
go run ./cmd/importer import --patch 16.20 --kind guide-previews --file docs/guide-previews/garen.json --source "海斗小助手公开样例：盖伦" --source-url "https://lolhaidou.cn/hero/garen.html" --source-license "未确认开放转载许可" --source-attribution "海斗小助手；公开内容整理改写；页面更新 2026-09-24；未核对国服当前版本"
```

新增 `katarina.json`、`varus.json`、`karma.json`、`soraka.json` 也按同样命令逐个导入，使用各英雄单独的来源名称与英雄页 URL，避免覆盖来源链接。每个文件均为一条预览记录。

预览服务需显式设置 `GUIDE_PREVIEW_ENABLED=1`。浏览器访问 `/mini/?guides=preview#detail?kind=champions&id=TahmKench`，可以选择流派并切换三个栏目；微信开发版默认读取已审核表，本地查看样例时需手动将 `app.js` 的 `globalData.previewGuides` 设置为 `true`，并连接已启用预览的开发后端，上传前恢复为 `false`。体验版与正式版始终读取已审核表。正常 `/api/v1/guides` 始终读取 `gameplay_guides`，只有启用预览后的 `?preview=1` 才读取样例表。预览开关默认关闭，不用于部署正式服务。

核对授权、当前版本和实际联动后，再按正常 `--kind guides` 流程录入审核内容。装备候选不是购买顺序，不补造来源没有提供的详细联动理由。
