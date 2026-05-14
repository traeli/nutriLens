<template>
  <view class="login-page">
    <!-- 协议详情视图 -->
    <view class="legal-view" v-if="viewType">
      <view class="legal-header">
        <text class="legal-back" @tap="viewType = ''">‹ 返回</text>
        <text class="legal-title">{{ viewType === 'privacy' ? '隐私保护政策' : '用户服务协议' }}</text>
        <text class="legal-date">更新日期：2026年5月8日</text>
      </view>
      <scroll-view class="legal-scroll" scroll-y>
        <!-- 隐私政策内容 -->
        <view v-if="viewType === 'privacy'">
          <text class="sec-title">一、我们收集的信息</text>
          <text class="sec-text">为了向您提供营养分析服务，我们会收集以下信息：</text>
          <text class="sec-text">1. 注册信息：微信授权获取的 OpenID 用于身份识别</text>
          <text class="sec-text">2. 个人健康信息：身高、体重、年龄、性别，用于提供个性化营养建议</text>
          <text class="sec-text">3. 饮食记录数据：您录入或拍照识别的食物信息、卡路里、营养成分数据</text>
          <text class="sec-text">4. 图片信息：您上传的食物照片，仅用于食物识别分析</text>

          <text class="sec-title">二、信息使用目的</text>
          <text class="sec-text">1. 食物识别和营养成分分析</text>
          <text class="sec-text">2. 个性化饮食建议和健康管理</text>
          <text class="sec-text">3. 饮食记录的统计和趋势展示</text>
          <text class="sec-text">4. 产品功能优化和服务质量提升</text>

          <text class="sec-title">三、信息存储与安全</text>
          <text class="sec-text">1. 您的信息存储在安全的服务器上，采用加密传输（HTTPS）保护数据安全</text>
          <text class="sec-text">2. 我们使用行业标准的数据库加密技术保护您的个人数据</text>
          <text class="sec-text">3. 严格限制数据访问权限，仅授权人员可访问用户数据</text>
          <text class="sec-text">4. 定期进行安全审计和漏洞检查</text>

          <text class="sec-title">四、信息共享</text>
          <text class="sec-text">1. 我们不会将您的个人信息出售或分享给任何第三方</text>
          <text class="sec-text">2. 食物图片仅发送至AI模型进行分析，不会被保存或用于其他用途</text>
          <text class="sec-text">3. 不会在您未明确同意的情况下向第三方共享您的饮食数据</text>

          <text class="sec-title">五、您的权利</text>
          <text class="sec-text">1. 您有权随时查看和修改您的个人资料信息</text>
          <text class="sec-text">2. 您有权删除您的饮食记录数据</text>
          <text class="sec-text">3. 您有权注销账户并删除所有相关数据</text>
          <text class="sec-text">4. 您有权撤回已授予的信息使用同意</text>

          <text class="sec-title">六、未成年人保护</text>
          <text class="sec-text">我们高度重视对未成年人个人信息的保护。若您是未满14周岁的未成年人，请在监护人指导下使用本产品，并在取得监护人同意后再提交个人信息。</text>

          <text class="sec-title">七、政策更新</text>
          <text class="sec-text">我们可能会不时更新本隐私政策。更新后的政策将在本页面发布，并通过应用内通知等方式告知您。</text>

          <text class="sec-title">八、联系我们</text>
          <text class="sec-text">如您对本隐私政策有任何疑问或建议，请通过以下方式联系我们：</text>
          <text class="sec-text">邮箱：trae.li@outlook.com</text>
        </view>

        <!-- 用户协议内容 -->
        <view v-if="viewType === 'agreement'">
          <text class="sec-title">一、服务说明</text>
          <text class="sec-text">NutriLens 营养镜头（以下简称"本产品"）是一款基于人工智能技术的智能饮食分析助手，为用户提供食物识别、营养成分分析、饮食记录管理及个性化健康建议等服务。</text>

          <text class="sec-title">二、用户注册与账号</text>
          <text class="sec-text">1. 您通过微信授权登录本产品，即视为注册成为用户</text>
          <text class="sec-text">2. 您应确保提供的个人健康信息（身高、体重、年龄、性别）真实准确</text>
          <text class="sec-text">3. 您有责任保管好自己的账号，因账号保管不当造成的损失由您自行承担</text>

          <text class="sec-title">三、服务内容</text>
          <text class="sec-text">1. 拍照识别：通过AI技术识别食物并分析营养成分</text>
          <text class="sec-text">2. 文字记录：通过文字描述分析食物营养信息</text>
          <text class="sec-text">3. 饮食记录：记录和管理每日饮食数据</text>
          <text class="sec-text">4. 营养建议：基于您的身体信息和饮食数据提供个性化建议</text>
          <text class="sec-text">5. 今天吃什么：随机推荐菜品，帮助解决选择困难</text>

          <text class="sec-title">四、免责声明</text>
          <text class="sec-text">1. 本产品的营养分析结果由AI模型生成，仅供参考，不构成医学或营养专业建议</text>
          <text class="sec-text">2. 食物识别结果可能存在误差，具体营养数据请以实际食品标注为准</text>
          <text class="sec-text">3. 如您患有糖尿病、食物过敏等疾病，请在专业医生指导下进行饮食管理</text>
          <text class="sec-text">4. 本产品不对因使用营养分析数据导致的任何健康问题承担责任</text>

          <text class="sec-title">五、用户行为规范</text>
          <text class="sec-text">1. 不得上传违法、违规或侵犯他人权益的内容</text>
          <text class="sec-text">2. 不得利用本产品从事任何违法违规活动</text>
          <text class="sec-text">3. 不得尝试破坏或干扰本产品的正常运行</text>

          <text class="sec-title">六、知识产权</text>
          <text class="sec-text">本产品的所有内容（包括但不限于界面设计、图标、文字、软件代码）均受知识产权法保护，未经授权不得复制、传播或修改。</text>

          <text class="sec-title">七、服务变更与终止</text>
          <text class="sec-text">1. 我们有权根据业务需要调整或变更服务内容</text>
          <text class="sec-text">2. 如您违反本协议，我们有权暂停或终止向您提供服务</text>
          <text class="sec-text">3. 您有权随时停止使用本产品并注销账号</text>

          <text class="sec-title">八、协议修改</text>
          <text class="sec-text">我们可能会不时修改本协议。修改后的协议将在应用内公布，继续使用本产品即表示您同意受修改后的协议约束。</text>

          <text class="sec-title">九、联系我们</text>
          <text class="sec-text">如您对本协议有任何疑问，请联系：</text>
          <text class="sec-text">邮箱：trae.li@outlook.com</text>
        </view>
      </scroll-view>
    </view>

    <!-- 隐私协议弹窗 -->
    <view class="privacy-modal" v-if="!privacyAgreed && !viewType">
      <view class="privacy-content card">
        <text class="privacy-title">用户隐私保护提示</text>
        <scroll-view class="privacy-text" scroll-y>
          <text user-select>欢迎使用 NutriLens 营养镜头！在使用我们的服务之前，请您仔细阅读并同意以下协议。我们将严格保护您的个人信息安全，仅收集必要信息用于提供个性化营养建议。</text>
        </scroll-view>
        <view class="privacy-links">
          <text class="link-btn" @tap="viewType = 'privacy'">查看《隐私保护政策》 ›</text>
          <text class="link-btn" @tap="viewType = 'agreement'">查看《用户服务协议》 ›</text>
        </view>
        <view class="privacy-check">
          <view class="checkbox-wrap" @tap="checked = !checked">
            <view class="checkbox-box" :class="{ checked }">
              <text class="checkbox-icon" v-if="checked">✓</text>
            </view>
            <text class="checkbox-label">我已阅读并同意《隐私保护政策》和《用户服务协议》</text>
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
    <view class="login-content" v-if="privacyAgreed && !viewType">
      <view class="logo-section">
        <text class="app-name">NutriLens</text>
        <text class="app-slogan">AI 智能饮食分析助手</text>
      </view>

      <view class="features">
        <view class="feature-item">
          <text class="feature-icon">📸</text>
          <text class="feature-text">拍照识别卡路里</text>
        </view>
        <view class="feature-item">
          <text class="feature-icon">🎯</text>
          <text class="feature-text">AI个性化建议</text>
        </view>
        <view class="feature-item">
          <text class="feature-icon">🎡</text>
          <text class="feature-text">今天吃什么转盘</text>
        </view>
      </view>

      <view class="login-agreement">
        <view class="checkbox-wrap" @tap="loginChecked = !loginChecked">
          <view class="checkbox-box" :class="{ checked: loginChecked }">
            <text class="checkbox-icon" v-if="loginChecked">✓</text>
          </view>
          <text class="checkbox-label">登录即表示同意
            <text class="link-inline" @tap.stop="viewType = 'privacy'">《隐私保护政策》</text>和
            <text class="link-inline" @tap.stop="viewType = 'agreement'">《用户服务协议》</text>
          </text>
        </view>
      </view>

      <button class="login-btn" @tap="doLogin" :loading="loginLoading" :disabled="!loginChecked">
        微信一键登录
      </button>

      <view class="login-links">
        <text class="login-link" @tap="viewType = 'privacy'">隐私保护政策</text>
        <text class="login-sep">|</text>
        <text class="login-link" @tap="viewType = 'agreement'">用户服务协议</text>
      </view>
    </view>
  </view>
</template>

<script>
import { api } from '@/api/request.js'

export default {
  data() {
    return {
      privacyAgreed: false,
      checked: false,
      loginChecked: false,
      countdown: 3,
      timer: null,
      loginLoading: false,
      viewType: '', // '', 'privacy', 'agreement'
      inviterId: 0,
    }
  },
  computed: {
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
    this.privacyAgreed = uni.getStorageSync('privacy_agreed')
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
      uni.setStorageSync('privacy_agreed', true)
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
            uni.setStorageSync('token', res.token)
            uni.setStorageSync('user_id', res.user_id)
            uni.setStorageSync('has_profile', res.has_profile)

            try {
              await api.agreePrivacy({ agreement_type: 'privacy', version: '1.0' })
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
  background: linear-gradient(180deg, #e8f5e9 0%, #ffffff 50%);
}

/* 协议详情视图 */
.legal-view {
  min-height: 100vh;
  background: #fff;
  padding: 30rpx;
}

.legal-header {
  padding-bottom: 24rpx;
  border-bottom: 1rpx solid #f0f0f0;
  margin-bottom: 30rpx;
}

.legal-back {
  font-size: 36rpx;
  color: #4CAF50;
  display: block;
  margin-bottom: 16rpx;
}

.legal-title {
  font-size: 38rpx;
  font-weight: bold;
  color: #333;
  display: block;
  margin-bottom: 8rpx;
}

.legal-date {
  font-size: 24rpx;
  color: #999;
  display: block;
}

.legal-scroll {
  height: calc(100vh - 200rpx);
}

.sec-title {
  font-size: 30rpx;
  font-weight: bold;
  color: #333;
  display: block;
  margin-top: 30rpx;
  margin-bottom: 16rpx;
}

.sec-text {
  font-size: 26rpx;
  color: #666;
  line-height: 1.8;
  display: block;
  margin-bottom: 8rpx;
}

/* 隐私弹窗 */
.privacy-modal {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 999;
}

.privacy-content {
  width: 88%;
  max-height: 78vh;
  padding: 40rpx;
}

.privacy-title {
  font-size: 36rpx;
  font-weight: bold;
  text-align: center;
  margin-bottom: 20rpx;
  display: block;
}

.privacy-text {
  max-height: 200rpx;
  margin-bottom: 20rpx;
  font-size: 26rpx;
  color: #666;
  line-height: 1.8;
}

.privacy-links {
  display: flex;
  flex-direction: column;
  gap: 16rpx;
  margin-bottom: 24rpx;
  padding: 20rpx;
  background: #f9f9f9;
  border-radius: 12rpx;
}

.link-btn {
  font-size: 28rpx;
  color: #4CAF50;
  font-weight: bold;
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
  border: 2rpx solid #ccc;
  border-radius: 6rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  margin-top: 4rpx;
}

.checkbox-box.checked {
  background: #4CAF50;
  border-color: #4CAF50;
}

.checkbox-icon {
  color: #fff;
  font-size: 24rpx;
}

.checkbox-label {
  font-size: 24rpx;
  color: #666;
  line-height: 1.5;
}

.privacy-buttons {
  display: flex;
  gap: 20rpx;
}

.btn-deny {
  flex: 1;
  background: #f5f5f5;
  color: #999;
  border: none;
  border-radius: 50rpx;
  font-size: 28rpx;
}

.btn-agree {
  flex: 1;
  background: linear-gradient(135deg, #4CAF50, #45a049);
  color: #fff;
  border: none;
  border-radius: 50rpx;
  font-size: 28rpx;
}

.btn-agree[disabled] {
  background: #ccc;
  color: #fff;
}

/* 登录页 */
.login-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 120rpx 60rpx;
}

.logo-section {
  text-align: center;
  margin-bottom: 80rpx;
}

.app-name {
  font-size: 64rpx;
  font-weight: bold;
  color: #4CAF50;
  display: block;
  margin-bottom: 16rpx;
}

.app-slogan {
  font-size: 28rpx;
  color: #999;
}

.features {
  display: flex;
  gap: 40rpx;
  margin-bottom: 100rpx;
}

.feature-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12rpx;
}

.feature-icon {
  font-size: 48rpx;
}

.feature-text {
  font-size: 24rpx;
  color: #666;
}

.login-agreement {
  width: 100%;
  margin-bottom: 30rpx;
}

.link-inline {
  color: #4CAF50;
  font-weight: bold;
}

.login-btn {
  width: 100%;
  background: linear-gradient(135deg, #4CAF50, #45a049);
  color: #fff;
  border: none;
  border-radius: 50rpx;
  padding: 28rpx 0;
  font-size: 32rpx;
  margin-bottom: 30rpx;
}

.login-btn::after {
  border: none;
}

.login-btn[disabled] {
  background: #ccc;
  color: #fff;
}

.login-links {
  display: flex;
  align-items: center;
  gap: 16rpx;
}

.login-link {
  font-size: 24rpx;
  color: #4CAF50;
}

.login-sep {
  font-size: 24rpx;
  color: #ddd;
}
</style>
