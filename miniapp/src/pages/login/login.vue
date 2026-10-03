<template>
  <view class="login-page">
    <!-- 协议详情视图 -->
    <view class="legal-view" v-if="viewType" :style="{ paddingTop: `${navLayout.contentTop}px` }">
      <view class="legal-header">
        <text class="legal-back" @tap="viewType = ''">‹ 返回</text>
        <text class="legal-title">{{ currentAgreement.title }}</text>
        <text class="legal-date">更新日期：{{ agreementUpdatedAt }}</text>
      </view>
      <scroll-view class="legal-scroll" scroll-y>
        <view v-for="section in currentAgreement.sections" :key="section.title">
          <text class="sec-title">{{ section.title }}</text>
          <text v-for="line in section.lines" :key="line" class="sec-text">{{ line }}</text>
        </view>
      </scroll-view>
    </view>

    <!-- 隐私协议弹窗 -->
    <view class="privacy-modal" v-if="!privacyAgreed && !viewType">
      <view class="privacy-content card">
        <text class="privacy-title">用户隐私保护提示</text>
        <scroll-view class="privacy-text" scroll-y>
          <text user-select>欢迎使用饭友记。在开始记录之前，请阅读并同意相关协议。我们只会在提供账号、到店记录和个性化服务所必需的范围内使用你的信息。</text>
        </scroll-view>
        <view class="privacy-links">
          <text class="link-btn" @tap="viewType = 'privacy'">查看《隐私保护政策》 ›</text>
          <text class="link-btn" @tap="viewType = 'user_service'">查看《用户服务协议》 ›</text>
          <text class="link-btn" @tap="viewType = 'community'">查看《社区内容规范》 ›</text>
        </view>
        <view class="privacy-check">
          <view class="checkbox-wrap" @tap="checked = !checked">
            <view class="checkbox-box" :class="{ checked }">
              <text class="checkbox-icon" v-if="checked">✓</text>
            </view>
            <text class="checkbox-label">我已阅读并同意《隐私保护政策》《用户服务协议》和《社区内容规范》</text>
          </view>
        </view>
        <view class="privacy-buttons">
          <button class="btn-deny" @tap="denyPrivacy">不同意</button>
          <button class="btn-agree" :disabled="!canAgree" @tap="agreePrivacy">
            {{ agreeBtnText }}
          </button>
        </view>
      </view>
    </view>

    <!-- 登录页 -->
    <view class="login-content" v-if="!viewType" :style="{ paddingTop: `${navLayout.contentTop}px` }">
      <image class="login-background" :src="loginBackground" mode="widthFix" />
      <view class="login-tone"></view>
      <view class="login-fade"></view>

      <view class="login-brand">
        <view class="brand-spark"><view></view></view>
        <text>饭友记</text>
      </view>

      <view class="login-panel">
        <view class="login-agreement">
          <view class="checkbox-wrap" @tap="loginChecked = !loginChecked">
            <view class="checkbox-box" :class="{ checked: loginChecked }">
              <text class="checkbox-icon" v-if="loginChecked">✓</text>
            </view>
            <text class="checkbox-label">我已阅读并同意
              <text class="link-inline" @tap.stop="viewType = 'privacy'">《隐私保护政策》</text>和
              <text class="link-inline" @tap.stop="viewType = 'user_service'">《用户服务协议》</text>和
              <text class="link-inline" @tap.stop="viewType = 'community'">《社区内容规范》</text>
            </text>
          </view>
        </view>

        <button class="login-btn" @tap="doLogin" :loading="loginLoading" :disabled="!loginChecked">
          <text>微信登录</text>
          <text class="login-arrow">↗</text>
        </button>
      </view>
    </view>
  </view>
</template>

<script>
import { api } from '@/api/request.js'
import { assetUrl } from '@/utils/assets.js'
import { AGREEMENTS, AGREEMENT_TYPES, AGREEMENT_UPDATED_AT, AGREEMENT_VERSION } from '@/content/agreements.js'

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
      navLayout: getNavLayout(),
      loginBackground: assetUrl('/static/dining/login-cheers-background-v2.jpg'),
      privacyAgreed: false,
      checked: false,
      loginChecked: false,
      countdown: 3,
      timer: null,
      loginLoading: false,
      viewType: '', // 可选值：空、隐私政策、用户协议、社区规范
      inviterId: 0,
    }
  },
  computed: {
    currentAgreement() {
      return AGREEMENTS[this.viewType] || { title: '', sections: [] }
    },
    agreementUpdatedAt() { return AGREEMENT_UPDATED_AT },
    canAgree() {
      return this.checked && this.countdown <= 0
    },
    agreeBtnText() {
      if (this.countdown > 0) return `请阅读(${this.countdown}s)`
      if (!this.checked) return '请勾选同意'
      return '我已阅读并同意'
    },
  },
  onLoad(options) {
    this.privacyAgreed = uni.getStorageSync('privacy_agreed_version') === AGREEMENT_VERSION
    if (options && options.inviter_id) {
      this.inviterId = parseInt(options.inviter_id) || 0
    }
    if (!this.privacyAgreed) {
      this.startCountdown()
    }
    this.checkLogin()
  },
  onUnload() {
    if (this.timer) clearInterval(this.timer)
  },
  onShow() {
    this.navLayout = getNavLayout()
  },
  methods: {
    startCountdown() {
      this.countdown = 3
      this.timer = setInterval(() => {
        this.countdown--
        if (this.countdown <= 0) {
          clearInterval(this.timer)
          this.timer = null
        }
      }, 1000)
    },
    checkLogin() {
      const token = uni.getStorageSync('token')
      if (token) {
        const hasProfile = uni.getStorageSync('has_profile')
        if (!hasProfile) {
          uni.redirectTo({ url: '/pages/profile/profile' })
        } else {
          uni.switchTab({ url: '/pages/home/home' })
        }
      }
    },
    agreePrivacy() {
      if (!this.canAgree) return
      this.privacyAgreed = true
      uni.setStorageSync('privacy_agreed_version', AGREEMENT_VERSION)
    },
    denyPrivacy() {
      uni.showToast({ title: '需要同意隐私协议才能使用', icon: 'none' })
    },
    doLogin() {
      if (this.loginLoading) return
      this.loginLoading = true

      uni.login({
        provider: 'weixin',
        success: async (loginRes) => {
          try {
            const res = await api.wxLogin(loginRes.code, this.inviterId)
            uni.setStorageSync('user_id', res.user_id)
            uni.setStorageSync('has_profile', res.has_profile)

            try {
              await Promise.all(AGREEMENT_TYPES.map(agreement_type => api.agreePrivacy({ agreement_type, version: AGREEMENT_VERSION })))
            } catch (e) {
              console.log('agree privacy api failed', e)
            }

            if (!res.has_profile) {
              uni.redirectTo({ url: '/pages/profile/profile' })
            } else {
              uni.switchTab({ url: '/pages/home/home' })
            }
          } catch (err) {
            uni.showToast({ title: '登录失败: ' + err.message, icon: 'none' })
          } finally {
            this.loginLoading = false
          }
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
.login-page {
  min-height: 100vh;
  background: #F7F5F0;
}

/* 协议详情视图 */
.legal-view {
  min-height: 100vh;
  padding: calc(env(safe-area-inset-top) + 28rpx) 34rpx 40rpx;
  background: #F7F5F0;
}

.legal-header {
  padding-bottom: 24rpx;
  border-bottom: 1rpx solid #DDD8D0;
  margin-bottom: 30rpx;
}

.legal-back {
  font-size: 36rpx;
  color: #9A5A48;
  display: block;
  margin-bottom: 16rpx;
}

.legal-title {
  font-size: 38rpx;
  font-weight: 720;
  color: #25231F;
  display: block;
  margin-bottom: 8rpx;
}

.legal-date {
  font-size: 24rpx;
  color: #928C84;
  display: block;
}

.legal-scroll {
  height: calc(100vh - 200rpx);
}

.sec-title {
  font-size: 30rpx;
  font-weight: 700;
  color: #25231F;
  display: block;
  margin-top: 30rpx;
  margin-bottom: 16rpx;
}

.sec-text {
  font-size: 26rpx;
  color: #69645E;
  line-height: 1.8;
  display: block;
  margin-bottom: 8rpx;
}

/* 隐私弹窗 */
.privacy-modal {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  padding: 34rpx;
  background: rgba(24, 24, 22, 0.62);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 999;
}

.privacy-content {
  width: 100%;
  max-height: 78vh;
  padding: 40rpx 34rpx 32rpx;
  border: 1rpx solid rgba(255,255,255,.6);
  border-radius: 28rpx;
  background: #FFFEFB;
  box-shadow: 0 24rpx 80rpx rgba(22,22,20,.22);
}

.privacy-title {
  font-size: 36rpx;
  font-weight: 720;
  color: #25231F;
  margin-bottom: 20rpx;
  display: block;
}

.privacy-text {
  max-height: 200rpx;
  margin-bottom: 20rpx;
  font-size: 26rpx;
  color: #69645E;
  line-height: 1.8;
}

.privacy-links {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
  margin-bottom: 24rpx;
  padding: 20rpx;
  border: 1rpx solid #E2DDD5;
  border-radius: 16rpx;
  background: #F7F5F0;
}

.link-btn {
  font-size: 28rpx;
  color: #4F6B57;
  font-weight: 650;
}

.privacy-check {
  margin-bottom: 24rpx;
}

.checkbox-wrap {
  display: flex;
  align-items: flex-start;
  gap: 12rpx;
}

.checkbox-box {
  width: 36rpx;
  height: 36rpx;
  border: 2rpx solid #B8B2A9;
  border-radius: 8rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  margin-top: 4rpx;
}

.checkbox-box.checked {
  border-color: #4F6B57;
  background: #4F6B57;
}

.checkbox-icon {
  color: #fff;
  font-size: 24rpx;
}

.checkbox-label {
  font-size: 24rpx;
  color: #69645E;
  line-height: 1.5;
}

.privacy-buttons {
  display: flex;
  gap: 20rpx;
}

.btn-deny {
  flex: 1;
  background: #EFECE6;
  color: #77716A;
  border-radius: 16rpx;
  font-size: 28rpx;
}

.btn-agree {
  flex: 1;
  background: #181816;
  color: #fff;
  border-radius: 16rpx;
  font-size: 28rpx;
}

.btn-agree[disabled] {
  background: #D2CEC7;
  color: #F7F5F0;
}

/* 登录页 */
.login-content {
  min-height: 100vh;
  padding-bottom: calc(env(safe-area-inset-bottom) + 28rpx);
}

.login-hero {
  position: relative;
  height: 650rpx;
  overflow: hidden;
  border-radius: 0 0 42rpx 42rpx;
  background: #EDE1CA;
}

.login-hero image, .hero-shade {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
}

.hero-shade {
  background: linear-gradient(90deg, rgba(244,232,207,.2), transparent 58%), linear-gradient(180deg, transparent 50%, rgba(245,232,207,.28));
}

.cover-binding {
  position: absolute;
  left: 13rpx;
  top: 92rpx;
  bottom: 74rpx;
  display: flex;
  flex-direction: column;
  justify-content: space-around;
}

.cover-binding view {
  width: 22rpx;
  height: 7rpx;
  border-radius: 6rpx;
  background: rgba(58,52,43,.46);
  box-shadow: 10rpx 0 0 rgba(246,236,216,.82);
}

.brand-copy {
  position: absolute;
  left: 52rpx;
  bottom: 72rpx;
  display: flex;
  flex-direction: column;
  color: #2C2924;
}

.brand-kicker {
  color: #9A5A48;
  font-size: 17rpx;
  font-weight: 700;
  letter-spacing: 3rpx;
}

.app-name {
  margin-top: 12rpx;
  font-size: 72rpx;
  font-weight: 760;
  letter-spacing: 6rpx;
  text-shadow: 2rpx 2rpx 0 rgba(154,90,72,.12);
}

.app-slogan {
  margin-top: 8rpx;
  color: #625B51;
  font-size: 25rpx;
  letter-spacing: 2rpx;
}

.hero-stamp {
  position: absolute;
  right: 34rpx;
  bottom: 48rpx;
  width: 104rpx;
  height: 104rpx;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  border: 2rpx solid rgba(154,90,72,.7);
  border-radius: 50%;
  color: #9A5A48;
  font-size: 17rpx;
  line-height: 1.5;
  transform: rotate(8deg);
}

.login-panel {
  position: relative;
  z-index: 2;
  margin: -38rpx 28rpx 0;
  padding: 36rpx 30rpx 28rpx;
  border: 1rpx solid #DCD2C2;
  border-radius: 8rpx 24rpx 24rpx 20rpx;
  background: #FFFDF7;
  box-shadow: 0 18rpx 50rpx rgba(45,42,37,.08);
}

.login-tape {
  position: absolute;
  top: -18rpx;
  left: 50%;
  width: 116rpx;
  height: 37rpx;
  background: rgba(196,167,112,.38);
  transform: translateX(-50%) rotate(-2deg);
}

.welcome-copy {
  display: flex;
  flex-direction: column;
}

.welcome-copy > text:first-child {
  color: #25231F;
  font-size: 34rpx;
  font-weight: 720;
}

.welcome-copy > text:last-child {
  margin-top: 10rpx;
  color: #7D7770;
  font-size: 22rpx;
  line-height: 1.55;
}

.feature-list {
  margin-top: 28rpx;
  border-top: 1rpx dashed #D8CCBA;
  border-bottom: 1rpx dashed #D8CCBA;
}

.feature-list > view {
  min-height: 92rpx;
  display: flex;
  align-items: center;
  gap: 20rpx;
  border-bottom: 1rpx dashed #E2D8C9;
}

.feature-list > view:last-child { border-bottom: 0; }
.feature-list > view > text { color: #9A5A48; font-size: 18rpx; font-weight: 700; letter-spacing: 1rpx; }
.feature-list > view > view { display: flex; flex-direction: column; }
.feature-list > view > view text:first-child { color: #302D29; font-size: 24rpx; font-weight: 650; }
.feature-list > view > view text:last-child { margin-top: 3rpx; color: #918B83; font-size: 18rpx; }

.login-agreement {
  margin: 27rpx 2rpx 22rpx;
}

.login-agreement .checkbox-label {
  flex: 1;
  font-size: 21rpx;
  line-height: 1.6;
}

.link-inline {
  color: #9A5A48;
  font-weight: 650;
}

.login-btn {
  width: 100%;
  height: 94rpx;
  margin: 0;
  padding: 0;
  border-radius: 18rpx;
  background: #181816;
  color: #fff;
  font-size: 29rpx;
  font-weight: 650;
  line-height: 94rpx;
}

.login-btn[disabled] {
  background: #D2CEC7;
  color: #F7F5F0;
}

.safe-note {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 9rpx;
  margin-top: 16rpx;
  color: #8C867E;
  font-size: 18rpx;
}

.safe-note view {
  width: 9rpx;
  height: 9rpx;
  border-radius: 50%;
  background: #4F6B57;
}

.login-links {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 14rpx;
  margin-top: 25rpx;
}

.login-link {
  color: #77716A;
  font-size: 20rpx;
}

.login-sep {
  color: #BBB5AC;
  font-size: 20rpx;
}

/* 统一视觉语言：柔和绿色、黑色与亮青柠色。 */
.login-page {
  min-height: 100vh;
  background: #f5faec;
  color: #171914;
}

.login-content {
  box-sizing: border-box;
  min-height: 100vh;
  padding-right: 24rpx;
  padding-left: 24rpx;
  padding-bottom: calc(env(safe-area-inset-bottom) + 38rpx);
}

.login-hero {
  height: 520rpx;
  border-radius: 42rpx;
  background: #11120f;
  box-shadow: 0 22rpx 50rpx rgba(17, 18, 15, .18);
}

.login-hero image {
  left: auto;
  right: 0;
  width: 58%;
  height: 100%;
  opacity: .88;
  clip-path: polygon(25% 0, 100% 0, 100% 100%, 0 100%, 17% 72%, 8% 43%);
}

.hero-shade {
  background: linear-gradient(90deg, #11120f 0%, #11120f 42%, rgba(17,18,15,.72) 61%, rgba(17,18,15,.18) 100%), linear-gradient(180deg, transparent 55%, rgba(17,18,15,.75));
}

.cover-binding {
  display: none;
}

.brand-copy {
  left: 31rpx;
  bottom: 45rpx;
  z-index: 2;
  width: 61%;
  color: #ffffff;
}

.brand-kicker {
  color: #c7ff35;
  font-size: 15rpx;
  font-weight: 850;
  letter-spacing: 3rpx;
}

.app-name {
  margin-top: 17rpx;
  color: #ffffff;
  font-family: "Kaiti SC", STKaiti, cursive;
  font-size: 78rpx;
  font-weight: 900;
  letter-spacing: -2rpx;
  line-height: 1;
  text-shadow: none;
}

.app-slogan {
  margin-top: 17rpx;
  color: rgba(255, 255, 255, .7);
  font-size: 21rpx;
  font-weight: 650;
  letter-spacing: 0;
}

.hero-stamp {
  right: 25rpx;
  bottom: 24rpx;
  z-index: 3;
  width: 103rpx;
  height: 103rpx;
  border: 0;
  border-radius: 35% 65% 48% 52%;
  background: #c7ff35;
  color: #11120f;
  font-size: 16rpx;
  font-weight: 850;
  transform: rotate(7deg);
}

.hero-mark {
  position: absolute;
  z-index: 3;
  color: #c7ff35;
  text-shadow: 0 0 20rpx rgba(199,255,53,.7);
}

.mark-one { left: 33rpx; top: 31rpx; font-size: 28rpx; }
.mark-two { left: 155rpx; top: 77rpx; font-size: 44rpx; }

.login-panel {
  margin: 16rpx 0 0;
  padding: 31rpx 28rpx 27rpx;
  border: 0;
  border-radius: 38rpx;
  background: #ffffff;
  box-shadow: none;
}

.login-tape {
  display: none;
}

.welcome-copy > text:first-child {
  color: #171914;
  font-family: "Kaiti SC", STKaiti, cursive;
  font-size: 41rpx;
  font-weight: 900;
}

.welcome-copy > text:last-child {
  margin-top: 8rpx;
  color: #747b71;
  font-size: 20rpx;
  line-height: 1.55;
}

.feature-list {
  display: grid;
  gap: 10rpx;
  margin-top: 24rpx;
  border: 0;
}

.feature-list > view {
  min-height: 90rpx;
  gap: 15rpx;
  padding: 14rpx 17rpx;
  border: 0;
  border-radius: 24rpx;
  background: #f0f4eb;
}

.feature-list > view > text {
  width: 43rpx;
  height: 43rpx;
  flex: none;
  border-radius: 50%;
  background: #11120f;
  color: #c7ff35;
  font-size: 15rpx;
  font-weight: 900;
  letter-spacing: 0;
  line-height: 43rpx;
  text-align: center;
}

.feature-list > view > view text:first-child {
  color: #242823;
  font-size: 22rpx;
  font-weight: 850;
}

.feature-list > view > view text:last-child {
  margin-top: 3rpx;
  color: #798076;
  font-size: 16rpx;
}

.login-agreement {
  margin: 24rpx 2rpx 20rpx;
  padding: 18rpx;
  border-radius: 23rpx;
  background: #f5f7f2;
}

.checkbox-box {
  width: 38rpx;
  height: 38rpx;
  border-color: #aeb5aa;
  border-radius: 12rpx;
}

.checkbox-box.checked {
  border-color: #11120f;
  background: #11120f;
}

.checkbox-label,
.login-agreement .checkbox-label {
  color: #656c62;
  font-size: 19rpx;
}

.link-inline,
.link-btn {
  color: #4f8529;
  font-weight: 800;
}

.login-btn {
  height: 98rpx;
  border-radius: 49rpx;
  background: #c7ff35;
  color: #11120f;
  font-size: 26rpx;
  font-weight: 900;
  line-height: 98rpx;
  box-shadow: 0 13rpx 26rpx rgba(73, 103, 20, .15);
}

.login-btn::after,
.btn-deny::after,
.btn-agree::after {
  border: 0;
}

.login-btn[disabled] {
  background: #e0e5dc;
  color: #969d93;
  box-shadow: none;
}

.safe-note {
  color: #7b8278;
}

.safe-note view {
  background: #67ad39;
}

.login-links {
  margin-top: 22rpx;
}

.login-link {
  color: #687064;
  font-size: 18rpx;
}

.privacy-modal {
  box-sizing: border-box;
  padding: 25rpx;
  background: rgba(8, 10, 7, .62);
}

.privacy-content {
  box-sizing: border-box;
  padding: 34rpx 28rpx 27rpx;
  border: 0;
  border-radius: 38rpx;
  background: #f9fcf5;
  box-shadow: 0 28rpx 80rpx rgba(8, 10, 7, .25);
}

.privacy-title {
  color: #171914;
  font-family: "Kaiti SC", STKaiti, cursive;
  font-size: 39rpx;
  font-weight: 900;
}

.privacy-text {
  color: #666e63;
  font-size: 23rpx;
}

.privacy-links {
  gap: 5rpx;
  padding: 13rpx 18rpx;
  border: 0;
  border-radius: 23rpx;
  background: #edf2e8;
}

.link-btn {
  min-height: 58rpx;
  font-size: 23rpx;
  line-height: 58rpx;
}

.privacy-buttons {
  gap: 12rpx;
}

.btn-deny,
.btn-agree {
  height: 86rpx;
  border-radius: 43rpx;
  font-size: 23rpx;
  line-height: 86rpx;
}

.btn-deny {
  background: #e7ece2;
  color: #626a5f;
}

.btn-agree {
  background: #11120f;
  color: #c7ff35;
}

.btn-agree[disabled] {
  background: #dde2d9;
  color: #999f96;
}

.legal-view {
  box-sizing: border-box;
  background: #f5faec;
}

.legal-header {
  border-bottom-color: #dce3d7;
}

.legal-back {
  color: #4f8529;
  font-size: 28rpx;
  font-weight: 800;
}

.legal-title {
  color: #171914;
  font-family: "Kaiti SC", STKaiti, cursive;
  font-size: 44rpx;
  font-weight: 900;
}

.sec-title {
  color: #171914;
  font-family: "Kaiti SC", STKaiti, cursive;
  font-weight: 900;
}

.sec-text {
  color: #60685d;
}

/* 年轻化食物贴纸登录封面。 */
.login-content {
  overflow: hidden;
  padding-right: 28rpx;
  padding-left: 28rpx;
  background:
    radial-gradient(circle at 6% 31%, rgba(255, 112, 68, .16) 0 8rpx, transparent 9rpx),
    radial-gradient(circle at 94% 18%, rgba(70, 146, 255, .18) 0 10rpx, transparent 11rpx),
    #f7faef;
}

.login-brand {
  height: 74rpx;
  display: flex;
  align-items: center;
  gap: 13rpx;
  padding: 0 8rpx;
  color: #11120f;
}

.login-brand text {
  font-family: "Kaiti SC", STKaiti, cursive;
  font-size: 32rpx;
  font-weight: 900;
  letter-spacing: 2rpx;
}

.brand-spark {
  width: 42rpx;
  height: 42rpx;
  border: 4rpx solid #11120f;
  border-radius: 50% 45% 52% 42%;
  background: #ff7648;
  color: #11120f;
  font-size: 22rpx;
  font-weight: 900;
  line-height: 42rpx;
  text-align: center;
  transform: rotate(-8deg);
}

.login-hero {
  height: 670rpx;
  overflow: hidden;
  border: 5rpx solid #11120f;
  border-radius: 52rpx 52rpx 80rpx 48rpx;
  background: #c7ff35;
  box-shadow: 11rpx 13rpx 0 #11120f;
  transform: rotate(-.4deg);
}

.login-hero::before {
  content: '';
  position: absolute;
  right: -75rpx;
  top: -100rpx;
  width: 290rpx;
  height: 290rpx;
  border: 5rpx solid #11120f;
  border-radius: 50%;
  background: #ff7648;
}

.login-hero::after {
  content: '';
  position: absolute;
  left: -33rpx;
  bottom: 81rpx;
  width: 90rpx;
  height: 90rpx;
  border: 5rpx solid #11120f;
  border-radius: 50%;
  background: #61a8ff;
}

.sticker-hero,
.login-hero image.sticker-hero {
  position: absolute;
  z-index: 3;
  right: -47rpx;
  bottom: -52rpx;
  left: auto;
  width: 108%;
  height: 108%;
  opacity: 1;
  clip-path: none;
  transform: rotate(.4deg);
}

.hero-orbit {
  position: absolute;
  z-index: 1;
  border: 4rpx solid #11120f;
  border-radius: 50%;
}

.orbit-one {
  left: 35rpx;
  top: 39rpx;
  width: 40rpx;
  height: 40rpx;
  background: #fff8e8;
}

.orbit-two {
  left: 76rpx;
  top: 84rpx;
  width: 16rpx;
  height: 16rpx;
  background: #ff7648;
}

.hero-doodle {
  position: absolute;
  z-index: 4;
  color: #11120f;
  font-family: "Kaiti SC", STKaiti, cursive;
  font-weight: 900;
}

.doodle-arrow {
  left: 35rpx;
  bottom: 25rpx;
  font-size: 68rpx;
  transform: rotate(-20deg);
}

.doodle-star {
  right: 22rpx;
  top: 40rpx;
  font-size: 44rpx;
  transform: rotate(10deg);
}

.hero-smile {
  position: absolute;
  z-index: 5;
  right: 31rpx;
  bottom: 31rpx;
  width: 83rpx;
  height: 83rpx;
  border: 5rpx solid #11120f;
  border-radius: 31% 69% 52% 48%;
  background: #fff8e8;
  transform: rotate(9deg);
}

.hero-smile::before,
.hero-smile::after {
  content: '';
  position: absolute;
  top: 23rpx;
  width: 8rpx;
  height: 8rpx;
  border-radius: 50%;
  background: #11120f;
}

.hero-smile::before { left: 22rpx; }
.hero-smile::after { right: 22rpx; }

.hero-smile view {
  position: absolute;
  left: 23rpx;
  bottom: 19rpx;
  width: 36rpx;
  height: 18rpx;
  border-bottom: 5rpx solid #11120f;
  border-radius: 0 0 50% 50%;
}

.login-panel {
  margin: 34rpx 6rpx 0;
  padding: 0 4rpx 24rpx;
  border-radius: 0;
  background: transparent;
}

.welcome-copy {
  position: relative;
  display: inline-flex;
  width: auto;
  margin-bottom: 23rpx;
}

.welcome-copy > text:first-child {
  position: relative;
  z-index: 2;
  color: #11120f;
  font-family: "Kaiti SC", STKaiti, cursive;
  font-size: 52rpx;
  font-weight: 900;
  line-height: 1.05;
  transform: rotate(-2deg);
}

.title-underline {
  position: absolute;
  right: -13rpx;
  bottom: -4rpx;
  left: 4rpx;
  height: 15rpx;
  border-radius: 50%;
  background: #ff7648;
  transform: rotate(-2deg);
}

.login-agreement {
  margin: 0 2rpx 19rpx;
  padding: 0;
  background: transparent;
}

.login-agreement .checkbox-wrap {
  align-items: center;
}

.login-agreement .checkbox-label {
  color: #697064;
  font-size: 18rpx;
  line-height: 1.55;
}

.login-agreement .checkbox-box {
  margin-top: 0;
  border: 4rpx solid #11120f;
  background: #ffffff;
}

.login-agreement .checkbox-box.checked {
  border-color: #11120f;
  background: #c7ff35;
}

.login-agreement .checkbox-icon {
  color: #11120f;
  font-weight: 900;
}

.login-agreement .link-inline {
  color: #3f711f;
}

.login-btn {
  box-sizing: border-box;
  height: 100rpx;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 20rpx 0 37rpx;
  border: 5rpx solid #11120f;
  border-radius: 50rpx;
  background: #11120f;
  color: #ffffff;
  font-size: 27rpx;
  line-height: 1;
  box-shadow: none;
}

.login-btn[disabled] {
  border-color: #cbd0c5;
  background: #e0e4dc;
  color: #92988e;
}

.login-arrow {
  width: 61rpx;
  height: 61rpx;
  border-radius: 50%;
  background: #c7ff35;
  color: #11120f;
  font-size: 37rpx;
  line-height: 57rpx;
  text-align: center;
}

.login-btn[disabled] .login-arrow {
  background: #f0f2ed;
  color: #9ba096;
}

.safe-note {
  margin-top: 13rpx;
  color: #737a70;
  font-size: 17rpx;
}

.safe-note view {
  background: #ff7648;
}

.login-links {
  margin-top: 17rpx;
}

/* 登录控件放置在全屏插画上方较安静的视觉区域。 */
.login-brand {
  align-self: center;
  height: 76rpx;
  gap: 14rpx;
}

.login-brand > text {
  font-size: 46rpx;
  letter-spacing: 3rpx;
}

.brand-spark {
  width: 45rpx;
  height: 45rpx;
}

.login-panel {
  width: 100%;
  margin: 76rpx 0 0;
  padding: 0 5rpx;
}

.login-agreement {
  margin: 0 6rpx 18rpx;
}

.login-agreement .checkbox-label {
  color: #24291f;
  font-size: 18rpx;
  font-weight: 650;
}

.login-agreement .link-inline {
  color: #11120f;
  font-weight: 900;
}

.login-btn {
  height: 96rpx;
  border: 4rpx solid #11120f;
  background: #11120f;
  box-shadow: 0 10rpx 0 rgba(39, 69, 43, .22);
}

.login-btn[disabled] {
  border-color: rgba(17, 18, 15, .42);
  background: rgba(245, 250, 236, .72);
  color: rgba(17, 18, 15, .55);
  box-shadow: none;
}

/* 低饱和插画搭配常规底部操作区域。 */
.login-background {
  opacity: .82;
}

.login-tone,
.login-fade {
  position: absolute;
  z-index: 1;
  top: 0;
  right: 0;
  bottom: 0;
  left: 0;
  pointer-events: none;
}

.login-tone {
  background: rgba(245, 250, 236, .08);
}

.login-fade {
  background: linear-gradient(
    180deg,
    rgba(245, 250, 236, 0) 0%,
    rgba(245, 250, 236, 0) 48%,
    rgba(245, 250, 236, .16) 59%,
    rgba(245, 250, 236, .72) 72%,
    rgba(245, 250, 236, .96) 82%,
    #f5faec 90%,
    #f5faec 100%
  );
}

.login-brand {
  z-index: 3;
}

.login-panel {
  z-index: 3;
  margin: auto 0 0;
  padding: 0 5rpx 4rpx;
}

.login-agreement {
  margin-bottom: 17rpx;
}

.login-btn {
  box-shadow: none;
}

.login-link,
.login-sep {
  font-size: 17rpx;
}

/* 参考作品集式引导结构，并适配当前产品视觉体系。 */
.login-content {
  overflow: hidden;
  padding-right: 28rpx;
  padding-left: 28rpx;
  background: #f5faec;
}

.login-brand {
  box-sizing: border-box;
  height: 82rpx;
  gap: 13rpx;
  padding: 0 4rpx;
}

.login-brand > text {
  color: #11120f;
  font-family: "Kaiti SC", STKaiti, cursive;
  font-size: 33rpx;
  font-weight: 900;
  letter-spacing: 1rpx;
}

.brand-spark {
  box-sizing: border-box;
  width: 43rpx;
  height: 43rpx;
  border: 0;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #11120f;
  transform: none;
}

.brand-spark view {
  width: 14rpx;
  height: 14rpx;
  border-radius: 50%;
  background: #c7ff35;
}

.brand-edition {
  height: 38rpx;
  margin-left: auto;
  padding: 0 17rpx;
  border: 2rpx solid #11120f;
  border-radius: 20rpx;
  color: #11120f;
  font-size: 13rpx;
  font-weight: 900;
  letter-spacing: 2rpx;
  line-height: 38rpx;
}

.login-hero {
  box-sizing: border-box;
  height: 790rpx;
  overflow: hidden;
  border: 0;
  border-radius: 58rpx;
  background: #11120f;
  box-shadow: none;
  transform: none;
}

.login-hero::before,
.login-hero::after {
  display: none;
  content: none;
}

.hero-meta {
  position: absolute;
  z-index: 8;
  top: 31rpx;
  right: 31rpx;
  left: 31rpx;
  display: flex;
  align-items: center;
  gap: 13rpx;
  color: rgba(255, 255, 255, .66);
  font-size: 14rpx;
  font-weight: 800;
  letter-spacing: 1rpx;
}

.hero-meta > text:first-child {
  color: #c7ff35;
}

.hero-meta view {
  width: 35rpx;
  height: 2rpx;
  background: rgba(255, 255, 255, .32);
}

.food-scene {
  position: absolute;
  top: 82rpx;
  right: 0;
  left: 0;
  height: 515rpx;
}

.city-capsule {
  position: absolute;
  left: 31rpx;
  top: 26rpx;
  width: 302rpx;
  height: 428rpx;
  overflow: hidden;
  border-radius: 151rpx;
  background: #f5faec;
}

.city-capsule::before {
  content: '';
  position: absolute;
  right: -33rpx;
  bottom: -66rpx;
  width: 245rpx;
  height: 245rpx;
  border-radius: 50%;
  background: #27452b;
}

.city-block {
  position: absolute;
  z-index: 2;
  bottom: 0;
  background: #11120f;
}

.city-block::before,
.city-block::after {
  content: '';
  position: absolute;
  left: 17rpx;
  width: 8rpx;
  height: 17rpx;
  border-radius: 5rpx;
  background: #c7ff35;
}

.city-block::before { top: 29rpx; }
.city-block::after { top: 57rpx; }
.city-block-one { left: 44rpx; width: 50rpx; height: 176rpx; }
.city-block-two { left: 111rpx; width: 72rpx; height: 259rpx; }
.city-block-three { left: 201rpx; width: 57rpx; height: 141rpx; }

.lime-capsule {
  position: absolute;
  right: 23rpx;
  top: 91rpx;
  width: 304rpx;
  height: 390rpx;
  border-radius: 152rpx;
  background: #c7ff35;
}

.plate-ring {
  position: absolute;
  z-index: 4;
  border: 10rpx solid #f5faec;
  border-radius: 50%;
}

.ring-one { left: 202rpx; top: 240rpx; width: 232rpx; height: 232rpx; }
.ring-two { left: 238rpx; top: 276rpx; width: 160rpx; height: 160rpx; }
.ring-three { left: 274rpx; top: 312rpx; width: 88rpx; height: 88rpx; background: #ff7048; }

.food-dot {
  position: absolute;
  z-index: 5;
  border-radius: 50%;
}

.food-dot-orange {
  left: 79rpx;
  top: 316rpx;
  width: 133rpx;
  height: 133rpx;
  background: #ff7048;
}

.food-dot-blue {
  right: 60rpx;
  top: 355rpx;
  width: 94rpx;
  height: 94rpx;
  background: #61a8ff;
}

.food-leaf {
  position: absolute;
  z-index: 6;
  width: 64rpx;
  height: 112rpx;
  border-radius: 100% 0 100% 0;
  background: #27452b;
}

.leaf-one { right: 163rpx; top: 311rpx; transform: rotate(18deg); }
.leaf-two { right: 118rpx; top: 319rpx; transform: rotate(73deg); }

.route-stroke {
  position: absolute;
  z-index: 3;
  left: 153rpx;
  top: 82rpx;
  width: 353rpx;
  height: 270rpx;
  border: 7rpx solid #f5faec;
  border-top-color: transparent;
  border-right-color: transparent;
  border-radius: 45% 42% 58% 39%;
  transform: rotate(-19deg);
}

.location-pin {
  position: absolute;
  z-index: 8;
  right: 67rpx;
  top: 7rpx;
  width: 72rpx;
  height: 88rpx;
  border-radius: 50% 50% 50% 8rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #ff7048;
  transform: rotate(45deg);
}

.location-pin view {
  width: 22rpx;
  height: 22rpx;
  border-radius: 50%;
  background: #f5faec;
}

.fork-mark {
  position: absolute;
  z-index: 7;
  right: 89rpx;
  top: 172rpx;
  width: 45rpx;
  height: 191rpx;
  border-radius: 23rpx;
  background: #11120f;
  transform: rotate(31deg);
}

.fork-mark::after {
  content: '';
  position: absolute;
  left: -18rpx;
  top: 58rpx;
  width: 81rpx;
  height: 22rpx;
  border-radius: 12rpx;
  background: #11120f;
}

.fork-mark view {
  position: absolute;
  top: 29rpx;
  width: 8rpx;
  height: 52rpx;
  border-radius: 5rpx;
  background: #f5faec;
}

.fork-mark view:nth-child(1) { left: 7rpx; }
.fork-mark view:nth-child(2) { left: 19rpx; }
.fork-mark view:nth-child(3) { left: 31rpx; }

.hero-bottom {
  position: absolute;
  right: 33rpx;
  bottom: 36rpx;
  left: 34rpx;
  display: flex;
  align-items: flex-end;
  justify-content: space-between;
}

.hero-title {
  display: flex;
  flex-direction: column;
  color: #ffffff;
  font-family: "Kaiti SC", STKaiti, cursive;
  font-size: 66rpx;
  font-weight: 900;
  letter-spacing: 3rpx;
  line-height: .93;
}

.hero-title text:last-child {
  color: #c7ff35;
}

.hero-arrow {
  width: 80rpx;
  height: 80rpx;
  border-radius: 50%;
  background: #c7ff35;
  color: #11120f;
  font-size: 39rpx;
  font-weight: 900;
  line-height: 76rpx;
  text-align: center;
}

.login-panel {
  margin: 28rpx 4rpx 0;
  padding: 0 0 25rpx;
}

.login-agreement {
  margin: 0 2rpx 19rpx;
}

.login-btn {
  height: 101rpx;
  padding-right: 18rpx;
  border: 0;
  border-radius: 52rpx;
  background: #c7ff35;
  color: #11120f;
}

.login-btn[disabled] {
  border: 0;
}

.login-arrow {
  background: #11120f;
  color: #c7ff35;
}

.login-btn[disabled] .login-arrow {
  background: #aeb5aa;
  color: #edf0e9;
}

.safe-note view {
  background: #27452b;
}

/* 满版共享餐桌登录插画。 */
.login-page {
  min-height: 100vh;
  background: #f5faec;
}

.login-content {
  position: relative;
  box-sizing: border-box;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  padding-right: 28rpx;
  padding-left: 28rpx;
  padding-bottom: calc(env(safe-area-inset-bottom) + 25rpx);
  background: #f5faec;
}

.login-background {
  position: absolute;
  z-index: 0;
  top: 0;
  right: 0;
  left: 0;
  width: 100%;
  height: auto;
}

.login-brand {
  position: relative;
  z-index: 2;
  box-sizing: border-box;
  height: 70rpx;
  align-self: flex-start;
  gap: 12rpx;
  padding: 0 3rpx;
}

.login-brand > text {
  color: #11120f;
  font-family: "Kaiti SC", STKaiti, cursive;
  font-size: 34rpx;
  font-weight: 900;
  letter-spacing: 1rpx;
}

.brand-spark {
  width: 42rpx;
  height: 42rpx;
  background: #11120f;
}

.brand-spark view {
  width: 14rpx;
  height: 14rpx;
  background: #c7ff35;
}

.login-panel {
  position: relative;
  z-index: 2;
  width: 100%;
  margin: auto 0 0;
  padding: 0 4rpx;
  border: 0;
  border-radius: 0;
  background: transparent;
  box-shadow: none;
}

.login-agreement {
  margin: 0 2rpx 18rpx;
  padding: 0;
  background: transparent;
}

.login-agreement .checkbox-wrap {
  align-items: center;
}

.login-agreement .checkbox-box {
  margin-top: 0;
  border: 3rpx solid #11120f;
  background: rgba(255, 255, 255, .78);
}

.login-agreement .checkbox-box.checked {
  border-color: #11120f;
  background: #c7ff35;
}

.login-agreement .checkbox-label {
  color: #596055;
  font-size: 18rpx;
}

.login-btn {
  height: 102rpx;
  padding: 0 19rpx 0 38rpx;
  border: 0;
  border-radius: 52rpx;
  background: #11120f;
  color: #ffffff;
  box-shadow: none;
}

.login-btn[disabled] {
  border: 0;
  background: #dfe4da;
  color: #949b91;
}

.login-arrow {
  background: #c7ff35;
  color: #11120f;
}

.login-btn[disabled] .login-arrow {
  background: #f1f3ee;
  color: #a0a69c;
}

.safe-note {
  margin-top: 14rpx;
  color: #737a70;
  font-size: 17rpx;
}

.safe-note view {
  background: #ff7048;
}

.login-links {
  margin-top: 17rpx;
}
</style>
