<template>
  <view class="legal-page">
    <view class="legal-header">
      <text class="legal-title">隐私保护政策</text>
      <text class="legal-date">更新日期：2026年9月18日</text>
    </view>
    <view class="legal-body">
      <text class="section-title">一、我们收集的信息</text>
      <text class="section-text">为了提供城市餐饮记录和私人营养记录服务，我们会在必要范围内处理以下信息：</text>
      <text class="section-text">1. 注册信息：微信授权获取的 OpenID 用于身份识别</text>
      <text class="section-text">2. 昵称和头像：用于个人资料及公开记录的作者展示</text>
      <text class="section-text">3. 经微信授权的手机号：仅用于公开发布者身份核验，不向其他用户公开</text>
      <text class="section-text">4. 您主动选择的餐厅位置、记录文字和图片：用于保存私人记录或在审核通过后公开展示</text>
      <text class="section-text">5. 消费凭证：仅供审核核验，默认不公开，并按约定期限删除</text>
      <text class="section-text">6. 私人营养记录：仅本人可见，不用于形成餐厅评价</text>

      <text class="section-title">二、信息使用目的</text>
      <text class="section-text">我们收集的信息将用于：</text>
      <text class="section-text">1. 保存和展示本人到店记录、路线进度与私人足迹</text>
      <text class="section-text">2. 对申请公开的文字和图片进行内容安全检测、人工审核和必要脱敏</text>
      <text class="section-text">3. 处理举报、申诉、权利投诉与账号安全风险</text>
      <text class="section-text">4. 保存和汇总仅本人可见的营养记录</text>

      <text class="section-title">三、信息存储与安全</text>
      <text class="section-text">1. 您的信息存储在安全的服务器上，采用加密传输（HTTPS）保护数据安全</text>
      <text class="section-text">2. 我们使用行业标准的数据库加密技术保护您的个人数据</text>
      <text class="section-text">3. 严格限制数据访问权限，仅授权人员可访问用户数据</text>
      <text class="section-text">4. 定期进行安全审计和漏洞检查</text>

      <text class="section-title">四、信息共享</text>
      <text class="section-text">1. 我们不会将您的个人信息出售或分享给任何第三方</text>
      <text class="section-text">2. 微信会处理登录、手机号授权和内容安全检测所需数据；地图能力仅在您主动选店或导航时使用</text>
      <text class="section-text">3. 公开图片通过安全检测和脱敏后才会展示，消费凭证不会出现在普通详情接口</text>

      <text class="section-title">五、您的权利</text>
      <text class="section-text">1. 您有权随时查看和修改您的个人资料信息</text>
      <text class="section-text">2. 您有权删除自己的到店记录和私人营养记录</text>
      <text class="section-text">3. 您有权注销账户并删除所有相关数据</text>
      <text class="section-text">4. 您有权撤回已授予的信息使用同意</text>

      <text class="section-title">六、未成年人保护</text>
      <text class="section-text">我们高度重视对未成年人个人信息的保护。若您是未满14周岁的未成年人，请在监护人指导下使用本产品，并在取得监护人同意后再提交个人信息。</text>

      <text class="section-title">七、政策更新</text>
      <text class="section-text">发生处理目的、信息类型或第三方处理方等重大变化时，我们会重新告知并在必要时重新取得您的同意。</text>

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
