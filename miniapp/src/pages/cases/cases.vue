<template>
  <view class="cases-page">
    <text class="title">举报与申诉</text>
    <view v-for="item in combined" :key="item.key" class="case-card">
      <view><text>{{ item.type }}</text><text>{{ item.status }}</text></view>
      <text>{{ item.text }}</text><text>{{ item.date }}</text>
    </view>
    <view v-if="!combined.length" class="empty">暂无举报或申诉记录</view>
  </view>
</template>
<script>
import { api } from '@/api/request.js'
export default {
  data() { return { reports: [], appeals: [] } },
  computed: { combined() { return [...this.reports.map(item => ({ key: 'r' + item.id, type: '内容举报', status: item.status, text: item.description || item.reason_code, date: this.date(item.created_at) })), ...this.appeals.map(item => ({ key: 'a' + item.id, type: '用户申诉', status: item.status, text: item.claim_text, date: this.date(item.created_at) }))] } },
  onShow() { this.load() },
  methods: { date(value) { return String(value || '').slice(0, 10) }, async load() { try { const [reports, appeals] = await Promise.all([api.getMyReports(), api.getMyAppeals()]); this.reports = reports.items || []; this.appeals = appeals.items || [] } catch (error) { uni.showToast({ title: error.message || '加载失败', icon: 'none' }) } } },
}
</script>
<style scoped>
.cases-page { min-height: 100vh; padding: 30rpx; background: #F5FAEC; }.title { display: block; margin: 15rpx 0 30rpx; font-size: 44rpx; font-weight: 800; }.case-card { margin-bottom: 18rpx; padding: 26rpx; border-radius: 22rpx; background: #fff; }.case-card view { display: flex; justify-content: space-between; font-weight: 700; }.case-card > text { display: block; margin-top: 12rpx; color: #686e62; }.case-card > text:last-child { font-size: 20rpx; }.empty { padding: 100rpx 20rpx; color: #858b7c; text-align: center; }
</style>
