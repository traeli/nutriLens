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
          <image v-if="group.imageURL" :src="baseUrl + group.imageURL" class="record-image" mode="aspectFill" />
          <view class="record-info">
            <view class="food-item" v-for="item in group.items" :key="item.id">
              <text class="food-name">{{ item.food_name }}</text>
              <text class="food-weight" v-if="item.unit_amount">{{ (item.unit_amount * 1000).toFixed(0) }}g</text>
              <text class="food-cal">{{ item.calories }} kcal</text>
            </view>
            <view class="food-nutrients-row" v-if="group.totalSugar > 0 || group.totalVitC > 0">
              <text class="nutrient-mini sugar" v-if="group.totalSugar > 0">糖分 {{ group.totalSugar.toFixed(1) }}g</text>
              <text class="nutrient-mini" v-if="group.totalVitC > 0">VitC {{ group.totalVitC.toFixed(1) }}mg</text>
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
    </view>
  </view>
</template>

<script>
import { api } from '@/api/request.js'

export default {
  data() {
    return {
      records: [],
      summary: {},
      currentDate: '',
      showDatePicker: false,
      baseUrl: 'http://localhost:8080',
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
            totalSugar: 0,
            totalVitC: 0,
          }
          result.push(map[gid])
        }
        map[gid].items.push(r)
        map[gid].totalCal += r.calories
        if (r.nutrients) {
          map[gid].totalSugar += r.nutrients.sugar || 0
          map[gid].totalVitC += r.nutrients.vitamin_c || 0
        }
      }
      return result
    },
  },
  methods: {
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
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8rpx 0;
}

.food-item + .food-item {
  border-top: 1rpx dashed #f0f0f0;
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

.food-weight {
  font-size: 22rpx;
  color: #999;
  background: #f5f5f5;
  padding: 2rpx 10rpx;
  border-radius: 6rpx;
}

.food-nutrients-row {
  display: flex;
  gap: 12rpx;
  margin-top: 8rpx;
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

.food-name {
  font-size: 30rpx;
  font-weight: bold;
  color: #333;
  display: block;
  margin-bottom: 8rpx;
}

.food-cal {
  font-size: 28rpx;
  color: #FF7043;
  display: block;
  margin-bottom: 4rpx;
}

.food-desc {
  font-size: 24rpx;
  color: #999;
  display: block;
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
</style>
