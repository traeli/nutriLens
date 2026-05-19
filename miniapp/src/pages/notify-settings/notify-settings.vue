<template>
  <view class="page">
    <view class="tip-card">
      <text class="tip-text">开启消息推送后，我们会在您设定的饭点提醒您按时吃饭，并根据您的饮食记录和偏好推荐合适的菜品。</text>
    </view>

    <!-- Subscribe button -->
    <view class="card subscribe-card">
      <text class="card-title">订阅授权</text>
      <text class="card-desc">首次使用推送功能需要授权订阅消息，点击下方按钮进行授权。</text>
      <button class="subscribe-btn" @tap="requestSubscribe">授权订阅消息</button>
    </view>

    <!-- Breakfast -->
    <view class="card">
      <view class="setting-row">
        <view class="setting-info">
          <text class="setting-label">🌅 早餐提醒</text>
          <text class="setting-desc">每天早上推送早餐建议</text>
        </view>
        <switch :checked="settings.breakfast_enabled" @change="onToggle('breakfast_enabled', $event)" color="#4CAF50" />
      </view>
      <view class="time-row" v-if="settings.breakfast_enabled">
        <text class="time-label">提醒时间</text>
        <picker mode="time" :value="settings.breakfast_time" @change="onTimeChange('breakfast_time', $event)" class="time-picker">
          <text class="time-value">{{ settings.breakfast_time }}</text>
        </picker>
      </view>
    </view>

    <!-- Lunch -->
    <view class="card">
      <view class="setting-row">
        <view class="setting-info">
          <text class="setting-label">☀️ 午餐提醒</text>
          <text class="setting-desc">每天中午推送午餐建议</text>
        </view>
        <switch :checked="settings.lunch_enabled" @change="onToggle('lunch_enabled', $event)" color="#4CAF50" />
      </view>
      <view class="time-row" v-if="settings.lunch_enabled">
        <text class="time-label">提醒时间</text>
        <picker mode="time" :value="settings.lunch_time" @change="onTimeChange('lunch_time', $event)" class="time-picker">
          <text class="time-value">{{ settings.lunch_time }}</text>
        </picker>
      </view>
    </view>

    <!-- Dinner -->
    <view class="card">
      <view class="setting-row">
        <view class="setting-info">
          <text class="setting-label">🌙 晚餐提醒</text>
          <text class="setting-desc">每天晚上推送晚餐建议</text>
        </view>
        <switch :checked="settings.dinner_enabled" @change="onToggle('dinner_enabled', $event)" color="#4CAF50" />
      </view>
      <view class="time-row" v-if="settings.dinner_enabled">
        <text class="time-label">提醒时间</text>
        <picker mode="time" :value="settings.dinner_time" @change="onTimeChange('dinner_time', $event)" class="time-picker">
          <text class="time-value">{{ settings.dinner_time }}</text>
        </picker>
      </view>
    </view>

    <!-- Note -->
    <view class="note-card">
      <text class="note-title">温馨提示</text>
      <text class="note-text">1. 每次推送前我们会检查您是否已记录该餐次的食物，已记录则不会重复提醒。</text>
      <text class="note-text">2. 推送内容由AI根据您的饮食偏好和当日摄入情况个性化生成。</text>
      <text class="note-text">3. 需要先在微信中授权订阅消息，否则无法接收推送。</text>
    </view>
  </view>
</template>

<script>
import { api } from '@/api/request.js'

export default {
  data() {
    return {
      settings: {
        breakfast_enabled: true,
        breakfast_time: '08:00',
        lunch_enabled: true,
        lunch_time: '12:00',
        dinner_enabled: true,
        dinner_time: '18:00',
      },
    }
  },
  onShow() {
    this.loadSettings()
  },
  methods: {
    async loadSettings() {
      try {
        const res = await api.getNotifySettings()
        this.settings = res
      } catch (e) {
        console.log('load notify settings failed', e)
      }
    },
    requestSubscribe() {
      // Call WeChat subscribe message API
      // tmplIds should be updated with the actual template ID from WeChat MP backend
      uni.requestSubscribeMessage({
        tmplIds: [],
        success: (res) => {
          console.log('subscribe result:', res)
          uni.showToast({ title: '授权成功', icon: 'success' })
        },
        fail: (err) => {
          console.log('subscribe failed:', err)
          uni.showToast({ title: '授权失败，请稍后再试', icon: 'none' })
        },
      })
    },
    onToggle(field, e) {
      const val = e.detail.value
      this.settings[field] = val
      this.updateSetting({ [field]: val })
    },
    onTimeChange(field, e) {
      const val = e.detail.value
      this.settings[field] = val
      this.updateSetting({ [field]: val })
    },
    async updateSetting(data) {
      try {
        await api.updateNotifySettings(data)
      } catch (e) {
        console.log('update notify settings failed', e)
        uni.showToast({ title: '保存失败', icon: 'none' })
        // Revert on failure
        this.loadSettings()
      }
    },
  },
}
</script>

<style scoped>
.page {
  padding: 30rpx;
  min-height: 100vh;
  background: #f5f5f5;
}

.tip-card {
  background: #E8F5E9;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 24rpx;
}

.tip-text {
  font-size: 26rpx;
  color: #2E7D32;
  line-height: 1.6;
}

.card {
  background: #fff;
  border-radius: 16rpx;
  padding: 30rpx;
  margin-bottom: 20rpx;
}

.subscribe-card {
  text-align: center;
}

.card-title {
  font-size: 30rpx;
  font-weight: bold;
  color: #333;
  display: block;
  margin-bottom: 12rpx;
}

.card-desc {
  font-size: 24rpx;
  color: #999;
  display: block;
  margin-bottom: 24rpx;
}

.subscribe-btn {
  background: linear-gradient(135deg, #4CAF50, #66BB6A);
  color: #fff;
  border-radius: 50rpx;
  font-size: 28rpx;
  padding: 16rpx 60rpx;
  display: inline-block;
  border: none;
}

.setting-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.setting-info {
  flex: 1;
}

.setting-label {
  font-size: 30rpx;
  color: #333;
  display: block;
}

.setting-desc {
  font-size: 24rpx;
  color: #999;
  display: block;
  margin-top: 4rpx;
}

.time-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-top: 20rpx;
  padding-top: 20rpx;
  border-top: 1rpx solid #f5f5f5;
}

.time-label {
  font-size: 26rpx;
  color: #666;
}

.time-picker {
  background: #f5f5f5;
  border-radius: 10rpx;
  padding: 12rpx 24rpx;
}

.time-value {
  font-size: 28rpx;
  color: #4CAF50;
  font-weight: bold;
}

.note-card {
  background: #fff;
  border-radius: 16rpx;
  padding: 30rpx;
  margin-top: 10rpx;
}

.note-title {
  font-size: 28rpx;
  font-weight: bold;
  color: #333;
  display: block;
  margin-bottom: 16rpx;
}

.note-text {
  font-size: 24rpx;
  color: #999;
  display: block;
  line-height: 1.8;
  margin-bottom: 6rpx;
}
</style>
