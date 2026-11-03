<template>
  <view class="mine-page">
    <view class="mine-head" :style="{ paddingTop: `${navLayout.contentTop}px` }">
      <view class="head-row">
        <view class="page-title">
          <view class="title-ic" :style="dz('face', '#3E9B6F', 38)"></view>
          <text class="title-h">我的</text>
          <text class="title-p">认真吃饭的人</text>
        </view>
        <view class="profile-inline">
          <view class="profile-chip" @tap="goProfile">
            <image v-if="profile.avatar_url" class="avatar" :src="profile.avatar_url" mode="aspectFill" />
            <view v-else class="avatar avatar-text">{{ avatarText }}</view>
            <text class="profile-name">{{ profile.nickname || '城市记录者' }}</text>
            <view class="profile-caret" :style="dz('caret', '#8B8271', 18)"></view>
          </view>
          <view class="settings-btn" @tap="showSettings">
            <view class="settings-ic" :style="dz('pen', '#332E26', 28)"></view>
          </view>
        </view>
      </view>
    </view>

    <view class="resume-card" @tap="startRecord">
      <view class="resume-copy">
        <text class="resume-eyebrow">继续最近行为</text>
        <text class="resume-title">记录今天吃过的一家店</text>
        <text class="resume-note">草稿会留在到店记录流程中，完成后形成个人足迹</text>
      </view>
      <text class="resume-arrow">›</text>
    </view>

    <view class="status-grid">
      <view class="score-card">
        <view class="tape"></view>
        <view class="card-top">
          <text class="card-label">本月贡献</text>
          <view class="card-medal" :style="dzBg('blob-cream')"><view class="card-medal-ic" :style="dz('stamp', '#E0A62F', 32)"></view></view>
        </view>
        <text class="score-num">{{ contribution.month_points }}</text>
        <view class="progress"><view class="progress-fill" :style="{ width: progressPercent + '%' }"></view></view>
        <text class="card-note">{{ nextLevelCopy }}</text>
      </view>
      <view class="streak-card">
        <view class="card-top">
          <text class="card-label">累计记录</text>
          <view class="card-medal" :style="dzBg('blob-white')"><view class="card-medal-ic" :style="dz('pen', '#3E9B6F', 32)"></view></view>
        </view>
        <view class="streak-num-row">
          <text class="score-num">{{ metrics.records }}</text>
          <text class="streak-unit">条</text>
        </view>
        <text class="card-note">认真记录每一店</text>
      </view>
    </view>

    <view class="mini-stats">
      <view class="stat-col">
        <view class="stat-medal" :style="dzBg('blob-cream')"><view class="stat-ic" :style="dz('pin', '#3E9B6F', 34)"></view></view>
        <text class="stat-num">{{ metrics.records }}</text>
        <text class="stat-label">到店足迹</text>
      </view>
      <view class="stat-col">
        <view class="stat-medal" :style="dzBg('blob-cream')"><view class="stat-ic" :style="dz('check', '#2F7D58', 34)"></view></view>
        <text class="stat-num">{{ metrics.helpful }}</text>
        <text class="stat-label">有用投票</text>
      </view>
      <view class="stat-col">
        <view class="stat-medal" :style="dzBg('blob-cream')"><view class="stat-ic" :style="dz('flag', '#E0A62F', 34)"></view></view>
        <text class="stat-num">{{ unlockedBadges.length }}</text>
        <text class="stat-label">已解锁徽章</text>
      </view>
    </view>

    <view class="sec content-section">
      <view class="sec-ic" :style="dz('compass', '#3E9B6F', 38)"></view>
      <view class="sec-title">
        <text class="sec-h">我的内容</text>
        <view class="sec-scribble" :style="dz('scribble', '#3E9B6F', 56)"></view>
      </view>
    </view>
    <view class="list-card content-list">
      <view v-for="item in quickActions" :key="item.label" class="list-row" @tap="openAction(item)">
        <view class="row-medal" :style="dzBg('blob-cream')"><view class="row-ic" :style="dz(quickIcon(item), '#332E26', 32)"></view></view>
        <view class="row-copy">
          <text class="row-label">{{ item.label }}</text>
          <text class="row-desc">{{ item.desc }}</text>
        </view>
        <text class="row-arrow">›</text>
      </view>
    </view>

    <view class="sec account-section">
      <view class="sec-ic" :style="dz('stamp', '#3E9B6F', 38)"></view>
      <view class="sec-title">
        <text class="sec-h">账号与规范</text>
        <view class="sec-scribble" :style="dz('scribble', '#3E9B6F', 56)"></view>
      </view>
    </view>
    <view class="list-card account-list">
      <view v-for="item in accountActions" :key="item.label" class="list-row" @tap="openAction(item)">
        <view class="row-medal" :style="dzBg('blob-white')"><view class="row-ic" :style="dz(accountIcon(item), '#332E26', 32)"></view></view>
        <view class="row-copy">
          <text class="row-label">{{ item.label }}</text>
          <text class="row-desc">{{ item.desc }}</text>
        </view>
        <text class="row-arrow">›</text>
      </view>
    </view>

    <view class="page-quote">
      <view class="quote-swirl" :style="dz('swirl', '#C9BFA8', 34)"></view>
      <text>认真吃饭的人，值得被认真对待</text>
      <view class="quote-swirl flip" :style="dz('swirl', '#C9BFA8', 34)"></view>
    </view>
  </view>
</template>

<script>
import { api, clearSession } from '@/api/request.js'
import { getSelectedCity } from '@/store/city.js'
import { syncCustomTabBar } from '@/utils/tab-bar.js'
import { getPrimaryNavLayout } from '@/utils/nav-layout.js'
import { doodleBg, doodleStyle } from '@/utils/doodles.js'

const previewBadges = [
  { code: 'first_note', name: '初次落笔', unlocked_at: null },
  { code: 'city_walker', name: '城市漫游', unlocked_at: null },
  { code: 'trusted_voice', name: '可靠声音', unlocked_at: null },
]

export default {
  data() {
    return {
      navLayout: getPrimaryNavLayout(),
      city: getSelectedCity(), profile: { nickname: uni.getStorageSync('nickname') || '林初', avatar_url: '' },
      contribution: { total_points: 0, month_points: 0 }, trust: { trust_level: 'new' }, badges: [],
      metrics: { records: 0, helpful: 0, badges: 0 },
      quickActions: [
        { label: '最近记录', desc: '查看公开记录与审核状态', action: 'records' },
        { label: '我的足迹', desc: '查看已经点亮的城市与街区', action: 'footprints' },
        { label: '收藏的店与路线', desc: '继续上次想去的地方', action: 'favorites' },
        { label: '营养历史', desc: '查看私人饮食与营养数据', action: 'diet' },
      ],
      accountActions: [
        { label: '贡献与徽章', desc: '查看内容贡献和成长反馈', action: 'contribution' },
        { label: '举报与申诉', desc: '查看处理进度', action: 'cases' },
        { label: '隐私保护政策', desc: '了解信息如何被使用', icon: '隐', url: '/pages/privacy/privacy' },
        { label: '用户服务协议', desc: '平台规则与内容规范', icon: '约', url: '/pages/agreement/agreement' },
        { label: '社区内容规范', desc: '公开内容与审核规则', icon: '规', url: '/pages/community/community' },
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
  onShow() { this.navLayout = getPrimaryNavLayout(); syncCustomTabBar(this, 3); this.city = getSelectedCity(); this.loadMine() },
  methods: {
    dz(key, color, size) { return doodleStyle(key, color, size) },
    dzBg(key) { return doodleBg(key) },
    quickIcon(item) { return ({ records: 'quote', footprints: 'pin', favorites: 'star-hand', diet: 'mifan', cases: 'stamp' })[item.action] || 'check' },
    accountIcon(item) { return ({ settings: 'umbrella', '/pages/privacy/privacy': 'magnifier', '/pages/agreement/agreement': 'flag', '/pages/community/community': 'check' })[item.url || item.action] || 'check' },
    async loadMine() {
      if (!uni.getStorageSync('token')) return
      const results = await Promise.allSettled([api.getProfile(), api.getContribution(), api.getBadges(), api.getMyReviews({ limit: 20 }), api.getFootprintSummary()])
      const [profileResult, contributionResult, badgeResult, recordResult, footprintResult] = results
      if (profileResult.status === 'fulfilled') this.profile = profileResult.value
      if (contributionResult.status === 'fulfilled') {
        this.contribution = contributionResult.value.account || this.contribution
        this.trust = contributionResult.value.trust || this.trust
      }
      if (badgeResult.status === 'fulfilled') this.badges = badgeResult.value.badges || []
      if (recordResult.status === 'fulfilled') {
        const records = recordResult.value.items || []
        this.metrics.helpful = records.reduce((sum, item) => sum + Number(item.review.helpful_count || 0), 0)
        const submittedCount = records.length
        const pendingCount = records.filter(item => item.review.publish_status === 'pending').length
        this.quickActions[0].desc = submittedCount ? `${submittedCount} 条餐厅评论` : '还没有公开评论'
        if (pendingCount) this.quickActions[0].desc += `，${pendingCount} 条待审核`
      }
      if (footprintResult.status === 'fulfilled') this.metrics.records = Number(footprintResult.value.total_record_count || 0)
      this.metrics.badges = this.unlockedBadges.length
      const failed = results.find(item => item.status === 'rejected')
      if (failed) console.warn('[mine] some data use preview:', failed.reason && failed.reason.message ? failed.reason.message : failed.reason)
    },
    goProfile() { uni.navigateTo({ url: '/pages/profile/profile' }) },
    startRecord() { uni.switchTab({ url: '/pages/record/record' }) },
    openAction(item) {
      if (item.url) return uni.navigateTo({ url: item.url })
      if (item.action === 'records') return uni.navigateTo({ url: '/pages/my-records/my-records' })
      if (item.action === 'footprints') return uni.switchTab({ url: '/pages/footprints/footprints' })
      if (item.action === 'favorites') return uni.showToast({ title: '收藏列表将在路线数据接入后展示', icon: 'none' })
      if (item.action === 'diet') return uni.navigateTo({ url: '/pages/nutrition/nutrition' })
      if (item.action === 'cases') return uni.navigateTo({ url: '/pages/cases/cases' })
      if (item.action === 'contribution') return uni.showToast({ title: `${this.contribution.total_points || 0} 点贡献 · ${this.unlockedBadges.length} 枚徽章`, icon: 'none' })
      if (item.action === 'settings') return this.showSettings()
      uni.showToast({ title: `${item.label}将在数据接入后开放`, icon: 'none' })
    },
    showSettings() {
      uni.showActionSheet({ itemList: ['编辑个人资料', '清理本地缓存', '退出登录', '注销账号'], success: ({ tapIndex }) => {
        if (tapIndex === 0) this.goProfile()
        if (tapIndex === 1) uni.showToast({ title: '缓存已整理', icon: 'success' })
        if (tapIndex === 2) { clearSession(); uni.reLaunch({ url: '/pages/login/login' }) }
        if (tapIndex === 3) this.confirmDeleteAccount()
      } })
    },
    confirmDeleteAccount() {
      uni.showModal({
        title: '永久注销账号？',
        content: '账号、私人记录、公开记录、图片、消费凭证和其他关联数据将被永久删除，且无法恢复。',
        confirmText: '永久注销',
        confirmColor: '#C43D2D',
        success: async ({ confirm }) => {
          if (!confirm) return
          try {
            await api.deleteAccount()
            uni.clearStorageSync()
            uni.showToast({ title: '账号已注销', icon: 'success' })
            setTimeout(() => uni.reLaunch({ url: '/pages/login/login' }), 500)
          } catch (error) {
            uni.showToast({ title: error.message || '注销失败，请稍后重试', icon: 'none' })
          }
        },
      })
    },
  },
}
</script>

<style scoped>
/* —— V6 贴纸手帐：token 同 home.vue —— */
.mine-page {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  padding: 0 32rpx calc(150rpx + env(safe-area-inset-bottom));
  overflow: hidden;
  background: linear-gradient(180deg, #FFFDF9 0%, #FFFBF3 36%, #FFF3DF 100%);
  color: #332E26;
  font-family: "PingFang SC", "Helvetica Neue", Helvetica, sans-serif;
}
.mine-head { order: 1; }
.resume-card { order: 2; display: flex; align-items: center; gap: 20rpx; margin-top: 28rpx; padding: 30rpx; border-radius: 28rpx; background: #2F7D58; color: #FFFFFF; }
.resume-copy { min-width: 0; flex: 1; display: flex; flex-direction: column; }
.resume-eyebrow { color: #E5F3EA; font-size: 22rpx; font-weight: 700; }
.resume-title { margin-top: 8rpx; font-size: 30rpx; font-weight: 850; }
.resume-note { margin-top: 8rpx; color: rgba(255,255,255,.82); font-size: 23rpx; line-height: 1.5; }
.resume-arrow { flex: none; font-size: 42rpx; }
.content-section { order: 3; }
.content-list { order: 4; }
.status-grid { order: 5; }
.mini-stats { order: 6; }
.account-section { order: 7; }
.account-list { order: 8; }
.page-quote { order: 9; }

/* 头部 */
.mine-head { margin-bottom: 8rpx; }
.head-row { display: flex; align-items: center; gap: 16rpx; }
.page-title { min-width: 0; display: flex; align-items: center; gap: 14rpx; }
.title-ic { flex: none; }
.title-h { font-size: 34rpx; font-weight: 900; letter-spacing: 1rpx; }
.title-p { font-size: 23rpx; color: #8B8271; padding-top: 6rpx; }
.profile-inline { margin-left: auto; min-width: 0; display: flex; align-items: center; gap: 12rpx; }
.profile-chip {
  min-width: 0; display: flex; align-items: center; gap: 10rpx; height: 60rpx;
  padding: 0 20rpx 0 8rpx; box-sizing: border-box;
  background: #FFF4DF; border: 3rpx solid #EFE5D0; border-radius: 999rpx;
}
.avatar { width: 44rpx; height: 44rpx; border-radius: 50%; background: #FFFFFF; }
.avatar-text { display: flex; align-items: center; justify-content: center; font-size: 20rpx; font-weight: 800; }
.profile-name { max-width: 140rpx; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 22rpx; font-weight: 700; }
.profile-caret { flex: none; margin-top: 2rpx; }
.settings-btn {
  flex: none; width: 60rpx; height: 60rpx; border-radius: 50%; box-sizing: border-box;
  background: #FFF4DF; border: 3rpx solid #EFE5D0;
  display: flex; align-items: center; justify-content: center;
}

/* 双状态卡 */
.status-grid { display: flex; gap: 24rpx; margin-top: 24rpx; }
.score-card, .streak-card {
  position: relative; flex: 1; min-width: 0; min-height: 264rpx; display: flex; flex-direction: column;
  padding: 26rpx 28rpx; border-radius: 36rpx; box-sizing: border-box;
}
.score-card { background: #FFFFFF; border: 3rpx solid #EFE5D0; box-shadow: 0 4rpx 0 rgba(51, 46, 38, .05); }
.streak-card { background: #FFF4DF; border: 3rpx solid #EADFC2; }
.tape {
  position: absolute; top: -16rpx; left: 24rpx; width: 96rpx; height: 32rpx;
  transform: rotate(-3deg); border-radius: 6rpx;
  background: rgba(255, 196, 75, .5); border: 2rpx solid rgba(224, 166, 47, .35);
  box-sizing: border-box;
}
.card-top { display: flex; align-items: center; gap: 12rpx; }
.card-label { font-size: 24rpx; font-weight: 800; }
.card-medal {
  margin-left: auto; flex: none; width: 80rpx; height: 80rpx; box-sizing: border-box;
  display: flex; align-items: center; justify-content: center; transform: rotate(-5deg);
}
.card-medal-ic { flex: none; }
.score-num { margin-top: 18rpx; font-size: 76rpx; font-weight: 900; line-height: 1; }
.streak-num-row { display: flex; align-items: baseline; gap: 8rpx; margin-top: 18rpx; }
.streak-unit { font-size: 24rpx; font-weight: 700; color: #8B8271; }
.progress { height: 12rpx; margin-top: 22rpx; border-radius: 999rpx; background: #F1E7CE; overflow: hidden; }
.progress-fill { height: 100%; border-radius: 999rpx; background: linear-gradient(90deg, #6FBF94 0%, #3E9B6F 60%, #2F7D58 100%); }
.card-note { margin-top: auto; padding-top: 16rpx; font-size: 20rpx; color: #8B8271; line-height: 1.5; }

/* 三列小数据 */
.mini-stats {
  display: flex; margin-top: 24rpx; padding: 28rpx 0;
  background: #FFFFFF; border: 3rpx solid #EFE5D0; border-radius: 36rpx;
  box-shadow: 0 4rpx 0 rgba(51, 46, 38, .05);
}
.stat-col { flex: 1; min-width: 0; display: flex; flex-direction: column; align-items: center; gap: 8rpx; }
.stat-medal {
  width: 96rpx; height: 96rpx; margin-bottom: 6rpx; box-sizing: border-box;
  display: flex; align-items: center; justify-content: center;
}
.stat-col:nth-child(1) .stat-medal { transform: rotate(-5deg); }
.stat-col:nth-child(3) .stat-medal { transform: rotate(5deg); }
.stat-ic { flex: none; }
.stat-num { font-size: 34rpx; font-weight: 900; line-height: 1; }
.stat-label { font-size: 20rpx; color: #8B8271; }

/* 区块头 */
.sec { display: flex; align-items: center; gap: 16rpx; margin: 44rpx 4rpx 20rpx; }
.sec-ic { flex: none; }
.sec-title { position: relative; }
.sec-h { font-size: 34rpx; font-weight: 900; letter-spacing: 1rpx; }
.sec-scribble { position: absolute; left: -4rpx; bottom: -14rpx; }

/* 列表卡 */
.list-card {
  background: #FFFFFF; border: 3rpx solid #EFE5D0; border-radius: 36rpx;
  box-shadow: 0 4rpx 0 rgba(51, 46, 38, .05);
}
.list-row { display: flex; align-items: center; gap: 20rpx; padding: 22rpx 28rpx; border-bottom: 2rpx solid #F5EEDF; }
.list-row:last-child { border-bottom: 0; }
.row-medal {
  flex: none; width: 80rpx; height: 80rpx; box-sizing: border-box;
  display: flex; align-items: center; justify-content: center;
}
.list-row:nth-child(even) .row-medal { transform: rotate(4deg); }
.row-ic { flex: none; }
.row-copy { flex: 1; min-width: 0; display: flex; flex-direction: column; gap: 4rpx; }
.row-label { font-size: 26rpx; font-weight: 800; }
.row-desc { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 21rpx; color: #8B8271; }
.row-arrow { flex: none; font-size: 32rpx; font-weight: 700; color: #C9BFA8; }

/* 页脚手写引言 */
.page-quote {
  display: flex; align-items: center; justify-content: center; gap: 16rpx;
  margin-top: 48rpx; color: #8B8271;
  font-family: "Kaiti SC", "STKaiti", "PingFang SC", sans-serif; font-size: 24rpx; letter-spacing: 2rpx;
}
.quote-swirl { flex: none; opacity: .7; }
.quote-swirl.flip { transform: scaleX(-1); }
</style>
