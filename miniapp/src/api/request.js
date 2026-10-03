import { resolveAssetFields } from '@/utils/assets.js'

const BASE_URL = (import.meta.env.VITE_BASE_URL || 'http://localhost:8080') + '/api/v1'

let refreshPromise = null

export function clearSession() {
  for (const key of ['token', 'refresh_token', 'token_expires_at']) uni.removeStorageSync(key)
}

function saveSession(pair) {
  uni.setStorageSync('token', pair.token)
  uni.setStorageSync('refresh_token', pair.refresh_token)
  uni.setStorageSync('token_expires_at', Date.now() + pair.expires_in * 1000)
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
        const error = new Error((apiError && (apiError.message || apiError.code)) || '请求失败')
        error.status = res.statusCode
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
      if (error.status === 401 && uni.getStorageSync('token') === previousToken) clearSession()
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
    if (!previousToken || !uni.getStorageSync('token')) throw error
    if (uni.getStorageSync('token') === previousToken) {
      if (!uni.getStorageSync('refresh_token')) { clearSession(); throw error }
      await refreshSession()
    }
    const retryToken = uni.getStorageSync('token')
    try {
      return await send(url, options, filePath)
    } catch (retryError) {
      if (retryError.status === 401 && uni.getStorageSync('token') === retryToken) clearSession()
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
  refreshToken: () => refreshSession(),

  // 用户
  getProfile: () => request('/user/profile'),
  updateProfile: (data) => request('/user/profile', { method: 'PUT', data }),
  deleteAccount: () => request('/user/account', { method: 'DELETE' }),

  // 语音
  transcribeVisitAudio: filePath => uploadFile('/speech/transcribe', filePath, {
    name: 'audio',
    formData: { scene: 'visit_experience' },
  }),

  // 隐私协议
  agreePrivacy: (data) => request('/agreements/accept', { method: 'POST', data }),

  // 城市餐饮
  getCities: () => request('/cities'),
  getCityPosterTheme: code => request('/cities/' + code + '/poster-theme'),
  getHomeSummary: (cityCode) => request('/home/summary' + buildQuery({ city_code: cityCode })),
  getRoute: id => request('/routes/' + id),
  getNearbyRoute: params => request('/routes/nearby' + buildQuery(params)),
  searchPlaces: params => request('/places/search' + buildQuery(params)),
  createPlace: (data) => request('/places', { method: 'POST', data }),
  getPlace: id => request('/places/' + id),
  getPlaceExperiences: (id, params) => request('/places/' + id + '/experiences' + buildQuery(params)),
  getExperiences: params => request('/experiences' + buildQuery(params)),
  getExperience: id => request('/experiences/' + id),
  getTags: () => request('/tags'),
  createVisitRecord: (data, idempotencyKey) => request('/records', { method: 'POST', data, header: idempotencyKey ? { 'Idempotency-Key': idempotencyKey } : {} }),
  getVisitRecord: id => request('/records/' + id),
  updateVisitRecord: (id, data) => request('/records/' + id, { method: 'PATCH', data }),
  deleteVisitRecord: id => request('/records/' + id, { method: 'DELETE' }),
  submitVisitRecord: id => request('/records/' + id + '/submit-public', { method: 'POST' }),
  getVisitRecordReviewStatus: id => request('/records/' + id + '/review-status'),
  getMyRecords: params => request('/me/records' + buildQuery(params)),
  createRestaurantReview: visitId => request(`/visits/${visitId}/review`, { method: 'POST' }),
  getRestaurantReviewStatus: id => request(`/reviews/${id}/status`),
  getMyReviews: params => request('/me/reviews' + buildQuery(params)),
  getFootprintSummary: () => request('/me/footprints/summary'),
  getFootprintMap: () => request('/me/footprints/map'),
  getContribution: () => request('/me/contribution'),
  getBadges: () => request('/me/badges'),
  uploadRecordMedia: (id, filePath, mediaType = 'photo') => uploadFile(`/records/${id}/media`, filePath, { formData: { media_type: mediaType } }),
  uploadEvidence: (id, filePath, evidenceType = 'receipt') => uploadFile(`/records/${id}/evidences`, filePath, { formData: { evidence_type: evidenceType } }),
  deleteRecordMedia: (recordId, mediaId) => request(`/records/${recordId}/media/${mediaId}`, { method: 'DELETE' }),
  getPublisherVerification: () => request('/publisher-verification/status'),
  verifyPublisherPhone: code => request('/publisher-verification/phone', { method: 'POST', data: { code } }),
  favoritePlace: id => request(`/places/${id}/favorite`, { method: 'POST' }),
  unfavoritePlace: id => request(`/places/${id}/favorite`, { method: 'DELETE' }),
  getFavoritePlaces: () => request('/me/favorite-places'),
  helpfulExperience: id => request(`/experiences/${id}/helpful`, { method: 'POST' }),
  unhelpfulExperience: id => request(`/experiences/${id}/helpful`, { method: 'DELETE' }),
  markExperienceOutdated: (id, reason = '') => request(`/experiences/${id}/outdated`, { method: 'POST', data: { reason } }),
  createReport: data => request('/reports', { method: 'POST', data }),
  getMyReports: () => request('/me/reports'),
  createAppeal: (recordId, text) => request(`/records/${recordId}/appeals`, { method: 'POST', data: { text } }),
  getMyAppeals: () => request('/me/appeals'),
  favoriteRoute: id => request(`/routes/${id}/favorite`, { method: 'POST' }),
  unfavoriteRoute: id => request(`/routes/${id}/favorite`, { method: 'DELETE' }),
  startRoute: id => request(`/routes/${id}/start`, { method: 'POST' }),
  updateRouteJourney: (id, data) => request(`/route-journeys/${id}`, { method: 'PATCH', data }),
  getMyRouteJourneys: () => request('/me/route-journeys'),
  createNutritionRecord: data => request('/nutrition/records', { method: 'POST', data }),
  getNutritionRecords: params => request('/nutrition/records' + buildQuery(params)),
  deleteNutritionRecord: id => request(`/nutrition/records/${id}`, { method: 'DELETE' }),
  getNutritionSummary: () => request('/nutrition/summary'),
}

export default api
