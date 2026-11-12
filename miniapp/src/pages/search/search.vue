<template>
  <view class="search-page page-shell">
    <view class="search-field surface">
      <text class="search-symbol">⌕</text>
      <input v-model="keyword" class="search-input" placeholder="搜索店名、小吃或商圈" confirm-type="search" @input="onKeywordInput" @confirm="loadPlaces" />
      <text v-if="keyword" class="clear" @tap="clearKeyword">×</text>
    </view>
    <text class="search-hint">可搜索店名、小吃名或商圈</text>

    <view class="city-line"><text class="pin">●</text><text>当前城市</text><text class="city">{{ city.name }}</text></view>

    <view class="result-head"><view><text class="result-title">{{ recommendedMode ? '推荐店铺' : '匹配的店铺' }}</text><text class="count"> {{ places.length }} 家</text></view><button class="filter-btn" @tap="showFilter">▽ 筛选</button></view>

    <view class="place-list">
      <view v-if="loading" class="result-state">正在搜索这座城市…</view>
      <view v-else-if="error" class="result-state error" @tap="loadPlaces">{{ error }}，点击重试</view>
      <view v-else-if="!places.length" class="result-state">没有找到匹配地点</view>
      <view v-for="place in places" :key="place.id" class="place-row" @tap="choosePlace(place)">
        <image class="place-image" :src="place.image" mode="aspectFill" />
        <view class="place-copy">
          <text class="place-name">{{ place.name }}</text>
          <text class="place-meta">{{ place.category }} · {{ place.area }} · {{ place.distance }}</text>
          <text class="experience-count">▢ {{ place.count }} 篇发现帖</text>
        </view>
        <text class="row-arrow">›</text>
      </view>
    </view>

    <view v-if="related.length" class="related-head"><text>相关发现帖</text><text class="muted">编辑整理</text></view>
    <view v-if="related.length" class="related-grid">
      <view v-for="item in related" :key="item.id" class="related-card surface" @tap="goDetail(item.id)">
        <image :src="item.image" mode="aspectFill" />
        <text>{{ item.title }}</text>
      </view>
    </view>
  </view>
</template>

<script>
import { api } from '@/api/request.js'
import { diningImages } from '@/mock/city-dining.js'
import { getSelectedCity } from '@/store/city.js'
export default {
  data() { return { keyword: '', places: [], related: [], images: diningImages, city: getSelectedCity(), loading: false, error: '', recommendedMode: false, searchTimer: null } },
  onLoad(options) {
    if (options.keyword) this.keyword = decodeURIComponent(options.keyword)
    this.recommendedMode = options.recommended === '1'
    if (this.recommendedMode) uni.setNavigationBarTitle({ title: '每日推荐' })
    this.loadPlaces()
    this.loadRelated()
  },
  onShow() {
    const city = getSelectedCity()
    if (city.code !== this.city.code) { this.city = city; this.loadPlaces(); this.loadRelated() }
  },
  onUnload() { if (this.searchTimer) clearTimeout(this.searchTimer) },
  methods: {
    onKeywordInput() {
      if (this.searchTimer) clearTimeout(this.searchTimer)
      this.searchTimer = setTimeout(() => this.loadPlaces(), 320)
    },
    clearKeyword() { this.keyword = ''; this.loadPlaces() },
    async loadPlaces() {
      this.loading = true
      this.error = ''
      try {
        const page = await api.searchPlaces({ city_code: this.city.code, q: this.keyword.trim(), recommended: this.recommendedMode ? 1 : '', limit: 30 })
        this.places = (page.items || []).map((item, index) => ({
          ...item,
          category: ({ restaurant: '餐厅', stall: '小吃摊', night_market: '夜市档口', drink: '饮品店' })[item.category] || '餐饮地点',
          area: item.business_area || item.district || '本城',
          distance: item.address || '地址待补充',
          count: Number(item.experience_count || 0),
          image: index % 3 === 0 ? diningImages.noodles : index % 3 === 1 ? diningImages.shop : diningImages.cafe,
        }))
      } catch (error) {
        this.places = []
        this.error = error.message || '搜索失败'
      } finally { this.loading = false }
    },
    async loadRelated() {
      try {
        const page = await api.getDiscoveryPosts({ city_code: this.city.code, limit: 2 })
        this.related = (page.items || []).map((item, index) => ({ id: item.post.id, title: `${item.place.name} · ${String(item.post.content || '饭友记编辑发现').slice(0, 24)}`, image: (item.images && item.images[0] && item.images[0].image_url) || (index ? diningImages.cafe : diningImages.noodles) }))
      } catch (error) { this.related = [] }
    },
    choosePlace(place) {
      uni.navigateTo({ url: `/pages/place-detail/place-detail?id=${place.id}` })
    },
    goDetail(id) { uni.navigateTo({ url: `/pages/discover-detail/discover-detail?id=${id}` }) },
    showFilter() { uni.showToast({ title: '筛选面板将在数据接入阶段完善', icon: 'none' }) },
  },
}
</script>

<style scoped>
.search-page { padding-top: 24rpx; }
.search-field { height: 94rpx; display: flex; align-items: center; padding: 0 22rpx; border-color: #BDB8B0; }
.search-symbol { font-size: 50rpx; margin-right: 14rpx; transform: rotate(-20deg); }
.search-input { flex: 1; font-size: 30rpx; }
.clear { width: 44rpx; height: 44rpx; border-radius: 50%; background: #ABA8A3; color: #fff; line-height: 40rpx; text-align: center; font-size: 38rpx; }
.search-hint { display: block; margin: 14rpx 4rpx 34rpx; color: #8D8880; font-size: 23rpx; }
.city-line { display: flex; align-items: center; gap: 16rpx; padding: 12rpx 4rpx 34rpx; font-size: 27rpx; }
.pin { color: #44413D; font-size: 20rpx; }
.city { margin-left: 8rpx; font-size: 31rpx; font-weight: 600; }
.result-head, .related-head { display: flex; align-items: center; justify-content: space-between; margin: 8rpx 2rpx 14rpx; }
.result-title, .related-head > text:first-child { font-size: 34rpx; font-weight: 700; }
.count { color: #C84B31; font-size: 28rpx; font-weight: 600; }
.filter-btn { margin: 0; padding: 0 20rpx; height: 64rpx; line-height: 62rpx; border: 1rpx solid #CFCAC2; border-radius: 16rpx; background: transparent; font-size: 25rpx; }
.result-state { padding: 70rpx 20rpx; color: #8D8880; text-align: center; font-size: 24rpx; }.result-state.error { color: #B7432C; }
.place-row { min-height: 190rpx; display: flex; align-items: center; gap: 22rpx; padding: 20rpx 2rpx; border-bottom: 1rpx solid #DEDAD2; }
.place-image { width: 150rpx; height: 150rpx; flex: none; border-radius: 18rpx; }
.place-copy { flex: 1; min-width: 0; display: flex; flex-direction: column; }
.place-name { font-size: 31rpx; font-weight: 650; }
.place-meta { margin: 8rpx 0 13rpx; color: #79746D; font-size: 23rpx; white-space: nowrap; }
.experience-count { align-self: flex-start; padding: 9rpx 14rpx; border-radius: 12rpx; background: #F8E9E4; color: #B7432C; font-size: 22rpx; }
.row-arrow { color: #333; font-size: 48rpx; }
.related-head { margin-top: 38rpx; }
.related-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 14rpx; }
.related-card { overflow: hidden; }
.related-card image { width: 100%; height: 160rpx; display: block; }
.related-card text { display: block; min-height: 100rpx; padding: 16rpx; font-size: 23rpx; line-height: 1.5; }
.add-place { display: flex; flex-direction: column; align-items: center; margin-top: 44rpx; color: #7F7A73; }
.add-place button { margin-top: 18rpx; padding: 0 42rpx; border: 1rpx solid #C84B31; border-radius: 16rpx; background: transparent; color: #C84B31; font-size: 27rpx; }
.place-mask { position: fixed; inset: 0; z-index: 30; display: flex; align-items: flex-end; background: rgba(28, 26, 22, .5); }
.place-sheet { width: 100%; box-sizing: border-box; display: flex; flex-direction: column; padding: 18rpx 30rpx 44rpx; border-radius: 36rpx 36rpx 0 0; background: #F6F2EA; }
.place-sheet .panel-handle { width: 72rpx; height: 7rpx; margin: 0 auto; border-radius: 5rpx; background: #C9BFAF; }
.place-sheet-title { margin-top: 28rpx; font-size: 34rpx; font-weight: 750; text-align: center; }
.place-sheet-desc { margin-top: 8rpx; color: #8D8880; font-size: 21rpx; text-align: center; }
.place-field { display: flex; flex-direction: column; margin-top: 26rpx; }
.place-field > text { margin-bottom: 12rpx; color: #55514B; font-size: 23rpx; font-weight: 650; }
.place-field input { height: 88rpx; padding: 0 24rpx; border: 1rpx solid #DEDAD2; border-radius: 18rpx; background: #FFFEFB; font-size: 27rpx; }
.place-location-button { width: 100%; height: 88rpx; margin: 0; border: 1rpx solid #C84B31; border-radius: 18rpx; background: #FFFEFB; color: #C84B31; font-size: 26rpx; line-height: 86rpx; }
.picked-location { display: flex; align-items: center; gap: 18rpx; padding: 20rpx 22rpx; border: 1rpx solid #D7C9B9; border-radius: 18rpx; background: #FFFEFB; }
.picked-location > view { min-width: 0; display: flex; flex: 1; flex-direction: column; }
.picked-location view text:first-child { color: #2D2A26; font-size: 26rpx; font-weight: 700; }
.picked-location view text:last-child { margin-top: 6rpx; overflow: hidden; color: #807A72; font-size: 21rpx; text-overflow: ellipsis; white-space: nowrap; }
.picked-location > text { flex: none; color: #C84B31; font-size: 22rpx; font-weight: 700; }
.place-submit { margin-top: 38rpx; height: 92rpx; line-height: 90rpx; border: 0; border-radius: 18rpx; background: #C84B31; color: #fff; font-size: 28rpx; font-weight: 700; }
.place-submit[disabled] { opacity: .6; color: #fff; }
.place-cancel { margin-top: 24rpx; color: #8D8880; font-size: 24rpx; text-align: center; }
</style>
