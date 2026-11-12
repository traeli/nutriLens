<template>
  <view class="mine-page">
    <view class="mine-head" :style="{ paddingTop: `${navLayout.contentTop}px` }">
      <view><text class="eyebrow">饭友记</text><text class="title">我的</text><text class="subtitle">收藏想去的地方，也给有用的信息点个赞</text></view>
      <view class="profile" @tap="goProfile"><image v-if="profile.avatar_url" :src="profile.avatar_url" mode="aspectFill" /><view v-else>{{ avatarText }}</view><text>{{ profile.nickname || (isLoggedIn ? '饭友' : '登录') }}</text></view>
    </view>
    <view class="discover-card" @tap="goDiscover"><view class="tape"></view><text class="card-kicker">编辑正在整理</text><text class="card-title">下一家好吃的店，可能就在发现里</text><text class="card-note">这里展示运营编辑统一整理的店铺信息。</text><view class="card-link">去发现看看 →</view></view>
    <text class="section-title">常用功能</text>
    <view class="list-card"><view v-for="item in actions" :key="item.label" class="row" @tap="openAction(item)"><view class="row-icon">{{ item.icon }}</view><view><text class="row-label">{{ item.label }}</text><text class="row-desc">{{ item.desc }}</text></view><text class="arrow">›</text></view></view>
    <text class="section-title">账号与规则</text>
    <view class="list-card"><view v-for="item in accountActions" :key="item.label" class="row" @tap="openAction(item)"><view class="row-icon pale">{{ item.icon }}</view><view><text class="row-label">{{ item.label }}</text><text class="row-desc">{{ item.desc }}</text></view><text class="arrow">›</text></view></view>
    <text class="footer-note">认真整理每一份信息，也尊重每一次真实选择。</text>
  </view>
</template>
<script>
import { api, clearSession } from '@/api/request.js'
import { syncCustomTabBar } from '@/utils/tab-bar.js'
import { getPrimaryNavLayout } from '@/utils/nav-layout.js'
export default {
  data() { return { navLayout: getPrimaryNavLayout(), profile: { nickname: uni.getStorageSync('nickname') || '', avatar_url: '' }, actions: [
    { label: '我赞过的发现', desc: '回到发现页查看已标记有用的帖子', icon: '👍', action: 'likes' },
    { label: '收藏的店与路线', desc: '继续上次想去的地方', icon: '☆', action: 'favorites' },
    { label: '举报记录', desc: '查看内容举报处理进度', icon: '!', url: '/pages/cases/cases' },
  ], accountActions: [
    { label: '隐私保护政策', desc: '了解信息如何被使用', icon: '隐', url: '/pages/privacy/privacy' },
    { label: '用户服务协议', desc: '平台规则与内容说明', icon: '约', url: '/pages/agreement/agreement' },
    { label: '账号与安全', desc: '资料、退出登录与注销账号', icon: '安', action: 'settings' },
  ] } },
  computed: { isLoggedIn() { return Boolean(uni.getStorageSync('token')) }, avatarText() { return (this.profile.nickname || '饭').slice(0, 1) } },
  onShow() { this.navLayout = getPrimaryNavLayout(); syncCustomTabBar(this, 2); this.loadProfile() },
  methods: {
    async loadProfile() { if (!this.isLoggedIn) return; try { this.profile = await api.getProfile() } catch (error) { console.warn('[mine] profile load failed:', error.message) } },
    goProfile() { uni.navigateTo({ url: this.isLoggedIn ? '/pages/profile/profile' : '/pages/login/login' }) },
    goDiscover() { uni.switchTab({ url: '/pages/discover/discover' }) },
    openAction(item) { if (item.url) return uni.navigateTo({ url: item.url }); if (item.action === 'likes') return this.goDiscover(); if (item.action === 'favorites') return uni.showToast({ title: '收藏列表将在下一版集中展示', icon: 'none' }); if (item.action === 'settings') return this.showSettings() },
    showSettings() { if (!this.isLoggedIn) return uni.navigateTo({ url: '/pages/login/login' }); uni.showActionSheet({ itemList: ['编辑个人资料', '清理本地缓存', '退出登录', '注销账号'], success: ({ tapIndex }) => { if (tapIndex === 0) this.goProfile(); if (tapIndex === 1) uni.showToast({ title: '缓存已整理', icon: 'success' }); if (tapIndex === 2) { clearSession(); uni.reLaunch({ url: '/pages/login/login' }) }; if (tapIndex === 3) this.confirmDeleteAccount() } }) },
    confirmDeleteAccount() { uni.showModal({ title: '永久注销账号？', content: '账号资料、收藏、点赞和其他关联数据将被永久删除。', confirmText: '永久注销', confirmColor: '#C43D2D', success: async ({ confirm }) => { if (!confirm) return; try { await api.deleteAccount(); uni.clearStorageSync(); uni.showToast({ title: '账号已注销', icon: 'success' }); setTimeout(() => uni.reLaunch({ url: '/pages/home/home' }), 500) } catch (error) { uni.showToast({ title: error.message || '注销失败，请稍后重试', icon: 'none' }) } } }) },
  },
}
</script>
<style scoped>
.mine-page{min-height:100vh;padding:0 30rpx calc(150rpx + env(safe-area-inset-bottom));box-sizing:border-box;background:#FFF9EF;color:#332E26}.mine-head{display:flex;align-items:flex-start;gap:20rpx;padding-bottom:32rpx}.mine-head>view:first-child{flex:1;display:flex;flex-direction:column}.eyebrow{color:#D86642;font-size:21rpx;font-weight:850;letter-spacing:3rpx}.title{margin-top:5rpx;font-family:"Kaiti SC",serif;font-size:52rpx;font-weight:900}.subtitle{margin-top:8rpx;color:#827765;font-size:22rpx}.profile{display:flex;flex-direction:column;align-items:center;gap:7rpx;color:#706654;font-size:20rpx}.profile image,.profile>view{width:76rpx;height:76rpx;border:3rpx solid #E5D5B8;border-radius:50%;background:#F5D36C;line-height:76rpx;text-align:center;font-size:30rpx;font-weight:850}.discover-card{position:relative;display:flex;flex-direction:column;margin:8rpx 0 40rpx;padding:34rpx;border:3rpx solid #DED0B6;border-radius:28rpx;background:#FFF1C9;transform:rotate(-.35deg)}.tape{position:absolute;top:-14rpx;left:50%;width:120rpx;height:30rpx;background:rgba(245,190,78,.65);transform:translateX(-50%) rotate(2deg)}.card-kicker{color:#D86642;font-size:21rpx;font-weight:850}.card-title{margin-top:13rpx;font-family:"Kaiti SC",serif;font-size:37rpx;font-weight:900;line-height:1.4}.card-note{margin-top:13rpx;color:#6F624C;font-size:22rpx;line-height:1.65}.card-link{align-self:flex-start;margin-top:22rpx;padding:12rpx 18rpx;border-radius:999rpx;background:#332E26;color:#fff;font-size:22rpx;font-weight:800}.section-title{display:block;margin:30rpx 4rpx 16rpx;font-family:"Kaiti SC",serif;font-size:34rpx;font-weight:900}.list-card{overflow:hidden;border:2rpx solid #E8DCC7;border-radius:25rpx;background:#FFFDF9}.row{min-height:108rpx;display:flex;align-items:center;gap:18rpx;padding:16rpx 22rpx;border-bottom:2rpx dashed #E9DECB}.row:last-child{border:0}.row-icon{width:62rpx;height:62rpx;flex:none;border-radius:19rpx;background:#FFE1D5;line-height:62rpx;text-align:center;font-size:27rpx;font-weight:850;transform:rotate(-2deg)}.row-icon.pale{background:#F3E9D7}.row>view:nth-child(2){min-width:0;flex:1;display:flex;flex-direction:column}.row-label{font-size:26rpx;font-weight:850}.row-desc{margin-top:7rpx;color:#887D6A;font-size:21rpx}.arrow{color:#887D6A;font-size:39rpx}.footer-note{display:block;padding:46rpx 10rpx 20rpx;color:#A09582;text-align:center;font-size:21rpx}
</style>
