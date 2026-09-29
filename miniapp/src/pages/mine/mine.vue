<template>
  <view class="mine-page">
    <view class="mine-head" :style="{ paddingTop: `${navLayout.contentTop}px` }">
      <view class="state-title">
        <text>健康概览</text>
        <view class="profile-inline" @tap="goProfile">
          <text>{{ profile.nickname || '城市记录者' }}</text>
          <image v-if="profile.avatar_url" class="avatar" :src="profile.avatar_url" mode="aspectFill" />
          <view v-else class="avatar avatar-text">{{ avatarText }}</view>
          <text class="settings" @tap.stop="showSettings">⚙</text>
        </view>
      </view>
    </view>

    <view class="body">
      <view class="status-grid">
        <view class="score-card">
          <text>本月贡献</text>
          <view><text>{{ contribution.month_points }}</text><view class="score-ring"><view></view></view></view>
          <text>{{ nextLevelCopy }}</text>
        </view>
        <view class="streak-card">
          <text>累计记录</text><text>{{ metrics.records }}</text><text>条</text><text>认真记录每一店</text>
        </view>
      </view>

      <view class="mini-stats">
        <view><text>♜</text><text>{{ metrics.records }}</text><text>我的记录</text></view>
        <view><text>♨</text><text>{{ metrics.helpful }}</text><text>有用投票</text></view>
        <view><text>ϟ</text><text>{{ unlockedBadges.length }}</text><text>已解锁徽章</text></view>
      </view>

      <view class="quick-grid">
        <view v-for="item in quickActions" :key="item.label" class="quick-item" @tap="openAction(item)">
          <view class="quick-icon" :class="item.tone">{{ item.icon }}</view><text>{{ item.label }}</text><text>{{ item.desc }}</text>
        </view>
      </view>

      <view class="account-menu surface">
        <view v-for="item in accountActions" :key="item.label" @tap="openAction(item)"><text class="menu-mark">{{ item.icon }}</text><view><text>{{ item.label }}</text><text>{{ item.desc }}</text></view><text class="arrow">›</text></view>
      </view>

    </view>
  </view>
</template>

<script>
import { api } from '@/api/request.js'
import { getSelectedCity } from '@/store/city.js'
import { syncCustomTabBar } from '@/utils/tab-bar.js'

const previewBadges = [
  { code: 'first_note', name: '初次落笔', unlocked_at: null },
  { code: 'city_walker', name: '城市漫游', unlocked_at: null },
  { code: 'trusted_voice', name: '可靠声音', unlocked_at: null },
]

function getNavLayout() {
  const info = typeof uni.getWindowInfo === 'function' ? uni.getWindowInfo() : uni.getSystemInfoSync()
  const statusBarHeight = Number(info.statusBarHeight || 20)
  let menu = null
  try { menu = uni.getMenuButtonBoundingClientRect() } catch (error) { menu = null }

  const menuTop = menu && Number(menu.top) > 0 ? Number(menu.top) : statusBarHeight + 6
  const menuHeight = menu && Number(menu.height) > 0 ? Number(menu.height) : 32
  const menuBottom = menu && Number(menu.bottom) > menuTop ? Number(menu.bottom) : menuTop + menuHeight

  // Content begins below both the system status area (including Dynamic Island)
  // and the WeChat capsule. The extra gap keeps handwritten ascenders clear.
  return { contentTop: Math.max(statusBarHeight, menuBottom) + 16 }
}

export default {
  data() {
    return {
      navLayout: getNavLayout(),
      city: getSelectedCity(), profile: { nickname: uni.getStorageSync('nickname') || '林初', avatar_url: '' },
      contribution: { total_points: 0, month_points: 0 }, trust: { trust_level: 'new' }, badges: [],
      metrics: { records: 0, helpful: 0, badges: 0 },
      quickActions: [
        { label: '我的记录', desc: '查看提交与审核状态', icon: '记', tone: 'red', action: 'records' },
        { label: '草稿与审核', desc: '暂无待处理', icon: '审', tone: 'green', action: 'review' },
        { label: '举报与申诉', desc: '处理进度', icon: '诉', tone: 'ink', action: 'cases' },
      ],
      accountActions: [
        { label: '隐私保护政策', desc: '了解信息如何被使用', icon: '隐', url: '/pages/privacy/privacy' },
        { label: '用户服务协议', desc: '平台规则与内容规范', icon: '约', url: '/pages/agreement/agreement' },
        { label: '账号与安全', desc: '登录设备、退出账号', icon: '安', action: 'settings' },
      ],
    }
  },
  computed: {
    avatarText() { return (this.profile.nickname || '城').slice(0, 1) },
    trustLabel() { return ({ new: '新记录者', normal: '城市记录者', reliable: '可靠记录者', trusted: '资深记录者' })[this.trust.trust_level] || '城市记录者' },
    levelInitial() { return this.trustLabel().slice(0, 1) },
    progressPercent() { return Math.min(100, Math.max(8, Number(this.contribution.month_points || 0))) },
    nextLevelCopy() { return this.progressPercent >= 100 ? '本月目标已经完成，感谢每一次认真记录。' : `距离本月目标还差 ${100 - this.progressPercent} 点贡献值` },
    unlockedBadges() { return this.badges.filter(item => item.unlocked_at) },
    displayBadges() { return this.badges.length ? this.badges : previewBadges },
  },
  onShow() { this.navLayout = getNavLayout(); syncCustomTabBar(this, 3); this.city = getSelectedCity(); this.loadMine() },
  methods: {
    async loadMine() {
      if (!uni.getStorageSync('token')) return
      const results = await Promise.allSettled([api.getProfile(), api.getContribution(), api.getBadges(), api.getMyRecords({ limit: 20 })])
      const [profileResult, contributionResult, badgeResult, recordResult] = results
      if (profileResult.status === 'fulfilled') this.profile = profileResult.value
      if (contributionResult.status === 'fulfilled') {
        this.contribution = contributionResult.value.account || this.contribution
        this.trust = contributionResult.value.trust || this.trust
      }
      if (badgeResult.status === 'fulfilled') this.badges = badgeResult.value.badges || []
      if (recordResult.status === 'fulfilled') {
        const records = recordResult.value.items || []
        this.metrics.records = records.length
        this.metrics.helpful = records.reduce((sum, item) => sum + Number(item.record.helpful_count || 0), 0)
        const submittedCount = records.filter(item => ['pending_review', 'published', 'rejected'].includes(item.record.publish_status)).length
        const pendingCount = records.filter(item => item.record.publish_status === 'pending_review').length
        this.quickActions[0].desc = submittedCount ? `${submittedCount} 条提交记录` : '还没有提交记录'
        this.quickActions[1].desc = pendingCount ? `${pendingCount} 条待审核` : '暂无待审核'
      }
      this.metrics.badges = this.unlockedBadges.length
      const failed = results.find(item => item.status === 'rejected')
      if (failed) console.warn('[mine] some data use preview:', failed.reason && failed.reason.message ? failed.reason.message : failed.reason)
    },
    goProfile() { uni.navigateTo({ url: '/pages/profile/profile' }) },
    openAction(item) {
      if (item.url) return uni.navigateTo({ url: item.url })
      if (item.action === 'records') return uni.navigateTo({ url: '/pages/my-records/my-records' })
      if (item.action === 'review') return uni.navigateTo({ url: '/pages/my-records/my-records?status=pending_review' })
      if (item.action === 'cases') return uni.navigateTo({ url: '/pages/cases/cases' })
      if (item.action === 'settings') return this.showSettings()
      uni.showToast({ title: `${item.label}将在数据接入后开放`, icon: 'none' })
    },
    showSettings() {
      uni.showActionSheet({ itemList: ['编辑个人资料', '清理本地缓存', '退出登录'], success: ({ tapIndex }) => {
        if (tapIndex === 0) this.goProfile()
        if (tapIndex === 1) uni.showToast({ title: '缓存已整理', icon: 'success' })
        if (tapIndex === 2) { uni.removeStorageSync('token'); uni.reLaunch({ url: '/pages/login/login' }) }
      } })
    },
  },
}
</script>

<style scoped>
.mine-page { min-height: 100vh; padding-bottom: calc(128rpx + env(safe-area-inset-bottom)); background: #F3F0EA; }.hero { position: relative; overflow: hidden; padding: calc(env(safe-area-inset-top) + 30rpx) 32rpx 36rpx; border-radius: 0 0 42rpx 42rpx; background: #222A24; color: #fff; }.hero-orbit { position: absolute; border: 1rpx solid rgba(255,255,255,.08); border-radius: 50%; }.orbit-one { right: -180rpx; top: -230rpx; width: 570rpx; height: 570rpx; }.orbit-two { right: -40rpx; top: -80rpx; width: 310rpx; height: 310rpx; }.hero-head { position: relative; z-index: 1; display: flex; align-items: center; justify-content: space-between; }.eyebrow { color: #B7CDBB; font-size: 17rpx; font-weight: 700; letter-spacing: 5rpx; }.settings { padding: 8rpx 4rpx 18rpx 24rpx; color: rgba(255,255,255,.75); font-size: 30rpx; letter-spacing: 4rpx; }.identity { position: relative; z-index: 1; display: flex; align-items: center; gap: 20rpx; margin-top: 28rpx; }.avatar { width: 106rpx; height: 106rpx; flex: none; border: 3rpx solid rgba(255,255,255,.7); border-radius: 50%; }.avatar-text { display: flex; align-items: center; justify-content: center; background: #C84B31; font-size: 38rpx; font-weight: 700; }.identity-copy { display: flex; flex: 1; flex-direction: column; }.nickname { font-size: 38rpx; font-weight: 720; }.identity-copy text:last-child { margin-top: 7rpx; color: rgba(255,255,255,.58); font-size: 21rpx; }.profile-link { color: rgba(255,255,255,.68); font-size: 20rpx; }.hero-metrics { position: relative; z-index: 1; display: grid; grid-template-columns: repeat(3, 1fr); margin-top: 40rpx; padding-top: 28rpx; border-top: 1rpx solid rgba(255,255,255,.12); }.hero-metrics view { display: flex; flex-direction: column; border-right: 1rpx solid rgba(255,255,255,.12); text-align: center; }.hero-metrics view:last-child { border: 0; }.hero-metrics view text:first-child { font-size: 37rpx; font-weight: 720; }.hero-metrics view text:last-child { margin-top: 5rpx; color: rgba(255,255,255,.52); font-size: 19rpx; }
.body { padding-top: 24rpx; }.month-card { padding: 30rpx; border-radius: 26rpx; background: #C84B31; color: #fff; }.month-top { display: flex; align-items: center; justify-content: space-between; }.month-top > view:first-child { display: flex; flex-direction: column; }.section-kicker { color: rgba(255,255,255,.64); font-size: 17rpx; font-weight: 700; letter-spacing: 3rpx; }.month-points { margin-top: 5rpx; font-size: 54rpx; font-weight: 750; }.month-points text { font-size: 20rpx; font-weight: 500; }.level-seal { width: 90rpx; height: 90rpx; display: flex; flex-direction: column; align-items: center; justify-content: center; border: 1rpx solid rgba(255,255,255,.5); border-radius: 50%; }.level-seal text:first-child { font-size: 27rpx; font-weight: 700; }.level-seal text:last-child { margin-top: 2rpx; font-size: 14rpx; }.progress { height: 7rpx; margin-top: 24rpx; overflow: hidden; border-radius: 5rpx; background: rgba(255,255,255,.23); }.progress view { height: 100%; border-radius: 5rpx; background: #fff; }.progress-copy { display: block; margin-top: 12rpx; color: rgba(255,255,255,.72); font-size: 19rpx; }
.quick-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 14rpx; margin-top: 20rpx; }.quick-item { min-height: 178rpx; display: flex; flex-direction: column; padding: 22rpx; border: 1rpx solid #E0DBD3; border-radius: 22rpx; background: #FFFEFB; }.quick-icon { width: 54rpx; height: 54rpx; margin-bottom: 20rpx; border-radius: 15rpx; line-height: 54rpx; text-align: center; font-size: 21rpx; font-weight: 700; }.quick-icon.red { background: #F7E5DF; color: #B7462F; }.quick-icon.green { background: #E5EEE5; color: #4F6B57; }.quick-icon.sand { background: #F2EAD9; color: #8B682D; }.quick-icon.ink { background: #E8E8E5; color: #343633; }.quick-item > text:nth-child(2) { font-size: 27rpx; font-weight: 680; }.quick-item > text:last-child { margin-top: 6rpx; color: #8A847D; font-size: 19rpx; }
.badges-section { margin-top: 38rpx; }.section-head { display: flex; align-items: flex-end; justify-content: space-between; margin: 0 3rpx 16rpx; }.section-head > view { display: flex; flex-direction: column; }.section-head .section-kicker { color: #9A5A48; }.section-head view text:last-child { margin-top: 5rpx; font-size: 31rpx; font-weight: 700; }.section-head > text { color: #8B857E; font-size: 20rpx; }.badges-scroll { width: calc(100vw - 30rpx); }.badge-row { display: inline-flex; gap: 13rpx; padding-right: 30rpx; }.badge-card { width: 190rpx; display: flex; flex-direction: column; align-items: center; padding: 24rpx 14rpx; border: 1rpx solid #DCD7CF; border-radius: 21rpx; background: #FFFEFB; }.badge-mark { width: 72rpx; height: 72rpx; margin-bottom: 14rpx; border: 1rpx solid #C84B31; border-radius: 50%; color: #C84B31; line-height: 70rpx; text-align: center; font-size: 26rpx; font-weight: 700; }.badge-card > text:nth-child(2) { font-size: 23rpx; font-weight: 650; }.badge-card > text:last-child { margin-top: 5rpx; color: #928C84; font-size: 17rpx; }.badge-card.locked { opacity: .48; filter: grayscale(1); }
.account-menu { margin-top: 32rpx; padding: 0 24rpx; }.account-menu > view { min-height: 106rpx; display: flex; align-items: center; gap: 18rpx; border-bottom: 1rpx solid #E3DED6; }.account-menu > view:last-child { border: 0; }.menu-mark { width: 50rpx; height: 50rpx; border-radius: 50%; background: #EFEBE5; color: #4E4A45; line-height: 50rpx; text-align: center; font-size: 18rpx; font-weight: 700; }.account-menu view view { display: flex; flex: 1; flex-direction: column; }.account-menu view view text:first-child { font-size: 25rpx; font-weight: 650; }.account-menu view view text:last-child { margin-top: 4rpx; color: #918B83; font-size: 18rpx; }.arrow { color: #938D85; font-size: 38rpx; }.footer-note { display: block; margin: 34rpx auto 10rpx; color: #9B958D; text-align: center; font-size: 18rpx; letter-spacing: 1rpx; }

.mine-page { position: relative; overflow: hidden; background: #F3F3EF; color: #1B1B19; }
.mine-art { position: absolute; top: 560rpx; right: 0; width: 100%; height: 1180rpx; opacity: .075; }
.hero, .body { position: relative; z-index: 1; }
.hero { border-bottom: 1rpx solid #D7D5CF; border-radius: 0; background: rgba(248,248,244,.96); color: #1B1B19; }
.hero-orbit { border-color: rgba(66,106,84,.09); }
.eyebrow { color: #426A54; }
.settings { color: #5F5E5A; }
.avatar { border-color: #FFFFFF; box-shadow: 0 6rpx 18rpx rgba(25,24,21,.1); }
.avatar-text { background: #DFDDD8; color: #57534E; }
.identity-copy text:last-child { color: #666560; }
.profile-link { color: #D64B32; font-weight: 650; }
.hero-metrics { border-top-color: #D7D5CF; }
.hero-metrics view { border-right-color: #D7D5CF; }
.hero-metrics view text:last-child { color: #666560; }
.month-card { background: #1B1B19; }
.month-card .progress view { background: #D64B32; }
.quick-item, .badge-card, .account-menu { border-color: #D7D5CF; background: #F8F8F4; }
.quick-icon.red { background: #F4E3DE; color: #D64B32; }
.quick-icon.green { background: #E3EAE2; color: #426A54; }
.quick-icon.sand { background: #E7E0D4; color: #745F43; }
.section-head .section-kicker { color: #426A54; }
.badge-mark { border-color: #D64B32; color: #D64B32; }
.menu-mark { background: #E3EAE2; color: #426A54; }

.mine-page { padding: 0 24rpx calc(145rpx + env(safe-area-inset-bottom)); background: #f3f8ef; }
.mine-head { padding-right: 8rpx; padding-left: 8rpx; }
.avatar { width: 62rpx; height: 62rpx; border: 6rpx solid #d9eed2; box-shadow: none; }
.avatar-text { display: flex; align-items: center; justify-content: center; background: #b8e5ad; color: #111; font-size: 25rpx; font-weight: 850; }
.settings { padding: 12rpx 0 12rpx 2rpx; color: #111; font-size: 28rpx; }
.state-title { display: flex; align-items: center; justify-content: space-between; gap: 16rpx; }
.state-title > text:first-child { flex: none; font-family: "Kaiti SC", STKaiti, cursive; font-size: 57rpx; font-weight: 900; line-height: 1; }
.profile-inline { min-width: 0; display: flex; align-items: center; justify-content: flex-end; gap: 11rpx; }
.profile-inline > text:first-child { max-width: 150rpx; overflow: hidden; font-size: 21rpx; font-weight: 800; text-overflow: ellipsis; white-space: nowrap; }
.body { padding: 24rpx 0 0; }
.status-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 13rpx; }
.score-card, .streak-card { min-height: 300rpx; display: flex; flex-direction: column; padding: 27rpx; border-radius: 34rpx; }
.score-card { background: linear-gradient(145deg,#d9ff61,#b9f532); }
.score-card > text:first-child, .streak-card > text:first-child { font-size: 24rpx; font-weight: 850; }
.score-card > view { display: flex; align-items: center; justify-content: space-between; margin-top: 17rpx; }
.score-card > view > text { font-size: 83rpx; font-weight: 950; line-height: 1; }
.score-card > text:last-child { margin-top: auto; font-size: 18rpx; }
.score-ring { position: relative; width: 82rpx; height: 82rpx; border: 13rpx solid rgba(255,255,255,.7); border-left-color: #5baa31; border-radius: 50%; transform: rotate(35deg); }
.streak-card { position: relative; background: #11130f; color: #fff; }
.streak-card > text:nth-child(2) { margin-top: 19rpx; font-size: 82rpx; font-weight: 950; line-height: 1; }
.streak-card > text:nth-child(3) { position: absolute; left: 122rpx; top: 103rpx; font-size: 23rpx; }
.streak-card > text:last-child { margin-top: auto; color: rgba(255,255,255,.6); font-size: 18rpx; }
.mini-stats { display: grid; grid-template-columns: repeat(3,1fr); gap: 13rpx; margin-top: 14rpx; }
.mini-stats > view { min-width: 0; height: 145rpx; display: flex; flex-direction: column; justify-content: center; padding: 19rpx; border-radius: 29rpx; background: #fff; }
.mini-stats > view > text:first-child { color: #62ba37; font-size: 25rpx; }
.mini-stats > view > text:nth-child(2) { margin-top: 7rpx; overflow: hidden; font-size: 31rpx; font-weight: 900; text-overflow: ellipsis; white-space: nowrap; }
.mini-stats > view > text:last-child { margin-top: 4rpx; color: #858a84; font-size: 16rpx; }
.quick-grid { grid-template-columns: 1fr 1fr; gap: 13rpx; margin-top: 14rpx; }
.quick-item { min-height: 135rpx; padding: 18rpx 20rpx; border: 0; border-radius: 27rpx; background: #fff; }
.quick-icon { width: 44rpx; height: 44rpx; margin-bottom: 9rpx; border-radius: 13rpx; line-height: 44rpx; }
.quick-item > text:nth-child(2) { font-size: 23rpx; }
.quick-item > text:last-child { font-size: 17rpx; }
.account-menu { margin-top: 25rpx; border: 0; border-radius: 28rpx; background: #fff; }
</style>
