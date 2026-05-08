<template>
  <view class="home-page">
    <!-- Header -->
    <view class="header">
      <text class="greeting">{{ greeting }} 👋</text>
      <text class="date-text">{{ today }}</text>
    </view>

    <!-- Today Summary Card -->
    <view class="card summary-card" @tap="goHistory">
      <view class="summary-header">
        <text class="summary-title">今日摄入</text>
        <text class="summary-link">查看详情 ›</text>
      </view>
      <view class="summary-body">
        <view class="calorie-main">
          <text class="calorie-num">{{ summary.total_calories || 0 }}</text>
          <text class="calorie-unit">kcal</text>
        </view>
        <view class="nutrients-row" v-if="summary.total_nutrients">
          <view class="nutrient-item">
            <text class="nutrient-val">{{ (summary.total_nutrients.protein || 0).toFixed(1) }}</text>
            <text class="nutrient-label">蛋白质g</text>
          </view>
          <view class="nutrient-item">
            <text class="nutrient-val">{{ (summary.total_nutrients.carbs || 0).toFixed(1) }}</text>
            <text class="nutrient-label">碳水g</text>
          </view>
          <view class="nutrient-item">
            <text class="nutrient-val">{{ (summary.total_nutrients.fat || 0).toFixed(1) }}</text>
            <text class="nutrient-label">脂肪g</text>
          </view>
          <view class="nutrient-item">
            <text class="nutrient-val sugar-val">{{ (summary.total_nutrients.sugar || 0).toFixed(1) }}</text>
            <text class="nutrient-label">糖分g</text>
          </view>
        </view>
      </view>
    </view>

    <!-- Action Buttons -->
    <view class="actions">
      <view class="action-card card" @tap="takePhoto">
        <view class="action-icon-wrap" style="background: #e8f5e9;">
          <text class="action-icon">📸</text>
        </view>
        <text class="action-title">拍照识别</text>
        <text class="action-desc">拍张照片分析卡路里</text>
      </view>

      <view class="action-card card" @tap="inputText">
        <view class="action-icon-wrap" style="background: #e3f2fd;">
          <text class="action-icon">✏️</text>
        </view>
        <text class="action-title">文字记录</text>
        <text class="action-desc">告诉我你吃了什么</text>
      </view>
    </view>

    <!-- Quick Wheel -->
    <view class="card wheel-card" @tap="goWheel">
      <view class="wheel-card-content">
        <text class="wheel-emoji">🎡</text>
        <view class="wheel-text">
          <text class="wheel-title">不知道吃什么？</text>
          <text class="wheel-desc">转一转大转盘帮你决定</text>
        </view>
        <text class="wheel-arrow">›</text>
      </view>
    </view>

    <!-- Analyze Result Cards -->
    <view class="result-section" v-if="analyzeResult" @tap="analyzeResult = null">
      <view class="card result-wrap" @tap.stop>
        <view class="result-header">
          <text class="result-section-title">AI 分析结果</text>
          <text class="result-close-btn" @tap="analyzeResult = null">✕</text>
        </view>
        <view class="food-card" v-for="item in analyzeResult.records" :key="item.id">
          <view class="food-card-header">
            <text class="food-card-name">{{ item.food_name }}</text>
            <text class="food-card-amount" v-if="item.unit_amount">{{ (item.unit_amount * 1000).toFixed(0) }}g</text>
            <text class="food-card-cal">{{ item.calories }} kcal</text>
          </view>
          <view class="food-card-nutrients" v-if="item.nutrients">
            <view class="nutrient-chip" v-if="item.nutrients.protein">
              <text class="chip-val">{{ item.nutrients.protein.toFixed(1) }}</text>
              <text class="chip-label">蛋白质g</text>
            </view>
            <view class="nutrient-chip" v-if="item.nutrients.carbs">
              <text class="chip-val">{{ item.nutrients.carbs.toFixed(1) }}</text>
              <text class="chip-label">碳水g</text>
            </view>
            <view class="nutrient-chip" v-if="item.nutrients.fat">
              <text class="chip-val">{{ item.nutrients.fat.toFixed(1) }}</text>
              <text class="chip-label">脂肪g</text>
            </view>
            <view class="nutrient-chip sugar-chip" v-if="item.nutrients.sugar">
              <text class="chip-val">{{ item.nutrients.sugar.toFixed(1) }}</text>
              <text class="chip-label">糖分g</text>
            </view>
            <view class="nutrient-chip" v-if="item.nutrients.fiber">
              <text class="chip-val">{{ item.nutrients.fiber.toFixed(1) }}</text>
              <text class="chip-label">纤维g</text>
            </view>
            <view class="nutrient-chip" v-if="item.nutrients.vitamin_c">
              <text class="chip-val">{{ item.nutrients.vitamin_c.toFixed(1) }}</text>
              <text class="chip-label">VitC mg</text>
            </view>
          </view>
        </view>
        <view class="result-total">
          <text class="total-text">合计: {{ resultTotalCal }} kcal</text>
        </view>
        <view class="result-suggestion" v-if="analyzeResult.suggestion">
          <text class="suggestion-label">AI 建议</text>
          <text class="suggestion-text">{{ analyzeResult.suggestion }}</text>
        </view>
        <text class="result-close">点击空白处关闭</text>
      </view>
    </view>

    <!-- Text Input Modal -->
    <view class="modal" v-if="showTextInput" @tap="onModalMaskTap">
      <view class="modal-content card" @tap.stop>
        <text class="modal-title">记录你吃了什么</text>
        <textarea class="text-input" :value="foodText" @input="onFoodInput"
          placeholder="例如: 一碗米饭、半斤红烧肉、一杯牛奶..." :maxlength="500"
          :auto-focus="true" />
        <view class="meal-type-picker">
          <text class="meal-label">用餐类型:</text>
          <view class="meal-options">
            <text class="meal-tag" v-for="m in mealTypes" :key="m.value"
              :class="{ active: mealType === m.value }" @tap="mealType = m.value">
              {{ m.label }}
            </text>
          </view>
        </view>
        <button class="btn-primary" @tap="submitText" :loading="analyzing">
          {{ analyzing ? 'AI分析中...' : '开始分析' }}
        </button>
      </view>
    </view>
  </view>
</template>

<script>
import { api } from '@/api/request.js'

export default {
  data() {
    return {
      summary: {},
      showTextInput: false,
      foodText: '',
      mealType: 1,
      analyzing: false,
      analyzeResult: null,
      mealTypes: [
        { label: '早餐', value: 1 },
        { label: '午餐', value: 2 },
        { label: '晚餐', value: 3 },
        { label: '加餐', value: 4 },
      ],
    }
  },
  computed: {
    greeting() {
      const h = new Date().getHours()
      if (h < 12) return '早上好'
      if (h < 18) return '下午好'
      return '晚上好'
    },
    today() {
      const d = new Date()
      return `${d.getMonth() + 1}月${d.getDate()}日`
    },
    resultTotalCal() {
      if (!this.analyzeResult || !this.analyzeResult.records) return 0
      return this.analyzeResult.records.reduce((s, r) => s + r.calories, 0).toFixed(0)
    },
  },
  onShow() {
    const token = uni.getStorageSync('token')
    if (!token) {
      uni.reLaunch({ url: '/pages/login/login' })
      return
    }
    this.loadSummary()
  },
  methods: {
    onFoodInput(e) {
      this.foodText = e.detail.value
    },
    onModalMaskTap() {
      this.showTextInput = false
    },
    async loadSummary() {
      try {
        const res = await api.getDailySummary()
        this.summary = res
      } catch (e) {
        console.log('load summary failed', e)
      }
    },
    takePhoto() {
      uni.chooseImage({
        count: 1,
        sourceType: ['camera', 'album'],
        success: (res) => {
          const filePath = res.tempFilePaths[0]
          uni.navigateTo({
            url: `/pages/analyze/analyze?image=${encodeURIComponent(filePath)}`
          })
        },
      })
    },
    inputText() {
      this.showTextInput = true
    },
    async submitText() {
      if (!this.foodText.trim()) {
        uni.showToast({ title: '请输入食物描述', icon: 'none' })
        return
      }
      this.analyzing = true
      try {
        const res = await api.analyzeText({
          description: this.foodText,
          meal_type: this.mealType,
        })
        if (res.is_food === false) {
          uni.showToast({ title: '未识别到食物，请重新输入', icon: 'none', duration: 2000 })
          return
        }
        this.showTextInput = false
        this.foodText = ''
        this.analyzeResult = { records: res.records, suggestion: res.suggestion }
        this.loadSummary()
      } catch (err) {
        uni.showToast({ title: '分析失败: ' + err.message, icon: 'none' })
      } finally {
        this.analyzing = false
      }
    },
    goHistory() {
      uni.switchTab({ url: '/pages/history/history' })
    },
    goWheel() {
      uni.switchTab({ url: '/pages/wheel/wheel' })
    },
  },
}
</script>

<style scoped>
.home-page {
  padding: 30rpx;
  min-height: 100vh;
}

.header {
  margin-bottom: 30rpx;
}

.greeting {
  font-size: 44rpx;
  font-weight: bold;
  color: #333;
  display: block;
  margin-bottom: 8rpx;
}

.date-text {
  font-size: 26rpx;
  color: #999;
}

.summary-card {
  background: linear-gradient(135deg, #4CAF50, #66BB6A);
  color: #fff;
  margin-bottom: 30rpx;
}

.summary-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20rpx;
}

.summary-title {
  font-size: 28rpx;
  opacity: 0.9;
}

.summary-link {
  font-size: 24rpx;
  opacity: 0.8;
}

.calorie-main {
  display: flex;
  align-items: baseline;
  gap: 8rpx;
  margin-bottom: 20rpx;
}

.calorie-num {
  font-size: 72rpx;
  font-weight: bold;
}

.calorie-unit {
  font-size: 28rpx;
  opacity: 0.8;
}

.nutrients-row {
  display: flex;
  gap: 40rpx;
  padding-top: 20rpx;
  border-top: 1rpx solid rgba(255,255,255,0.3);
}

.nutrient-item {
  display: flex;
  flex-direction: column;
  gap: 4rpx;
}

.nutrient-val {
  font-size: 32rpx;
  font-weight: bold;
}

.nutrient-label {
  font-size: 22rpx;
  opacity: 0.8;
}

.nutrient-val.sugar-val {
  color: #FFD54F;
}

.actions {
  display: flex;
  gap: 20rpx;
  margin-bottom: 20rpx;
}

.action-card {
  flex: 1;
  text-align: center;
  padding: 36rpx 20rpx;
}

.action-icon-wrap {
  width: 80rpx;
  height: 80rpx;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 16rpx;
}

.action-icon {
  font-size: 36rpx;
}

.action-title {
  font-size: 30rpx;
  font-weight: bold;
  color: #333;
  display: block;
  margin-bottom: 8rpx;
}

.action-desc {
  font-size: 22rpx;
  color: #999;
}

.wheel-card {
  margin-bottom: 20rpx;
}

.wheel-card-content {
  display: flex;
  align-items: center;
  gap: 20rpx;
}

.wheel-emoji {
  font-size: 48rpx;
}

.wheel-text {
  flex: 1;
}

.wheel-title {
  font-size: 30rpx;
  font-weight: bold;
  color: #333;
  display: block;
  margin-bottom: 4rpx;
}

.wheel-desc {
  font-size: 24rpx;
  color: #999;
}

.wheel-arrow {
  font-size: 40rpx;
  color: #ccc;
}

.modal {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0,0,0,0.5);
  display: flex;
  align-items: flex-end;
  z-index: 999;
}

.modal-content {
  width: 100%;
  border-radius: 30rpx 30rpx 0 0;
  padding: 40rpx;
}

.modal-title {
  font-size: 34rpx;
  font-weight: bold;
  display: block;
  margin-bottom: 20rpx;
}

.text-input {
  background: #f5f5f5;
  border-radius: 16rpx;
  padding: 24rpx;
  width: 100%;
  min-height: 200rpx;
  font-size: 28rpx;
  margin-bottom: 20rpx;
}

.meal-type-picker {
  margin-bottom: 30rpx;
}

.meal-label {
  font-size: 26rpx;
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

.result-section {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 998;
}

.result-wrap {
  width: 90%;
  max-height: 75vh;
  overflow-y: auto;
}

.result-section-title {
  font-size: 32rpx;
  font-weight: bold;
  color: #333;
}

.result-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24rpx;
}

.result-close-btn {
  width: 56rpx;
  height: 56rpx;
  line-height: 56rpx;
  text-align: center;
  font-size: 32rpx;
  color: #999;
  background: #f0f0f0;
  border-radius: 50%;
}

.food-card {
  background: #f9f9f9;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 16rpx;
}

.food-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12rpx;
}

.food-card-name {
  font-size: 30rpx;
  font-weight: bold;
  color: #333;
}

.food-card-cal {
  font-size: 28rpx;
  font-weight: bold;
  color: #FF7043;
}

.food-card-amount {
  font-size: 24rpx;
  color: #999;
  background: #f5f5f5;
  padding: 4rpx 12rpx;
  border-radius: 8rpx;
}

.sugar-chip {
  background: #fff3e0 !important;
}

.chip-val.sugar-val {
  color: #FF7043;
}

.food-card-nutrients {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
}

.nutrient-chip {
  display: flex;
  align-items: center;
  gap: 4rpx;
  background: #e8f5e9;
  padding: 6rpx 16rpx;
  border-radius: 20rpx;
}

.chip-val {
  font-size: 24rpx;
  font-weight: bold;
  color: #4CAF50;
}

.chip-label {
  font-size: 20rpx;
  color: #666;
}

.result-total {
  text-align: center;
  padding: 20rpx 0;
  border-top: 1rpx solid #f0f0f0;
  margin-top: 8rpx;
}

.total-text {
  font-size: 30rpx;
  font-weight: bold;
  color: #FF7043;
}

.result-suggestion {
  background: #fff8e1;
  border-radius: 12rpx;
  padding: 20rpx;
  margin-top: 16rpx;
}

.suggestion-label {
  font-size: 24rpx;
  color: #FFA726;
  font-weight: bold;
  display: block;
  margin-bottom: 8rpx;
}

.suggestion-text {
  font-size: 26rpx;
  color: #666;
  line-height: 1.6;
}

.result-close {
  display: block;
  text-align: center;
  font-size: 22rpx;
  color: #ccc;
  margin-top: 20rpx;
}
</style>
