const fs = require('node:fs');
const path = require('node:path');
const { createHash } = require('node:crypto');

function prepare({ patch, sourcesDir, curationFile, backupFile, out }) {
  const read = file => JSON.parse(fs.readFileSync(file, 'utf8'));
  const curation = read(curationFile);
  const backup = read(backupFile);
  if (curation.patch !== patch || backup.patch !== patch) throw new Error('Curation and backup must match the target patch');
  if (!curation.source_hashes_file) throw new Error('Curation must pin the reviewed source hashes');
  const hashes = read(path.join(path.dirname(curationFile), curation.source_hashes_file));
  const pages = read(path.join(sourcesDir, 'pages.json'));
  const catalog = kind => read(path.join('data', patch, kind + '.json'));
  const champions = new Map(catalog('champions').map(c => [c.id.toLowerCase(), c]));
  champions.set('huijinqumoren', champions.get('locke'));
  const augments = new Map(catalog('augments').map(a => [a.name, a.id]));
  const augmentIDs = new Set(augments.values());
  augments.set('任务：钢化你心', 'ARAM_Quest_SteelYourHeart');
  const items = new Map(catalog('items').map(i => [i.name, i.id]));
  const slugOf = p => path.basename(new URL(p.url).pathname, '.html');
  const existing = new Map(backup.rows.filter(r => r.table === 'gameplay_guide_previews').map(r => [r.key, r.content]));
  const report = { patch, source_pages: pages.length, builds_available: 0, mechanisms_available: 0, build_records: 0, basic_records: 0, inserts: 0, updates: [], retained_existing_ids: [], held_builds: [], missing_heroes: [], unresolved_links: [], files: [] };
  const produced = new Set();
  fs.mkdirSync(out, { recursive: true });
  for (const p of pages) {
    const slug = slugOf(p);
    if (hashes[p.url] !== p.sha256) throw new Error('Source changed; review the curation before preparing: ' + p.url);
    const rawHash = createHash('sha256').update(fs.readFileSync(path.join(sourcesDir, p.file))).digest('hex');
    if (rawHash !== p.sha256) throw new Error('Source snapshot hash mismatch: ' + p.url);
    const champion = champions.get(slug);
    if (!champion) throw new Error('Unmapped hero: ' + slug);
    const mechanics = p.blocks.filter(b => b.kind === 'box-special' || b.kind === 'box-trap');
    const edits = curation.mechanics[slug] || [];
    if (mechanics.length !== edits.length) throw new Error('Unreviewed mechanics: ' + slug);
    if ((p.basics || []).length && !curation.basics[slug] && curation.basics[slug] !== null) throw new Error('Unreviewed basic text: ' + slug);
    report.builds_available += (p.builds || []).length;
    report.mechanisms_available += mechanics.length;
    const references = [{ name: '海斗小助手：' + champion.name, url: p.url, note: '公开页面；来源日期 ' + p.source_modified + '；原文哈希 ' + p.sha256 }];
    const versionNote = { title: '版本提示', body: '当前内容关联 ' + patch + ' 图鉴，技能、海克斯和装备引用已匹配。来源页面未固定国服补丁，具体触发、数值及循环条件仍需在当前国服核对。' };
    const records = [];
    if ((p.builds || []).length) {
      if (p.builds.length !== 1) throw new Error('Multiple builds require separate editorial IDs: ' + slug);
      const b = p.builds[0];
      if (curation.skip_builds[slug]) {
        report.held_builds.push({ hero: champion.id, source_name: b.name, reason: curation.skip_builds[slug] });
      } else {
        const edit = curation.builds[slug];
        if (!edit || edit.length !== 4 || edit.some(x => !x.trim())) throw new Error('Unreviewed build: ' + slug);
        const core = b.augments.map(a => {
          const id = augments.get(a.name);
          if (!id) throw new Error('Unmapped core augment: ' + slug + ': ' + a.name);
          return id;
        });
        core.push(...(curation.extra_core[slug] || []));
        if (core.some(id => !augmentIDs.has(id))) throw new Error('Missing current-patch core augment: ' + slug);
        const itemIDs = [...new Set(b.items.map(name => {
          const id = items.get(name);
          if (!id) throw new Error('Unmapped item: ' + slug + ': ' + name);
          return id;
        }))];
        const reasons = curation.build_reasons?.[slug] || [];
        if (core.length > 1 && reasons.length !== core.length) throw new Error('Each core augment needs its own reason: ' + slug);
        records.push({ id: 'haidou-' + slug + '-build', title: champion.name + '：' + edit[0], category: 'fun', playstyle_name: edit[0], summary: edit[1], champion_ids: [champion.id], tags: [], core_augment_ids: [...new Set(core)], alternative_augment_ids: [], recommendations: [...new Set(core)].map((id, i) => ({ id, priority: 'core', reason: reasons[i] || edit[2] })), stages: itemIDs.length ? [{ label: '装备候选（非购买顺序）', item_ids: itemIDs, reason: edit[3] }] : [], basic_sections: [{ title: '玩法思路', body: edit[2] }], sections: [versionNote], references: [...references, ...b.augments.map(a => ({ name: a.name, url: a.url }))], reviewed: false });
        report.build_records++;
      }
    }
    const sections = [];
    const recommendations = new Map();
    const onlyVersion = [];
    if (curation.basics[slug]) sections.push({ title: '操作要点', body: curation.basics[slug] });
    mechanics.forEach((m, index) => {
      const label = (m.links[0]?.name || '机制提示').replace(/^\[|\]$/g, '');
      const id = augments.get(label);
      if (!id) {
        report.unresolved_links.push({ hero: champion.id, name: label, url: m.links[0]?.url });
        return;
      }
      const versionOnly = (curation.version_only_mechanics[slug] || []).includes(index);
      const section = { title: (m.kind === 'box-trap' ? '避坑：' : '联动：') + label, body: edits[index] };
      if (versionOnly) onlyVersion.push(section.title + '。' + section.body);
      else sections.push(section);
      // Negative interactions remain text and must never become hero recommendations.
      if (!versionOnly && m.kind === 'box-special') {
        const previous = recommendations.get(id);
        recommendations.set(id, { id, priority: 'alternative', reason: previous ? previous.reason + ' ' + edits[index] : edits[index] });
      }
    });
    mechanics.forEach((m, index) => {
      const label = (m.links[0]?.name || '').replace(/^\[|\]$/g, '');
      const id = augments.get(label);
      if (m.kind === 'box-trap' && recommendations.has(id)) {
        const r = recommendations.get(id);
        recommendations.set(id, { ...r, priority: 'situational', reason: r.reason + ' ' + edits[index] });
      }
    });
    if (sections.length || onlyVersion.length) {
      const oldKey = slug === 'garen' ? 'sample-garen-basic' : slug === 'tahmkench' ? 'sample-tahm-basic' : null;
      const old = oldKey ? existing.get(oldKey) : null;
      if (oldKey && !old) throw new Error('Expected existing guide not found: ' + oldKey);
      const mergedRecommendations = new Map((old?.recommendations || []).map(r => [r.id, r]));
      for (const [id, r] of recommendations) if (!mergedRecommendations.has(id)) mergedRecommendations.set(id, r);
      const oldSections = (old?.sections || []).filter(s => s.title !== '版本提示');
      const allSections = [...oldSections, ...sections];
      if (onlyVersion.length) allSections.push({ title: '版本提示', body: onlyVersion.join('\n\n') });
      if (old) allSections.push(...old.sections.filter(s => s.title === '版本提示'));
      allSections.push(versionNote);
      const firstCurrentMechanic = mechanics.findIndex(m => augments.has((m.links[0]?.name || '').replace(/^\[|\]$/g, '')));
      const summary = old?.summary || curation.basics[slug] || edits[firstCurrentMechanic];
      records.push({ ...(old || {}), id: oldKey || 'haidou-' + slug + '-basic', title: old?.title || champion.name + '：技能与海克斯选择', category: 'basic', summary, champion_ids: [champion.id], tags: old?.tags || [], core_augment_ids: old?.core_augment_ids || [], alternative_augment_ids: [...new Set([...(old?.alternative_augment_ids || []), ...recommendations.keys()])], recommendations: [...mergedRecommendations.values()], stages: old?.stages || [], sections: allSections, references: [...(old?.references || []), ...references, ...mechanics.filter(m => augments.has((m.links[0]?.name || '').replace(/^\[|\]$/g, ''))).flatMap(m => m.links.map(a => ({ name: a.name, url: a.url })))].filter((r, i, all) => all.findIndex(x => x.url === r.url) === i), reviewed: false });
      report.basic_records++;
    }
    if (!records.length && ![...existing.values()].some(g => g.champion_ids.includes(champion.id))) report.missing_heroes.push(champion.id);
    if (!records.length) continue;
    for (const g of records) {
      if ((curation.excluded_augment_names || []).some(name => JSON.stringify(g).includes(name))) throw new Error('Unavailable augment remains in curated guide: ' + g.id);
      if (produced.has(g.id)) throw new Error('Duplicate guide ID: ' + g.id);
      produced.add(g.id);
      if (existing.has(g.id)) report.updates.push(g.id); else report.inserts++;
    }
    const file = path.join(out, slug + '.json');
    fs.writeFileSync(file, JSON.stringify(records, null, 2) + '\n');
    report.files.push({ file: path.basename(file), source: '海斗小助手公开资料：' + champion.id + ' ' + patch, source_url: p.url, source_license: '未确认开放转载许可', source_attribution: '海斗小助手；公开内容整理改写；页面日期 ' + p.source_modified + '；未核对国服当前版本' });
  }
  report.retained_existing_ids = [...existing.keys()].filter(id => !produced.has(id));
  fs.writeFileSync(path.join(out, 'batch.json'), JSON.stringify({ patch, files: report.files }, null, 2) + '\n');
  fs.writeFileSync(path.join(out, 'coverage.json'), JSON.stringify(report, null, 2) + '\n');
  console.log(JSON.stringify({ ...report, files: report.files.length }, null, 2));
  return report;
}

if (require.main === module) {
  const [patch, sourcesDir, curationFile, backupFile, out] = process.argv.slice(2);
  if (!out) throw new Error('Usage: node scripts/prepare-guide-expansion.cjs PATCH SOURCES_DIR CURATION_JSON BACKUP_JSON OUT_DIR');
  prepare({ patch, sourcesDir, curationFile, backupFile, out });
}
module.exports = { prepare };
