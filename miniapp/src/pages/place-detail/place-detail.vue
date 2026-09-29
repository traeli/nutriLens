<template>
  <view class="place-page">
    <view v-if="loading" class="page-state">正在打开店铺…</view>
    <view v-else-if="error" class="page-state error" @tap="loadPlace">{{ error }}，点击重试</view>
    <template v-else>
      <image class="place-hero" :src="place.image_url || images.shop" mode="aspectFill" />
      <view class="place-card">
        <view class="title-line"><view><text class="place-name">{{ place.name }}</text><text class="place-meta">{{ categoryName }} · {{ areaName }}</text></view><view class="place-buttons"><button @tap="toggleFavorite">{{ favorited ? '已收藏' : '收藏' }}</button><button @tap="openMap">地图</button></view></view>
        <text class="address">⌖ {{ place.address || '地址待补充' }}</text>
        <view class="stats">
          <view><text>{{ place.experience_count || 0 }}</text><text>真实记录</text></view>
          <view><text>{{ place.recommend_count || 0 }}</text><text>推荐体验</text></view>
          <view><text>{{ averageCost }}</text><text>参考人均</text></view>
        </view>
        <text class="section-title">店铺简介</text>
        <text class="description">{{ place.description || '这家店的简介还在完善中。下方内容均来自用户亲自到店后的真实记录。' }}</text>
      </view>

      <view class="reviews">
        <view class="review-head"><text>全部真实评论</text><text>{{ place.experience_count || 0 }} 条</text></view>
        <view v-if="reviewsLoading && !reviews.length" class="review-state">正在加载真实记录…</view>
        <view v-else-if="reviewsError && !reviews.length" class="review-state error" @tap="loadReviews(true)">{{ reviewsError }}，点击重试</view>
        <view v-else-if="!reviews.length" class="review-state">暂时还没有公开评论</view>
        <view v-for="item in reviews" :key="item.id" class="review-card" @tap="goExperience(item.id)">
          <view class="review-user"><text class="avatar">{{ item.author.slice(0, 1) }}</text><text>{{ item.author }}</text><text>{{ item.date }}</text></view>
          <text class="review-conclusion">{{ item.conclusion }}</text>
          <text class="review-content">{{ item.content }}</text>
          <view class="review-foot"><text v-if="item.cost">人均 ¥{{ item.cost }}</text><text>查看详情 ›</text></view>
        </view>
        <view v-if="reviewsLoading && reviews.length" class="loading-more">正在加载更多…</view>
        <view v-else-if="!hasMore && reviews.length" class="loading-more">已经看完全部真实评论</view>
      </view>
    </template>
  </view>
</template>

<script>
import { api } from '@/api/request.js'
import { diningImages } from '@/mock/city-dining.js'

export default {
  data() {
    return { id: '', place: {}, reviews: [], cursor: '', hasMore: true, loading: true, error: '', reviewsLoading: false, reviewsError: '', images: diningImages, favorited: false }
  },
  computed: {
    categoryName() { return ({ restaurant: '餐厅', stall: '小吃摊', night_market: '夜市档口', drink: '饮品店', other: '餐饮地点' })[this.place.category] || '餐饮地点' },
    areaName() { return this.place.business_area || this.place.district || '本城' },
    averageCost() { return this.place.average_cost == null ? '—' : `¥${Math.round(Number(this.place.average_cost))}` },
  },
  onLoad(options) {
    this.id = String(options.id || '')
    this.loadPlace()
  },
  onReachBottom() { this.loadReviews(false) },
  methods: {
    async loadPlace() {
      if (!/^\d+$/.test(this.id)) { this.loading = false; this.error = '店铺编号无效'; return }
      this.loading = true
      this.error = ''
      try {
        this.place = await api.getPlace(this.id)
        if (uni.getStorageSync('token')) {
          const favorites = await api.getFavoritePlaces()
          this.favorited = (favorites.items || []).some(item => Number(item.id) === Number(this.id))
        }
        await this.loadReviews(true)
      } catch (error) {
        this.error = error.message || '店铺加载失败'
      } finally { this.loading = false }
    },
    async loadReviews(reset) {
      if (this.reviewsLoading || (!reset && !this.hasMore)) return
      if (reset) { this.cursor = ''; this.hasMore = true; this.reviews = []; this.reviewsError = '' }
      this.reviewsLoading = true
      try {
        const page = await api.getPlaceExperiences(this.id, { cursor: this.cursor, limit: 10 })
        const items = (page.items || []).map(item => {
          const version = item.version || {}
          return {
            id: item.record && item.record.id,
            author: (item.author && item.author.nickname) || '城市食客',
            date: this.formatDate(version.visit_date),
            conclusion: ({ recommend: '推荐', neutral: '一般', caution: '谨慎选择', hot: '很喜欢' })[version.conclusion] || '真实体验',
            content: version.content || '这位用户留下了一条真实到店记录。',
            cost: version.average_cost == null ? '' : Math.round(Number(version.average_cost)),
          }
        }).filter(item => item.id)
        this.reviews = reset ? items : this.reviews.concat(items)
        this.cursor = page.next_cursor || ''
        this.hasMore = Boolean(page.has_more)
      } catch (error) {
        this.reviewsError = error.message || '评论加载失败'
      } finally { this.reviewsLoading = false }
    },
    formatDate(value) {
      const date = new Date(value)
      return Number.isNaN(date.getTime()) ? '最近到店' : `${date.getFullYear()}年${date.getMonth() + 1}月${date.getDate()}日`
    },
    openMap() {
      const longitude = Number(this.place.longitude)
      const latitude = Number(this.place.latitude)
      if (!Number.isFinite(longitude) || !Number.isFinite(latitude) || (!longitude && !latitude)) return uni.showToast({ title: '这家店暂无有效坐标', icon: 'none' })
      uni.openLocation({ longitude, latitude, name: this.place.name, address: this.place.address || '', scale: 18 })
    },
    goExperience(id) { uni.navigateTo({ url: `/pages/experience-detail/experience-detail?id=${id}` }) },
    async toggleFavorite() { try { if (this.favorited) await api.unfavoritePlace(this.id); else await api.favoritePlace(this.id); this.favorited = !this.favorited } catch (error) { uni.showToast({ title: error.message || '操作失败', icon: 'none' }) } },
  },
}
</script>

<style scoped>
.place-page { min-height: 100vh; padding-bottom: calc(50rpx + env(safe-area-inset-bottom)); background: #F3F3EF; color: #171714; }
.page-state { min-height: 75vh; display: flex; align-items: center; justify-content: center; color: #817C74; font-size: 26rpx; }.page-state.error,.review-state.error { color: #B34A35; }
.place-hero { width: 100%; height: 390rpx; display: block; background: #DDD9D1; }
.place-card { margin: -32rpx 24rpx 0; position: relative; padding: 34rpx 30rpx; border-radius: 32rpx; background: #FFFEFA; box-shadow: 0 10rpx 30rpx rgba(30,26,20,.08); }
.title-line { display: flex; align-items: flex-start; justify-content: space-between; gap: 20rpx; }.title-line > view { min-width: 0; display: flex; flex-direction: column; }.place-name { font-size: 43rpx; font-weight: 850; }.place-meta { margin-top: 10rpx; color: #777169; font-size: 23rpx; }
.title-line button { width: 150rpx; height: 66rpx; margin: 0; padding: 0; border: 0; border-radius: 34rpx; background: #C7FF35; color: #151612; font-size: 23rpx; font-weight: 800; line-height: 66rpx; }
.place-buttons { display: flex; gap: 10rpx; }.place-buttons button { width: 112rpx; }
.address { display: block; margin-top: 25rpx; color: #5F5A53; font-size: 23rpx; line-height: 1.6; }.stats { display: grid; grid-template-columns: repeat(3,1fr); margin: 30rpx 0; padding: 25rpx 0; border-top: 1rpx solid #E4E0D9; border-bottom: 1rpx solid #E4E0D9; }.stats view { display: flex; flex-direction: column; align-items: center; border-right: 1rpx solid #E4E0D9; }.stats view:last-child { border: 0; }.stats view text:first-child { font-size: 31rpx; font-weight: 850; }.stats view text:last-child { margin-top: 7rpx; color: #88827A; font-size: 20rpx; }
.section-title { font-size: 28rpx; font-weight: 800; }.description { display: block; margin-top: 13rpx; color: #5D5851; font-size: 24rpx; line-height: 1.8; }
.reviews { padding: 38rpx 24rpx; }.review-head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20rpx; }.review-head text:first-child { font-size: 34rpx; font-weight: 850; }.review-head text:last-child { color: #8A847C; font-size: 23rpx; }.review-state,.loading-more { padding: 65rpx 15rpx; color: #89847D; font-size: 24rpx; text-align: center; }
.review-card { position: relative; margin-bottom: 18rpx; padding: 25rpx; border-radius: 24rpx; background: #FFFEFA; }.review-user { display: flex; align-items: center; gap: 12rpx; color: #37342F; font-size: 23rpx; }.review-user text:last-child { margin-left: auto; color: #989189; font-size: 20rpx; }.avatar { width: 48rpx; height: 48rpx; border-radius: 50%; background: #E5E2DA; line-height: 48rpx; text-align: center; }.review-conclusion { display: inline-block; margin-top: 20rpx; padding: 7rpx 13rpx; border-radius: 10rpx; background: #EEFBCB; font-size: 20rpx; font-weight: 800; }.review-content { display: block; margin-top: 14rpx; font-size: 25rpx; line-height: 1.7; }.review-foot { display: flex; justify-content: space-between; margin-top: 18rpx; color: #8B857D; font-size: 21rpx; }.review-foot text:last-child { margin-left: auto; color: #B7432C; font-weight: 700; }
</style>
