const page = document.body.dataset.page;
const $ = (selector) => document.querySelector(selector);
const escapeHTML = (value) => String(value ?? '').replace(/[&<>"']/g, (character) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[character]);
const percent = (value) => value == null ? '暂无数据' : `${(value * 100).toFixed(1)}%`;
const number = (value) => Number(value || 0).toLocaleString('zh-CN');
const rarityNames = { Silver: '白银', Gold: '黄金', Prismatic: '棱彩' };

async function api(path) {
  const response = await fetch(path);
  if (!response.ok) {
    const error = new Error(response.status === 404 ? '暂无已发布版本或对应数据' : '数据加载失败，请稍后重试');
    error.status = response.status;
    throw error;
  }
  return response.json();
}

async function loadPatch() {
  try {
    const patch = await api('/api/v1/patches/latest');
    $('#patch').textContent = `PATCH ${patch.version}`;
  } catch {
    $('#patch').textContent = '暂无已发布版本';
  }
}

function sourceMarkup(source) {
  if (!source?.name) return '<span>来源：暂无信息</span>';
  const label = escapeHTML(source.name);
  let url;
  try { url = new URL(source.url); } catch { return `<span>来源：${label}</span>`; }
  if (!['https:', 'http:'].includes(url.protocol)) return `<span>来源：${label}</span>`;
  return `<span>来源：<a href="${escapeHTML(url.href)}" target="_blank" rel="noopener noreferrer">${label}</a></span>`;
}

function iconMarkup(item) {
  let iconURL = item.icon_url;
  if (!iconURL) return '<span class="entity-icon icon-fallback" aria-hidden="true">✦</span>';
  try {
    const url = new URL(iconURL,location.origin);
    if (url.protocol !== 'https:' || !['ddragon.leagueoflegends.com', 'raw.communitydragon.org'].includes(url.hostname)) throw new Error('invalid icon URL');
    return `<img class="entity-icon" src="${escapeHTML(url.href)}" alt="" loading="lazy" onerror="this.hidden=true;this.nextElementSibling.hidden=false"><span class="entity-icon icon-fallback" aria-hidden="true" hidden>✦</span>`;
  } catch { return '<span class="entity-icon icon-fallback" aria-hidden="true">✦</span>'; }
}

function rankMarkup(index) { return `<span class="rank-number">${String(index + 1).padStart(2, '0')}</span>`; }
function rateMarkup(value) { return `<span class="rate-value">${percent(value)}</span>${value == null ? '' : `<span class="rate-bar"><i style="width:${Math.min(100, value * 150)}%"></i></span>`}`; }

function championMarkup(item, index) {
  const stats = item.stats || {};
  return `<a class="ranking-row" href="/champion.html?id=${encodeURIComponent(item.id)}">${rankMarkup(index)}<span class="entity-cell">${iconMarkup(item)}<span><strong>${escapeHTML(item.name_zh)}</strong><small>${escapeHTML(item.name_en || item.id)}</small></span></span><span class="rate-cell">${rateMarkup(stats.win_rate)}</span><span>${percent(stats.pick_rate)}</span><span class="recommendation-pending">专属数据待接入</span><span class="row-action">查看思路 →</span></a>`;
}

function augmentMarkup(item, index) {
  const stats = item.stats || {};
  return `<div class="ranking-row" data-rarity="${escapeHTML(item.rarity || 'unknown')}">${rankMarkup(index)}<span class="entity-cell">${iconMarkup(item)}<span><strong>${escapeHTML(item.name_zh)}</strong><small>${escapeHTML(item.description || item.name_en || '暂无效果描述')}</small></span></span><span class="rate-cell">${rateMarkup(stats.win_rate)}</span><span>${percent(stats.pick_rate)}</span><span class="rarity rarity-${escapeHTML(item.rarity || 'unknown')}">${escapeHTML(rarityNames[item.rarity] || item.rarity || '未分级')}</span></div>`;
}

function itemMarkup(item, index) {
  const categories = (item.categories || []).join(' · ');
  return `<div class="ranking-row">${rankMarkup(index)}<span class="entity-cell">${iconMarkup(item)}<span><strong>${escapeHTML(item.name_zh)}</strong><small class="item-mobile-category">${escapeHTML(categories)}</small></span></span><span class="item-category">${escapeHTML(categories)}</span><span class="item-price">${number(item.price)}</span><span class="item-effect">${escapeHTML(item.description)}${item.required_champion ? `<small>专属英雄：${escapeHTML(item.required_champion)}</small>` : ''}</span></div>`;
}

function comboMarkup(item) {
  const champions = item.champions || [];
  const names = champions.map((champion) => escapeHTML(champion.name_zh)).join(' <span class="plus">＋</span> ');
  const links = champions.map((champion) => `<a href="/champion.html?id=${encodeURIComponent(champion.id)}">${escapeHTML(champion.name_zh)} →</a>`).join('');
  return `<article class="data-card"><span class="eyebrow">双英雄搭配</span><h2>${names}</h2><div class="card-stats"><span>胜率 <strong class="positive">${percent(item.win_rate)}</strong></span><span>选取率 <strong>${percent(item.pick_rate)}</strong></span><span>样本 <strong>${number(item.sample_count)}</strong></span></div><div class="card-links">${links}</div>${sourceMarkup(item.source)}</article>`;
}

function setupList() {
  const endpoint = `/api/v1/${page}`;
  const list = $('#list');
  const status = $('#list-status');
  const retry = $('#load-more');
  const search = $('#search');
  const sort = $('#sort');
  const rarity = $('#rarity');
  const category = $('#category');
  const rarityTabs = [...document.querySelectorAll('[data-rarity]')];
  const pagers = ['上方分页', '下方分页'].map((label) => {
    const pager = document.createElement('nav');
    pager.className = 'pagination';
    pager.setAttribute('aria-label', label);
    pager.hidden = true;
    return pager;
  });
  status.after(pagers[0]);
  retry.before(pagers[1]);
  const initialPage = Number(new URLSearchParams(location.search).get('page'));
  let currentPage = Number.isSafeInteger(initialPage) && initialPage > 0 ? initialPage : 1;
  let pageSize = 20;
  let items = [];
  let loading = false;
  let loaded = false;
  let selectedRarity = '';

  function render(scroll = false) {
    if (!loaded) return;
    const query = search.value.trim().toLocaleLowerCase();
    let visible = items.filter((item) => {
      if (page === 'combos') return (item.champions || []).some((champion) => `${champion.name_zh} ${champion.id}`.toLocaleLowerCase().includes(query));
      const text = `${item.name_zh} ${item.name_en || ''} ${item.id} ${['augments', 'items'].includes(page) ? item.description || '' : ''}`;
      return text.toLocaleLowerCase().includes(query) && (!rarity || !rarity.value || item.rarity === rarity.value) && (!selectedRarity || item.rarity === selectedRarity) && (!category?.value || item.categories?.includes(category.value));
    });
    if (sort?.value === 'win') visible.sort((first, second) => (second.stats?.win_rate ?? second.win_rate ?? -1) - (first.stats?.win_rate ?? first.win_rate ?? -1));
    if (sort?.value === 'pick') visible.sort((first, second) => (second.stats?.pick_rate ?? second.pick_rate ?? -1) - (first.stats?.pick_rate ?? first.pick_rate ?? -1));
    if (sort?.value === 'samples') visible.sort((first, second) => second.sample_count - first.sample_count);
    if (sort?.value === 'price') visible.sort((first, second) => first.price - second.price);
    if (sort?.value === 'name') visible.sort((first, second) => first.name_zh.localeCompare(second.name_zh, 'zh-CN'));
    const pageCount = Math.max(1, Math.ceil(visible.length / pageSize));
    currentPage = Math.min(currentPage, pageCount);
    const start = (currentPage - 1) * pageSize;
    const markup = page === 'champions' ? championMarkup : page === 'augments' ? augmentMarkup : page === 'items' ? itemMarkup : comboMarkup;
    list.innerHTML = visible.slice(start, start + pageSize).map((item, index) => markup(item, start + index)).join('');
    if (page !== 'combos') {
      $('#hero-count').textContent = `${items.length} ${page === 'champions' ? '位英雄' : page === 'items' ? '件装备' : '个符文'}`;
    }
    if (['champions', 'augments'].includes(page)) {
      const leader = items.filter((item) => item.stats?.win_rate != null).sort((first, second) => second.stats.win_rate - first.stats.win_rate)[0];
      $('#leader').innerHTML = leader ? `<small>当前版本胜率领先</small><strong>${percent(leader.stats.win_rate)}</strong><span>${escapeHTML(leader.name_zh)}</span>` : '<small>当前版本胜率领先</small><strong>—</strong><span>暂无统计数据</span>';
    }
    status.textContent = visible.length ? `共 ${visible.length} 条 · 显示 ${start + 1}–${Math.min(start + pageSize, visible.length)} 条` : items.length ? '没有匹配的结果' : '暂无数据';
    if (!visible.length) list.innerHTML = `<p class="empty-state">${items.length ? '没有匹配的结果，请调整搜索条件。' : '当前版本暂无数据。'}</p>`;
    const url = new URL(location.href);
    if (currentPage === 1) url.searchParams.delete('page');
    else url.searchParams.set('page', currentPage);
    history.replaceState(null, '', url);
    const pageNumbers = new Set([1, pageCount]);
    for (let index = Math.max(1, currentPage - 1); index <= Math.min(pageCount, currentPage + 1); index++) pageNumbers.add(index);
    let lastPage = 0;
    const pages = [...pageNumbers].sort((a, b) => a - b).map((index) => {
      const gap = lastPage && index > lastPage + 1 ? '<span class="pagination-gap" aria-hidden="true">…</span>' : '';
      lastPage = index;
      return `${gap}<button type="button" data-page-number="${index}" aria-label="第 ${index} 页"${index === currentPage ? ' aria-current="page"' : ''}>${index}</button>`;
    }).join('');
    pagers.forEach((pager, index) => {
      pager.hidden = !visible.length;
      pager.innerHTML = `<label class="page-size" for="page-size-${index}">每页 <select id="page-size-${index}">${[20, 50, 100].map((size) => `<option value="${size}"${size === pageSize ? ' selected' : ''}>${size} 条</option>`).join('')}</select></label><div class="page-buttons"><button type="button" data-page-number="${currentPage - 1}" aria-label="上一页" title="上一页"${currentPage === 1 ? ' disabled' : ''}><img src="/icons/chevron-left.svg" alt="" width="18" height="18"></button>${pages}<button type="button" data-page-number="${currentPage + 1}" aria-label="下一页" title="下一页"${currentPage === pageCount ? ' disabled' : ''}><img src="/icons/chevron-right.svg" alt="" width="18" height="18"></button></div><span class="page-summary">${currentPage} / ${pageCount} 页</span>`;
    });
    if (scroll) pagers[0].scrollIntoView({ block: 'start' });
  }

  async function loadItems() {
    if (loading) return;
    loading = true;
    retry.hidden = true;
    status.textContent = '加载中…';
    try {
      const records = [];
      let patch = '';
      // Load this small versioned catalog once so filtering and sorting apply
      // to every record, including records beyond the current display page.
      while (true) {
        const params = new URLSearchParams({ limit: '100', offset: String(records.length) });
        if (patch) params.set('patch', patch);
        const result = await api(`${endpoint}?${params}`);
        patch = result.patch;
        const batch = result.items || [];
        records.push(...batch);
        if (batch.length < 100) break;
      }
      items = records;
      loaded = true;
      render();
    } catch (error) {
      status.textContent = error.message;
      retry.hidden = false;
      retry.textContent = '重试加载';
    } finally {
      loading = false;
    }
  }

  const resetPage = () => {
    currentPage = 1;
    render();
  };
  search.addEventListener('input', resetPage);
  sort?.addEventListener('change', resetPage);
  rarity?.addEventListener('change', resetPage);
  category?.addEventListener('change', resetPage);
  rarityTabs.forEach((tab) => tab.addEventListener('click', () => {
    selectedRarity = tab.dataset.rarity;
    rarityTabs.forEach((button) => button.classList.toggle('active', button === tab));
    resetPage();
  }));
  pagers.forEach((pager) => {
    pager.addEventListener('click', (event) => {
      const button = event.target.closest('button[data-page-number]');
      if (!button || button.disabled) return;
      currentPage = Number(button.dataset.pageNumber);
      render(true);
    });
    pager.addEventListener('change', (event) => {
      if (!event.target.matches('select')) return;
      pageSize = Number(event.target.value);
      currentPage = 1;
      render(true);
    });
  });
  retry.addEventListener('click', loadItems);
  loadItems();
}

async function loadChampion() {
  const detail = $('#detail');
  const id = new URLSearchParams(location.search).get('id');
  if (!id) { detail.textContent = '缺少英雄 ID'; return; }
  try {
    const result = await api(`/api/v1/champions/${encodeURIComponent(id)}`);
    const item = result.item;
    const stats = item.stats;
    document.title = `${item.name_zh} · HEXIS`;
    detail.innerHTML = `<span class="eyebrow">CHAMPION · ${escapeHTML(result.patch)}</span><div class="detail-title">${iconMarkup(item)}<div><h1>${escapeHTML(item.name_zh)}</h1><p class="english-name">${escapeHTML(item.name_en || item.id)}</p></div></div><p class="guide-intro">这局选了 ${escapeHTML(item.name_zh)}？先看强化符文与出装方向，再根据对手阵容调整。</p><div class="guide-grid"><section class="guide-card"><span class="eyebrow">01 · 强化符文</span><h2>优先考虑什么符文</h2><p>当前版本尚未导入该英雄与强化符文的关联数据。可先查看全部符文效果，但这里不会将符文整体胜率当作英雄专属推荐。</p><a href="/augments.html">浏览强化符文 →</a></section><section class="guide-card"><span class="eyebrow">02 · 装备路线</span><h2>先出什么装备</h2><p>当前版本尚未导入该英雄的出装路线和适用场景。装备图鉴可用于查看效果，暂不提供未经验证的出装顺序。</p><a href="/items.html">浏览装备图鉴 →</a></section></div><div class="detail-stats"><div><small>英雄胜率</small><strong class="positive">${percent(stats?.win_rate)}</strong></div><div><small>选取率</small><strong>${percent(stats?.pick_rate)}</strong></div></div><div class="source-block"><h2>数据来源</h2><p>${sourceMarkup(item.source)}</p>${stats ? `<p>统计：${sourceMarkup(stats.source)}</p>` : '<p>当前版本尚无统计数据。</p>'}</div>`;
  } catch (error) {
    detail.textContent = error.status === 404 ? '未找到这个英雄，或暂无已发布版本。' : error.message;
  }
}

loadPatch();
if (['champions', 'augments', 'combos', 'items'].includes(page)) setupList();
if (page === 'champion') loadChampion();
