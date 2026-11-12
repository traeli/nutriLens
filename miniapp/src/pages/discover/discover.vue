<template>
  <view class="discover-page">
    <view class="page-head" :style="{ paddingTop: `${navLayout.contentTop}px` }">
      <view class="head-copy">
        <text class="eyebrow">饭友记编辑整理</text>
        <text class="title">发现好吃的</text>
        <text class="subtitle">从公开信息里，慢慢挑出值得去的店</text>
      </view>
      <view class="city-pill" @tap="goHome"><text>{{ city.name || '当前城市' }}</text><text>⌄</text></view>
    </view>

    <view class="notice-card">
      <view class="notice-pin"></view>
      <text>每篇内容都由饭友记编辑统一整理，店铺营业信息请以现场为准。</text>
    </view>

    <view v-if="loading && !posts.length" class="state-card">正在翻编辑的手帐…</view>
    <view v-else-if="error && !posts.length" class="state-card error" @tap="loadPosts(true)">{{ error }}，点一下重试</view>
    <view v-else-if="!posts.length" class="state-card empty">
      <text class="empty-mark">✎</text>
      <text>这座城市的发现帖还在整理</text>
    </view>

    <view v-for="(item, index) in posts" :key="item.id" class="post-card" :class="{ tilt: index % 2 }" @tap="openPost(item.id)">
      <view v-if="item.images.length" class="photo-grid" :class="`count-${Math.min(item.images.length, 3)}`">
        <image v-for="image in item.images.slice(0, 3)" :key="image.id" :src="image.image_url" mode="aspectFill" />
        <text v-if="item.images.length > 3" class="photo-more">+{{ item.images.length - 3 }}</text>
      </view>
      <view v-else class="photo-placeholder"><text>{{ item.place.slice(0, 1) }}</text><text>图片整理中</text></view>
      <view class="post-body">
        <view class="author-line"><text class="author-stamp">编辑发现</text><text>{{ item.date }}</text></view>
        <text class="post-title">{{ item.title }}</text>
        <view class="place-line"><text class="place-pin">⌖</text><view><text class="place-name">{{ item.place }}</text><text class="place-address">{{ item.address }}</text></view></view>
        <text class="excerpt">{{ item.excerpt }}</text>
        <view class="post-foot">
          <text class="source">{{ item.source || '饭友记编辑整理' }}</text>
          <view class="like" :class="{ active: item.liked }" @tap.stop="toggleLike(item)"><text>👍</text><text>{{ item.likes || '有用' }}</text></view>
        </view>
      </view>
    </view>
    <view v-if="loading && posts.length" class="more-state">继续翻找中…</view>
    <view v-else-if="posts.length && !hasMore" class="more-state">已经看到这里啦</view>
  </view>
</template>

<script>
import { api } from '@/api/request.js'
import { getSelectedCity } from '@/store/city.js'
import { syncCustomTabBar } from '@/utils/tab-bar.js'
import { getPrimaryNavLayout } from '@/utils/nav-layout.js'

export default {
  data() {
    return { navLayout: getPrimaryNavLayout(), city: getSelectedCity(), posts: [], cursor: '', hasMore: true, loading: false, error: '', likedIDs: new Set() }
  },
  onShow() {
    this.navLayout = getPrimaryNavLayout()
    syncCustomTabBar(this, 1)
    const city = getSelectedCity()
    const changed = city.code !== this.city.code
    this.city = city
    if (changed || !this.posts.length) this.loadPosts(true)
    else this.loadLikes()
  },
  onPullDownRefresh() { this.loadPosts(true).finally(() => uni.stopPullDownRefresh()) },
  onReachBottom() { this.loadPosts(false) },
  methods: {
    async loadLikes() {
      if (!uni.getStorageSync('token')) { this.likedIDs = new Set(); this.applyLikes(); return }
      try {
        const result = await api.getMyDiscoveryPostLikes()
        this.likedIDs = new Set((result.post_ids || []).map(Number))
        this.applyLikes()
      } catch (error) { console.warn('[discover] likes load failed:', error.message) }
    },
    applyLikes() { this.posts = this.posts.map(item => ({ ...item, liked: this.likedIDs.has(Number(item.id)) })) },
    async loadPosts(reset) {
      if (this.loading || (!reset && !this.hasMore)) return
      if (reset) { this.cursor = ''; this.hasMore = true; this.error = '' }
      this.loading = true
      try {
        const page = await api.getDiscoveryPosts({ city_code: this.city.code, cursor: this.cursor, limit: 10 })
        const items = (page.items || []).map(this.normalizePost)
        this.posts = reset ? items : this.posts.concat(items)
        this.cursor = page.next_cursor || ''
        this.hasMore = Boolean(page.has_more)
        await this.loadLikes()
      } catch (error) {
        if (reset) this.posts = []
        this.error = error.message || '发现内容加载失败'
      } finally { this.loading = false }
    },
    normalizePost(item) {
      const post = item.post || {}
      const place = item.place || {}
      return { id: post.id, title: post.title || place.name || '编辑发现', excerpt: String(post.content || '').slice(0, 110), author: post.author_name || '饭友记编辑', source: post.source_name || '', likes: Number(post.like_count || 0), liked: false, date: this.formatDate(post.published_at || post.created_at), place: place.name || '店铺待补充', address: place.address || '地址待补充', images: (item.images || []).slice(0, 6) }
    },
    formatDate(value) { const date = new Date(value); return Number.isNaN(date.getTime()) ? '最近整理' : `${date.getMonth() + 1}月${date.getDate()}日` },
    async toggleLike(item) {
      if (!uni.getStorageSync('token')) return uni.navigateTo({ url: '/pages/login/login' })
      try {
        if (item.liked) await api.unlikeDiscoveryPost(item.id)
        else await api.likeDiscoveryPost(item.id)
        item.liked = !item.liked
        item.likes = Math.max(0, item.likes + (item.liked ? 1 : -1))
      } catch (error) { uni.showToast({ title: error.message || '操作失败', icon: 'none' }) }
    },
    openPost(id) { uni.navigateTo({ url: `/pages/discover-detail/discover-detail?id=${id}` }) },
    goHome() { uni.switchTab({ url: '/pages/home/home' }) },
  },
}
</script>

<style scoped>
.discover-page { min-height: 100vh; padding: 0 28rpx calc(150rpx + env(safe-area-inset-bottom)); box-sizing: border-box; background: #FFF9EF; color: #332E26; }
.page-head { display: flex; align-items: flex-start; gap: 20rpx; padding-bottom: 30rpx; }
.head-copy { flex: 1; display: flex; flex-direction: column; }.eyebrow { color: #D86642; font-size: 22rpx; font-weight: 800; letter-spacing: 2rpx; }.title { margin-top: 8rpx; font-family: "Kaiti SC", serif; font-size: 54rpx; font-weight: 900; }.subtitle { margin-top: 8rpx; color: #807664; font-size: 23rpx; }
.city-pill { display: flex; gap: 10rpx; margin-top: 8rpx; padding: 14rpx 18rpx; border: 2rpx solid #E7D8BD; border-radius: 999rpx; background: #FFFDF8; font-size: 23rpx; }
.notice-card { position: relative; margin-bottom: 28rpx; padding: 26rpx 28rpx; border: 2rpx dashed #D8C59F; border-radius: 24rpx; background: #FFF3CF; color: #6F6047; font-size: 22rpx; line-height: 1.65; transform: rotate(-.4deg); }.notice-pin { position: absolute; top: -10rpx; left: 48rpx; width: 46rpx; height: 18rpx; background: rgba(247,188,76,.75); transform: rotate(-4deg); }
.state-card,.more-state { padding: 70rpx 20rpx; color: #887E6D; text-align: center; font-size: 24rpx; }.state-card.error { color: #B64C36; }.state-card.empty { display: flex; flex-direction: column; gap: 18rpx; }.empty-mark { font-size: 54rpx; }
.post-card { overflow: hidden; margin-bottom: 34rpx; border: 3rpx solid #E9DDC6; border-radius: 30rpx; background: #FFFDF9; box-shadow: 0 8rpx 0 rgba(92,72,42,.05); }.post-card.tilt { transform: rotate(.25deg); }
.photo-grid { position: relative; display: grid; gap: 4rpx; height: 360rpx; background: #EEE5D5; }.photo-grid image { width: 100%; height: 100%; min-width: 0; min-height: 0; }.photo-grid.count-1 { grid-template-columns: 1fr; }.photo-grid.count-2 { grid-template-columns: 1fr 1fr; }.photo-grid.count-3 { grid-template-columns: 1.35fr 1fr; grid-template-rows: 1fr 1fr; }.photo-grid.count-3 image:first-child { grid-row: 1 / 3; }.photo-more { position: absolute; right: 14rpx; bottom: 14rpx; padding: 7rpx 13rpx; border-radius: 999rpx; background: rgba(30,26,20,.72); color: white; font-size: 21rpx; }
.photo-placeholder { height: 250rpx; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 12rpx; background: #F2E6CD; color: #8E7C5D; }.photo-placeholder text:first-child { font-family: "Kaiti SC", serif; font-size: 72rpx; }
.post-body { padding: 27rpx; }.author-line { display: flex; align-items: center; justify-content: space-between; color: #9A8E7A; font-size: 20rpx; }.author-stamp { padding: 6rpx 12rpx; border: 2rpx solid #D86642; color: #D86642; font-weight: 800; transform: rotate(-2deg); }.post-title { display: block; margin-top: 20rpx; font-family: "Kaiti SC", serif; font-size: 38rpx; font-weight: 900; line-height: 1.35; }
.place-line { display: flex; gap: 12rpx; margin-top: 20rpx; padding: 18rpx; border-radius: 18rpx; background: #F8F0E2; }.place-pin { color: #3E9B6F; font-size: 28rpx; }.place-line view { display: flex; flex-direction: column; min-width: 0; }.place-name { font-size: 25rpx; font-weight: 800; }.place-address { margin-top: 6rpx; overflow: hidden; color: #817664; font-size: 21rpx; text-overflow: ellipsis; white-space: nowrap; }
.excerpt { display: block; margin-top: 20rpx; color: #554E42; font-size: 25rpx; line-height: 1.75; }.post-foot { display: flex; align-items: center; gap: 18rpx; margin-top: 24rpx; padding-top: 20rpx; border-top: 2rpx dashed #E7DCC8; }.source { flex: 1; color: #9A8E7A; font-size: 20rpx; }.like { display: flex; align-items: center; gap: 8rpx; padding: 12rpx 18rpx; border-radius: 999rpx; background: #F4ECDD; color: #6F6452; font-size: 22rpx; font-weight: 800; }.like.active { background: #FFE0D5; color: #C84F34; }
</style>
