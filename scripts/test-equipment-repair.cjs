const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const { prepare } = require('./prepare-equipment-repair.cjs');
const { create } = require('../miniprogram/lib/core');

const patch = '16.20';
const dir = 'data/16.20/equipment-repair';
const read = file => JSON.parse(fs.readFileSync(file, 'utf8'));
const backupFile = 'data/16.20/backups/guides-before-equipment-repair-20261020.json';
const backup = read(backupFile);
const original = new Map(backup.rows.filter(r => r.table === 'gameplay_guide_previews').map(r => [r.key, r.content]));
const guides = new Map(original);
const batch = read(path.join(dir, 'batch.json'));
const catalog = kind => read(`data/${patch}/${kind}.json`);
const itemIDs = new Set(catalog('items').map(i => i.id));
const inserted = [], updated = [];
for (const file of batch.files) for (const guide of read(path.join(dir, file.file))) {
  assert.equal(guide.category, 'basic');
  assert.equal(guide.reviewed, false);
  assert(guide.stages.length);
  for (const stage of guide.stages) {
    assert(stage.item_ids.every(id => itemIDs.has(id)));
    assert.match(stage.label, /装备候选/);
    assert.match(stage.reason, /非固定购买顺序/);
  }
  const old = original.get(guide.id);
  if (old) {
    assert.deepEqual({ ...guide, stages: old.stages }, old, 'existing content outside equipment must remain unchanged');
    assert.deepEqual(guide.stages.slice(0, old.stages.length), old.stages);
    updated.push(guide.id);
  } else {
    assert(guide.sections.every(s => s.title === '版本提示'), 'no gameplay is invented for equipment-only records');
    assert.deepEqual(guide.recommendations, []);
    inserted.push(guide.id);
  }
  guides.set(guide.id, guide);
}
assert.equal(updated.length, 41);
assert.equal(inserted.length, 9);
assert.equal(guides.size, 293);
assert.deepEqual(guides.get('haidou-jarvaniv-basic').sections, original.get('haidou-jarvaniv-basic').sections, 'retired augment cleanup is preserved');
for (const guide of guides.values()) assert(!/活力再生|最终都市列车|卡皮巴拉空投/.test(JSON.stringify(guide)));
const tmp = fs.mkdtempSync(path.join(os.tmpdir(), 'hexis-equipment-test-'));
try {
  const merged = { ...backup, rows: backup.rows.map(row => ({ ...row, content: guides.get(row.key) || row.content })) };
  for (const id of inserted) merged.rows.push({ table: 'gameplay_guide_previews', key: id, content: guides.get(id), source: 'test' });
  const mergedFile = path.join(tmp, 'backup.json');
  fs.writeFileSync(mergedFile, JSON.stringify(merged));
  const options = { patch, sourcesDir: 'data/16.20/guide-sources', curationFile: 'docs/guide-expansion/16.20-equipment-curation.json', backupFile: mergedFile, out: path.join(tmp, 'rerun') };
  const rerun = prepare(options);
  assert.equal(rerun.unchanged.length, 50, 'rerunning cannot duplicate equipment stages');
  guides.get('haidou-vayne-basic').stages[0].reason = 'newer manual edit';
  fs.writeFileSync(mergedFile, JSON.stringify(merged));
  assert.throws(() => prepare({ ...options, out: path.join(tmp, 'conflict') }), /Existing equipment changed/, 'manual edits must not be overwritten');
} finally { fs.rmSync(tmp, { recursive: true, force: true }); }

const data = { champions: catalog('champions').map(r => ({ ...r, name_zh: r.name })), items: catalog('items').map(r => ({ ...r, name_zh: r.name })), augments: catalog('augments').map(r => ({ ...r, name_zh: r.name })), guides: [...guides.values()] };
data.guides.find(g => g.id === 'haidou-vayne-basic').stages[0].reason = '海克斯搭配：双发快射。装备按对局取舍，非固定购买顺序。';
const app = create({ previewGuides: true, read: () => null, write: () => {}, render: () => {}, async request(url) {
  if (url.includes('/latest')) return { version: patch };
  const u = new URL(url, 'http://local'), kind = u.pathname.split('/').pop(), offset = Number(u.searchParams.get('offset'));
  return { patch, items: data[kind].slice(offset, offset + 100) };
} });
(async () => {
  await app.load();
  for (const [hero, route, count] of [['Vayne', '双发快射暴击特效流', 6], ['Lux', '法强爆发流', 7], ['Sivir', '暴击特效流', 10]]) {
    app.enter('detail', { kind: 'champions', id: hero });
    const view = app.view();
    const stage = view.commonBasics.flatMap(g => g.stages).find(s => s.label.startsWith(route));
    assert(stage, hero + ': missing equipment on hero page');
    assert.equal(stage.items.length, count);
    assert(stage.items.every(i => i.name));
    if (hero === 'Sivir') assert.equal(view.heroGuideHeading, '装备参考');
  }
  console.log('PASS: 50 equipment routes, preserved existing text/cleanup/equipment, current item references, repeat import protection and hero-page display');
})().catch(error => { console.error(error); process.exitCode = 1; });
