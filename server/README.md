# 食迹簿后端

这是面向城市餐饮发现和真实到店记录的新后端。旧 NutriLens 营养、转盘、积分、成就、提醒和 Webhook 业务不在本项目中。

## 本地启动

```bash
go run ./cmd/api
```

默认从当前工作目录读取 `config/config.yaml`，也可以通过 `CONFIG_PATH` 指定其他文件。当前 Docker Compose 会将 `server/config/config.yaml` 只读挂载到容器内 `/app/config/config.yaml`，后端业务配置直接从该文件读取。部署前必须填写新的微信密钥、百炼 Key、JWT 密钥和数据加密密钥，且不得提交生产使用的真实密钥。

开发环境可以向 `/api/v1/auth/wx-login` 提交 `code=the code is a mock one` 创建本地测试账号。

## 数据库

项目不会自动修改数据库结构。首次部署前显式执行：

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

迁移也不会清理旧业务遗留表。是否归档或删除旧营养业务数据，应在确认备份与保留期限后通过单独的数据迁移处理。

## 当前接口范围

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
