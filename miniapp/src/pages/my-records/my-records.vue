<template>
  <view class="records-page">
    <view class="records-heading">
      <text>我的记录</text>
      <text>每一次提交，都能在这里看到审核进度</text>
    </view>

    <view class="status-summary">
      <view><text>{{ statusCounts.pending_review }}</text><text>待审核</text></view>
      <view><text>{{ statusCounts.published }}</text><text>审核通过</text></view>
      <view><text>{{ statusCounts.rejected }}</text><text>审核不通过</text></view>
    </view>

    <scroll-view class="filter-scroll" scroll-x>
      <view class="filter-row">
        <text v-for="item in filters" :key="item.value" :class="{ active: selectedStatus === item.value }" @tap="selectedStatus = item.value">{{ item.label }}</text>
      </view>
    </scroll-view>

    <view v-if="loading" class="state-card">正在整理你的记录…</view>
    <view v-else-if="!filteredRecords.length" class="state-card empty">
      <text>这里还没有记录</text>
      <text>提交一条到店体验后，审核状态会出现在这里。</text>
      <button @tap="startRecord">去记录</button>
    </view>

    <view v-else class="record-list">
      <view v-for="item in filteredRecords" :key="item.id" class="record-card" @tap="openRecord(item)">
        <view class="record-top">
          <view class="place-mark">{{ item.place.slice(0, 1) }}</view>
          <view class="record-copy"><text>{{ item.place }}</text><text>{{ item.city }} · {{ item.date }}</text></view>
          <text class="status-pill" :class="`status-${item.status}`">{{ item.statusLabel }}</text>
        </view>
        <view class="record-bottom"><text>{{ item.conclusion }}</text><text>{{ item.statusHint }}</text><text>›</text></view>
      </view>
    </view>
  </view>
</template>

<script>
import { api } from '@/api/request.js'

export default {
  data() {
    return {
      loading: false,
      records: [],
      selectedStatus: '',
      filters: [
        { value: '', label: '全部' },
        { value: 'pending_review', label: '待审核' },
        { value: 'published', label: '审核通过' },
        { value: 'rejected', label: '审核不通过' },
      ],
    }
  },
  computed: {
    filteredRecords() { return this.selectedStatus ? this.records.filter(item => item.status === this.selectedStatus) : this.records },
    statusCounts() {
      return this.records.reduce((result, item) => { result[item.status] = (result[item.status] || 0) + 1; return result }, { pending_review: 0, published: 0, rejected: 0 })
    },
  },
  onLoad(options) {
    const status = String((options && options.status) || '')
    if (this.filters.some(item => item.value === status)) this.selectedStatus = status
  },
  onShow() { this.loadRecords() },
  onPullDownRefresh() { this.loadRecords().finally(() => uni.stopPullDownRefresh()) },
  methods: {
    async loadRecords() {
      if (!uni.getStorageSync('token')) { uni.navigateTo({ url: '/pages/login/login' }); return }
      this.loading = true
      try {
        const page = await api.getMyRecords({ limit: 50 })
        this.records = (page.items || []).filter(item => ['pending_review', 'published', 'rejected'].includes(item.record.publish_status)).map(this.normalizeRecord)
      } catch (error) {
        uni.showToast({ title: error.message || '记录加载失败', icon: 'none' })
      } finally { this.loading = false }
    },
    normalizeRecord(item) {
      const status = item.record.publish_status
      const labels = { pending_review: '待审核', published: '审核通过', rejected: '审核不通过' }
      const hints = { pending_review: '内容已提交，等待审核', published: '已经公开展示', rejected: '查看原因并修改后重新提交' }
      const conclusions = { hot: '夯 · 值得专程去', recommend: '推荐 · 愿意再来', neutral: '一般 · 如实记录', caution: '谨慎 · 建议留意' }
      const date = String(item.version.visit_date || '').slice(0, 10)
      return {
        id: item.record.id, status, statusLabel: labels[status], statusHint: hints[status],
        place: item.place.name || '未命名餐厅', city: item.place.city_code || '当前城市',
        date: date ? date.replace(/^\d{4}-/, '').replace('-', '月') + '日' : '最近提交',
        conclusion: conclusions[item.version.conclusion] || '真实到店记录',
      }
    },
    openRecord(item) {
      const url = item.status === 'published' ? `/pages/experience-detail/experience-detail?id=${item.id}` : `/pages/review-status/review-status?id=${item.id}`
      uni.navigateTo({ url })
    },
    startRecord() { uni.switchTab({ url: '/pages/record/record' }) },
  },
}
</script>

<style scoped>
.records-page { min-height: 100vh; box-sizing: border-box; padding: 30rpx 24rpx 70rpx; background: #f5faec; color: #171914; }
.records-heading { display: flex; flex-direction: column; padding: 10rpx 5rpx 24rpx; }
.records-heading text:first-child { font-family: "Kaiti SC", STKaiti, cursive; font-size: 60rpx; font-weight: 900; line-height: 1; }
.records-heading text:last-child { margin-top: 13rpx; color: #737b70; font-size: 20rpx; }
.status-summary { display: grid; grid-template-columns: repeat(3, 1fr); gap: 11rpx; }
.status-summary view { min-width: 0; display: flex; flex-direction: column; padding: 24rpx 18rpx; border-radius: 27rpx; background: #fff; }
.status-summary view:first-child { background: #11120f; color: #fff; }
.status-summary view:nth-child(2) { background: #c7ff35; }
.status-summary text:first-child { font-size: 43rpx; font-weight: 950; line-height: 1; }
.status-summary text:last-child { margin-top: 10rpx; color: #747b70; font-size: 17rpx; font-weight: 750; white-space: nowrap; }
.status-summary view:first-child text:last-child { color: rgba(255,255,255,.58); }
.filter-scroll { width: calc(100vw - 24rpx); margin-top: 26rpx; white-space: nowrap; }
.filter-row { display: inline-flex; gap: 10rpx; padding-right: 24rpx; }
.filter-row text { padding: 13rpx 22rpx; border-radius: 28rpx; background: #e7ece1; color: #62695f; font-size: 19rpx; font-weight: 750; }
.filter-row text.active { background: #11120f; color: #c7ff35; }
.state-card { margin-top: 18rpx; padding: 50rpx 25rpx; border-radius: 30rpx; background: #fff; color: #737a70; text-align: center; font-size: 21rpx; }
.state-card.empty { display: flex; flex-direction: column; align-items: center; }
.state-card.empty text:first-child { color: #171914; font-family: "Kaiti SC", STKaiti, cursive; font-size: 34rpx; font-weight: 900; }
.state-card.empty text:nth-child(2) { margin-top: 9rpx; }
.state-card button { margin-top: 25rpx; padding: 0 38rpx; border-radius: 35rpx; background: #11120f; color: #c7ff35; font-size: 21rpx; }
.record-list { margin-top: 18rpx; }
.record-card { padding: 21rpx; border-radius: 30rpx; background: #fff; }
.record-card + .record-card { margin-top: 12rpx; }
.record-top { display: flex; align-items: center; gap: 15rpx; }
.place-mark { width: 68rpx; height: 68rpx; flex: none; border-radius: 21rpx; background: #11120f; color: #c7ff35; font-family: "Kaiti SC", STKaiti, cursive; font-size: 30rpx; font-weight: 900; line-height: 68rpx; text-align: center; }
.record-copy { min-width: 0; display: flex; flex: 1; flex-direction: column; }
.record-copy text:first-child { overflow: hidden; font-size: 24rpx; font-weight: 850; text-overflow: ellipsis; white-space: nowrap; }
.record-copy text:last-child { margin-top: 5rpx; color: #7e857b; font-size: 17rpx; }
.status-pill { flex: none; padding: 8rpx 13rpx; border-radius: 18rpx; font-size: 17rpx; font-weight: 800; }
.status-pending_review { background: #fff0c6; color: #8a6411; }
.status-published { background: #e1f6d8; color: #387425; }
.status-rejected { background: #f9e1dc; color: #b23e2b; }
.record-bottom { display: grid; grid-template-columns: auto 1fr auto; gap: 12rpx; margin-top: 18rpx; padding-top: 17rpx; border-top: 1rpx solid #edf0e9; align-items: center; }
.record-bottom text:first-child { font-size: 19rpx; font-weight: 800; }
.record-bottom text:nth-child(2) { overflow: hidden; color: #7e857b; font-size: 17rpx; text-align: right; text-overflow: ellipsis; white-space: nowrap; }
.record-bottom text:last-child { color: #777e74; font-size: 31rpx; }
</style>
