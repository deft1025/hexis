const core = require('./core');
const tabModes = core.tabs.map(tab => tab.id);

module.exports = function createPage(mode) {
  return {
    data: { mode, loading: true },
    onLoad(options) {
      this._app = core.create({
        previewGuides: getApp().globalData.previewGuides === true,
        read: key => wx.getStorageSync(key),
        write: (key, value) => wx.setStorageSync(key, value),
        notice: title => wx.showToast({ title, icon: 'none' }),
        render: view => { this.setData(view); wx.setNavigationBarTitle({ title: view.title || '海克斯手册' }); },
        request: path => new Promise((resolve, reject) => wx.request({
          url: getApp().globalData.apiBase + path,
          timeout: 20000,
          success: response => response.statusCode === 200 ? resolve(response.data) : reject(new Error(response.statusCode === 404 ? '暂无对应版本资料' : '资料加载失败，请重试')),
          fail: () => reject(new Error('网络连接失败，请重试'))
        }))
      });
      this._app.enter(mode, options);
      this._app.load();
    },
    onShow() {
      this._app?.refreshSaved();
      const pending = getApp().globalData.pendingQuery;
      if (pending?.mode === mode && this._app) {
        this._app.change('query', pending.query);
        getApp().globalData.pendingQuery = null;
      }
    },
    onPullDownRefresh() { this._app.load().finally(() => wx.stopPullDownRefresh()); },
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
        wx.navigateTo({ url: `/pages/detail/index?kind=${d.kind}&id=${encodeURIComponent(d.id)}${d.playstyle ? '&playstyle=' + encodeURIComponent(d.playstyle) : ''}` });
      } else if (d.action === 'navigate') {
        const path = `/pages/${d.mode}/index`;
        if (tabModes.includes(d.mode)) {
          getApp().globalData.pendingQuery = d.query ? { mode: d.mode, query: d.query } : null;
          wx.switchTab({ url: path });
        } else wx.navigateTo({ url: path + (d.query ? `?query=${encodeURIComponent(d.query)}` : '') });
      } else if (d.action === 'change') {
        this._app.change(d.field, d.value);
        if (d.field === 'page') wx.pageScrollTo({ scrollTop: 0, duration: 150 });
      }
      else if (d.action === 'favorite') this._app.toggleFavorite(d.kind, d.id);
      else if (d.action === 'skills') this._app.skills();
      else if (d.action === 'skillsRetry') this._app.skills().then(() => this._app.skills());
      else if (d.action === 'retry') this._app.load();
      else if (d.action === 'retryGuides') this._app.retryGuides();
      else if (d.action === 'clear') wx.showModal({ title: '清除最近查看？', content: '收藏内容会保留。', success: result => { if (result.confirm) this._app.clearRecent(); } });
      else if (d.action === 'copySource' && this.data.source.url) wx.setClipboardData({ data: this.data.source.url });
      else if (d.action === 'copyReference' && /^https?:\/\//.test(d.url || '')) wx.setClipboardData({ data: d.url });
      else if (d.action === 'settings') wx.showModal({ title: '资料与版本', content: `当前版本 ${this.data.patch || '未加载'}。图鉴来自 Riot Data Dragon 与 CommunityDragon；玩法仅展示已审核的编辑内容。收藏保存在当前设备。`, showCancel: false });
    }
  };
};
