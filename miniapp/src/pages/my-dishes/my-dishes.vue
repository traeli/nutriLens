<template>
  <view class="my-dishes-page">
    <!-- Add Bar -->
    <view class="add-bar">
      <view class="add-row">
        <input class="add-input" :value="newDishName" @input="e => newDishName = e.detail.value"
          placeholder="输入菜品名称，AI自动分析营养" />
        <button class="add-btn" @tap="addDishAI" :loading="adding">{{ adding ? '分析中...' : '添加' }}</button>
      </view>
    </view>

    <!-- Dish List -->
    <view class="dish-list" v-if="dishes.length > 0">
      <view class="dish-card card" v-for="dish in dishes" :key="dish.id">
        <view class="dish-header">
          <view class="dish-title-row">
            <text class="dish-name">{{ dish.name }}</text>
            <text class="dish-category">{{ dish.category }}</text>
          </view>
          <text class="dish-delete" @tap="confirmDelete(dish)">删除</text>
        </view>
        <view class="dish-info-row">
          <text class="dish-cal">{{ dish.calories }} kcal</text>
          <text class="dish-weight" v-if="dish.unit_amount">{{ (dish.unit_amount * 1000).toFixed(0) }}g</text>
        </view>
        <view class="dish-nutrients" v-if="dish.nutrients">
          <view class="nutrient-tag" v-if="dish.nutrients.protein">
            <text class="nt-val">{{ dish.nutrients.protein.toFixed(1) }}</text>
            <text class="nt-label">蛋白g</text>
          </view>
          <view class="nutrient-tag" v-if="dish.nutrients.carbs">
            <text class="nt-val">{{ dish.nutrients.carbs.toFixed(1) }}</text>
            <text class="nt-label">碳水g</text>
          </view>
          <view class="nutrient-tag" v-if="dish.nutrients.fat">
            <text class="nt-val">{{ dish.nutrients.fat.toFixed(1) }}</text>
            <text class="nt-label">脂肪g</text>
          </view>
          <view class="nutrient-tag sugar" v-if="dish.nutrients.sugar">
            <text class="nt-val">{{ dish.nutrients.sugar.toFixed(1) }}</text>
            <text class="nt-label">糖分g</text>
          </view>
          <view class="nutrient-tag" v-if="dish.nutrients.vitamin_c">
            <text class="nt-val">{{ dish.nutrients.vitamin_c.toFixed(1) }}</text>
            <text class="nt-label">VitC mg</text>
          </view>
        </view>
        <view class="dish-recipe" v-if="dish.recipe">
          <text class="recipe-label">做法</text>
          <text class="recipe-text">{{ dish.recipe }}</text>
        </view>
        <view class="dish-stars">
          <text class="star" v-for="s in 5" :key="s"
            :class="{ filled: s <= starsFromWeight(dish.weight) }"
            @tap="setStars(dish, s)">★</text>
        </view>
      </view>
    </view>

    <view class="empty" v-else>
      <text class="empty-text">还没有自定义菜品</text>
      <text class="empty-hint">输入菜品名称添加</text>
    </view>
  </view>
</template>

<script>
import { api } from '@/api/request.js'

export default {
  data() {
    return {
      dishes: [],
      newDishName: '',
      adding: false,
    }
  },
  onShow() {
    if (!uni.getStorageSync('token')) {
      uni.reLaunch({ url: '/pages/login/login' })
      return
    }
    this.loadDishes()
  },
  methods: {
    async loadDishes() {
      try {
        const res = await api.getDishes()
        this.dishes = (res.dishes || []).filter(d => !d.is_system)
      } catch (e) {
        console.log('load dishes failed', e)
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
          weight: 50,
        })
        if (res.is_food === false) {
          uni.showToast({ title: '未识别到食物，请重新输入', icon: 'none', duration: 2000 })
          return
        }
        this.newDishName = ''
        this.loadDishes()
        uni.showToast({ title: 'AI分析完成，已添加', icon: 'success' })
      } catch (err) {
        uni.showToast({ title: '添加失败: ' + err.message, icon: 'none' })
      } finally {
        this.adding = false
      }
    },
    starsFromWeight(w) {
      if (w <= 10) return 1
      if (w <= 30) return 2
      if (w <= 50) return 3
      if (w <= 70) return 4
      return 5
    },
    setStars(dish, stars) {
      const weight = stars * 20
      dish.weight = weight
      api.createDish({
        name: dish.name,
        calories: dish.calories,
        category: dish.category,
        weight: weight,
        recipe: dish.recipe,
      }).catch(e => console.log('update pref failed', e))
    },
    confirmDelete(dish) {
      uni.showModal({
        title: '删除菜品',
        content: `确定删除 "${dish.name}" 吗？`,
        success: async (res) => {
          if (res.confirm) {
            try {
              await api.deleteDish(dish.id)
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
.my-dishes-page {
  padding: 20rpx 30rpx;
  min-height: 100vh;
}

.add-bar {
  margin-bottom: 20rpx;
}

.add-row {
  display: flex;
  gap: 16rpx;
}

.add-input {
  flex: 1;
  background: #f5f5f5;
  border-radius: 16rpx;
  padding: 20rpx 24rpx;
  font-size: 28rpx;
}

.add-btn {
  background: #4CAF50;
  color: #fff;
  border: none;
  border-radius: 16rpx;
  font-size: 26rpx;
  padding: 0 32rpx;
  white-space: nowrap;
}

.add-btn::after {
  border: none;
}

.dish-card {
  margin-bottom: 16rpx;
}

.dish-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12rpx;
}

.dish-title-row {
  display: flex;
  align-items: center;
  gap: 12rpx;
}

.dish-name {
  font-size: 30rpx;
  font-weight: bold;
  color: #333;
}

.dish-category {
  font-size: 22rpx;
  color: #fff;
  background: #81C784;
  padding: 4rpx 14rpx;
  border-radius: 8rpx;
}

.dish-delete {
  font-size: 24rpx;
  color: #FF7043;
}

.dish-info-row {
  display: flex;
  gap: 24rpx;
  margin-bottom: 12rpx;
}

.dish-cal {
  font-size: 28rpx;
  color: #FF7043;
  font-weight: bold;
}

.dish-weight {
  font-size: 24rpx;
  color: #999;
}

.dish-nutrients {
  display: flex;
  flex-wrap: wrap;
  gap: 10rpx;
  margin-bottom: 12rpx;
}

.nutrient-tag {
  display: flex;
  align-items: center;
  gap: 4rpx;
  background: #e8f5e9;
  padding: 6rpx 14rpx;
  border-radius: 12rpx;
}

.nutrient-tag.sugar {
  background: #fff3e0;
}

.nt-val {
  font-size: 22rpx;
  font-weight: bold;
  color: #4CAF50;
}

.nutrient-tag.sugar .nt-val {
  color: #FF7043;
}

.nt-label {
  font-size: 18rpx;
  color: #999;
}

.dish-recipe {
  background: #f5f5f5;
  border-radius: 12rpx;
  padding: 16rpx;
  margin-bottom: 12rpx;
}

.recipe-label {
  font-size: 22rpx;
  color: #4CAF50;
  font-weight: bold;
  display: block;
  margin-bottom: 8rpx;
}

.recipe-text {
  font-size: 24rpx;
  color: #666;
  line-height: 1.6;
}

.dish-stars {
  display: flex;
  gap: 8rpx;
}

.star {
  font-size: 40rpx;
  color: #ddd;
}

.star.filled {
  color: #FFB800;
}

.empty {
  text-align: center;
  padding: 100rpx 0;
}

.empty-text {
  font-size: 28rpx;
  color: #ccc;
  display: block;
  margin-bottom: 8rpx;
}

.empty-hint {
  font-size: 24rpx;
  color: #ddd;
}
</style>
