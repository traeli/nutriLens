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
        reject(new Error(err.errMsg || '网络请求失败'))
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

// COS upload: get presigned URL from COS SDK, then PUT file directly
async function uploadToCOS(filePath, bizType) {
  const filename = filePath.split('/').pop() || 'image.jpg'
  console.log('[COS] step1: requesting presign...', { bizType, filename, filePath })

  const { upload_url, object_key, object_url } = await request('/upload/presign', {
    method: 'POST',
    data: { biz_type: bizType, filename },
  })
  console.log('[COS] step1 OK: presign result', { upload_url, object_key })

  // Read file as ArrayBuffer and PUT to COS presigned URL
  const fs = uni.getFileSystemManager()
  const fileData = fs.readFileSync(filePath)

  console.log('[COS] step2: PUT to COS...')
  await new Promise((resolve, reject) => {
    uni.request({
      url: upload_url,
      method: 'PUT',
      data: fileData,
      success(res) {
        console.log('[COS] step2 response:', res.statusCode)
        if (res.statusCode >= 200 && res.statusCode < 300) {
          resolve()
        } else {
          console.error('[COS] step2 FULL ERROR:', res.data)
          reject(new Error('COS upload failed: ' + res.statusCode))
        }
      },
      fail(err) {
        console.error('[COS] step2 FAILED:', err)
        reject(new Error(err.errMsg || 'COS upload failed'))
      },
    })
  })

  console.log('[COS] upload success:', { object_key, object_url })
  return { object_key, object_url }
}

export const api = {
  // Auth
  wxLogin: (code, inviterId) => request('/auth/wx-login', { method: 'POST', data: { code, inviter_id: inviterId || 0 } }),

  // User
  getProfile: () => request('/user/profile'),
  updateProfile: (data) => request('/user/profile', { method: 'PUT', data }),

  // Upload
  getPresignedUrl: (data) => request('/upload/presign', { method: 'POST', data }),
  uploadToCOS,

  // Food — image analyze: upload to COS first, then analyze
  async analyzeImage(filePath, mealType) {
    const { object_key, object_url } = await uploadToCOS(filePath, 'food')
    const res = await request('/food/analyze/image', {
      method: 'POST',
      data: { image_key: object_key, meal_type: mealType },
    })
    res._image_url = object_url
    return res
  },
  analyzeText: (data) => request('/food/analyze/text', { method: 'POST', data }),
  getFoodRecords: (params) => request('/food/records' + buildQuery(params)),
  getFoodRecord: (id) => request('/food/records/' + id),
  deleteFoodRecord: (id) => request('/food/records/' + id, { method: 'DELETE' }),
  getDailySummary: (date) => request('/food/daily-summary' + (date ? '?date=' + date : '')),
  getMonthlySummary: (month) => request('/food/monthly-summary?month=' + month),
  getDailyAnalysis: (date) => request('/food/daily-analysis' + (date ? '?date=' + date : '')),

  // Wheel
  getDishes: (category) => request('/wheel/dishes' + (category ? '?category=' + category : '')),
  spinWheel: (category) => request('/wheel/spin' + (category ? '?category=' + category : ''), { method: 'POST' }),
  createDish: (data) => request('/wheel/dishes', { method: 'POST', data }),
  createDishAI: (data) => request('/wheel/dishes/ai', { method: 'POST', data }),
  deleteDish: (id) => request('/wheel/dishes/' + id, { method: 'DELETE' }),

  // Privacy
  agreePrivacy: (data) => request('/privacy/agree', { method: 'POST', data }),
  getPrivacyStatus: () => request('/privacy/status'),

  // Feedback
  createFeedback: (data) => request('/feedback', { method: 'POST', data }),
  getFeedbacks: () => request('/feedback'),

  // Score / Rank
  getMyScore: () => request('/score/my'),
  getRank: (params) => request('/score/rank' + buildQuery(params)),
  getScoreLogs: (params) => request('/score/log' + buildQuery(params)),
  toggleRankVisibility: () => request('/user/rank-visibility', { method: 'PUT' }),

  // Share
  getSharePoster: () => request('/share/poster'),
  recordShare: (shareType) => request('/share/record', { method: 'POST', data: { share_type: shareType || 'poster' } }),

  // Achievement
  getAchievements: () => request('/achievement/list'),
  setTitle: (data) => request('/achievement/title', { method: 'PUT', data }),
  checkAchievements: () => request('/achievement/check', { method: 'POST' }),

  // Notify
  getNotifySettings: () => request('/notify/settings'),
  updateNotifySettings: (data) => request('/notify/settings', { method: 'PUT', data }),
}

export default api
