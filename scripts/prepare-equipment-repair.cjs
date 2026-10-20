const fs = require('node:fs');
const path = require('node:path');
const { createHash } = require('node:crypto');

function prepare({ patch, sourcesDir, curationFile, backupFile, out, catalogDir = path.join('data', patch) }) {
  const read = file => JSON.parse(fs.readFileSync(file, 'utf8'));
  const curation = read(curationFile), backup = read(backupFile);
  if (curation.patch !== patch || backup.patch !== patch) throw new Error('Patch mismatch');
  const hashes = read(path.join(path.dirname(curationFile), curation.source_hashes_file));
  const champions = new Map(read(path.join(catalogDir, 'champions.json')).map(c => [c.id.toLowerCase(), c]));
  const items = new Map(read(path.join(catalogDir, 'items.json')).map(i => [i.name, i.id]));
  const augments = read(path.join(catalogDir, 'augments.json'));
  const byName = new Map(augments.map(a => [a.name, a]));
  for (const [alias, id] of Object.entries(curation.augment_aliases || {})) {
    const augment = augments.find(a => a.id === id);
    if (!augment) throw new Error('Unknown augment alias: ' + alias);
    byName.set(alias, augment);
  }
  const report = { patch, inserts: [], updates: [], unchanged: [], files: [] };
  const seen = new Set();
  fs.mkdirSync(out, { recursive: true });
  for (const p of read(path.join(sourcesDir, 'pages.json'))) {
    if (!p.equipment_builds?.length) continue;
    const slug = path.basename(new URL(p.url).pathname, '.html');
    const routes = curation.routes[slug];
    if (!routes || routes.length !== p.equipment_builds.length || routes.some(s => !s.trim())) throw new Error('Unreviewed equipment routes: ' + slug);
    seen.add(slug);
    const rawHash = createHash('sha256').update(fs.readFileSync(path.join(sourcesDir, p.file))).digest('hex');
    if (hashes[p.url] !== p.sha256 || rawHash !== p.sha256) throw new Error('Source changed: ' + slug);
    const champion = champions.get(slug);
    if (!champion) throw new Error('Unknown champion: ' + slug);
    const id = 'haidou-' + slug + '-basic';
    const old = backup.rows.find(r => r.table === 'gameplay_guide_previews' && r.key === id);
    if (old && (old.content.category !== 'basic' || old.content.reviewed !== false || JSON.stringify(old.content.champion_ids) !== JSON.stringify([champion.id]))) throw new Error('Unexpected existing record: ' + id);
    const stages = p.equipment_builds.map((b, index) => {
      const itemIDs = [...new Set(b.items.map(name => {
        if (!items.has(name)) throw new Error('Unknown item: ' + slug + ': ' + name);
        return items.get(name);
      }))];
      const names = b.augments.map(a => {
        if (!byName.has(a.name)) throw new Error('Unknown augment: ' + slug + ': ' + a.name);
        return byName.get(a.name).name;
      });
      return { label: routes[index] + ' · 装备候选', item_ids: itemIDs, reason: `${names.length ? '海克斯搭配：' + names.join('、') + '。' : ''}装备按对局取舍，非固定购买顺序。` };
    });
    const reference = { name: '海斗小助手：' + champion.name, url: p.url, note: '公开装备清单；原文哈希 ' + p.sha256 };
    const guide = structuredClone(old?.content || {
      id, title: champion.name + '：装备参考', category: 'basic', summary: routes.join('、') + '的装备候选。', champion_ids: [champion.id], tags: [],
      core_augment_ids: [], alternative_augment_ids: [], recommendations: [], stages: [],
      sections: [{ title: '版本提示', body: `当前装备已匹配 ${patch} 图鉴；搭配条件与效果仍需核对当前国服版本。` }], references: [reference], reviewed: false
    });
    for (const stage of stages) {
      const existing = guide.stages.find(s => s.label === stage.label);
      if (existing && JSON.stringify(existing) !== JSON.stringify(stage)) throw new Error('Existing equipment changed; review before replacing: ' + id);
      if (!existing) guide.stages.push(stage);
    }
    if (!guide.references.some(r => r.url === p.url)) guide.references.push(reference);
    if (old && JSON.stringify(guide) === JSON.stringify(old.content)) { report.unchanged.push(id); continue; }
    (old ? report.updates : report.inserts).push(id);
    const filename = slug + '.json';
    fs.writeFileSync(path.join(out, filename), JSON.stringify([guide], null, 2) + '\n');
    report.files.push({ file: filename, source: old?.source || `海斗小助手公开资料：${champion.id} ${patch}`,
      source_url: old?.source_url?.Valid ? old.source_url.String : p.url,
      source_license: old?.source_license?.Valid ? old.source_license.String : '未确认开放转载许可',
      source_attribution: old?.source_attribution?.Valid ? old.source_attribution.String : `海斗小助手；公开装备清单整理；页面日期 ${p.source_modified}；未核对国服当前版本` });
  }
  for (const slug of Object.keys(curation.routes)) if (!seen.has(slug)) throw new Error('Reviewed equipment route missing: ' + slug);
  fs.writeFileSync(path.join(out, 'batch.json'), JSON.stringify({ patch, files: report.files }, null, 2) + '\n');
  fs.writeFileSync(path.join(out, 'coverage.json'), JSON.stringify(report, null, 2) + '\n');
  console.log(`Equipment repair: ${report.updates.length} updates, ${report.inserts.length} inserts, ${report.unchanged.length} unchanged; database unchanged`);
  return report;
}

if (require.main === module) {
  const [patch, sourcesDir, curationFile, backupFile, out] = process.argv.slice(2);
  if (!out) throw new Error('Usage: node scripts/prepare-equipment-repair.cjs PATCH SOURCES_DIR CURATION_JSON BACKUP_JSON OUT_DIR');
  prepare({ patch, sourcesDir, curationFile, backupFile, out });
}
module.exports = { prepare };
