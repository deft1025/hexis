const assert = require('node:assert/strict');
const { create } = require('../miniprogram/lib/core');
const heroes = Array.from({ length: 30 }, (_, i) => ({ id: i === 29 ? 'TahmKench' : `hero${i}`, name_zh: i === 29 ? '河流之王' : `英雄${i}`, icon_url: 'https://ddragon.leagueoflegends.com/cdn/16.20.1/img/champion/TahmKench.png' }));
const augment = { id: 'control', name_zh: '控制联动', rarity: 'Prismatic', description: '定身触发效果' };
const item = { id: '3083', name_zh: '狂徒铠甲', categories: ['防御'], price: 3300, recipe: [] };
const guide = { id: 'tank-control', title: '控制与坦度', category: 'fun', summary: '说明', champion_ids: ['TahmKench'], tags: ['坦克'], core_augment_ids: ['control'], alternative_augment_ids: [], recommendations: [{ id: 'control', priority: 'core', reason: '定身触发' }], stages: [{ label: '核心', item_ids: ['3083'], reason: '生命值' }], sections: [{ title: '联动', body: '说明' }], reviewed: true };
let storage = {};
let last;
let failGuides = false;
let failAll = false;
const data = { champions: heroes, augments: [augment], items: [item], guides: [guide] };
const app = create({
  read: key => storage[key], write: (key, value) => { storage[key] = value; }, notice: () => {}, render: value => { last = value; },
  async request(path) {
    if (failAll) throw new Error('offline');
    if (path.includes('patches/latest')) return { version: '16.20' };
    if (path.includes('/skills')) return { item: { title: '塔姆', roles: ['坦克'], skills: [] } };
    const url = new URL(path, 'http://local');
    const kind = url.pathname.split('/').pop();
    assert.equal(url.searchParams.has('preview'), false);
    if (kind === 'guides' && failGuides) throw new Error('unavailable');
    assert.equal(url.searchParams.get('patch'), '16.20');
    const offset = Number(url.searchParams.get('offset'));
    return { patch: '16.20', items: data[kind].slice(offset, offset + 100) };
  }
});
(async () => {
  await app.load();
  assert.equal(last.guideStatusText, '玩法仅展示已审核的编辑内容。');
  app.enter('champions');
  assert.equal(last.rows.length, 24);
  app.change('page', 2);
  assert.equal(last.rows.length, 6);
  app.change('query', '蛤蟆');
  assert.equal(last.rows[0].id, 'TahmKench');
  assert.equal(last.rows[0].icon_url, heroes[29].icon_url);
  assert.equal(last.page, 1);
  app.enter('detail', { kind: 'champions', id: 'TahmKench' });
  assert.equal(last.heroGuides[0].core[0].id, 'control');
  app.change('tab', 'augments');
  assert.equal(last.recommendations[0].reason, '定身触发');
  await app.skills();
  assert.equal(last.skills.roleText, '坦克');
  app.toggleFavorite('champions', 'TahmKench');
  app.enter('mine');
  assert.equal(last.favorites[0].id, 'TahmKench');
  assert.equal(last.recent[0].id, 'TahmKench');
  app.enter('detail', { kind: 'augments', id: 'control' });
  assert.equal(last.relatedHeroes[0].id, 'TahmKench');
  assert.equal(last.relatedItems[0].id, '3083');
  app.enter('search');
  app.change('query', '控制');
  assert.equal(last.resultCount, 2);
  app.rememberSearch();
  app.clearRecent();
  assert.equal(last.favoritesCount, 1);
  assert.equal(last.searches.length, 0);
  failGuides = true;
  await app.load();
  assert.equal(last.error, '');
  assert.match(last.guideError, /无法加载/);
  failGuides = false;
  await app.retryGuides();
  assert.equal(last.guideError, '');
  failAll = true;
  await app.load();
  assert.equal(last.error, 'offline');
  const previewApp = create({
    previewGuides: true,
    read: () => null, write: () => {}, render: value => { last = value; },
    async request(path) {
      if (path.includes('patches/latest')) return { version: '16.20' };
      const url = new URL(path, 'http://local');
      const kind = url.pathname.split('/').pop();
      assert.equal(url.searchParams.get('preview'), kind === 'guides' ? '1' : null);
      return { patch: '16.20', items: kind === 'guides' ? [{ ...guide, reviewed: false }] : data[kind] };
    }
  });
  await previewApp.load();
  assert.match(last.guideStatusText, /预览模式.*尚未审核/);
  assert(!last.guideStatusText.includes('仅展示已审核'));
  previewApp.enter('detail', { kind: 'champions', id: 'TahmKench' });
  assert.match(last.heroGuides[0].meta, /样例待核对/);
  previewApp.change('tab', 'augments');
  assert.match(last.recommendations[0].priority, /样例待核对/);
  const secondFlow = { ...guide, id: 'second-flow', title: '另一流派', playstyle_name: '另一流派', basic_sections: [{ title: '另一操作', body: '独立操作条件' }], recommendations: [{ id: 'control', priority: 'situational', reason: '第二流派的不同触发条件' }, { id: 'silver', priority: 'alternative', reason: '白银备选' }] };
  const common = { ...guide, id: 'common-basic', title: '通用基础', category: 'basic', recommendations: [{ id: 'control', priority: 'alternative', reason: '通用建议' }] };
  const multiApp = create({
    read: () => ({ favorites: [{ key: 'guides:common-basic', kind: 'guides', id: 'common-basic', name: '旧收藏标题' }], recent: [{ key: 'guides:common-basic', kind: 'guides', id: 'common-basic', name: '旧历史标题' }] }), write: () => {}, render: value => { last = value; },
    async request(path) {
      if (path.includes('patches/latest')) return { version: '16.20' };
      const kind = new URL(path, 'http://local').pathname.split('/').pop();
      const records = { ...data, guides: [guide, secondFlow, common], augments: [augment, { ...augment, id: 'silver', rarity: 'Silver' }] };
      return { patch: '16.20', items: records[kind] };
    }
  });
  multiApp.enter('detail', { kind: 'champions', id: 'TahmKench', playstyle: 'second-flow' });
  await multiApp.load();
  assert.equal(last.playstyle, 'second-flow', 'selection must survive loading');
  assert.deepEqual(last.heroGuides.map(row => row.id), ['second-flow']);
  multiApp.change('tab', 'basic');
  assert.equal(last.playstyle, 'second-flow');
  assert.equal(last.commonBasics[0].id, 'common-basic');
  assert.equal(last.playstyleBasics[0].basicSections[0].body, '独立操作条件');
  multiApp.change('tab', 'augments');
  assert.deepEqual(last.recommendationGroups.map(row => row.id), ['second-flow']);
  assert.equal(last.commonRecommendationGroups[0].recommendations[0].reason, '通用建议');
  multiApp.change('playstyle', '');
  assert.equal(last.recommendationGroups.length, 2);
  assert.equal(last.recommendationGroups[0].recommendations[0].reason, '定身触发');
  assert.equal(last.recommendationGroups[1].recommendations[0].reason, '第二流派的不同触发条件');
  multiApp.change('filter', 'Silver');
  assert.deepEqual(last.recommendationGroups.map(row => row.id), ['second-flow']);
  assert.deepEqual(last.recommendationGroups[0].recommendations.map(row => row.id), ['silver']);
  assert.equal(last.commonRecommendationGroups.length, 0);
  multiApp.change('playstyle', guide.id);
  assert.equal(last.filter, '');
  assert.equal(last.recommendationGroups[0].id, guide.id);
  await multiApp.load();
  assert.equal(last.playstyle, guide.id, 'selection must survive refresh');
  multiApp.enter('detail', { kind: 'champions', id: 'hero0' });
  assert.equal(last.playstyle, '');
  assert.equal(last.recommendationGroups.length, 0);
  multiApp.enter('detail', { kind: 'champions', id: 'TahmKench', playstyle: 'missing' });
  assert.equal(last.playstyle, '');
  assert.equal(last.heroGuides.length, 2);
  for (const [sections, heading] of [
    [[{ title: '基础', body: '操作建议' }], '基础思路'],
    [[{ title: '避坑：技能', body: '触发限制' }, { title: '版本提示', body: '待核对' }], '机制与避坑'],
    [[{ title: '基础', body: '操作建议' }, { title: '联动：技能', body: '触发条件' }], '基础思路与机制'],
    [[{ title: '版本提示', body: '待核对' }], '装备参考']
  ]) {
    common.sections = sections;
    await multiApp.load();
    const expected = `河流之王：${heading}`;
    multiApp.enter('mine');
    assert.equal(last.favorites.find(row => row.id === common.id).name, expected, 'existing favorites use the current display name');
    assert.equal(last.recent.find(row => row.id === common.id).name, expected, 'existing history uses the current display name');
    multiApp.enter('search', { query: '通用基础' });
    const result = last.groups.find(group => group.kind === 'guides').items.find(row => row.id === common.id);
    assert.equal(result.name, expected);
    assert.equal(result.title, '通用基础', 'source title remains unchanged');
    multiApp.enter('detail', { kind: 'guides', id: common.id });
    assert.equal(last.detail.name, expected);
    multiApp.toggleFavorite('guides', common.id);
    multiApp.toggleFavorite('guides', common.id);
    assert.equal(last.favorites.find(row => row.id === common.id).name, expected);
  }
  multiApp.enter('detail', { kind: 'guides', id: guide.id });
  assert.equal(last.detail.name, guide.title, 'independent playstyle names remain unchanged');
  common.sections = [{ title: '版本提示', body: '待核对' }];
  common.stages = [{ label: '装备候选', item_ids: [item.id], reason: '非固定购买顺序' }];
  await multiApp.load();
  multiApp.enter('detail', { kind: 'champions', id: 'TahmKench' });
  assert.equal(last.heroGuideHeading, '装备参考');
  assert.equal(last.commonBasics[0].name, '河流之王：装备参考');
  assert.equal(last.commonBasics[0].stages[0].items[0].name, item.name_zh);
  multiApp.enter('detail', { kind: 'guides', id: common.id });
  assert.equal(last.detail.readingHeading, '装备参考');
  console.log('PASS: catalog, favorites, skills, failure/retry, preview isolation, multi-flow selection across tabs and refresh, shared augments with separate reasons, common basics, rarity filtering and hero switching');
})().catch(error => { console.error(error); process.exitCode = 1; });
