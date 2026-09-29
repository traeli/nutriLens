export function syncCustomTabBar(vm, selected) {
  const page = vm && (vm.$scope || (vm.$mp && vm.$mp.page))
  if (!page || typeof page.getTabBar !== 'function') return
  const tabBar = page.getTabBar()
  if (tabBar && typeof tabBar.setData === 'function') tabBar.setData({ selected })
}
