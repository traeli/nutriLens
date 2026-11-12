import { resolveAssetFields } from '@/utils/assets.js'

const BASE_URL = (import.meta.env.VITE_BASE_URL || 'http://localhost:8080') + '/api/v1'

let refreshPromise = null
let loginRedirecting = false

export function clearSession() {
  for (const key of ['token', 'refresh_token', 'token_expires_at', 'user_id', 'has_profile', 'nickname']) uni.removeStorageSync(key)
}

function redirectToLogin() {
  clearSession()
  if (loginRedirecting) return
  let currentRoute = ''
  try {
    const pages = typeof getCurrentPages === 'function' ? getCurrentPages() : []
    currentRoute = pages.length ? pages[pages.length - 1].route : ''
  } catch {}
  if (currentRoute === 'pages/login/login') return
  loginRedirecting = true
  uni.reLaunch({
    url: '/pages/login/login',
    complete() { loginRedirecting = false },
  })
}

function saveSession(pair) {
  uni.setStorageSync('token', pair.token)
  uni.setStorageSync('refresh_token', pair.refresh_token)
  uni.setStorageSync('token_expires_at', Date.now() + pair.expires_in * 1000)
}

function logHTTPError(url, options, filePath, statusCode, apiError) {
  console.error('[API] HTTP 请求失败', {
    method: options.method || (filePath ? 'POST' : 'GET'),
    path: url,
    upload: Boolean(filePath),
    statusCode,
    code: (apiError && apiError.code) || '',
    message: (apiError && apiError.message) || '请求失败',
    requestId: (apiError && apiError.request_id) || '',
  })
}

function send(url, options = {}, filePath) {
  const token = uni.getStorageSync('token')
  const header = { ...(filePath ? {} : { 'Content-Type': 'application/json' }), ...options.header }
  if (token && !options.public) header.Authorization = 'Bearer ' + token
  return new Promise((resolve, reject) => {
    const callbacks = {
      success(res) {
        let data = res.data
        if (filePath) { try { data = JSON.parse(data) } catch {} }
        if (res.statusCode >= 200 && res.statusCode < 300) {
          resolve(resolveAssetFields(data))
          return
        }
        const apiError = data && data.error
        logHTTPError(url, options, filePath, res.statusCode, apiError)
        const error = new Error((apiError && (apiError.message || apiError.code)) || '请求失败')
        error.status = res.statusCode
        error.code = (apiError && apiError.code) || ''
        error.requestId = (apiError && apiError.request_id) || ''
        reject(error)
      },
      fail(err) { reject(new Error(err.errMsg || '网络请求失败')) },
    }
    if (filePath) {
      uni.uploadFile({ url: BASE_URL + url, filePath, name: options.name || 'file', formData: options.formData, header, ...callbacks })
    } else {
      uni.request({ url: BASE_URL + url, method: options.method || 'GET', data: options.data, header, ...callbacks })
    }
  })
}

function refreshSession() {
  if (refreshPromise) return refreshPromise
  const previousToken = uni.getStorageSync('token')
  const refreshToken = uni.getStorageSync('refresh_token')
  if (!previousToken || !refreshToken) return Promise.reject(new Error('请重新登录'))
  refreshPromise = send('/auth/refresh', { method: 'POST', public: true, data: { refresh_token: refreshToken } })
    .then(pair => {
      // 请求期间发生退出或重新登录时，不允许旧请求恢复之前的会话。
      if (uni.getStorageSync('token') !== previousToken) throw new Error('登录状态已变更')
      saveSession(pair)
      return pair
    })
    .catch(error => {
      if (error.status === 401 && uni.getStorageSync('token') === previousToken) redirectToLogin()
      throw error
    })
    .finally(() => { refreshPromise = null })
  return refreshPromise
}

async function authenticatedRequest(url, options, filePath) {
  if (options.public) return send(url, options, filePath)
  const expiresAt = Number(uni.getStorageSync('token_expires_at'))
  if (uni.getStorageSync('token') && uni.getStorageSync('refresh_token') && expiresAt && expiresAt - Date.now() < 60000) {
    await refreshSession()
  }
  const previousToken = uni.getStorageSync('token')
  try {
    return await send(url, options, filePath)
  } catch (error) {
    if (error.status !== 401) throw error
    if (!previousToken || !uni.getStorageSync('token')) {
      redirectToLogin()
      throw error
    }
    if (uni.getStorageSync('token') === previousToken) {
      if (!uni.getStorageSync('refresh_token')) { redirectToLogin(); throw error }
      await refreshSession()
    }
    const retryToken = uni.getStorageSync('token')
    try {
      return await send(url, options, filePath)
    } catch (retryError) {
      if (retryError.status === 401 && uni.getStorageSync('token') === retryToken) redirectToLogin()
      throw retryError
    }
  }
}

function request(url, options = {}) { return authenticatedRequest(url, options) }
function uploadFile(url, filePath, options = {}) { return authenticatedRequest(url, options, filePath) }

function buildQuery(params) {
  if (!params) return ''
  const parts = []
  for (const key in params) {
    if (params[key] !== undefined && params[key] !== '') {
      parts.push(encodeURIComponent(key) + '=' + encodeURIComponent(params[key]))
    }
  }
  return parts.length ? '?' + parts.join('&') : ''
}

export const api = {
  // 认证
  wxLogin: async (code, inviterId) => {
    const result = await request('/auth/wx-login', { method: 'POST', public: true, data: { code, inviter_id: inviterId || 0 } })
    saveSession(result)
    return result
  },
  sendEmailCode: email => request('/auth/getemailcode', { method: 'POST', public: true, data: { email } }),
  emailLogin: async (email, emailCode) => {
    const result = await request('/auth/email-login', { method: 'POST', public: true, data: { email, email_code: emailCode } })
    saveSession(result)
    return result
  },
  refreshToken: () => refreshSession(),

  // 用户
  getProfile: () => request('/user/profile'),
  updateProfile: (data) => request('/user/profile', { method: 'PUT', data }),
  deleteAccount: () => request('/user/account', { method: 'DELETE' }),

  // 隐私协议
  agreePrivacy: (data) => request('/agreements/accept', { method: 'POST', data }),

  // 城市餐饮
  getCities: () => request('/cities', { public: true }),
  getCityPosterTheme: code => request('/cities/' + code + '/poster-theme', { public: true }),
  getHomeSummary: (cityCode) => request('/home/summary' + buildQuery({ city_code: cityCode }), { public: true }),
  getRoute: id => request('/routes/' + id, { public: true }),
  getNearbyRoute: params => request('/routes/nearby' + buildQuery(params), { public: true }),
  searchPlaces: params => request('/places/search' + buildQuery(params), { public: true }),
  getPlace: id => request('/places/' + id, { public: true }),
  getPlaceDiscoveryPosts: (id, params) => request('/places/' + id + '/discovery-posts' + buildQuery(params), { public: true }),
  getDiscoveryPosts: params => request('/discovery/posts' + buildQuery(params), { public: true }),
  getDiscoveryPost: id => request('/discovery/posts/' + id, { public: true }),
  getTags: () => request('/tags', { public: true }),
  favoritePlace: id => request(`/places/${id}/favorite`, { method: 'POST' }),
  unfavoritePlace: id => request(`/places/${id}/favorite`, { method: 'DELETE' }),
  getFavoritePlaces: () => request('/me/favorite-places'),
  likeDiscoveryPost: id => request(`/discovery/posts/${id}/like`, { method: 'POST' }),
  unlikeDiscoveryPost: id => request(`/discovery/posts/${id}/like`, { method: 'DELETE' }),
  getMyDiscoveryPostLikes: () => request('/me/discovery-post-likes'),
  createReport: data => request('/reports', { method: 'POST', data }),
  getMyReports: () => request('/me/reports'),
  favoriteRoute: id => request(`/routes/${id}/favorite`, { method: 'POST' }),
  unfavoriteRoute: id => request(`/routes/${id}/favorite`, { method: 'DELETE' }),
  startRoute: id => request(`/routes/${id}/start`, { method: 'POST' }),
  updateRouteJourney: (id, data) => request(`/route-journeys/${id}`, { method: 'PATCH', data }),
  getMyRouteJourneys: () => request('/me/route-journeys'),
}

export default api
