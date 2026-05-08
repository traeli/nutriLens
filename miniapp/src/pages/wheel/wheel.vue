<template>
  <view class="wheel-page">
    <!-- Category Filter -->
    <scroll-view class="category-bar" scroll-x>
      <text class="category-tag" :class="{ active: !selectedCategory }" @tap="selectCategory('')">全部</text>
      <text class="category-tag" :class="{ active: selectedCategory === cat }"
        v-for="cat in categories" :key="cat" @tap="selectCategory(cat)">
        {{ cat }}
      </text>
    </scroll-view>

    <!-- Wheel Canvas -->
    <view class="wheel-container">
      <canvas canvas-id="wheelCanvas" class="wheel-canvas"></canvas>
      <view class="wheel-pointer" :class="{ spinning: isSpinning }">▼</view>
    </view>

    <!-- Spin Button -->
    <button class="spin-btn" @tap="spin" :disabled="isSpinning">
      {{ isSpinning ? '旋转中...' : '转一转' }}
    </button>

    <!-- Result -->
    <view class="result-card card" v-if="result">
      <text class="result-name">{{ result.name }}</text>
      <text class="result-category">{{ result.category }}</text>
      <view class="result-main">
        <view class="result-cal-row">
          <text class="result-cal-num">{{ result.calories }}</text>
          <text class="result-cal-unit">kcal</text>
        </view>
        <text class="result-amount" v-if="result.unit_amount">{{ (result.unit_amount * 1000).toFixed(0) }}g</text>
      </view>
      <view class="result-nutrients" v-if="result.nutrients">
        <view class="nutrient-detail" v-if="result.nutrients.protein">
          <text class="nd-val">{{ result.nutrients.protein.toFixed(1) }}</text>
          <text class="nd-label">蛋白质g</text>
        </view>
        <view class="nutrient-detail" v-if="result.nutrients.carbs">
          <text class="nd-val">{{ result.nutrients.carbs.toFixed(1) }}</text>
          <text class="nd-label">碳水g</text>
        </view>
        <view class="nutrient-detail" v-if="result.nutrients.fat">
          <text class="nd-val">{{ result.nutrients.fat.toFixed(1) }}</text>
          <text class="nd-label">脂肪g</text>
        </view>
        <view class="nutrient-detail sugar" v-if="result.nutrients.sugar">
          <text class="nd-val">{{ result.nutrients.sugar.toFixed(1) }}</text>
          <text class="nd-label">糖分g</text>
        </view>
        <view class="nutrient-detail" v-if="result.nutrients.fiber">
          <text class="nd-val">{{ result.nutrients.fiber.toFixed(1) }}</text>
          <text class="nd-label">纤维g</text>
        </view>
        <view class="nutrient-detail" v-if="result.nutrients.vitamin_c">
          <text class="nd-val">{{ result.nutrients.vitamin_c.toFixed(1) }}</text>
          <text class="nd-label">VitC mg</text>
        </view>
      </view>
    </view>

    <!-- Add Dish -->
    <view class="card add-section">
      <view class="section-header">
        <text class="section-title">添加菜品</text>
      </view>
      <input class="add-input" :value="newDishName" @input="onDishInput" placeholder="输入菜品名称，AI自动分析营养" />
      <view class="pref-row">
        <text class="pref-text">喜爱程度</text>
        <view class="star-row">
          <text class="star" v-for="s in 5" :key="s"
            :class="{ filled: s <= newDishStars }"
            @tap="newDishStars = s">★</text>
        </view>
      </view>
      <button class="add-submit-btn" @tap="addDishAI" :loading="adding">
        {{ adding ? 'AI 分析中...' : '添加菜品' }}
      </button>
    </view>

    <!-- My Dishes List -->
    <view class="card" v-if="myDishes.length > 0">
      <view class="section-header">
        <text class="section-title">我的菜品</text>
      </view>
      <view class="dish-item" v-for="dish in myDishes" :key="dish.id">
        <view class="dish-left">
          <text class="dish-name">{{ dish.name }}</text>
          <text class="dish-meta">{{ dish.category }} · {{ dish.calories }} kcal</text>
          <view class="dish-nutrients" v-if="dish.nutrients">
            <text class="mini-chip" v-if="dish.nutrients.sugar">糖 {{ dish.nutrients.sugar.toFixed(1) }}g</text>
          </view>
        </view>
        <view class="dish-right">
          <view class="dish-stars-mini">
            <text class="mini-s" v-for="s in 5" :key="s"
              :class="{ on: s <= Math.round(dish.weight / 20) }">★</text>
          </view>
          <text class="dish-delete" @tap="deleteDish(dish.id)">删除</text>
        </view>
      </view>
    </view>
  </view>
</template>

<script>
import { api } from '@/api/request.js'

export default {
  data() {
    return {
      dishes: [],
      myDishes: [],
      selectedCategory: '',
      categories: [],
      isSpinning: false,
      result: null,
      newDishName: '',
      newDishStars: 3,
      adding: false,
      rotation: 0,
    }
  },
  onReady() {
    if (!uni.getStorageSync('token')) {
      uni.reLaunch({ url: '/pages/login/login' })
      return
    }
    this.loadDishes()
  },
  methods: {
    onDishInput(e) {
      this.newDishName = e.detail.value
    },
    async loadDishes() {
      try {
        const res = await api.getDishes(this.selectedCategory)
        this.dishes = res.dishes || []
        this.myDishes = this.dishes.filter(d => !d.is_system)
        const cats = new Set(this.dishes.map(d => d.category))
        this.categories = [...cats]
        this.$nextTick(() => {
          this.drawWheel()
        })
      } catch (e) {
        console.log('load dishes failed', e)
      }
    },
    selectCategory(cat) {
      this.selectedCategory = cat
      this.loadDishes()
    },
    drawWheel() {
      if (this.dishes.length === 0) return

      const ctx = uni.createCanvasContext('wheelCanvas', this)
      const w = 300
      const h = 300
      const cx = w / 2
      const cy = h / 2
      const r = Math.min(cx, cy) - 10

      const colors = ['#4CAF50', '#66BB6A', '#81C784', '#A5D6A7',
                      '#FF7043', '#FF8A65', '#FFA726', '#FFB74D',
                      '#42A5F5', '#64B5F6', '#AB47BC', '#BA68C8']
      const sliceAngle = (2 * Math.PI) / this.dishes.length

      this.dishes.forEach((dish, i) => {
        const startAngle = this.rotation + i * sliceAngle
        const endAngle = startAngle + sliceAngle

        ctx.beginPath()
        ctx.moveTo(cx, cy)
        ctx.arc(cx, cy, r, startAngle, endAngle)
        ctx.closePath()
        ctx.setFillStyle(colors[i % colors.length])
        ctx.fill()
        ctx.setStrokeStyle('#fff')
        ctx.setLineWidth(2)
        ctx.stroke()

        ctx.save()
        ctx.translate(cx, cy)
        ctx.rotate(startAngle + sliceAngle / 2)
        ctx.setFillStyle('#fff')
        ctx.setFontSize(12)
        ctx.setTextAlign('center')
        const text = dish.name.length > 4 ? dish.name.substring(0, 4) + '..' : dish.name
        ctx.fillText(text, r * 0.6, 4)
        ctx.restore()
      })

      ctx.beginPath()
      ctx.arc(cx, cy, 20, 0, 2 * Math.PI)
      ctx.setFillStyle('#fff')
      ctx.fill()
      ctx.setStrokeStyle('#4CAF50')
      ctx.setLineWidth(3)
      ctx.stroke()

      ctx.draw()
    },
    async spin() {
      if (this.isSpinning || this.dishes.length === 0) return
      this.isSpinning = true
      this.result = null

      try {
        const res = await api.spinWheel(this.selectedCategory)
        const selectedDish = res.dish
        const selectedIndex = this.dishes.findIndex(d => d.id === selectedDish.id)

        const sliceAngle = (2 * Math.PI) / this.dishes.length
        const targetAngle = 2 * Math.PI * 5 + (2 * Math.PI - selectedIndex * sliceAngle - sliceAngle / 2 - Math.PI / 2)
        const duration = 4000
        const startTime = Date.now()
        const startRotation = this.rotation

        const animate = () => {
          const elapsed = Date.now() - startTime
          const progress = Math.min(elapsed / duration, 1)
          const eased = 1 - Math.pow(1 - progress, 3)
          this.rotation = startRotation + targetAngle * eased

          this.drawWheel()

          if (progress < 1) {
            setTimeout(animate, 30)
          } else {
            this.isSpinning = false
            this.result = selectedDish
          }
        }
        animate()
      } catch (err) {
        this.isSpinning = false
        uni.showToast({ title: '转盘失败: ' + err.message, icon: 'none' })
      }
    },
    async addDishAI() {
      if (!this.newDishName.trim()) {
        uni.showToast({ title: '请输入菜品名称', icon: 'none' })
        return
      }
      this.adding = true
      try {
        const res = await api.createDishAI({
          name: this.newDishName.trim(),
          weight: this.newDishStars * 20,
        })
        if (res.is_food === false) {
          uni.showToast({ title: '未识别到食物，请重新输入', icon: 'none', duration: 2000 })
          return
        }
        this.newDishName = ''
        this.newDishStars = 3
        this.loadDishes()
        uni.showToast({ title: 'AI分析完成，已添加', icon: 'success' })
      } catch (err) {
        uni.showToast({ title: '添加失败: ' + err.message, icon: 'none' })
      } finally {
        this.adding = false
      }
    },
    async deleteDish(id) {
      uni.showModal({
        title: '删除菜品',
        content: '确定删除这个菜品吗？',
        success: async (res) => {
          if (res.confirm) {
            try {
              await api.deleteDish(id)
              this.loadDishes()
              uni.showToast({ title: '已删除', icon: 'success' })
            } catch (err) {
              uni.showToast({ title: '删除失败', icon: 'none' })
            }
          }
        },
      })
    },
  },
}
</script>

<style scoped>
.wheel-page {
  padding: 20rpx 30rpx;
  min-height: 100vh;
}

.category-bar {
  white-space: nowrap;
  margin-bottom: 20rpx;
}

.category-tag {
  display: inline-block;
  padding: 12rpx 28rpx;
  border-radius: 30rpx;
  background: #fff;
  font-size: 26rpx;
  color: #666;
  margin-right: 16rpx;
}

.category-tag.active {
  background: #4CAF50;
  color: #fff;
}

.wheel-container {
  position: relative;
  display: flex;
  justify-content: center;
  margin: 20rpx 0;
}

.wheel-canvas {
  width: 600rpx;
  height: 600rpx;
}

.wheel-pointer {
  position: absolute;
  top: -10rpx;
  left: 50%;
  transform: translateX(-50%);
  font-size: 40rpx;
  color: #FF7043;
  z-index: 10;
}

.spin-btn {
  background: linear-gradient(135deg, #4CAF50, #45a049);
  color: #fff;
  border: none;
  border-radius: 50rpx;
  padding: 24rpx 0;
  font-size: 34rpx;
  font-weight: bold;
  text-align: center;
  margin: 20rpx auto;
  width: 60%;
}

.spin-btn::after {
  border: none;
}

.spin-btn[disabled] {
  opacity: 0.6;
}

.result-card {
  margin-bottom: 20rpx;
}

.result-name {
  font-size: 40rpx;
  font-weight: bold;
  color: #4CAF50;
  display: block;
  margin-bottom: 6rpx;
}

.result-category {
  font-size: 24rpx;
  color: #999;
  display: block;
  margin-bottom: 16rpx;
}

.result-main {
  display: flex;
  align-items: baseline;
  gap: 16rpx;
  margin-bottom: 20rpx;
}

.result-cal-row {
  display: flex;
  align-items: baseline;
  gap: 6rpx;
}

.result-cal-num {
  font-size: 56rpx;
  font-weight: bold;
  color: #FF7043;
}

.result-cal-unit {
  font-size: 26rpx;
  color: #999;
}

.result-amount {
  font-size: 24rpx;
  color: #999;
  background: #f5f5f5;
  padding: 4rpx 14rpx;
  border-radius: 8rpx;
}

.result-nutrients {
  display: flex;
  flex-wrap: wrap;
  gap: 12rpx;
  padding-top: 16rpx;
  border-top: 1rpx solid #f0f0f0;
}

.nutrient-detail {
  display: flex;
  flex-direction: column;
  align-items: center;
  background: #e8f5e9;
  padding: 10rpx 20rpx;
  border-radius: 12rpx;
  min-width: 100rpx;
}

.nutrient-detail.sugar {
  background: #fff3e0;
}

.nd-val {
  font-size: 28rpx;
  font-weight: bold;
  color: #4CAF50;
}

.nutrient-detail.sugar .nd-val {
  color: #FF7043;
}

.nd-label {
  font-size: 18rpx;
  color: #999;
}

.add-section {
  margin-bottom: 20rpx;
}

.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16rpx;
}

.section-title {
  font-size: 30rpx;
  font-weight: bold;
  color: #333;
}

.add-input {
  background: #f5f5f5;
  border-radius: 16rpx;
  padding: 20rpx 24rpx;
  font-size: 28rpx;
  margin-bottom: 16rpx;
}

.pref-row {
  display: flex;
  align-items: center;
  gap: 16rpx;
  margin-bottom: 20rpx;
}

.pref-text {
  font-size: 26rpx;
  color: #666;
  white-space: nowrap;
}

.star-row {
  display: flex;
  gap: 8rpx;
}

.star {
  font-size: 44rpx;
  color: #ddd;
}

.star.filled {
  color: #FFB800;
}

.add-submit-btn {
  background: linear-gradient(135deg, #4CAF50, #45a049);
  color: #fff;
  border: none;
  border-radius: 50rpx;
  font-size: 28rpx;
  padding: 20rpx 0;
}

.add-submit-btn::after {
  border: none;
}

.dish-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 20rpx 0;
  border-bottom: 1rpx solid #f0f0f0;
}

.dish-item:last-child {
  border-bottom: none;
}

.dish-left {
  flex: 1;
}

.dish-name {
  font-size: 28rpx;
  font-weight: bold;
  color: #333;
  display: block;
  margin-bottom: 4rpx;
}

.dish-meta {
  font-size: 22rpx;
  color: #999;
  display: block;
  margin-bottom: 6rpx;
}

.dish-nutrients {
  display: flex;
  gap: 8rpx;
}

.mini-chip {
  font-size: 20rpx;
  color: #FF7043;
  background: #fff3e0;
  padding: 2rpx 10rpx;
  border-radius: 8rpx;
}

.dish-right {
  display: flex;
  flex-direction: column;
  align-items: flex-end;
  gap: 10rpx;
}

.dish-stars-mini {
  display: flex;
  gap: 2rpx;
}

.mini-s {
  font-size: 26rpx;
  color: #ddd;
}

.mini-s.on {
  color: #FFB800;
}

.dish-delete {
  font-size: 22rpx;
  color: #FF7043;
  padding: 4rpx 12rpx;
}
</style>
