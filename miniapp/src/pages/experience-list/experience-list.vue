<template>
  <view class="list-page">
    <view class="list-head"><text>用户真实记录</text><text>{{ city.name }} · 亲自吃过才值得被看见</text></view>
    <view v-if="loading && !records.length" class="list-state">正在加载真实记录…</view>
    <view v-else-if="error && !records.length" class="list-state error" @tap="loadRecords(true)">{{ error }}，点击重试</view>
    <view v-else-if="!records.length" class="list-state">这座城市还没有公开记录</view>
    <view v-for="item in records" :key="item.id" class="record-row" @tap="goDetail(item.id)">
      <image :src="item.image || images.shop" mode="aspectFill" />
      <view class="record-copy">
        <view class="user-line"><text>{{ item.author }}</text><text>{{ item.date }}</text></view>
        <text class="place-name">{{ item.place }}</text>
        <text class="record-content">{{ item.content }}</text>
        <text class="conclusion">{{ item.conclusion }}</text>
      </view>
      <text class="arrow">›</text>
    </view>
    <view v-if="loading && records.length" class="load-more">正在加载更多…</view>
    <view v-else-if="!hasMore && records.length" class="load-more">已经看完了</view>
  </view>
</template>

<script>
import { api } from '@/api/request.js'
import { diningImages } from '@/mock/city-dining.js'
import { getSelectedCity } from '@/store/city.js'

export default {
  data() { return { city: getSelectedCity(), records: [], cursor: '', hasMore: true, loading: false, error: '', images: diningImages } },
  onLoad() { this.loadRecords(true) },
  onPullDownRefresh() { this.loadRecords(true).finally(() => uni.stopPullDownRefresh()) },
  onReachBottom() { this.loadRecords(false) },
  methods: {
    async loadRecords(reset) {
      if (this.loading || (!reset && !this.hasMore)) return
      if (reset) { this.cursor = ''; this.hasMore = true; this.records = []; this.error = '' }
      this.loading = true
      try {
        const page = await api.getExperiences({ city_code: this.city.code, cursor: this.cursor, limit: 12 })
        const items = (page.items || []).map(item => {
          const version = item.version || {}
          return {
            id: item.record && item.record.id,
            place: (item.place && item.place.name) || '未命名地点',
            author: (item.author && item.author.nickname) || '城市食客',
            date: this.formatDate(version.visit_date),
            content: version.content || '一条真实到店记录',
            conclusion: ({ recommend: '推荐', neutral: '一般', caution: '谨慎选择', hot: '很喜欢' })[version.conclusion] || '真实体验',
            image: item.media && item.media[0] && item.media[0].public_url,
          }
        }).filter(item => item.id)
        this.records = reset ? items : this.records.concat(items)
        this.cursor = page.next_cursor || ''
        this.hasMore = Boolean(page.has_more)
      } catch (error) { this.error = error.message || '加载失败' } finally { this.loading = false }
    },
    formatDate(value) { const date = new Date(value); return Number.isNaN(date.getTime()) ? '最近' : `${date.getMonth() + 1}月${date.getDate()}日` },
    goDetail(id) { uni.navigateTo({ url: `/pages/experience-detail/experience-detail?id=${id}` }) },
  },
}
</script>

<style scoped>
.list-page { min-height: 100vh; box-sizing: border-box; padding: 32rpx 26rpx calc(60rpx + env(safe-area-inset-bottom)); background: #F3F3EF; color: #171714; }.list-head { display: flex; flex-direction: column; margin-bottom: 28rpx; }.list-head text:first-child { font-size: 42rpx; font-weight: 850; }.list-head text:last-child { margin-top: 8rpx; color: #858078; font-size: 23rpx; }.list-state { padding: 150rpx 20rpx; color: #8A857D; text-align: center; font-size: 25rpx; }.list-state.error { color: #B7432C; }
.record-row { display: flex; align-items: center; gap: 20rpx; margin-bottom: 18rpx; padding: 20rpx; border-radius: 24rpx; background: #FFFEFA; }.record-row image { width: 160rpx; height: 180rpx; flex: none; border-radius: 18rpx; }.record-copy { min-width: 0; display: flex; flex: 1; flex-direction: column; }.user-line { display: flex; justify-content: space-between; color: #8D877F; font-size: 20rpx; }.place-name { margin-top: 11rpx; font-size: 29rpx; font-weight: 800; }.record-content { margin-top: 8rpx; overflow: hidden; color: #5E5952; font-size: 22rpx; line-height: 1.5; text-overflow: ellipsis; white-space: nowrap; }.conclusion { align-self: flex-start; margin-top: 12rpx; padding: 6rpx 12rpx; border-radius: 9rpx; background: #EEFBCB; font-size: 19rpx; font-weight: 750; }.arrow { color: #817B73; font-size: 42rpx; }.load-more { padding: 35rpx; color: #908A82; font-size: 22rpx; text-align: center; }
</style>
