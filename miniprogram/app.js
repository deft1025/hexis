App({
  globalData: { apiBase: 'https://hexis.xuninfo.top', previewGuides: false },
  onLaunch() {
    const localPreview = wx.getAccountInfoSync().miniProgram.envVersion === 'develop' &&
      wx.getDeviceInfo().platform === 'devtools';
    if (localPreview) {
      this.globalData.apiBase = 'http://127.0.0.1:8196';
      this.globalData.previewGuides = true;
    }
  }
});
