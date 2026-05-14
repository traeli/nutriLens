<template>
  <view class="achievement-page">
    <!-- Summary -->
    <view class="ach-summary">
      <view class="ach-summary-item">
        <text class="ach-summary-num">{{ unlockedCount }}</text>
        <text class="ach-summary-label">已解锁</text>
      </view>
      <view class="ach-summary-divider"></view>
      <view class="ach-summary-item">
        <text class="ach-summary-num">{{ totalCount }}</text>
        <text class="ach-summary-label">总成就</text>
      </view>
      <view class="ach-summary-divider"></view>
      <view class="ach-summary-item">
        <text class="ach-summary-num">{{ legendaryCount }}</text>
        <text class="ach-summary-label">传说</text>
      </view>
    </view>

    <!-- Current title -->
    <view class="ach-current-title" v-if="currentTitle">
      <text class="ach-current-label">当前称号</text>
      <text class="ach-current-name">{{ currentTitle }}</text>
    </view>

    <!-- Rarity filter -->
    <view class="ach-filter">
      <view
        v-for="f in filters"
        :key="f.key"
        :class="['ach-filter-item', activeFilter === f.key && 'ach-filter-active']"
        @tap="activeFilter = f.key"
      >
        <text>{{ f.label }}</text>
      </view>
    </view>

    <!-- Achievement grid -->
    <scroll-view scroll-y class="ach-grid">
      <view class="ach-card" v-for="item in filteredAchievements" :key="item.id"
        :class="{ 'ach-card-locked': !item.unlocked }"
        @tap="onAchievementTap(item)"
      >
        <view :class="['ach-icon-wrap', `rarity-${item.rarity}`]">
          <text class="ach-icon">{{ rarityIcon(item.rarity) }}</text>
        </view>
        <text class="ach-name">{{ item.name }}</text>
        <text class="ach-desc">{{ item.description }}</text>
        <view v-if="item.unlocked" class="ach-badge">已解锁</view>
        <view v-if="item.unlocked && item.reward_type" class="ach-reward-tag">
          {{ rewardLabel(item.reward_type, item.reward_value) }}
        </view>
      </view>
    </scroll-view>

    <!-- New unlock celebration -->
    <view class="ach-celebration" v-if="showCelebration" @tap="showCelebration = false">
      <view class="ach-celebrate-card">
        <text class="ach-celebrate-title">恭喜解锁新成就！</text>
        <text class="ach-celebrate-name">{{ newAchievement.name }}</text>
        <text class="ach-celebrate-desc">{{ newAchievement.description }}</text>
        <view v-if="newAchievement.reward_type" class="ach-celebrate-reward">
          奖励：{{ rewardLabel(newAchievement.reward_type, newAchievement.reward_value) }}
        </view>
        <text class="ach-celebrate-close">点击关闭</text>
      </view>
    </view>
  </view>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { api } from '../../api/request.js'

const filters = [
  { key: 'all', label: '全部' },
  { key: 'common', label: '普通' },
  { key: 'rare', label: '稀有' },
  { key: 'epic', label: '史诗' },
  { key: 'legendary', label: '传说' },
]

const activeFilter = ref('all')
const achievements = ref([])
const showCelebration = ref(false)
const newAchievement = ref({})

const unlockedCount = computed(() => achievements.value.filter(a => a.unlocked).length)
const totalCount = computed(() => achievements.value.length)
const legendaryCount = computed(() => achievements.value.filter(a => a.unlocked && a.rarity === 'legendary').length)
const currentTitle = computed(() => {
  const equipped = achievements.value.find(a => a.unlocked && a.is_equipped)
  return equipped ? equipped.name : ''
})

const filteredAchievements = computed(() => {
  if (activeFilter.value === 'all') return achievements.value
  return achievements.value.filter(a => a.rarity === activeFilter.value)
})

onMounted(() => {
  loadAchievements()
  checkNewUnlocks()
})

async function loadAchievements() {
  try {
    const res = await api.getAchievements()
    achievements.value = res.achievements || []
  } catch (e) {
    console.error('load achievements failed:', e)
  }
}

async function checkNewUnlocks() {
  try {
    const res = await api.checkAchievements()
    const unlocks = res.new_achievements || []
    if (unlocks.length > 0) {
      newAchievement.value = unlocks[0]
      showCelebration.value = true
      await loadAchievements()
    }
  } catch (e) {
    console.error('check achievements failed:', e)
  }
}

async function onAchievementTap(item) {
  if (!item.unlocked) {
    uni.showToast({ title: '尚未解锁', icon: 'none' })
    return
  }
  uni.showActionSheet({
    itemList: ['佩戴此称号'],
    success: async () => {
      try {
        await api.setTitle({ achievement_id: item.id })
        uni.showToast({ title: '称号已更换', icon: 'success' })
        await loadAchievements()
      } catch (e) {
        uni.showToast({ title: '设置失败', icon: 'none' })
      }
    },
  })
}

function rarityIcon(rarity) {
  const map = { common: '⭐', rare: '💎', epic: '🔥', legendary: '👑' }
  return map[rarity] || '⭐'
}

function rewardLabel(type, value) {
  if (type === 'quota') return `+${value}次识别`
  if (type === 'vip') return '永久不限次'
  if (type === 'poster_template') return '专属海报模板'
  return ''
}
</script>

<style scoped>
.achievement-page {
  min-height: 100vh;
  background: #f5f5f5;
}

.ach-summary {
  display: flex;
  justify-content: space-around;
  padding: 32rpx;
  background: linear-gradient(135deg, #FF9800, #FF5722);
  color: #fff;
}

.ach-summary-item {
  display: flex;
  flex-direction: column;
  align-items: center;
}

.ach-summary-num {
  font-size: 40rpx;
  font-weight: bold;
}

.ach-summary-label {
  font-size: 22rpx;
  opacity: 0.8;
}

.ach-summary-divider {
  width: 1rpx;
  height: 60rpx;
  background: rgba(255, 255, 255, 0.3);
}

.ach-current-title {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12rpx;
  padding: 16rpx;
  background: #FFF3E0;
}

.ach-current-label {
  font-size: 24rpx;
  color: #999;
}

.ach-current-name {
  font-size: 28rpx;
  color: #FF9800;
  font-weight: bold;
}

.ach-filter {
  display: flex;
  padding: 16rpx;
  gap: 12rpx;
  overflow-x: auto;
  white-space: nowrap;
}

.ach-filter-item {
  padding: 8rpx 24rpx;
  border-radius: 24rpx;
  font-size: 24rpx;
  background: #fff;
  color: #666;
  flex-shrink: 0;
}

.ach-filter-active {
  background: #4CAF50;
  color: #fff;
}

.ach-grid {
  height: calc(100vh - 380rpx);
  padding: 0 16rpx;
}

.ach-card {
  display: flex;
  align-items: center;
  padding: 20rpx;
  background: #fff;
  border-radius: 16rpx;
  margin-bottom: 12rpx;
  position: relative;
}

.ach-card-locked {
  opacity: 0.5;
}

.ach-icon-wrap {
  width: 64rpx;
  height: 64rpx;
  border-radius: 16rpx;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 16rpx;
  flex-shrink: 0;
}

.rarity-common { background: #F5F5F5; }
.rarity-rare { background: #E3F2FD; }
.rarity-epic { background: #F3E5F5; }
.rarity-legendary { background: #FFF3E0; }

.ach-icon {
  font-size: 32rpx;
}

.ach-name {
  font-size: 28rpx;
  font-weight: bold;
  color: #333;
  display: block;
}

.ach-desc {
  font-size: 22rpx;
  color: #999;
  display: block;
}

.ach-badge {
  position: absolute;
  top: 12rpx;
  right: 16rpx;
  font-size: 20rpx;
  color: #4CAF50;
  background: #E8F5E9;
  padding: 2rpx 12rpx;
  border-radius: 8rpx;
}

.ach-reward-tag {
  font-size: 20rpx;
  color: #FF9800;
  margin-left: auto;
  flex-shrink: 0;
}

.ach-celebration {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 999;
}

.ach-celebrate-card {
  width: 80%;
  background: #fff;
  border-radius: 24rpx;
  padding: 48rpx 32rpx;
  text-align: center;
}

.ach-celebrate-title {
  font-size: 32rpx;
  font-weight: bold;
  color: #FF9800;
  display: block;
  margin-bottom: 16rpx;
}

.ach-celebrate-name {
  font-size: 40rpx;
  font-weight: bold;
  display: block;
  margin-bottom: 12rpx;
}

.ach-celebrate-desc {
  font-size: 26rpx;
  color: #666;
  display: block;
  margin-bottom: 16rpx;
}

.ach-celebrate-reward {
  font-size: 28rpx;
  color: #4CAF50;
  font-weight: bold;
  margin-bottom: 24rpx;
}

.ach-celebrate-close {
  font-size: 24rpx;
  color: #999;
}
</style>
