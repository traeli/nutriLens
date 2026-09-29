export const diningImages = {
  shop: '/static/dining/noodle-shop.jpg',
  noodles: '/static/dining/noodle-detail.jpg',
  cafe: '/static/dining/corner-cafe.jpg',
}

export const featuredExperience = {
  id: 'exp-001',
  place: '梧桐面馆',
  area: '静安寺',
  city: '上海',
  visitMonth: '2026年8月到店',
  conclusion: '推荐',
  author: '林初',
  completeness: '信息完整度高',
  average: 68,
  wait: 12,
  meal: '晚餐',
  content: '面条筋道，浇头现炒。周五晚上等位约十二分钟，店内座位不多，建议错峰来。',
  tags: ['口味稳定', '明码标价', '排队'],
  helpful: 128,
  verified: true,
  image: diningImages.noodles,
}

export const places = [
  { id: 'p1', name: '梧桐面馆', category: '面馆', area: '静安寺', distance: '1.2km', count: 128, image: diningImages.shop },
  { id: 'p2', name: '梧桐里小馆', category: '本帮菜', area: '南京西路', distance: '1.8km', count: 36, image: diningImages.noodles },
  { id: 'p3', name: '梧桐咖啡', category: '咖啡', area: '巨鹿路', distance: '2.3km', count: 12, image: diningImages.cafe },
]

export const recentExperiences = [
  { id: 'exp-002', place: '山野小馆', area: '新天地', excerpt: '两个人点三道菜刚好，时令蔬菜比招牌菜更有记忆点。', author: '阿圆', time: '昨天更新', image: diningImages.noodles },
  { id: 'exp-003', place: '街角咖啡', area: '淮海中路', excerpt: '下午四点靠窗位置最安静，适合一个人带本书坐一会儿。', author: '林初', time: '3 天前', image: diningImages.cafe },
  { id: 'exp-004', place: '阿婆馄饨', area: '曹家渡', excerpt: '小碗也有十二只，荠菜肉馅清爽，附近居民来得很多。', author: '木木', time: '本周更新', image: diningImages.shop },
]

export const cityGuides = [
  { id: 'guide-01', number: '01', title: '沿着武康路，吃一顿不赶时间的午饭', meta: '徐汇 · 半日路线 · 4 站', tone: 'brick' },
  { id: 'guide-02', number: '02', title: '下班后还亮着灯的社区小馆', meta: '静安 · 晚餐 · 6 家', tone: 'green' },
  { id: 'guide-03', number: '03', title: '从一碗面开始认识老城厢', meta: '黄浦 · 步行路线 · 5 站', tone: 'blue' },
]

export const collectionCampaigns = [
  { id: 'campaign-01', kicker: '小店线索簿', title: '你家巷口那间，外地朋友很少知道的小店', count: 128, deadline: '长期征集' },
  { id: 'campaign-02', kicker: '长期征集', title: '把一间认真做饭的小店，留在城市地图上', count: 76, deadline: '持续开放' },
]

export const featuredAlleyRoute = {
  id: 'route-caojiadu-breakfast',
  district: '普陀 × 静安',
  title: '从曹家渡菜场出发，吃一顿上海人的早饭',
  description: '不追着名店跑。沿着菜场和居民区慢慢走，四小份刚好留住每一种味道。',
  duration: '2.5 小时',
  distance: '步行 2.8 km',
  budget: '约 ¥55',
  verifiedAt: '9月8日走访',
  stops: [
    { time: '07:30', name: '先喝一碗咸豆浆', note: '趁油条刚出锅' },
    { time: '08:10', name: '菜场边吃葱油饼', note: '一张两人分' },
    { time: '09:00', name: '拐进弄堂吃小馄饨', note: '留一点肚子' },
    { time: '09:40', name: '老点心铺带一份回去', note: '路线收尾' },
  ],
}

export const hiddenFoodClues = [
  {
    id: 'exp-004',
    place: '阿婆馄饨',
    area: '曹家渡 · 菜场后门',
    story: '没有套餐，也不催着翻台。附近居民端着搪瓷碗来，一小锅一小锅地下。',
    order: '先点：荠菜肉馄饨',
    verifiedAt: '9月8日核验',
    image: diningImages.shop,
  },
  {
    id: 'exp-002',
    place: '山野小馆',
    area: '新天地往南 · 老弄堂里',
    story: '门脸很窄，菜单跟着当天买到的菜变。来这里更适合问一句“今天什么新鲜”。',
    order: '先问：今日时蔬',
    verifiedAt: '9月5日核验',
    image: diningImages.noodles,
  },
]

export const myRecords = [
  { id: 'r1', place: '梧桐面馆', date: '8月18日', area: '静安区 · 乌鲁木齐中路', conclusion: '推荐', status: '已公开', image: diningImages.noodles },
  { id: 'r2', place: '山野小馆', date: '8月16日', area: '新天地', conclusion: '一般', status: '仅自己可见', image: diningImages.shop },
  { id: 'r3', place: '街角咖啡', date: '8月12日', area: '淮海中路', conclusion: '推荐', status: '审核中', image: diningImages.cafe },
]
