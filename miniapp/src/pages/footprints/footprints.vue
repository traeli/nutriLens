<template>
  <view class="footprints-page">
    <view class="page-head" :style="{ paddingTop: `${navLayout.contentTop}px` }">
      <view class="head-row">
        <view class="page-title">
          <view class="title-ic" :style="dz('pin', '#3E9B6F', 38)"></view>
          <text class="title-h">吃过的地方</text>
          <text class="title-p">一餐一饭，慢慢组成自己的生活</text>
        </view>
        <view class="city-pill" @tap="backToCityPicker">
          <view class="city-dot"></view>
          <text class="city-name">{{ city.name || '选择城市' }}</text>
          <view class="city-caret" :style="dz('caret', '#8B8271', 18)"></view>
        </view>
      </view>
    </view>

    <view class="map-card">
      <view class="map-card-head">
        <view>
          <text class="map-title">{{ city.name || '当前城市' }}点亮图</text>
          <text class="map-note">每完成一次到店记录，就点亮一个去过的街区</text>
        </view>
        <text class="map-progress">{{ summary.business_area_count }} 个街区</text>
      </view>
      <view class="city-map" :class="{ empty: !mapAreas.length }">
        <view v-for="(area, index) in displayAreas" :key="area.name" class="map-region" :class="{ lit: area.lit }" :style="{ '--region-index': index }">
          <text>{{ area.name }}</text>
        </view>
      </view>
      <text v-if="!mapAreas.length" class="map-empty-copy">地图还没有点亮。留下第一条记录，这里会长出你的城市食记。</text>
    </view>

    <view class="summary-card">
      <view class="summary-medal" :style="dzBg('blob-green')"><view class="summary-medal-ic" :style="dz('compass', '#3E9B6F', 40)"></view></view>
      <view class="summary-body">
        <view class="summary-col">
          <text class="summary-num">{{ summary.place_count }}</text>
          <text class="summary-label">家店</text>
        </view>
        <view class="summary-divider"></view>
        <view class="summary-col">
          <text class="summary-num">{{ summary.business_area_count }}</text>
          <text class="summary-label">个街区</text>
        </view>
      </view>
      <text class="summary-sticker">足迹</text>
    </view>

    <view class="sec">
      <view class="sec-ic" :style="dz('compass', '#3E9B6F', 38)"></view>
      <view class="sec-title">
        <text class="sec-h">{{ city.name || '当前城市' }}的最近足迹</text>
        <view class="sec-scribble" :style="dz('scribble', '#3E9B6F', 56)"></view>
      </view>
      <text class="sec-p">{{ records.length }} 条</text>
    </view>
    <view v-if="loading" class="state-card">
      <view class="loading-dot"></view>
      <text>正在整理你的城市足迹</text>
    </view>
    <view v-if="!records.length && !loading" class="empty-card">
      <view class="empty-trail">
        <view class="trail-stop" :style="dzBg('blob-cream')"><view class="trail-ic" :style="dz('bowl', '#332E26', 38)"></view></view>
        <view class="trail-squig" :style="dz('squiggle', '#D8CCB0', 34)"></view>
        <view class="trail-stop flip" :style="dzBg('blob-cream')"><view class="trail-ic" :style="dz('noodle', '#332E26', 38)"></view></view>
        <view class="trail-squig" :style="dz('squiggle', '#D8CCB0', 34)"></view>
        <view class="trail-stop" :style="dzBg('blob-cream')"><view class="trail-ic" :style="dz('coffee', '#332E26', 38)"></view></view>
      </view>
      <text class="empty-title">还没有到店记录</text>
      <text class="empty-note">完成第一条记录，开始收藏认真吃饭的日子。</text>
      <view class="empty-cta" @tap="startRecord">开始记录</view>
    </view>
    <view v-if="records.length" class="timeline-card">
      <view v-for="record in records" :key="record.id" class="timeline-row" @tap="goRecord(record)">
        <view class="timeline-date">
          <view class="tl-star" :style="dz('star-hand', '#3E9B6F', 30)"></view>
          <text>{{ record.date }}</text>
        </view>
        <view class="tl-photo">
          <image class="timeline-img" :src="record.image" mode="aspectFill" />
        </view>
        <view class="timeline-copy">
          <text class="tl-place">{{ record.place }}</text>
          <text class="tl-area">{{ record.area }}</text>
          <text class="tl-status">{{ record.status }}</text>
        </view>
        <text class="row-arrow">›</text>
      </view>
    </view>

    <view class="poster-compact" @tap="openPoster">
      <view class="poster-compact-copy">
        <text class="poster-compact-title">生成城市食记海报</text>
        <text class="poster-compact-note">把已经点亮的足迹保存成一张图</text>
      </view>
      <text class="poster-compact-arrow">›</text>
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
import { getPrimaryNavLayout } from '@/utils/nav-layout.js'
import { doodleBg, doodleStyle } from '@/utils/doodles.js'

const FALLBACK_BACKGROUND = assetUrl('/static/posters/food-memory-night-v1.jpg')

export default {
  data() {
    return {
      navLayout: getPrimaryNavLayout(), city: getSelectedCity(), loading: false, records: [], mapAreas: [],
      fallbackBackground: FALLBACK_BACKGROUND,
      summary: { place_count: 0, business_area_count: 0 },
    }
  },
  computed: {
    displayAreas() {
      if (this.mapAreas.length) return this.mapAreas.slice(0, 8).map(name => ({ name, lit: true }))
      return ['待点亮', '待探索', '下一站', '新街区'].map(name => ({ name, lit: false }))
    },
  },
  onShow() {
    this.navLayout = getPrimaryNavLayout()
    this.city = getSelectedCity()
    syncCustomTabBar(this, 2)
    this.loadFootprints()
  },
  methods: {
    dz(key, color, size) { return doodleStyle(key, color, size) },
    dzBg(key) { return doodleBg(key) },
    async loadFootprints() {
      if (!uni.getStorageSync('token')) { this.records = []; this.mapAreas = []; this.summary = { place_count: 0, business_area_count: 0 }; return }
      this.loading = true
      try {
        const [mapResult, recordPage] = await Promise.all([api.getFootprintMap(), api.getMyRecords({ limit: 50 })])
        const cityCode = String(this.city.code || '')
        const points = (mapResult.points || []).filter(point => !cityCode || String(point.city_code || '') === cityCode)
        const items = (recordPage.items || []).filter(item => !cityCode || String((item.place && item.place.city_code) || '') === cityCode)
        this.records = items.map(this.normalizeRecord)
        this.mapAreas = [...new Set(points.map(point => point.business_area || point.district).filter(Boolean))]
        this.summary = { place_count: points.length, business_area_count: this.mapAreas.length }
      } catch (error) {
        this.records = []; this.mapAreas = []; this.summary = { place_count: 0, business_area_count: 0 }
        console.warn('[footprints] load failed:', error.message)
      } finally { this.loading = false }
    },
    normalizeRecord(item) {
      return { id: item.record.id, place: item.place.name, date: this.formatDate(item.version.visit_date), area: [item.place.district, item.place.business_area].filter(Boolean).join(' · '), status: '私人足迹', image: (item.media && item.media[0] && item.media[0].public_url) || diningImages.shop }
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
    goRecord() { uni.showToast({ title: '这是仅自己可见的到店足迹', icon: 'none' }) },
    startRecord() { uni.switchTab({ url: '/pages/record/record' }) },
  },
}
</script>

<style scoped>
/* —— V6 贴纸手帐：token 同 home.vue —— */
.footprints-page {
  min-height: 100vh;
  padding: 0 32rpx calc(150rpx + env(safe-area-inset-bottom));
  overflow: hidden;
  background: linear-gradient(180deg, #FFFDF9 0%, #FFFBF3 36%, #FFF3DF 100%);
  color: #332E26;
  font-family: "PingFang SC", "Helvetica Neue", Helvetica, sans-serif;
}

/* 海报 canvas 保持离屏，openPoster 依赖 */
.poster-canvas { position: fixed; left: -2000px; top: 0; width: 650px; height: 920px; }

/* 头部 */
.page-head { margin-bottom: 8rpx; }
.head-row { display: flex; align-items: center; gap: 16rpx; }
.page-title { min-width: 0; display: flex; align-items: center; gap: 14rpx; }
.title-ic { flex: none; }
.title-h { font-size: 34rpx; font-weight: 900; letter-spacing: 1rpx; white-space: nowrap; }
.title-p { font-size: 23rpx; color: #8B8271; padding-top: 6rpx; white-space: nowrap; }
.city-pill {
  margin-left: auto; min-width: 0; flex: none; display: flex; align-items: center; gap: 12rpx;
  height: 60rpx; padding: 0 22rpx 0 24rpx; box-sizing: border-box;
  background: #FFF4DF; border: 3rpx solid #EFE5D0; border-radius: 999rpx;
}
.city-dot { flex: none; width: 12rpx; height: 12rpx; border-radius: 50%; background: #3E9B6F; }
.city-name { font-size: 25rpx; font-weight: 700; white-space: nowrap; }
.city-caret { flex: none; margin-top: 2rpx; }

/* 足迹主资产：即使为空也保留可被逐步点亮的城市框架。 */
.map-card { margin-top: 28rpx; padding: 30rpx; border: 3rpx solid #E7DCC5; border-radius: 28rpx; background: #FFFFFF; }
.map-card-head { display: flex; align-items: flex-start; justify-content: space-between; gap: 24rpx; }
.map-card-head > view { min-width: 0; display: flex; flex-direction: column; }
.map-title { font-size: 32rpx; font-weight: 850; }
.map-note { margin-top: 8rpx; color: #6F675B; font-size: 24rpx; line-height: 1.5; }
.map-progress { flex: none; padding: 8rpx 14rpx; border-radius: 12rpx; background: #E5F3EA; color: #246647; font-size: 22rpx; font-weight: 800; }
.city-map { display: grid; grid-template-columns: repeat(2, 1fr); gap: 12rpx; margin-top: 28rpx; padding: 22rpx; border-radius: 28rpx 42rpx 34rpx 48rpx; background: #FFF4DF; transform: rotate(-.5deg); }
.map-region { display: flex; align-items: center; justify-content: center; min-height: 96rpx; padding: 12rpx; box-sizing: border-box; border: 2rpx dashed #CFC2A8; border-radius: 40rpx 24rpx 36rpx 22rpx; color: #8B8271; font-size: 23rpx; text-align: center; }
.map-region:nth-child(3n+2) { border-radius: 24rpx 44rpx 22rpx 38rpx; }
.map-region.lit { border-style: solid; border-color: #2F7D58; background: #E5F3EA; color: #246647; font-weight: 800; }
.map-empty-copy { display: block; margin-top: 20rpx; color: #6F675B; font-size: 24rpx; line-height: 1.6; text-align: center; }

/* 概览卡 */
.summary-card {
  position: relative; display: flex; align-items: center; gap: 20rpx; margin-top: 24rpx; padding: 28rpx 28rpx;
  background: #FFFFFF; border: 3rpx solid #EFE5D0; border-radius: 36rpx;
  box-shadow: 0 4rpx 0 rgba(51, 46, 38, .05); box-sizing: border-box;
}
.poster-compact { display: flex; align-items: center; gap: 20rpx; min-height: 112rpx; margin-top: 32rpx; padding: 20rpx 26rpx; box-sizing: border-box; border: 2rpx solid #E7DCC5; border-radius: 24rpx; background: #FFF4DF; }
.poster-compact-copy { min-width: 0; flex: 1; display: flex; flex-direction: column; }
.poster-compact-title { font-size: 27rpx; font-weight: 800; }
.poster-compact-note { margin-top: 6rpx; color: #6F675B; font-size: 23rpx; }
.poster-compact-arrow { color: #2F7D58; font-size: 40rpx; }
.summary-medal {
  flex: none; width: 112rpx; height: 112rpx; box-sizing: border-box; transform: rotate(-5deg);
  display: flex; align-items: center; justify-content: center;
}
.summary-medal-ic { flex: none; }
.summary-body { flex: 1; min-width: 0; display: flex; align-items: center; }
.summary-col { flex: 1; min-width: 0; display: flex; align-items: baseline; justify-content: center; gap: 10rpx; }
.summary-num { font-size: 64rpx; font-weight: 900; line-height: 1; }
.summary-label { font-size: 24rpx; font-weight: 700; color: #8B8271; }
.summary-divider { width: 0; height: 56rpx; border-left: 3rpx dashed #E3D9C2; }
.summary-sticker {
  position: absolute; top: -16rpx; right: 28rpx; transform: rotate(2deg);
  padding: 6rpx 16rpx; border-radius: 14rpx;
  background: linear-gradient(180deg, #FFD05C 0%, #FFC64B 60%, #F7B93E 100%); box-shadow: 0 3rpx 0 #E0A62F;
  font-family: "Kaiti SC", "STKaiti", "PingFang SC", sans-serif; font-size: 25rpx; letter-spacing: 2rpx;
}

/* 海报入口 */
.poster-entry {
  position: relative; height: 500rpx; margin-top: 32rpx; overflow: hidden;
  border-radius: 44rpx; border: 3rpx solid #EFE5D0;
  box-shadow: 0 4rpx 0 rgba(51, 46, 38, .05); background: #332E26;
}
.poster-tape {
  position: absolute; top: 22rpx; left: 50%; width: 120rpx; height: 34rpx; z-index: 2;
  transform: translateX(-50%) rotate(-2deg); border-radius: 6rpx;
  background: rgba(255, 251, 243, .55); border: 2rpx solid rgba(255, 251, 243, .35);
  box-sizing: border-box;
}
.poster-sprinkles {
  position: absolute; right: 26rpx; bottom: 132rpx; width: 118rpx; height: 82rpx; z-index: 2; opacity: .9;
}
.poster-img { width: 100%; height: 100%; }
.poster-shade { position: absolute; inset: 0; background: linear-gradient(180deg, rgba(30, 26, 18, .1), rgba(30, 26, 18, .34) 46%, rgba(30, 26, 18, .86)); }
.poster-copy { position: absolute; left: 32rpx; right: 32rpx; top: 34rpx; display: flex; flex-direction: column; color: #FFFBF3; }
.poster-eyebrow { color: #FFC64B; font-size: 21rpx; font-weight: 800; letter-spacing: 2rpx; }
.poster-title { margin-top: 14rpx; font-family: "Kaiti SC", "STKaiti", "PingFang SC", sans-serif; font-size: 52rpx; font-weight: 900; }
.poster-subtitle { width: 72%; margin-top: 10rpx; color: #EFE5D0; font-size: 22rpx; line-height: 1.5; }
.poster-cta {
  position: absolute; left: 28rpx; right: 28rpx; bottom: 28rpx;
  display: flex; align-items: center; justify-content: center; gap: 12rpx;
  height: 84rpx; border-radius: 999rpx;
  background: linear-gradient(180deg, #4BAB7E 0%, #3E9B6F 52%, #35855A 100%);
  color: #FFFFFF; font-size: 27rpx; font-weight: 800; box-shadow: 0 6rpx 0 #2F7D58;
}
.poster-arrow { flex: none; }

/* 区块头 */
.sec { display: flex; align-items: center; gap: 16rpx; margin: 44rpx 4rpx 20rpx; }
.sec-ic { flex: none; }
.sec-title { position: relative; }
.sec-h { font-size: 34rpx; font-weight: 900; letter-spacing: 1rpx; white-space: nowrap; }
.sec-scribble { position: absolute; left: -4rpx; bottom: -14rpx; }
.sec-p { font-size: 23rpx; color: #8B8271; padding-top: 6rpx; white-space: nowrap; }

/* 加载 / 空态 */
.state-card {
  display: flex; flex-direction: column; align-items: center; gap: 14rpx;
  padding: 44rpx 24rpx; border-radius: 36rpx;
  background: #FFFFFF; border: 3rpx solid #EFE5D0; color: #8B8271; font-size: 23rpx;
}
.loading-dot { width: 20rpx; height: 20rpx; border-radius: 50%; background: #3E9B6F; animation: pulse 1s infinite alternate; }
.empty-card {
  display: flex; flex-direction: column; align-items: center; gap: 8rpx;
  padding: 44rpx 32rpx; border-radius: 36rpx;
  background: #FFFFFF; border: 3rpx dashed #EFE5D0;
}
.empty-trail { display: flex; align-items: center; gap: 10rpx; margin-bottom: 16rpx; }
.trail-stop {
  width: 92rpx; height: 92rpx; box-sizing: border-box;
  display: flex; align-items: center; justify-content: center;
}
.trail-stop.flip { transform: rotate(5deg) scaleX(-1); }
.trail-stop.flip .trail-ic { transform: scaleX(-1); }
.trail-ic { flex: none; }
.trail-squig { flex: none; }
.empty-title { font-size: 28rpx; font-weight: 800; color: #332E26; }
.empty-note { font-size: 22rpx; color: #8B8271; text-align: center; line-height: 1.6; }
.empty-cta {
  margin-top: 20rpx; display: flex; align-items: center; justify-content: center;
  height: 76rpx; padding: 0 56rpx; border-radius: 999rpx;
  background: linear-gradient(180deg, #4BAB7E 0%, #3E9B6F 52%, #35855A 100%);
  color: #FFFFFF; font-size: 26rpx; font-weight: 800;
  box-shadow: 0 6rpx 0 #2F7D58;
}

/* 时间线列表 */
.timeline-card {
  position: relative; background: #FFFFFF; border: 3rpx solid #EFE5D0; border-radius: 36rpx;
  box-shadow: 0 4rpx 0 rgba(51, 46, 38, .05);
}
.timeline-card::before {
  content: ''; position: absolute; left: 31rpx; top: 34rpx; bottom: 34rpx; width: 0;
  border-left: 3rpx dashed #E3D9C2;
}
.timeline-row {
  position: relative; display: flex; align-items: center; gap: 18rpx;
  padding: 22rpx 24rpx; border-bottom: 2rpx solid #F5EEDF;
}
.timeline-row:last-child { border-bottom: 0; }
.timeline-date { flex: none; width: 110rpx; display: flex; flex-direction: column; align-items: flex-start; gap: 8rpx; }
.tl-star { width: 30rpx; height: 30rpx; margin-left: -8rpx; }
.timeline-date > text { font-size: 19rpx; font-weight: 700; color: #8B8271; line-height: 1.3; }
.tl-photo {
  flex: none; padding: 8rpx; border-radius: 16rpx; box-sizing: content-box;
  background: #FFFFFF; border: 3rpx solid #EFE5D0;
  box-shadow: 0 3rpx 0 rgba(51, 46, 38, .05); transform: rotate(-1.8deg);
}
.timeline-row:nth-child(even) .tl-photo { transform: rotate(1.5deg); }
.timeline-img { display: block; width: 140rpx; height: 96rpx; border-radius: 8rpx; background: #FFF4DF; }
.timeline-copy { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 6rpx; }
.tl-place { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 26rpx; font-weight: 800; }
.tl-area { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 21rpx; color: #8B8271; }
.tl-status { font-size: 20rpx; font-weight: 700; color: #2F7D58; }
.row-arrow { flex: none; font-size: 32rpx; font-weight: 700; color: #C9BFA8; }

@keyframes pulse { to { opacity: .25; transform: scale(.7); } }
</style>
