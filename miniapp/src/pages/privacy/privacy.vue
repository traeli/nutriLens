<template>
  <view class="legal-page">
    <view class="legal-header">
      <text class="legal-title">隐私保护政策</text>
      <text class="legal-date">更新日期：2026年5月8日</text>
    </view>
    <view class="legal-body">
      <text class="section-title">一、我们收集的信息</text>
      <text class="section-text">为了向您提供营养分析服务，我们会收集以下信息：</text>
      <text class="section-text">1. 注册信息：微信授权获取的 OpenID 用于身份识别</text>
      <text class="section-text">2. 个人健康信息：身高、体重、年龄、性别，用于提供个性化营养建议</text>
      <text class="section-text">3. 饮食记录数据：您录入或拍照识别的食物信息、卡路里、营养成分数据</text>
      <text class="section-text">4. 图片信息：您上传的食物照片，仅用于食物识别分析</text>

      <text class="section-title">二、信息使用目的</text>
      <text class="section-text">我们收集的信息将用于：</text>
      <text class="section-text">1. 食物识别和营养成分分析</text>
      <text class="section-text">2. 个性化饮食建议和健康管理</text>
      <text class="section-text">3. 饮食记录的统计和趋势展示</text>
      <text class="section-text">4. 产品功能优化和服务质量提升</text>

      <text class="section-title">三、信息存储与安全</text>
      <text class="section-text">1. 您的信息存储在安全的服务器上，采用加密传输（HTTPS）保护数据安全</text>
      <text class="section-text">2. 我们使用行业标准的数据库加密技术保护您的个人数据</text>
      <text class="section-text">3. 严格限制数据访问权限，仅授权人员可访问用户数据</text>
      <text class="section-text">4. 定期进行安全审计和漏洞检查</text>

      <text class="section-title">四、信息共享</text>
      <text class="section-text">1. 我们不会将您的个人信息出售或分享给任何第三方</text>
      <text class="section-text">2. 食物图片仅发送至AI模型进行分析，不会被保存或用于其他用途</text>
      <text class="section-text">3. 不会在您未明确同意的情况下向第三方共享您的饮食数据</text>

      <text class="section-title">五、您的权利</text>
      <text class="section-text">1. 您有权随时查看和修改您的个人资料信息</text>
      <text class="section-text">2. 您有权删除您的饮食记录数据</text>
      <text class="section-text">3. 您有权注销账户并删除所有相关数据</text>
      <text class="section-text">4. 您有权撤回已授予的信息使用同意</text>

      <text class="section-title">六、未成年人保护</text>
      <text class="section-text">我们高度重视对未成年人个人信息的保护。若您是未满14周岁的未成年人，请在监护人指导下使用本产品，并在取得监护人同意后再提交个人信息。</text>

      <text class="section-title">七、政策更新</text>
      <text class="section-text">我们可能会不时更新本隐私政策。更新后的政策将在本页面发布，并通过应用内通知等方式告知您。继续使用本产品即表示您同意受修订后的隐私政策约束。</text>

      <text class="section-title">八、联系我们</text>
      <text class="section-text">如您对本隐私政策有任何疑问或建议，请通过以下方式联系我们：</text>
      <text class="section-text">邮箱：trae.li@outlook.com</text>
    </view>
  </view>
</template>

<style scoped>
.legal-page {
  padding: 30rpx;
  min-height: 100vh;
  background: #fff;
}

.legal-header {
  text-align: center;
  margin-bottom: 40rpx;
  padding-bottom: 30rpx;
  border-bottom: 1rpx solid #f0f0f0;
}

.legal-title {
  font-size: 38rpx;
  font-weight: bold;
  color: #333;
  display: block;
  margin-bottom: 12rpx;
}

.legal-date {
  font-size: 24rpx;
  color: #999;
}

.section-title {
  font-size: 30rpx;
  font-weight: bold;
  color: #333;
  display: block;
  margin-top: 30rpx;
  margin-bottom: 16rpx;
}

.section-text {
  font-size: 26rpx;
  color: #666;
  line-height: 1.8;
  display: block;
  margin-bottom: 8rpx;
}
</style>
