export function getPrimaryNavLayout(extraGap = 14) {
  const info = typeof uni.getWindowInfo === 'function' ? uni.getWindowInfo() : uni.getSystemInfoSync()
  const statusBarHeight = Number(info.statusBarHeight || 20)
  let menu = null
  try { menu = uni.getMenuButtonBoundingClientRect() } catch (error) { menu = null }

  const top = menu && Number(menu.top) > 0 ? Number(menu.top) : statusBarHeight + 6
  const height = menu && Number(menu.height) > 0 ? Number(menu.height) : 32
  const bottom = menu && Number(menu.bottom) > top ? Number(menu.bottom) : top + height
  const pageGutter = Number(info.windowWidth || 375) * 26 / 750
  const right = menu && Number(menu.left) > 0
    ? Math.max(72, Number(info.windowWidth) - Number(menu.left) - pageGutter + 8)
    : 8

  return {
    top,
    height,
    right,
    totalHeight: bottom + 12,
    contentTop: Math.max(statusBarHeight, bottom) + extraGap,
  }
}
