const BASE_URL = (import.meta.env.VITE_BASE_URL || 'http://localhost:8080') + '/api/v1'

export const SERVER_URL = import.meta.env.VITE_BASE_URL || 'http://localhost:8080'

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
          reject(new Error(res.data.error || '请求失败'))
        }
      },
      fail(err) {
        reject(err)
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

function uploadFile(url, filePath, name = 'image', formData = {}) {
  const token = uni.getStorageSync('token')
  return new Promise((resolve, reject) => {
    uni.uploadFile({
      url: BASE_URL + url,
      filePath,
      name,
      formData,
      header: {
        Authorization: token ? 'Bearer ' + token : '',
      },
      success(res) {
        const data = JSON.parse(res.data)
        if (res.statusCode >= 200 && res.statusCode < 300) {
          resolve(data)
        } else {
          reject(new Error(data.error || '上传失败'))
        }
      },
      fail(err) {
        reject(err)
      },
    })
  })
}

export const api = {
  // Auth
  wxLogin: (code) => request('/auth/wx-login', { method: 'POST', data: { code } }),

  // User
  getProfile: () => request('/user/profile'),
  updateProfile: (data) => request('/user/profile', { method: 'PUT', data }),

  // Food
  analyzeImage: (filePath, mealType) => uploadFile('/food/analyze/image', filePath, 'image', { meal_type: String(mealType) }),
  analyzeText: (data) => request('/food/analyze/text', { method: 'POST', data }),
  getFoodRecords: (params) => request('/food/records' + buildQuery(params)),
  getFoodRecord: (id) => request('/food/records/' + id),
  deleteFoodRecord: (id) => request('/food/records/' + id, { method: 'DELETE' }),
  getDailySummary: (date) => request('/food/daily-summary' + (date ? '?date=' + date : '')),
  getMonthlySummary: (month) => request('/food/monthly-summary?month=' + month),

  // Wheel
  getDishes: (category) => request('/wheel/dishes' + (category ? '?category=' + category : '')),
  spinWheel: (category) => request('/wheel/spin' + (category ? '?category=' + category : ''), { method: 'POST' }),
  createDish: (data) => request('/wheel/dishes', { method: 'POST', data }),
  createDishAI: (data) => request('/wheel/dishes/ai', { method: 'POST', data }),
  deleteDish: (id) => request('/wheel/dishes/' + id, { method: 'DELETE' }),

  // Privacy
  agreePrivacy: (data) => request('/privacy/agree', { method: 'POST', data }),
  getPrivacyStatus: () => request('/privacy/status'),

  // Upload
  uploadImage: (filePath) => uploadFile('/upload/image', filePath),
}

export default api
