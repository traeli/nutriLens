<template>
  <view class="detail-page">
    <view v-if="loading" class="detail-state">正在打开这次体验…</view>
    <view v-else-if="error" class="detail-state" @tap="loadExperience">{{ error }}，点击重试</view>
    <template v-else>
    <view class="hero-wrap">
      <image class="hero" :src="record.image" mode="aspectFill" />
      <view class="back" :style="{ top: `${navTop}px` }" @tap="back">‹</view>
      <view class="more" :style="{ top: `${navTop}px` }">•••</view>
    </view>

    <view class="detail-sheet">
      <view class="title-row"><view><text class="place">{{ record.place }}</text><text class="place-meta">{{ record.city }} · {{ record.area }} · {{ record.visitMonth }}</text></view><text class="status">◒ {{ record.conclusion }}</text></view>
      <view class="author-row">
        <view class="author-avatar">{{ record.author.slice(0, 1) }}</view>
        <view class="author-copy"><text class="author">{{ record.author }}</text><text class="muted">{{ record.completeness }}</text></view>
        <text v-if="record.verified" class="verified">◇ 凭证已核验</text>
      </view>
      <view class="facts">
        <view><text class="fact-label">人均</text><text class="fact-value">¥{{ record.average }}</text></view>
        <view><text class="fact-label">等位</text><text class="fact-value">{{ record.wait }} 分钟</text></view>
        <view><text class="fact-label">用餐</text><text class="fact-value">{{ record.meal }}</text></view>
      </view>
      <text class="section-title">这次体验</text>
      <text class="content">{{ record.content }}</text>
      <view class="tags"><text v-for="tag in record.tags" :key="tag">{{ tag }}</text></view>
      <text class="update">更新于 {{ record.updated }}</text>
    </view>

    <view class="actions">
      <view @tap="openPlace"><text class="action-icon">⌖</text><text>打开地图</text></view>
      <view @tap="helpfulAction"><text class="action-icon">♥</text><text>有帮助 {{ record.helpful }}</text></view>
      <view @tap="outdated"><text class="action-icon">!</text><text>信息过时</text></view>
      <view @tap="report"><text class="action-icon">⚑</text><text>举报</text></view>
    </view>
    </template>
  </view>
</template>

<script>
import { api } from '@/api/request.js'
import { diningImages } from '@/mock/city-dining.js'

function getNavTop() {
  const info = typeof uni.getWindowInfo === 'function' ? uni.getWindowInfo() : uni.getSystemInfoSync()
  const statusBarHeight = Number(info.statusBarHeight || 20)
  let menu = null
  try { menu = uni.getMenuButtonBoundingClientRect() } catch (error) { menu = null }
  const menuBottom = menu && Number(menu.bottom) > 0 ? Number(menu.bottom) : statusBarHeight + 44
  return Math.max(statusBarHeight, menuBottom) + 10
}
export default {
  data() { return { navTop: getNavTop(), id: '', record: { author: '', tags: [], helpful: 0 }, helpful: false, loading: true, error: '' } },
  onLoad(options) { this.navTop = getNavTop(); this.id = options.id || ''; this.loadExperience() },
  methods: {
    async loadExperience() {
      if (!this.id || !/^\d+$/.test(String(this.id))) { this.loading = false; this.error = '体验编号无效'; return }
      this.loading = true
      this.error = ''
      try {
        const item = await api.getExperience(this.id)
        const version = item.version || {}
        const place = item.place || {}
        const conclusion = ({ recommend: '推荐', neutral: '一般', caution: '谨慎选择' })[version.conclusion] || '真实体验'
        const meal = ({ breakfast: '早餐', lunch: '午餐', dinner: '晚餐', late_night: '夜宵' })[version.meal_period] || '未填写'
        this.record = {
          place: place.name || '未命名地点', city: place.district || '', area: place.business_area || place.address || '', visitMonth: this.formatDate(version.visit_date),
          conclusion, author: (item.author && item.author.nickname) || '城市记录者', completeness: item.evidence_verified ? '消费凭证已核验' : '公开到店记录',
          verified: Boolean(item.evidence_verified), average: version.average_cost == null ? '—' : version.average_cost, wait: version.wait_minutes == null ? '—' : version.wait_minutes,
          meal, content: version.content || '记录者暂未补充文字。', tags: (item.tags || []).map(tag => tag.name), helpful: Number(item.record.helpful_count || 0),
          image: (item.media && item.media[0] && item.media[0].public_url) || diningImages.shop, updated: this.formatDate(item.record.updated_at),
          address: place.address || '', longitude: Number(place.longitude), latitude: Number(place.latitude),
        }
      } catch (error) { this.error = error.message || '加载失败' } finally { this.loading = false }
    },
    formatDate(value) {
      const date = new Date(value)
      return Number.isNaN(date.getTime()) ? '最近' : `${date.getMonth() + 1}月${date.getDate()}日`
    },
    back() { uni.navigateBack() },
    openPlace() {
      if (!Number.isFinite(this.record.longitude) || !Number.isFinite(this.record.latitude) || (!this.record.longitude && !this.record.latitude)) {
        uni.showToast({ title: '这家店暂时没有准确位置', icon: 'none' })
        return
      }
      uni.openLocation({
        latitude: this.record.latitude,
        longitude: this.record.longitude,
        name: this.record.place,
        address: this.record.address,
        scale: 18,
      })
    },
    async helpfulAction() {
      try {
        if (this.helpful) await api.unhelpfulExperience(this.id)
        else await api.helpfulExperience(this.id)
        this.helpful = !this.helpful
        this.record.helpful = Math.max(0, this.record.helpful + (this.helpful ? 1 : -1))
      } catch (error) { uni.showToast({ title: error.message || '操作失败', icon: 'none' }) }
    },
    outdated() {
      uni.showModal({ title: '标记信息过时', editable: true, placeholderText: '可选：说明哪里已经变化', success: async result => {
        if (!result.confirm) return
        try { await api.markExperienceOutdated(this.id, result.content || ''); uni.showToast({ title: '已提交', icon: 'success' }) } catch (error) { uni.showToast({ title: error.message || '提交失败', icon: 'none' }) }
      } })
    },
    report() {
      const labels = ['虚假经历', '侮辱诽谤', '隐私或肖像', '广告引流', '同行恶意', '内容过时', '其他']
      const codes = ['false_experience', 'defamation', 'privacy_portrait', 'ad', 'malicious_competitor', 'outdated', 'other']
      uni.showActionSheet({ itemList: labels, success: ({ tapIndex }) => {
        uni.showModal({ title: '补充举报说明', editable: true, placeholderText: '请描述具体问题（选填）', success: async result => {
          if (!result.confirm) return
          try { await api.createReport({ target_type: 'record', target_id: Number(this.id), reason_code: codes[tapIndex], description: result.content || '' }); uni.showToast({ title: '举报已提交', icon: 'success' }) } catch (error) { uni.showToast({ title: error.message || '举报失败', icon: 'none' }) }
        } })
      } })
    },
  },
}
</script>

<style scoped>
.detail-page { min-height: 100vh; background: #181816; padding-bottom: 116rpx; }
.detail-state { min-height: 100vh; display: flex; align-items: center; justify-content: center; padding: 40rpx; background: #F7F5F0; color: #756F68; font-size: 26rpx; text-align: center; }
.hero-wrap { position: relative; height: 550rpx; }
.hero { width: 100%; height: 100%; }
.back, .more { position: absolute; top: calc(env(safe-area-inset-top) + 20rpx); width: 68rpx; height: 68rpx; display: flex; align-items: center; justify-content: center; border-radius: 50%; background: rgba(24,24,22,.65); color: #fff; }
.back { left: 28rpx; font-size: 56rpx; padding-bottom: 8rpx; }
.more { right: 28rpx; font-size: 26rpx; letter-spacing: 4rpx; }
.detail-sheet { position: relative; margin-top: -18rpx; padding: 46rpx 36rpx 52rpx; border-radius: 30rpx 30rpx 0 0; background: #F7F5F0; }
.title-row { display: flex; justify-content: space-between; gap: 20rpx; padding-bottom: 32rpx; border-bottom: 1rpx solid #DDD8D0; }
.title-row > view { display: flex; flex-direction: column; }
.place { font-size: 44rpx; font-weight: 700; }
.place-meta { margin-top: 8rpx; color: #88837C; font-size: 24rpx; }
.status { align-self: center; flex: none; padding: 16rpx 20rpx; border-radius: 32rpx; background: #E4ECE2; color: #4F6B57; font-size: 27rpx; font-weight: 650; }
.author-row { display: flex; align-items: center; gap: 18rpx; padding: 30rpx 0; }
.author-avatar { width: 72rpx; height: 72rpx; display: flex; align-items: center; justify-content: center; border-radius: 50%; background: #4F6B57; color: #fff; }
.author-copy { display: flex; flex: 1; flex-direction: column; }
.author { font-size: 29rpx; font-weight: 650; }
.author-copy .muted { margin-top: 3rpx; font-size: 22rpx; }
.verified { padding: 12rpx 15rpx; border: 1rpx solid #C84B31; border-radius: 12rpx; color: #C84B31; font-size: 21rpx; }
.facts { display: grid; grid-template-columns: repeat(3, 1fr); padding: 24rpx 0; border-radius: 18rpx; background: #EFEBE4; }
.facts > view { display: flex; flex-direction: column; align-items: center; border-right: 1rpx solid #DAD5CD; }
.facts > view:last-child { border: 0; }
.fact-label { color: #858079; font-size: 22rpx; }
.fact-value { margin-top: 4rpx; font-size: 29rpx; font-weight: 650; }
.section-title { display: block; margin: 42rpx 0 18rpx; font-size: 34rpx; font-weight: 700; }
.content { font-size: 29rpx; line-height: 1.75; }
.tags { display: flex; flex-wrap: wrap; gap: 14rpx; margin-top: 28rpx; }
.tags text { padding: 13rpx 22rpx; border-radius: 28rpx; background: #F6E7E1; color: #B6432C; font-size: 23rpx; }
.update { display: block; margin-top: 36rpx; padding-top: 28rpx; border-top: 1rpx solid #DDD8D0; color: #99948D; font-size: 22rpx; }
.actions { position: fixed; z-index: 10; left: 0; right: 0; bottom: 0; height: calc(106rpx + env(safe-area-inset-bottom)); padding-bottom: env(safe-area-inset-bottom); display: grid; grid-template-columns: repeat(4, 1fr); background: #FFFEFB; box-shadow: 0 -8rpx 30rpx rgba(24,24,22,.08); }
.actions > view { display: flex; align-items: center; justify-content: center; gap: 9rpx; font-size: 23rpx; }
.actions > view + view { border-left: 1rpx solid #E1DDD6; }
.action-icon { font-size: 30rpx; }
.actions .active { color: #C84B31; }

.detail-page { background: #f5faec; }
.hero-wrap { height: 620rpx; }
.back, .more { background: rgba(12,13,11,.84); }
.detail-sheet { margin-top: -42rpx; padding: 42rpx 28rpx 54rpx; border-radius: 48rpx 48rpx 0 0; background: #f5faec; }
.title-row { border: 0; }
.place { font-family: "Kaiti SC", STKaiti, cursive; font-size: 47rpx; font-weight: 900; }
.status { border-radius: 33rpx; background: #c7ff35; color: #111; }
.author-row { padding: 22rpx 0; }
.author-avatar { background: #111; }
.verified { border-color: #9bd319; color: #4f7600; }
.facts { gap: 12rpx; padding: 0; background: transparent; }
.facts > view { min-height: 135rpx; justify-content: center; border: 0; border-radius: 24rpx; background: #fff; }
.fact-value { font-size: 36rpx; font-weight: 900; }
.section-title { font-family: "Kaiti SC", STKaiti, cursive; font-size: 38rpx; font-weight: 900; }
.content { color: #4f554e; font-size: 26rpx; }
.tags text { border: 2rpx solid #91c919; background: transparent; color: #26340e; }
.actions { left: 22rpx; right: 22rpx; bottom: calc(12rpx + env(safe-area-inset-bottom)); height: 96rpx; padding: 0; overflow: hidden; border-radius: 50rpx; background: #10110f; color: #fff; box-shadow: 0 10rpx 30rpx rgba(15,16,14,.2); }
.actions > view + view { border-left-color: rgba(255,255,255,.22); }
</style>
