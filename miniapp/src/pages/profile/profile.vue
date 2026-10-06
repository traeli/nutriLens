<template>
  <view class="profile-page">
    <image class="profile-art" :src="profileArt" mode="aspectFill" />
    <view class="profile-form">
      <button class="avatar-button" open-type="chooseAvatar" @chooseavatar="onChooseAvatar">
        <image v-if="form.avatar_url" class="avatar-image" :src="form.avatar_url" mode="aspectFill" />
        <view v-else class="avatar-placeholder">
          <text>{{ avatarInitial }}</text>
        </view>
        <view class="avatar-edit">＋</view>
      </button>

      <view class="form-card">
        <view class="form-item">
          <text class="field-label">昵称</text>
          <input
            class="nickname-input"
            type="nickname"
            v-model="form.nickname"
            maxlength="64"
            placeholder="请输入昵称"
            placeholder-class="input-placeholder"
            confirm-type="done"
            :disabled="loadingProfile || submitting"
            @blur="onNicknameBlur"
          />
        </view>

        <button
          class="save-button"
          :disabled="!nicknameReady || loadingProfile || submitting"
          :loading="submitting"
          @tap="submit"
        >
          保存
        </button>
      </view>
    </view>
  </view>
</template>

<script>
import { api } from '@/api/request.js'
import { assetUrl } from '@/utils/assets.js'

export default {
  data() {
    return {
      profileArt: assetUrl('/static/dining/visit-rating-art-v2.jpg'),
      form: {
        nickname: '',
        avatar_url: '',
      },
      avatarTempPath: '',
      loadingProfile: true,
      submitting: false,
    }
  },
  computed: {
    nicknameReady() {
      return Boolean(this.form.nickname.trim())
    },
    avatarInitial() {
      return this.form.nickname.trim().slice(0, 1) || '我'
    },
  },
  onLoad() {
    if (!uni.getStorageSync('token')) {
      uni.reLaunch({ url: '/pages/login/login' })
      return
    }
    this.loadProfile()
  },
  methods: {
    async loadProfile() {
      this.loadingProfile = true
      try {
        const res = await api.getProfile()
        this.form.nickname = res.nickname || ''
        this.form.avatar_url = res.avatar_url || ''
      } catch (error) {
        if (error.status !== 401) console.warn('[profile] 个人资料加载失败', error.message || error)
      } finally {
        this.loadingProfile = false
      }
    },
    onChooseAvatar(e) {
      const tempPath = e.detail.avatarUrl
      if (!tempPath) return
      this.avatarTempPath = tempPath
      this.form.avatar_url = tempPath
    },
    onNicknameBlur(e) {
      if (e.detail.value) {
        this.form.nickname = e.detail.value
      }
    },
    async submit() {
      if (!this.nicknameReady || this.submitting) return

      this.submitting = true
      const nickname = this.form.nickname.trim()
      try {
        await api.updateProfile({
          nickname,
          avatar_url: this.avatarTempPath ? '' : this.form.avatar_url,
        })
        uni.setStorageSync('has_profile', true)
        uni.setStorageSync('nickname', nickname)
        uni.showToast({ title: '保存成功', icon: 'success' })
        setTimeout(() => {
          uni.switchTab({ url: '/pages/home/home' })
        }, 800)
      } catch (err) {
        uni.showToast({ title: '保存失败，请稍后重试', icon: 'none' })
      } finally {
        this.submitting = false
      }
    },
  },
}
</script>

<style scoped>
.profile-page {
  position: relative;
  min-height: 100vh;
  overflow: hidden;
  background: #F3F3EF;
}

.profile-art {
  position: absolute;
  right: 0;
  bottom: 0;
  left: 0;
  width: 100%;
  height: 68%;
  opacity: .13;
}

.profile-form {
  position: relative;
  z-index: 1;
  padding: 160rpx 42rpx 70rpx;
}

.avatar-button {
  position: relative;
  z-index: 2;
  width: 142rpx;
  height: 142rpx;
  margin: 0 auto -70rpx;
  padding: 7rpx;
  overflow: visible;
  border-radius: 50%;
  background: #F8F8F4;
  box-shadow: 0 10rpx 28rpx rgba(25, 24, 21, .1);
  line-height: normal;
}

.avatar-button::after {
  border: 0;
}

.avatar-image,
.avatar-placeholder {
  width: 128rpx;
  height: 128rpx;
  border-radius: 50%;
}

.avatar-placeholder {
  display: flex;
  align-items: center;
  justify-content: center;
  background: #DFDDD8;
  color: #57534E;
}

.avatar-placeholder text {
  font-size: 44rpx;
  font-weight: 650;
}

.avatar-edit {
  position: absolute;
  right: -1rpx;
  bottom: 4rpx;
  width: 42rpx;
  height: 42rpx;
  border: 5rpx solid #F8F8F4;
  border-radius: 50%;
  background: #D64B32;
  color: #FFFFFF;
  font-size: 25rpx;
  line-height: 32rpx;
  text-align: center;
}

.form-card {
  padding: 106rpx 32rpx 32rpx;
  border: 1rpx solid #D7D5CF;
  border-radius: 24rpx;
  background: #F8F8F4;
  box-shadow: 0 12rpx 31rpx rgba(25, 24, 21, .07);
}

.form-item {
  padding: 0 3rpx;
}

.field-label {
  display: block;
  margin-bottom: 14rpx;
  color: #1B1B19;
  font-size: 25rpx;
  font-weight: 650;
}

.nickname-input {
  height: 94rpx;
  padding: 0 24rpx;
  border: 1rpx solid #D7D5CF;
  border-radius: 18rpx;
  background: #FFFFFF;
  color: #1B1B19;
  font-size: 29rpx;
}

.input-placeholder {
  color: #88827A;
}

.save-button {
  height: 94rpx;
  margin: 32rpx 0 0;
  border-radius: 18rpx;
  background: #D64B32;
  color: #FFFFFF;
  font-size: 29rpx;
  font-weight: 650;
  line-height: 94rpx;
}

.save-button[disabled] {
  background: #D9D7D1;
  color: #88827A;
  opacity: 1;
}
</style>
