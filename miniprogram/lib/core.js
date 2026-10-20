(function (root, factory) {
  if (typeof module === 'object' && module.exports) module.exports = factory();
  else root.HexisCore = factory();
})(typeof globalThis !== 'undefined' ? globalThis : this, function () {
  const titles = { home: '海克斯手册', champions: '英雄', augments: '海克斯', items: '装备', mine: '我的', search: '搜索', guides: '玩法手册' };
  const rarityNames = { Prismatic: '棱彩', Gold: '黄金', Silver: '白银' };
  const aliases = { TahmKench: '塔姆 蛤蟆', Graves: '格雷福斯 男枪', Jinx: '金克丝', Ahri: '阿狸', Garen: '盖伦', Lucian: '卢锡安', Vayne: '薇恩', Lux: '拉克丝', MissFortune: '厄运小姐 女枪', MasterYi: '易 剑圣', MonkeyKing: '孙悟空 猴子', Sett: '瑟提', Caitlyn: '凯特琳 女警', Yasuo: '亚索', Yone: '永恩', Aatrox: '亚托克斯 剑魔', LeeSin: '李青 盲僧', Ezreal: '伊泽瑞尔 EZ', Malphite: '墨菲特 石头人' };
  const tabs = [{ id: 'home', title: '手册', icon: 'book-open' }, { id: 'champions', title: '英雄', icon: 'users' }, { id: 'augments', title: '海克斯', icon: 'sparkles' }, { id: 'items', title: '装备', icon: 'backpack' }, { id: 'mine', title: '我的', icon: 'bookmark' }];
  const priorities = { core: '优先考虑', alternative: '备选联动', situational: '按局势选择' };
  const kinds = { champions: '英雄', augments: '海克斯', items: '装备', guides: '玩法' };
  const storageKey = 'hexis.handbook.v1';
  const searchText = row => [row.name_zh, row.name_en, row.title, row.summary, row.description, row.id, aliases[row.id], ...(row.tags || [])].join(' ').toLowerCase();
  const compareNames = (a, b) => (a.name_zh || a.title).localeCompare(b.name_zh || b.title, 'zh-CN');
  const basicHeading = (sections, hasEquipment) => {
    if (!sections.length) return hasEquipment ? '装备参考' : '版本提示';
    const mechanics = sections.filter(section => /^(联动|避坑)[:：]/.test(section.title));
    return mechanics.length ? (mechanics.length === sections.length ? '机制与避坑' : '基础思路与机制') : '基础思路';
  };
  const guideSections = row => [...(row.sections || []), ...(row.basic_sections || [])]
    .filter((section, index, all) => all.findIndex(other => other.title === section.title && other.body === section.body) === index);

  function create(adapter) {
    const imageURL = url => /^https:\/\/(ddragon\.leagueoflegends\.com|raw\.communitydragon\.org)\//.test(url || '') ? url : '';
    const saved = adapter.read(storageKey) || {};
    const state = { mode: 'home', kind: '', id: '', query: '', filter: '', linkedOnly: '', examplesExpanded: '', sort: 'name', page: 1, tab: 'fun', playstyle: '', patch: '', loading: true, error: '', skills: null, skillsOpen: false, skillsLoading: false, skillsError: '', guideError: '', guidesLoading: false };
    let catalogs = { champions: [], augments: [], items: [], guides: [] };
    let favorites = Array.isArray(saved.favorites) ? saved.favorites.slice(0, 200) : [];
    let recent = Array.isArray(saved.recent) ? saved.recent.slice(0, 20) : [];
    let searches = Array.isArray(saved.searches) ? saved.searches.slice(0, 8) : [];
    let generation = 0;
    let skillGeneration = 0;
    const key = (kind, id) => `${kind}:${id}`;
    const isSaved = (kind, id) => favorites.some(row => row.key === key(kind, id));
    const persist = () => adapter.write(storageKey, { favorites, recent, searches });
    const find = (kind, id) => (catalogs[kind] || []).find(row => row.id === id);
    function guideName(row) {
      if (row.category !== 'basic') return row.title;
      const heroes = (row.champion_ids || []).map(id => find('champions', id)?.name_zh).filter(Boolean);
      return heroes.length ? `${heroes.join('、')}：${basicHeading(guideSections(row).filter(section => section.title !== '版本提示'), row.stages?.length)}` : row.title;
    }
    function decorate(kind, row) {
      if (!row) return null;
      const linked = kind === 'augments' ? augmentGuides(row.id) : [];
      const linkedHeroCount = new Set(linked.flatMap(g => g.champion_ids || [])).size;
      return Object.assign({}, row, { kind, name: kind === 'guides' ? guideName(row) : row.name_zh || row.title, rarityName: rarityNames[row.rarity] || '', rarityClass: row.rarity || '', saved: isSaved(kind, row.id), linkageLabel: linkedHeroCount ? `${linkedHeroCount} 位英雄 · ${linked.length} 条玩法` : '', meta: kind === 'champions' ? (aliases[row.id] || row.name_en || row.id) : kind === 'augments' ? (rarityNames[row.rarity] || '') : kind === 'items' ? `${(row.categories || []).join(' / ')} · ${row.price} 金币` : `${({ fun: '趣味玩法', basic: '基础思路', general: '通用玩法' })[row.category] || '编辑玩法'}${row.reviewed === false ? ' · 样例待核对' : ''}`, icon_url: imageURL(row.icon_url || (row.champion_ids?.length ? find('champions', row.champion_ids[0])?.icon_url : '') || '') });
    }
    function guideView(row) {
      const sections = guideSections(row);
      const readingSections = sections.filter(section => section.title !== '版本提示');
      const readingHeading = row.category === 'basic' ? basicHeading(readingSections, row.stages?.length) : '怎么玩';
      const heroes = (row.champion_ids || []).map(id => decorate('champions', find('champions', id))).filter(Boolean);
      const augmentIDs = [...new Set([...(row.core_augment_ids || []), ...(row.alternative_augment_ids || []), ...(row.recommendations || []).map(rec => rec.id)])];
      return Object.assign(decorate('guides', row), {
        heroes, readingHeading,
        introSummary: readingSections.some(section => section.body.trim() === (row.summary || '').trim()) ? '' : row.summary,
        core: (row.core_augment_ids || []).map(id => decorate('augments', find('augments', id))).filter(Boolean),
        alternatives: (row.alternative_augment_ids || []).map(id => decorate('augments', find('augments', id))).filter(Boolean),
        tags: row.tags || [],
        sections: row.sections || [],
        readingSections,
        versionNotes: sections.filter(section => section.title === '版本提示'),
        augmentChoices: augmentIDs.map(id => {
          const augment = decorate('augments', find('augments', id));
          if (!augment) return null;
          const rec = (row.recommendations || []).find(rec => rec.id === id);
          const priority = rec?.priority || ((row.core_augment_ids || []).includes(id) ? 'core' : 'alternative');
          const reason = rec?.reason || '';
          return Object.assign(augment, { reason, displayReason: row.category === 'basic' && readingSections.some(section => section.body.trim() === reason.trim()) ? '' : reason, priority: priorities[priority] });
        }).filter(Boolean),
        basicSections: row.basic_sections?.length ? row.basic_sections : row.sections || [],
        references: row.references || [],
        playstyleName: row.playstyle_name || row.title,
        stages: (row.stages || []).map(stage => Object.assign({}, stage, { items: (stage.item_ids || []).map(id => decorate('items', find('items', id))).filter(Boolean) })),
        recommendations: (row.recommendations || []).map(rec => Object.assign(decorate('augments', find('augments', rec.id)) || {}, { reason: rec.reason, priority: `${priorities[rec.priority]}${row.reviewed === false ? ' · 样例待核对' : ''}` })).filter(rec => rec.id)
      });
    }
    const heroGuides = id => catalogs.guides.filter(g => (g.champion_ids || []).includes(id));
    const augmentGuides = id => catalogs.guides.filter(g => [...(g.core_augment_ids || []), ...(g.alternative_augment_ids || []), ...(g.recommendations || []).map(r => r.id)].includes(id));
    function matchesGuide(row, query) {
      const heroes = (row.champion_ids || []).map(id => find('champions', id)).filter(Boolean);
      const augments = [...(row.core_augment_ids || []), ...(row.alternative_augment_ids || []), ...(row.recommendations || []).map(rec => rec.id)].map(id => find('augments', id)).filter(Boolean);
      return [searchText(row), row.playstyle_name, ...heroes.map(searchText), ...augments.map(augment => augment.name_zh)].join(' ').toLowerCase().includes(query.trim().toLowerCase());
    }
    function matchesCatalog(kind, row, query) {
      return kind === 'guides' ? matchesGuide(row, query) : searchText(row).includes(query.trim().toLowerCase());
    }
    function augmentRelation(g, augmentID) {
      const rec = (g.recommendations || []).find(row => row.id === augmentID);
      const priority = rec?.priority || ((g.core_augment_ids || []).includes(augmentID) ? 'core' : 'alternative');
      return { guideID: g.id, guideTitle: g.title, playstyle: g.playstyle_name || g.title, priority,
        priorityText: `${priorities[priority]}${g.reviewed === false ? ' · 样例待核对' : ''}`, reason: rec?.reason || g.summary };
    }
    function heroExamples(guides, augmentID) {
      const rows = new Map();
      for (const g of guides) {
        const relation = augmentRelation(g, augmentID);
        for (const id of g.champion_ids || []) {
          const hero = decorate('champions', find('champions', id));
          if (!hero) continue;
          if (!rows.has(id)) rows.set(id, Object.assign(hero, { relations: [] }));
          rows.get(id).relations.push(relation);
        }
      }
      const order = { core: 0, alternative: 1, situational: 2 };
      return [...rows.values()].map(hero => {
        hero.relations.sort((a, b) => order[a.priority] - order[b.priority] || a.guideTitle.localeCompare(b.guideTitle, 'zh-CN'));
        return Object.assign(hero, { playstyle: hero.relations[0].guideID });
      }).sort((a, b) => order[a.relations[0].priority] - order[b.relations[0].priority] || a.name.localeCompare(b.name, 'zh-CN'));
    }
    function recommendationGroup(g) {
      const row = guideView(g);
      row.recommendations = row.recommendations.filter(rec => !state.filter || rec.rarity === state.filter);
      return row;
    }
    function sourceView(source) {
      return { name: source?.name || '暂无来源', url: source?.url || '', attribution: source?.attribution || '' };
    }
    function entries(records) {
      return records.map(record => Object.assign(decorate(record.kind, find(record.kind, record.id)) || { kind: record.kind, id: record.id, name: record.name, icon_url: '', missing: true, meta: '当前版本未收录', saved: isSaved(record.kind, record.id) }, { key: record.key }));
    }
    function view() {
      const row = find(state.kind, state.id);
      const detail = state.kind === 'guides' && row ? guideView(row) : decorate(state.kind, row);
      const guideDetail = state.kind === 'guides' ? detail : null;
      let rows = [];
      if (['champions', 'augments', 'items', 'guides'].includes(state.mode)) {
        rows = catalogs[state.mode].filter(row => (state.mode !== 'guides' || row.category !== 'basic') && matchesCatalog(state.mode, row, state.query) && (!state.filter || (state.mode === 'augments' ? row.rarity === state.filter : state.mode === 'items' ? row.categories?.includes(state.filter) : true)) && (state.mode !== 'augments' || !state.linkedOnly || augmentGuides(row.id).length));
        rows = rows.slice().sort(state.sort === 'price' ? (a, b) => a.price - b.price : compareNames);
      }
      const pageCount = Math.max(1, Math.ceil(rows.length / 24));
      state.page = Math.min(state.page, pageCount);
      const heroGuideRows = detail && state.kind === 'champions' ? heroGuides(detail.id) : [];
      const playstyleRows = heroGuideRows.filter(row => row.category === 'fun');
      const activePlaystyle = playstyleRows.some(row => row.id === state.playstyle) ? state.playstyle : '';
      const selectedPlaystyles = playstyleRows.filter(row => !activePlaystyle || row.id === activePlaystyle);
      const commonBasics = heroGuideRows.filter(row => row.category === 'basic');
      const commonBasicViews = commonBasics.map(guideView);
      const recommendationGroups = selectedPlaystyles.map(recommendationGroup).filter(row => row.recommendations.length);
      const commonRecommendationGroups = commonBasics.map(recommendationGroup).filter(row => row.recommendations.length);
      const playstyles = [{ id: '', name: '全部流派' }, ...playstyleRows.map(row => ({ id: row.id, name: row.playstyle_name || row.title }))];
      const playstyleIndex = Math.max(0, playstyles.findIndex(row => row.id === activePlaystyle));
      const related = detail && state.kind === 'augments' ? augmentGuides(detail.id) : [];
      const relatedHeroes = heroExamples(related, detail?.id);
      const relatedItems = [...new Set(related.flatMap(g => (g.stages || []).flatMap(s => s.item_ids)))].map(id => decorate('items', find('items', id))).filter(Boolean);
      const groups = state.mode === 'search' && state.query.trim() ? Object.keys(catalogs).map(kind => {
        const matched = catalogs[kind].filter(row => matchesCatalog(kind, row, state.query)).sort(compareNames);
        return { kind, title: kinds[kind], count: matched.length, items: matched.slice(0, 20).map(row => decorate(kind, row)) };
      }).filter(group => group.count) : [];
      return Object.assign({}, state, {
        guideStatusText: adapter.previewGuides ? '当前为玩法预览模式，包含尚未审核的资料，机制与版本仍待核对。' : '玩法仅展示已审核的编辑内容。',
        title: state.mode === 'detail' ? detail?.name || '详情' : titles[state.mode], tabs, detail, guideDetail,
        counts: { champions: catalogs.champions.length, augments: catalogs.augments.length, items: catalogs.items.length, guides: catalogs.guides.length },
        rows: rows.slice((state.page - 1) * 24, state.page * 24).map(row => state.mode === 'guides' ? guideView(row) : decorate(state.mode, row)), total: rows.length, pageCount,
        heroes: ['TahmKench', 'Graves', 'Jinx', 'Ahri', 'Garen', 'Lucian', 'Vayne', 'Lux'].map(id => decorate('champions', find('champions', id))).filter(Boolean),
        featured: catalogs.guides.filter(row => row.category !== 'basic').slice(0, 3).map(guideView), general: catalogs.guides.filter(row => row.category === 'general').map(guideView),
        recent: entries(recent).slice(0, 6), favorites: entries(favorites), favoritesCount: favorites.length, searches,
        groups, resultCount: groups.reduce((sum, group) => sum + group.count, 0),
        playstyles, playstyle: activePlaystyle, playstyleIndex, playstyleName: playstyles[playstyleIndex].name,
        heroGuides: selectedPlaystyles.map(guideView),
        heroGuideHeading: commonBasics.length ? basicHeading(commonBasicViews.flatMap(row => row.readingSections), commonBasicViews.some(row => row.stages.length)) : '英雄玩法',
        heroGuideNotice: playstyleRows.length ? '' : commonBasics.length
          ? (adapter.previewGuides ? '独立玩法待补充' : '暂无已审核的独立玩法')
          : (adapter.previewGuides ? '该英雄的玩法与基础思路待补充' : '该英雄暂无已审核的玩法与基础思路'),
        commonBasics: commonBasicViews, playstyleBasics: selectedPlaystyles.map(guideView),
        recommendationGroups, commonRecommendationGroups,
        recommendations: [...recommendationGroups, ...commonRecommendationGroups].flatMap(row => row.recommendations),
        linkedAugmentCount: catalogs.augments.filter(row => augmentGuides(row.id).length).length,
        relatedHeroes, relatedExamples: state.examplesExpanded ? relatedHeroes : relatedHeroes.slice(0, 3),
        moreRelatedExamples: !state.examplesExpanded && relatedHeroes.length > 3,
        relatedItems, relatedGuides: related.map(g => Object.assign(guideView(g), { augmentRelation: augmentRelation(g, detail.id) })),
        recipe: state.kind === 'items' && detail ? (detail.recipe || []).map(id => decorate('items', find('items', id)) || { id, kind: 'items', name: `装备 ${id}`, meta: '当前模式未收录', missing: true }).filter(Boolean) : [],
        source: sourceView(detail?.source)
      });
    }
    function emit() { adapter.render(view()); }
    async function catalog(kind, patch) {
      const rows = [];
      for (;;) {
        const body = await adapter.request(`/api/v1/${kind}?patch=${encodeURIComponent(patch)}&limit=100&offset=${rows.length}${kind === 'guides' && adapter.previewGuides ? '&preview=1' : ''}`);
        if (body.patch !== patch || !Array.isArray(body.items)) throw new Error('版本资料不一致，请重试');
        rows.push(...body.items);
        if (body.items.length < 100) return rows;
        if (rows.length > 10000) throw new Error('资料数量异常');
      }
    }
    async function loadGuides(patch, run) {
      state.guidesLoading = true;
      state.guideError = '';
      emit();
      try { const rows = await catalog('guides', patch); if (run === generation) catalogs.guides = rows; }
      catch (_) { if (run === generation) state.guideError = '玩法资料暂时无法加载'; }
      finally { if (run === generation) { state.guidesLoading = false; emit(); } }
    }
    async function load() {
      const run = ++generation;
      state.loading = true;
      state.error = '';
      emit();
      try {
        const patch = await adapter.request('/api/v1/patches/latest');
        const results = await Promise.all(['champions', 'augments', 'items'].map(kind => catalog(kind, patch.version)));
        if (run !== generation) return;
        catalogs = { champions: results[0], augments: results[1], items: results[2], guides: [] };
        state.patch = patch.version;
        rememberEntry();
        state.loading = false;
        emit();
        await loadGuides(patch.version, run);
      } catch (error) {
        if (run !== generation) return;
        state.error = error.message || '资料加载失败，请重试';
        state.loading = false;
        emit();
      }
    }
    function rememberEntry() {
      if (state.mode !== 'detail') return;
      const row = decorate(state.kind, find(state.kind, state.id));
      if (!row) return;
      recent = [{ key: key(row.kind, row.id), kind: row.kind, id: row.id, name: row.name, icon_url: row.icon_url }, ...recent.filter(r => r.key !== key(row.kind, row.id))].slice(0, 20);
      persist();
    }
    function enter(mode, options) {
      options = options || {};
      ++skillGeneration;
      Object.assign(state, { mode, kind: options.kind || '', id: options.id || '', query: options.query || '', filter: '', linkedOnly: options.linkedOnly === '1' ? '1' : '', examplesExpanded: '', sourceOpen: '', basicSourceOpen: '', sort: 'name', page: 1, tab: 'fun', playstyle: options.playstyle || '', skills: null, skillsOpen: false, skillsLoading: false, skillsError: '' });
      rememberEntry();
      emit();
    }
    async function skills() {
      state.skillsOpen = !state.skillsOpen;
      emit();
      if (!state.skillsOpen || state.skills || state.skillsLoading) return;
      const run = ++skillGeneration;
      const id = state.id;
      state.skillsLoading = true;
      state.skillsError = '';
      emit();
      try {
        const body = await adapter.request(`/api/v1/champions/${encodeURIComponent(id)}/skills?patch=${encodeURIComponent(state.patch)}`);
        if (run === skillGeneration) state.skills = Object.assign({}, body.item, { roleText: (body.item.roles || []).join(' / '), skills: (body.item.skills || []).map(skill => Object.assign({},skill,{icon_url:imageURL(skill.icon_url)})) });
      } catch (_) { if (run === skillGeneration) state.skillsError = '技能资料暂时无法加载'; }
      finally { if (run === skillGeneration) { state.skillsLoading = false; emit(); } }
    }
    function toggleFavorite(kind, id) {
      if (isSaved(kind, id)) favorites = favorites.filter(r => r.key !== key(kind, id));
      else {
        const row = decorate(kind, find(kind, id));
        if (!row) return;
        if (favorites.length >= 200) { adapter.notice('最多收藏 200 条'); return; }
        favorites.unshift({ key: key(kind, id), kind, id, name: row.name, icon_url: row.icon_url });
      }
      persist(); emit();
    }
    return {
      load, enter, view, skills, toggleFavorite,
      change(field, value) {
        if (!['query', 'filter', 'sort', 'tab', 'page', 'playstyle', 'linkedOnly', 'examplesExpanded', 'sourceOpen', 'basicSourceOpen'].includes(field)) return;
        state[field] = field === 'page' ? Math.max(1, Number(value) || 1) : value;
        if (field !== 'page') state.page = 1;
        if (field === 'tab' || field === 'playstyle') state.filter = '';
        emit();
      },
      rememberSearch() {
        const q = state.query.trim();
        if (q) { searches = [q, ...searches.filter(s => s !== q)].slice(0, 8); persist(); emit(); }
      },
      retryGuides: () => state.patch ? loadGuides(state.patch, generation) : load(),
      clearRecent() { recent = []; searches = []; persist(); emit(); },
      refreshSaved() { const s = adapter.read(storageKey) || {}; favorites = s.favorites || []; recent = s.recent || []; searches = s.searches || []; emit(); }
    };
  }
  return { create, tabs, rarityNames };
});
