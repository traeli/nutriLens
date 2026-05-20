<template>
  <view class="share-poster-page">
    <!-- Loading -->
    <view class="loading-section" v-if="loading">
      <view class="spinner" />
      <text class="loading-text">加载海报中...</text>
    </view>

    <!-- Error -->
    <view class="empty-hint" v-if="!loading && errorMsg">
      <text class="empty-icon">😞</text>
      <text class="empty-text">{{ errorMsg }}</text>
      <button class="btn-primary" @tap="goHome">返回首页</button>
    </view>

    <!-- Preview -->
    <view class="preview-section" v-if="!loading && !errorMsg && posterUrl">
      <image
        :src="posterUrl"
        mode="widthFix"
        class="poster-preview"
        @tap="previewFull"
        :show-menu-by-longpress="true"
      />
    </view>

    <!-- Actions -->
    <view class="action-section" v-if="!loading && !errorMsg && posterUrl">
      <button class="action-btn save-btn" @tap="saveToAlbum">
        保存到相册
      </button>
      <button class="action-btn share-btn" @tap="sharePoster">
        分享给好友
      </button>
    </view>
  </view>
</template>

<script>
import { api } from '@/api/request.js'

export default {
  data() {
    return {
      posterUrl: '',
      loading: true,
      errorMsg: '',
    }
  },
  onLoad() {
    uni.setNavigationBarTitle({ title: '分享海报' })
    this.fetchPoster()
  },
  onShareAppMessage() {
    const userId = uni.getStorageSync('user_id') || ''
    return {
      title: 'NutriLens — AI食物热量分析',
      path: '/pages/home/home?inviter_id=' + userId,
    }
  },
  methods: {
    async fetchPoster() {
      this.loading = true
      try {
        const res = await api.getSharePoster()
        this.posterUrl = res.image_url || ''
        if (!this.posterUrl) {
          this.errorMsg = '暂无海报'
        }
      } catch (e) {
        this.errorMsg = '获取海报失败，请稍后重试'
      } finally {
        this.loading = false
      }
    },

    async saveToAlbum() {
      try {
        // Download remote image to temp first
        const tempPath = await this.downloadImage(this.posterUrl)
        const hasPermission = await this.checkPhotoPermission()
        if (!hasPermission) {
          uni.showModal({
            title: '需要相册权限',
            content: '请在设置中开启相册权限以保存海报',
            confirmText: '去设置',
            success: (res) => {
              if (res.confirm) uni.openSetting({})
            },
          })
          return
        }

        await new Promise((resolve, reject) => {
          uni.saveImageToPhotosAlbum({
            filePath: tempPath,
            success: () => {
              uni.showToast({ title: '已保存到相册', icon: 'success' })
              resolve()
            },
            fail: reject,
          })
        })

        try { await api.recordShare('poster') } catch (e) { /* ok */ }
      } catch (err) {
        console.error('Save failed:', err)
        uni.showToast({ title: '保存失败', icon: 'none' })
      }
    },

    downloadImage(url) {
      return new Promise((resolve, reject) => {
        uni.downloadFile({
          url,
          success: (res) => {
            if (res.statusCode === 200) resolve(res.tempFilePath)
            else reject(new Error('下载失败'))
          },
          fail: reject,
        })
      })
    },

    checkPhotoPermission() {
      return new Promise((resolve) => {
        uni.getSetting({
          success: (res) => {
            if (res.authSetting['scope.writePhotosAlbum'] === false) {
              resolve(false)
            } else {
              uni.authorize({
                scope: 'scope.writePhotosAlbum',
                success: () => resolve(true),
                fail: () => resolve(false),
              })
            }
          },
          fail: () => resolve(false),
        })
      })
    },

    sharePoster() {
      // Download to temp first, then share
      uni.downloadFile({
        url: this.posterUrl,
        success: (res) => {
          if (res.statusCode === 200) {
            uni.showShareImageMenu({
              path: res.tempFilePath,
              success: () => {
                try { api.recordShare('poster') } catch (e) { /* ok */ }
              },
              fail: () => {
                uni.showModal({
                  title: '分享提示',
                  content: '请长按海报图片，使用微信分享给好友',
                  showCancel: false,
                })
              },
            })
          }
        },
        fail: () => {
          uni.showModal({
            title: '分享提示',
            content: '请长按海报图片，使用微信分享给好友',
            showCancel: false,
          })
        },
      })
    },

    goHome() {
      uni.switchTab({ url: '/pages/home/home' })
    },

    previewFull() {
      uni.previewImage({
        urls: [this.posterUrl],
        current: this.posterUrl,
      })
    },
  },
}
</script>

<style scoped>
.share-poster-page {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 30rpx;
  padding-bottom: 60rpx;
  min-height: 100vh;
  background: #f5f5f5;
}

.preview-section {
  width: 100%;
  max-width: 670rpx;
  margin-bottom: 40rpx;
}

.poster-preview {
  width: 100%;
  border-radius: 16rpx;
  box-shadow: 0 4rpx 24rpx rgba(0, 0, 0, 0.12);
}

.loading-section {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 120rpx 0;
}

.spinner {
  width: 64rpx;
  height: 64rpx;
  border: 4rpx solid #E0E0E0;
  border-top-color: #4CAF50;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  margin-bottom: 24rpx;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.loading-text {
  font-size: 28rpx;
  color: #999;
}

.empty-hint {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 80rpx 40rpx;
  text-align: center;
}

.empty-icon {
  font-size: 100rpx;
  margin-bottom: 30rpx;
}

.empty-text {
  font-size: 32rpx;
  color: #333;
  margin-bottom: 40rpx;
}

.btn-primary {
  background: linear-gradient(135deg, #4CAF50, #66BB6A);
  color: #fff;
  border-radius: 44rpx;
  font-size: 28rpx;
  padding: 0 60rpx;
  height: 80rpx;
  line-height: 80rpx;
  border: none;
}

.action-section {
  width: 100%;
  max-width: 670rpx;
  display: flex;
  flex-direction: column;
  gap: 20rpx;
}

.action-btn {
  width: 100%;
  height: 88rpx;
  line-height: 88rpx;
  border-radius: 44rpx;
  font-size: 30rpx;
  font-weight: 500;
  text-align: center;
  border: none;
}

.save-btn {
  background: linear-gradient(135deg, #4CAF50, #66BB6A);
  color: #fff;
}

.share-btn {
  background: #fff;
  color: #4CAF50;
  border: 2rpx solid #4CAF50;
}
</style>
