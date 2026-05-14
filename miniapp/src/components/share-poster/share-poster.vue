<template>
  <view class="poster-overlay" v-if="visible" @tap="close">
    <view class="poster-container" @tap.stop>
      <view class="poster-header">
        <text class="poster-title">生成分享海报</text>
        <text class="poster-close" @tap="close">✕</text>
      </view>
      <view class="poster-card" :style="{ background: copy.bgColor }">
        <!-- Header -->
        <view class="poster-top">
          <view class="poster-avatar">{{ nickname.charAt(0) || '?' }}</view>
          <view class="poster-user-info">
            <text class="poster-nickname">{{ nickname || '匿名食客' }}</text>
            <text class="poster-date">{{ today }}</text>
          </view>
        </view>
        <!-- Main copy -->
        <view class="poster-main" :style="{ borderColor: copy.color }">
          <text class="poster-main-title" :style="{ color: copy.color }">{{ copy.title }}</text>
          <text class="poster-main-sub">{{ copy.subtitle }}</text>
        </view>
        <!-- Calories display -->
        <view class="poster-calories">
          <text class="poster-cal-num" :style="{ color: copy.color }">{{ Math.round(totalCalories) }}</text>
          <text class="poster-cal-unit">千卡</text>
        </view>
        <!-- Nutrients bar -->
        <view class="poster-nutrients">
          <view class="poster-nut-item" v-for="(val, key) in nutrients" :key="key">
            <text class="poster-nut-val" :style="{ color: getNutrientColor(key) }">{{ typeof val === 'number' ? val.toFixed(1) : val }}</text>
            <text class="poster-nut-label">{{ getNutrientLabel(key) }}</text>
          </view>
        </view>
        <!-- Score -->
        <view class="poster-score">
          <text>{{ copy.score }}</text>
        </view>
        <!-- Tagline -->
        <view class="poster-tagline">
          <text>{{ copy.tagline }}</text>
        </view>
        <!-- Footer -->
        <view class="poster-footer">
          <text class="poster-app-name">NutriLens · 热量判官</text>
        </view>
      </view>
      <view class="poster-actions">
        <button class="poster-btn poster-btn-save" @tap="savePoster">保存到相册</button>
        <button class="poster-btn poster-btn-share" open-type="share">分享给好友</button>
      </view>
    </view>
  </view>
</template>

<script setup>
import { computed, ref } from 'vue'
import { getPosterCopy, getNutrientLabel, getNutrientColor } from '../../utils/poster-template.js'

const props = defineProps({
  visible: Boolean,
  nickname: { type: String, default: '' },
  totalCalories: { type: Number, default: 0 },
  nutrients: { type: Object, default: () => ({}) },
})

const emit = defineEmits(['close', 'share'])

const today = new Date().toLocaleDateString('zh-CN', { month: 'long', day: 'numeric' })

const copy = computed(() => getPosterCopy(props.totalCalories))

function close() {
  emit('close')
}

function savePoster() {
  // In a real implementation, this would use Canvas to render and save
  uni.showActionSheet({
    itemList: ['保存海报图片'],
    success() {
      uni.showToast({ title: '长按海报图片保存', icon: 'none' })
    },
  })
}
</script>

<style scoped>
.poster-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.6);
  z-index: 999;
  display: flex;
  align-items: center;
  justify-content: center;
}

.poster-container {
  width: 85%;
  max-width: 340px;
}

.poster-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16rpx 24rpx;
}

.poster-title {
  font-size: 32rpx;
  font-weight: bold;
  color: #fff;
}

.poster-close {
  font-size: 36rpx;
  color: #fff;
  padding: 8rpx;
}

.poster-card {
  border-radius: 24rpx;
  padding: 32rpx;
  overflow: hidden;
}

.poster-top {
  display: flex;
  align-items: center;
  margin-bottom: 24rpx;
}

.poster-avatar {
  width: 64rpx;
  height: 64rpx;
  border-radius: 50%;
  background: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28rpx;
  font-weight: bold;
  color: #333;
  margin-right: 16rpx;
}

.poster-user-info {
  display: flex;
  flex-direction: column;
}

.poster-nickname {
  font-size: 28rpx;
  font-weight: bold;
  color: #333;
}

.poster-date {
  font-size: 22rpx;
  color: #666;
}

.poster-main {
  border-left: 6rpx solid;
  padding: 16rpx 20rpx;
  margin-bottom: 24rpx;
}

.poster-main-title {
  font-size: 36rpx;
  font-weight: bold;
  display: block;
  margin-bottom: 8rpx;
}

.poster-main-sub {
  font-size: 26rpx;
  color: #555;
  display: block;
}

.poster-calories {
  text-align: center;
  margin: 24rpx 0;
}

.poster-cal-num {
  font-size: 72rpx;
  font-weight: bold;
}

.poster-cal-unit {
  font-size: 28rpx;
  color: #666;
  margin-left: 8rpx;
}

.poster-nutrients {
  display: flex;
  justify-content: space-around;
  margin: 24rpx 0;
  flex-wrap: wrap;
}

.poster-nut-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  width: 25%;
  margin-bottom: 12rpx;
}

.poster-nut-val {
  font-size: 28rpx;
  font-weight: bold;
}

.poster-nut-label {
  font-size: 20rpx;
  color: #888;
}

.poster-score {
  text-align: center;
  font-size: 26rpx;
  color: #555;
  margin: 16rpx 0;
}

.poster-tagline {
  text-align: center;
  font-size: 24rpx;
  color: #888;
  font-style: italic;
  margin-bottom: 16rpx;
}

.poster-footer {
  text-align: center;
  padding-top: 16rpx;
  border-top: 1rpx solid rgba(0, 0, 0, 0.1);
}

.poster-app-name {
  font-size: 22rpx;
  color: #999;
}

.poster-actions {
  display: flex;
  gap: 16rpx;
  margin-top: 24rpx;
}

.poster-btn {
  flex: 1;
  height: 80rpx;
  line-height: 80rpx;
  border-radius: 40rpx;
  font-size: 28rpx;
  border: none;
}

.poster-btn-save {
  background: #fff;
  color: #333;
}

.poster-btn-share {
  background: #4CAF50;
  color: #fff;
}
</style>
