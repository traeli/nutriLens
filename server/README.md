# 食迹簿后端

这是面向城市餐饮发现和真实到店记录的新后端。旧 NutriLens 营养、转盘、积分、成就、提醒和 Webhook 业务不在本项目中。

## 本地启动

```bash
go run ./cmd/api
```

默认从当前工作目录读取 `config/config.yaml`，也可以通过 `CONFIG_PATH` 指定其他文件。当前 Docker Compose 会将 `server/config/config.yaml` 只读挂载到容器内 `/app/config/config.yaml`，后端业务配置直接从该文件读取。部署前必须填写新的微信密钥、百炼 Key、JWT 密钥和数据加密密钥，且不得提交生产使用的真实密钥。

开发环境可以向 `/api/v1/auth/wx-login` 提交 `code=the code is a mock one` 创建本地测试账号。

## 数据库

兼容已有 SQL 表结构时，`User.OpenID`、`Tag.Code`、`Badge.Code`、`PublisherVerification.PhoneHash` 使用 GORM 的 `unique` 标签，对应原 SQL 的单字段 `UNIQUE` 约束。不要仅改成 `uniqueIndex`：GORM 会将其视作另一类定义，可能尝试按生成的名称删除原约束而导致启动失败。原先仅有独立唯一索引的数据库可能会新增唯一约束，旧索引不会在本次修改中自动清理。

启动时还会根据 `redis.dsn`（可由 `REDIS_DSN` 覆盖）创建 Redis 客户端并执行 PING；配置无效或连接失败会停止启动，正常退出时关闭连接池。本地 Go 进程连接 Docker Redis 可使用 `redis://127.0.0.1:6379/0`，同一 Compose 网络中的后端使用 `redis://redis:6379/0`。需要认证时支持 `redis://用户名:密码@主机:端口/数据库编号`，凭据中的特殊字符需进行 URL 编码；真实凭据不应提交到仓库。

`internal/model/redis.Client` 封装底层 `go-redis` 客户端和 `v1:` 键名前缀，并负责登录会话存储和原子刷新。可在项目根目录运行 `docker compose up -d redis` 启动本地 Redis。

后端连接数据库成功后，会调用 `pgsql.AutoMigrate`，根据 `internal/model/pgsql` 中注册的业务模型自动创建缺失的表、字段及索引；迁移失败则停止启动。重复启动不会清空已有数据，也不会自动删除旧表或旧字段，但 GORM 可能调整已有字段的类型或约束。数据库本身仍需预先创建，连接账号需要建表及修改结构的权限。

自动建表不等于导入初始数据：城市、标签、路线、店铺和功能开关等仍由以下 SQL 脚本初始化。首次部署需要这些内容时执行（已有库请先备份，部分种子脚本会更新运营配置）：

```bash
psql "$DATABASE_DSN" -f migrations/000001_core.sql
psql "$DATABASE_DSN" -f migrations/000002_cities.sql
psql "$DATABASE_DSN" -f migrations/000004_city_poster_themes.sql
psql "$DATABASE_DSN" -f migrations/000005_city_routes.sql
psql "$DATABASE_DSN" -f migrations/000006_place_seeds.sql
psql "$DATABASE_DSN" -f migrations/000007_city_route_stops.sql
psql "$DATABASE_DSN" -f migrations/000008_place_descriptions.sql
psql "$DATABASE_DSN" -f migrations/000009_governance_and_engagement.sql
```

迁移使用原有核心表名，包括 `nutrilens_users`、`places`、`visit_records` 和 `visit_record_versions`。`000002_cities.sql` 创建城市配置表；`000004_city_poster_themes.sql` 创建非地理足迹海报主题。已经停用的 `city_maps` 原型表仅为兼容历史开发数据库保留，服务端不再注册读取接口。后续新增城市海报只需写入 `city_poster_themes`，无需修改小程序渲染器。`000005_city_routes.sql` 创建首页城市路线表（轮播卡片与精选手绘路线）并写入各城市初始路线；`000007_city_route_stops.sql` 将路线站点规范化关联到 `places`，供攻略详情逐站打开地图。`000006_place_seeds.sql` 为各城市写入初始店铺目录（`status='active'`，同名地点已存在时跳过），供「找一家店」搜索冷启动使用。

自动迁移仅覆盖当前注册的业务模型及必要的唯一索引，不保证与 SQL 完全等价：SQL 中的外键、额外检查约束、管理字段，以及 `moderation_actions`、`case_materials` 管理表没有全部映射到模型。需要完整 SQL 结构时应在首次自动建表之前执行上述脚本；在自动建表之后运行 `CREATE TABLE IF NOT EXISTS` 不会补齐已有表中缺失的外键或管理字段，需要另行编写增量迁移。

迁移也不会清理旧业务遗留表。是否归档或删除旧营养业务数据，应在确认备份与保留期限后通过单独的数据迁移处理。

## 当前接口范围

到店记录提交已放宽：无需手机号核验，标签和正文选填（正文 0～500 字），成功提交后仍为 `pending_review` 并创建人工审核任务。非空正文继续调用文本安全检测；空正文的检测记录标为 `not_applicable`，不向微信发送空文本。登录、协议、所有权、账号状态、提交开关和每日额度检查仍保留。独立手机号核验接口继续兼容，但记录页面不再调用。

- 微信登录与个人资料
- 城市、地点搜索、公开体验和标签
- 登录用户的短音频语音转写
- 私人到店记录的创建、读取、版本化编辑与软删除
- 公开审核提交、发布者微信手机号核验、文本内容安全检查、我的记录、足迹汇总与地图点位
- 公开图片和私密消费凭证上传、收藏、有帮助、信息过时、举报与申诉
- 路线收藏、开始路线和完成站点；私人营养记录与汇总
- 贡献账户、可信等级和徽章

审核管理界面不由本服务提供；小程序侧只创建和查询审核、举报及申诉数据。快速记录支持用店名创建待确认地点；待确认地点不会直接进入公开推荐。旧 `/food`、`/wheel`、旧积分、成就、通知和 Webhook 路由不再注册，私人营养数据使用独立的 `/nutrition/*` 接口。

生产环境必须配置 `DATABASE_DSN`、`JWT_SECRET`、`WECHAT_APP_ID`、`WECHAT_APP_SECRET`、`DATA_ENCRYPTION_KEY`（32 字节密钥的 Base64）、`UPLOAD_DIR` 和 `PUBLIC_BASE_URL`。公开内容提交前会调用微信文本内容安全接口；调用失败时保持草稿，不会绕过检查。

账号注销使用 `DELETE /api/v1/user/account`。注销会永久清除用户账号、到店记录、营养记录、上传图片和消费凭证等关联数据，并使注销前签发的 JWT 失效。

## Docker Compose

```bash
# 先填写 server/config/config.yaml，并确保 database.dsn 使用 host=postgres。
cd ..
docker compose up -d postgres server
curl --fail http://127.0.0.1:8080/health
```

Docker 构建默认通过 `https://goproxy.cn` 下载 Go 模块。若部署环境需要其他代理，可以在构建时覆盖：

```bash
docker compose build --build-arg GOPROXY=https://your-go-proxy.example,direct server
```

服务仅绑定宿主机回环地址 `127.0.0.1:8080`，应由 HTTPS 反向代理对外提供服务。上传目录固定为容器内 `/app/uploads` 并挂载持久化卷。可选管理服务不会默认启动；确有需要时填写 Directus 独立密钥后使用 `--profile admin` 启动。
运行镜像包含 `tzdata`，以支持数据库 DSN 中的 `TimeZone=Asia/Shanghai`。

## 验证

```bash
gofmt -w ./cmd ./config ./internal
go test ./...
go build ./...
```


### 初始化与登录会话

PostgreSQL 连接、迁移和所有持久化业务结构位于 `internal/model/pgsql`；Redis 客户端、会话结构与读写位于同级 `internal/model/redis`。微信、百炼客户端保留在 `internal/platform`，请求 DTO 和业务服务仍归各自层。

`main` 创建实例后显式调用 `Init` 检查连接，失败即 panic（不输出连接凭据）。Go 的内置小写 `init()` 无法由 main 主动调用，故使用导出的 `Init`。PostgreSQL、Redis 使用 PING；微信预取服务端 access_token；已配置百炼 Key 时调用 `/models` 检查可达性及认证，不产生转写。未配置百炼 Key 时保留语音禁用行为；微信仅在允许 mock 且未配置凭据时跳过检查。正常退出关闭数据库和 Redis 连接池。

所有 `/api/v1` 业务接口必须使用 `Authorization: Bearer <token>`，登录 `/auth/wx-login`（兼容 `/auth/login`）和 `/auth/refresh` 除外。`/health` 和公开静态资源 `/uploads` 仍可直接访问。鉴权同时校验 JWT、Redis 中的 Token 摘要及账号状态；Redis 不可用返回 503，凭据过期/被替换返回 401。

登录返回 `token`、`refresh_token`、`expires_in`、`refresh_expires_in` 及原有用户信息。`POST /api/v1/auth/refresh` 请求体为 `{"refresh_token":"<refresh-token>"}`，返回新的双 Token。刷新通过 Redis Lua 原子替换，旧访问和刷新 Token 同时失效。Redis 仅存 SHA-256 摘要，默认 TTL 为 7200 秒，与 JWT 有效期一致，沿用 `jwt.expire_hours` / `JWT_EXPIRE_HOURS` 配置（默认 2）。每个登录会话独立，多端互不覆盖。

小程序在请求前发现不足一分钟到期时刷新，并合并并发刷新；401 后最多刷新并重试一次，上传同样适用。双 Token 都过期后需要重新微信登录。刷新接口返回 503 或网络错误时保留本地凭据，401 时清除。发布此次改动后，旧的纯 JWT 会话需重新登录。

会话集成测试自动启动独立的 `redis-server`（仅 Unix socket、不持久化、退出清理），缺少该程序时会跳过对应测试。小程序请求回归测试：`node --test tests/*.test.mjs`。


### 语音接口排查

`POST /api/v1/speech/transcribe` 使用有效登录 Token，以 `multipart/form-data` 上传 `audio` 文件；小程序录音格式为 MP3。当前空 `bailian.api_key` 不启用语音识别，需要在服务进程的环境变量中配置 `BAILIAN_API_KEY`，或在私有配置文件填写 `bailian.api_key`，并重启后端。Key、`BAILIAN_BASE_URL` 和 `BAILIAN_ASR_MODEL` 必须对应同一区域和可用模型，不要把 Key 放入小程序或提交到仓库。启动日志会提示未配置语音服务。

- 401：登录凭据缺失或 Redis 会话失效，重新登录后上传。
- 400：检查 multipart 的字段名 `audio`、文件格式及 7MB 大小限制。
- 503 / `SPEECH_NOT_CONFIGURED`：服务端未配置百炼识别能力。
- 503 / `SERVICE_UNAVAILABLE`：已进入识别服务，但上游请求失败或没有识别结果。

百炼请求结构采用官方 [Qwen-ASR API](https://www.alibabacloud.com/help/zh/model-studio/qwen-asr-api-reference) 的 `/chat/completions`、`input_audio.data` 和 Base64 Data URL 格式。

### 邮箱验证码登录

新增两个无需登录凭据的入口：

- `POST /api/v1/auth/getemailcode`，请求 `{"email":"user@example.com"}`，成功返回 `{"message":"验证码已发送"}`。
- `POST /api/v1/auth/email-login`，请求 `{"email":"user@example.com","email_code":"001234"}`。不需要微信 `code`；成功返回现有 `LoginResult`（双 Token、有效期、用户 ID 和资料）。

在本地私有配置或服务进程环境变量中填写 SMTP 参数，随后重启后端。默认全部留空时禁用邮箱登录，接口返回 503 / `EMAIL_NOT_CONFIGURED`；填写不完整时配置校验失败。已启用时 `main` 调用 SMTP `Init` 检查连接和认证，失败 panic，检查本身不发送邮件。

| YAML 字段 | 环境变量 | 含义 |
|---|---|---|
| `smtp.host` | `SMTP_HOST` | 邮件服务商提供的 SMTP 主机 |
| `smtp.port` | `SMTP_PORT` | 隐式 TLS 通常为 465，STARTTLS 通常为 587，以服务商配置为准 |
| `smtp.username` | `SMTP_USERNAME` | SMTP 登录账号 |
| `smtp.password` | `SMTP_PASSWORD` | 邮箱服务商提供的 SMTP 授权码或密码 |
| `smtp.from` | `SMTP_FROM` | 服务商允许该账号使用的发件邮箱，不带显示名 |
| `smtp.tls_mode` | `SMTP_TLS_MODE` | `tls` 或 `starttls`，禁止明文降级 |

QQ、163 和企业邮箱需先在邮件服务商处启用 SMTP；密钥只填写在本地或部署环境，不放入小程序或提交到仓库。切换 STARTTLS 时同时修改端口。客户端支持 AUTH PLAIN 和 AUTH LOGIN，连接及发送最长 15 秒，跟随请求取消。SMTP 成功表示服务商接受邮件，收件箱仍可能存在投递延迟或垃圾邮件归类。

验证码使用安全随机源生成六位数字，5 分钟过期；Redis 只保存 HMAC-SHA256 摘要（复用 JWT 密钥并用用途域隔离），邮箱和 IP 键也使用摘要。邮件被接受后才启用验证码。发送失败撤销当前批次，但保留发送冷却与额度；SMTP 接受后 Redis 激活失败也不会报告成功。

频率限制：同邮箱 60 秒一次；同邮箱及同 IP 每小时分别最多 10 次发送；同 IP 每分钟最多 30 次校验。每份验证码最多输错 5 次，成功校验后原子删除，不能重复使用。为防伪造转发头，IP 取 TCP 对端；反向代理部署时多个用户可能共享该额度，需在可信代理方案确定后调整。Redis Cluster 中邮箱验证键使用统一 `{email-auth}` hash tag，以支持邮箱/IP 联合限流的 Lua 脚本。

首次邮箱登录在同一 PostgreSQL 事务中创建用户、信用档案与贡献账户；邮箱唯一索引处理并发注册。`open_id` 使用 `email_<uuid>` 占位，保留原有非空唯一约束；不会调用微信或自动合并已有微信账号。邮箱用户尚未绑定微信身份，依赖真实 OpenID 的微信能力不因此自动可用。验证码消费后如果数据库或 Token 存储失败，用户需在冷却结束后重新取码。

已有数据迁移：`model/pgsql.User.Email` 为可空字符串，原微信账号保持 `NULL`。启动时的 `AutoMigrate` 会添加缺失字段和唯一索引；也可在备份后执行增量脚本 `migrations/000010_user_email.sql`。无需重建表或重跑种子数据。SQL 迁移期间应避开高峰；回退应用时可保留新增字段，避免丢失邮箱数据。

验证：`go test ./...` 包含临时 Redis 和本地 TLS SMTP 测试，不发送真实邮件。若设置专用测试库 `EMAIL_TEST_DATABASE_DSN`，还会创建独立测试 schema 验证已有用户迁移、并发注册和事务回滚；不设置则跳过该集成测试。不要使用生产库作为测试库。
