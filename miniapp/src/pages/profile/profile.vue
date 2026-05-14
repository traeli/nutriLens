<template>
  <view class="container">
    <view class="card">
      <text class="section-title">完善个人资料</text>
      <text class="section-desc">用于提供更精准的营养建议</text>

      <view class="form">
        <!-- Avatar -->
        <view class="form-item avatar-item">
          <text class="label">头像</text>
          <button class="avatar-btn" open-type="chooseAvatar" @chooseavatar="onChooseAvatar">
            <image v-if="form.avatar_url" class="avatar-img" :src="form.avatar_url" mode="aspectFill" />
            <view v-else class="avatar-placeholder">
              <text class="avatar-placeholder-text">+</text>
            </view>
          </button>
        </view>

        <!-- Nickname (WeChat nickname capability) -->
        <view class="form-item">
          <text class="label">昵称</text>
          <input class="input nickname-input" type="nickname" v-model="form.nickname"
            placeholder="点击获取微信昵称" @blur="onNicknameBlur" />
        </view>

        <view class="form-item">
          <text class="label">性别</text>
          <view class="gender-picker">
            <view class="gender-option" :class="{ active: form.gender === 1 }" @tap="form.gender = 1">
              <text>♂ 男</text>
            </view>
            <view class="gender-option" :class="{ active: form.gender === 2 }" @tap="form.gender = 2">
              <text>♀ 女</text>
            </view>
          </view>
        </view>

        <view class="form-item">
          <text class="label">身高 (cm)</text>
          <input class="input" v-model="form.height" type="digit" placeholder="如: 170" />
        </view>

        <view class="form-item">
          <text class="label">体重 (kg)</text>
          <input class="input" v-model="form.weight" type="digit" placeholder="如: 65" />
        </view>

        <view class="form-item">
          <text class="label">年龄</text>
          <input class="input" v-model="form.age" type="number" placeholder="如: 25" />
        </view>
      </view>
    </view>

    <button class="btn-primary" @tap="submit">保存</button>
  </view>
</template>

<script>
import { api } from '@/api/request.js'

export default {
  data() {
    return {
      form: {
        nickname: '',
        avatar_url: '',
        gender: 1,
        height: '',
        weight: '',
        age: '',
      },
      avatarTempPath: '',
    }
  },
  onLoad() {
    this.loadProfile()
  },
  methods: {
    async loadProfile() {
      try {
        const res = await api.getProfile()
        if (res.height > 0) {
          this.form.nickname = res.nickname || ''
          this.form.avatar_url = res.avatar_url || ''
          this.form.gender = res.gender || 1
          this.form.height = String(res.height)
          this.form.weight = String(res.weight)
          this.form.age = String(res.age)
        }
      } catch (e) {
        // New user, form stays empty
      }
    },
    async onChooseAvatar(e) {
      const tempPath = e.detail.avatarUrl
      this.avatarTempPath = tempPath
      this.form.avatar_url = tempPath
    },
    onNicknameBlur(e) {
      // type="nickname" input fills the value from WeChat on blur
      if (e.detail.value) {
        this.form.nickname = e.detail.value
      }
    },
    async submit() {
      if (!this.form.height || !this.form.weight || !this.form.age) {
        uni.showToast({ title: '请填写完整信息', icon: 'none' })
        return
      }

      let avatarUrl = this.form.avatar_url

      // Upload avatar to COS if it's a local temp file
      if (this.avatarTempPath && this.avatarTempPath === avatarUrl) {
        try {
          uni.showLoading({ title: '上传头像...' })
          const res = await api.uploadToCOS(this.avatarTempPath, 'avatar')
          avatarUrl = res.object_url
          this.form.avatar_url = avatarUrl
        } catch (e) {
          console.error('avatar upload failed:', e)
          // Continue without avatar upload
        } finally {
          uni.hideLoading()
        }
      }

      try {
        await api.updateProfile({
          nickname: this.form.nickname,
          avatar_url: avatarUrl,
          gender: this.form.gender,
          height: parseFloat(this.form.height),
          weight: parseFloat(this.form.weight),
          age: parseInt(this.form.age),
        })
        uni.setStorageSync('has_profile', true)
        uni.setStorageSync('nickname', this.form.nickname)
        uni.showToast({ title: '保存成功', icon: 'success' })
        setTimeout(() => {
          uni.switchTab({ url: '/pages/home/home' })
        }, 1000)
      } catch (err) {
        uni.showToast({ title: '保存失败: ' + err.message, icon: 'none' })
      }
    },
  },
}
</script>

<style scoped>
.section-title {
  font-size: 40rpx;
  font-weight: bold;
  color: #333;
  display: block;
  margin-bottom: 8rpx;
}

.section-desc {
  font-size: 26rpx;
  color: #999;
  display: block;
  margin-bottom: 40rpx;
}

.form {
  margin-top: 20rpx;
}

.form-item {
  margin-bottom: 36rpx;
}

.avatar-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.avatar-btn {
  width: 120rpx;
  height: 120rpx;
  border-radius: 50%;
  padding: 0;
  margin: 0;
  background: #f5f5f5;
  border: none;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
  line-height: normal;
}

.avatar-btn::after {
  border: none;
}

.avatar-img {
  width: 120rpx;
  height: 120rpx;
  border-radius: 50%;
}

.avatar-placeholder {
  width: 120rpx;
  height: 120rpx;
  display: flex;
  align-items: center;
  justify-content: center;
}

.avatar-placeholder-text {
  font-size: 48rpx;
  color: #ccc;
}

.label {
  font-size: 28rpx;
  color: #666;
  display: block;
  margin-bottom: 16rpx;
}

.input {
  background: #f5f5f5;
  border-radius: 16rpx;
  padding: 24rpx;
  font-size: 30rpx;
}

.nickname-input {
  background: #FFF8E1;
}

.gender-picker {
  display: flex;
  gap: 20rpx;
}

.gender-option {
  flex: 1;
  text-align: center;
  padding: 20rpx;
  border-radius: 16rpx;
  background: #f5f5f5;
  font-size: 30rpx;
  color: #999;
}

.gender-option.active {
  background: #e8f5e9;
  color: #4CAF50;
  font-weight: bold;
}

.btn-primary {
  margin-top: 40rpx;
}
</style>
