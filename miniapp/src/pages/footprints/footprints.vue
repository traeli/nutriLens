<template>
  <view class="footprints-page">
    <view class="header" :style="{ paddingTop: `${navLayout.contentTop}px` }">
      <view class="city-chip" @tap="backToCityPicker">{{ city.name || '选择城市' }}⌄</view>
      <text class="sky-mark">✦</text>
    </view>

    <view class="page-heading">
      <view class="page-title"><text>吃过的地方</text><view><view></view><view></view></view></view>
      <text class="page-caption">一餐一饭，慢慢组成自己的生活</text>
    </view>

    <view class="summary-hero">
      <view><text class="count">{{ summary.place_count }}</text><text> 家店</text></view>
      <view><text class="count">{{ summary.business_area_count }}</text><text> 个街区</text></view>
      <text class="summary-note">不画地图<br/>只收藏生活</text>
    </view>

    <view class="poster-entry" @tap="openPoster">
      <image :src="fallbackBackground" mode="aspectFill" />
      <view class="poster-shade"></view>
      <view class="poster-copy">
        <text class="poster-eyebrow">非地理示意 · 私人回忆</text>
        <text class="poster-title">食光纪念海报</text>
        <text class="poster-subtitle">把认真吃饭的日子，留成一张生活插画</text>
      </view>
      <view class="poster-action"><text>生成我的海报</text><text>↗</text></view>
      <text class="poster-star star-a">✦</text><text class="poster-star star-b">·</text><text class="poster-star star-c">✦</text>
    </view>

    <view class="content">
      <view class="records-head"><text>{{ city.name || '当前城市' }}的最近足迹</text><text>{{ records.length }} 条</text></view>
      <view v-if="loading" class="state-card"><text class="loading-dot"></text><text>正在整理你的城市足迹</text></view>
      <view v-if="!records.length && !loading" class="empty">
        <text>还没有到店记录</text>
        <text>完成第一条记录，开始收藏认真吃饭的日子。</text>
        <button @tap="startRecord">开始记录</button>
      </view>
      <view v-for="(record, index) in records" :key="record.id" class="timeline-row" @tap="goRecord(record)">
        <view class="timeline-date"><text>{{ record.date }}</text><view></view></view>
        <image :src="record.image" mode="aspectFill" />
        <view class="timeline-copy"><text>{{ record.place }}</text><text>{{ record.area }}</text><text>★ 4.{{ 6 - Math.min(index, 2) }} · {{ record.status }}</text></view>
        <text class="arrow">›</text>
      </view>
    </view>
    <canvas canvas-id="footprintPoster" class="poster-canvas"></canvas>
  </view>
</template>

<script>
import { api } from '@/api/request.js'
import { assetUrl } from '@/utils/assets.js'
import { diningImages } from '@/mock/city-dining.js'
import { getSelectedCity } from '@/store/city.js'
import { syncCustomTabBar } from '@/utils/tab-bar.js'

const FALLBACK_BACKGROUND = assetUrl('/static/posters/food-memory-night-v1.jpg')

function getNavLayout() {
  const info = typeof uni.getWindowInfo === 'function' ? uni.getWindowInfo() : uni.getSystemInfoSync()
  const statusBarHeight = Number(info.statusBarHeight || 20)
  let menu = null
  try { menu = uni.getMenuButtonBoundingClientRect() } catch (error) { menu = null }
  const menuTop = menu && Number(menu.top) > 0 ? Number(menu.top) : statusBarHeight + 6
  const menuHeight = menu && Number(menu.height) > 0 ? Number(menu.height) : 32
  const menuBottom = menu && Number(menu.bottom) > menuTop ? Number(menu.bottom) : menuTop + menuHeight
  return { contentTop: Math.max(statusBarHeight, menuBottom) + 14 }
}

export default {
  data() {
    return {
      navLayout: getNavLayout(), city: getSelectedCity(), loading: false, records: [],
      fallbackBackground: FALLBACK_BACKGROUND,
      summary: { place_count: 0, business_area_count: 0 },
    }
  },
  onShow() {
    this.navLayout = getNavLayout()
    this.city = getSelectedCity()
    syncCustomTabBar(this, 2)
    this.loadFootprints()
  },
  methods: {
    async loadFootprints() {
      if (!uni.getStorageSync('token')) { this.records = []; this.summary = { place_count: 0, business_area_count: 0 }; return }
      this.loading = true
      try {
        const [mapResult, recordPage] = await Promise.all([api.getFootprintMap(), api.getMyRecords({ limit: 50 })])
        const cityCode = String(this.city.code || '')
        const points = (mapResult.points || []).filter(point => !cityCode || String(point.city_code || '') === cityCode)
        const items = (recordPage.items || []).filter(item => !cityCode || String((item.place && item.place.city_code) || '') === cityCode)
        this.records = items.map(this.normalizeRecord)
        this.summary = { place_count: points.length, business_area_count: new Set(points.map(point => point.business_area).filter(Boolean)).size }
      } catch (error) {
        this.records = []; this.summary = { place_count: 0, business_area_count: 0 }
        console.warn('[footprints] load failed:', error.message)
      } finally { this.loading = false }
    },
    normalizeRecord(item) {
      const statusMap = { draft: ['草稿', 'private'], pending_review: ['审核中', 'reviewing'], published: ['已公开', 'published'], rejected: ['需修改', 'warning'], withdrawn: ['已撤回', 'private'] }
      const state = statusMap[item.record.publish_status] || ['仅自己可见', 'private']
      return { id: item.record.id, place: item.place.name, date: this.formatDate(item.version.visit_date), area: [item.place.district, item.place.business_area].filter(Boolean).join(' · '), status: state[0], publishStatus: item.record.publish_status, image: (item.media && item.media[0] && item.media[0].public_url) || diningImages.shop }
    },
    formatDate(value) { return value ? String(value).slice(0, 10).replace(/^\d{4}-/, '').replace('-', '月') + '日' : '日期待补充' },
    async openPoster() {
      if (!this.records.length) return uni.showToast({ title: '先记录一餐再生成海报', icon: 'none' })
      let theme = {}
      try { theme = await api.getCityPosterTheme(this.city.code) } catch {}
      const ctx = uni.createCanvasContext('footprintPoster', this)
      const accent = theme.accent_color || '#C7FF35'
      ctx.setFillStyle('#10140F'); ctx.fillRect(0, 0, 650, 920)
      ctx.setFillStyle(accent); ctx.fillRect(42, 44, 118, 10)
      ctx.setFillStyle('#F5F1E7'); ctx.setFontSize(48); ctx.fillText(theme.title || `${this.city.name}食光纪念`, 42, 125)
      ctx.setFillStyle('#AEB6A8'); ctx.setFontSize(24); ctx.fillText(theme.subtitle || '认真吃饭的日子，慢慢连成生活', 42, 172)
      ctx.setFillStyle(accent); ctx.setFontSize(76); ctx.fillText(String(this.summary.place_count || 0), 42, 280)
      ctx.setFillStyle('#F5F1E7'); ctx.setFontSize(25); ctx.fillText('家店 · 私人足迹', 150, 270)
      const visible = this.records.slice(0, 7)
      visible.forEach((item, index) => {
        const y = 350 + index * 70
        ctx.setFillStyle(index % 2 ? '#789080' : accent); ctx.beginPath(); ctx.arc(58, y - 8, 8, 0, Math.PI * 2); ctx.fill()
        if (index < visible.length - 1) { ctx.setStrokeStyle('#506054'); ctx.moveTo(58, y); ctx.lineTo(58, y + 54); ctx.stroke() }
        ctx.setFillStyle('#F5F1E7'); ctx.setFontSize(25); ctx.fillText(String(item.place || '').slice(0, 15), 88, y)
        ctx.setFillStyle('#8E9A8D'); ctx.setFontSize(19); ctx.fillText(`${item.date} · ${item.area || '城市一角'}`.slice(0, 28), 88, y + 28)
      })
      ctx.setFillStyle('#879185'); ctx.setFontSize(18); ctx.fillText('非地理示意 · 仅由你的私人记录生成', 42, 875)
      ctx.draw(false, () => setTimeout(() => uni.canvasToTempFilePath({ canvasId: 'footprintPoster', width: 650, height: 920, destWidth: 1300, destHeight: 1840, success: result => uni.previewImage({ urls: [result.tempFilePath] }), fail: () => uni.showToast({ title: '海报生成失败', icon: 'none' }) }, this), 120))
    },
    backToCityPicker() { uni.switchTab({ url: '/pages/home/home' }) },
    goRecord(record) { uni.navigateTo({ url: record.publishStatus === 'published' ? `/pages/experience-detail/experience-detail?id=${record.id}` : `/pages/review-status/review-status?id=${record.id}` }) },
    startRecord() { uni.switchTab({ url: '/pages/record/record' }) },
  },
}
</script>

<style scoped>
.footprints-page { min-height: 100vh; padding: 0 24rpx calc(145rpx + env(safe-area-inset-bottom)); background: #f5faec; color: #171914; }
.poster-canvas { position: fixed; left: -2000px; top: 0; width: 650px; height: 920px; }
.header { display: flex; align-items: center; justify-content: space-between; padding: calc(env(safe-area-inset-top) + 22rpx) 4rpx 15rpx; }.city-chip { padding: 13rpx 24rpx; border-radius: 30rpx; background: #e8ece2; font-size: 23rpx; font-weight: 750; }.sky-mark { padding-right: 14rpx; font-size: 44rpx; }
.page-heading { margin: 4rpx 4rpx 16rpx; }.page-title { position: relative; display: inline-flex; }.page-title > text { font-family: "Kaiti SC", STKaiti, cursive; font-size: 72rpx; font-weight: 900; line-height: 1.08; }.page-title > view { position: absolute; right: -70rpx; top: 38rpx; width: 70rpx; height: 60rpx; }.page-title > view > view { position: absolute; width: 10rpx; height: 40rpx; border-radius: 8rpx; background: #9cdd19; transform: rotate(28deg); }.page-title > view > view:last-child { left: 30rpx; top: 15rpx; transform: rotate(65deg); }.page-caption { display: block; margin-top: 9rpx; color: #72796c; font-size: 21rpx; }
.summary-hero { position: relative; width: 54%; display: flex; flex-direction: column; margin: 0 4rpx 20rpx; padding: 17rpx 28rpx; border-radius: 54rpx 70rpx 54rpx 35rpx; background: #c7ff35; }.summary-hero > view { display: flex; align-items: baseline; font-size: 25rpx; font-weight: 800; line-height: 1.2; }.summary-hero .count { margin-right: 5rpx; font-size: 47rpx; font-weight: 900; }.summary-note { position: absolute; left: calc(100% + 38rpx); top: 31rpx; width: 210rpx; font-family: "Kaiti SC", STKaiti, cursive; font-size: 24rpx; font-weight: 700; line-height: 1.45; transform: rotate(-7deg); }
.poster-entry { position: relative; height: 500rpx; overflow: hidden; border-radius: 38rpx; background: #070a08; box-shadow: 0 22rpx 46rpx rgba(14,17,11,.22); }.poster-entry > image { width: 100%; height: 100%; }.poster-shade { position: absolute; inset: 0; background: linear-gradient(180deg, rgba(2,4,3,.12), rgba(2,4,3,.34) 46%, rgba(2,4,3,.88)); }.poster-copy { position: absolute; left: 30rpx; right: 30rpx; top: 32rpx; display: flex; flex-direction: column; color: #f5f2e8; }.poster-eyebrow { color: #c7ff35; font-size: 17rpx; font-weight: 800; letter-spacing: 2rpx; }.poster-title { margin-top: 14rpx; font-family: "Kaiti SC", STKaiti, cursive; font-size: 54rpx; font-weight: 900; }.poster-subtitle { width: 72%; margin-top: 9rpx; color: #d2d1c5; font-size: 21rpx; line-height: 1.5; }.poster-action { position: absolute; left: 28rpx; right: 28rpx; bottom: 25rpx; display: flex; align-items: center; justify-content: space-between; padding: 20rpx 23rpx; border-radius: 27rpx; background: #c7ff35; color: #11140f; font-size: 23rpx; font-weight: 850; }.poster-star { position: absolute; color: #efffd0; text-shadow: 0 0 18rpx #c7ff35; }.star-a { right: 70rpx; top: 72rpx; }.star-b { right: 160rpx; top: 210rpx; font-size: 44rpx; }.star-c { right: 55rpx; top: 290rpx; font-size: 20rpx; }
.records-head { display: flex; align-items: center; justify-content: space-between; margin: 34rpx 3rpx 10rpx; }.records-head text:first-child { font-family: "Kaiti SC", STKaiti, cursive; font-size: 34rpx; font-weight: 900; }.records-head text:last-child { color: #747b6e; font-size: 20rpx; }.state-card,.empty { display: flex; flex-direction: column; align-items: center; margin-top: 16rpx; padding: 42rpx 24rpx; border-radius: 30rpx; background: #fff; color: #747a70; }.loading-dot { width: 20rpx; height: 20rpx; margin-bottom: 14rpx; border-radius: 50%; background: #9cdd19; animation: pulse 1s infinite alternate; }.empty text:first-child { color: #292d25; font-size: 28rpx; font-weight: 800; }.empty text:nth-child(2) { margin-top: 11rpx; font-size: 20rpx; text-align: center; }.empty button { margin-top: 24rpx; padding: 0 34rpx; border-radius: 40rpx; background: #10130e; color: #fff; font-size: 22rpx; }
.timeline-row { min-height: 128rpx; display: grid; grid-template-columns: 115rpx 150rpx 1fr 28rpx; align-items: center; gap: 15rpx; padding: 14rpx 0; border-bottom: 1rpx solid #dfe3da; }.timeline-date { position: relative; align-self: stretch; display: flex; align-items: center; border-right: 2rpx solid #171813; }.timeline-date text { width: 100rpx; font-size: 19rpx; font-weight: 700; }.timeline-date view { position: absolute; right: -9rpx; top: 50%; width: 15rpx; height: 15rpx; margin-top: -8rpx; border: 3rpx solid #171813; border-radius: 50%; background: #c7ff35; }.timeline-row > image { width: 150rpx; height: 104rpx; border-radius: 21rpx; }.timeline-copy { min-width: 0; display: flex; flex-direction: column; }.timeline-copy text:first-child { font-size: 25rpx; font-weight: 850; }.timeline-copy text:nth-child(2) { margin-top: 5rpx; overflow: hidden; color: #747a72; font-size: 18rpx; text-overflow: ellipsis; white-space: nowrap; }.timeline-copy text:nth-child(3) { margin-top: 7rpx; font-size: 18rpx; font-weight: 650; }.arrow { color: #8d9486; font-size: 40rpx; }
@keyframes pulse { to { opacity: .25; transform: scale(.7); } }
</style>
