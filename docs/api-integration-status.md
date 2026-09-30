# 前后端接口联调状态

更新日期：2026-09-11

## 第一轮已接通

| 页面 | 接口 | 状态 |
|---|---|---|
| 登录 | `POST /api/v1/auth/wx-login` | 已统一 `token`、`user_id`、`has_profile` 和 `profile` |
| 登录 | `POST /api/v1/agreements/accept` | 已接通隐私协议确认 |
| 个人资料 | `GET/PUT /api/v1/user/profile` | 已改为昵称和头像 URL，不再提交健康字段 |
| 账号注销 | `DELETE /api/v1/user/account` | 已支持永久清理账号、关联业务数据和上传文件；旧 Token 随即失效 |
| 首页 | `GET /api/v1/experiences` | 已读取当前城市公开体验，不再展示虚构评分 |
| 搜索 | `GET /api/v1/places/search` | 已支持城市、关键词、加载、空数据、错误重试和体验数量 |
| 搜索 | `GET /api/v1/places/:id/experiences` | 已支持从店铺进入第一条公开体验 |
| 地点提交 | `POST /api/v1/places` | 已接入微信地图选点，提交店名、地址和 GCJ-02 坐标；不再允许纯文本地点直接生成 |
| 吃喝攻略 | `GET /api/v1/routes/:id` | 已支持有序站点、逐站打开微信地图和复制完整路线 |
| 体验详情 | `GET /api/v1/experiences/:id` | 已接入地点、作者摘要、价格、等位、标签、正文和核验状态 |
| 记录 | `GET /api/v1/tags` | 已读取后端标签字典 |
| 记录 | `POST /api/v1/speech/transcribe` | 已接入百炼短音频转写，识别结果自动填入体验感受 |
| 记录 | `POST/PATCH /api/v1/records` | 已支持创建和版本化保存草稿 |
| 记录 | `POST /api/v1/records/:id/submit-public` | 已支持提交公开审核 |
| 审核状态 | `GET /api/v1/records/:id/review-status` | 已按真实状态展示 |
| 我的记录 | `GET /api/v1/me/records` | 记录页展示最近 3 条；“我的记录”展示待审核、审核通过、审核不通过 |
| 足迹 | `GET /api/v1/me/records` | 已接入本人记录列表 |
| 足迹 | `GET /api/v1/me/footprints/summary` | 已接入地点、商圈和本月记录统计 |
| 足迹 | `GET /api/v1/me/footprints/map` | 已统一地图点位字段 |
| 足迹海报 | 暂未开放 | 入口保留，点击提示“功能正在开发中”；不展示城市、店名或日期 |
| 我的 | `GET /api/v1/me/contribution` | 已统一 `account` 和 `trust` 结构 |
| 我的 | `GET /api/v1/me/badges` | 已统一徽章列表结构 |

## 第二轮已接通

- 图片和消费凭证上传：使用受 JWT 保护的 multipart 接口；凭证存储目录不公开。
- 发布者核验：使用微信手机号授权 code，手机号加密保存，仅返回核验状态。
- 文本内容安全：提交公开前调用微信内容安全接口，并建立人工审核任务。
- 收藏、有帮助、信息过时、举报与申诉：小程序页面和用户侧接口已接通。
- 路线收藏、开始路线、站点完成进度：已接通。
- 私人营养记录：通过独立 `/nutrition/*` 接口保存、汇总、筛选和删除，不进入餐厅评价。
- 隐私政策、用户服务协议和社区内容规范共用版本 `2.0`，登录前均可查看并在登录后记录同意。
- 足迹海报：使用城市主题和本人足迹在本地 Canvas 生成，不上传用户足迹。

## 仍需外部配置

- 图片内容安全、脱敏和审核结果需要审核运营流程更新对应状态；只有检测、脱敏均通过的图片才会出现在公开详情。
- 生产环境需填写合法 HTTPS `PUBLIC_BASE_URL`、微信内容安全权限以及 32 字节数据加密密钥。
- 营养模块当前提供独立的手动记录闭环；图片/文字 AI 营养估算仍需单独接入模型供应商。

## 本地联调前提

1. 执行 `server/migrations/000001_core.sql`。
2. 配置 `DATABASE_DSN`、长度至少 32 位的 `JWT_SECRET`、`bailian.api_key`，以及真实微信环境需要的 `WECHAT_APP_ID`、`WECHAT_APP_SECRET`。
3. 启动后端：`cd server && go run ./cmd/api`。
4. 小程序开发工具使用 `VITE_BASE_URL` 指向后端；真机调试不能使用 `localhost`，需要配置 HTTPS 合法域名或局域网可访问地址。

本轮已使用隔离的临时 PostgreSQL 完成迁移和真实 HTTP 冒烟测试，覆盖登录、资料、地点搜索、草稿创建、提交审核、审核状态、足迹、贡献、公开发现流和公开详情。临时 API、测试 Token 与数据库容器均已在验证后清理；实际开发数据库仍需按上述步骤配置。
