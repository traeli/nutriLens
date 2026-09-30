<template>
  <view class="legal-page">
    <view class="legal-header">
      <text class="legal-title">{{ document.title }}</text>
      <text class="legal-date">更新日期：{{ updatedAt }}</text>
    </view>
    <view class="legal-body">
      <view v-for="section in document.sections" :key="section.title">
        <text class="section-title">{{ section.title }}</text>
        <text v-for="line in section.lines" :key="line" class="section-text">{{ line }}</text>
      </view>
    </view>
  </view>
</template>

<script>
import { AGREEMENTS, AGREEMENT_UPDATED_AT } from '@/content/agreements.js'

export default {
  data() { return { document: AGREEMENTS.user_service, updatedAt: AGREEMENT_UPDATED_AT } },
}
</script>

<style scoped>
.legal-page { padding: 30rpx; min-height: 100vh; background: #fff; }
.legal-header { text-align: center; margin-bottom: 40rpx; padding-bottom: 30rpx; border-bottom: 1rpx solid #f0f0f0; }
.legal-title { display: block; margin-bottom: 12rpx; color: #333; font-size: 38rpx; font-weight: bold; }
.legal-date { color: #999; font-size: 24rpx; }
.section-title { display: block; margin-top: 30rpx; margin-bottom: 16rpx; color: #333; font-size: 30rpx; font-weight: bold; }
.section-text { display: block; margin-bottom: 8rpx; color: #666; font-size: 26rpx; line-height: 1.8; }
</style>
