const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const { create } = require('../miniprogram/lib/core');
const guides = ['tahmkench', 'garen', 'katarina', 'varus', 'karma', 'soraka'].flatMap(name => JSON.parse(fs.readFileSync(path.join(__dirname, '../docs/guide-previews', name + '.json'), 'utf8')));
const champions = [...new Set(guides.flatMap(g => g.champion_ids))].map(id => ({ id, name_zh: id }));
const augmentIDs = [...new Set(guides.flatMap(g => [...g.core_augment_ids, ...g.alternative_augment_ids, ...g.recommendations.map(r => r.id)]))];
const augments = [...augmentIDs.map(id => ({ id, name_zh: id, rarity: id === 'ARAM_EtherealWeapon' ? 'Gold' : 'Silver' })), { id: 'no-guide', name_zh: '尚无资料', rarity: 'Gold' }];
const items = [...new Set(guides.flatMap(g => g.stages.flatMap(s => s.item_ids)))].map(id => ({ id, name_zh: id, categories: [], price: 100 }));
let view;
let failGuides = false;
const app = create({
  read: () => null, write: () => {}, render: result => { view = result; },
  async request(requestPath) {
    if (requestPath.includes('patches/latest')) return { version: '16.20' };
    const url = new URL(requestPath, 'http://local');
    const kind = url.pathname.split('/').pop();
    if (kind === 'guides' && failGuides) throw new Error('guide outage');
    return { patch: '16.20', items: { champions, augments, items, guides }[kind] };
  }
});
(async () => {
  await app.load();
  app.enter('guides');
  assert.equal(view.total, 8);
  assert.ok(view.rows.every(row => row.category !== 'basic'));
  app.change('query', ' 蛤蟆 ');
  assert.equal(view.total, 3);
  assert.ok(view.rows.every(row => row.champion_ids.includes('TahmKench')));
  app.change('query', 'aram_firstaidkit');
  assert.deepEqual(view.rows.map(row => row.id).sort(), ['sample-tahm-first-aid', 'sample-karma-shield', 'sample-soraka-spirit-bomb'].sort());
  app.change('query', '没有这条玩法');
  assert.equal(view.total, 0);
  app.change('query', '');
  assert.equal(view.total, 8);
  app.enter('home');
  assert.ok(view.featured.every(row => row.category !== 'basic'));
  app.enter('detail', { kind: 'guides', id: 'sample-tahm-basic' });
  assert.equal(view.guideDetail.id, 'sample-tahm-basic');
  app.enter('augments', { linkedOnly: '1' });
  assert.equal(view.total, 10);
  assert.equal(view.rows.some(row => row.id === 'no-guide'), false);
  assert.equal(view.linkedAugmentCount, 10);
  app.change('filter', 'Gold');
  assert.deepEqual(view.rows.map(row => row.id), ['ARAM_EtherealWeapon']);
  assert.equal(view.rows[0].linkageLabel, '3 位英雄 · 3 条玩法');
  for (const [id, expected] of [
    ['ARAM_EtherealWeapon', ['TahmKench', 'Katarina', 'Varus']],
    ['ARAM_FirstAidKit', ['TahmKench', 'Karma', 'Soraka']],
    ['ARAM_SpinToWin', ['Garen', 'Katarina']]
  ]) {
    app.enter('detail', { kind: 'augments', id });
    assert.deepEqual(view.relatedHeroes.map(row => row.id).sort(), expected.sort());
    assert.equal(view.relatedExamples.length, expected.length);
    for (const hero of view.relatedExamples) {
      assert.ok(hero.relations.length > 0);
      assert.ok(hero.relations.every(r => r.reason && r.guideID && r.priorityText.includes('样例待核对')));
      assert.ok(view.relatedGuides.some(g => g.id === hero.playstyle && g.champion_ids.includes(hero.id)));
    }
    for (const g of view.relatedGuides) {
      assert.equal(g.augmentRelation.reason, g.recommendations.find(r => r.id === id).reason);
    }
  }
  app.enter('detail', { kind: 'augments', id: 'no-guide' });
  assert.equal(view.relatedHeroes.length, 0);
  assert.equal(view.relatedGuides.length, 0);
  app.enter('detail', { kind: 'guides', id: 'sample-karma-shield' });
  assert.match(view.guideDetail.references[0].url, /firstaidkit/);
  for (const guide of guides) {
    app.enter('detail', { kind: 'guides', id: guide.id });
    const detail = view.guideDetail;
    assert.equal(detail.id, guide.id);
    assert.deepEqual(detail.stages.map(stage => stage.item_ids), guide.stages.map(stage => stage.item_ids));
    assert.deepEqual(detail.readingSections, [...guide.sections, ...(guide.basic_sections || [])].filter(section => section.title !== '版本提示'));
    assert.deepEqual(detail.versionNotes, guide.sections.filter(section => section.title === '版本提示'));
    assert.equal(new Set(detail.augmentChoices.map(row => row.id)).size, detail.augmentChoices.length);
    for (const augment of detail.augmentChoices) {
      assert.equal(augment.reason, guide.recommendations.find(row => row.id === augment.id)?.reason || '');
    }
  }
  app.enter('detail', { kind: 'champions', id: 'TahmKench', playstyle: 'sample-tahm-first-aid' });
  assert.deepEqual(view.heroGuides.map(row => row.id), ['sample-tahm-first-aid']);
  assert.deepEqual(view.heroGuides[0].stages, []);
  app.change('playstyle', '');
  assert.equal(view.heroGuides.length, 3);
  assert.equal(view.commonBasics.length, 1);
  // The same champion can have distinct reasons in multiple guides, without duplicate hero cards.
  const extra = { ...guides[0], id: 'test-second-ethereal', recommendations: [{ id: 'ARAM_EtherealWeapon', priority: 'situational', reason: '另一流派的条件' }] };
  guides.push(extra);
  const multiHero = { ...extra, id: 'test-multi-hero', champion_ids: ['Karma', 'Soraka'] };
  guides.push(multiHero);
  guides.push({ ...extra, id: 'test-merged-content', core_augment_ids: ['ARAM_EtherealWeapon'], alternative_augment_ids: ['ARAM_FirstAidKit'], recommendations: [], sections: [{ title: '条件', body: '独立条件' }, { title: '版本提示', body: '仍待核对' }], basic_sections: [{ title: '条件', body: '独立条件' }] });
  await app.load();
  app.enter('detail', { kind: 'guides', id: 'test-merged-content' });
  assert.equal(view.guideDetail.readingSections.length, 1);
  assert.equal(view.guideDetail.versionNotes.length, 1);
  assert.deepEqual(view.guideDetail.augmentChoices.map(row => [row.id, row.reason, row.priority]), [['ARAM_EtherealWeapon', '', '优先考虑'], ['ARAM_FirstAidKit', '', '备选联动']]);
  guides.pop();
  await app.load();
  app.enter('detail', { kind: 'augments', id: 'ARAM_EtherealWeapon' });
  assert.equal(view.relatedHeroes.length, 5);
  assert.equal(view.relatedExamples.length, 3);
  assert.equal(view.moreRelatedExamples, true);
  app.change('examplesExpanded', '1');
  assert.equal(view.relatedExamples.length, 5);
  const tahm = view.relatedExamples.find(row => row.id === 'TahmKench');
  assert.equal(tahm.relations.length, 2);
  assert.equal(tahm.relations[1].reason, '另一流派的条件');
  failGuides = true;
  await app.load();
  assert.ok(view.guideError);
  assert.equal(view.relatedExamples.length, 0);
  console.log('PASS: source-backed hero examples, augment coverage and rarity filters, per-guide reasons, source links, deduplicated hero cards, expanded examples and guide outage');
})().catch(error => { console.error(error); process.exitCode = 1; });
