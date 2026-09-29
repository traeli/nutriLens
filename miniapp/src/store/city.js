import { api } from '@/api/request.js'

const CITY_CACHE_KEY = 'supported_cities'
const SELECTED_CITY_KEY = 'selected_city'
const EMPTY_CITY = { code: '', name: '选择城市', desc: '', image: '', is_default: false }

let pendingRequest = null

function normalizeCity(city) {
  if (!city || !String(city.code || '').trim() || !String(city.name || '').trim()) return null
  return {
    code: String(city.code).trim(),
    name: String(city.name).trim(),
    desc: String(city.desc || '').trim(),
    image: String(city.image || '').trim(),
    is_default: Boolean(city.is_default),
  }
}

export function getCityOptions() {
  const cached = uni.getStorageSync(CITY_CACHE_KEY)
  if (!Array.isArray(cached)) return []
  return cached.map(normalizeCity).filter(Boolean)
}

export function setCityOptions(items) {
  const cities = Array.isArray(items) ? items.map(normalizeCity).filter(Boolean) : []
  uni.setStorageSync(CITY_CACHE_KEY, cities)

  if (!cities.length) return cities
  const saved = uni.getStorageSync(SELECTED_CITY_KEY)
  const savedCode = saved && typeof saved === 'object' ? String(saved.code || '') : ''
  const savedName = typeof saved === 'string' ? saved : ''
  const selected = cities.find(item => item.code === savedCode)
    || cities.find(item => item.name === savedName)
    || cities.find(item => item.is_default)
    || cities[0]
  uni.setStorageSync(SELECTED_CITY_KEY, selected)
  return cities
}

export async function refreshCityOptions() {
  if (pendingRequest) return pendingRequest
  pendingRequest = api.getCities()
    .then(result => setCityOptions(result && result.items))
    .finally(() => { pendingRequest = null })
  return pendingRequest
}

export function getSelectedCity() {
  const cities = getCityOptions()
  const saved = uni.getStorageSync(SELECTED_CITY_KEY)
  if (saved && typeof saved === 'object') {
    const match = cities.find(item => item.code === String(saved.code || ''))
    if (match) return match
    const normalized = normalizeCity(saved)
    if (normalized && !cities.length) return normalized
  }
  if (typeof saved === 'string') {
    const match = cities.find(item => item.name === saved)
    if (match) return match
  }
  return cities.find(item => item.is_default) || cities[0] || { ...EMPTY_CITY }
}

export function setSelectedCity(city) {
  const cities = getCityOptions()
  const selected = typeof city === 'string'
    ? cities.find(item => item.name === city || item.code === city)
    : cities.find(item => item.code === String((city && city.code) || ''))
  if (!selected) return
  uni.setStorageSync(SELECTED_CITY_KEY, selected)
  uni.$emit('city:changed', selected)
}
