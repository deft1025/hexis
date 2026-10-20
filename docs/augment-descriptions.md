# 海克斯描述修复

2026-10-19 按用户要求修复当前 `16.20` 的描述，并优化客户端跨页面读取。基础图鉴、技能、玩法、来源、版本状态和统计均未重新导入；仅在 `augment_patch_profiles` 更新了 51 条描述。原先 53 条带未解析提示，修复后为 6 条。

## 解析范围

`internal/staticdata/augment_descriptions.go` 读取相同版本的 KIWI 清单、模式资源和简中文本，支持：

- `mSpellCalculations` 中可识别的常量、命名常量、标量倍乘、属性加成、等级插值与简单分段成长。
- 已核对的属性：法术强度、护甲、攻击力、魔法抗性、生命值，以及总值/额外值区别；不认识的属性或计算类型保留未解析状态，不代入假想英雄面板。
- FNV-1a 名称哈希，解决原文里具名变量对应资源哈希键的问题；百分比显示去除浮点误差。
- 近战/远程分支。攻击距离优先使用明确的 `MeleeRangeIncrease` / `RangedRangeIncrease`，避免近似乘数造成 50.025 等显示值。
- 非技能专属任务通过关联 `ModesQuestData.Milestones` 获取首阶段条件，多阶段任务标注后续目标提高。技能相关任务的动态条件不直接替换为原始里程碑数字。
- 已知 Q/W/E 保留具体技能键；对局决定目标的 `IsAbilityAugment` 用“所选技能”表达，不猜固定技能。

本次示例：万用瞄准镜为近战 75 / 远程 50 攻击距离；旋转至胜为 30% 额外伤害、终极技能 50%；巨像的勇气保留等级成长与最大生命值公式；升级焚天为 5 秒、9% 已损失生命值、500 金币。

`fetch` 的后续新版本采集也会经过同一解析器。新版本仍须核对实际原文与结果，不代表未知新计算自动可用。

## 当前未解析

以下六条仍保留未确认提示，未靠删除提示或填固定数值消除：

| ID | 未确认内容 |
| --- | --- |
| `ARAM_OkBoomerang` | 模板 `Cooldown` 与原始 `BaseCooldown` 的运行时关系 |
| `Twinfire` | 额外飞弹数量所用属性/换算 |
| `ARAM_Juiced` | 法力相关的特殊计算类型 |
| `ARAM_Multishot` | 技能相关任务所需命中次数 |
| `ARAM_SpecializedRecursion` | 技能相关任务所需命中次数 |
| `SpecializedEmpowerment` | 技能相关任务所需命中次数 |

“所选技能”是动态机制的通用表达，不等于已经获取每位英雄对应的专属技能清单。原始摘要中本来没有写出的伤害、治疗等数值不会自动扩写，不能把本次结果称为全机制详解。

## 修复命令

仅在用户要求修复已存在版本时使用，不因阅读此文档自行执行。数据库配置来自现有环境，不在命令或文档填写凭据。

```bash
# 固定快照重算、核对数据库原文、输出差异；不修改数据库
go run ./cmd/importer repair-augment-descriptions \
  --dir data/16.20 --out data/16.20/backups/augment-descriptions-plan-NEW.json

# 逐条检查对照，再用新的备份文件名应用
go run ./cmd/importer repair-augment-descriptions \
  --dir data/16.20 --out data/16.20/backups/augment-descriptions-applied-NEW.json --apply
```

工具校验 manifest 的固定版本、KIWI 模式、简中、三份原始资源的 SHA-256 和来源版本，再核对数据库与快照的数量、ID、名称及描述。数据库描述必须与原始快照或本次重算结果相同；碰到人工改写则停止要求复核。报告含每条变更前后的描述、版本和未解析列表，独占创建、权限 `0600`，不覆盖旧备份。

`--apply` 写入前先保存报告，再在同一事务中更新该版本的描述。仅工具成功退出才表示提交成功；报告的 `apply_requested` 只记录调用意图，不能独立证明事务已提交。失败则回滚。重复执行会跳过已经相同的描述。`augments` 公共身份表、来源和其他补丁均不修改，API 使用版本资料表。

本次实际备份为 `data/16.20/backups/augment-descriptions-applied-20261019.json`。`data/16.20/raw/`、原始 `augments.json` 和原始 manifest 未改，仍用于追溯原导入；修复结果在数据库和差异报告里。不要重导原始旧文件覆盖修复。

## 验证

```bash
go test ./...
node scripts/test-mini-core.cjs
node scripts/test-mini-relations.cjs
node scripts/test-mini-cache.cjs
node scripts/test-guide-expansion.cjs
```

客户端缓存实现和使用方式见 [小程序说明](../miniprogram/README.md)。缓存不改变资料从本项目 API/数据库读取的边界，图片文件仍按已有外部 URL 加载。
