<template>
  <view class="route-page">
    <view v-if="loading" class="route-state">正在整理这条吃喝路线…</view>
    <view v-else-if="error" class="route-state error" @tap="loadRoute">{{ error }}，点击重试</view>
    <template v-else>
      <view class="route-hero">
        <text class="route-tag">{{ route.tag || '吃喝攻略' }}</text>
        <text class="route-title">{{ route.title }}</text>
        <text class="route-meta">{{ route.meta || '按顺序慢慢吃，实际出行交给地图 App' }}</text>
      </view>

      <view class="route-note">
        <text>路线只标记建议顺序</text>
        <text>交通方式、营业状态和实时路况请以地图 App 为准。</text>
      </view>

      <view class="stop-list">
        <view v-if="!route.stops.length" class="empty-stops">这条攻略的具体站点还在整理中。</view>
        <view v-for="(stop, index) in route.stops" :key="`${stop.place_id || stop.name}-${index}`" class="stop-row">
          <view class="route-line" @tap="toggleStop(index)"><text :class="{ done: completedStops.includes(index) }">{{ completedStops.includes(index) ? '✓' : index + 1 }}</text><view v-if="index < route.stops.length - 1"></view></view>
          <view class="stop-copy" @tap="openPlace(stop)">
            <text class="stop-name">{{ stop.name }}</text>
            <text class="stop-address">{{ stop.address || '地址待补充' }}</text>
            <text v-if="stop.distance_meters" class="stop-distance">距你约 {{ formatDistance(stop.distance_meters) }} · {{ stop.experience_count || 0 }} 条真实记录</text>
            <text v-if="stop.note" class="stop-note">{{ stop.note }}</text>
          </view>
          <button class="map-button" :disabled="!hasCoordinate(stop)" @tap.stop="openStop(stop)">打开地图</button>
        </view>
      </view>

      <view v-if="route.stops.length" class="route-actions">
        <button class="copy-button" @tap="toggleFavorite">{{ favorited ? '取消收藏' : '收藏路线' }}</button>
        <button class="copy-button journey-button" @tap="startOrComplete">{{ journeyId ? (completedStops.length === route.stops.length ? '完成路线' : '保存进度') : '开始路线' }}</button>
        <button class="copy-button" @tap="copyRoute">复制完整路线</button>
        <text>复制后可粘贴到高德、百度等地图中逐站添加。</text>
      </view>
    </template>
  </view>
</template>

<script>
import { api } from '@/api/request.js'

export default {
  data() {
    return { id: '', nearby: false, nearbyQuery: {}, route: { stops: [] }, loading: true, error: '', favorited: false, journeyId: 0, completedStops: [] }
  },
  onLoad(options) {
    this.id = String(options.id || '')
    this.nearby = options.nearby === '1'
    this.nearbyQuery = { city_code: options.city_code || '', longitude: options.longitude || '', latitude: options.latitude || '', limit: 4 }
    this.loadRoute()
  },
  methods: {
    async loadRoute() {
      if (!this.nearby && !/^\d+$/.test(this.id)) {
        this.loading = false
        this.error = '攻略编号无效'
        return
      }
      this.loading = true
      this.error = ''
      try {
        const route = this.nearby ? await api.getNearbyRoute(this.nearbyQuery) : await api.getRoute(this.id)
        this.route = { ...route, stops: Array.isArray(route.stops) ? route.stops : [] }
      } catch (error) {
        this.error = error.message || '攻略加载失败'
      } finally {
        this.loading = false
      }
    },
    hasCoordinate(stop) {
      const longitude = Number(stop.longitude)
      const latitude = Number(stop.latitude)
      return Number.isFinite(longitude) && Number.isFinite(latitude) && Boolean(longitude || latitude)
    },
    openStop(stop) {
      if (!this.hasCoordinate(stop)) return
      uni.openLocation({
        latitude: Number(stop.latitude),
        longitude: Number(stop.longitude),
        name: stop.name,
        address: stop.address || '',
        scale: 18,
      })
    },
    openPlace(stop) {
      if (stop.place_id) uni.navigateTo({ url: `/pages/place-detail/place-detail?id=${stop.place_id}` })
    },
    formatDistance(value) {
      const meters = Number(value || 0)
      return meters >= 1000 ? `${(meters / 1000).toFixed(1)}km` : `${Math.max(1, Math.round(meters))}m`
    },
    copyRoute() {
      const text = this.route.stops.map((stop, index) => `${index + 1}. ${stop.name}${stop.address ? `（${stop.address}）` : ''}`).join('\n')
      uni.setClipboardData({
        data: `${this.route.title}\n${text}`,
        success: () => uni.showToast({ title: '路线已复制', icon: 'success' }),
      })
    },
    async toggleFavorite() {
      if (this.nearby) return uni.showToast({ title: '临时路线暂不支持收藏', icon: 'none' })
      try { if (this.favorited) await api.unfavoriteRoute(this.id); else await api.favoriteRoute(this.id); this.favorited = !this.favorited } catch (error) { uni.showToast({ title: error.message || '操作失败', icon: 'none' }) }
    },
    toggleStop(index) {
      if (!this.journeyId) return
      this.completedStops = this.completedStops.includes(index) ? this.completedStops.filter(item => item !== index) : [...this.completedStops, index]
    },
    async startOrComplete() {
      if (this.nearby) return uni.showToast({ title: '临时路线暂不支持进度', icon: 'none' })
      try {
        if (!this.journeyId) { const journey = await api.startRoute(this.id); this.journeyId = journey.id; return }
        const complete = this.completedStops.length === this.route.stops.length
        await api.updateRouteJourney(this.journeyId, { completed_stop_ids: this.completedStops, complete })
        uni.showToast({ title: complete ? '路线已完成' : '进度已保存', icon: 'success' })
      } catch (error) { uni.showToast({ title: error.message || '操作失败', icon: 'none' }) }
    },
  },
}
</script>

<style scoped>
.route-page { min-height: 100vh; box-sizing: border-box; padding: 30rpx 28rpx calc(60rpx + env(safe-area-inset-bottom)); background: #F5FAEC; color: #12130F; }
.route-state { min-height: 70vh; display: flex; align-items: center; justify-content: center; color: #78806F; font-size: 26rpx; text-align: center; }
.route-state.error { color: #A94430; }
.route-hero { display: flex; flex-direction: column; padding: 38rpx 32rpx; border-radius: 34rpx; background: #11120F; color: #fff; }
.route-tag { align-self: flex-start; padding: 8rpx 16rpx; border-radius: 22rpx; background: #C7FF35; color: #11120F; font-size: 20rpx; font-weight: 800; }
.route-title { margin-top: 24rpx; font-family: "Kaiti SC", STKaiti, serif; font-size: 48rpx; font-weight: 900; line-height: 1.25; }
.route-meta { margin-top: 15rpx; color: rgba(255,255,255,.58); font-size: 23rpx; line-height: 1.6; }
.route-note { display: flex; flex-direction: column; margin: 20rpx 0 30rpx; padding: 23rpx 25rpx; border: 1rpx solid #D8E1C9; border-radius: 22rpx; background: #fff; }
.route-note text:first-child { font-size: 25rpx; font-weight: 800; }
.route-note text:last-child { margin-top: 6rpx; color: #76806B; font-size: 21rpx; line-height: 1.5; }
.stop-list { padding: 7rpx 8rpx; }
.empty-stops { padding: 90rpx 20rpx; color: #7D8476; font-size: 25rpx; text-align: center; }
.stop-row { display: flex; align-items: flex-start; min-height: 150rpx; }
.route-line { width: 58rpx; align-self: stretch; display: flex; flex-direction: column; align-items: center; flex: none; }
.route-line > text { width: 48rpx; height: 48rpx; border-radius: 50%; background: #11120F; color: #C7FF35; font-size: 22rpx; font-weight: 900; line-height: 48rpx; text-align: center; }
.route-line > text.done { background: #C7FF35; color: #11120F; }
.route-line > view { width: 2rpx; flex: 1; background: #A8B39C; }
.stop-copy { min-width: 0; display: flex; flex: 1; flex-direction: column; padding: 3rpx 20rpx 42rpx 12rpx; }
.stop-name { font-size: 31rpx; font-weight: 850; }
.stop-address { margin-top: 8rpx; color: #747C6D; font-size: 21rpx; line-height: 1.5; }
.stop-distance { margin-top: 8rpx; color: #9A6235; font-size: 20rpx; }
.stop-note { margin-top: 10rpx; color: #4E5945; font-size: 21rpx; }
.map-button { width: 142rpx; height: 64rpx; margin: 0; padding: 0; border: 1rpx solid #161712; border-radius: 32rpx; background: transparent; color: #161712; font-size: 21rpx; font-weight: 800; line-height: 62rpx; }
.map-button[disabled] { border-color: #CDD2C7; color: #ABB1A5; }
.route-actions { display: flex; flex-direction: column; align-items: center; margin-top: 20rpx; }
.copy-button { width: 100%; height: 94rpx; border: 0; border-radius: 47rpx; background: #C7FF35; color: #11120F; font-size: 27rpx; font-weight: 900; line-height: 94rpx; }
.route-actions .copy-button + .copy-button { margin-top: 14rpx; }
.journey-button { background: #11120F; color: #fff; }
.route-actions > text { margin-top: 15rpx; color: #7B8374; font-size: 20rpx; text-align: center; }
</style>
