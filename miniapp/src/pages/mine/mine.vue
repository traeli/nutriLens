<template>
  <view class="mine-page">
    <!-- Profile Card -->
    <view class="profile-card card">
      <view class="avatar-wrap">
        <image v-if="profile.avatar_url" class="avatar-img" :src="profile.avatar_url" mode="aspectFill" />
        <text v-else class="avatar-text">{{ avatarText }}</text>
      </view>
      <view class="profile-info">
        <text class="nickname">{{ profile.nickname || '未设置昵称' }}</text>
        <text class="profile-desc" v-if="profile.height">
          {{ profile.height }}cm · {{ profile.weight }}kg · {{ profile.age }}岁 · {{ genderText }}
        </text>
        <text class="profile-desc" v-else>点击完善个人资料</text>
      </view>
      <text class="edit-arrow" @tap="goProfile">›</text>
    </view>

    <!-- BMI Card -->
    <view class="card bmi-card" v-if="profile.height && profile.weight">
      <text class="bmi-label">BMI 指数</text>
      <text class="bmi-value">{{ bmi }}</text>
      <text class="bmi-status" :class="bmiClass">{{ bmiStatus }}</text>
    </view>

    <!-- Menu -->
    <view class="menu-card card">
      <view class="menu-item" @tap="goProfile">
        <text class="menu-icon">👤</text>
        <text class="menu-text">个人资料</text>
        <text class="menu-arrow">›</text>
      </view>
      <view class="menu-item" @tap="goFoodLog">
        <text class="menu-icon">📋</text>
        <text class="menu-text">饮食记录</text>
        <text class="menu-arrow">›</text>
      </view>
      <view class="menu-item" @tap="goMyDishes">
        <text class="menu-icon">🍽</text>
        <text class="menu-text">我的菜品</text>
        <text class="menu-arrow">›</text>
      </view>
      <view class="menu-item" @tap="goNotifySettings">
        <text class="menu-icon">🔔</text>
        <text class="menu-text">消息推送</text>
        <text class="menu-arrow">›</text>
      </view>
      <view class="menu-item" @tap="showAbout">
        <text class="menu-icon">ℹ️</text>
        <text class="menu-text">关于</text>
        <text class="menu-arrow">›</text>
      </view>
      <view class="menu-item" @tap="goFeedback">
        <text class="menu-icon">💬</text>
        <text class="menu-text">意见反馈</text>
        <text class="menu-arrow">›</text>
      </view>
      <view class="menu-item" @tap="goRank">
        <text class="menu-icon">🏆</text>
        <text class="menu-text">排行榜</text>
        <text class="menu-arrow">›</text>
      </view>
      <view class="menu-item" @tap="goAchievement">
        <text class="menu-icon">🎖</text>
        <text class="menu-text">我的成就</text>
        <text class="menu-arrow">›</text>
      </view>
    </view>

    <!-- Logout -->
    <button class="logout-btn" @tap="logout">退出登录</button>

    <!-- ICP Filing -->
    <view class="icp-footer">
      <text class="icp-text">湘ICP备2024071320号-4</text>
    </view>
  </view>
</template>

<script>
import { api } from '@/api/request.js'

export default {
  data() {
    return {
      profile: {},
    }
  },
  computed: {
    avatarText() {
      return (this.profile.nickname || '?').charAt(0).toUpperCase()
    },
    genderText() {
      return this.profile.gender === 1 ? '男' : this.profile.gender === 2 ? '女' : '未知'
    },
    bmi() {
      if (!this.profile.height || !this.profile.weight) return '-'
      const h = this.profile.height / 100
      return (this.profile.weight / (h * h)).toFixed(1)
    },
    bmiClass() {
      const v = parseFloat(this.bmi)
      if (v < 18.5) return 'underweight'
      if (v < 24) return 'normal'
      if (v < 28) return 'overweight'
      return 'obese'
    },
    bmiStatus() {
      const v = parseFloat(this.bmi)
      if (v < 18.5) return '偏瘦'
      if (v < 24) return '正常'
      if (v < 28) return '偏胖'
      return '肥胖'
    },
  },
  onShow() {
    if (!uni.getStorageSync('token')) {
      uni.reLaunch({ url: '/pages/login/login' })
      return
    }
    this.loadProfile()
  },
  methods: {
    async loadProfile() {
      try {
        const res = await api.getProfile()
        this.profile = res
      } catch (e) {
        console.log('load profile failed', e)
      }
    },
    goProfile() {
      uni.navigateTo({ url: '/pages/profile/profile' })
    },
    goFoodLog() {
      uni.navigateTo({ url: '/pages/food-log/food-log' })
    },
    goMyDishes() {
      uni.navigateTo({ url: '/pages/my-dishes/my-dishes' })
    },
    goNotifySettings() {
      uni.navigateTo({ url: '/pages/notify-settings/notify-settings' })
    },
    goFeedback() {
      uni.navigateTo({ url: '/pages/feedback/feedback' })
    },
    goRank() {
      uni.navigateTo({ url: '/pages/rank/rank' })
    },
    goAchievement() {
      uni.navigateTo({ url: '/pages/achievement/achievement' })
    },
    showAbout() {
      uni.showModal({
        title: 'NutriLens 营养镜头',
        content: '版本: 1.0.0\nAI智能饮食分析助手\n拍照识别食物卡路里，获取个性化营养建议\n\n湘ICP备2024071320号-4',
        showCancel: false,
      })
    },
    logout() {
      uni.showModal({
        title: '退出登录',
        content: '确定要退出登录吗？',
        success: (res) => {
          if (res.confirm) {
            uni.removeStorageSync('token')
            uni.removeStorageSync('user_id')
            uni.removeStorageSync('has_profile')
            uni.reLaunch({ url: '/pages/login/login' })
          }
        },
      })
    },
  },
}
</script>

<style scoped>
.mine-page {
  padding: 30rpx;
  min-height: 100vh;
}

.profile-card {
  display: flex;
  align-items: center;
  gap: 24rpx;
}

.avatar-wrap {
  width: 100rpx;
  height: 100rpx;
  border-radius: 50%;
  background: linear-gradient(135deg, #4CAF50, #66BB6A);
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
}

.avatar-img {
  width: 100rpx;
  height: 100rpx;
  border-radius: 50%;
}

.avatar-text {
  font-size: 40rpx;
  color: #fff;
  font-weight: bold;
}

.profile-info {
  flex: 1;
}

.nickname {
  font-size: 32rpx;
  font-weight: bold;
  color: #333;
  display: block;
  margin-bottom: 8rpx;
}

.profile-desc {
  font-size: 24rpx;
  color: #999;
}

.edit-arrow {
  font-size: 40rpx;
  color: #ccc;
  padding: 10rpx;
}

.bmi-card {
  text-align: center;
  padding: 30rpx;
}

.bmi-label {
  font-size: 26rpx;
  color: #999;
  display: block;
  margin-bottom: 10rpx;
}

.bmi-value {
  font-size: 56rpx;
  font-weight: bold;
  color: #333;
  display: block;
  margin-bottom: 8rpx;
}

.bmi-status {
  font-size: 26rpx;
  display: block;
}

.bmi-status.normal { color: #4CAF50; }
.bmi-status.underweight { color: #42A5F5; }
.bmi-status.overweight { color: #FFA726; }
.bmi-status.obese { color: #FF7043; }

.menu-card {
  padding: 0;
}

.menu-item {
  display: flex;
  align-items: center;
  padding: 30rpx;
  border-bottom: 1rpx solid #f5f5f5;
}

.menu-item:last-child {
  border-bottom: none;
}

.menu-icon {
  font-size: 36rpx;
  margin-right: 20rpx;
}

.menu-text {
  flex: 1;
  font-size: 28rpx;
  color: #333;
}

.menu-arrow {
  font-size: 32rpx;
  color: #ccc;
}

.logout-btn {
  margin-top: 40rpx;
  background: #fff;
  color: #FF7043;
  border: 1rpx solid #FF7043;
  border-radius: 50rpx;
  font-size: 28rpx;
}

.icp-footer {
  text-align: center;
  margin-top: 60rpx;
  padding-bottom: 40rpx;
}

.icp-text {
  font-size: 22rpx;
  color: #bbb;
}
</style>
