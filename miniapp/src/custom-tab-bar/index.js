const tabs = [
  { pagePath: '/pages/home/home', text: '发现' },
  { pagePath: '/pages/record/record', text: '记录' },
  { pagePath: '/pages/footprints/footprints', text: '足迹' },
  { pagePath: '/pages/mine/mine', text: '我的' },
]

Component({
  data: { tabs, selected: 0 },
  lifetimes: {
    attached() { this.syncCurrentRoute() },
  },
  pageLifetimes: {
    show() { this.syncCurrentRoute() },
  },
  methods: {
    syncCurrentRoute() {
      const pages = getCurrentPages()
      const route = pages.length ? `/${pages[pages.length - 1].route}` : tabs[0].pagePath
      const index = tabs.findIndex(item => item.pagePath === route)
      if (index >= 0) this.setData({ selected: index })
    },
    switchTab(event) {
      const index = Number(event.currentTarget.dataset.index)
      const item = tabs[index]
      if (!item || index === this.data.selected) return
      this.setData({ selected: index })
      wx.switchTab({
        url: item.pagePath,
        fail: () => this.syncCurrentRoute(),
      })
    },
  },
})
