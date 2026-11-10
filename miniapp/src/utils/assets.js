const ENV = import.meta.env || {}
const ASSET_BASE_URL = (ENV.VITE_ASSET_BASE_URL || '').trim().replace(/\/+$/, '')

// 仅迁移随包的餐饮插图和海报；用户上传图片、头像和本地 tabBar 图标保持原地址。
export function assetUrl(path) {
  if (typeof path !== 'string' || !ASSET_BASE_URL) return path
  if (!/^\/?static\/(dining|posters|route-theme-icons)\//.test(path)) return path
  return ASSET_BASE_URL + '/' + path.replace(/^\//, '')
}

const IMAGE_FIELDS = new Set(['image', 'image_url', 'background_image', 'public_url', 'theme_icon_url'])

// 后端种子数据仍使用 /static 路径，在前端边界转换，避免迁移数据库中的运营数据。
export function resolveAssetFields(value) {
  if (Array.isArray(value)) return value.map(resolveAssetFields)
  if (!value || typeof value !== 'object') return value
  return Object.fromEntries(Object.entries(value).map(([key, item]) => [
    key,
    IMAGE_FIELDS.has(key) && typeof item === 'string' ? assetUrl(item) : resolveAssetFields(item),
  ]))
}
