<template>
  <view class="container">
    <view class="card">
      <text class="section-title">完善个人资料</text>
      <text class="section-desc">用于提供更精准的营养建议</text>

      <view class="form">
        <view class="form-item">
          <text class="label">昵称</text>
          <input class="input" v-model="form.nickname" placeholder="请输入昵称" />
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
        gender: 1,
        height: '',
        weight: '',
        age: '',
      },
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
          this.form.gender = res.gender || 1
          this.form.height = String(res.height)
          this.form.weight = String(res.weight)
          this.form.age = String(res.age)
        }
      } catch (e) {
        // New user, form stays empty
      }
    },
    async submit() {
      if (!this.form.height || !this.form.weight || !this.form.age) {
        uni.showToast({ title: '请填写完整信息', icon: 'none' })
        return
      }
      try {
        await api.updateProfile({
          nickname: this.form.nickname,
          gender: this.form.gender,
          height: parseFloat(this.form.height),
          weight: parseFloat(this.form.weight),
          age: parseInt(this.form.age),
        })
        uni.setStorageSync('has_profile', true)
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
