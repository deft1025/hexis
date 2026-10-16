App({
  globalData: { apiBase: 'https://hexis.xuninfo.top', previewGuides: false },
  onLaunch() {
    this.globalData.previewGuides = wx.getAccountInfoSync().miniProgram.envVersion === 'develop';
  }
});
