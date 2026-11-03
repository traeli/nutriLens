const ICON_SOURCES = {
  bowl: '/static/food-icons/bowl.png',
  coffee: '/static/food-icons/coffee.png',
  douhua: '/static/food-icons/douhua.png',
  noodle: '/static/food-icons/noodle.png',
  shengjian: '/static/food-icons/shengjian.png',
}

const ICON_ALIASES = {
  dumpling: 'shengjian',
  jiaozi: 'shengjian',
  noodles: 'noodle',
  soy_milk: 'douhua',
}

// 接口只传递稳定的语义键；具体图片版本由小程序控制。
export function foodIconSource(iconKey) {
  const normalized = ICON_ALIASES[iconKey] || iconKey
  return ICON_SOURCES[normalized] || ICON_SOURCES.bowl
}
