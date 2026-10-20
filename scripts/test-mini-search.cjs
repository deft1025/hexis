const assert = require('node:assert/strict');
const { create } = require('../miniprogram/lib/core');

const champions = [{ id: 'Graves', name_zh: '法外狂徒', name_en: 'Graves' }];
const augments = ['核心联动', '备选联动', '按需联动'].map((name, i) => ({ id: `augment-${i}`, name_zh: name }));
const guide = {
  id: 'build', title: '持续输出', playstyle_name: '装填循环', category: 'fun',
  summary: '持续作战', tags: ['拉扯'], champion_ids: ['Graves'],
  core_augment_ids: ['augment-0'], alternative_augment_ids: ['augment-1'],
  recommendations: [{ id: 'augment-2', priority: 'situational', reason: '按局势选择' }]
};
const guides = [
  ...Array.from({ length: 27 }, (_, i) => ({ ...guide, id: `build-${i}`, title: `持续输出-${String(26 - i).padStart(2, '0')}` })),
  { ...guide, id: 'basic', title: '通用基础', category: 'basic' },
  { id: 'unrelated', title: '另一玩法', category: 'fun', champion_ids: ['missing'], core_augment_ids: ['missing'] }
];
const records = { champions, augments, items: [{ id: 'item', name_zh: '测试装备', description: '护盾效果' }], guides };
const app = create({
  read: () => null, write: () => {}, render: () => {},
  async request(path) {
    if (path.endsWith('/latest')) return { version: '16.20' };
    const url = new URL(path, 'http://local');
    const offset = Number(url.searchParams.get('offset'));
    return { patch: '16.20', items: records[url.pathname.split('/').pop()].slice(offset, offset + 100) };
  }
});

(async () => {
  await app.load();
  for (const query of ['男枪', '法外狂徒', 'GRAVES', '  Graves  ', '核心联动', '备选联动', '按需联动', '装填循环', '持续作战', '拉扯', 'build-1', '不存在']) {
    app.enter('search', { query });
    const group = app.view().groups.find(row => row.kind === 'guides');
    const globalIDs = (group?.items || []).filter(row => row.category !== 'basic').map(row => row.id);
    app.enter('guides', { query });
    const view = app.view();
    assert.deepEqual(view.rows.slice(0, globalIDs.length).map(row => row.id), globalIDs, `matching/order differs for ${query}`);
    const expected = query === '不存在' ? 0 : query === 'build-1' ? 11 : 27;
    assert.equal(view.total, expected, `handbook count differs for ${query}`);
    assert.equal(group?.count || 0, expected + (expected === 27 ? 1 : 0), `global count differs for ${query}`);
    if (query !== '不存在') assert(view.total > 0, `expected matches for ${query}`);
  }
  app.enter('search', { query: '  男枪  ' });
  const globalView = app.view();
  const guideGroup = globalView.groups.find(row => row.kind === 'guides');
  assert.equal(guideGroup.count, 28);
  assert.equal(guideGroup.items.length, 20);
  assert.equal(globalView.resultCount, globalView.groups.reduce((sum, group) => sum + group.count, 0));
  app.enter('guides', { query: '  男枪  ' });
  assert.equal(app.view().total, 27);
  app.change('page', 2);
  assert.equal(app.view().rows.length, 3);
  assert(!app.view().rows.some(row => row.id === 'basic'));
  for (const [kind, query, id] of [['champions', '  GRAVES  ', 'Graves'], ['augments', '  核心联动  ', 'augment-0'], ['items', '  护盾效果  ', 'item']]) {
    app.enter('search', { query });
    assert.equal(app.view().groups.find(row => row.kind === kind).items[0].id, id);
    app.enter(kind, { query });
    assert.equal(app.view().rows[0].id, id);
  }
  app.enter('search', { query: '   ' });
  assert.equal(app.view().resultCount, 0);
  app.enter('guides', { query: '   ' });
  assert.equal(app.view().total, 28);
  console.log('PASS: global/handbook search parity for hero aliases, linked augments, playstyle names, case/whitespace, ordering, counts and pagination; global basics remain searchable');
})().catch(error => { console.error(error); process.exitCode = 1; });
