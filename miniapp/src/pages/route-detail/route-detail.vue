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
        <text>出发前提醒</text>
        <text>路线按建议顺序排列；雨天替代、闭店风险和实时路况请在出发前再次确认。</text>
      </view>

      <view class="overview-card">
        <text class="section-title">路线概览</text>
        <view class="overview-grid">
          <view><text>站点</text><text class="overview-value">{{ route.stops.length }} 站</text></view>
          <view><text>用时 / 距离</text><text class="overview-value">{{ route.meta || '以实际路线为准' }}</text></view>
          <view><text>适合</text><text class="overview-value">{{ route.tag || '城市漫游' }}</text></view>
        </view>
      </view>

      <view v-if="route.stops.length" class="route-sketch-card">
        <view class="sketch-heading">
          <view>
            <text class="section-title">路线手记</text>
            <view class="sketch-title-mark"></view>
          </view>
          <view class="sketch-palette"><view></view><view></view><view></view></view>
        </view>
        <text class="sketch-caption">跟着坐标走，线条故意画得不那么规矩。</text>
        <view class="route-sketch">
          <canvas id="route-line-canvas" canvas-id="route-line-canvas" class="route-line-canvas"></canvas>
          <view v-for="point in sketchLayout.points" :key="`point-${point.index}`" class="sketch-stop" :style="pointStyle(point)" @tap="openPlace(point.stop)">
            <view class="sketch-icon-wrap" :style="{ borderBottomColor: themeAccent(point.stop) }">
              <image class="sketch-icon" :src="themeIcon(point.stop)" mode="aspectFit" />
            </view>
            <text class="sketch-index">{{ String(point.index + 1).padStart(2, '0') }}</text>
            <text class="sketch-name">{{ point.stop.name }}</text>
          </view>
          <text class="sketch-side-note">慢慢走，沿路吃</text>
        </view>
        <view class="sketch-memo-row">
          <view class="sketch-memo yellow"><text>路线</text><text>{{ route.meta || `${route.stops.length} 站` }}</text></view>
          <view class="sketch-memo pink"><text>主题</text><text>{{ route.tag || '城市漫游' }}</text></view>
        </view>
      </view>

      <view class="stop-list">
        <text class="section-title">按顺序吃喝</text>
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
        <button class="copy-button" @tap="copyRoute">复制完整路线</button>
        <text>复制后可粘贴到高德、百度等地图中逐站添加。</text>
      </view>
      <view v-if="route.stops.length" class="fixed-journey">
        <button class="journey-button" @tap="advanceJourney">{{ journeyActionLabel }}</button>
        <text v-if="journeyId && nextStop" class="skip-action" @tap="skipNextStop">跳过这一站</text>
      </view>
    </template>
  </view>
</template>

<script>
import { api } from '@/api/request.js'
import { foodIconSource, foodThemeAccent } from '@/utils/food-icons.js'
import { routeSketchLayout } from '@/utils/route-layout.js'

export default {
  data() {
    return { id: '', nearby: false, nearbyQuery: {}, route: { stops: [] }, loading: true, error: '', favorited: false, journeyId: 0, completedStops: [] }
  },
  onLoad(options) {
    this.id = String(options.id || '')
    this.nearby = options.nearby === '1'
    this.nearbyQuery = {
      city_code: options.city_code || '',
      longitude: options.longitude || '',
      latitude: options.latitude || '',
      limit: 4,
      budget: options.budget || '',
      scene: options.scene || '',
    }
    this.loadRoute()
  },
  computed: {
    sketchLayout() { return routeSketchLayout(this.route.stops) },
    nextStopIndex() { return this.route.stops.findIndex((stop, index) => !this.completedStops.includes(index)) },
    nextStop() { return this.nextStopIndex >= 0 ? this.route.stops[this.nextStopIndex] : null },
    journeyActionLabel() {
      if (!this.journeyId) return '开始这条路线'
      if (!this.nextStop) return '完成路线'
      return `下一站 · ${this.nextStop.name}`
    },
  },
  methods: {
    pointStyle(point) { return { left: `${point.x}%`, top: `${point.y}%` } },
    themeIcon(stop) { return stop.theme_icon_url || foodIconSource(stop.icon_key || stop.theme_code || 'general') },
    themeAccent(stop) { return foodThemeAccent(stop.theme_code || 'general') },
    drawRouteSketch() {
      if (!this.sketchLayout.curves.length) return
      this.$nextTick(() => {
        uni.createSelectorQuery().in(this).select('#route-line-canvas').boundingClientRect(canvasInfo => {
          if (!canvasInfo || !canvasInfo.width || !canvasInfo.height) return
          const context = uni.createCanvasContext('route-line-canvas', this)
          context.clearRect(0, 0, canvasInfo.width, canvasInfo.height)
          context.setStrokeStyle('#4A453D')
          context.setLineWidth(2.35)
          context.setLineCap('round')
          context.setLineJoin('round')
          const percentX = value => value * canvasInfo.width / 100
          const percentY = value => value * canvasInfo.height / 100
          const first = this.sketchLayout.points[0]
          context.beginPath()
          context.moveTo(percentX(first.x), percentY(first.y))
          this.sketchLayout.curves.forEach(curve => {
            context.quadraticCurveTo(
              percentX(curve.control.x), percentY(curve.control.y),
              percentX(curve.end.x), percentY(curve.end.y),
            )
          })
          context.stroke()
          context.draw(false)
        }).exec()
      })
    },
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
        if (!this.error && this.route.stops.length) this.drawRouteSketch()
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
    async advanceJourney() {
      if (!this.journeyId) return this.startOrComplete()
      if (this.nextStopIndex >= 0) this.completedStops = [...this.completedStops, this.nextStopIndex]
      return this.startOrComplete()
    },
    skipNextStop() {
      if (this.nextStopIndex < 0) return
      this.completedStops = [...this.completedStops, this.nextStopIndex]
      uni.showToast({ title: '已跳过，可继续下一站', icon: 'none' })
    },
  },
}
</script>

<style scoped>
.route-page { min-height: 100vh; box-sizing: border-box; padding: 30rpx 28rpx calc(180rpx + env(safe-area-inset-bottom)); background: #FFF9EE; color: #332E26; }
.route-state { min-height: 70vh; display: flex; align-items: center; justify-content: center; color: #78806F; font-size: 26rpx; text-align: center; }
.route-state.error { color: #A94430; }
.route-hero { display: flex; flex-direction: column; padding: 38rpx 32rpx; border-radius: 34rpx; background: #11120F; color: #fff; }
.route-tag { align-self: flex-start; padding: 8rpx 16rpx; border-radius: 22rpx; background: #C7FF35; color: #11120F; font-size: 20rpx; font-weight: 800; }
.route-title { margin-top: 24rpx; font-family: "Kaiti SC", STKaiti, serif; font-size: 48rpx; font-weight: 900; line-height: 1.25; }
.route-meta { margin-top: 15rpx; color: rgba(255,255,255,.58); font-size: 23rpx; line-height: 1.6; }
.route-note { display: flex; flex-direction: column; margin: 20rpx 0 30rpx; padding: 23rpx 25rpx; border: 1rpx solid #D8E1C9; border-radius: 22rpx; background: #fff; }
.route-note text:first-child { font-size: 25rpx; font-weight: 800; }
.route-note text:last-child { margin-top: 6rpx; color: #76806B; font-size: 21rpx; line-height: 1.5; }
.overview-card { margin-bottom: 30rpx; padding: 28rpx; border: 1rpx solid #E7DCC5; border-radius: 28rpx; background: #FFFFFF; }
.section-title { display: block; margin-bottom: 22rpx; color: #332E26; font-size: 32rpx; font-weight: 800; }
.overview-grid { display: grid; grid-template-columns: repeat(3, 1fr); gap: 12rpx; }
.overview-grid > view { display: flex; flex-direction: column; min-width: 0; padding: 18rpx; border-radius: 20rpx; background: #FFF4DF; color: #6F675B; font-size: 22rpx; }
.overview-value { margin-top: 8rpx; color: #332E26; font-size: 24rpx; font-weight: 800; line-height: 1.35; }
.route-sketch-card { position: relative; margin-bottom: 36rpx; padding: 32rpx 20rpx 28rpx; border-radius: 8rpx 24rpx 12rpx 20rpx; background: #FFFDFC; box-shadow: 0 10rpx 28rpx rgba(68,55,36,.07); overflow: hidden; }
.route-sketch-card::before { content: ''; position: absolute; left: 0; right: 0; top: 0; height: 2rpx; background: rgba(72,63,51,.1); }
.sketch-heading { position: relative; z-index: 3; display: flex; align-items: center; justify-content: space-between; padding: 0 12rpx; }
.sketch-heading > view:first-child { position: relative; }
.sketch-heading .section-title { position: relative; z-index: 2; margin-bottom: 0; font-family: "Kaiti SC", STKaiti, serif; font-size: 38rpx; transform: rotate(-2deg); }
.sketch-title-mark { position: absolute; z-index: 1; left: -6rpx; right: -18rpx; bottom: 0; height: 12rpx; border-radius: 40% 60% 46% 54%; background: rgba(248,207,62,.65); transform: rotate(-1deg); }
.sketch-palette { display: flex; align-items: center; gap: 8rpx; transform: rotate(3deg); }
.sketch-palette view { width: 26rpx; height: 26rpx; border-radius: 55% 45% 51% 49%; background: #F3C846; }
.sketch-palette view:nth-child(2) { background: #EE7654; transform: translateY(5rpx); }
.sketch-palette view:nth-child(3) { background: #5E9FBD; transform: translateY(-3rpx); }
.sketch-caption { position: relative; z-index: 3; display: block; margin: 12rpx 12rpx 0; color: #777066; font-family: "Kaiti SC", STKaiti, serif; font-size: 21rpx; transform: rotate(-.5deg); }
.route-sketch { position: relative; height: 640rpx; margin-top: 6rpx; }
.route-line-canvas { position: absolute; z-index: 1; inset: 0; width: 100%; height: 100%; pointer-events: none; }
.sketch-stop { position: absolute; z-index: 2; display: flex; flex-direction: column; align-items: center; width: 190rpx; transform: translate(-50%, -50%); }
.sketch-stop:nth-of-type(2n) { transform: translate(-50%, -50%) rotate(1.3deg); }
.sketch-icon-wrap { width: 96rpx; height: 80rpx; border-bottom: 9rpx solid; border-radius: 48% 52% 56% 44%; box-sizing: border-box; transform: rotate(-3deg); }
.sketch-icon { width: 96rpx; height: 96rpx; transform: translateY(-12rpx); }
.sketch-index { margin-top: 2rpx; color: #332E26; font-family: "Kaiti SC", STKaiti, serif; font-size: 23rpx; font-weight: 900; transform: rotate(-3deg); }
.sketch-name { max-width: 190rpx; margin-top: 2rpx; padding: 2rpx 7rpx 5rpx; border-bottom: 5rpx solid rgba(248,207,62,.48); background: rgba(255,253,248,.82); color: #332E26; font-family: "Kaiti SC", STKaiti, serif; font-size: 22rpx; font-weight: 800; line-height: 1.2; text-align: center; }
.sketch-side-note { position: absolute; z-index: 3; right: 2%; top: 48%; color: #E96E4E; font-family: "Kaiti SC", STKaiti, serif; font-size: 21rpx; transform: rotate(-7deg); }
.sketch-side-note::after { content: ''; display: block; width: 132rpx; height: 4rpx; margin-top: 2rpx; border-radius: 50%; background: #E96E4E; transform: rotate(-3deg); opacity: .68; }
.sketch-memo-row { display: flex; justify-content: flex-end; gap: 18rpx; padding: 0 12rpx 4rpx; }
.sketch-memo { display: flex; flex-direction: column; width: 190rpx; min-height: 88rpx; padding: 14rpx 16rpx; box-sizing: border-box; box-shadow: 4rpx 7rpx 0 rgba(68,55,36,.07); font-family: "Kaiti SC", STKaiti, serif; transform: rotate(-2deg); }
.sketch-memo.pink { background: #F3B9D6; transform: rotate(2deg) translateY(-8rpx); }
.sketch-memo.yellow { background: #F6D65A; }
.sketch-memo text:first-child { font-size: 18rpx; font-weight: 800; }
.sketch-memo text:last-child { margin-top: 5rpx; font-size: 20rpx; line-height: 1.25; }
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
.fixed-journey { position: fixed; z-index: 10; left: 0; right: 0; bottom: 0; display: flex; align-items: center; gap: 20rpx; padding: 18rpx 28rpx calc(18rpx + env(safe-area-inset-bottom)); border-top: 1rpx solid #E7DCC5; background: rgba(255,249,238,.96); }
.journey-button { flex: 1; height: 92rpx; margin: 0; border: 0; border-radius: 46rpx; background: #2F7D58; color: #fff; font-size: 27rpx; font-weight: 900; line-height: 92rpx; }
.journey-button::after { border: 0; }
.skip-action { flex: none; min-height: 88rpx; padding: 0 8rpx; color: #6F675B; font-size: 24rpx; font-weight: 700; line-height: 88rpx; }
.route-actions > text { margin-top: 15rpx; color: #7B8374; font-size: 20rpx; text-align: center; }
</style>
