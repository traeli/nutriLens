/**
 * 微信小程序端的装饰图标兼容层。
 *
 * 当前页面保留文字作为完整信息来源。微信端曾出现 SVG data URI + CSS mask
 * 退化为实心方块的问题，因此装饰图标统一隐藏，不再打包未使用的 SVG 源文件。
 */
export function doodleStyle() {
  return { display: 'none' }
}

export function doodleBg() {
  return {}
}
