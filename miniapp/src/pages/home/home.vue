<template>
  <view class="home-page">
    <!-- Header: slogan + date -->
    <view class="header">
      <text class="slogan-text">{{ dailySlogan }}</text>
      <text class="date-text">{{ today }}</text>
    </view>

    <!-- Hero: Photo Action -->
    <view class="hero-action" @tap="takePhoto">
      <view class="hero-bg-circle hero-circle-1"></view>
      <view class="hero-bg-circle hero-circle-2"></view>
      <view class="hero-bg-circle hero-circle-3"></view>
      <view class="hero-content">
        <view class="hero-btn-icon">📸</view>
        <text class="hero-title">拍照识别</text>
        <text class="hero-sub">咔嚓一下，卡路里无处遁形</text>
        <view class="hero-cta">
          <text class="hero-cta-text">点击拍照</text>
        </view>
      </view>
    </view>

    <!-- Secondary: Text Input -->
    <view class="text-action" @tap="inputText">
      <text class="text-action-icon">✏️</text>
      <view class="text-action-info">
        <text class="text-action-title">文字记录</text>
        <text class="text-action-desc">说出来，AI帮你算</text>
      </view>
      <text class="text-action-arrow">›</text>
    </view>

    <!-- Today Summary -->
    <view class="summary-section" @tap="goHistory">
      <view class="summary-top">
        <view class="summary-left">
          <text class="summary-label">今日摄入</text>
          <view class="calorie-row">
            <text class="calorie-num">{{ summary.total_calories || 0 }}</text>
            <text class="calorie-unit">kcal</text>
          </view>
          <text class="calorie-comment">{{ calorieComment }}</text>
        </view>
        <view class="summary-right">
          <text class="summary-emoji">{{ calorieEmoji }}</text>
        </view>
      </view>
      <view class="nutrients-row" v-if="summary.total_nutrients">
        <view class="nutrient-pill protein-pill">
          <text class="pill-val">{{ (summary.total_nutrients.protein || 0).toFixed(1) }}</text>
          <text class="pill-label">蛋白质</text>
        </view>
        <view class="nutrient-pill carbs-pill">
          <text class="pill-val">{{ (summary.total_nutrients.carbs || 0).toFixed(1) }}</text>
          <text class="pill-label">碳水</text>
        </view>
        <view class="nutrient-pill fat-pill">
          <text class="pill-val">{{ (summary.total_nutrients.fat || 0).toFixed(1) }}</text>
          <text class="pill-label">脂肪</text>
        </view>
        <view class="nutrient-pill sugar-pill">
          <text class="pill-val">{{ (summary.total_nutrients.sugar || 0).toFixed(1) }}</text>
          <text class="pill-label">糖分</text>
        </view>
      </view>
      <text class="summary-link">查看详情 ›</text>
    </view>

    <!-- Daily Tip -->
    <view class="tip-card">
      <view class="tip-header">
        <text class="tip-tag">每日毒鸡汤</text>
      </view>
      <text class="tip-text">{{ dailyTip }}</text>
    </view>

    <!-- Quick Wheel + Rank -->
    <view class="shortcut-row">
      <view class="shortcut-card card" @tap="goWheel">
        <text class="shortcut-emoji">🎡</text>
        <text class="shortcut-title">吃什么</text>
        <text class="shortcut-desc">转盘决定</text>
      </view>
      <view class="shortcut-card card" @tap="goRank">
        <text class="shortcut-emoji">🏆</text>
        <text class="shortcut-title">排行榜</text>
        <text class="shortcut-desc">冲个榜</text>
      </view>
    </view>

    <!-- Analyze Result Cards -->
    <view class="result-section" v-if="analyzeResult" @tap="analyzeResult = null">
      <view class="card result-wrap" @tap.stop>
        <view class="result-header">
          <text class="result-section-title">AI 分析结果</text>
          <text class="result-close-btn" @tap="analyzeResult = null">✕</text>
        </view>
        <view class="food-card" v-for="item in analyzeResult.records" :key="item.id">
          <view class="food-card-header">
            <text class="food-card-name">{{ item.food_name }}</text>
            <text class="food-card-amount" v-if="item.unit_amount">{{ (item.unit_amount * 1000).toFixed(0) }}g</text>
            <text class="food-card-cal">{{ item.calories }} kcal</text>
          </view>
          <view class="food-card-nutrients" v-if="item.nutrients">
            <view class="nutrient-chip" v-if="item.nutrients.protein">
              <text class="chip-val">{{ item.nutrients.protein.toFixed(1) }}</text>
              <text class="chip-label">蛋白质g</text>
            </view>
            <view class="nutrient-chip" v-if="item.nutrients.carbs">
              <text class="chip-val">{{ item.nutrients.carbs.toFixed(1) }}</text>
              <text class="chip-label">碳水g</text>
            </view>
            <view class="nutrient-chip" v-if="item.nutrients.fat">
              <text class="chip-val">{{ item.nutrients.fat.toFixed(1) }}</text>
              <text class="chip-label">脂肪g</text>
            </view>
            <view class="nutrient-chip sugar-chip" v-if="item.nutrients.sugar">
              <text class="chip-val sugar-val">{{ item.nutrients.sugar.toFixed(1) }}</text>
              <text class="chip-label">糖分g</text>
            </view>
            <view class="nutrient-chip" v-if="item.nutrients.fiber">
              <text class="chip-val">{{ item.nutrients.fiber.toFixed(1) }}</text>
              <text class="chip-label">纤维g</text>
            </view>
            <view class="nutrient-chip" v-if="item.nutrients.vitamin_c">
              <text class="chip-val">{{ item.nutrients.vitamin_c.toFixed(1) }}</text>
              <text class="chip-label">VitC mg</text>
            </view>
          </view>
        </view>
        <view class="result-total">
          <text class="total-text">合计: {{ resultTotalCal }} kcal</text>
        </view>
        <view class="result-suggestion" v-if="analyzeResult.suggestion">
          <text class="suggestion-label">AI 建议</text>
          <text class="suggestion-text">{{ analyzeResult.suggestion }}</text>
        </view>
        <text class="result-close">点击空白处关闭</text>
      </view>
    </view>

    <!-- Text Input Modal -->
    <view class="modal" v-if="showTextInput" @tap="onModalMaskTap">
      <view class="modal-content card" @tap.stop>
        <text class="modal-title">记录你吃了什么</text>
        <textarea class="text-input" :value="foodText" @input="onFoodInput"
          placeholder="例如: 一碗米饭、半斤红烧肉、一杯牛奶..." :maxlength="500"
          :auto-focus="true" :adjust-position="false" cursor-spacing="20" />
        <view class="meal-type-picker">
          <text class="meal-label">用餐类型:</text>
          <view class="meal-options">
            <text class="meal-tag" v-for="m in mealTypes" :key="m.value"
              :class="{ active: mealType === m.value }" @tap="mealType = m.value">
              {{ m.label }}
            </text>
          </view>
        </view>
        <button class="btn-primary" @tap="submitText" :loading="analyzing">
          {{ analyzing ? 'AI分析中...' : '开始分析' }}
        </button>
      </view>
    </view>
  </view>
</template>

<script>
import { api } from '@/api/request.js'

const SLOGANS = [
  '吃好喝好，长生不老',
  '别饿着自己，毕竟你只有一个胃',
  '认真吃饭的人，运气不会太差',
  '你不是胖，你只是瘦得不明显',
  '三分练七分吃，剩下九十分靠基因',
  '今天也要做个快乐的饭桶',
  '没有什么是一顿饭解决不了的',
  '吃饭不积极，思想有问题',
  '干饭人，干饭魂',
  '你不吃饭，饭会伤心的',
]

const TIPS = [
  '你知道吗？大脑消耗了你每天20%的热量，所以多思考可以减肥（并不能）',
  '一根香蕉的热量约等于散步15分钟，但谁会在吃香蕉前去散步呢？',
  '辣椒素确实能加速代谢，但靠吃辣减肥约等于靠买彩票发财',
  '喝冷水确实会消耗热量——大约半个苹果的热量，恭喜你',
  '所谓的"负卡路里食物"是个美丽的谎言，就像"再吃最后一口"',
  '空腹可以跑步吗？可以，但你会跑得比思考人生还慢',
  '膳食纤维让你有饱腹感，但别想靠吃草变成羊',
  '每天喝8杯水？其实你从食物里已经偷喝了不少',
  '早餐确实很重要，但比起不吃早餐，吃油条豆浆更致命',
  '减肥最好的方法就是——别看美食视频（你正在看这个APP）',
]

export default {
  data() {
    return {
      summary: {},
      showTextInput: false,
      foodText: '',
      mealType: 1,
      analyzing: false,
      analyzeResult: null,
      mealTypes: [
        { label: '早餐', value: 1 },
        { label: '午餐', value: 2 },
        { label: '晚餐', value: 3 },
        { label: '加餐', value: 4 },
      ],
      dailySlogan: '',
      dailyTip: '',
    }
  },
  computed: {
    today() {
      const d = new Date()
      const weekDays = ['周日', '周一', '周二', '周三', '周四', '周五', '周六']
      return `${d.getMonth() + 1}月${d.getDate()}日 ${weekDays[d.getDay()]}`
    },
    calorieComment() {
      const cal = this.summary.total_calories || 0
      if (!cal) return '胃在抗议中...'
      if (cal < 500) return '这点能量，怎么拯救世界？'
      if (cal < 1500) return '稳住，我们能赢'
      if (cal <= 2000) return '完美，你是饮食界的天选之子'
      return '差不多了，放下手中的鸡腿'
    },
    calorieEmoji() {
      const cal = this.summary.total_calories || 0
      if (!cal) return '😴'
      if (cal < 500) return '🥬'
      if (cal < 1500) return '💪'
      if (cal <= 2000) return '🎉'
      return '🙈'
    },
    resultTotalCal() {
      if (!this.analyzeResult || !this.analyzeResult.records) return 0
      return this.analyzeResult.records.reduce((s, r) => s + r.calories, 0).toFixed(0)
    },
  },
  onShow() {
    const token = uni.getStorageSync('token')
    if (!token) {
      uni.reLaunch({ url: '/pages/login/login' })
      return
    }
    this.loadSummary()
  },
  onLoad(options) {
    const dayIndex = new Date().getDate() - 1
    this.dailySlogan = SLOGANS[dayIndex % SLOGANS.length]
    this.dailyTip = TIPS[dayIndex % TIPS.length]
    this.mealType = this.autoMealType()
    if (options && options.inviter_id) {
      uni.setStorageSync('inviter_id', options.inviter_id)
    }
  },
  methods: {
    onFoodInput(e) {
      this.foodText = e.detail.value
    },
    onModalMaskTap() {
      this.showTextInput = false
    },
    async loadSummary() {
      try {
        const res = await api.getDailySummary()
        this.summary = res
      } catch (e) {
        console.log('load summary failed', e)
      }
    },
    takePhoto() {
      uni.chooseImage({
        count: 1,
        sourceType: ['camera', 'album'],
        success: (res) => {
          const filePath = res.tempFilePaths[0]
          uni.navigateTo({
            url: `/pages/analyze/analyze?image=${encodeURIComponent(filePath)}`
          })
        },
      })
    },
    inputText() {
      this.showTextInput = true
    },
    async submitText() {
      if (!this.foodText.trim()) {
        uni.showToast({ title: '请输入食物描述', icon: 'none' })
        return
      }
      this.analyzing = true
      try {
        const res = await api.analyzeText({
          description: this.foodText,
          meal_type: this.mealType,
        })
        if (res.is_food === false) {
          uni.showToast({ title: '未识别到食物，请重新输入', icon: 'none', duration: 2000 })
          return
        }
        this.showTextInput = false
        this.foodText = ''
        this.analyzeResult = { records: res.records, suggestion: res.suggestion }
        this.loadSummary()
      } catch (err) {
        uni.showToast({ title: '分析失败: ' + err.message, icon: 'none' })
      } finally {
        this.analyzing = false
      }
    },
    goHistory() {
      uni.switchTab({ url: '/pages/history/history' })
    },
    goWheel() {
      uni.switchTab({ url: '/pages/wheel/wheel' })
    },
    goRank() {
      uni.navigateTo({ url: '/pages/rank/rank' })
    },
    autoMealType() {
      const h = new Date().getHours()
      if (h >= 5 && h < 11) return 1   // 早餐
      if (h >= 11 && h < 14) return 2   // 午餐
      if (h >= 14 && h < 20) return 3   // 晚餐
      return 4                           // 加餐
    },
  },
  onShareAppMessage() {
    const userId = uni.getStorageSync('user_id') || ''
    return {
      title: '饭友记 — AI拍照识别食物热量，超好用！',
      path: '/pages/home/home?inviter_id=' + userId,
    }
  },
  onShareTimeline() {
    const userId = uni.getStorageSync('user_id') || ''
    return {
      title: '饭友记 — AI拍照识别食物热量',
      path: '/pages/home/home?inviter_id=' + userId,
    }
  },
}
</script>

<style scoped>
.home-page {
  padding: 30rpx;
  min-height: 100vh;
}

/* ========== Header ========== */
.header {
  margin-bottom: 36rpx;
}

.slogan-text {
  font-size: 40rpx;
  font-weight: bold;
  color: #333;
  display: block;
  margin-bottom: 8rpx;
}

.date-text {
  font-size: 24rpx;
  color: #bbb;
}

/* ========== Hero Action ========== */
.hero-action {
  position: relative;
  background: linear-gradient(135deg, #FF9A76, #FECDA6);
  border-radius: 28rpx;
  padding: 48rpx 36rpx;
  margin-bottom: 20rpx;
  overflow: hidden;
}

.hero-bg-circle {
  position: absolute;
  border-radius: 50%;
}

.hero-circle-1 {
  width: 260rpx;
  height: 260rpx;
  background: rgba(255, 255, 255, 0.12);
  top: -80rpx;
  right: -40rpx;
}

.hero-circle-2 {
  width: 160rpx;
  height: 160rpx;
  background: rgba(255, 255, 255, 0.08);
  bottom: -50rpx;
  left: 40rpx;
}

.hero-circle-3 {
  width: 100rpx;
  height: 100rpx;
  background: rgba(255, 255, 255, 0.06);
  top: 40rpx;
  right: 100rpx;
}

.hero-content {
  position: relative;
  z-index: 1;
}

.hero-icon {
  font-size: 56rpx;
  display: block;
  margin-bottom: 12rpx;
}

.hero-btn-icon {
  width: 96rpx;
  height: 96rpx;
  background: rgba(255, 255, 255, 0.25);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 48rpx;
  margin-bottom: 20rpx;
}

.hero-title {
  font-size: 38rpx;
  font-weight: bold;
  color: #fff;
  display: block;
  margin-bottom: 8rpx;
  text-shadow: 0 2rpx 8rpx rgba(0, 0, 0, 0.08);
}

.hero-sub {
  font-size: 24rpx;
  color: rgba(255, 255, 255, 0.85);
  display: block;
  margin-bottom: 24rpx;
}

.hero-cta {
  display: inline-flex;
  background: #fff;
  border-radius: 40rpx;
  padding: 14rpx 40rpx;
}

.hero-cta-text {
  font-size: 28rpx;
  font-weight: bold;
  color: #FF9A76;
}

/* ========== Text Action ========== */
.text-action {
  display: flex;
  align-items: center;
  gap: 20rpx;
  background: #fff;
  border-radius: 20rpx;
  padding: 28rpx 30rpx;
  margin-bottom: 28rpx;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.06);
}

.text-action-icon {
  font-size: 36rpx;
}

.text-action-info {
  flex: 1;
}

.text-action-title {
  font-size: 28rpx;
  font-weight: bold;
  color: #333;
  display: block;
  margin-bottom: 4rpx;
}

.text-action-desc {
  font-size: 22rpx;
  color: #999;
}

.text-action-arrow {
  font-size: 36rpx;
  color: #ddd;
}

/* ========== Summary ========== */
.summary-section {
  background: #fff;
  border-radius: 24rpx;
  padding: 32rpx;
  margin-bottom: 20rpx;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.06);
}

.summary-top {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 28rpx;
}

.summary-label {
  font-size: 24rpx;
  color: #999;
  display: block;
  margin-bottom: 12rpx;
}

.calorie-row {
  display: flex;
  align-items: baseline;
  gap: 6rpx;
  margin-bottom: 8rpx;
}

.calorie-num {
  font-size: 72rpx;
  font-weight: bold;
  color: #333;
  line-height: 1;
}

.calorie-unit {
  font-size: 26rpx;
  color: #999;
}

.calorie-comment {
  font-size: 24rpx;
  color: #FF9A76;
  font-weight: 500;
}

.summary-emoji {
  font-size: 80rpx;
}

.nutrients-row {
  display: flex;
  gap: 16rpx;
  margin-bottom: 20rpx;
}

.nutrient-pill {
  flex: 1;
  text-align: center;
  padding: 16rpx 0;
  border-radius: 16rpx;
}

.protein-pill {
  background: #FFF3E0;
}

.carbs-pill {
  background: #E8F5E9;
}

.fat-pill {
  background: #FCE4EC;
}

.sugar-pill {
  background: #FFF9C4;
}

.pill-val {
  font-size: 28rpx;
  font-weight: bold;
  display: block;
  margin-bottom: 4rpx;
}

.protein-pill .pill-val { color: #E65100; }
.carbs-pill .pill-val { color: #2E7D32; }
.fat-pill .pill-val { color: #C62828; }
.sugar-pill .pill-val { color: #F57F17; }

.pill-label {
  font-size: 20rpx;
  color: #999;
}

.summary-link {
  font-size: 22rpx;
  color: #ccc;
  text-align: center;
  display: block;
}

/* ========== Tip Card ========== */
.tip-card {
  background: #fff;
  border-radius: 20rpx;
  padding: 28rpx 30rpx;
  margin-bottom: 20rpx;
  box-shadow: 0 2rpx 12rpx rgba(0, 0, 0, 0.06);
  border-left: 6rpx solid #FFB74D;
}

.tip-header {
  margin-bottom: 8rpx;
}

.tip-tag {
  font-size: 20rpx;
  color: #fff;
  background: linear-gradient(135deg, #FFB74D, #FF9800);
  padding: 4rpx 16rpx;
  border-radius: 8rpx;
}

.tip-text {
  font-size: 24rpx;
  color: #795548;
  line-height: 1.7;
}

/* ========== Shortcuts ========== */
.shortcut-row {
  display: flex;
  gap: 20rpx;
  margin-bottom: 20rpx;
}

.shortcut-card {
  flex: 1;
  text-align: center;
  padding: 32rpx 20rpx;
}

.shortcut-emoji {
  font-size: 48rpx;
  display: block;
  margin-bottom: 12rpx;
}

.shortcut-title {
  font-size: 28rpx;
  font-weight: bold;
  color: #333;
  display: block;
  margin-bottom: 4rpx;
}

.shortcut-desc {
  font-size: 22rpx;
  color: #999;
}

/* ========== Modal ========== */
.modal {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: flex-end;
  z-index: 999;
}

.modal-content {
  width: 100%;
  border-radius: 30rpx 30rpx 0 0;
  padding: 40rpx;
  padding-bottom: calc(40rpx + env(safe-area-inset-bottom));
}

.modal-title {
  font-size: 34rpx;
  font-weight: bold;
  display: block;
  margin-bottom: 20rpx;
}

.text-input {
  background: #f5f5f5;
  border-radius: 16rpx;
  padding: 24rpx;
  width: 100%;
  min-height: 200rpx;
  font-size: 28rpx;
  margin-bottom: 20rpx;
}

.meal-type-picker {
  margin-bottom: 30rpx;
}

.meal-label {
  font-size: 26rpx;
  color: #666;
  display: block;
  margin-bottom: 16rpx;
}

.meal-options {
  display: flex;
  gap: 16rpx;
}

.meal-tag {
  padding: 12rpx 28rpx;
  border-radius: 30rpx;
  background: #f5f5f5;
  font-size: 26rpx;
  color: #666;
}

.meal-tag.active {
  background: #FFF3E0;
  color: #FF9A76;
}

/* ========== Result Overlay ========== */
.result-section {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
  background: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 998;
}

.result-wrap {
  width: 90%;
  max-height: 75vh;
  overflow-y: auto;
}

.result-section-title {
  font-size: 32rpx;
  font-weight: bold;
  color: #333;
}

.result-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 24rpx;
}

.result-close-btn {
  width: 56rpx;
  height: 56rpx;
  line-height: 56rpx;
  text-align: center;
  font-size: 32rpx;
  color: #999;
  background: #f0f0f0;
  border-radius: 50%;
}

.food-card {
  background: #f9f9f9;
  border-radius: 16rpx;
  padding: 24rpx;
  margin-bottom: 16rpx;
}

.food-card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12rpx;
}

.food-card-name {
  font-size: 30rpx;
  font-weight: bold;
  color: #333;
}

.food-card-cal {
  font-size: 28rpx;
  font-weight: bold;
  color: #FF7043;
}

.food-card-amount {
  font-size: 24rpx;
  color: #999;
  background: #f5f5f5;
  padding: 4rpx 12rpx;
  border-radius: 8rpx;
}

.sugar-chip {
  background: #fff3e0 !important;
}

.chip-val.sugar-val {
  color: #FF7043;
}

.food-card-nutrients {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
}

.nutrient-chip {
  display: flex;
  align-items: center;
  gap: 4rpx;
  background: #e8f5e9;
  padding: 6rpx 16rpx;
  border-radius: 20rpx;
}

.chip-val {
  font-size: 24rpx;
  font-weight: bold;
  color: #4CAF50;
}

.chip-label {
  font-size: 20rpx;
  color: #666;
}

.result-total {
  text-align: center;
  padding: 20rpx 0;
  border-top: 1rpx solid #f0f0f0;
  margin-top: 8rpx;
}

.total-text {
  font-size: 30rpx;
  font-weight: bold;
  color: #FF7043;
}

.result-suggestion {
  background: #fff8e1;
  border-radius: 12rpx;
  padding: 20rpx;
  margin-top: 16rpx;
}

.suggestion-label {
  font-size: 24rpx;
  color: #FFA726;
  font-weight: bold;
  display: block;
  margin-bottom: 8rpx;
}

.suggestion-text {
  font-size: 26rpx;
  color: #666;
  line-height: 1.6;
}

.result-close {
  display: block;
  text-align: center;
  font-size: 22rpx;
  color: #ccc;
  margin-top: 20rpx;
}
</style>
