const root = document.querySelector('#app');
const esc = value => String(value ?? '').replace(/[&<>"']/g, char => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[char]);
const icon = (name, extra = '') => `<img class="icon ${extra}" src="./assets/${name}.png" alt="" width="20" height="20">`;
const button = (action, label, attrs = '', cls = 'text-action') => `<button class="${cls}" data-action="${action}" ${attrs}>${label}</button>`;
const route = (mode, query = '') => `#${mode}${query ? `?query=${encodeURIComponent(query)}` : ''}`;
const detailLink = row => `#detail?kind=${encodeURIComponent(row.kind)}&id=${encodeURIComponent(row.id)}`;
const heroExampleLink = row => `${detailLink(row)}&playstyle=${encodeURIComponent(row.playstyle)}`;
const safeURL = value => { try { const url = new URL(value); return ['http:', 'https:'].includes(url.protocol) ? url.href : ''; } catch (_) { return ''; } };
const avatar = (row, cls = 'avatar') => row.icon_url ? `<img class="${cls} ${esc(row.rarityClass)}" src="${esc(row.icon_url)}" alt="" loading="lazy" onerror="this.style.visibility='hidden'">` : `<span class="${cls} placeholder">${icon('book-open')}</span>`;
const link = (mode, text, cls = 'text-action', query = '') => `<a class="${cls}" href="${route(mode, query)}">${text}${icon('chevron-right')}</a>`;
const section = (text, right = '') => `<div class="section-heading"><h2 class="section-title">${text}</h2>${right}</div>`;
const empty = text => `<div class="empty-inline">${text}</div>`;
const tokens = rows => `<div class="token-list">${rows.map(row => row.missing ? `<span class="token missing">${esc(row.name)}<small>当前模式未收录</small></span>` : `<a class="token" href="${detailLink(row)}">${avatar(row, 'token-icon')}<span>${esc(row.name)}</span></a>`).join('')}</div>`;
function entry(row) {
  return `<a class="entry ${row.kind === 'augments' ? 'augment-entry' : ''}" href="${row.missing ? '#' : detailLink(row)}" ${row.missing ? 'data-action="missing"' : ''}>${avatar(row)}<span class="entry-copy"><strong class="entry-name">${esc(row.name)}</strong><span class="entry-meta">${esc(row.meta)}</span>${row.linkageLabel ? `<span class="linkage-count">${esc(row.linkageLabel)}</span>` : ''}${row.kind === 'augments' ? `<span class="entry-effect">${esc(row.description)}</span>` : ''}${row.reason ? `<span class="entry-effect">${esc(row.priority)} · ${esc(row.reason)}</span>` : ''}</span>${icon('chevron-right', 'chevron')}</a>`;
}
function guide(row) {
  return `<article class="guide-card" data-guide-id="${esc(row.id)}"><a class="guide-title" href="${detailLink(row)}"><div><span class="overline">${esc(row.meta)}</span><h3 class="section-title">${esc(row.title)}</h3></div></a><a class="body-copy guide-summary" href="${detailLink(row)}">${esc(row.augmentRelation ? row.augmentRelation.reason : row.summary)}</a>${row.core.length ? `<div class="guide-core"><span class="band-label">核心</span>${tokens(row.core)}</div>` : ''}<a class="guide-open" href="${detailLink(row)}">查看玩法${icon('chevron-right')}</a></article>`;
}
function heroExample(row) {
  return `<article class="hero-example"><a class="hero-example-heading" href="${heroExampleLink(row)}">${avatar(row)}<span class="entry-copy"><strong class="entry-name">${esc(row.name)}</strong><span class="entry-meta">${esc(row.meta)}</span></span>${icon('chevron-right')}</a>${row.relations.map(relation => `<div class="hero-example-reason"><span class="overline">${esc(relation.priorityText)}</span><p class="body-copy">${esc(relation.reason)}</p><a class="text-action" href="#detail?kind=guides&id=${encodeURIComponent(relation.guideID)}">${esc(relation.playstyle)}${icon('arrow-up-right')}</a></div>`).join('')}</article>`;
}
function segments(options, active, field = 'filter') {
  return `<div class="segments" role="tablist">${options.map(([id, text]) => button('change', esc(text), `data-field="${field}" data-value="${id}" role="tab" aria-selected="${active === id}"`, active === id ? 'selected' : '')).join('')}</div>`;
}
function rarityFilter(active) { return segments([['', '全部'], ['Prismatic', '棱彩'], ['Gold', '黄金'], ['Silver', '白银']], active); }
function search(v) {
  return `<div class="search-control">${icon('search')}<input id="query" type="search" aria-label="搜索资料" placeholder="${v.mode === 'search' ? '英雄、海克斯、装备或玩法' : v.mode === 'guides' ? '搜索英雄、玩法或海克斯' : '搜索名称或效果'}" value="${esc(v.query)}" autocomplete="off">${v.query ? button('clear-search', icon('x'), 'title="清空搜索"', 'icon-button') : ''}</div>`;
}
function guideState(v, blank) {
  return v.guidesLoading ? empty('玩法加载中') : v.guideError ? empty(`${esc(v.guideError)}${button('retryGuides', '重试')}`) : empty(blank);
}
function home(v) {
  return `<div class="home-heading"><span class="overline">HEXIS / 游戏手册</span><div class="heading-line"><h1 class="page-title">海克斯大乱斗</h1>${button('settings', `<span class="status-dot"></span>${esc(v.patch)}`, '', 'version')}</div></div><a class="search-control" href="#search">${icon('search')}<span>搜索英雄、海克斯、装备与玩法</span>${icon('arrow-up-right')}</a><div class="quick-links"><div><a class="inline-count" href="#guides">${icon('sparkles-active')}<div><strong class="entry-name">玩法手册</strong><span class="entry-meta">联动与出装思路</span></div></a>${link('guides', '')}</div><div><a class="inline-count" href="#items">${icon('backpack-active')}<div><strong class="entry-name">装备速查</strong><span class="entry-meta">${v.counts.items} 件装备</span></div></a>${link('items', '')}</div></div>${section('英雄速查', link('champions', `全部 ${v.counts.champions}`))}<div class="hero-grid">${v.heroes.map(row => `<a class="hero-tile" href="${detailLink(row)}">${avatar(row, 'hero-avatar')}<span>${esc(row.name)}</span><span class="entry-meta">${esc(row.meta)}</span></a>`).join('')}</div>${section('玩法精选', link('guides', '全部玩法'))}${v.featured.length ? v.featured.map(guide).join('') : guideState(v, `${icon('book-open')}暂无已审核玩法`)}${v.recent.length ? `${section('最近查看', link('mine', '查看全部'))}${v.recent.map(entry).join('')}` : ''}<div class="catalog-counts"><span>${v.counts.champions} 英雄</span><span>${v.counts.augments} 海克斯</span><span>${v.counts.items} 装备</span></div>`;
}
function list(v) {
  const filters = v.mode === 'augments' ? `${rarityFilter(v.filter)}<div class="filter-wrap">${button('change', '全部海克斯', `data-field="linkedOnly" data-value=""`, `chip ${!v.linkedOnly ? 'selected' : ''}`)}${button('change', `有联动 · ${v.linkedAugmentCount}`, 'data-field="linkedOnly" data-value="1"', `chip ${v.linkedOnly ? 'selected' : ''}`)}</div>${v.guidesLoading || v.guideError ? guideState(v, '') : ''}` : v.mode === 'items' ? `<div class="filter-wrap">${['', '攻击', '法术', '防御', '功能', '鞋子', '消耗品', '其他'].map(text => button('change', text || '全部', `data-field="filter" data-value="${text}"`, `chip ${v.filter === text ? 'selected' : ''}`)).join('')}${button('change', `${icon('arrow-down-up')}<span>${v.sort === 'price' ? '按价格' : '按名称'}</span>`, `data-field="sort" data-value="${v.sort === 'price' ? 'name' : 'price'}"`, 'sort-control')}</div>` : '';
  return `${search(v)}${filters}<div class="result-line"><span>${v.total} 条${v.mode === 'guides' ? (v.query ? '匹配玩法' : '玩法') : (v.query ? '匹配资料' : '资料')}</span>${v.pageCount > 1 ? `<span>${v.page} / ${v.pageCount} 页</span>` : ''}</div><div class="${v.mode !== 'guides' ? 'catalog-layout' : ''}">${v.rows.map(v.mode === 'guides' ? guide : entry).join('')}</div>${v.mode === 'guides' && (v.guidesLoading || v.guideError || !v.rows.length) ? guideState(v, v.query ? '没有找到匹配玩法，试试英雄名或海克斯名' : '暂无已审核玩法') : !v.rows.length ? empty('没有找到匹配资料') : ''}${v.pageCount > 1 ? `<div class="pager">${button('change', icon('chevron-left'), `data-field="page" data-value="${v.page - 1}" aria-label="上一页" title="上一页" ${v.page === 1 ? 'disabled' : ''}`, '')}<span>第 ${v.page} / ${v.pageCount} 页</span>${button('change', icon('chevron-right'), `data-field="page" data-value="${v.page + 1}" aria-label="下一页" title="下一页" ${v.page === v.pageCount ? 'disabled' : ''}`, '')}</div>` : ''}`;
}
function searchPage(v) {
  return `${search(v)}${!v.query ? `${section('最近搜索')}<div class="filter-wrap">${v.searches.map(text => button('change', esc(text), `data-field="query" data-value="${esc(text)}"`, 'chip')).join('') || '<p class="body-copy">暂无搜索记录</p>'}</div>` : !v.groups.length ? empty('没有找到匹配资料') : v.groups.map(group => `${section(group.title, `<span class="muted">${group.count} 条</span>`)}<div class="catalog-layout">${group.items.map(entry).join('')}</div>${group.count > 20 ? link(group.kind, '查看全部', 'text-action', v.query) : ''}`).join('')}`;
}
function mine(v) {
  return `<div class="personal-heading">${avatar({ icon_url: './assets/bookmark-active.png' }, 'personal-icon')}<div><h2 class="section-title">我的手册</h2><p class="body-copy">${v.favoritesCount} 项收藏 · 本机保存</p></div>${button('settings', icon('info'), 'title="资料与版本"', 'icon-button')}</div>${section('收藏')}${v.favorites.length ? v.favorites.map(entry).join('') : empty('暂无收藏')}${section('最近查看', button('clear', icon('trash-2'), 'title="清除最近查看"', 'icon-button'))}${v.recent.length ? v.recent.map(entry).join('') : empty('暂无查看记录')}${button('settings', `<span>资料与版本</span><span>${esc(v.patch)}</span>${icon('chevron-right')}`, '', 'source-line')}`;
}
function hero(v) {
  const skillContent = v.skillsOpen ? `<div class="skills">${v.skillsLoading ? empty('技能加载中') : v.skillsError ? empty(`${v.skillsError}${button('skills-retry', '重试')}`) : (v.skills?.skills || []).map(skill => `<div class="skill">${avatar(skill)}<div><strong class="entry-name">${esc(skill.key)} · ${esc(skill.name)}</strong><p class="body-copy">${esc(skill.description)}</p></div></div>`).join('')}${v.skills ? `<span class="source-caption">Data Dragon ${esc(v.skills.version)} · 技能数值及模式调整以游戏内为准</span>` : ''}</div>` : '';
  const content = v.guidesLoading || v.guideError ? guideState(v, '') : `${v.heroGuides.length ? v.heroGuides.map(guide).join('') : empty('该英雄暂无已审核的趣味玩法')}${v.commonBasics.length ? `${section('英雄通用基础')}${v.commonBasics.map(guide).join('')}` : ''}`;
  return `${button('skills', `${icon('swords')}<span>英雄技能</span>${icon(v.skillsOpen ? 'x' : 'chevron-right')}`, `aria-expanded="${v.skillsOpen}"`, 'source-line skill-toggle')}${skillContent}${section('英雄玩法', v.playstyle ? button('change', `全部玩法${icon('chevron-right')}`, 'data-field="playstyle" data-value=""') : '')}${content}<div class="detail-links">${link('augments', '全部海克斯')}${link('items', '装备速查')}</div>`;
}
function detail(v) {
  const row = v.detail;
  if (!row) return empty('当前版本未收录该资料');
  const heading = `<div class="detail-heading ${v.kind === 'guides' ? 'guide-heading' : ''}">${avatar(row, 'detail-avatar')}<div class="detail-copy"><span class="overline">${({ champions: '英雄档案', augments: '强化符文', items: '装备档案', guides: '编辑玩法' })[v.kind]} / ${esc(v.patch)}</span><h1 class="detail-title">${esc(row.name)}</h1><span class="body-copy">${esc(v.skills ? `${v.skills.title} · ${v.skills.roles.join(' / ')}` : row.meta)}</span></div>${button('favorite', icon(row.saved ? 'bookmark-active' : 'bookmark'), `data-kind="${v.kind}" data-id="${esc(v.id)}" title="${row.saved ? '取消收藏' : '收藏'}" aria-pressed="${row.saved}"`, `icon-button ${row.saved ? 'is-saved' : ''}`)}</div>`;
  let body = '';
  if (v.kind === 'champions') body = hero(v);
  if (v.kind === 'augments' || v.kind === 'items') {
    body = `${section(v.kind === 'augments' ? '海克斯效果' : '属性与效果', `<span class="muted">${esc(row.rarityName)}</span>`)}<p class="effect-body">${esc(row.description)}</p>`;
    if (v.kind === 'items') body += `<div class="equipment-prices"><div><span class="overline">总价</span><span>${row.price} <small class="muted">金币</small></span></div><div><span class="overline">卖价</span><span>${row.sell_price} <small class="muted">金币</small></span></div><span class="tag">${row.purchasable ? '可购买' : '变形或特殊装备'}</span></div>${row.required_champion ? `<p class="body-copy">专属英雄：${esc(row.required_champion)}</p>` : ''}${v.recipe.length ? `${section('合成配方')}${tokens(v.recipe)}` : ''}`;
    else body += `${section('联动英雄', `<span class="muted">${v.relatedHeroes.length} 位实例</span>`)}${v.relatedExamples.length ? v.relatedExamples.map(heroExample).join('') : guideState(v, '该海克斯的英雄实例待补充')}${v.moreRelatedExamples ? button('change', '展开全部英雄实例', 'data-field="examplesExpanded" data-value="1"') : ''}${section('关联玩法', `<span class="muted">${v.relatedGuides.length} 条</span>`)}${v.relatedGuides.length ? v.relatedGuides.map(guide).join('') : guideState(v, '该海克斯的关联玩法待补充')}${!v.relatedGuides.length ? `<a class="text-action" href="#augments?linkedOnly=1">查看已有联动的海克斯${icon('chevron-right')}</a>` : ''}`;
  }
  if (v.kind === 'guides') {
    const g = v.guideDetail;
    body = `<div class="guide-intro"><p class="effect-body">${esc(g.summary)}</p>${g.heroes.length ? tokens(g.heroes) : ''}</div>${g.readingSections.length ? `${section('怎么玩')}${g.readingSections.map(s => `<div class="basic-section"><h3 class="band-label">${esc(s.title)}</h3><p class="effect-body">${esc(s.body)}</p></div>`).join('')}` : ''}${g.augmentChoices.length ? `${section('海克斯搭配')}${g.augmentChoices.map(augment => `<div class="augment-choice"><a class="entry" href="${detailLink(augment)}">${avatar(augment)}<span class="entry-copy"><span class="overline">${esc(augment.priority)}</span><strong class="entry-name">${esc(augment.name)}</strong></span>${icon('chevron-right', 'chevron')}</a>${augment.reason ? `<p class="effect-body">${esc(augment.reason)}</p>` : ''}</div>`).join('')}` : ''}${g.category !== 'basic' || g.stages.length ? `${section('装备选择')}${g.stages.length ? g.stages.map(stage => `<div class="guide-band"><span class="band-label">${esc(stage.label)}</span>${tokens(stage.items)}<p class="body-copy">${esc(stage.reason)}</p></div>`).join('') : empty('配套装备待补充')}` : ''}${g.versionNotes.length ? `<div class="version-notes"><h3 class="band-label">版本提示</h3>${g.versionNotes.map(note => `<p class="body-copy">${esc(note.body)}</p>`).join('')}</div>` : ''}`;
  }
  let url;
  try { const parsed = new URL(v.source.url); if (['http:', 'https:'].includes(parsed.protocol)) url = parsed.href; } catch (_) {}
  return `${heading}<div class="detail-content">${body}<div class="source-footer"><span class="overline">资料来源</span><span class="body-copy">${esc(v.source.name)}</span>${v.source.attribution ? `<span class="body-copy">${esc(v.source.attribution)}</span>` : ''}${url ? `<a class="text-action" href="${esc(url)}" target="_blank" rel="noopener noreferrer">查看来源${icon('arrow-up-right')}</a>` : ''}${(v.guideDetail?.references || []).map(ref => safeURL(ref.url) ? `<a class="text-action" href="${esc(safeURL(ref.url))}" target="_blank" rel="noopener noreferrer">${esc(ref.name)}${icon('arrow-up-right')}</a>${ref.note ? `<span class="body-copy">${esc(ref.note)}</span>` : ''}` : '').join('')}</div></div>`;
}
function render(v) {
  const active = document.activeElement;
  const focused = active?.id === 'query';
  const cursor = focused ? active.selectionStart : null;
  let content = v.loading ? `<div class="state-view"><div class="loading-line"></div><span>正在加载版本资料</span></div>` : v.error ? `<div class="state-view">${icon('wifi-off')}<h2 class="section-title">暂时无法连接</h2><span class="body-copy">${esc(v.error)}</span>${button('retry', '重新加载', '', 'command')}</div>` : v.mode === 'home' ? home(v) : v.mode === 'mine' ? mine(v) : v.mode === 'detail' ? detail(v) : v.mode === 'search' ? searchPage(v) : list(v);
  root.innerHTML = `<div class="workspace">${!['home', 'detail'].includes(v.mode) ? `<div class="heading-line page-heading"><h1 class="page-title">${esc(v.title)}</h1><span class="muted">${v.mode === 'mine' ? `${v.favoritesCount} 项收藏` : esc(v.patch)}</span></div>` : ''}${content}</div>`;
  document.title = `${v.title} · HEXIS`;
  document.querySelector('#tabs').innerHTML = v.tabs.map(tab => `<a href="#${tab.id}" ${v.mode === tab.id ? 'aria-current="page"' : ''}><img src="./assets/${tab.icon}${v.mode === tab.id ? '-active' : ''}.png" alt=""><span>${tab.title}</span></a>`).join('');
  document.querySelector('#header-action').innerHTML = ['detail', 'search', 'guides'].includes(v.mode) ? button('back', `${icon('chevron-left')}<span>返回</span>`, '', 'back-action') : button('settings', icon('info'), 'title="资料与版本"', 'icon-button');
  if (focused) { const input = document.querySelector('#query'); input?.focus(); if (cursor != null) { try { input.setSelectionRange(cursor, cursor); } catch (_) {} } }
}
let toastTimer;
const app = HexisCore.create({
  previewGuides: new URLSearchParams(location.search).get('guides') === 'preview',
  read: key => { try { return JSON.parse(localStorage.getItem(key)); } catch (_) { return null; } },
  write: (key, value) => { try { localStorage.setItem(key, JSON.stringify(value)); } catch (_) {} },
  notice: message => { const toast = document.querySelector('#toast'); toast.textContent = message; toast.hidden = false; clearTimeout(toastTimer); toastTimer = setTimeout(() => { toast.hidden = true; }, 2000); },
  render,
  request: async path => { const response = await fetch(path); if (!response.ok) throw new Error(response.status === 404 ? '暂无对应版本资料' : '资料加载失败，请重试'); return response.json(); }
});
function navigate() {
  const [mode, query] = location.hash.slice(1).split('?');
  const params = Object.fromEntries(new URLSearchParams(query || ''));
  app.enter(['home', 'champions', 'augments', 'items', 'mine', 'detail', 'search', 'guides'].includes(mode) ? mode : 'home', params);
  window.scrollTo(0, 0);
}
window.addEventListener('hashchange', navigate);
root.addEventListener('input', event => { if (event.target.id === 'query') app.change('query', event.target.value); });
root.addEventListener('change', event => { if (event.target.id === 'playstyle') app.change('playstyle', event.target.value); });
root.addEventListener('keydown', event => { if (event.target.id === 'query' && event.key === 'Enter') { app.rememberSearch(); event.target.blur(); } });
document.addEventListener('click', async event => {
  const target = event.target.closest('[data-action]');
  if (!target) return;
  event.preventDefault();
  const d = target.dataset;
  if (d.action === 'change') { app.change(d.field, d.value); if (d.field === 'page') window.scrollTo(0, 0); }
  else if (d.action === 'clear-search') app.change('query', '');
  else if (d.action === 'favorite') app.toggleFavorite(d.kind, d.id);
  else if (d.action === 'skills') app.skills();
  else if (d.action === 'skills-retry') { await app.skills(); app.skills(); }
  else if (d.action === 'retry') app.load();
  else if (d.action === 'retryGuides') app.retryGuides();
  else if (d.action === 'back') { if (history.length > 1) history.back(); else location.hash = 'home'; }
  else if (d.action === 'clear') { if (confirm('清除最近查看和搜索记录？收藏内容会保留。')) app.clearRecent(); }
  else if (d.action === 'missing') { const toast = document.querySelector('#toast'); toast.textContent = '当前版本未收录'; toast.hidden = false; setTimeout(() => { toast.hidden = true; }, 2000); }
  else if (d.action === 'close-dialog') document.querySelector('#info').close();
  else if (d.action === 'settings') { document.querySelector('#info-content').innerHTML = `<span class="version"><span class="status-dot"></span>版本 ${esc(app.view().patch)}</span><p class="body-copy">图鉴来自 Riot Data Dragon 与 CommunityDragon。玩法仅展示已审核的编辑内容。</p><p class="body-copy">收藏与查看记录保存在当前设备。</p>`; document.querySelector('#info').showModal(); }
});
navigate();
app.load();
