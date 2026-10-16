App({
  globalData: { apiBase: 'http://127.0.0.1:8182', previewGuides: false },
  onLaunch() {
    this.globalData.previewGuides = wx.getAccountInfoSync().miniProgram.envVersion === 'develop';
  }
});
