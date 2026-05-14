<template>
  <view class="container">
    <!-- Image Preview -->
    <view class="card" v-if="imageUrl">
      <image :src="imageUrl" mode="aspectFit" class="preview-image" />
    </view>

    <!-- Meal Type Selection -->
    <view class="card">
      <text class="label">用餐类型</text>
      <view class="meal-options">
        <text class="meal-tag" v-for="m in mealTypes" :key="m.value"
          :class="{ active: mealType === m.value }" @tap="mealType = m.value">
          {{ m.label }}
        </text>
      </view>
    </view>

    <!-- Analyzing State -->
    <view class="card analyzing-card" v-if="analyzing">
      <view class="analyzing-content">
        <view class="loading-spinner"></view>
        <text class="analyzing-text">AI 正在分析你的食物...</text>
        <text class="analyzing-sub">正在识别食物并计算营养成分</text>
      </view>
    </view>

    <!-- Result -->
    <view v-if="records.length > 0 && !analyzing">
      <!-- Total Calories Card -->
      <view class="card result-card">
        <text class="result-title">{{ records.length > 1 ? '共 ' + records.length + ' 种食物' : records[0].food_name }}</text>
        <view class="calorie-display">
          <text class="calorie-value">{{ totalCalories }}</text>
          <text class="calorie-unit">kcal</text>
        </view>
      </view>

      <!-- Per-food items -->
      <view class="card" v-for="(item, idx) in records" :key="item.id || idx">
        <view class="food-header">
          <text class="food-name-result">{{ item.food_name }}</text>
          <text class="food-weight-result" v-if="item.unit_amount">{{ (item.unit_amount * 1000).toFixed(0) }}g</text>
        </view>
        <text class="food-cal-result">{{ item.calories }} kcal</text>
        <view class="nutrients-grid" v-if="item.nutrients">
          <view class="nutrient-item">
            <text class="nutrient-name">蛋白质</text>
            <text class="nutrient-val">{{ (item.nutrients.protein || 0).toFixed(1) }}g</text>
          </view>
          <view class="nutrient-item">
            <text class="nutrient-name">碳水化合物</text>
            <text class="nutrient-val">{{ (item.nutrients.carbs || 0).toFixed(1) }}g</text>
          </view>
          <view class="nutrient-item">
            <text class="nutrient-name">脂肪</text>
            <text class="nutrient-val">{{ (item.nutrients.fat || 0).toFixed(1) }}g</text>
          </view>
          <view class="nutrient-item">
            <text class="nutrient-name">膳食纤维</text>
            <text class="nutrient-val">{{ (item.nutrients.fiber || 0).toFixed(1) }}g</text>
          </view>
          <view class="nutrient-item">
            <text class="nutrient-name">糖分</text>
            <text class="nutrient-val sugar-val">{{ (item.nutrients.sugar || 0).toFixed(1) }}g</text>
          </view>
          <view class="nutrient-item">
            <text class="nutrient-name">维C</text>
            <text class="nutrient-val">{{ (item.nutrients.vitamin_c || 0).toFixed(1) }}mg</text>
          </view>
        </view>
      </view>

      <!-- AI Suggestion -->
      <view class="card suggestion-card" v-if="suggestion">
        <text class="card-title">AI 建议</text>
        <text class="suggestion-text">{{ suggestion }}</text>
      </view>

      <button class="btn-primary" @tap="goHome">返回首页</button>
      <button class="btn-share" @tap="sharePoster">生成分享海报</button>
    </view>

    <!-- Analyze Button (before analysis) -->
    <button class="btn-primary" v-if="!analyzing && records.length === 0" @tap="analyze">
      开始分析
    </button>
  </view>
</template>

<script>
import { api } from '@/api/request.js'

export default {
  data() {
    return {
      imageUrl: '',
      mealType: 1,
      analyzing: false,
      records: [],
      suggestion: '',
      mealTypes: [
        { label: '早餐', value: 1 },
        { label: '午餐', value: 2 },
        { label: '晚餐', value: 3 },
        { label: '加餐', value: 4 },
      ],
    }
  },
  computed: {
    totalCalories() {
      return this.records.reduce((sum, r) => sum + (r.calories || 0), 0)
    },
  },
  onLoad(options) {
    if (options.image) {
      this.imageUrl = decodeURIComponent(options.image)
    }
    // Auto detect meal type by time
    const h = new Date().getHours()
    if (h < 10) this.mealType = 1
    else if (h < 14) this.mealType = 2
    else if (h < 20) this.mealType = 3
    else this.mealType = 4
  },
  onShareAppMessage() {
    const userId = uni.getStorageSync('user_id') || ''
    return {
      title: `我刚用NutriLens分析了食物热量，今天吃了${Math.round(this.totalCalories)}千卡！`,
      path: '/pages/home/home?inviter_id=' + userId,
    }
  },
  onShareTimeline() {
    const userId = uni.getStorageSync('user_id') || ''
    return {
      title: `NutriLens — AI食物热量分析，今天吃了${Math.round(this.totalCalories)}千卡`,
      path: '/pages/home/home?inviter_id=' + userId,
    }
  },
  methods: {
    async analyze() {
      if (!this.imageUrl) {
        uni.showToast({ title: '请先拍照', icon: 'none' })
        return
      }
      this.analyzing = true
      try {
        const res = await api.analyzeImage(this.imageUrl, this.mealType)
        if (res.is_food === false) {
          uni.showModal({
            title: '未识别到食物',
            content: '图片中没有检测到食物，请重新拍照',
            showCancel: false,
          })
          return
        }
        this.records = res.records || []
        this.suggestion = res.suggestion || ''
      } catch (err) {
        uni.showToast({ title: '分析失败: ' + err.message, icon: 'none' })
      } finally {
        this.analyzing = false
      }
    },
    goHome() {
      uni.switchTab({ url: '/pages/home/home' })
    },
    async sharePoster() {
      try {
        await api.recordShare('poster')
        uni.showToast({ title: '分享已记录', icon: 'success' })
      } catch (e) {
        console.error('record share failed:', e)
      }
      // Trigger share via WeChat
      uni.showActionSheet({
        itemList: ['分享到微信好友', '生成海报图片'],
        success: (res) => {
          if (res.tapIndex === 0) {
            // WeChat share is handled by onShareAppMessage
            uni.showToast({ title: '请点击右上角分享给好友', icon: 'none' })
          } else {
            uni.showToast({ title: '海报功能开发中', icon: 'none' })
          }
        },
      })
    },
  },
}
</script>

<style scoped>
.preview-image {
  width: 100%;
  height: 400rpx;
  border-radius: 16rpx;
}

.label {
  font-size: 28rpx;
  color: #666;
  display: block;
  margin-bottom: 16rpx;
}

.meal-options {
  display: flex;
  gap: 16rpx;
}

.meal-tag {
  padding: 12rpx 28rpx;
  border-radius: 30rpx;
  background: #f5f5f5;
  font-size: 26rpx;
  color: #666;
}

.meal-tag.active {
  background: #e8f5e9;
  color: #4CAF50;
}

.analyzing-card {
  text-align: center;
  padding: 60rpx 30rpx;
}

.analyzing-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 16rpx;
}

.loading-spinner {
  width: 60rpx;
  height: 60rpx;
  border: 4rpx solid #e0e0e0;
  border-top: 4rpx solid #4CAF50;
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.analyzing-text {
  font-size: 32rpx;
  font-weight: bold;
  color: #333;
}

.analyzing-sub {
  font-size: 24rpx;
  color: #999;
}

.result-card {
  text-align: center;
  background: linear-gradient(135deg, #4CAF50, #66BB6A);
  color: #fff;
}

.result-title {
  font-size: 36rpx;
  font-weight: bold;
  display: block;
  margin-bottom: 16rpx;
}

.calorie-display {
  display: flex;
  align-items: baseline;
  justify-content: center;
  gap: 8rpx;
}

.calorie-value {
  font-size: 72rpx;
  font-weight: bold;
}

.calorie-unit {
  font-size: 28rpx;
  opacity: 0.8;
}

.food-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8rpx;
}

.food-name-result {
  font-size: 32rpx;
  font-weight: bold;
  color: #333;
}

.food-weight-result {
  font-size: 24rpx;
  color: #999;
  background: #f5f5f5;
  padding: 4rpx 12rpx;
  border-radius: 8rpx;
}

.food-cal-result {
  font-size: 28rpx;
  color: #FF7043;
  font-weight: bold;
  display: block;
  margin-bottom: 16rpx;
}

.card-title {
  font-size: 30rpx;
  font-weight: bold;
  color: #333;
  display: block;
  margin-bottom: 20rpx;
}

.nutrients-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20rpx;
}

.nutrient-item {
  background: #f5f5f5;
  padding: 20rpx;
  border-radius: 16rpx;
  display: flex;
  flex-direction: column;
  gap: 8rpx;
}

.nutrient-name {
  font-size: 24rpx;
  color: #999;
}

.nutrient-val {
  font-size: 32rpx;
  font-weight: bold;
  color: #333;
}

.sugar-val {
  color: #FF7043;
}

.suggestion-text {
  font-size: 28rpx;
  color: #666;
  line-height: 1.8;
}

.btn-primary {
  margin-top: 30rpx;
  margin-bottom: 16rpx;
}

.btn-share {
  margin-top: 0;
  margin-bottom: 60rpx;
  background: #FF9800;
  color: #fff;
  border-radius: 50rpx;
  font-size: 28rpx;
}
</style>
