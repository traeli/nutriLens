# 前后端接口联调状态

> 2026-11-12：小程序主导航已改为“首页 / 发现 / 我的”。发现页读取新的 `discovery_posts` 数据域，用户记录、足迹、营养记录、语音和地点投稿接口已停止注册。完整接口以 `docs/discovery-posts-design.md` 为准。

更新日期：2026-11-12

## 第一轮已接通

| 页面 | 接口 | 状态 |
|---|---|---|
| 登录 | `POST /api/v1/auth/wx-login` | 已统一 `token`、`user_id`、`has_profile` 和 `profile` |
| 登录 | `POST /api/v1/agreements/accept` | 已接通隐私协议确认 |
| 个人资料 | `GET/PUT /api/v1/user/profile` | 已改为昵称和头像 URL，不再提交健康字段 |
| 账号注销 | `DELETE /api/v1/user/account` | 已支持永久清理账号、关联业务数据和上传文件；旧 Token 随即失效 |
| 首页 | `GET /api/v1/experiences` | 只展示运营方统一维护的编辑食记，公开作者固定为“饭友记编辑” |
| 首页聚合 | `GET /api/v1/home/summary` | `scenario_filters` 返回城市/全局场景配置；`featured_route.stops` 返回 `recommended_food`、`food_category_code` 和 `icon_key`，客户端只负责本地图标渲染 |
| 搜索 | `GET /api/v1/places/search` | 已支持城市、关键词、加载、空数据、错误重试和体验数量 |
| 搜索 | `GET /api/v1/places/:id/experiences` | 已支持从店铺进入运营方编辑食记 |
| 地点提交 | `POST /api/v1/places` | 已接入微信地图选点，提交店名、地址和 GCJ-02 坐标；不再允许纯文本地点直接生成 |
| 吃喝攻略 | `GET /api/v1/routes/:id` | 已支持有序站点、逐站打开微信地图和复制完整路线 |
| 实时路线 | `GET /api/v1/routes/nearby` | 首页和详情页已透传 `budget`、`scene`；后端按位置、时间、场景、公开体验信号及站点多样性即时生成 1/2/3 站，不读取预设路线表 |
| 食记详情 | `GET /api/v1/experiences/:id` | 已接入运营方作者、价格、等位、标签、正文和“有帮助”互动 |
| 记录 | `GET /api/v1/tags` | 已读取后端标签字典 |
| 记录 | `POST /api/v1/speech/transcribe` | 已接入百炼短音频转写，识别结果自动填入体验感受 |
| 私人足迹 | `POST/PATCH /api/v1/records` | 已支持创建和版本化保存，不进入审核或公开统计 |
| 用户公开发布 | — | 个人主体版本已删除提交公开、评论审核、发布者核验和申诉接口 |
| 足迹 | `GET /api/v1/me/records` | 已接入本人记录列表 |
| 足迹 | `GET /api/v1/me/footprints/summary` | 已接入地点、商圈和本月记录统计 |
| 足迹 | `GET /api/v1/me/footprints/map` | 已统一地图点位字段 |
| 足迹海报 | 暂未开放 | 入口保留，点击提示“功能正在开发中”；不展示城市、店名或日期 |
| 我的 | `GET /api/v1/me/footprints/summary` | 只展示本人的记录、地点、商圈和本月数量 |

## 第二轮已接通

- 图片和消费凭证上传：使用受 JWT 保护的 multipart 接口；凭证存储目录不公开。
- 普通用户只保存私人到店记录，不再创建公开评论、进入审核或申请公开。
- 运营方编辑食记保留收藏、有帮助、信息过时和举报；“有帮助”只展示聚合数量，不展示用户身份或互动关系。
- 路线收藏、开始路线、站点完成进度：已接通。
- 私人营养记录：`POST /nutrition/records` 在只提交饮食描述时调用服务端 LLM 估算营养并保存；查询、汇总和删除继续通过独立 `/nutrition/*` 接口完成，不进入餐厅评价。
- 私人足迹只在本人记录和足迹页，运营方食记只在公开发现，饮食记录只在饮食页面。
- 隐私政策和用户服务协议使用版本 `2.1`，明确用户记录不提供公开发布能力。
- 足迹海报：使用城市主题和本人足迹在本地 Canvas 生成，不上传用户足迹。

## 仍需外部配置

- 运营方应只发布拥有使用权的图片，并持续维护公开食记内容。
- 生产环境需填写合法 HTTPS `PUBLIC_BASE_URL` 和必要的数据加密配置。
- 营养文字估算已接入独立的 OpenAI-compatible LLM 接口，默认使用百炼 `qwen-plus`；可通过 `LLM_API_KEY`、`LLM_BASE_URL` 和 `LLM_MODEL` 切换供应商或模型。图片营养识别仍需单独接入。

## 本地联调前提

1. 执行 `server/migrations/000001_core.sql`。
2. 配置 `DATABASE_DSN`、长度至少 32 位的 `JWT_SECRET`、`bailian.api_key`，以及真实微信环境需要的 `WECHAT_APP_ID`、`WECHAT_APP_SECRET`。
3. 启动后端：`cd server && go run ./cmd/api`。
4. 小程序开发工具使用 `VITE_BASE_URL` 指向后端；真机调试不能使用 `localhost`，需要配置 HTTPS 合法域名或局域网可访问地址。

本轮验证覆盖登录、资料、地点搜索、私人记录、足迹、运营方食记发现流和食记详情；普通用户公开提交相关路由应保持未注册状态。


### Redis 鉴权与记录城市参数

城市、首页摘要、公开路线、店铺和已发布体验接口允许游客读取；个人资料、私人数据和写入接口继续校验 Redis 会话。登录返回双 Token，`/auth/refresh` 默认有效期为两小时；小程序受保护请求和上传支持预刷新、并发合并及一次重试。首页不再自动申请定位，只有用户主动生成附近路线或选择门店时才触发对应授权。记录页创建 `/places` 前确保城市初始化，避免 `city_code` 为空，并向 `/records` 同步传递城市编码。

### 邮箱验证码登录

`POST /auth/getemailcode` 与 `POST /auth/email-login` 已完成 SMTP、Redis 验证码及账户业务。请求封装新增 `api.sendEmailCode(email)` 和 `api.emailLogin(email, emailCode)`，邮箱登录会保存双 Token；尚未增加邮箱登录页面。真实发信需要本地配置 SMTP 账号与授权码。
