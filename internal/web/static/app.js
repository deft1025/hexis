const page = document.body.dataset.page;
const $ = (selector) => document.querySelector(selector);
const escapeHTML = (value) => String(value ?? '').replace(/[&<>"']/g, (character) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[character]);
const percent = (value) => value == null ? '暂无数据' : `${(value * 100).toFixed(1)}%`;
const number = (value) => Number(value || 0).toLocaleString('zh-CN');
const rarityNames = { Silver: '白银', Gold: '黄金', Prismatic: '棱彩' };
const augmentIcons = {
  '珠光护手': 'jeweledgauntlet', '法术回响': 'spellwake', '巨人杀手': 'giantslayer',
  '不可触碰': 'canttouchthis', '灵能使者': 'mindtomatter', '大法师': 'bigbrain',
  '利刃华尔兹': 'bladewaltz', '极速': 'withhaste', '缩小术': 'eureka', '坚韧不倒': 'outlawsgrit'
};
const itemCatalog = [
  ['3031', '无尽之刃', '暴击伤害', '攻击', 3450], ['3078', '三相之力', '全面提升攻击能力', '攻击', 3333],
  ['3094', '疾射火炮', '充能攻击与攻击距离', '攻击', 2650], ['3142', '幽梦之灵', '主动提升移动速度', '攻击', 2800],
  ['3161', '朔极之矛', '提升技能伤害', '攻击', 3100], ['3046', '幻影之舞', '攻击速度与移动速度', '攻击', 2650],
  ['3508', '夺萃之镰', '暴击与技能急速', '攻击', 2900], ['3089', '灭世者的死亡之帽', '巨幅提升法术强度', '法术', 3500],
  ['3157', '中娅沙漏', '主动进入凝滞状态', '法术', 3250], ['3135', '虚空之杖', '提升魔法穿透', '法术', 3000],
  ['3116', '瑞莱的冰晶节杖', '技能使敌人减速', '法术', 2600], ['3102', '女妖面纱', '周期性阻挡敌方技能', '法术', 3000],
  ['3083', '狂徒铠甲', '大量生命值与生命回复', '防御', 3100], ['3026', '守护天使', '阵亡后复活', '防御', 3200],
  ['3075', '荆棘之甲', '护甲与重伤效果', '防御', 2450], ['3084', '心之钢', '持续提升生命值', '防御', 3000],
  ['3190', '钢铁烈阳之匣', '为队友提供护盾', '功能', 2200], ['2065', '舒瑞娅的战歌', '为队友提供移动速度', '功能', 2200]
].map(([id, name_zh, description, category, price]) => ({ id, name_zh, description, category, price, icon_url: `https://ddragon.leagueoflegends.com/cdn/15.21.1/img/item/${id}.png`, stats: null }));

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
  if (!iconURL && ['champions', 'champion'].includes(page)) iconURL = `https://ddragon.leagueoflegends.com/cdn/15.21.1/img/champion/${encodeURIComponent(item.id)}.png`;
  if (!iconURL && page === 'augments' && augmentIcons[item.name_zh]) iconURL = `https://raw.communitydragon.org/latest/game/assets/ux/cherry/augments/icons/${augmentIcons[item.name_zh]}_small.png`;
  if (!iconURL) return '<span class="entity-icon icon-fallback" aria-hidden="true">✦</span>';
  try {
    const url = new URL(iconURL);
    if (!['http:', 'https:'].includes(url.protocol)) throw new Error('invalid icon URL');
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
  return `<div class="ranking-row">${rankMarkup(index)}<span class="entity-cell">${iconMarkup(item)}<span><strong>${escapeHTML(item.name_zh)}</strong><small>${escapeHTML(item.description || item.name_en || '暂无效果描述')}</small></span></span><span class="rate-cell">${rateMarkup(stats.win_rate)}</span><span>${percent(stats.pick_rate)}</span><span class="rarity rarity-${escapeHTML(item.rarity || 'unknown')}">${escapeHTML(rarityNames[item.rarity] || item.rarity || '未分级')}</span></div>`;
}

function itemMarkup(item, index) {
  return `<div class="ranking-row">${rankMarkup(index)}<span class="entity-cell">${iconMarkup(item)}<span><strong>${escapeHTML(item.name_zh)}</strong></span></span><span>${escapeHTML(item.category)}</span><span>${number(item.price)}</span><span class="item-effect">${escapeHTML(item.description)}</span></div>`;
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
  const more = $('#load-more');
  const search = $('#search');
  const sort = $('#sort');
  const rarity = $('#rarity');
  const rarityTabs = [...document.querySelectorAll('[data-rarity]')];
  const pageSize = 100;
  let items = [];
  let hasMore = false;
  let loading = false;
  let selectedRarity = '';

  function render() {
    const query = search.value.trim().toLocaleLowerCase();
    let visible = items.filter((item) => {
      if (page === 'combos') return (item.champions || []).some((champion) => `${champion.name_zh} ${champion.id}`.toLocaleLowerCase().includes(query));
      const text = `${item.name_zh} ${item.name_en || ''} ${item.id} ${page === 'augments' ? item.description || '' : ''}`;
      return text.toLocaleLowerCase().includes(query) && (!rarity || !rarity.value || item.rarity === rarity.value) && (!selectedRarity || item.rarity === selectedRarity);
    });
    if (sort?.value === 'win') visible.sort((first, second) => (second.stats?.win_rate ?? second.win_rate ?? -1) - (first.stats?.win_rate ?? first.win_rate ?? -1));
    if (sort?.value === 'pick') visible.sort((first, second) => (second.stats?.pick_rate ?? second.pick_rate ?? -1) - (first.stats?.pick_rate ?? first.pick_rate ?? -1));
    if (sort?.value === 'name') visible.sort((first, second) => first.name_zh.localeCompare(second.name_zh, 'zh-CN'));
    list.innerHTML = visible.map(page === 'champions' ? championMarkup : page === 'augments' ? augmentMarkup : comboMarkup).join('');
    if (page !== 'combos') {
      $('#hero-count').textContent = `${items.length}${hasMore ? '+' : ''} ${page === 'champions' ? '位英雄' : '个符文'}`;
      const leader = items.filter((item) => item.stats?.win_rate != null).sort((first, second) => second.stats.win_rate - first.stats.win_rate)[0];
      $('#leader').innerHTML = leader ? `<small>当前已加载胜率领先</small><strong>${percent(leader.stats.win_rate)}</strong><span>${escapeHTML(leader.name_zh)}</span>` : '<small>当前已加载胜率领先</small><strong>—</strong><span>暂无统计数据</span>';
    }
    if (!loading) status.textContent = items.length ? `显示 ${visible.length} 条 · 已加载 ${items.length} 条${hasMore ? '，可继续加载' : ''}` : '暂无数据';
    if (!visible.length && items.length) list.innerHTML = '<p class="empty-state">没有匹配的结果，请调整搜索条件。</p>';
  }

  async function loadMore() {
    if (loading) return;
    loading = true;
    more.disabled = true;
    status.textContent = '加载中…';
    let failed = false;
    try {
      const result = await api(`${endpoint}?limit=${pageSize}&offset=${items.length}`);
      const batch = result.items || [];
      items = items.concat(batch);
      hasMore = batch.length === pageSize;
      more.hidden = !hasMore;
      render();
    } catch (error) {
      failed = true;
      status.textContent = error.message;
      more.hidden = false;
      more.textContent = '重试加载';
    } finally {
      loading = false;
      more.disabled = false;
      if (!failed) {
        more.textContent = '加载更多';
        render();
      }
    }
  }

  search.addEventListener('input', render);
  sort?.addEventListener('change', render);
  rarity?.addEventListener('change', render);
  rarityTabs.forEach((tab) => tab.addEventListener('click', () => {
    selectedRarity = tab.dataset.rarity;
    rarityTabs.forEach((button) => button.classList.toggle('active', button === tab));
    render();
  }));
  more.addEventListener('click', loadMore);
  loadMore();
}

function setupItems() {
  const list = $('#list');
  const status = $('#list-status');
  const search = $('#search');
  const category = $('#category');
  $('#hero-count').textContent = `${itemCatalog.length} 件装备`;
  function render() {
    const query = search.value.trim().toLocaleLowerCase();
    const visible = itemCatalog.filter((item) => `${item.name_zh} ${item.description} ${item.id}`.toLocaleLowerCase().includes(query) && (!category.value || item.category === category.value));
    list.innerHTML = visible.map(itemMarkup).join('') || '<p class="empty-state">没有匹配的装备。</p>';
    status.textContent = `显示 ${visible.length} 件装备 · 表现统计待接入`;
  }
  search.addEventListener('input', render);
  category.addEventListener('change', render);
  render();
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
if (['champions', 'augments', 'combos'].includes(page)) setupList();
if (page === 'items') setupItems();
if (page === 'champion') loadChampion();
