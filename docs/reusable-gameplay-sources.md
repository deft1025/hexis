# 可复用玩法数据源调查

调查日期：2026-10-14。范围为已有资料，不生成原创攻略，不导入胜率或选取率。本次只核查与读取公开样本，没有写入生产数据库。

## 结论

目前找到明确允许按条件复用的机制资料，以及开放源码项目中的标签和特性文件；尚未找到同时满足完整攻略、明确数据许可、当前国服版本的成品数据集。不能把“找到公开 JSON”报告为“获得整套攻略库”。

| 来源 | 实际内容 | 许可与来源 | 适用结论 |
| --- | --- | --- | --- |
| League of Legends Wiki | 海克斯效果、机制 Notes、技能交互、历史变更；部分数据有 Lua 源文件 | 页面明确标注 CC BY-SA 3.0；复用须署名、链接许可、标记修改或翻译，并满足相同方式共享要求；图片等素材须单独核查 | 优先候选，可复用机制说明；不能填充不存在的完整出装路线 |
| kiwiiiii-eth/aram-mayhem-analyzer | `data/augments.json` 223 个海克斯及标签，`data/champions.json` 173 个英雄及能力特征 | 仓库 MIT；README 说明英文描述来自 Wiki，Wiki 派生内容仍须按 CC BY-SA 处理；名称、图标来自 CommunityDragon | 标签/特征可以继续核查，不能直接视为实测联动或攻略正文 |
| tsuiho6/aram-mayhem-guide | `data/augment-build-guides.json` 包含 15 个英雄的 16 条联动出装；另有阶段路线与情境规则文件 | 根目录未发现 LICENSE，README 未授予明确复用许可；样本注明来源为 ARAMMayhem 编辑攻略 | 内容贴合，但复用条件未明确，暂不导入 |
| xzq-xu/open-hex-assistant | 装备、英雄专属海克斯的统计推荐与本地规则程序 | 代码 MIT；README 指向 `data.v2.iesdev.com` 的统计、`hextech.dtodo.cn` 的元数据，未验证上游再分发许可 | 不是独立开放攻略库，不以代码 MIT 代替上游数据许可 |
| joechiboo/aram-mayhem-handbook | 211 个海克斯、英雄前 12 个海克斯和选择原则 | README 明确列出 ARAMKit、Mobalytics 等来源，声明统计版权归原来源；未确认数据开放许可 | 偏统计，当前范围不接入 |

一个名为 `NEMA-Anonymous/ARAM-MAYHEM-COMPANION` 的 MIT 项目核查时仅有 LICENSE，没有实际攻略数据，不作为可用数据源。

## 下载与来源地址

### Wiki 机制资料

- 模式说明：https://wiki.leagueoflegends.com/en-us/ARAM:_Mayhem
- 海克斯效果与 Notes：https://wiki.leagueoflegends.com/en-us/ARAM:_Mayhem/Augments
- 结构化模块：https://wiki.leagueoflegends.com/en-us/Module:MayhemAugmentData/data?action=raw
- 页面标注的许可：https://creativecommons.org/licenses/by-sa/3.0/
- 附加许可说明：https://meta.weirdgloop.org/w/Licensing

本次核实了页面许可和海克斯页面包含英雄技能交互说明。结构化模块地址由开源重建脚本引用，本次没有解析该模块，不能保证 Notes 均包含在模块中。网页文本和 Lua 不是完整中文攻略 JSON。

### 标签与英雄能力文件

- 仓库：https://github.com/kiwiiiii-eth/aram-mayhem-analyzer
- 海克斯文件：https://raw.githubusercontent.com/kiwiiiii-eth/aram-mayhem-analyzer/main/data/augments.json
- 英雄文件：https://raw.githubusercontent.com/kiwiiiii-eth/aram-mayhem-analyzer/main/data/champions.json
- MIT 许可：https://github.com/kiwiiiii-eth/aram-mayhem-analyzer/blob/main/LICENSE
- 数据来源说明：https://github.com/kiwiiiii-eth/aram-mayhem-analyzer/blob/main/README.md

样本结果：223 个海克斯 `key` 全部精确匹配本地 `16.20` 海克斯 ID；173 个英雄 `id` 全部精确匹配英雄 ID。这只证明 ID 一致，不证明数据版本、机制或推荐有效。

重建脚本使用 CommunityDragon `latest` 与最新 Data Dragon，数据没有显式固定版本元数据。英雄能力由技能描述关键词识别并配人工覆盖；例如样本将亚托克斯的 R 标为含硬控，将阿狸的治疗能力标为 false，因此不能把这些特征作为无误的英雄事实直接展示。`aram_balance` 来自 Wiki 的 `aram` 字段，也不能直接当成 KIWI 专属平衡参数。繁中名称需用我们已有简中图鉴映射。

### 现成联动出装候选

- 仓库：https://github.com/tsuiho6/aram-mayhem-guide
- 文件：https://raw.githubusercontent.com/tsuiho6/aram-mayhem-guide/main/data/augment-build-guides.json
- 来源说明：https://github.com/tsuiho6/aram-mayhem-guide/blob/main/README.md

样本自述版本 `26.19`，整理日期 `2026-09-26`，来源 `ARAMMayhem editorial combo pages`。15 个英雄、16 条联动出装，存在 `requiredAugments`、装备 ID、说明和原始来源 URL。许可和原始攻略再分发条件均未确认；不能通过改写文本、去掉评级或转换格式来解决许可缺口。内容也需复核，样本存在将破败王者之刃作用描述为维持护盾等可疑说明。

## 接入边界

优先采用已有资料，不运行根据标签自动生成玩法的方案。先核查 Wiki 中实际存在的技能交互 Notes，保存原页面、修订版本、许可与翻译记录，匹配现有英雄和海克斯 ID；核对国服当前 KIWI 机制后才能作为联动资料发布。

当前 `gameplay_guides` 导入器要求玩法说明和装备阶段都非空。纯机制 Notes 通常没有完整出装信息，不应为了通过校验自动补造装备路线。若最终选用此类源，需要调整内容模型或另设联动说明表，再让海克斯联动栏目读取它。基础思路和趣味玩法仍需找到已整理且允许复用的攻略，不能声称一个机制源已覆盖全部栏目。

如获现成攻略数据许可，再按原始结构映射到 `gameplay_guides`，保留来源、许可、版本和审核记录。调查阶段不写入未审核内容，不使用国际服统计替代国服统计。
