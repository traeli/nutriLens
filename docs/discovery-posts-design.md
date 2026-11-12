# 发现帖子功能设计与发布说明

更新日期：2026-11-12

## 产品边界

- 小程序主导航为“首页 / 发现 / 我的”，不再提供记录、足迹或私人营养记录入口。
- 发现帖子只由运营人员在 Directus 后台创建、编辑和发布；普通用户没有帖子、评论、图片或语音上传接口。
- 用户只能浏览、收藏地点、对发现帖子标记“有用”和举报异常内容。
- 点赞仅展示聚合数量，不提供点赞用户列表、关注关系、私信或排行榜。

## 数据结构

- `discovery_posts`：标题、正文、作者、来源、状态、发布时间、地点关系和聚合点赞数。
- `discovery_post_images`：远程 COS 图片、排序和尺寸。数据库使用 `(post_id, sort_order)` 唯一索引，并将 `sort_order` 约束在 `0..5`，因此单帖最多六张图。
- `discovery_post_likes`：`(post_id, user_id)` 复合主键，保证每个用户对每篇帖子最多一次标记。
- 店铺名称和地址统一从 `places` 读取，不在帖子表重复维护。

## API

| Method | Path | 说明 |
|---|---|---|
| GET | `/api/v1/discovery/posts` | 按城市分页读取已发布帖子 |
| GET | `/api/v1/discovery/posts/:id` | 读取已发布帖子详情 |
| GET | `/api/v1/places/:id/discovery-posts` | 读取店铺关联帖子 |
| POST | `/api/v1/discovery/posts/:id/like` | 登录用户标记有用 |
| DELETE | `/api/v1/discovery/posts/:id/like` | 取消有用标记 |
| GET | `/api/v1/me/discovery-post-likes` | 获取本人已点赞帖子编号 |

原 `/records`、`/nutrition/records`、`/speech/transcribe`、用户地点提交和 `/experiences` 路由不再注册。

## Directus 维护

服务端启动完成自动建表后，在 Directus 刷新 Schema，并将三个新表加入数据模型：

1. `discovery_posts.place_id` 关联 `places.id`。
2. `discovery_post_images.post_id` 关联 `discovery_posts.id`，按 `sort_order` 升序。
3. 运营角色拥有帖子和图片的创建、更新、删除权限；小程序用户不通过 Directus 写入。
4. 只有 `status=published` 的帖子会被公开 API 返回。
5. 图片先上传 COS，再将地址写入 `image_url`；每篇使用 `sort_order=0..5`。

历史 `restaurant_reviews` 中已发布的运营食记会在启动时幂等迁移到新表；历史点赞同步迁移。旧私人数据保留用于账号注销清理，但不再提供产品入口。
