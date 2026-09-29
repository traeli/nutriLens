<template>
  <view class="nutrition-page">
    <view class="intro"><text>私人营养记录</text><text>仅自己可见，不参与餐厅评价。当前支持手动记录，识别结果也可写入同一数据结构。</text></view>
    <view class="form">
      <picker :range="periodLabels" @change="mealPeriod = periods[$event.detail.value]"><view>用餐时段：{{ periodName }}</view></picker>
      <textarea v-model="description" maxlength="1000" placeholder="写下这一餐吃了什么" />
      <view class="numbers"><input v-model="calories" type="digit" placeholder="热量 kcal" /><input v-model="protein" type="digit" placeholder="蛋白质 g" /></view>
      <view class="numbers"><input v-model="fat" type="digit" placeholder="脂肪 g" /><input v-model="carbs" type="digit" placeholder="碳水 g" /></view>
      <button :loading="saving" @tap="save">保存私人记录</button>
    </view>
    <view class="summary"><text>累计 {{ summary.meal_count || 0 }} 餐</text><text>{{ Number(summary.calories || 0).toFixed(0) }} kcal</text></view>
    <view v-for="item in records" :key="item.id" class="record">
      <view><text>{{ labelOf(item.meal_period) }} · {{ formatDate(item.eaten_at) }}</text><text>{{ item.description || '未填写描述' }}</text></view>
      <view><text>{{ item.calories == null ? '—' : item.calories }} kcal</text><text @tap="remove(item.id)">删除</text></view>
    </view>
    <view v-if="!records.length" class="empty">还没有营养记录。</view>
  </view>
</template>

<script>
import { api } from '@/api/request.js'
export default {
  data() { return { periods: ['breakfast', 'lunch', 'dinner', 'snack'], periodLabels: ['早餐', '午餐', '晚餐', '加餐'], mealPeriod: 'lunch', description: '', calories: '', protein: '', fat: '', carbs: '', saving: false, records: [], summary: {} } },
  computed: { periodName() { return this.labelOf(this.mealPeriod) } },
  onShow() { this.load() },
  methods: {
    labelOf(value) { return ({ breakfast: '早餐', lunch: '午餐', dinner: '晚餐', snack: '加餐' })[value] || '用餐' },
    formatDate(value) { const date = new Date(value); return Number.isNaN(date.getTime()) ? '' : `${date.getMonth() + 1}月${date.getDate()}日` },
    number(value) { return value === '' ? null : Number(value) },
    async load() { if (!uni.getStorageSync('token')) return; try { const [list, summary] = await Promise.all([api.getNutritionRecords({ limit: 50 }), api.getNutritionSummary()]); this.records = list.items || []; this.summary = summary || {} } catch (error) { uni.showToast({ title: error.message || '加载失败', icon: 'none' }) } },
    async save() { if (!this.description.trim()) return uni.showToast({ title: '请先填写食物描述', icon: 'none' }); this.saving = true; try { await api.createNutritionRecord({ meal_period: this.mealPeriod, eaten_at: new Date().toISOString(), description: this.description.trim(), foods: [], calories: this.number(this.calories), protein_grams: this.number(this.protein), fat_grams: this.number(this.fat), carbohydrate_grams: this.number(this.carbs) }); this.description = ''; this.calories = this.protein = this.fat = this.carbs = ''; await this.load(); uni.showToast({ title: '已保存', icon: 'success' }) } catch (error) { uni.showToast({ title: error.message || '保存失败', icon: 'none' }) } finally { this.saving = false } },
    remove(id) { uni.showModal({ title: '删除这条记录？', success: async result => { if (!result.confirm) return; try { await api.deleteNutritionRecord(id); await this.load() } catch (error) { uni.showToast({ title: error.message || '删除失败', icon: 'none' }) } } }) },
  },
}
</script>

<style scoped>
.nutrition-page { min-height: 100vh; padding: 28rpx; background: #F5FAEC; color: #171914; }
.intro, .form, .record, .summary { margin-bottom: 20rpx; padding: 28rpx; border-radius: 24rpx; background: #fff; }
.intro, .intro text, .record view { display: flex; flex-direction: column; }
.intro text:first-child { font-size: 38rpx; font-weight: 800; }.intro text:last-child { margin-top: 12rpx; color: #747a6d; font-size: 22rpx; line-height: 1.6; }
.form textarea, .form input, .form picker { box-sizing: border-box; width: 100%; margin-top: 16rpx; padding: 20rpx; border: 1rpx solid #dce2d2; border-radius: 16rpx; }
.form textarea { height: 180rpx; }.numbers { display: flex; gap: 14rpx; }.form button { margin-top: 20rpx; background: #C7FF35; color: #111; }
.summary { display: flex; justify-content: space-between; font-weight: 700; }.record { display: flex; justify-content: space-between; }.record view:first-child text:last-child { margin-top: 8rpx; color: #676c61; }.record view:last-child { align-items: flex-end; }.record view:last-child text:last-child { margin-top: 15rpx; color: #b24634; }.empty { padding: 80rpx; color: #858b7c; text-align: center; }
</style>
