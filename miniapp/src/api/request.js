const BASE_URL = (import.meta.env.VITE_BASE_URL || 'http://localhost:8080') + '/api/v1'

function request(url, options = {}) {
  const token = uni.getStorageSync('token')
  const header = {
    'Content-Type': 'application/json',
    ...options.header,
  }
  if (token) {
    header['Authorization'] = 'Bearer ' + token
  }

  return new Promise((resolve, reject) => {
    uni.request({
      url: BASE_URL + url,
      method: options.method || 'GET',
      data: options.data,
      header,
      success(res) {
        if (res.statusCode === 401) {
          uni.removeStorageSync('token')
          // 不自动跳转，由各页面自行判断登录态
          reject(new Error('未登录'))
          return
        }
        if (res.statusCode >= 200 && res.statusCode < 300) {
          resolve(res.data)
        } else {
          const apiError = res.data && res.data.error
          reject(new Error((apiError && (apiError.message || apiError.code)) || apiError || '请求失败'))
        }
      },
      fail(err) {
        reject(new Error(err.errMsg || '网络请求失败'))
      },
    })
  })
}

function uploadFile(url, filePath, options = {}) {
  const token = uni.getStorageSync('token')
  const header = { ...options.header }
  if (token) header.Authorization = 'Bearer ' + token

  return new Promise((resolve, reject) => {
    uni.uploadFile({
      url: BASE_URL + url,
      filePath,
      name: options.name || 'file',
      formData: options.formData,
      header,
      success(res) {
        let data = res.data
        try { data = JSON.parse(res.data) } catch {}
        if (res.statusCode === 401) {
          uni.removeStorageSync('token')
          reject(new Error('未登录'))
          return
        }
        if (res.statusCode >= 200 && res.statusCode < 300) {
          resolve(data)
          return
        }
        const apiError = data && data.error
        reject(new Error((apiError && (apiError.message || apiError.code)) || apiError || '上传失败'))
      },
      fail(err) {
        reject(new Error(err.errMsg || '上传失败'))
      },
    })
  })
}

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
  // Auth
  wxLogin: (code, inviterId) => request('/auth/wx-login', { method: 'POST', data: { code, inviter_id: inviterId || 0 } }),

  // User
  getProfile: () => request('/user/profile'),
  updateProfile: (data) => request('/user/profile', { method: 'PUT', data }),
  deleteAccount: () => request('/user/account', { method: 'DELETE' }),

  // Speech
  transcribeVisitAudio: filePath => uploadFile('/speech/transcribe', filePath, {
    name: 'audio',
    formData: { scene: 'visit_experience' },
  }),

  // Privacy
  agreePrivacy: (data) => request('/agreements/accept', { method: 'POST', data }),

  // City dining
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
