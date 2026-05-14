<template>
  <view class="history-page">
    <!-- Date Selector -->
    <view class="date-bar">
      <text class="date-arrow" @tap="prevDay">‹</text>
      <text class="date-text" @tap="showDatePicker = true">{{ currentDate }}</text>
      <text class="date-arrow" @tap="nextDay">›</text>
    </view>

    <!-- Summary -->
    <view class="card summary" v-if="summary.total_calories">
      <text class="summary-title">今日摄入 {{ summary.total_calories }} kcal</text>
      <view class="nutrients-bar" v-if="summary.total_nutrients">
        <view class="nutrient-chip">
          <text>蛋白质 {{ (summary.total_nutrients.protein || 0).toFixed(1) }}g</text>
        </view>
        <view class="nutrient-chip">
          <text>碳水 {{ (summary.total_nutrients.carbs || 0).toFixed(1) }}g</text>
        </view>
        <view class="nutrient-chip">
          <text>脂肪 {{ (summary.total_nutrients.fat || 0).toFixed(1) }}g</text>
        </view>
        <view class="nutrient-chip">
          <text>膳食纤维 {{ (summary.total_nutrients.fiber || 0).toFixed(1) }}g</text>
        </view>
        <view class="nutrient-chip sugar-chip">
          <text>糖分 {{ (summary.total_nutrients.sugar || 0).toFixed(1) }}g</text>
        </view>
        <view class="nutrient-chip">
          <text>维C {{ (summary.total_nutrients.vitamin_c || 0).toFixed(1) }}mg</text>
        </view>
      </view>
    </view>

    <!-- Records List -->
    <view class="records">
      <view class="record-card card" v-for="group in groupedRecords" :key="group.id"
        @longpress="confirmDeleteGroup(group)">
        <view class="record-header">
          <text class="meal-type">{{ mealTypeLabel(group.mealType) }}</text>
          <text class="record-time">{{ formatTime(group.time) }}</text>
        </view>
        <view class="record-body">
          <image v-if="group.imageURL" :src="resolveImageUrl(group.imageURL)" class="record-image" mode="aspectFill" />
          <view class="record-info">
            <view class="food-item" v-for="item in group.items" :key="item.id">
              <view class="food-main-row">
                <text class="food-name">{{ item.food_name }}</text>
                <text class="food-weight" v-if="item.unit_amount">{{ (item.unit_amount * 1000).toFixed(0) }}g</text>
                <text class="food-cal">{{ item.calories }} kcal</text>
              </view>
              <view class="food-nutrients-row" v-if="item.nutrients">
                <text class="nutrient-mini" v-if="item.nutrients.protein">蛋白质 {{ item.nutrients.protein.toFixed(1) }}g</text>
                <text class="nutrient-mini" v-if="item.nutrients.carbs">碳水 {{ item.nutrients.carbs.toFixed(1) }}g</text>
                <text class="nutrient-mini" v-if="item.nutrients.fat">脂肪 {{ item.nutrients.fat.toFixed(1) }}g</text>
                <text class="nutrient-mini" v-if="item.nutrients.fiber">膳食纤维 {{ item.nutrients.fiber.toFixed(1) }}g</text>
                <text class="nutrient-mini sugar" v-if="item.nutrients.sugar">糖分 {{ item.nutrients.sugar.toFixed(1) }}g</text>
                <text class="nutrient-mini" v-if="item.nutrients.vitamin_c">VitC {{ item.nutrients.vitamin_c.toFixed(1) }}mg</text>
              </view>
            </view>
            <text class="food-desc" v-if="group.description">{{ group.description }}</text>
          </view>
        </view>
        <view class="record-total" v-if="group.items.length > 1">
          <text class="total-label">合计</text>
          <text class="total-cal">{{ group.totalCal.toFixed(0) }} kcal</text>
        </view>
        <view class="record-suggestion" v-if="group.suggestion">
          <text class="suggestion-label">AI建议</text>
          <text class="suggestion-text">{{ group.suggestion }}</text>
        </view>
      </view>

      <view class="empty" v-if="records.length === 0">
        <text class="empty-text">暂无饮食记录</text>
        <text class="empty-hint">去首页拍照或文字记录吧</text>
      </view>

      <!-- AI Daily Analysis -->
      <view class="ai-section" v-if="records.length > 0">
        <button class="ai-btn" @tap="toggleAnalysis" :loading="aiLoading" v-if="!dailyAnalysis || !showAnalysis">
          {{ aiLoading ? 'AI 分析中...' : '查看今日健康建议' }}
        </button>
        <view class="card ai-card" v-if="dailyAnalysis && showAnalysis">
          <view class="ai-card-header">
            <text class="ai-card-title">今日健康建议</text>
            <text class="ai-card-collapse" @tap="showAnalysis = false">收起</text>
          </view>
          <!-- Structured JSON display -->
          <view v-if="parsedAnalysis">
            <view class="ai-block">
              <view class="ai-block-header">
                <text class="ai-block-icon">🍎</text>
                <text class="ai-block-title">营养评估</text>
              </view>
              <text class="ai-block-text">{{ parsedAnalysis.nutrition_eval }}</text>
            </view>
            <view class="ai-block">
              <view class="ai-block-header">
                <text class="ai-block-icon">🥗</text>
                <text class="ai-block-title">饮食建议</text>
              </view>
              <text class="ai-block-text">{{ parsedAnalysis.diet_advice }}</text>
            </view>
            <view class="ai-block">
              <view class="ai-block-header">
                <text class="ai-block-icon">🏃</text>
                <text class="ai-block-title">运动建议</text>
              </view>
              <text class="ai-block-text">{{ parsedAnalysis.exercise_advice }}</text>
            </view>
          </view>
          <!-- Fallback for old plain text -->
          <text class="ai-block-text" v-else>{{ dailyAnalysis.analysis_text }}</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script>
import { api, SERVER_URL } from '@/api/request.js'

export default {
  data() {
    return {
      records: [],
      summary: {},
      dailyAnalysis: null,
      showAnalysis: false,
      aiLoading: false,
      currentDate: '',
      showDatePicker: false,
      baseUrl: SERVER_URL,
    }
  },
  onShow() {
    if (!uni.getStorageSync('token')) {
      uni.reLaunch({ url: '/pages/login/login' })
      return
    }
    this.currentDate = this.formatDate(new Date())
    this.loadData()
  },
  computed: {
    groupedRecords() {
      const map = {}
      const result = []
      for (const r of this.records) {
        const gid = r.group_id || String(r.id)
        if (!map[gid]) {
          map[gid] = {
            id: gid,
            mealType: r.meal_type,
            time: r.created_at,
            imageURL: r.image_url,
            description: r.description,
            suggestion: r.ai_suggestion,
            items: [],
            totalCal: 0,
          }
          result.push(map[gid])
        }
        map[gid].items.push(r)
        map[gid].totalCal += r.calories
      }
      return result
    },
    parsedAnalysis() {
      if (!this.dailyAnalysis || !this.dailyAnalysis.analysis_text) return null
      try {
        const obj = JSON.parse(this.dailyAnalysis.analysis_text)
        if (obj.nutrition_eval || obj.diet_advice || obj.exercise_advice) return obj
      } catch (e) {}
      return null
    },
  },
  methods: {
    resolveImageUrl(url) {
      if (!url) return ''
      if (url.startsWith('http')) return url
      return this.baseUrl + url
    },
    async loadData() {
      try {
        const [recordsRes, summaryRes] = await Promise.all([
          api.getFoodRecords({ date: this.currentDate }),
          api.getDailySummary(this.currentDate),
        ])
        this.records = recordsRes.records || []
        this.summary = summaryRes
      } catch (e) {
        console.log('load data failed', e)
      }
    },
    prevDay() {
      const d = new Date(this.currentDate)
      d.setDate(d.getDate() - 1)
      this.currentDate = this.formatDate(d)
      this.loadData()
    },
    nextDay() {
      const d = new Date(this.currentDate)
      d.setDate(d.getDate() + 1)
      if (d <= new Date()) {
        this.currentDate = this.formatDate(d)
        this.loadData()
      }
    },
    mealTypeLabel(type) {
      const labels = { 1: '早餐', 2: '午餐', 3: '晚餐', 4: '加餐' }
      return labels[type] || '其他'
    },
    formatDate(d) {
      const y = d.getFullYear()
      const m = String(d.getMonth() + 1).padStart(2, '0')
      const day = String(d.getDate()).padStart(2, '0')
      return `${y}-${m}-${day}`
    },
    formatTime(dateStr) {
      const d = new Date(dateStr)
      return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
    },
    async toggleAnalysis() {
      if (this.dailyAnalysis) {
        this.showAnalysis = !this.showAnalysis
        return
      }
      this.aiLoading = true
      try {
        const res = await api.getDailyAnalysis(this.currentDate)
        this.dailyAnalysis = res
        this.showAnalysis = true
      } catch (e) {
        console.log('load daily analysis failed', e)
        uni.showToast({ title: '分析失败，请稍后再试', icon: 'none' })
      } finally {
        this.aiLoading = false
      }
    },
    async loadDailyAnalysis() {
      this.aiLoading = true
      try {
        const res = await api.getDailyAnalysis(this.currentDate)
        this.dailyAnalysis = res
        this.showAnalysis = true
      } catch (e) {
        console.log('load daily analysis failed', e)
        uni.showToast({ title: '分析失败，请稍后再试', icon: 'none' })
      } finally {
        this.aiLoading = false
      }
    },
    confirmDelete(record) {
      uni.showModal({
        title: '删除记录',
        content: `确定删除 "${record.food_name}" 的记录吗？`,
        success: async (res) => {
          if (res.confirm) {
            try {
              await api.deleteFoodRecord(record.id)
              this.loadData()
              uni.showToast({ title: '已删除', icon: 'success' })
            } catch (err) {
              uni.showToast({ title: '删除失败', icon: 'none' })
            }
          }
        },
      })
    },
    confirmDeleteGroup(group) {
      const names = group.items.map(i => i.food_name).join('、')
      uni.showModal({
        title: '删除记录',
        content: `确定删除 "${names}" 的记录吗？`,
        success: async (res) => {
          if (res.confirm) {
            try {
              for (const item of group.items) {
                await api.deleteFoodRecord(item.id)
              }
              this.loadData()
              uni.showToast({ title: '已删除', icon: 'success' })
            } catch (err) {
              uni.showToast({ title: '删除失败', icon: 'none' })
            }
          }
        },
      })
    },
  },
}
</script>

<style scoped>
.history-page {
  padding: 20rpx 30rpx;
  min-height: 100vh;
}

.date-bar {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 30rpx;
  padding: 20rpx 0;
  margin-bottom: 20rpx;
}

.date-text {
  font-size: 32rpx;
  font-weight: bold;
  color: #333;
}

.date-arrow {
  font-size: 40rpx;
  color: #4CAF50;
  padding: 10rpx 20rpx;
}

.summary {
  background: linear-gradient(135deg, #4CAF50, #66BB6A);
  color: #fff;
  margin-bottom: 20rpx;
}

.summary-title {
  font-size: 30rpx;
  font-weight: bold;
  display: block;
  margin-bottom: 16rpx;
}

.nutrients-bar {
  display: flex;
  gap: 16rpx;
  flex-wrap: wrap;
}

.nutrient-chip {
  background: rgba(255,255,255,0.2);
  padding: 8rpx 20rpx;
  border-radius: 20rpx;
  font-size: 22rpx;
}

.sugar-chip {
  background: rgba(255, 112, 67, 0.3);
}

.record-card {
  margin-bottom: 20rpx;
}

.record-header {
  display: flex;
  justify-content: space-between;
  margin-bottom: 16rpx;
}

.meal-type {
  font-size: 24rpx;
  color: #4CAF50;
  font-weight: bold;
}

.record-time {
  font-size: 24rpx;
  color: #999;
}

.record-body {
  display: flex;
  gap: 20rpx;
}

.record-image {
  width: 120rpx;
  height: 120rpx;
  border-radius: 12rpx;
  flex-shrink: 0;
}

.record-info {
  flex: 1;
}

.food-item {
  padding: 8rpx 0;
}

.food-item + .food-item {
  border-top: 1rpx dashed #f0f0f0;
}

.food-main-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6rpx;
}

.food-name {
  font-size: 30rpx;
  font-weight: bold;
  color: #333;
  flex: 1;
}

.food-weight {
  font-size: 22rpx;
  color: #999;
  background: #f5f5f5;
  padding: 2rpx 10rpx;
  border-radius: 6rpx;
  margin: 0 12rpx;
}

.food-cal {
  font-size: 28rpx;
  color: #FF7043;
  flex-shrink: 0;
}

.food-nutrients-row {
  display: flex;
  gap: 10rpx;
  flex-wrap: wrap;
  margin-top: 6rpx;
}

.nutrient-mini {
  font-size: 20rpx;
  color: #666;
  background: #e8f5e9;
  padding: 4rpx 12rpx;
  border-radius: 12rpx;
}

.nutrient-mini.sugar {
  background: #fff3e0;
  color: #FF7043;
}

.food-desc {
  font-size: 24rpx;
  color: #999;
  display: block;
}

.record-total {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 12rpx;
  padding-top: 12rpx;
  border-top: 1rpx solid #eee;
}

.total-label {
  font-size: 24rpx;
  color: #999;
}

.total-cal {
  font-size: 28rpx;
  font-weight: bold;
  color: #FF7043;
}

.record-suggestion {
  margin-top: 16rpx;
  padding-top: 16rpx;
  border-top: 1rpx solid #f0f0f0;
}

.suggestion-label {
  font-size: 22rpx;
  color: #4CAF50;
  display: block;
  margin-bottom: 8rpx;
}

.suggestion-text {
  font-size: 24rpx;
  color: #666;
  line-height: 1.6;
}

.empty {
  text-align: center;
  padding: 100rpx 0;
}

.empty-text {
  font-size: 30rpx;
  color: #ccc;
  display: block;
  margin-bottom: 10rpx;
}

.empty-hint {
  font-size: 24rpx;
  color: #ddd;
}

.ai-section {
  margin-top: 30rpx;
  padding-bottom: 40rpx;
}

.ai-btn {
  background: linear-gradient(135deg, #FF8F00, #FFA726);
  color: #fff;
  font-size: 28rpx;
  border-radius: 16rpx;
  border: none;
}

.ai-card {
  background: #fffdf5;
}

.ai-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20rpx;
}

.ai-card-title {
  font-size: 28rpx;
  font-weight: bold;
  color: #FF8F00;
}

.ai-card-collapse {
  font-size: 24rpx;
  color: #FFA726;
  padding: 6rpx 20rpx;
  background: rgba(255, 167, 38, 0.15);
  border-radius: 16rpx;
}

.ai-block {
  margin-bottom: 20rpx;
}

.ai-block:last-child {
  margin-bottom: 0;
}

.ai-block-header {
  display: flex;
  align-items: center;
  gap: 8rpx;
  margin-bottom: 8rpx;
}

.ai-block-icon {
  font-size: 28rpx;
}

.ai-block-title {
  font-size: 26rpx;
  font-weight: bold;
  color: #5D4037;
}

.ai-block-text {
  font-size: 26rpx;
  color: #666;
  line-height: 1.6;
  padding-left: 36rpx;
}
</style>
