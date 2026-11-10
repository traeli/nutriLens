import { assetUrl } from './assets.js'

const ICON_SOURCES = {
  dumpling_bun: assetUrl('/static/route-theme-icons/dumpling_bun.png'),
  noodles: assetUrl('/static/route-theme-icons/noodles.png'),
  rice_meal: assetUrl('/static/route-theme-icons/rice_meal.png'),
  local_dishes: assetUrl('/static/route-theme-icons/local_dishes.png'),
  hotpot: assetUrl('/static/route-theme-icons/hotpot.png'),
  bbq_skewers: assetUrl('/static/route-theme-icons/bbq_skewers.png'),
  street_snack: assetUrl('/static/route-theme-icons/street_snack.png'),
  breakfast: assetUrl('/static/route-theme-icons/breakfast.png'),
  coffee_tea: assetUrl('/static/route-theme-icons/coffee_tea.png'),
  dessert: assetUrl('/static/route-theme-icons/dessert.png'),
  bakery: assetUrl('/static/route-theme-icons/bakery.png'),
  seafood: assetUrl('/static/route-theme-icons/seafood.png'),
  vegetarian: assetUrl('/static/route-theme-icons/vegetarian.png'),
  bar_drinks: assetUrl('/static/route-theme-icons/bar_drinks.png'),
  general: assetUrl('/static/route-theme-icons/general.png'),
}

const ICON_ALIASES = {
  dumpling: 'dumpling_bun',
  jiaozi: 'dumpling_bun',
  shengjian: 'dumpling_bun',
  noodle: 'noodles',
  soy_milk: 'breakfast',
  douhua: 'breakfast',
  bowl: 'general',
  coffee: 'coffee_tea',
  dumpling_bun: 'dumpling_bun',
  noodles: 'noodles',
  rice_meal: 'rice_meal',
  local_dishes: 'local_dishes',
  hotpot: 'hotpot',
  bbq_skewers: 'bbq_skewers',
  street_snack: 'street_snack',
  breakfast: 'breakfast',
  coffee_tea: 'coffee_tea',
  dessert: 'dessert',
  bakery: 'bakery',
  seafood: 'seafood',
  vegetarian: 'vegetarian',
  bar_drinks: 'bar_drinks',
  general: 'bowl',
}

const THEME_ACCENTS = {
  dumpling_bun: '#F2C94C', noodles: '#EE7D52', rice_meal: '#E5B85C', local_dishes: '#D96C4A',
  hotpot: '#E55039', bbq_skewers: '#C85B3C', street_snack: '#F1A94E', breakfast: '#F2C94C',
  coffee_tea: '#A77955', dessert: '#D98FA6', bakery: '#D9A25F', seafood: '#5F9DB8',
  vegetarian: '#78A86B', bar_drinks: '#8067A8', general: '#D6B86B',
}

// 接口只传递稳定的语义键；具体图片版本由小程序控制。
export function foodIconSource(iconKey) {
  const normalized = ICON_ALIASES[iconKey] || iconKey
  return ICON_SOURCES[normalized] || ICON_SOURCES.general
}

export function foodThemeAccent(themeCode) {
  return THEME_ACCENTS[themeCode] || THEME_ACCENTS.general
}
