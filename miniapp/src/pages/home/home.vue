<template>
  <view class="home-page">
    <view class="nav-space" :style="{ height: `${navLayout.totalHeight}px` }">
      <view class="topbar" :style="{ top: `${navLayout.top}px`, height: `${navLayout.height}px`, paddingRight: `${navLayout.right}px` }">
        <text class="brand">饭友记</text>
        <view class="city-trigger" @tap="chooseCity">
          <text class="pin">●</text><text>{{ city.name }}</text><text class="chevron">⌄</text>
        </view>
        <view class="avatar" @tap="goMine"><text>{{ avatarText }}</text></view>
      </view>
    </view>

    <view class="search-float" @tap="goSearch">
      <view class="search-mark"></view>
      <text>搜索店名、小吃或街区</text>
      <view class="scan-mark"><view></view><view></view><view></view><view></view></view>
    </view>

    <swiper class="route-swiper" :current="routeIndex" circular :duration="360" @change="onRouteChange">
      <swiper-item v-for="guide in routeGuides" :key="guide.id">
        <view class="route-card" @tap="openGuide(guide)">
          <view class="route-copy">
            <text class="route-tag">城市路线</text>
            <text class="route-title">{{ guide.title }}</text>
            <text class="route-meta">用一顿早饭，认识一座城市</text>
            <view class="route-go">→</view>
          </view>
          <image :src="guide.image || images.shop" mode="aspectFill" />
          <text class="route-note">Good<br/>Morning!</text>
        </view>
      </swiper-item>
    </swiper>

    <view class="section-head recent-heading">
      <text>每日推荐</text><text @tap="goRecommendedPlaces">查看更多 ›</text>
    </view>
    <view class="recent-list">
      <view v-if="!displayRecent.length" class="recent-empty">这座城市还没有公开体验，去留下第一条真实记录吧。</view>
      <view v-for="(item, index) in displayRecent" :key="item.placeId" class="recent-item" @tap="goPlaceDetail(item.placeId)">
        <image :src="item.image" mode="aspectFill" />
        <view class="recent-copy"><text>{{ item.place }}</text><text>{{ item.area }} · {{ item.time }}</text></view>
        <view class="recent-score"><text>{{ item.count }} 条真实记录</text><text>{{ item.verdict }}</text></view>
      </view>
    </view>

    <view class="guide-section">
      <view class="section-head guide-heading">
        <view><text>吃喝攻略</text><text>跟着路线，慢慢认识一座城</text></view>
        <text @tap="openRoute">查看路线 ›</text>
      </view>

      <view class="hand-route-card" v-if="featuredRoute" @tap="openGuide(featuredRoute)">
        <view class="route-card-head">
          <view><text>{{ featuredRoute.title }}</text><text>{{ featuredRoute.subtitle || featuredRoute.meta }}</text></view>
          <view class="route-open">→</view>
        </view>

        <view class="hand-map">
          <view class="street street-one"></view>
          <view class="street street-two"></view>
          <view class="street street-three"></view>
          <view class="river"></view>
          <view class="route-dash dash-one"></view>
          <view class="route-dash dash-two"></view>
          <view class="route-dash dash-three"></view>
          <view
            v-for="(stop, stopIndex) in featuredRoute.stops.slice(0, 4)"
            :key="stop.index"
            class="route-point"
            :class="'point-' + (stopIndex + 1)"
          >
            <text>{{ stop.index }}</text><text>{{ stop.name }}</text>
          </view>
          <text class="map-note">Walk, eat,<br/>remember.</text>
        </view>

        <view class="route-card-foot"><text>{{ featuredRoute.footnote || '跟着路线，慢慢认识一座城' }}</text><text>打开路线</text></view>
      </view>
    </view>

    <view class="community-section">
      <view class="section-head community-heading">
        <view><text>用户真实记录</text><text>亲自吃过，才值得被看见</text></view>
        <text @tap="goExperienceList">查看更多 ›</text>
      </view>

      <view class="community-list">
        <view v-for="item in displayCommunityRecords" :key="`community-${item.id}`" class="community-row" @tap="goDetail(item.id)">
          <image :src="item.image" mode="aspectFill" />
          <view class="community-copy">
            <view class="community-user"><text class="user-avatar">{{ (item.author || '食').slice(0, 1) }}</text><text>{{ item.author || '城市食客' }}</text><text>· {{ item.time }}</text></view>
            <text class="community-place">{{ item.place }}</text>
            <text class="community-note">“{{ item.excerpt || '这家店值得留下一条真实记录。' }}”</text>
          </view>
          <text class="community-arrow">›</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script>
import { api } from '@/api/request.js'
import { collectionCampaigns, diningImages } from '@/mock/city-dining.js'
import { getCityOptions, getSelectedCity, refreshCityOptions, setCityOptions, setSelectedCity } from '@/store/city.js'
import { syncCustomTabBar } from '@/utils/tab-bar.js'

function getNavLayout() {
  const info = typeof uni.getWindowInfo === 'function' ? uni.getWindowInfo() : uni.getSystemInfoSync()
  let menu = null
  try { menu = uni.getMenuButtonBoundingClientRect() } catch (error) { menu = null }
  const statusBarHeight = Number(info.statusBarHeight || 20)
  const top = menu && menu.top > 0 ? menu.top : statusBarHeight + 6
  const height = menu && menu.height > 0 ? menu.height : 32
  const bottom = menu && menu.bottom > top ? menu.bottom : top + height
  const pageGutter = Number(info.windowWidth || 375) * 26 / 750
  const right = menu && menu.left > 0
    ? Math.max(72, Number(info.windowWidth) - menu.left - pageGutter + 8)
    : 8
  return { top, height, right, totalHeight: bottom + 12 }
}

export default {
  data() {
    const cities = getCityOptions()
    const selected = getSelectedCity()
    return {
      navLayout: getNavLayout(),
      cities,
      city: selected,
      cityIndex: Math.max(0, cities.findIndex(item => item.code === selected.code)),
      routeIndex: 0,
      locationCityCode: selected.code,
      images: diningImages,
      campaigns: collectionCampaigns,
      recent: [],
      communityRecords: [],
      featuredRoute: null,
      fallbackFeaturedRoute: null,
      nearbyLocation: null,
      routeGuides: [
        { id: 'route-01', number: '01', tag: '早餐路线', title: '曹家渡的四站早餐', meta: '2.5小时 · 步行2.8km', distance: '2.8km', keyword: '曹家渡', stops: ['曹家渡', '武定路', '余姚路', '长寿路'] },
        { id: 'route-02', number: '02', tag: '小巷路线', title: '从平江路拐进寻常巷陌', meta: '1.8小时 · 步行1.6km', distance: '1.6km', keyword: '小巷', stops: ['临顿路', '钮家巷', '平江路'] },
      ],
    }
  },
  computed: {
    avatarText() { return (uni.getStorageSync('nickname') || '我').slice(0, 1) },
    citySlides() {
      const source = this.cities.length ? this.cities : [this.city]
      return source.map(item => ({ ...item, tagline: item.desc || '', hero: item.image || this.images.shop }))
    },
    displayRecent() {
      return this.recent.slice(0, 2).map((item, index) => ({
        ...item,
        area: item.area || '本城',
        time: item.time || '最近',
        image: item.image || (index ? this.images.shop : this.images.noodles),
        count: Number(item.count || 0),
        verdict: item.verdict || '真实推荐',
      }))
    },
    displayCommunityRecords() {
      return this.communityRecords.slice(0, 3).map((item, index) => ({
        ...item,
        author: item.author || '城市食客',
        time: item.time || '最近更新',
        image: item.image || (index % 2 ? this.images.cafe : this.images.noodles),
      }))
    },
  },
  onShow() {
    syncCustomTabBar(this, 0)
    this.navLayout = getNavLayout()
    this.cities = getCityOptions()
    const selected = getSelectedCity()
    this.city = selected
    this.cityIndex = Math.max(0, this.cities.findIndex(item => item.code === selected.code))
    this.loadHomeSummary()
  },
  methods: {
    chooseCity() {
      if (!this.cities.length) {
        this.refreshCities()
        return uni.showToast({ title: '城市列表加载中', icon: 'none' })
      }
      uni.showActionSheet({ itemList: this.cities.map(item => item.name), success: ({ tapIndex }) => this.selectCity(tapIndex) })
    },
    selectCity(index) {
      const selected = this.cities[index]
      if (!selected) return
      const changed = selected.code !== this.city.code
      this.city = selected
      this.cityIndex = index
      this.routeIndex = 0
      setSelectedCity(selected)
      if (changed) this.loadHomeSummary()
    },
    onCityChange(event) { this.selectCity(Number(event.detail.current || 0)) },
    onRouteChange(event) { this.routeIndex = Number(event.detail.current || 0) },
    async loadHomeSummary() {
      if (!this.city.code) {
        return
      }
      try {
        const data = await api.getHomeSummary(this.city.code)
        if (Array.isArray(data.cities) && data.cities.length) {
          setCityOptions(data.cities)
          this.cities = getCityOptions()
          this.city = getSelectedCity()
          this.cityIndex = Math.max(0, this.cities.findIndex(item => item.code === this.city.code))
        }
        this.routeGuides = Array.isArray(data.route_guides) ? data.route_guides : []
        this.fallbackFeaturedRoute = data.featured_route || null
        this.featuredRoute = this.fallbackFeaturedRoute
        const picks = (data.daily_picks || []).map(card => this.normalizeDailyPlace(card)).filter(item => item.place && item.placeId)
        const community = (data.community_records || []).map(card => this.normalizeSummaryCard(card, true)).filter(item => item.place)
        this.recent = picks.slice(0, 2)
        this.communityRecords = community.slice(0, 3)
        this.loadNearbyRoute()
      } catch (error) {
        console.warn('[home] use preview summary:', error && error.message ? error.message : error)
        this.refreshCities()
      }
    },
    async refreshCities() {
      try {
        const cities = await refreshCityOptions()
        if (!cities.length) throw new Error('暂无可用城市')
        const previousCode = this.city.code
        this.cities = cities
        this.city = getSelectedCity()
        this.cityIndex = Math.max(0, cities.findIndex(item => item.code === this.city.code))
        if (this.city.code !== previousCode) this.loadHomeSummary()
      } catch (error) {
        console.warn('[home] keep cached cities:', error && error.message ? error.message : error)
      }
    },
    normalizeSummaryCard(card, withAuthor) {
      return {
        id: card.record_id || card.id,
        placeId: card.place_id,
        place: card.place_name || '',
        area: card.area || '本城',
        excerpt: card.excerpt || '',
        time: this.formatDate(card.visited_at),
        image: card.image_url || '',
        score: card.conclusion_label || '到店',
        verdict: card.verdict || '真实记录',
        author: withAuthor && card.author ? card.author.nickname : '',
      }
    },
    normalizeDailyPlace(place) {
      return {
        placeId: place.place_id,
        place: place.name || '',
        area: place.area || '本城',
        time: this.formatDate(place.latest_visit_at),
        image: place.image_url || '',
        count: Number(place.experience_count || 0),
        verdict: Number(place.recommend_count || 0) > 0 ? `${place.recommend_count} 人推荐` : '真实到店',
      }
    },
    loadNearbyRoute() {
      const cityCode = this.city.code
      uni.getLocation({
        type: 'gcj02',
        success: async location => {
          const longitude = Number(location.longitude)
          const latitude = Number(location.latitude)
          if (!Number.isFinite(longitude) || !Number.isFinite(latitude)) return
          try {
            const route = await api.getNearbyRoute({ city_code: cityCode, longitude, latitude, limit: 4 })
            if (cityCode === this.city.code && Array.isArray(route.stops) && route.stops.length) {
              this.nearbyLocation = { longitude, latitude }
              this.featuredRoute = route
            }
          } catch (error) {
            console.warn('[home] nearby route unavailable:', error && error.message ? error.message : error)
          }
        },
        fail: () => {
          this.nearbyLocation = null
          this.featuredRoute = this.fallbackFeaturedRoute
        },
      })
    },
    formatDate(value) {
      const date = new Date(value)
      if (Number.isNaN(date.getTime())) return '最近'
      return `${date.getMonth() + 1}月${date.getDate()}日`
    },
    quickRecord(mode) { uni.setStorageSync('record_entry_mode', mode); uni.switchTab({ url: '/pages/record/record' }) },
    openGuide(guide) {
      if (guide && guide.source === 'nearby' && this.nearbyLocation) {
        const query = [
          'nearby=1',
          `city_code=${encodeURIComponent(this.city.code)}`,
          `longitude=${this.nearbyLocation.longitude}`,
          `latitude=${this.nearbyLocation.latitude}`,
        ].join('&')
        uni.navigateTo({ url: `/pages/route-detail/route-detail?${query}` })
        return
      }
      if (guide && /^\d+$/.test(String(guide.id || ''))) {
        uni.navigateTo({ url: `/pages/route-detail/route-detail?id=${guide.id}` })
        return
      }
      const keyword = (guide && (guide.keyword || guide.title)) || ''
      uni.navigateTo({ url: `/pages/search/search?keyword=${encodeURIComponent(keyword)}` })
    },
    openRoute() { this.openGuide(this.featuredRoute || this.routeGuides[this.routeIndex] || this.routeGuides[0]) },
    openCampaign(campaign) {
      uni.showModal({
        title: campaign.title,
        content: '告诉我们店名、位置和你亲自吃过的那一道。线索会先核验，再公开推荐。',
        confirmText: '去记录',
        success: result => { if (result.confirm) this.quickRecord('text') },
      })
    },
    goSearch() { uni.navigateTo({ url: '/pages/search/search' }) },
    goRecommendedPlaces() { uni.navigateTo({ url: '/pages/search/search?recommended=1' }) },
    goExperienceList() { uni.navigateTo({ url: '/pages/experience-list/experience-list' }) },
    goPlaceDetail(id) { if (id) uni.navigateTo({ url: `/pages/place-detail/place-detail?id=${id}` }) },
    goDetail(id) { if (id) uni.navigateTo({ url: `/pages/experience-detail/experience-detail?id=${id}` }) },
    goFootprints() { uni.switchTab({ url: '/pages/footprints/footprints' }) },
    goMine() { uni.switchTab({ url: '/pages/mine/mine' }) },
  },
  onShareAppMessage() { return { title: `饭友记｜在${this.city.name}认真吃饭`, path: '/pages/home/home' } },
}
</script>

<style scoped>
.home-page {
  min-height: 100vh;
  padding: 0 26rpx calc(128rpx + env(safe-area-inset-bottom));
  overflow: hidden;
  background: #F3F3EF;
  color: #151513;
}

.nav-space { position: relative; }
.topbar { position: absolute; left: 0; right: 0; display: flex; align-items: center; gap: 20rpx; }
.brand { flex: none; font-family: "Songti SC", STSong, serif; font-size: 43rpx; font-weight: 800; letter-spacing: 2rpx; line-height: 1; }
.city-trigger { min-width: 0; display: flex; align-items: center; gap: 9rpx; font-size: 25rpx; font-weight: 650; white-space: nowrap; }
.pin { color: #D64B32; font-size: 17rpx; }
.chevron { margin-top: -6rpx; font-size: 24rpx; }
.avatar { width: 52rpx; height: 52rpx; display: flex; flex: none; align-items: center; justify-content: center; margin-left: auto; border-radius: 50%; background: #DFDDD8; color: #57534E; font-size: 20rpx; font-weight: 700; }

.hero-wrap { position: relative; }
.city-swiper { width: 100%; height: 380rpx; }
.city-hero { position: relative; height: 100%; overflow: hidden; border-radius: 25rpx; }
.city-hero > image, .hero-shade { position: absolute; inset: 0; width: 100%; height: 100%; }
.hero-shade { background: linear-gradient(90deg, rgba(14,13,11,.66), rgba(14,13,11,.05) 58%, rgba(14,13,11,.36)); }
.hero-copy { position: absolute; left: 31rpx; bottom: 45rpx; display: flex; flex-direction: column; color: #fff; }
.hero-city { font-family: "Songti SC", STSong, serif; font-size: 68rpx; font-weight: 500; line-height: 1; }
.hero-line { margin-top: 18rpx; font-size: 24rpx; letter-spacing: 2rpx; }
.hero-dash { width: 35rpx; height: 3rpx; margin-top: 18rpx; background: rgba(255,255,255,.84); }
.location-copy { position: absolute; right: 26rpx; bottom: 39rpx; display: flex; align-items: flex-start; gap: 12rpx; color: #fff; }
.location-dot { width: 12rpx; height: 12rpx; margin-top: 7rpx; border: 4rpx solid #fff; border-radius: 50%; }
.location-copy > view:last-child { display: flex; flex-direction: column; align-items: flex-end; gap: 10rpx; }
.location-copy text:first-child { font-size: 21rpx; font-weight: 650; }
.location-copy text:last-child { color: rgba(255,255,255,.82); font-size: 19rpx; }
.hero-pages { position: absolute; left: 0; right: 0; bottom: 15rpx; z-index: 2; display: flex; justify-content: center; gap: 8rpx; }
.hero-pages > view { width: 8rpx; height: 8rpx; border-radius: 50%; background: rgba(255,255,255,.45); transition: width .25s ease; }
.hero-pages > view.active { width: 25rpx; border-radius: 5rpx; background: #fff; }

.search-float { position: relative; z-index: 3; min-height: 88rpx; display: flex; align-items: center; gap: 23rpx; margin: -18rpx 8rpx 0; padding: 0 29rpx; border: 1rpx solid rgba(255,255,255,.8); border-radius: 46rpx; background: rgba(249,249,247,.96); box-shadow: 0 12rpx 31rpx rgba(25,24,21,.1); }
.search-float > text { flex: 1; color: #6E6B67; font-size: 25rpx; }
.search-mark { position: relative; width: 27rpx; height: 27rpx; border: 4rpx solid #666560; border-radius: 50%; }
.search-mark::after { content: ''; position: absolute; right: -10rpx; bottom: -7rpx; width: 13rpx; height: 4rpx; border-radius: 4rpx; background: #666560; transform: rotate(45deg); }
.scan-mark { position: relative; width: 31rpx; height: 31rpx; }
.scan-mark > view { position: absolute; width: 10rpx; height: 10rpx; }
.scan-mark view:nth-child(1) { left: 0; top: 0; border-left: 3rpx solid #5F5E5A; border-top: 3rpx solid #5F5E5A; }
.scan-mark view:nth-child(2) { right: 0; top: 0; border-right: 3rpx solid #5F5E5A; border-top: 3rpx solid #5F5E5A; }
.scan-mark view:nth-child(3) { left: 0; bottom: 0; border-left: 3rpx solid #5F5E5A; border-bottom: 3rpx solid #5F5E5A; }
.scan-mark view:nth-child(4) { right: 0; bottom: 0; border-right: 3rpx solid #5F5E5A; border-bottom: 3rpx solid #5F5E5A; }

.section-head { display: flex; align-items: center; justify-content: space-between; }
.section-head > text:first-child { font-family: "Songti SC", STSong, serif; font-size: 37rpx; font-weight: 800; }
.section-head > text:last-child { min-height: 60rpx; display: flex; align-items: center; color: #5F5D59; font-size: 22rpx; }
.route-heading { margin: 35rpx 10rpx 10rpx; }
.route-swiper { width: calc(100% + 18rpx); height: 264rpx; margin-left: 8rpx; }
.route-card { height: 246rpx; display: grid; grid-template-columns: 56% 44%; margin-right: 14rpx; overflow: hidden; border-bottom: 1rpx solid #D7D5CF; background: #F8F8F4; }
.route-copy { min-width: 0; padding: 15rpx 14rpx 0 2rpx; }
.route-top { display: grid; grid-template-columns: 91rpx minmax(0, 1fr); gap: 8rpx; }
.route-number { color: #426A54; font-family: Georgia, serif; font-size: 62rpx; line-height: 1; }
.route-summary { min-width: 0; display: flex; flex-direction: column; }
.route-tag { align-self: flex-start; padding: 5rpx 10rpx; border-radius: 9rpx; background: #E3EAE2; color: #426A54; font-size: 16rpx; }
.route-title { margin-top: 8rpx; font-size: 25rpx; font-weight: 750; line-height: 1.25; }
.route-meta { margin-top: 6rpx; color: #706D68; font-size: 18rpx; white-space: nowrap; }
.route-stops { position: relative; display: grid; margin-top: 29rpx; }
.route-line { position: absolute; left: 20rpx; right: 20rpx; top: 8rpx; height: 3rpx; background: #426A54; }
.route-stop { position: relative; z-index: 1; min-width: 0; display: flex; flex-direction: column; align-items: center; gap: 12rpx; }
.route-stop > view { width: 16rpx; height: 16rpx; border: 4rpx solid #426A54; border-radius: 50%; background: #F8F8F4; }
.route-stop text { width: 100%; overflow: hidden; font-size: 16rpx; text-align: center; text-overflow: ellipsis; white-space: nowrap; }
.map-preview { position: relative; margin: 15rpx 0 15rpx 5rpx; overflow: hidden; border-radius: 17rpx 0 0 17rpx; background: #E9EEE8; }
.map-grid { position: absolute; inset: -20rpx; background-image: linear-gradient(rgba(255,255,255,.86) 2rpx, transparent 2rpx), linear-gradient(90deg, rgba(255,255,255,.86) 2rpx, transparent 2rpx); background-size: 53rpx 47rpx; transform: rotate(-12deg); }
.path { position: absolute; z-index: 1; height: 5rpx; border-radius: 5rpx; background: #426A54; transform-origin: left center; }
.path-one { left: 35rpx; top: 170rpx; width: 78rpx; transform: rotate(-31deg); }.path-two { left: 99rpx; top: 129rpx; width: 72rpx; transform: rotate(-5deg); }.path-three { left: 167rpx; top: 123rpx; width: 66rpx; transform: rotate(-43deg); }
.map-node { position: absolute; z-index: 2; width: 15rpx; height: 15rpx; border: 4rpx solid #426A54; border-radius: 50%; background: #F8F8F4; }
.node-one { left: 29rpx; top: 165rpx; }.node-two { left: 92rpx; top: 122rpx; }.node-three { left: 164rpx; top: 116rpx; }.node-four { left: 208rpx; top: 74rpx; }
.map-preview > text { position: absolute; right: 17rpx; bottom: 14rpx; z-index: 2; font-family: Georgia, serif; font-size: 18rpx; font-style: italic; transform: rotate(-7deg); }

.recent-heading { margin: 18rpx 10rpx 3rpx; padding-top: 22rpx; border-top: 1rpx solid #D7D5CF; }
.recent-list { margin: 0 10rpx; }
.recent-empty { padding: 36rpx 20rpx; color: #88827A; text-align: center; font-size: 22rpx; line-height: 1.6; }
.recent-item { min-height: 120rpx; display: flex; align-items: center; gap: 18rpx; border-bottom: 1rpx solid #D9D7D1; }
.recent-item > image { width: 145rpx; height: 91rpx; flex: none; border-radius: 11rpx; }
.recent-copy { min-width: 0; display: flex; flex: 1; flex-direction: column; }
.recent-copy text:first-child { font-size: 25rpx; font-weight: 750; }
.recent-copy text:nth-child(2), .recent-copy text:nth-child(3) { margin-top: 4rpx; overflow: hidden; color: #6F6C68; font-size: 18rpx; text-overflow: ellipsis; white-space: nowrap; }
.recent-score { width: 67rpx; display: flex; flex: none; flex-direction: column; align-items: center; }
.recent-score text:first-child { color: #D64B32; font-family: Georgia, serif; font-size: 32rpx; }
.recent-score text:last-child { color: #66635F; font-size: 15rpx; white-space: nowrap; }
.row-arrow { color: #4B4946; font-size: 31rpx; }

.campaign { position: relative; min-height: 144rpx; display: flex; align-items: center; margin: 34rpx 7rpx 0; overflow: hidden; border-radius: 18rpx; background: #E7E0D4; }
.campaign::after { content: ''; position: absolute; inset: 0; background: linear-gradient(90deg, rgba(242,237,227,.99) 0%, rgba(242,237,227,.97) 60%, rgba(242,237,227,.08) 100%); }
.campaign > image { position: absolute; right: 0; top: 0; width: 36%; height: 100%; }
.campaign-brush { position: relative; z-index: 2; width: 43rpx; min-height: 144rpx; flex: none; background: #D2573F; clip-path: polygon(0 0,100% 4%,82% 21%,100% 38%,75% 58%,100% 79%,84% 100%,0 100%); }
.campaign-copy { position: relative; z-index: 2; min-width: 0; display: flex; flex: 1; flex-direction: column; padding: 0 15rpx; }
.campaign-copy text:first-child { font-family: Georgia, serif; font-size: 16rpx; letter-spacing: 2rpx; }
.campaign-copy text:nth-child(2) { margin-top: 8rpx; font-size: 25rpx; font-weight: 750; white-space: nowrap; }
.campaign-copy text:last-child { margin-top: 7rpx; color: #5F5C57; font-size: 18rpx; }
.campaign-arrow { position: relative; z-index: 3; width: 45rpx; height: 45rpx; flex: none; margin-right: 218rpx; border-radius: 50%; background: #1B1B19; color: #fff; line-height: 42rpx; text-align: center; font-size: 25rpx; }

/* 新视觉语言：手作展示字体与现代功能界面。 */
.home-page { padding: 0 24rpx calc(142rpx + env(safe-area-inset-bottom)); background: #f5faec; }
.topbar { gap: 18rpx; }
.brand { font-family: "PingFang SC", sans-serif; font-size: 32rpx; font-weight: 900; }
.city-trigger { margin-left: auto; padding: 15rpx 22rpx; border-radius: 32rpx; background: rgba(20,20,18,.06); }
.pin { color: #171813; }
.avatar { display: none; }
.search-float { min-height: 92rpx; margin: 18rpx 4rpx 22rpx; border: 2rpx solid #dedfd9; background: #fff; box-shadow: none; }
.search-float > text { color: #969a98; }
.route-swiper { width: 100%; height: 300rpx; margin: 0; }
.route-card { position: relative; height: 280rpx; display: block; margin: 0 4rpx; overflow: hidden; border: 0; border-radius: 34rpx; background: #10110e; }
.route-card > image { position: absolute; right: 0; top: 0; width: 43%; height: 100%; }
.route-card::after { content: ''; position: absolute; inset: 0; background: linear-gradient(90deg,#10110e 0%,#10110e 52%,rgba(16,17,14,.1) 80%); }
.route-copy { position: relative; z-index: 2; width: 64%; height: 100%; display: flex; flex-direction: column; padding: 27rpx 25rpx; color: #fff; }
.route-tag { align-self: flex-start; padding: 8rpx 19rpx; border-radius: 25rpx; background: #c7ff35; color: #11120e; font-size: 21rpx; font-weight: 800; }
.route-title { margin-top: 17rpx; font-family: "Kaiti SC", STKaiti, cursive; font-size: 39rpx; font-weight: 900; line-height: 1.12; }
.route-meta { margin-top: 12rpx; color: rgba(255,255,255,.58); font-size: 20rpx; }
.route-go { position: absolute; right: 1rpx; bottom: 25rpx; width: 60rpx; height: 60rpx; border-radius: 50%; background: #c7ff35; color: #111; font-size: 38rpx; line-height: 56rpx; text-align: center; }
.route-note { position: absolute; z-index: 3; right: 150rpx; top: 35rpx; color: #c7ff35; font-family: cursive; font-size: 22rpx; font-style: italic; line-height: 1.05; transform: rotate(-8deg); }
.recent-heading { margin: 18rpx 6rpx 12rpx; padding: 0; border: 0; }
.section-head > text:first-child { font-family: "Kaiti SC", STKaiti, cursive; font-size: 39rpx; font-weight: 900; }
.recent-list { display: grid; grid-template-columns: 1fr 1fr; gap: 14rpx; margin: 0 4rpx; }
.recent-item { min-width: 0; display: flex; flex-wrap: wrap; gap: 0; padding-bottom: 15rpx; overflow: hidden; border: 0; border-radius: 27rpx; background: #fff; }
.recent-item > image { width: 100%; height: 184rpx; border-radius: 0; }
.recent-copy { width: 100%; flex: none; padding: 13rpx 13rpx 0; }
.recent-copy text:first-child { font-size: 23rpx; }
.recent-copy text:nth-child(2) { font-size: 17rpx; }
.recent-score { width: 100%; flex-direction: row; justify-content: space-between; padding: 7rpx 13rpx 0; }
.recent-score text:first-child { color: #111; font-family: inherit; font-size: 19rpx; font-weight: 800; }
.recent-score text:last-child { color: #777d75; }
.guide-section { margin: 42rpx 4rpx 0; }
.guide-heading { align-items: flex-end; margin-bottom: 16rpx; }
.guide-heading > view { display: flex; flex-direction: column; }
.guide-heading > view text:first-child { font-family: "Kaiti SC", STKaiti, cursive; font-size: 41rpx; font-weight: 900; }
.guide-heading > view text:last-child { margin-top: 5rpx; color: #777d75; font-size: 18rpx; }
.guide-heading > text { color: #60655f; font-size: 20rpx; }
.hand-route-card { overflow: hidden; border-radius: 34rpx; background: #11120f; color: #fff; }
.route-card-head { min-height: 125rpx; display: flex; align-items: center; gap: 20rpx; padding: 24rpx 26rpx; }
.route-card-head > view:first-child { min-width: 0; display: flex; flex: 1; flex-direction: column; }
.route-card-head > view:first-child text:first-child { font-family: "Kaiti SC", STKaiti, cursive; font-size: 32rpx; font-weight: 900; }
.route-card-head > view:first-child text:last-child { margin-top: 7rpx; color: rgba(255,255,255,.55); font-size: 18rpx; }
.route-open { width: 59rpx; height: 59rpx; flex: none; border-radius: 50%; background: #c7ff35; color: #111; font-size: 36rpx; line-height: 55rpx; text-align: center; }
.hand-map { position: relative; height: 330rpx; overflow: hidden; margin: 0 14rpx; border-radius: 28rpx; background: #c7ff35; }
.street { position: absolute; height: 16rpx; border: 4rpx solid rgba(255,255,255,.84); border-right: 0; border-left: 0; transform-origin: left center; }
.street-one { left: -20rpx; top: 85rpx; width: 760rpx; transform: rotate(-12deg); }
.street-two { left: 80rpx; top: -30rpx; width: 470rpx; transform: rotate(70deg); }
.street-three { left: 265rpx; top: 310rpx; width: 560rpx; transform: rotate(-54deg); }
.river { position: absolute; right: -50rpx; top: -80rpx; width: 150rpx; height: 520rpx; border-radius: 50%; border-left: 28rpx solid rgba(83,208,181,.48); transform: rotate(13deg); }
.route-dash { position: absolute; z-index: 2; height: 7rpx; border-radius: 8rpx; background: repeating-linear-gradient(90deg,#111 0,#111 14rpx,transparent 14rpx,transparent 25rpx); transform-origin: left center; }
.dash-one { left: 105rpx; top: 95rpx; width: 190rpx; transform: rotate(34deg); }
.dash-two { left: 270rpx; top: 205rpx; width: 170rpx; transform: rotate(-25deg); }
.dash-three { left: 420rpx; top: 135rpx; width: 190rpx; transform: rotate(29deg); }
.route-point { position: absolute; z-index: 3; display: flex; flex-direction: column; align-items: center; }
.route-point > text:first-child { width: 50rpx; height: 50rpx; border: 5rpx solid rgba(255,255,255,.9); border-radius: 50%; background: #111; color: #fff; font-size: 22rpx; font-weight: 900; line-height: 40rpx; text-align: center; }
.route-point > text:last-child { margin-top: 5rpx; padding: 4rpx 8rpx; border-radius: 10rpx; background: rgba(199,255,53,.88); color: #111; font-size: 17rpx; font-weight: 850; white-space: nowrap; }
.point-one { left: 58rpx; top: 44rpx; }.point-two { left: 245rpx; top: 165rpx; }.point-three { left: 399rpx; top: 77rpx; }.point-four { right: 45rpx; top: 211rpx; }
.map-note { position: absolute; left: 46rpx; bottom: 25rpx; color: #111; font-family: cursive; font-size: 19rpx; font-style: italic; line-height: 1.05; transform: rotate(-7deg); }
.route-card-foot { min-height: 82rpx; display: flex; align-items: center; justify-content: space-between; padding: 0 27rpx; }
.route-card-foot text:first-child { color: rgba(255,255,255,.58); font-size: 18rpx; }
.route-card-foot text:last-child { color: #c7ff35; font-size: 21rpx; font-weight: 800; }
.community-section { margin: 43rpx 4rpx 0; }
.community-heading { align-items: flex-end; margin-bottom: 14rpx; }
.community-heading > view { display: flex; flex-direction: column; }
.community-heading > view text:first-child { font-family: "Kaiti SC", STKaiti, cursive; font-size: 40rpx; font-weight: 900; }
.community-heading > view text:last-child { margin-top: 5rpx; color: #777d75; font-size: 18rpx; }
.community-heading > text { color: #60655f; font-size: 20rpx; }
.community-list { overflow: hidden; border-radius: 30rpx; background: #fff; }
.community-row { min-height: 158rpx; display: flex; align-items: center; gap: 17rpx; padding: 18rpx; border-bottom: 1rpx solid #e2e6de; }
.community-row:last-child { border-bottom: 0; }
.community-row > image { width: 126rpx; height: 112rpx; flex: none; border-radius: 20rpx; }
.community-copy { min-width: 0; display: flex; flex: 1; flex-direction: column; }
.community-user { display: flex; align-items: center; gap: 7rpx; color: #858a83; font-size: 16rpx; }
.community-user text:nth-child(2) { color: #4f554e; font-weight: 750; }
.user-avatar { width: 31rpx; height: 31rpx; border-radius: 50%; background: #dff5d5; color: #356d20; font-size: 15rpx; font-weight: 850; line-height: 31rpx; text-align: center; }
.community-place { margin-top: 8rpx; font-size: 23rpx; font-weight: 850; }
.community-note { margin-top: 5rpx; overflow: hidden; color: #686e67; font-size: 18rpx; text-overflow: ellipsis; white-space: nowrap; }
.community-arrow { color: #6e746d; font-size: 35rpx; }
</style>
