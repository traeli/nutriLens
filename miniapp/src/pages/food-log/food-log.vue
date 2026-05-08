<template>
  <view class="food-log-page">
    <!-- Date Range Selector -->
    <view class="range-bar card">
      <view class="range-tabs">
        <text class="range-tab" :class="{ active: rangeType === 'week' }" @tap="setRange('week')">本周</text>
        <text class="range-tab" :class="{ active: rangeType === 'month' }" @tap="setRange('month')">本月</text>
        <text class="range-tab" :class="{ active: rangeType === 'day' }" @tap="setRange('day')">按日期</text>
      </view>
      <picker mode="date" fields="month" :value="selectedMonth" @change="onMonthChange" v-if="rangeType !== 'day'">
        <text class="month-picker">{{ displayMonth }}</text>
      </picker>
      <picker mode="date" :value="selectedDate" @change="onDateChange" v-else>
        <text class="month-picker">{{ displayDate }}</text>
      </picker>
    </view>

    <!-- Monthly Stats -->
    <view class="card stats-card" v-if="summary.avg_calories">
      <text class="stats-title">月度统计</text>
      <view class="stats-main">
        <view class="stats-avg">
          <text class="stats-avg-num">{{ summary.avg_calories.toFixed(0) }}</text>
          <text class="stats-avg-unit">日均 kcal</text>
        </view>
        <text class="stats-days">记录 {{ summary.days_with_records }}/{{ summary.days_in_month }} 天</text>
      </view>
      <view class="stats-nutrients" v-if="summary.total_nutrients">
        <view class="stats-nutrient">
          <text class="sn-val">{{ avgNutrient('protein').toFixed(1) }}</text>
          <text class="sn-label">日均蛋白质g</text>
        </view>
        <view class="stats-nutrient">
          <text class="sn-val">{{ avgNutrient('carbs').toFixed(1) }}</text>
          <text class="sn-label">日均碳水g</text>
        </view>
        <view class="stats-nutrient">
          <text class="sn-val">{{ avgNutrient('fat').toFixed(1) }}</text>
          <text class="sn-label">日均脂肪g</text>
        </view>
        <view class="stats-nutrient">
          <text class="sn-val sugar">{{ avgNutrient('sugar').toFixed(1) }}</text>
          <text class="sn-label">日均糖分g</text>
        </view>
      </view>
    </view>

    <!-- Meal Filter -->
    <scroll-view class="meal-filter" scroll-x>
      <text class="meal-tag" :class="{ active: mealFilter === 0 }" @tap="mealFilter = 0">全部</text>
      <text class="meal-tag" :class="{ active: mealFilter === mt.value }"
        v-for="mt in mealTypes" :key="mt.value" @tap="mealFilter = mt.value">
        {{ mt.label }}
      </text>
    </scroll-view>

    <!-- Day Timeline -->
    <view class="timeline">
      <view class="day-group" v-for="day in filteredDays" :key="day.date">
        <view class="day-header">
          <text class="day-date">{{ formatDateLabel(day.date) }}</text>
          <text class="day-week">{{ day.weekday }}</text>
        </view>
        <view class="day-meals">
          <view class="meal-row" v-for="mt in mealTypes" :key="mt.value">
            <view v-if="getMealRecords(day, mt.value).length > 0">
              <view class="meal-record" v-for="r in getMealRecords(day, mt.value)" :key="r.id">
                <text class="meal-icon">{{ mt.icon }}</text>
                <text class="meal-label-text">{{ mt.label }}</text>
                <text class="meal-food">{{ r.food_name }}</text>
                <text class="meal-weight" v-if="r.unit_amount">{{ (r.unit_amount * 1000).toFixed(0) }}g</text>
                <text class="meal-cal">{{ r.calories }} kcal</text>
                <text class="meal-delete" @tap="deleteRecord(r)">删除</text>
              </view>
            </view>
            <view class="meal-missing" v-else>
              <text class="meal-icon">{{ mt.icon }}</text>
              <text class="meal-label-text">{{ mt.label }}</text>
              <text class="missing-tip">未记录</text>
            </view>
          </view>
        </view>
      </view>

      <view class="empty" v-if="filteredDays.length === 0">
        <text class="empty-text">本月暂无饮食记录</text>
      </view>
    </view>
  </view>
</template>

<script>
import { api } from '@/api/request.js'

export default {
  data() {
    return {
      rangeType: 'month',
      selectedMonth: '',
      selectedDate: '',
      mealFilter: 0,
      summary: {},
      allRecords: [],
      mealTypes: [
        { label: '早餐', value: 1, icon: '☀' },
        { label: '午餐', value: 2, icon: '🌤' },
        { label: '晚餐', value: 3, icon: '🌙' },
        { label: '加餐', value: 4, icon: '🍵' },
      ],
    }
  },
  computed: {
    displayMonth() {
      if (!this.selectedMonth) return ''
      const parts = this.selectedMonth.split('-')
      return parts[0] + '年' + parseInt(parts[1]) + '月'
    },
    displayDate() {
      if (!this.selectedDate) return '选择日期'
      const parts = this.selectedDate.split('-')
      return parseInt(parts[1]) + '月' + parseInt(parts[2]) + '日'
    },
    filteredDays() {
      let days = this.groupedByDay
      if (this.rangeType === 'day' && this.selectedDate) {
        days = days.filter(d => d.date === this.selectedDate)
      }
      if (this.mealFilter > 0) {
        days = days.filter(d => d.meals && d.meals[this.mealFilter])
      }
      return days
    },
    groupedByDay() {
      const map = {}
      const weekdays = ['日', '一', '二', '三', '四', '五', '六']
      for (const r of this.allRecords) {
        const key = r.created_at.substring(0, 10)
        if (!map[key]) {
          const d = new Date(key)
          map[key] = {
            date: key,
            weekday: '周' + weekdays[d.getDay()],
            meals: {},
          }
        }
        const mt = r.meal_type
        if (!map[key].meals[mt]) map[key].meals[mt] = []
        map[key].meals[mt].push(r)
      }
      return Object.values(map).sort((a, b) => b.date.localeCompare(a.date))
    },
  },
  onLoad() {
    const now = new Date()
    this.selectedMonth = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
    this.selectedDate = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')}`
    this.loadData()
  },
  methods: {
    setRange(type) {
      this.rangeType = type
      this.loadData()
    },
    onMonthChange(e) {
      this.selectedMonth = e.detail.value
      this.loadData()
    },
    onDateChange(e) {
      this.selectedDate = e.detail.value
      const dateMonth = this.selectedDate.substring(0, 7)
      if (dateMonth !== this.selectedMonth) {
        this.selectedMonth = dateMonth
        this.loadData()
      }
    },
    async loadData() {
      const month = this.rangeType === 'day' && this.selectedDate
        ? this.selectedDate.substring(0, 7)
        : this.selectedMonth
      try {
        const res = await api.getMonthlySummary(month)
        this.summary = res
        this.allRecords = res.records || []
      } catch (e) {
        console.log('load monthly summary failed', e)
      }
    },
    avgNutrient(key) {
      if (!this.summary.total_nutrients || !this.summary.days_with_records) return 0
      return (this.summary.total_nutrients[key] || 0) / this.summary.days_with_records
    },
    getMealRecords(day, mealType) {
      return (day.meals && day.meals[mealType]) || []
    },
    sumMealCal(day, mealType) {
      const records = this.getMealRecords(day, mealType)
      return records.reduce((sum, r) => sum + r.calories, 0).toFixed(0)
    },
    formatDateLabel(dateStr) {
      const parts = dateStr.split('-')
      return `${parseInt(parts[1])}月${parseInt(parts[2])}日`
    },
    deleteRecord(record) {
      uni.showModal({
        title: '删除记录',
        content: `确定删除 "${record.food_name}" 吗？`,
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
  },
}
</script>

<style scoped>
.food-log-page {
  padding: 20rpx 30rpx;
  min-height: 100vh;
  background: #f5f5f5;
}

.range-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20rpx 30rpx;
  margin-bottom: 20rpx;
}

.range-tabs {
  display: flex;
  gap: 16rpx;
}

.range-tab {
  padding: 10rpx 24rpx;
  border-radius: 30rpx;
  font-size: 24rpx;
  color: #666;
  background: #f0f0f0;
}

.range-tab.active {
  background: #4CAF50;
  color: #fff;
}

.month-picker {
  font-size: 28rpx;
  font-weight: bold;
  color: #333;
  padding: 10rpx 20rpx;
}

.stats-card {
  margin-bottom: 20rpx;
}

.stats-title {
  font-size: 26rpx;
  color: #999;
  display: block;
  margin-bottom: 16rpx;
}

.stats-main {
  display: flex;
  align-items: baseline;
  gap: 20rpx;
  margin-bottom: 20rpx;
}

.stats-avg-num {
  font-size: 56rpx;
  font-weight: bold;
  color: #4CAF50;
}

.stats-avg-unit {
  font-size: 24rpx;
  color: #999;
}

.stats-days {
  font-size: 24rpx;
  color: #999;
}

.stats-nutrients {
  display: flex;
  gap: 10rpx;
  padding-top: 16rpx;
  border-top: 1rpx solid #f0f0f0;
}

.stats-nutrient {
  flex: 1;
  text-align: center;
}

.sn-val {
  font-size: 28rpx;
  font-weight: bold;
  color: #333;
  display: block;
}

.sn-label {
  font-size: 20rpx;
  color: #999;
}

.meal-filter {
  white-space: nowrap;
  margin-bottom: 20rpx;
}

.meal-tag {
  display: inline-block;
  padding: 10rpx 24rpx;
  border-radius: 30rpx;
  background: #fff;
  font-size: 24rpx;
  color: #666;
  margin-right: 12rpx;
}

.meal-tag.active {
  background: #4CAF50;
  color: #fff;
}

.timeline {
  /* timeline */
}

.day-group {
  margin-bottom: 24rpx;
}

.day-header {
  display: flex;
  align-items: baseline;
  gap: 12rpx;
  margin-bottom: 12rpx;
  padding-left: 8rpx;
}

.day-date {
  font-size: 28rpx;
  font-weight: bold;
  color: #333;
}

.day-week {
  font-size: 22rpx;
  color: #999;
}

.day-meals {
  background: #fff;
  border-radius: 16rpx;
  overflow: hidden;
}

.meal-row {
  padding: 20rpx 24rpx;
  border-bottom: 1rpx solid #f5f5f5;
}

.meal-row:last-child {
  border-bottom: none;
}

.meal-record {
  display: flex;
  align-items: center;
  gap: 12rpx;
}

.meal-icon {
  font-size: 28rpx;
}

.meal-label-text {
  font-size: 24rpx;
  color: #999;
  width: 70rpx;
}

.meal-food {
  flex: 1;
  font-size: 24rpx;
  color: #333;
  background: #e8f5e9;
  padding: 4rpx 14rpx;
  border-radius: 8rpx;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.meal-weight {
  font-size: 20rpx;
  color: #999;
  background: #f5f5f5;
  padding: 2rpx 10rpx;
  border-radius: 6rpx;
}

.sn-val.sugar {
  color: #FF7043;
}

.meal-cal {
  font-size: 24rpx;
  color: #FF7043;
  font-weight: bold;
}

.meal-delete {
  font-size: 22rpx;
  color: #FF7043;
  padding: 4rpx 12rpx;
}

.meal-missing {
  display: flex;
  align-items: center;
  gap: 12rpx;
}

.missing-tip {
  font-size: 22rpx;
  color: #FFA726;
  background: #FFF8E1;
  padding: 4rpx 14rpx;
  border-radius: 8rpx;
}

.empty {
  text-align: center;
  padding: 80rpx 0;
}

.empty-text {
  font-size: 28rpx;
  color: #ccc;
}
</style>
