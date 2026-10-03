# 城市餐饮体验：数据库表结构草案

> 版本：0.1
> 范围：私人足迹、公开餐厅评论、私人饮食记录；仅评论先审后公开
> 数据库：PostgreSQL + GORM

## 1. 结论

当前启动实现：`main` 显式调用 `internal/model/pgsql.Init()`；连接自检通过后，同文件中的 `AutoMigrate` 在事务中补齐业务表、索引和历史兼容数据。失败时回滚并停止启动；数据库本身需提前创建。新增模型需要加入该函数的注册列表。

迁移完成后，`internal/model/pgsql/seed.go` 的 `Seed` 会幂等生成城市、初始地点、海报主题、体验标签、城市路线和功能开关，不覆盖已有运营数据。后续结构与基础数据调整统一使用 Go 代码，不再增加 SQL migration 文件。

- 现有 `food_records` 继续作为独立的私人营养工具使用；转盘、旧积分和旧成就只做数据归档，不再作为新版小程序入口。
- 新版餐饮体验使用独立表，不把 `food_records` 转换为公开点评。
- `nutrition_records.advice` 保存营养识别时生成的本餐饮食建议，仅本人可见，不用于公开评价或医疗判断。
- 公开内容必须保留版本、检测、人工审核、举报和处置记录。
- 消费凭证和公开图片分开存储；凭证默认仅审核人员可访问。

建议新增 **25 张表**：MVP/P0 20 张，贡献与徽章 5 张。第二阶段可以先实现其中标为 `MVP` 的 15 张，再补齐管理与激励能力。

## 2. 可复用的现有表

| 现有表 | 使用方式 | 是否需要调整 |
|---|---|---|
| `nutrilens_users` | 账号主体、昵称、头像 | 增加账号状态字段，健康字段停止新增收集 |
| `privacy_agreements` | 记录新版隐私政策、用户协议、社区公约同意版本 | 建议增加 `ip_hash`、`client_version` |
| `feedback` | 普通产品意见反馈 | 不用于内容举报或商家投诉 |
| `notify_settings` / `notify_logs` | 审核结果和申诉进度通知 | 增加新版通知类型 |
| `food_records`、`dishes` | 私人餐食营养识别 | 独立使用；可复制菜品草稿，但不关联公开状态和评价结论 |
| `user_scores`、`score_logs` | 旧版积分 | 不转换为新版可信度或贡献分 |
| `achievements`、`user_achievements` | 旧版成就 | 与新版徽章隔离 |

## 3. 核心业务表

### 3.0 `cities`（MVP）

小程序支持城市与首页城市视觉配置的唯一数据源。

| 字段 | 类型 | 说明 |
|---|---|---|
| `code` | varchar(16) PK | 行政区划编码，也是稳定的接口标识 |
| `name` | varchar(64) | 城市展示名称 |
| `description` | varchar(240) | 首页短描述；接口字段名为 `desc`，避免数据库直接使用排序关键字 `DESC` |
| `image` | varchar(1024) | 城市主图地址，可为小程序静态资源路径或 HTTPS 地址 |
| `enabled` | boolean | 是否对小程序开放 |
| `is_default` | boolean | 是否为首次使用的默认城市，全表最多一条为 true |
| `sort_order` | integer | 运营展示顺序 |
| `created_at` / `updated_at` | timestamptz | 创建与更新时间 |

索引：`(enabled, sort_order, code)`；`is_default=true` 使用部分唯一索引。暂不加入省份、拼音、经纬度和时区字段：当前页面没有消费这些数据，需要定位、按拼音检索或跨时区运营时再按明确场景扩展。

### 3.0.1 `city_poster_themes`（MVP）

足迹页非地理食光海报的主题数据源，与城市运营配置一对一；新增城市主题不需要发布小程序新版本。

| 字段 | 类型 | 说明 |
|---|---|---|
| `city_code` | varchar(16) PK/FK | 对应 `cities.code`，删除城市时级联清理 |
| `version` | integer | 主题数据版本，用于客户端缓存刷新 |
| `title` / `subtitle` | varchar | 城市海报标题与生活化副文案 |
| `background_image` | varchar(1024) | 不含地理信息的艺术背景资源 |
| `accent_color` / `secondary_color` | varchar(16) | Canvas 文字、节点和连线配色 |
| `motifs` | jsonb | 城市生活意象词，不得存储道路、边界或坐标 |
| `created_at` / `updated_at` | timestamptz | 创建与更新时间 |

用户店名、到店日期和统计只在小程序本地 Canvas 合成，主题表不接收用户隐私数据。海报节点按时间排序生成艺术星座，不表达真实地理关系。

### 3.1 `places`（MVP）

餐厅、小吃摊、夜市档口、饮品店等地点主表。

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | bigint PK | 地点 ID |
| `name` | varchar(160) | 店名 |
| `category` | varchar(32) | restaurant/stall/night_market/drink/other |
| `city_code` | varchar(16) | 城市编码 |
| `district` | varchar(64) | 行政区 |
| `business_area` | varchar(64) | 商圈 |
| `address` | varchar(300) | 门店地址 |
| `description` | text | 运营维护的客观店铺简介，默认空字符串，不从用户评论复制 |
| `longitude` / `latitude` | decimal(10,7) | POI 坐标 |
| `poi_provider` | varchar(32) | 地图服务商 |
| `poi_id` | varchar(128) | 第三方 POI ID |
| `status` | varchar(24) | pending/active/closed/merged/rejected |
| `merged_into_id` | bigint nullable | 被合并到的地点 |
| `created_by` | bigint nullable | 用户补充地点时记录提交人 |

用户补充地点必须来自地图选点并具有非零 GCJ-02 坐标；`poi_provider=wechat` 表示位置由微信原生选点界面确认。没有稳定 POI ID 时，使用“同城 + 标准化名称 + 80 米距离”辅助去重。

### 3.1.1 `city_route_stops`

吃喝攻略的有序地点表。`route_id` 关联 `city_routes`，`place_id` 关联规范化地点，`sort_order` 表示建议访问顺序，`note` 保存该站的简短攻略说明。地址和坐标只保存在 `places`，地点修正后所有攻略自动生效。路线不存储地图折线或实时算路结果。
| `created_at` / `updated_at` | timestamptz | 时间戳 |

索引：`(city_code, name)`、`(poi_provider, poi_id)` 唯一索引、经纬度空间索引、`merged_into_id`。

### 3.2 `place_suggestions`（MVP）

用户找不到地点时提交的新地点或纠错请求。

核心字段：`id`、`user_id`、`suggestion_type`、`place_id`、`name`、`city_code`、`address`、`longitude`、`latitude`、`reason`、`status`、`reviewer_id`、`review_note`、时间戳。

### 3.3 `visit_records`（MVP）

用户本人的私人到店足迹。它不参与审核、公开发现流或餐厅统计；保留旧发布字段仅用于平滑迁移，服务写入固定为 `visibility=private`、`publish_status=draft`。

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | bigint PK | 记录 ID |
| `user_id` | bigint FK | 发布者 |
| `place_id` | bigint FK | 具体地点 |
| `visibility` | varchar(16) | private/public |
| `publish_status` | varchar(32) | draft/detecting/pending_review/published/rejected/reported/hidden/deleted |
| `risk_level` | varchar(16) | low/medium/high/critical |
| `current_version_id` | bigint nullable | 当前展示版本 |
| `submitted_at` | timestamptz nullable | 提交公开时间 |
| `published_at` | timestamptz nullable | 审核通过时间 |
| `last_confirmed_at` | timestamptz nullable | 用户确认信息仍有效时间 |
| `helpful_count` | int | 冗余统计 |
| `outdated_count` | int | 冗余统计 |
| `report_count` | int | 冗余统计 |
| `created_at` / `updated_at` / `deleted_at` | timestamptz | 时间戳和软删除 |

关键索引：`(user_id, created_at desc)`、`(place_id, publish_status, published_at desc)`、`(visibility, publish_status, published_at desc)`。

### 3.4 `visit_record_versions`（MVP）

每次保存或修改足迹都生成新版本。已创建的餐厅评论持有独立快照，之后修改足迹不会静默改变已提交评论。

核心字段：

- `id`、`record_id`、`version_no`、`editor_user_id`；
- `visit_date`、`consumer_type`、`conclusion`；
- `price_min`、`price_max`、`average_cost`、`wait_minutes`、`meal_period`；
- `dishes` jsonb、`content` text、`visibility`；
- `change_summary`、`created_at`。

约束：`UNIQUE(record_id, version_no)`；正文长度由服务端校验 20～500 字。

### 3.5 `restaurant_reviews` 与评论快照（MVP）

公开餐厅评论使用四张独立表：

- `restaurant_reviews`：评论主记录，关联 `user_id`、`place_id` 和来源 `visit_record_id`；每次足迹最多一条评论，状态为 `pending/published/rejected`。
- `restaurant_review_versions`：评论内容快照，保存到店日期、结论、消费、排队、菜品和正文。
- `restaurant_review_tag_links`：评论版本与标签的关系。
- `restaurant_review_media`：评论与已上传媒体的关系。

评论 ID 与来源足迹 ID 在兼容期保持一致，使已有体验详情、投票、举报和申诉外键能够平滑迁移。一个用户可以通过多次到店足迹对同一餐厅提交多条评论。发现流、餐厅平均值和“我的评论”只能查询这些评论表，不能读取饮食记录或私人足迹。

### 3.6 `tags`（MVP）

标签字典。核心字段：`id`、`code` 唯一、`name`、`group_name`、`enabled`、`sort_order`。

首批数据：`taste` 口味、`price` 价格、`service` 服务、`queue` 排队、`hygiene_observation` 卫生观感、`promotion_mismatch` 宣传不符、`other` 其他。

### 3.7 `visit_record_tag_links`（MVP）

字段：`record_version_id`、`tag_id`，联合主键。标签绑定具体版本，避免编辑后丢失历史。

### 3.8 `record_media`（MVP）

公开内容图片。

核心字段：`id`、`record_id`、`record_version_id`、`object_key`、`public_url`、`media_type`、`width`、`height`、`sort_order`、`safety_status`、`desensitize_status`、`processed_object_key`、`face_detected`、`created_at`。

公开页面只能返回已检测、已脱敏的 `processed_object_key`。

### 3.9 `record_evidences`（MVP）

消费凭证，仅审核可见。

核心字段：`id`、`record_id`、`user_id`、`evidence_type`、`original_object_key`、`masked_object_key`、`verify_status`、`verified_by`、`verified_at`、`retention_until`、`created_at`。

禁止在普通详情接口中返回原始路径。

### 3.10 `nutrition_records`（MVP）

私人饮食记录，仅关联 `user_id`，保存餐次、进食时间、食物明细、营养估算和建议。该表没有地点、可见性、审核或社交互动字段；数据只出现在饮食记录页面，不进入足迹、我的评论、城市最近足迹和餐厅平均统计。

## 4. 用户互动表

### 4.1 `place_favorites`（MVP）

字段：`user_id`、`place_id`、`created_at`；联合唯一索引 `(user_id, place_id)`。

### 4.2 `helpful_votes`（MVP）

字段：`user_id`、`record_id`、`created_at`；联合唯一索引 `(user_id, record_id)`。不能给自己的记录投票。

### 4.3 `outdated_signals`（MVP）

字段：`id`、`user_id`、`record_id`、`reason`、`status`、`handled_at`、`created_at`；同一用户同一记录只能保留一个未处理信号。

### 4.4 `content_reports`（MVP/P0）

内容举报工单。

物理表使用 `content_reports`，避免与旧系统中其他类型的 `reports` 表发生命名冲突；接口资源路径仍使用 `/reports`。

核心字段：`id`、`reporter_user_id`、`target_type`、`target_id`、`reason_code`、`description`、`status`、`priority`、`assigned_to`、`result`、`resolved_at`、`created_at`、`updated_at`。

`reason_code`：false_experience/defamation/privacy_portrait/ad/malicious_competitor/outdated/other。

## 5. 身份、审核与申诉表

### 5.1 `publisher_verifications`（P0）

公开发布者身份核验。

核心字段：`user_id` 唯一、`phone_encrypted`、`phone_hash`、`verification_channel`、`verified_at`、`status`、`provider_reference`、时间戳。

手机号必须加密保存，日志和普通接口不得返回完整值。

### 5.2 `content_safety_checks`（MVP/P0）

每个版本、每张图片的机器检测结果。

核心字段：`id`、`target_type`、`target_id`、`provider`、`check_type`、`result`、`risk_labels` jsonb、`raw_response_ref`、`checked_at`。

### 5.3 `moderation_tasks`（MVP/P0）

人工审核队列。

核心字段：`id`、`task_type`、`target_type`、`target_id`、`record_version_id`、`priority`、`risk_labels` jsonb、`status`、`assignee_id`、`due_at`、`created_at`、`updated_at`。

### 5.4 `moderation_actions`（P0）

所有审核和处置动作的不可变日志。

核心字段：`id`、`task_id`、`operator_id`、`target_type`、`target_id`、`action`、`before_status`、`after_status`、`reason_code`、`note`、`metadata` jsonb、`created_at`。

动作包括 approve/reject/request_changes/limit/hide/restore/delete/ban/unban/mask。

### 5.5 `appeal_cases`（P0）

用户申诉、权利人投诉和商家投诉统一案件主表。

核心字段：`id`、`case_type`、`record_id`、`complainant_type`、`complainant_user_id`、`merchant_claim_id`、`contact_encrypted`、`claim_text`、`status`、`decision`、`assigned_to`、`response_due_at`、`closed_at`、时间戳。

### 5.6 `case_materials`（P0）

字段：`id`、`case_id`、`submitted_by_type`、`submitted_by_id`、`material_type`、`object_key`、`text_content`、`access_level`、`created_at`。

用于保存商家主体材料、消费证明、不侵权声明和补充说明。

### 5.7 `merchant_claims`（P1）

商家认领与主体核验。

核心字段：`id`、`place_id`、`entity_name`、`credit_code_encrypted`、`contact_name_encrypted`、`contact_phone_encrypted`、`relation_type`、`status`、`verified_by`、`verified_at`、时间戳。

## 6. 可信度与贡献体系

这些表不进入第一批接口开发，但数据结构应与旧积分隔离。

### 6.1 `user_trust_profiles`

字段：`user_id` 唯一、`trust_level`、`account_age_score`、`verified_record_count`、`violation_count`、`daily_publish_limit`、`risk_flags` jsonb、`updated_at`。

### 6.2 `contribution_accounts`

字段：`user_id` 唯一、`total_points`、`month_points`、`updated_at`。

### 6.3 `contribution_logs`

字段：`id`、`user_id`、`action_code`、`points`、`record_id`、`idempotency_key` 唯一、`description`、`created_at`。

### 6.4 `badges`

字段：`id`、`code` 唯一、`name`、`description`、`icon_url`、`condition_type`、`condition_value` jsonb、`enabled`、时间戳。

### 6.5 `user_badges`

字段：`user_id`、`badge_id`、`unlocked_at`；联合唯一索引 `(user_id, badge_id)`。

## 7. 状态枚举

### 评论状态

```text
pending -> published
pending -> rejected
```

私人足迹固定为 `private/draft`，不进入状态机。首版所有评论必须经过 `pending`，不存在低风险自动直发。

### 审核任务状态

```text
pending -> assigned -> processing -> resolved
                         \-> cancelled
```

### 申诉案件状态

```text
submitted -> validating -> notice_sent -> waiting_response
          -> reviewing -> decided -> closed
```

## 8. 实施顺序

1. 第一批：`places`、`place_suggestions`、私人足迹表、餐厅评论表、`nutrition_records`、标签、媒体和凭证表。
2. 第二批：`publisher_verifications`、`content_safety_checks`、`moderation_tasks`、`moderation_actions`。
3. 第三批：`place_favorites`、`helpful_votes`、`outdated_signals`、`content_reports`、`appeal_cases`、`case_materials`。
4. 第四批：商家认领、用户可信度、贡献分与新版徽章。

生产环境与开发环境使用同一套 `pgsql.AutoMigrate` 和 `pgsql.Seed` 启动流程。


## Redis 登录会话

关系型模型、连接池、自动迁移和种子数据集中在 `server/internal/model/pgsql`。Redis 操作位于 `server/internal/cache/redis`；两者都由 `main` 显式控制生命周期。

会话键为 `v1:session:{<session-id>}`，类型 Hash，字段为 `user_id`、`access_hash`、`refresh_hash`；后两者为 Token 的 SHA-256 摘要，不存明文凭据。键 TTL 默认 7200 秒，与两种 JWT 的有效期一致。刷新通过 Lua 比较旧 refresh 摘要、替换两个摘要并重设 TTL，只有一个并发请求能成功；过期或丢失的键拒绝鉴权。账号注销后，账号状态检查立即拒绝该账号的所有会话，残留 Redis 键在 TTL 到期后回收。

## 邮箱登录增量字段与 Redis 验证码

`nutrilens_users.email` 为可空 `VARCHAR(254)`，唯一索引 `idx_nutrilens_users_email`。所有邮箱登录入口统一以小写地址存储和查询；微信用户邮箱为 `NULL`，不统一填空字符串。邮箱用户的 `open_id` 使用 `email_<uuid>`，原微信 OpenID 和全部已有数据保留。增量迁移见 `server/migrations/000010_user_email.sql`，与启动 `AutoMigrate` 的模型定义一致，不执行种子导入。回退应用可保留字段和索引；移除字段将丢失邮箱资料，因此不提供自动删除回滚。

验证码键为 `v1:{email-auth}:code:<email-hmac>`，Hash 字段为 `id`（随机发送批次）、`digest`（验证码 HMAC）、`attempts`、`ready`，TTL 为 300 秒。`ready=0` 不允许登录；SMTP 接受后变为 1，失败时按批次比对后删除。相关键 `cooldown` 为 60 秒、`send-email`/`send-ip` 为 3600 秒、`login-ip` 为 60 秒。验证码与身份键均不保存原始邮箱、IP 或验证码；HMAC 使用 JWT 密钥及独立用途域，密钥变更会使未消费验证码失效。

发送限流、校验计数及一次性消费通过 Redis Lua 完成。所有邮箱认证键使用同一 Cluster hash tag 保证多键原子执行；业务用户与附属信用/贡献数据在 PostgreSQL 事务中创建，并通过邮箱唯一索引避免并发重复注册。
