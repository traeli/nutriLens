<template>
  <view class="detail-page">
    <view v-if="loading" class="state">正在打开这份发现…</view>
    <view v-else-if="error" class="state error" @tap="loadPost">{{ error }}，点击重试</view>
    <template v-else>
      <view class="top-actions" :style="{ top: `${navTop}px` }"><view @tap="back">‹</view></view>
      <swiper v-if="post.images.length" class="gallery" circular :indicator-dots="post.images.length > 1" indicator-color="rgba(255,255,255,.5)" indicator-active-color="#FFFFFF">
        <swiper-item v-for="image in post.images" :key="image.id"><image :src="image.image_url" mode="aspectFill" /></swiper-item>
      </swiper>
      <view v-else class="gallery placeholder">饭友记编辑发现</view>
      <view class="sheet">
        <view class="meta-line"><text class="stamp">编辑发现</text><text>{{ post.date }}</text></view>
        <text class="title">{{ post.title }}</text>
        <view class="place-card" @tap="openPlace">
          <view class="place-icon">⌖</view>
          <view><text class="place-name">{{ post.place }}</text><text class="address">{{ post.address }}</text></view>
          <text class="arrow">›</text>
        </view>
        <view class="content-mark">✎</view>
        <text class="content">{{ post.content }}</text>
        <view v-if="post.source" class="source-card">
          <text class="source-label">信息来源</text><text>{{ post.source }}</text>
          <text v-if="post.sourceUrl" class="copy-source" @tap="copySource">复制来源链接</text>
        </view>
        <text class="disclaimer">内容由饭友记编辑根据公开信息整理，营业时间、菜单和价格请以商家现场为准。</text>
        <view class="action-row">
          <view class="map-action" @tap="openMap">打开地图</view>
          <view class="like-action" :class="{ active: post.liked }" @tap="toggleLike"><text>👍</text><text>{{ post.liked ? '已觉得有用' : '觉得有用' }}</text><text>{{ post.likes }}</text></view>
        </view>
      </view>
    </template>
  </view>
</template>

<script>
import { api } from '@/api/request.js'

function navTop() { const info = typeof uni.getWindowInfo === 'function' ? uni.getWindowInfo() : uni.getSystemInfoSync(); return Number(info.statusBarHeight || 20) + 10 }
export default {
  data() { return { id: '', navTop: navTop(), loading: true, error: '', post: { images: [], likes: 0, liked: false } } },
  onLoad(options) { this.id = String(options.id || ''); this.loadPost() },
  methods: {
    async loadPost() {
      if (!/^\d+$/.test(this.id)) { this.loading = false; this.error = '帖子编号无效'; return }
      this.loading = true; this.error = ''
      try {
        const item = await api.getDiscoveryPost(this.id)
        const raw = item.post || {}; const place = item.place || {}
        let liked = false
        if (uni.getStorageSync('token')) { const result = await api.getMyDiscoveryPostLikes(); liked = (result.post_ids || []).map(Number).includes(Number(this.id)) }
        this.post = { id: raw.id, title: raw.title || place.name || '编辑发现', content: raw.content || '内容正在整理。', author: raw.author_name || '饭友记编辑', source: raw.source_name || '', sourceUrl: raw.source_url || '', likes: Number(raw.like_count || 0), liked, date: this.formatDate(raw.published_at || raw.created_at), placeId: place.id, place: place.name || '店铺待补充', address: place.address || '地址待补充', longitude: Number(place.longitude), latitude: Number(place.latitude), images: (item.images || []).slice(0, 6) }
      } catch (error) { this.error = error.message || '帖子加载失败' } finally { this.loading = false }
    },
    formatDate(value) { const date = new Date(value); return Number.isNaN(date.getTime()) ? '最近整理' : `${date.getFullYear()}年${date.getMonth() + 1}月${date.getDate()}日` },
    async toggleLike() { if (!uni.getStorageSync('token')) return uni.navigateTo({ url: '/pages/login/login' }); try { if (this.post.liked) await api.unlikeDiscoveryPost(this.id); else await api.likeDiscoveryPost(this.id); this.post.liked = !this.post.liked; this.post.likes = Math.max(0, this.post.likes + (this.post.liked ? 1 : -1)) } catch (error) { uni.showToast({ title: error.message || '操作失败', icon: 'none' }) } },
    openMap() { if (!Number.isFinite(this.post.longitude) || !Number.isFinite(this.post.latitude)) return uni.showToast({ title: '店铺坐标待补充', icon: 'none' }); uni.openLocation({ longitude: this.post.longitude, latitude: this.post.latitude, name: this.post.place, address: this.post.address, scale: 18 }) },
    openPlace() { if (this.post.placeId) uni.navigateTo({ url: `/pages/place-detail/place-detail?id=${this.post.placeId}` }) },
    copySource() { uni.setClipboardData({ data: this.post.sourceUrl }) },
    back() { uni.navigateBack({ fail: () => uni.switchTab({ url: '/pages/discover/discover' }) }) },
  },
}
</script>

<style scoped>
.detail-page { min-height: 100vh; background: #FFF9EF; color: #332E26; }.state { min-height: 80vh; display: flex; align-items: center; justify-content: center; color: #847965; font-size: 25rpx; }.state.error { color: #B54D37; }.top-actions { position: fixed; left: 24rpx; z-index: 5; }.top-actions view { width: 72rpx; height: 72rpx; border-radius: 50%; background: rgba(32,28,22,.72); color: white; font-size: 52rpx; line-height: 64rpx; text-align: center; }.gallery { height: 610rpx; background: #E8DCC5; }.gallery image { width: 100%; height: 100%; }.gallery.placeholder { display: flex; align-items: center; justify-content: center; color: #88775D; font-family: "Kaiti SC", serif; font-size: 40rpx; }.sheet { position: relative; margin-top: -34rpx; padding: 38rpx 30rpx calc(70rpx + env(safe-area-inset-bottom)); border-radius: 36rpx 36rpx 0 0; background: #FFF9EF; }.meta-line { display: flex; align-items: center; justify-content: space-between; color: #958874; font-size: 21rpx; }.stamp { padding: 7rpx 13rpx; border: 2rpx solid #D86642; color: #D86642; font-weight: 800; transform: rotate(-2deg); }.title { display: block; margin-top: 24rpx; font-family: "Kaiti SC", serif; font-size: 48rpx; font-weight: 900; line-height: 1.35; }.place-card { display: flex; align-items: center; gap: 16rpx; margin-top: 26rpx; padding: 22rpx; border: 2rpx solid #E7DAC3; border-radius: 22rpx; background: #FFFDF9; }.place-icon { color: #3E9B6F; font-size: 34rpx; }.place-card view:nth-child(2) { min-width: 0; flex: 1; display: flex; flex-direction: column; }.place-name { font-size: 27rpx; font-weight: 850; }.address { margin-top: 7rpx; color: #817664; font-size: 21rpx; }.arrow { font-size: 40rpx; }.content-mark { margin-top: 36rpx; color: #D86642; font-size: 34rpx; }.content { display: block; margin-top: 10rpx; font-size: 28rpx; line-height: 1.95; white-space: pre-wrap; }.source-card { display: flex; flex-direction: column; gap: 9rpx; margin-top: 34rpx; padding: 22rpx; border-left: 6rpx solid #F1BC4B; background: #FFF1C9; color: #6D604A; font-size: 22rpx; }.source-label { font-weight: 850; }.copy-source { color: #2F7D58; font-weight: 800; }.disclaimer { display: block; margin-top: 24rpx; color: #978B78; font-size: 20rpx; line-height: 1.7; }.action-row { display: flex; gap: 16rpx; margin-top: 36rpx; }.map-action,.like-action { height: 84rpx; display: flex; align-items: center; justify-content: center; border-radius: 999rpx; font-size: 24rpx; font-weight: 850; }.map-action { width: 190rpx; border: 2rpx solid #3E9B6F; color: #2F7D58; }.like-action { flex: 1; gap: 10rpx; background: #332E26; color: white; }.like-action.active { background: #D86642; }
</style>
