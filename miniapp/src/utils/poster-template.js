/**
 * 海报文案模板库 — "热量判官"夸张搞笑风
 */

// 根据今日总热量选择文案
export function getPosterCopy(totalCalories) {
  if (totalCalories < 1200) {
    return {
      title: '小仙女本仙',
      subtitle: '今天连空气都不敢多吃',
      score: '自律指数 ★★★★★',
      tagline: '你的饮食比我减肥期还狠',
      color: '#4CAF50',
      bgColor: '#E8F5E9',
    }
  }
  if (totalCalories < 1800) {
    return {
      title: '热量守恒定律',
      subtitle: '你的午餐 = 3杯奶茶',
      score: '正常指数 ★★★☆☆',
      tagline: '科学饮食，从记录每一口开始',
      color: '#2196F3',
      bgColor: '#E3F2FD',
    }
  }
  if (totalCalories < 2500) {
    return {
      title: '热量超标警告',
      subtitle: '你今天吃的热量比隔壁胖子还多30%',
      score: '超标指数 ★★★★☆',
      tagline: '放下那块蛋糕，立地成佛',
      color: '#FF9800',
      bgColor: '#FFF3E0',
    }
  }
  return {
    title: '热量炸弹已引爆',
    subtitle: '今天的热量够跑一个马拉松了',
    score: '暴走指数 ★★★★★',
    tagline: '承认吧，你就是个吃货',
    color: '#F44336',
    bgColor: '#FFEBEE',
  }
}

// 营养素标签
export function getNutrientLabel(key) {
  const map = {
    protein: '蛋白质',
    carbs: '碳水',
    fat: '脂肪',
    fiber: '膳食纤维',
    sugar: '糖分',
    vitamin_c: '维C',
  }
  return map[key] || key
}

// 营养素颜色
export function getNutrientColor(key) {
  const map = {
    protein: '#E91E63',
    carbs: '#FF9800',
    fat: '#9C27B0',
    fiber: '#4CAF50',
    sugar: '#F44336',
    vitamin_c: '#FFC107',
  }
  return map[key] || '#607D8B'
}
