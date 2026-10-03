<template>
  <view class="nutrition-page">
    <view class="intro">
      <text class="intro-title">饮食记录</text>
      <text class="intro-copy">写下吃了什么，为你估算这一餐的营养成分。记录仅自己可见，不参与餐厅评价。</text>
    </view>

    <view class="entry-card">
      <picker :range="periodLabels" @change="mealPeriod = periods[$event.detail.value]">
        <view class="period-row"><text>用餐时段</text><text>{{ periodName }} ›</text></view>
      </picker>
      <textarea
        v-model="description"
        class="food-input"
        maxlength="1000"
        placeholder="输入饮食，例如：我吃了一个苹果"
        placeholder-class="input-placeholder"
        @input="analysis = null"
      />
      <button class="analyze-button" :disabled="saving || Boolean(analysis)" :loading="saving" @tap="analyze">
        {{ analysis ? '已完成并保存' : '分析营养成分' }}
      </button>
    </view>

    <view v-if="analysis" class="analysis-card">
      <view class="card-heading">
        <view><text class="eyebrow">本餐参考估算</text><text class="food-title">{{ description.trim() }}</text></view>
        <text class="energy">{{ formatNumber(analysis.calories) }}<text> kcal</text></text>
      </view>
      <view class="food-list">
        <view v-for="food in analysis.foods" :key="food.name" class="food-row">
          <text>{{ food.name }}</text><text>{{ food.amount }}</text>
        </view>
      </view>
      <view class="nutrients">
        <view><text>{{ formatNumber(analysis.protein_grams) }}g</text><text>蛋白质</text></view>
        <view><text>{{ formatNumber(analysis.fat_grams) }}g</text><text>脂肪</text></view>
        <view><text>{{ formatNumber(analysis.carbohydrate_grams) }}g</text><text>碳水</text></view>
      </view>
      <view v-if="analysis.advice" class="advice-box">
        <text>饮食建议</text>
        <text>{{ analysis.advice }}</text>
      </view>
      <text class="estimate-note">营养数据由智能模型估算并已保存，仅供日常记录参考。</text>
    </view>

    <view class="summary">
      <view><text>累计记录</text><text>{{ summary.meal_count || 0 }} 餐</text></view>
      <view><text>累计热量</text><text>{{ formatNumber(summary.calories || 0) }} kcal</text></view>
    </view>

    <view v-if="records.length" class="history-title">最近记录</view>
    <view v-for="item in records" :key="item.id" class="record-card">
      <view class="record-heading">
        <view><text>{{ labelOf(item.meal_period) }} · {{ formatDate(item.eaten_at) }}</text><text>{{ item.description || '未填写描述' }}</text></view>
        <text class="delete" @tap="remove(item.id)">删除</text>
      </view>
      <view class="record-energy"><text>{{ hasNutrition(item) ? formatNumber(item.calories) : '—' }}</text><text> kcal</text></view>
      <view class="record-nutrients">
        <view><text>{{ nutrientValue(item.protein_grams) }}</text><text>蛋白质</text></view>
        <view><text>{{ nutrientValue(item.fat_grams) }}</text><text>脂肪</text></view>
        <view><text>{{ nutrientValue(item.carbohydrate_grams) }}</text><text>碳水</text></view>
      </view>
      <view v-if="item.advice" class="record-advice"><text>建议</text><text>{{ item.advice }}</text></view>
    </view>
    <view v-if="!records.length" class="empty"><text>还没有饮食记录</text><text>写下今天吃的食物，开始第一条记录吧。</text></view>
  </view>
</template>

<script>
import { api } from '@/api/request.js'

export default {
  data() {
    return {
      periods: ['breakfast', 'lunch', 'dinner', 'snack'],
      periodLabels: ['早餐', '午餐', '晚餐', '加餐'],
      mealPeriod: 'lunch',
      description: '',
      analysis: null,
      saving: false,
      records: [],
      summary: {},
    }
  },
  computed: {
    periodName() { return this.labelOf(this.mealPeriod) },
  },
  onShow() { this.load() },
  methods: {
    labelOf(value) { return ({ breakfast: '早餐', lunch: '午餐', dinner: '晚餐', snack: '加餐' })[value] || '用餐' },
    formatDate(value) {
      const date = new Date(value)
      return Number.isNaN(date.getTime()) ? '' : `${date.getMonth() + 1}月${date.getDate()}日`
    },
    formatNumber(value) {
      const number = Number(value)
      if (!Number.isFinite(number)) return '0'
      return Number.isInteger(number) ? String(number) : number.toFixed(1)
    },
    nutrientValue(value) { return value == null ? '—' : `${this.formatNumber(value)}g` },
    hasNutrition(item) { return item.calories != null },
    async load() {
      if (!uni.getStorageSync('token')) return
      try {
        const [list, summary] = await Promise.all([api.getNutritionRecords({ limit: 50 }), api.getNutritionSummary()])
        this.records = list.items || []
        this.summary = summary || {}
      } catch (error) {
        uni.showToast({ title: error.message || '加载失败', icon: 'none' })
      }
    },
    async analyze() {
      const description = this.description.trim()
      if (!description) return uni.showToast({ title: '请先输入吃了什么', icon: 'none' })
      this.saving = true
      try {
        const result = await api.createNutritionRecord({
          meal_period: this.mealPeriod,
          eaten_at: new Date().toISOString(),
          description,
          foods: [],
        })
        this.analysis = result
        await this.load()
        uni.showToast({ title: '分析完成并已保存', icon: 'success' })
      } catch (error) {
        uni.showToast({ title: error.message || '营养分析失败', icon: 'none' })
      } finally {
        this.saving = false
      }
    },
    remove(id) {
      uni.showModal({
        title: '删除这条记录？',
        success: async result => {
          if (!result.confirm) return
          try {
            await api.deleteNutritionRecord(id)
            await this.load()
          } catch (error) {
            uni.showToast({ title: error.message || '删除失败', icon: 'none' })
          }
        },
      })
    },
  },
}
</script>

<style scoped>
.nutrition-page { box-sizing: border-box; min-height: 100vh; padding: 28rpx 28rpx calc(64rpx + env(safe-area-inset-bottom)); background: #f5faec; color: #171914; }
.intro, .entry-card, .analysis-card, .summary, .record-card { margin-bottom: 20rpx; padding: 28rpx; border-radius: 24rpx; background: #fff; }
.intro-title { display: block; font-size: 38rpx; font-weight: 800; }
.intro-copy { display: block; margin-top: 12rpx; color: #747a6d; font-size: 24rpx; line-height: 1.6; }
.period-row { display: flex; justify-content: space-between; padding: 18rpx 4rpx 24rpx; border-bottom: 1rpx solid #edf0e8; font-size: 28rpx; }
.period-row text:last-child { color: #62685b; }
.food-input { box-sizing: border-box; width: 100%; height: 180rpx; margin-top: 22rpx; padding: 22rpx; border: 1rpx solid #dce2d2; border-radius: 18rpx; background: #fbfcf9; font-size: 29rpx; line-height: 1.55; }
.input-placeholder { color: #a0a69a; }
.analyze-button { height: 88rpx; margin-top: 22rpx; border: 0; border-radius: 16rpx; background: #c7ff35; color: #111; font-size: 31rpx; font-weight: 700; line-height: 88rpx; }
.analyze-button::after { border: 0; }
.card-heading, .record-heading { display: flex; align-items: flex-start; justify-content: space-between; gap: 20rpx; }
.card-heading > view, .record-heading > view { display: flex; min-width: 0; flex: 1; flex-direction: column; }
.eyebrow { color: #77806e; font-size: 22rpx; }
.food-title { margin-top: 9rpx; overflow: hidden; font-size: 31rpx; font-weight: 750; text-overflow: ellipsis; white-space: nowrap; }
.energy { flex: none; font-size: 42rpx; font-weight: 800; }
.energy text { font-size: 20rpx; font-weight: 600; }
.food-list { margin-top: 24rpx; padding: 18rpx 20rpx; border-radius: 16rpx; background: #f6f8f2; }
.food-row { display: flex; justify-content: space-between; color: #676e61; font-size: 23rpx; line-height: 1.8; }
.food-row text:first-child { color: #282b25; font-weight: 650; }
.nutrients, .record-nutrients { display: flex; margin-top: 24rpx; }
.nutrients view, .record-nutrients view { display: flex; flex: 1; flex-direction: column; align-items: center; border-right: 1rpx solid #edf0e8; }
.nutrients view:last-child, .record-nutrients view:last-child { border-right: 0; }
.nutrients view text:first-child, .record-nutrients view text:first-child { font-size: 29rpx; font-weight: 750; }
.nutrients view text:last-child, .record-nutrients view text:last-child { margin-top: 7rpx; color: #848a7e; font-size: 21rpx; }
.advice-box { display: flex; margin-top: 26rpx; padding: 22rpx; border-radius: 16rpx; background: #f3f8e8; flex-direction: column; }
.advice-box text:first-child { color: #586843; font-size: 22rpx; font-weight: 750; }
.advice-box text:last-child { margin-top: 10rpx; color: #3f4737; font-size: 25rpx; line-height: 1.65; }
.estimate-note { display: block; margin-top: 24rpx; color: #92988c; font-size: 20rpx; line-height: 1.5; }
.summary { display: flex; padding-top: 24rpx; padding-bottom: 24rpx; }
.summary view { display: flex; flex: 1; flex-direction: column; }
.summary view:last-child { align-items: flex-end; }
.summary view text:first-child { color: #7d8377; font-size: 21rpx; }
.summary view text:last-child { margin-top: 5rpx; font-size: 29rpx; font-weight: 750; }
.history-title { margin: 34rpx 6rpx 18rpx; font-size: 28rpx; font-weight: 750; }
.record-heading view text:first-child { color: #777e71; font-size: 22rpx; }
.record-heading view text:last-child { margin-top: 8rpx; overflow: hidden; font-size: 29rpx; font-weight: 700; text-overflow: ellipsis; white-space: nowrap; }
.delete { flex: none; color: #aa5544; font-size: 22rpx; }
.record-energy { margin-top: 24rpx; }
.record-energy text:first-child { font-size: 42rpx; font-weight: 800; }
.record-energy text:last-child { font-size: 20rpx; font-weight: 600; }
.record-advice { display: flex; margin-top: 24rpx; padding-top: 20rpx; border-top: 1rpx solid #edf0e8; gap: 14rpx; color: #62695c; font-size: 23rpx; line-height: 1.55; }
.record-advice text:first-child { flex: none; color: #72805e; font-weight: 700; }
.empty { display: flex; padding: 90rpx 30rpx; flex-direction: column; align-items: center; color: #858b7c; text-align: center; }
.empty text:first-child { color: #555b50; font-size: 28rpx; font-weight: 700; }
.empty text:last-child { margin-top: 12rpx; font-size: 23rpx; }
</style>
