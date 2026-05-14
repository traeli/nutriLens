<template>
  <view class="rank-page">
    <!-- Tabs -->
    <view class="rank-tabs">
      <view
        v-for="tab in tabs"
        :key="tab.key"
        :class="['rank-tab', activeTab === tab.key && 'rank-tab-active']"
        @tap="switchTab(tab.key)"
      >
        <text>{{ tab.label }}</text>
      </view>
    </view>

    <!-- Rank opt-in prompt -->
    <view class="rank-optin" v-if="!showOnRank">
      <view class="rank-optin-card">
        <text class="rank-optin-icon">🏆</text>
        <text class="rank-optin-title">开启排行榜</text>
        <text class="rank-optin-desc">开启后，你的昵称和分数将在排行榜中展示，快来看看你的排名吧！</text>
        <button class="rank-optin-btn" @tap="enableRank">确认开启</button>
        <text class="rank-optin-hint">开启后可随时关闭</text>
      </view>
    </view>

    <!-- My rank -->
    <view class="my-rank" v-if="showOnRank && myScore">
      <view class="my-rank-left">
        <view class="my-rank-avatar">{{ (nickname || '?').charAt(0) }}</view>
        <view class="my-rank-info">
          <text class="my-rank-name">{{ nickname || '我' }}</text>
          <text class="my-rank-streak" v-if="myScore.streak_days > 0">
            连续打卡 {{ myScore.streak_days }} 天
          </text>
        </view>
      </view>
      <view class="my-rank-right">
        <text class="my-rank-pos">第 {{ myScore.week_rank || myScore.total_rank }} 名</text>
        <text class="my-rank-score">{{ myScore.week_score || myScore.total_score }} 分</text>
      </view>
    </view>

    <!-- Rank list -->
    <scroll-view scroll-y class="rank-list" v-if="showOnRank">
      <view v-if="rankList.length === 0" class="rank-empty">
        <text>暂无排名数据</text>
      </view>
      <view
        v-for="(item, index) in rankList"
        :key="item.user_id"
        :class="['rank-item', index < 3 && 'rank-item-top']"
      >
        <view :class="['rank-pos', index === 0 && 'rank-pos-gold', index === 1 && 'rank-pos-silver', index === 2 && 'rank-pos-bronze']">
          <text v-if="index < 3" class="rank-medal">{{ ['🥇', '🥈', '🥉'][index] }}</text>
          <text v-else>{{ index + 1 }}</text>
        </view>
        <view class="rank-avatar">{{ (item.nickname || '?').charAt(0) }}</view>
        <view class="rank-info">
          <view class="rank-name-row">
            <text class="rank-name">{{ item.nickname || '匿名' }}</text>
            <text class="rank-title" v-if="item.title">{{ item.title }}</text>
          </view>
          <text class="rank-streak" v-if="item.streak_days > 0">
            连续 {{ item.streak_days }} 天
          </text>
        </view>
        <view class="rank-score">
          <text class="rank-score-num">{{ item.week_score || item.score }}</text>
          <text class="rank-score-label">分</text>
        </view>
      </view>
    </scroll-view>

    <!-- Toggle button at bottom -->
    <view class="rank-toggle" v-if="showOnRank">
      <text class="rank-toggle-text" @tap="disableRank">隐藏我的排名</text>
    </view>
  </view>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { api } from '../../api/request.js'

const tabs = [
  { key: 'week', label: '周榜' },
  { key: 'total', label: '总榜' },
  { key: 'friend', label: '好友榜' },
]

const activeTab = ref('week')
const rankList = ref([])
const myScore = ref(null)
const nickname = ref('')
const showOnRank = ref(false)

onMounted(() => {
  nickname.value = uni.getStorageSync('nickname') || ''
  loadMyScore()
})

async function switchTab(key) {
  activeTab.value = key
  await loadRank()
}

async function loadRank() {
  try {
    const res = await api.getRank({ scope: activeTab.value })
    rankList.value = res.rank || []
  } catch (e) {
    console.error('load rank failed:', e)
  }
}

async function loadMyScore() {
  try {
    const res = await api.getMyScore()
    myScore.value = res
    showOnRank.value = res.show_on_rank || false
    if (showOnRank.value) {
      loadRank()
    }
  } catch (e) {
    console.error('load my score failed:', e)
  }
}

async function enableRank() {
  try {
    const res = await api.toggleRankVisibility()
    showOnRank.value = res.show_on_rank
    if (showOnRank.value) {
      loadRank()
    }
    uni.showToast({ title: '已开启排行榜', icon: 'success' })
  } catch (e) {
    uni.showToast({ title: '操作失败', icon: 'none' })
  }
}

async function disableRank() {
  uni.showModal({
    title: '隐藏排名',
    content: '隐藏后你将不会出现在排行榜中，确定吗？',
    success: async (res) => {
      if (!res.confirm) return
      try {
        const res = await api.toggleRankVisibility()
        showOnRank.value = res.show_on_rank
        rankList.value = []
        uni.showToast({ title: '已隐藏', icon: 'success' })
      } catch (e) {
        uni.showToast({ title: '操作失败', icon: 'none' })
      }
    },
  })
}
</script>

<style scoped>
.rank-page {
  min-height: 100vh;
  background: #f5f5f5;
}

.rank-tabs {
  display: flex;
  background: #fff;
  padding: 20rpx 0;
  border-bottom: 1rpx solid #eee;
}

.rank-tab {
  flex: 1;
  text-align: center;
  padding: 16rpx 0;
  font-size: 28rpx;
  color: #666;
  position: relative;
}

.rank-tab-active {
  color: #4CAF50;
  font-weight: bold;
}

.rank-tab-active::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 50%;
  transform: translateX(-50%);
  width: 48rpx;
  height: 4rpx;
  background: #4CAF50;
  border-radius: 2rpx;
}

/* ========== Opt-in Card ========== */
.rank-optin {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 80rpx 40rpx;
}

.rank-optin-card {
  background: #fff;
  border-radius: 24rpx;
  padding: 60rpx 40rpx;
  text-align: center;
  box-shadow: 0 4rpx 24rpx rgba(0, 0, 0, 0.08);
  width: 100%;
}

.rank-optin-icon {
  font-size: 80rpx;
  display: block;
  margin-bottom: 24rpx;
}

.rank-optin-title {
  font-size: 34rpx;
  font-weight: bold;
  color: #333;
  display: block;
  margin-bottom: 16rpx;
}

.rank-optin-desc {
  font-size: 26rpx;
  color: #999;
  line-height: 1.6;
  display: block;
  margin-bottom: 40rpx;
}

.rank-optin-btn {
  background: linear-gradient(135deg, #4CAF50, #45a049);
  color: #fff;
  border-radius: 50rpx;
  padding: 24rpx 0;
  font-size: 32rpx;
  font-weight: bold;
  border: none;
}

.rank-optin-hint {
  font-size: 22rpx;
  color: #ccc;
  display: block;
  margin-top: 20rpx;
}

/* ========== My Rank ========== */
.my-rank {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin: 20rpx;
  padding: 24rpx;
  background: linear-gradient(135deg, #4CAF50, #8BC34A);
  border-radius: 16rpx;
  color: #fff;
}

.my-rank-left {
  display: flex;
  align-items: center;
}

.my-rank-avatar {
  width: 64rpx;
  height: 64rpx;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.3);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28rpx;
  font-weight: bold;
  margin-right: 16rpx;
}

.my-rank-name {
  font-size: 28rpx;
  font-weight: bold;
  display: block;
}

.my-rank-streak {
  font-size: 22rpx;
  opacity: 0.8;
}

.my-rank-right {
  text-align: right;
}

.my-rank-pos {
  font-size: 24rpx;
  display: block;
  opacity: 0.9;
}

.my-rank-score {
  font-size: 36rpx;
  font-weight: bold;
}

/* ========== Rank List ========== */
.rank-list {
  height: calc(100vh - 380rpx);
}

.rank-empty {
  text-align: center;
  padding: 100rpx;
  color: #999;
}

.rank-item {
  display: flex;
  align-items: center;
  padding: 20rpx 24rpx;
  background: #fff;
  border-bottom: 1rpx solid #f0f0f0;
}

.rank-item-top {
  background: #FFFDE7;
}

.rank-pos {
  width: 56rpx;
  text-align: center;
  font-size: 28rpx;
  font-weight: bold;
  color: #999;
}

.rank-medal {
  font-size: 36rpx;
}

.rank-avatar {
  width: 56rpx;
  height: 56rpx;
  border-radius: 50%;
  background: #E8F5E9;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24rpx;
  color: #4CAF50;
  font-weight: bold;
  margin: 0 16rpx;
}

.rank-info {
  flex: 1;
}

.rank-name-row {
  display: flex;
  align-items: center;
  gap: 8rpx;
}

.rank-name {
  font-size: 28rpx;
  color: #333;
}

.rank-title {
  font-size: 20rpx;
  color: #FF9800;
  background: #FFF3E0;
  padding: 2rpx 10rpx;
  border-radius: 8rpx;
}

.rank-streak {
  font-size: 22rpx;
  color: #999;
}

.rank-score {
  text-align: right;
}

.rank-score-num {
  font-size: 32rpx;
  font-weight: bold;
  color: #4CAF50;
}

.rank-score-label {
  font-size: 22rpx;
  color: #999;
}

/* ========== Toggle ========== */
.rank-toggle {
  padding: 24rpx;
  text-align: center;
}

.rank-toggle-text {
  font-size: 24rpx;
  color: #999;
  text-decoration: underline;
}
</style>
