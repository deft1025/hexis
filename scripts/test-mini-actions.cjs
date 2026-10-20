const assert = require('node:assert/strict');
const { create } = require('../miniprogram/lib/core');
const createPage = require('../miniprogram/lib/page');

(async () => {
  let saved = { favorites: [
    { key: 'guides:old', kind: 'guides', id: 'old', name: '已下架玩法' },
    { key: 'champions:shared', kind: 'champions', id: 'shared', name: '英雄' },
    { key: 'items:shared', kind: 'items', id: 'shared', name: '装备' }
  ] };
  const records = { champions: [{ id: 'shared', name_zh: '英雄' }], items: [{ id: 'shared', name_zh: '装备' }], augments: [], guides: [] };
  const app = create({
    read: () => saved, write: (_, value) => { saved = value; }, render: () => {}, notice: () => {},
    async request(path) {
      if (path.endsWith('/latest')) return { version: '16.20' };
      return { patch: '16.20', items: records[new URL(path, 'http://local').pathname.split('/').pop()] };
    }
  });
  await app.load();
  app.enter('mine');
  assert.equal(app.view().favorites[0].missing, true);
  assert.equal(new Set(app.view().favorites.map(row => row.key)).size, 3);
  app.toggleFavorite('guides', 'old');
  assert.equal(app.view().favoritesCount, 2);
  assert(!saved.favorites.some(row => row.id === 'old'));
  app.toggleFavorite('guides', 'old');
  assert.equal(app.view().favoritesCount, 2, 'cannot add unavailable records');
  app.toggleFavorite('items', 'shared');
  assert.equal(app.view().favorites[0].kind, 'champions', 'same ID in another catalog must remain');
  app.toggleFavorite('items', 'shared');
  assert.equal(app.view().favoritesCount, 2);

  let depth = 1, failPush = false, failReplace = false;
  const calls = [];
  global.getCurrentPages = () => Array.from({ length: depth });
  global.wx = {
    navigateTo(options) {
      calls.push(['push', options.url]);
      if (failPush) options.fail({ errMsg: failPush });
      else depth++;
    },
    redirectTo(options) {
      calls.push(['replace', options.url]);
      if (failReplace) options.fail();
    },
    showToast(options) { calls.push(['toast', options.title]); },
    showModal(options) { calls.push(['modal', options.content]); }
  };
  const page = createPage('detail');
  page._app = app;
  const action = dataset => page.action({ currentTarget: { dataset } });
  const detail = { action: 'open', kind: 'champions', id: 'a/b', playstyle: 'flow & 1' };
  const url = '/pages/detail/index?kind=champions&id=a%2Fb&playstyle=flow%20%26%201';
  action(detail);
  assert.deepEqual(calls.pop(), ['push', url]);
  depth = 10;
  action(detail);
  assert.deepEqual(calls.pop(), ['replace', url]);
  for (let i = 0; i < 20; i++) action(detail);
  assert.equal(depth, 10, 'repeated detail navigation cannot grow the stack beyond its limit');
  assert(calls.every(call => call[0] === 'replace'));
  calls.length = 0;
  depth = 9;
  failPush = 'navigateTo:fail webview count limit exceed';
  action(detail);
  assert.deepEqual(calls, [['push', url], ['replace', url]]);
  calls.length = 0;
  failPush = 'navigateTo:fail page not found';
  action(detail);
  assert.deepEqual(calls.map(call => call[0]), ['push', 'toast']);
  calls.length = 0;
  depth = 10;
  failReplace = true;
  action(detail);
  assert.deepEqual(calls.map(call => call[0]), ['replace', 'toast']);
  calls.length = 0;
  failReplace = false;
  action({ action: 'navigate', mode: 'guides', query: '男枪 & 塔姆' });
  assert.deepEqual(calls, [['replace', '/pages/guides/index?query=%E7%94%B7%E6%9E%AA%20%26%20%E5%A1%94%E5%A7%86']]);
  calls.length = 0;
  action({ action: 'open', kind: 'guides', id: 'old', missing: true });
  assert.equal(calls[0][0], 'toast');
  for (const previewGuides of [false, true]) {
    const infoApp = create({ previewGuides, read: () => null });
    page.data = infoApp.view();
    calls.length = 0;
    action({ action: 'settings' });
    assert.equal(calls[0][0], 'modal');
    assert(calls[0][1].includes(page.data.guideStatusText));
    assert.equal(calls[0][1].includes('仅展示已审核'), !previewGuides);
    assert(!calls[0][1].includes('undefined'));
  }
  console.log('PASS: unavailable favorites removable and persisted; catalog keys isolated; native navigation handles stack limit, fallback, failures and encoded destinations');
})().catch(error => { console.error(error); process.exitCode = 1; });
