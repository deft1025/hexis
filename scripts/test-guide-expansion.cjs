const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const { create } = require('../miniprogram/lib/core');

const patch = process.argv[2] || '16.20';
const dir = process.argv[3] || path.join('data', patch, 'guide-expansion');
const backupFile = process.argv[4] || path.join('data', patch, 'backups/guides-before-expansion-20261019.json');
const read = file => JSON.parse(fs.readFileSync(file, 'utf8'));
const batch = read(path.join(dir, 'batch.json'));
const coverage = read(path.join(dir, 'coverage.json'));
const backup = read(backupFile);
const guides = new Map(backup.rows.filter(r => r.table === 'gameplay_guide_previews').map(r => [r.key, r.content]));
const imported = batch.files.flatMap(f => read(path.join(dir, f.file)));
assert.equal(new Set(imported.map(g => g.id)).size, imported.length);
for (const g of imported) guides.set(g.id, g);
const catalog = kind => read(path.join('data', patch, kind + '.json'));
const ids = kind => new Set(catalog(kind).map(x => x.id));
const championIDs = ids('champions');
const augmentIDs = ids('augments');
const itemIDs = ids('items');
for (const g of guides.values()) {
  assert.ok(!/活力再生|最终都市列车|卡皮巴拉空投|mobiuscoil|9189_train|orbitallaser/.test(JSON.stringify(g)), 'Unavailable augment remains: ' + g.id);
  assert.equal(g.reviewed, false);
  assert.ok(g.sections.length);
  assert.ok(!['sample_count', 'win_rate', 'pick_rate', 'tier'].some(key => key in g));
  for (const id of g.champion_ids) assert.ok(championIDs.has(id), g.id + ': hero ' + id);
  for (const id of [...g.core_augment_ids, ...g.alternative_augment_ids, ...g.recommendations.map(r => r.id)]) assert.ok(augmentIDs.has(id), g.id + ': augment ' + id);
  for (const s of g.stages) for (const id of s.item_ids) assert.ok(itemIDs.has(id), g.id + ': item ' + id);
}
for (const old of backup.rows.filter(r => r.table === 'gameplay_guide_previews')) {
  const g = guides.get(old.key);
  assert.ok(g, 'Existing record removed: ' + old.key);
  assert.deepEqual(g.stages, old.content.stages, 'Existing equipment replaced: ' + old.key);
  if (!coverage.updates.includes(old.key)) assert.deepEqual(g, old.content);
  for (const recommendation of old.content.recommendations) assert.ok(g.recommendations.some(r => r.id === recommendation.id && r.reason === recommendation.reason));
}
assert.equal(guides.size, backup.rows.filter(r => r.table === 'gameplay_guide_previews').length + coverage.inserts);
assert.ok(!guides.get('haidou-garen-basic'));
assert.ok(!guides.get('haidou-soraka-build'), 'Duplicate existing playstyle');
assert.ok(!guides.get('haidou-senna-build'), 'Unresolved core augment imported');
assert.ok(!guides.get('haidou-bard-build'), 'Unresolved core augment imported');
assert.ok(!guides.get('haidou-ambessa-basic').recommendations.some(r => r.id === 'ARAM_EtherealWeapon'), 'Negative relation became a recommendation');
const kaisa = guides.get('haidou-kaisa-basic').recommendations.find(r => r.id === catalog('augments').find(a => a.name === '扳机炼狱').id);
assert.equal(kaisa.priority, 'situational');
assert.match(kaisa.reason, /位置/);
assert.equal(guides.get('haidou-jinx-basic').summary, 'Q 炮形态未必获得回归基本功增伤，别直接按机枪形态推算。');

const data = {
  champions: catalog('champions').map(c => ({ ...c, name_zh: c.name })),
  augments: catalog('augments').map(a => ({ ...a, name_zh: a.name })),
  items: catalog('items').map(i => ({ ...i, name_zh: i.name })),
  guides: [...guides.values()]
};
const offsets = [];
let view;
const app = create({ previewGuides: true, read: () => null, write: () => {}, render: value => { view = value; },
  async request(requestPath) {
    if (requestPath.includes('patches/latest')) return { version: patch };
    const url = new URL(requestPath, 'http://local');
    const kind = url.pathname.split('/').pop();
    const offset = Number(url.searchParams.get('offset'));
    assert.equal(url.searchParams.get('patch'), patch);
    assert.equal(url.searchParams.get('preview'), kind === 'guides' ? '1' : null);
    if (kind === 'guides') offsets.push(offset);
    return { patch, items: data[kind].slice(offset, offset + 100) };
  }
});
(async () => {
  await app.load();
  assert.equal(view.error, '');
  assert.equal(view.guideError, '');
  assert.deepEqual(offsets, [0, 100, 200]);
  for (const guide of guides.values()) {
    app.enter('detail', { kind: 'guides', id: guide.id });
    assert.equal(view.guideDetail.id, guide.id);
    assert.deepEqual(view.guideDetail.stages.map(s => s.item_ids), guide.stages.map(s => s.item_ids));
    for (const a of view.guideDetail.augmentChoices) assert.equal(a.reason, guide.recommendations.find(r => r.id === a.id)?.reason || '');
  }
  console.log('PASS: ' + guides.size + ' preview records, current-patch references, preserved old content, negative relation handling, full pagination and independent guide details');
})().catch(error => { console.error(error); process.exitCode = 1; });
