# 城市餐饮体验：接口清单草案

> 版本：0.1
> 基础路径：`/api/v1`
> 原则：所有业务查询和写入必须登录，并通过 Redis 会话鉴权；私人足迹和饮食记录永不进入公开审核，只有餐厅评论审核通过后才会公开。

> 业务边界：`/records` 保存私人到店足迹，`/reviews` 保存公开餐厅评论，`/nutrition/*` 保存私人饮食记录。三类数据独立，不相互参与列表、统计或审核。

## 1. 结论

建议规划 **61 个接口**：

- 小程序公开浏览：8 个；
- 登录用户与发布流程：26 个；
- 商家/权利人投诉：4 个；
- 管理后台：15 个。

第二阶段第一批只需实现其中 **29 个 P0/MVP 接口**，即可跑通“找店 → 私人记录 → 提交公开 → 审核 → 公开查看 → 举报/申诉”的闭环。

## 2. 通用约定

### 鉴权

- 无访问 Token 的入口仅限登录和刷新；健康检查及公开静态资源不属于业务 API。
- `User`：微信登录 JWT + Redis 有效会话 + 可用账号。
- `Verified publisher`：JWT + 已完成手机号/身份核验。
- `Admin`：独立后台账号、角色权限与操作审计，不能复用普通用户 JWT。

### 分页

统一使用游标：`?cursor=xxx&limit=20`，响应返回 `next_cursor`、`has_more`。

### 幂等

创建记录、提交公开、审核决定、贡献积分等写操作支持 `Idempotency-Key` 请求头。

### 错误格式

```json
{
  "error": {
    "code": "RECORD_REVIEW_PENDING",
    "message": "该记录正在审核中",
    "request_id": "req_xxx",
    "details": {}
  }
}
```

## 3. 公开内容浏览接口（需要登录）

| 方法 | 路径 | 说明 | 阶段 |
|---|---|---|---|
| GET | `/cities` | 可用城市及首页展示配置 | MVP |
| GET | `/cities/:code/poster-theme` | 非地理城市食光海报的文案、配色和背景 | MVP |
| GET | `/places/search` | 按店名、小吃名、商圈、城市、距离搜索 | MVP |
| GET | `/places/:id` | 地点详情与体验统计 | MVP |
| GET | `/places/:id/experiences` | 某地点已公开体验 | MVP |
| GET | `/experiences` | 发现流；按城市、结论、标签、时间筛选 | MVP |
| GET | `/experiences/:id` | 公开体验详情 | MVP |
| GET | `/tags` | 可用结构化标签 | MVP |
| GET | `/discovery/summary` | 首页推荐位和最近真实体验聚合 | 可选优化 |

`GET /experiences` 只返回 `publish_status=published`，默认排序为“近期 + 信息完整度”，不提供纯负面榜单排序。

`GET /places/:id` 在地点基础信息外返回 `description`、`image_url`、`experience_count`、`recommend_count` 和 `average_cost`。店铺详情页在简介下方通过 `GET /places/:id/experiences` 分页展示全部公开真实评论；评论项继续跳转公开体验详情。

首页“每日推荐”以店铺为单位去重和聚合，返回明确的 `place_id`，点击进入店铺详情；“用户真实记录”返回明确的 `record_id` 和 `place_id`，点击进入对应体验详情。`GET /places/search?recommended=1` 按公开真实记录数优先展示推荐店铺。

`GET /cities` 只返回 `enabled=true` 的城市，按默认城市、运营排序和城市编码稳定排序。响应示例：

```json
{
  "items": [
    {
      "code": "310000",
      "name": "上海",
      "desc": "在巷子里 遇见生活",
      "image": "/static/dining/shanghai-city.jpg",
      "is_default": true
    }
  ]
}
```

`desc` 和 `image` 用于首页城市主图区域；`is_default` 用于首次进入且本地没有选择记录时确定默认城市。城市是否启用及展示顺序属于服务端运营配置，不由小程序写死。

`GET /cities/:code/poster-theme` 作为后续海报能力的预留接口。当前小程序仅保留海报入口并提示“功能正在开发中”，不生成海报，也不展示城市、店名或日期。

```json
{
  "city_code": "310000",
  "version": 1,
  "title": "上海食光星图",
  "subtitle": "梧桐影里，记下认真吃饭的夜晚",
  "background_image": "/static/posters/food-memory-night-v1.jpg",
  "accent_color": "#C7FF35",
  "secondary_color": "#F1E5C8",
  "motifs": ["梧桐", "弄堂", "夜色"]
}
```

## 4. 登录与发布者核验（4 个）

现有 `/auth/wx-login`、`/user/profile` 继续复用。

| 方法 | 路径 | 说明 | 阶段 |
|---|---|---|---|
| POST | `/publisher-verification/phone` | 使用微信手机号 code 完成发布者核验 | P0 |
| GET | `/publisher-verification/status` | 查询手机号核验状态，不作为记录提交资格判断 | P0 |
| GET | `/agreements/current` | 当前协议、隐私政策、社区公约版本 | P0 |
| POST | `/agreements/accept` | 接受指定版本协议 | P0；可改造现有 privacy 接口 |

## 5. 地点接口（3 个登录接口）

| 方法 | 路径 | 说明 | 阶段 |
|---|---|---|---|
| POST | `/place-suggestions` | 找不到地点时提交新地点 | MVP |
| POST | `/places/:id/corrections` | 提交地址、名称或停业纠错 | P1 |
| GET | `/me/place-suggestions` | 查看本人地点提交进度 | MVP |

POI 搜索由服务端代理地图服务，前端不直接持有第三方密钥。

## 6. 私人足迹、餐厅评论与饮食记录

| 方法 | 路径 | 说明 | 阶段 |
|---|---|---|---|
| POST | `/records` | 新建私人草稿 | MVP |
| GET | `/records/:id` | 查看本人的私人足迹 | MVP |
| PATCH | `/records/:id` | 修改本人的私人足迹 | MVP |
| DELETE | `/records/:id` | 软删除本人记录 | MVP |
| GET | `/me/records` | 本人足迹列表；筛选状态、日期、城市 | MVP |
| GET | `/me/footprints/summary` | 店铺数、商圈数、本月记录数 | MVP |
| GET | `/me/footprints/map` | 本人记录的地点聚合点位 | MVP |
| POST | `/records/:id/media/presign` | 获取公开图片上传凭证 | MVP；复用上传服务 |
| DELETE | `/records/:id/media/:media_id` | 删除草稿图片 | MVP |
| POST | `/records/:id/evidences/presign` | 获取私密凭证上传地址 | MVP/P0 |
| POST | `/visits/:id/review` | 从一次私人足迹创建一条餐厅评论 | P0 |
| GET | `/reviews/:id/status` | 查看评论审核状态 | MVP |
| GET | `/me/reviews` | 本人的评论列表 | MVP |

当前小程序采用服务端 multipart 上传：`POST /records/:id/media`（字段 `file`、`media_type`）和
`POST /records/:id/evidences`（字段 `file`、`evidence_type`）。服务端存储实现可在保持接口不变的情况下替换为对象存储。

### 6.1 短音频语音转写

`POST /speech/transcribe` 需要登录，用 `multipart/form-data` 上传字段名为 `audio` 的音频文件。当前限制为单文件不超过 7MB，支持 AAC、AMR、FLAC、MP3、MPEG、OGG、OPUS、WAV、WebM 和 WMA。服务端调用百炼 `qwen3-asr-flash`，小程序不接触百炼 API Key。7MB 上限为 Base64 编码后的体积预留空间，可确保不超过百炼同步接口的 10MB 输入限制。

语音服务未配置时返回 HTTP 503、错误码 `SPEECH_NOT_CONFIGURED`，提示“语音识别服务尚未配置，请联系管理员”；上游请求失败仍返回 503 / `SERVICE_UNAVAILABLE`，不向客户端暴露密钥或上游内部错误。


成功响应：

```json
{
  "text": "这家店的面很好吃，人均 80 元。"
}
```

空文件、超限或格式不支持返回 `400 INVALID_ARGUMENT`；百炼未配置、超时或识别失败返回 `503 SERVICE_UNAVAILABLE`。音频只在请求内存中用于当次转写，不在该接口中落库。

### 6.2 私人营养文字识别

`POST /nutrition/records` 需要登录。小程序提交饮食描述，由服务端调用独立配置的 OpenAI-compatible LLM（默认使用百炼 `qwen-plus`），校验结构化营养结果后保存并返回完整记录。LLM Key 只保存在服务端，不下发到小程序。

请求示例：

```json
{
  "meal_period": "lunch",
  "eaten_at": "2026-10-03T12:30:00+08:00",
  "description": "我吃了一个苹果",
  "foods": []
}
```

响应包含 `foods`、`calories`、`protein_grams`、`fat_grams`、`carbohydrate_grams` 和 `advice`，`source_type` 为 `llm_text`。`advice` 是针对本餐的简短均衡饮食建议，不提供疾病诊断、治疗或极端节食建议。识别成功后记录已经落库，客户端不得再次调用创建接口。未配置 LLM 时返回 `503 NUTRITION_ANALYSIS_NOT_CONFIGURED`；上游错误、空响应或结构校验失败返回 `503 SERVICE_UNAVAILABLE`，失败时不创建记录。营养数据为估算值，仅供私人日常记录参考。

`POST /records` 请求示例：

```json
{
  "place_name": "巷口面馆",
  "city_code": "shanghai",
  "visit_date": "2026-09-10",
  "consumer_type": "restaurant",
  "conclusion": "recommend",
  "tag_codes": ["taste", "queue"],
  "average_cost": 68,
  "wait_minutes": 12,
  "meal_period": "dinner",
  "dishes": ["葱油拌面"],
  "content": "面条筋道，浇头现炒，周五晚等位约十二分钟。",
  "visibility": "private"
}
```

`POST /records` 必须提交已经确认的 `place_id`。已有地点直接选择；未收录地点先通过微信原生地图选点，再调用 `POST /places` 提交名称、地址和 GCJ-02 坐标，最后使用返回的 `place_id` 创建记录。系统不再根据纯文本店名创建 `0,0` 坐标地点。

`POST /places` 的新地点请求至少包含 `name`、`city_code`、`address`、`longitude`、`latitude` 和 `poi_provider=wechat`。同城同名且相距 80 米以内的地点视为同一候选，避免重复提交；同名连锁门店可因坐标不同而分别存在。

## 6.1 吃喝攻略路线

- `GET /routes/:id` 返回攻略标题及有序站点。
- `GET /routes/nearby?city_code=310000&longitude=121.47&latitude=31.23&limit=4` 使用 GCJ-02 当前位置生成 3～4 个站点的临时路线，按直线距离由近到远排列；有至少 3 个具备公开记录的地点时，仅使用这些地点。
- 每个站点引用规范化 `places` 数据，包含名称、地址和 GCJ-02 坐标。
- 小程序只展示建议顺序和直线距离，不调用道路算路服务，也不保存用户当前位置。
- 单站通过 `wx.openLocation` 打开微信内置地图；整条路线支持复制为有序文本，供用户粘贴到高德、百度等地图。
- 不依赖高德/百度的私有 App Scheme 或多途经点调起参数，避免微信环境拦截和第三方协议变化。

当前小程序仍以一次按钮操作完成提交：先保存私人足迹，再以足迹 ID 调用 `POST /visits/:id/review`。评论服务复制当次足迹的当前版本、标签和图片作为独立快照，足迹本身始终保持 `private/draft`。同一足迹只生成一条评论；用户再次到店会形成新的足迹，因此可以对同一餐厅多次评论。

评论正文和标签均可为空，正文保留 500 字上限。提交只校验足迹所有权与账号状态；非空正文执行文本安全检测，随后创建 `pending` 评论及审核任务。评论审核状态为 `pending/published/rejected`。公开发现流、餐厅平均消费、推荐数和互动只读取 `published` 评论。旧 `/records/:id/submit-public` 与 `/records/:id/review-status` 暂保留一个兼容周期，但内部已转发到评论服务，不再修改足迹。

## 7. 收藏与内容反馈（7 个）

| 方法 | 路径 | 说明 | 阶段 |
|---|---|---|---|
| POST | `/places/:id/favorite` | 收藏地点，幂等 | MVP |
| DELETE | `/places/:id/favorite` | 取消收藏 | MVP |
| GET | `/me/favorite-places` | 我的收藏 | MVP |
| POST | `/experiences/:id/helpful` | 标记有帮助，幂等 | MVP |
| DELETE | `/experiences/:id/helpful` | 取消有帮助 | MVP |
| POST | `/experiences/:id/outdated` | 标记信息过时 | MVP |
| POST | `/reports` | 举报公开体验或媒体 | P0 |

举报请求必须包含 `target_type`、`target_id`、`reason_code`，说明和附件按原因选填。

## 8. 我的状态、申诉与贡献（7 个）

| 方法 | 路径 | 说明 | 阶段 |
|---|---|---|---|
| GET | `/me/reports` | 我的举报记录 | P0 |
| GET | `/me/reports/:id` | 举报处理详情 | P0 |
| POST | `/records/:id/appeals` | 对驳回、隐藏或删除提出申诉 | P0 |
| GET | `/me/appeals` | 我的申诉列表 | P0 |
| GET | `/me/appeals/:id` | 申诉进度 | P0 |
| GET | `/me/contribution` | 新版贡献分统计 | P1 |
| GET | `/me/badges` | 新版徽章 | P1 |

## 9. 商家与权利人投诉（4 个）

这组接口可以不要求注册为普通社区用户，但必须有验证码、限频、联系方式核验和案件查询凭证。

| 方法 | 路径 | 说明 | 阶段 |
|---|---|---|---|
| POST | `/rights-complaints` | 提交商家/权利人投诉 | P0 |
| POST | `/rights-complaints/:id/materials/presign` | 上传主体和侵权初步证据 | P0 |
| GET | `/rights-complaints/:id/status` | 使用案件号和查询凭证查看进度 | P0 |
| POST | `/merchant-claims` | 发起店铺认领 | P1 |

## 10. 管理后台接口（15 个）

管理后台建议使用 `/api/v1/admin` 前缀和独立 RBAC。

### 审核队列（5 个）

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/admin/moderation/tasks` | 按状态、风险、时间、店铺筛选任务 |
| GET | `/admin/moderation/tasks/:id` | 原文、版本、媒体、凭证和检测结果 |
| POST | `/admin/moderation/tasks/:id/assign` | 领取或分配审核任务 |
| POST | `/admin/moderation/tasks/:id/decision` | 通过、驳回、要求修改 |
| GET | `/admin/moderation/tasks/:id/actions` | 审核操作历史 |

### 内容与主体管理（5 个）

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/admin/search` | 检索帖子、图片、用户、店铺 |
| POST | `/admin/records/:id/actions` | 限流、隐藏、删除、恢复、打码 |
| POST | `/admin/users/:id/actions` | 限频、封禁、解封 |
| POST | `/admin/places/:id/merge` | 合并重复地点 |
| POST | `/admin/place-suggestions/:id/decision` | 处理新地点或纠错 |

### 举报和申诉（5 个）

| 方法 | 路径 | 说明 |
|---|---|---|
| GET | `/admin/cases` | 举报、权利投诉、用户申诉队列 |
| GET | `/admin/cases/:id` | 案件详情、固定版本和证据 |
| POST | `/admin/cases/:id/action` | 通知、限流、隐藏、恢复、删除等 |
| POST | `/admin/cases/:id/request-materials` | 要求一方补充材料 |
| POST | `/admin/cases/:id/decision` | 维持、修改、删除、恢复并结案 |

所有后台写接口都必须写入 `moderation_actions`，保存操作人、理由、前后状态、时间和请求 ID。

## 11. 第二阶段第一批接口

优先实现以下闭环：

1. `GET /cities`
2. `GET /cities/:code/poster-theme`
3. `GET /places/search`
4. `GET /places/:id`
5. `POST /place-suggestions`
6. `GET /experiences`
7. `GET /experiences/:id`
8. `GET /places/:id/experiences`
9. `GET /tags`
10. `POST /speech/transcribe`
11. `POST /records`
12. `GET /records/:id`
13. `PATCH /records/:id`
14. `DELETE /records/:id`
15. `GET /me/records`
16. `GET /me/footprints/summary`
17. `GET /me/footprints/map`
18. `POST /records/:id/media/presign`
19. `POST /records/:id/evidences/presign`
20. `POST /publisher-verification/phone`
21. `GET /publisher-verification/status`
22. `GET /agreements/current`
23. `POST /agreements/accept`
24. `POST /visits/:id/review`
25. `GET /reviews/:id/status`
26. `POST /reports`
27. `POST /records/:id/appeals`
28. `GET /admin/moderation/tasks`
29. `GET /admin/moderation/tasks/:id`
30. `POST /admin/moderation/tasks/:id/decision`

收藏、有帮助、信息过时、贡献分、徽章和商家认领可以在闭环稳定后接入。

## 12. 与现有接口的关系

- 继续复用：`POST /auth/wx-login`、`GET/PUT /user/profile`、上传签名底层能力、通知设置。
- 需要改造：协议版本接口、上传 `biz_type`、登录响应中的发布者核验状态。
- 独立保留：`/nutrition/*` 私人饮食记录能力，不与足迹或公开评论混用。
- 前端下线：`/wheel/*`、旧 `/score/*`、旧 `/achievement/*` 对应页面；服务端可在完成数据归档后另行移除。
- 不建议把新版接口继续放在 `FoodService` 中；新增 `PlaceService`、`RecordService`、`ModerationService`、`CaseService`、`ContributionService`。

P1 可新增无副作用的 `POST /record-assistant/extract`，只根据用户文字、菜品图片或小票返回到店记录字段建议。该接口不得直接创建记录、修改总体结论或触发公开发布。


## 登录会话刷新与城市参数补充

`POST /api/v1/auth/wx-login`（兼容 `/auth/login`）保留 `token`、`user_id`、`has_profile`、`profile`，新增 `refresh_token`、`expires_in`、`refresh_expires_in`。两个有效期字段以秒计，当前默认均为 7200。

`POST /api/v1/auth/refresh` 无需访问 Token，请求体为 `{"refresh_token":"<refresh-token>"}`。成功 200 返回 `token`、`refresh_token`、`expires_in`、`refresh_expires_in`；缺少凭据为 400，过期、重用、凭据类型错误或账号不可用为 401，Redis/账号检查不可用为 503。刷新原子替换当前会话内两个 Token，旧值不能继续使用。超过两小时未刷新，需要重新登录；小程序在到期前一分钟的业务请求中主动刷新。

`POST /places` 的 `city_code` 必须为已启用城市的非空编码（例如 `310000`），同时提交店名、详细地址、经纬度。记录页提交前从城市缓存取值；没有编码则等待 `/cities` 初始化，不允许发送空编码。`POST /records` 同步携带选定的 `city_code`，实际地点关联仍以 `place_id` 为准。
