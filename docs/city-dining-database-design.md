# 城市餐饮体验：数据库表结构草案

> 版本：0.1
> 范围：个人主体、具体餐饮地点、全量先审后公开
> 数据库：PostgreSQL + GORM

## 1. 结论

当前启动实现：数据库连接成功后，`cmd/api/main.go` 调用 `internal/model/pgsql.AutoMigrate`，在事务中根据已注册的 `internal/model/pgsql` 模型补齐业务表，并补齐协议去重、默认城市、地点 POI、路线、记录幂等键、记录版本、审核任务和过时反馈的唯一索引。失败时回滚本次迁移并停止启动；数据库本身需提前创建。新增模型需要加入 `migrate.go` 的注册列表。

自动迁移不清空数据、不删除旧表或旧字段、不导入种子数据；GORM 可能调整已有字段类型和约束。它不完全替代 SQL：未映射的外键、检查约束、管理字段和管理表仍由显式 SQL 管理。完整 SQL 初始化应在首次自动建表前执行；之后补齐已有表的约束应使用增量迁移。城市、标签、路线、店铺等初始数据仍需单独导入，详见 `server/README.md`。

- 现有 `food_records` 继续作为独立的私人营养工具使用；转盘、旧积分和旧成就只做数据归档，不再作为新版小程序入口。
- 新版餐饮体验使用独立表，不把 `food_records` 转换为公开点评。
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

一条到店体验的稳定主记录，保存所有者、当前版本和状态。

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

每次保存、修改、驳回后重提都生成新版本，用于审计和争议固定。

核心字段：

- `id`、`record_id`、`version_no`、`editor_user_id`；
- `visit_date`、`consumer_type`、`conclusion`；
- `price_min`、`price_max`、`average_cost`、`wait_minutes`、`meal_period`；
- `dishes` jsonb、`content` text、`visibility`；
- `change_summary`、`created_at`。

约束：`UNIQUE(record_id, version_no)`；正文长度由服务端校验 20～500 字。

### 3.5 `tags`（MVP）

标签字典。核心字段：`id`、`code` 唯一、`name`、`group_name`、`enabled`、`sort_order`。

首批数据：`taste` 口味、`price` 价格、`service` 服务、`queue` 排队、`hygiene_observation` 卫生观感、`promotion_mismatch` 宣传不符、`other` 其他。

### 3.6 `visit_record_tag_links`（MVP）

字段：`record_version_id`、`tag_id`，联合主键。标签绑定具体版本，避免编辑后丢失历史。

### 3.7 `record_media`（MVP）

公开内容图片。

核心字段：`id`、`record_id`、`record_version_id`、`object_key`、`public_url`、`media_type`、`width`、`height`、`sort_order`、`safety_status`、`desensitize_status`、`processed_object_key`、`face_detected`、`created_at`。

公开页面只能返回已检测、已脱敏的 `processed_object_key`。

### 3.8 `record_evidences`（MVP）

消费凭证，仅审核可见。

核心字段：`id`、`record_id`、`user_id`、`evidence_type`、`original_object_key`、`masked_object_key`、`verify_status`、`verified_by`、`verified_at`、`retention_until`、`created_at`。

禁止在普通详情接口中返回原始路径。

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

### 记录状态

```text
draft
  -> detecting
  -> pending_review
  -> published

pending_review -> rejected -> draft -> detecting
published -> reported -> limited/hidden -> published/deleted
```

首版所有 `visibility=public` 的记录都必须经过 `pending_review`，不存在低风险自动直发。

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

1. 第一批：`places`、`place_suggestions`、`visit_records`、`visit_record_versions`、`tags`、`visit_record_tag_links`、`record_media`、`record_evidences`。
2. 第二批：`publisher_verifications`、`content_safety_checks`、`moderation_tasks`、`moderation_actions`。
3. 第三批：`place_favorites`、`helpful_votes`、`outdated_signals`、`content_reports`、`appeal_cases`、`case_materials`。
4. 第四批：商家认领、用户可信度、贡献分与新版徽章。

生产环境建议使用显式 SQL migration，不继续只依赖启动时 `AutoMigrate`。


## Redis 登录会话

关系型模型和连接迁移逻辑集中在 `server/internal/model/pgsql`，本次目录调整不改变业务表字段。Redis 操作位于 `server/internal/model/redis`。

会话键为 `v1:session:{<session-id>}`，类型 Hash，字段为 `user_id`、`access_hash`、`refresh_hash`；后两者为 Token 的 SHA-256 摘要，不存明文凭据。键 TTL 默认 7200 秒，与两种 JWT 的有效期一致。刷新通过 Lua 比较旧 refresh 摘要、替换两个摘要并重设 TTL，只有一个并发请求能成功；过期或丢失的键拒绝鉴权。账号注销后，账号状态检查立即拒绝该账号的所有会话，残留 Redis 键在 TTL 到期后回收。
