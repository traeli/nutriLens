<template>
  <view class="login-page">
    <view class="hero" :style="{ paddingTop: `${navLayout.contentTop}px` }">
      <image class="hero-image" :src="loginBackground" mode="aspectFill" />
      <view class="hero-shade"></view>
      <view class="brand"><view class="brand-dot"></view><text>饭友记</text></view>
      <view class="hero-copy">
        <text class="hero-kicker">城市真实食记</text>
        <text class="hero-title">跟着真实食客，\n吃懂一座城</text>
        <text class="hero-subtitle">发现走得完、吃得好的路线，把真实到店经历留成自己的城市食记。</text>
      </view>
    </view>

    <view class="login-panel">
      <button class="login-button" :loading="loginLoading" :disabled="!privacyAgreed || loginLoading" @tap="doLogin">
        <text>微信登录</text><text class="login-arrow">→</text>
      </button>
      <text v-if="!privacyAgreed" class="unlock-tip">完成协议确认后即可登录</text>
      <view class="legal-note">
        <text>登录即代表你已阅读并同意</text>
        <view class="legal-links">
          <text @tap="openAgreement('privacy')">《隐私保护政策》</text>
          <text @tap="openAgreement('user_service')">《用户服务协议》</text>
          <text @tap="openAgreement('community')">《社区内容规范》</text>
        </view>
      </view>
    </view>

    <view v-if="!privacyAgreed" class="privacy-mask">
      <view class="privacy-dialog" role="dialog" aria-label="用户协议确认">
        <text class="privacy-title">开始使用饭友记</text>
        <text class="privacy-copy">为了提供账号、到店记录和个性化服务，请阅读并一次性确认以下协议。协议版本更新时，我们会再次征求你的同意。</text>
        <view class="privacy-links">
          <view class="privacy-link" @tap="openAgreement('privacy')"><text>隐私保护政策</text><text>›</text></view>
          <view class="privacy-link" @tap="openAgreement('user_service')"><text>用户服务协议</text><text>›</text></view>
          <view class="privacy-link" @tap="openAgreement('community')"><text>社区内容规范</text><text>›</text></view>
        </view>
        <view class="consent-row" @tap="checked = !checked">
          <view class="checkbox" :class="{ checked }"><text v-if="checked">✓</text></view>
          <text class="consent-copy">我已阅读并同意以上三份协议</text>
        </view>
        <view class="privacy-actions">
          <button class="secondary-button" @tap="denyPrivacy">暂不同意</button>
          <button class="primary-button" open-type="agreePrivacyAuthorization" :disabled="!checked" @agreeprivacyauthorization="agreePrivacy">同意并继续</button>
        </view>
      </view>
    </view>
  </view>
</template>

<script>
import { api } from '@/api/request.js'
import { assetUrl } from '@/utils/assets.js'
import { getPrimaryNavLayout } from '@/utils/nav-layout.js'
import { AGREEMENT_TYPES, AGREEMENT_VERSION } from '@/content/agreements.js'

const agreementPages = {
  privacy: '/pages/privacy/privacy',
  user_service: '/pages/agreement/agreement',
  community: '/pages/community/community',
}

export default {
  data() {
    return {
      navLayout: getPrimaryNavLayout(),
      loginBackground: assetUrl('/static/dining/login-cheers-background-v2.jpg'),
      privacyAgreed: false,
      checked: false,
      loginLoading: false,
      inviterId: 0,
    }
  },
  onLoad(options) {
    this.privacyAgreed = uni.getStorageSync('privacy_agreed_version') === AGREEMENT_VERSION
    this.checked = this.privacyAgreed
    this.inviterId = parseInt((options && options.inviter_id) || 0, 10) || 0
    this.checkLogin()
  },
  onShow() { this.navLayout = getPrimaryNavLayout() },
  methods: {
    openAgreement(type) { uni.navigateTo({ url: agreementPages[type] }) },
    agreePrivacy() {
      if (!this.checked) return
      this.privacyAgreed = true
      uni.setStorageSync('privacy_agreed_version', AGREEMENT_VERSION)
    },
    denyPrivacy() { uni.switchTab({ url: '/pages/home/home' }) },
    async checkLogin() {
      if (!uni.getStorageSync('token')) return
      try {
        const profile = await api.getProfile()
        const hasProfile = Boolean(profile && String(profile.nickname || '').trim())
        uni.setStorageSync('has_profile', hasProfile)
        if (hasProfile) uni.setStorageSync('nickname', profile.nickname)
        uni.switchTab({ url: '/pages/home/home' })
      } catch (error) {
        if (error.status !== 401) console.warn('[login] 登录状态校验失败', error.message || error)
      }
    },
    doLogin() {
      if (!this.privacyAgreed || this.loginLoading) return
      this.loginLoading = true
      uni.login({
        provider: 'weixin',
        success: async ({ code }) => {
          try {
            const res = await api.wxLogin(code, this.inviterId)
            uni.setStorageSync('user_id', res.user_id)
            uni.setStorageSync('has_profile', res.has_profile)
            try {
              await Promise.all(AGREEMENT_TYPES.map(agreement_type => api.agreePrivacy({ agreement_type, version: AGREEMENT_VERSION })))
            } catch (error) {
              console.warn('[login] 协议状态同步失败', error.message || error)
            }
            if (!uni.getStorageSync('token')) return
            uni.switchTab({ url: '/pages/home/home' })
          } catch (error) {
            uni.showToast({ title: `登录失败：${error.message}`, icon: 'none' })
          } finally { this.loginLoading = false }
        },
        fail: () => {
          this.loginLoading = false
          uni.showToast({ title: '微信登录失败', icon: 'none' })
        },
      })
    },
  },
}
</script>

<style scoped>
.login-page { min-height: 100vh; box-sizing: border-box; overflow: hidden; background: #fff9ee; color: #332e26; }
.hero { position: relative; height: 62vh; min-height: 760rpx; box-sizing: border-box; overflow: hidden; background: #246647; }
.hero-image, .hero-shade { position: absolute; inset: 0; width: 100%; height: 100%; }
.hero-shade { background: linear-gradient(180deg, rgba(20, 42, 30, .12), rgba(20, 42, 30, .78)); }
.brand { position: relative; z-index: 1; display: flex; align-items: center; gap: 14rpx; padding: 0 32rpx; color: #fff; font-size: 30rpx; font-weight: 800; letter-spacing: 2rpx; }
.brand-dot { width: 18rpx; height: 18rpx; border: 5rpx solid #fff; border-radius: 50%; }
.hero-copy { position: absolute; z-index: 1; left: 32rpx; right: 32rpx; bottom: 64rpx; display: flex; flex-direction: column; }
.hero-kicker { color: #e5f3ea; font-size: 24rpx; font-weight: 700; letter-spacing: 4rpx; }
.hero-title { margin-top: 18rpx; color: #fff; font-size: 52rpx; font-weight: 900; line-height: 1.25; }
.hero-subtitle { max-width: 620rpx; margin-top: 20rpx; color: rgba(255,255,255,.88); font-size: 26rpx; line-height: 1.7; }
.login-panel { padding: 40rpx 32rpx calc(36rpx + env(safe-area-inset-bottom)); }
.login-button { display: flex; align-items: center; justify-content: center; gap: 20rpx; width: 100%; height: 96rpx; margin: 0; border: 0; border-radius: 48rpx; background: #2f7d58; color: #fff; font-size: 30rpx; font-weight: 800; line-height: 96rpx; }
.login-button::after, .primary-button::after, .secondary-button::after { border: 0; }
.login-button[disabled] { background: #c9c3b6; color: #fff; }
.login-arrow { font-size: 34rpx; }
.unlock-tip { display: block; margin-top: 16rpx; color: #8b8271; font-size: 24rpx; text-align: center; }
.legal-note { margin-top: 22rpx; color: #6f675b; font-size: 22rpx; line-height: 1.7; text-align: center; }
.legal-links { display: flex; flex-wrap: wrap; justify-content: center; }
.legal-links text { color: #246647; }
.privacy-mask { position: fixed; z-index: 99; inset: 0; display: flex; align-items: center; justify-content: center; box-sizing: border-box; padding: max(32rpx, env(safe-area-inset-top)) 32rpx max(32rpx, env(safe-area-inset-bottom)); background: rgba(25, 28, 25, .66); }
.privacy-dialog { width: 100%; max-width: 686rpx; max-height: calc(100vh - 64rpx); box-sizing: border-box; overflow-y: auto; padding: 40rpx 32rpx 32rpx; border-radius: 28rpx; background: #fff; }
.privacy-title { display: block; font-size: 34rpx; font-weight: 800; }
.privacy-copy { display: block; margin-top: 18rpx; color: #6f675b; font-size: 26rpx; line-height: 1.7; }
.privacy-links { margin-top: 24rpx; border-top: 1rpx solid #e7dcc5; }
.privacy-link { display: flex; align-items: center; justify-content: space-between; min-height: 88rpx; border-bottom: 1rpx solid #e7dcc5; color: #332e26; font-size: 26rpx; }
.consent-row { display: flex; align-items: flex-start; gap: 16rpx; min-height: 88rpx; box-sizing: border-box; padding-top: 24rpx; }
.checkbox { flex: none; display: flex; align-items: center; justify-content: center; width: 40rpx; height: 40rpx; box-sizing: border-box; border: 3rpx solid #8b8271; border-radius: 10rpx; color: #fff; font-size: 26rpx; line-height: 1; }
.checkbox.checked { border-color: #2f7d58; background: #2f7d58; }
.consent-copy { min-width: 0; flex: 1; color: #332e26; font-size: 26rpx; line-height: 1.55; }
.privacy-actions { display: flex; gap: 16rpx; margin-top: 28rpx; }
.secondary-button, .primary-button { flex: 1; height: 88rpx; margin: 0; padding: 0 12rpx; border-radius: 44rpx; font-size: 26rpx; font-weight: 700; line-height: 88rpx; }
.secondary-button { border: 2rpx solid #e7dcc5; background: #fff; color: #6f675b; }
.primary-button { border: 0; background: #2f7d58; color: #fff; }
.primary-button[disabled] { background: #d4cfc5; color: #fff; }
@media (max-width: 350px) {
  .hero { min-height: 660rpx; }
  .hero-title { font-size: 46rpx; }
  .privacy-dialog { padding: 32rpx 24rpx 24rpx; }
  .privacy-actions { flex-direction: column-reverse; }
  .secondary-button, .primary-button { width: 100%; flex: none; }
}
</style>
