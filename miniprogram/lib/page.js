const core = require('./core');
const dataCache = require('./data-cache');
const tabModes = core.tabs.map(tab => tab.id);

function openPage(url) {
  const fail = () => wx.showToast({ title: '页面打开失败，请重试', icon: 'none' });
  const replace = () => wx.redirectTo({ url, fail });
  if (getCurrentPages().length >= 10) replace();
  else wx.navigateTo({ url, fail: error => /limit|webview count/i.test(error.errMsg || '') ? replace() : fail() });
}

module.exports = function createPage(mode) {
  return {
    data: { mode, loading: true },
    onLoad(options) {
      const { apiBase, previewGuides } = getApp().globalData;
      this._request = dataCache.create(JSON.stringify([apiBase, previewGuides === true]), path => new Promise((resolve, reject) => wx.request({
        url: apiBase + path,
        timeout: 20000,
        success: response => response.statusCode === 200 ? resolve(response.data) : reject(new Error(response.statusCode === 404 ? '暂无对应版本资料' : '资料加载失败，请重试')),
        fail: () => reject(new Error('网络连接失败，请重试'))
      })));
      this._app = core.create({
        previewGuides: previewGuides === true,
        read: key => wx.getStorageSync(key),
        write: (key, value) => wx.setStorageSync(key, value),
        notice: title => wx.showToast({ title, icon: 'none' }),
        render: view => { this.setData(view); wx.setNavigationBarTitle({ title: view.title || '海克斯手册' }); },
        request: this._request
      });
      this._app.enter(mode, options);
      this._app.load();
    },
    onShow() {
      this._app?.refreshSaved();
      if (this._app && !this.data.loading && this._request.isStale()) this._app.load();
      const pending = getApp().globalData.pendingQuery;
      if (pending?.mode === mode && this._app) {
        this._app.change('query', pending.query);
        getApp().globalData.pendingQuery = null;
      }
    },
    onPullDownRefresh() { this._request.invalidate(); this._app.load().finally(() => wx.stopPullDownRefresh()); },
    onShareAppMessage() {
      const { mode, kind, id, detail } = this.data;
      return { title: detail ? `${detail.name} · 海克斯手册` : '海克斯大乱斗 · 英雄与玩法手册', path: mode === 'detail' ? `/pages/detail/index?kind=${kind}&id=${encodeURIComponent(id)}` : `/pages/${mode}/index` };
    },
    input(event) { this._app.change('query', event.detail.value); },
    choosePlaystyle(event) {
      const option = this.data.playstyles[Number(event.detail.value)];
      if (option) this._app.change('playstyle', option.id);
    },
    confirmSearch() { this._app.rememberSearch(); },
    action(event) {
      const d = event.currentTarget.dataset;
      if (d.action === 'open') {
        if (d.missing) return wx.showToast({ title: '当前版本未收录', icon: 'none' });
        openPage(`/pages/detail/index?kind=${d.kind}&id=${encodeURIComponent(d.id)}${d.playstyle ? '&playstyle=' + encodeURIComponent(d.playstyle) : ''}`);
      } else if (d.action === 'navigate') {
        const path = `/pages/${d.mode}/index`;
        if (tabModes.includes(d.mode)) {
          getApp().globalData.pendingQuery = d.query ? { mode: d.mode, query: d.query } : null;
          wx.switchTab({ url: path });
        } else openPage(path + (d.query ? `?query=${encodeURIComponent(d.query)}` : ''));
      } else if (d.action === 'change') {
        this._app.change(d.field, d.value);
        if (d.field === 'page') wx.pageScrollTo({ scrollTop: 0, duration: 150 });
      }
      else if (d.action === 'favorite') this._app.toggleFavorite(d.kind, d.id);
      else if (d.action === 'skills') this._app.skills();
      else if (d.action === 'skillsRetry') this._app.skills().then(() => this._app.skills());
      else if (d.action === 'retry') { this._request.invalidate(); this._app.load(); }
      else if (d.action === 'retryGuides') this._app.retryGuides();
      else if (d.action === 'clear') wx.showModal({ title: '清除最近查看？', content: '收藏内容会保留。', success: result => { if (result.confirm) this._app.clearRecent(); } });
      else if (d.action === 'copySource' && this.data.source.url) wx.setClipboardData({ data: this.data.source.url });
      else if (d.action === 'copyReference' && /^https?:\/\//.test(d.url || '')) wx.setClipboardData({ data: d.url });
      else if (d.action === 'settings') wx.showModal({ title: '资料与版本', content: `当前版本 ${this.data.patch || '未加载'}。图鉴来自 Riot Data Dragon 与 CommunityDragon。${this.data.guideStatusText}收藏与查看记录保存在当前设备。`, showCancel: false });
    }
  };
};
