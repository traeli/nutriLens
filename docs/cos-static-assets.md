# 小程序静态图片迁移到 COS

## 当前项目使用的地址

实际上传保留了 `cos-assets/` 外层目录，开发和生产配置现已统一为：

```dotenv
VITE_ASSET_BASE_URL=https://nutrilens-1429775067.cos.ap-guangzhou.myqcloud.com/cos-assets
```

因此图片对象路径应为 `cos-assets/static/dining/noodle-shop.jpg` 等；完整地址为 `https://nutrilens-1429775067.cos.ap-guangzhou.myqcloud.com/cos-assets/static/dining/noodle-shop.jpg`。下面的 `nutrilens/v1/` 是原先规划示例，不需要为遵循示例重新搬动已上传文件。开发和生产现在都使用 COS；只有手动将开发配置的资源根地址留空，才恢复离线图片模式。修改配置后需重启编译。

## 本次整理

原 `miniapp/src/static` 文件内容共 2,818,321 字节（约 2.69 MiB）。uni-app 会将 static 目录直接复制到输出目录，即使图片没有被页面引用，也会占用包体积。

- 7 张在用大图移到 `miniapp/cos-assets/static`：1,876,854 字节（约 1.79 MiB）。这个目录用于上传 COS，不是小程序源码目录。
- 2 张未发现运行时代码或种子 SQL 引用的图片移到 `miniapp/asset-archive/dining`：935,364 字节（约 913 KiB），保留原图，不参与构建。
- `miniapp/src/static` 保留 logo 和 8 张底部导航图标：6,103 字节（约 5.96 KiB）。导航图标继续使用本地路径。

配置 COS 后可从包内移除约 2.68 MiB 图片内容。这个数值是源文件大小差，不是微信压缩后的上传包大小，也不保证整个主包小于截图中的 1.5 M。截图为代码质量检查；是否存在硬性上传限制，需结合完整上传错误判断。

本次未运行构建、测试或上传到 COS，现有 dist 中的旧产物不会因移动源文件立即改变，需要重新构建。

## 1. 上传目录

在 COS 中建立 `nutrilens/v1/` 目录，然后将本地 `miniapp/cos-assets/` **里面的 static 文件夹**上传进去。不要上传 cos-assets 外层文件夹，也不要只上传压缩包。

上传完成后，对象键应当是：

```text
nutrilens/v1/static/dining/corner-cafe.jpg
nutrilens/v1/static/dining/login-cheers-background-v2.jpg
nutrilens/v1/static/dining/noodle-detail.jpg
nutrilens/v1/static/dining/noodle-shop.jpg
nutrilens/v1/static/dining/shanghai-city.jpg
nutrilens/v1/static/dining/visit-rating-art-v2.jpg
nutrilens/v1/static/posters/food-memory-night-v1.jpg
```

| 相对路径 | 大小 | 当前用途 |
|---|---:|---|
| static/dining/corner-cafe.jpg | 205,396 B | 店铺默认图、城市和路线种子数据 |
| static/dining/login-cheers-background-v2.jpg | 243,908 B | 登录页背景 |
| static/dining/noodle-detail.jpg | 180,464 B | 餐食默认图、城市和路线种子数据 |
| static/dining/noodle-shop.jpg | 221,228 B | 店铺默认图、城市和路线种子数据 |
| static/dining/shanghai-city.jpg | 195,917 B | 上海城市种子图 |
| static/dining/visit-rating-art-v2.jpg | 624,222 B | 记录页、个人资料页插图 |
| static/posters/food-memory-night-v1.jpg | 205,719 B | 足迹页入口、海报主题种子数据 |

归档的 `login-journal-cover.jpg` 和 `shanghai-hero.jpg` 暂不需要上传。如果现有线上数据库中仍有人为配置了这两张图，可将对应归档文件额外上传到同一个 `static/dining/` 前缀，原路径转换仍然有效。归档判断只依据仓库引用，未读取线上数据库。

## 2. 在腾讯云控制台操作

1. 进入对象存储 COS → 存储桶列表，使用现有的公开静态资源桶，或新建专门的静态资源桶。选择适合用户所在地的地域。
2. 这些是公开页面插图，需要允许匿名读取。专用静态资源桶可选择“公有读私有写”；若复用私有桶，仅开放本次静态资源目录或对象的读取权限，不要开放整个私有数据桶。
3. 进入文件列表，创建 `nutrilens/v1/` 路径，在该路径中选择“上传文件夹”，上传上述 `static` 文件夹。核对对象键，避免出现重复 `static/static` 或多余 `cos-assets`。
4. 文件类型应为 `image/jpeg`。使用可长期访问的 HTTPS 地址，不要把有过期时间的临时签名链接写入前端配置。
5. 在控制台复制其中一张图片的对象地址，确认未登录腾讯云时也能下载到正确图片，且不是 403/404。部分新桶默认域名在浏览器中会触发下载而非直接预览，这与对象不存在不同；正式图片分发可配置自定义 HTTPS 域名或 CDN。

只上传本次整理的公开插图。用户消费凭证、手机号等私密内容不在本次迁移范围。小程序读取公开图片不需要 COS SDK、SecretId 或 SecretKey；不要把云账号密钥放进前端环境变量。

## 3. 填写资源根地址

例如对象完整地址为（示例，必须换成你自己的域名）：

```text
https://your-bucket-1250000000.cos.ap-guangzhou.myqcloud.com/nutrilens/v1/static/dining/noodle-shop.jpg
```

则 `miniapp/.env.production` 填写：

```dotenv
VITE_ASSET_BASE_URL=https://your-bucket-1250000000.cos.ap-guangzhou.myqcloud.com/nutrilens/v1
```

使用自定义域名时也一样：

```dotenv
VITE_ASSET_BASE_URL=https://assets.example.com/nutrilens/v1
```

不要在资源根地址末尾额外添加 `/static`。代码会将 `/static/dining/noodle-shop.jpg` 拼接到根地址后。

`VITE_BASE_URL` 是 Go 后端地址，`VITE_ASSET_BASE_URL` 是静态图片地址，两者用途不同。现有 `.env.production` 的后端仍指向 `https://nutrilens.cloud`，发布前需另行确认它是不是你的后端。

本地 `miniapp/.env.development` 的资源地址可留空：构建钩子会从 cos-assets 复制在用图片，以保留离线开发。想在开发时查看云端图片，可在开发配置中填同样的资源根地址。生产配置不能留空，构建会给出明确提示。

## 4. 微信域名配置

在微信公众平台的小程序后台进入开发管理/开发设置中的服务器域名，按实际使用的 COS 或 CDN HTTPS 主机配置 downloadFile 合法域名。只填协议和主机，不填 `/nutrilens/v1` 这样的路径。

例如：

```text
https://your-bucket-1250000000.cos.ap-guangzhou.myqcloud.com
```

普通 image 展示与 wx.downloadFile、getImageInfo 等接口的域名要求不同；如果后续下载图片、生成包含远程图片的画布，应确保对应域名满足下载接口要求。Go API 域名仍需配置在 request 合法域名中；本次没有增加小程序直传 COS 功能，不需要因此配置 COS uploadFile 域名。

正式发布前以真机、开启正常域名校验的结果为准，不依赖开发者工具跳过校验。若微信后台不接受默认 COS 域名，使用符合平台要求的自定义 HTTPS 域名。

## 5. 重新构建和上传

先上传图片并配置好域名，再在 miniapp 目录执行：

```bash
npm run build:mp-weixin
```

微信开发者工具打开 `miniapp/dist/build/mp-weixin`，不要上传 `dist/dev/mp-weixin`。没有配置 COS 的开发构建仍包含本地图片，体积不会达到生产构建的精简效果。

构建完成后需要人工确认：

- 输出目录 `static` 里不再有 dining 和 posters 大图文件夹。
- 首页城市/路线、登录背景、个人资料、记录入口、足迹页图片可以加载。
- 微信工具中查看实际主包/上传包体积，并重新扫描代码质量。

修改环境变量后必须重启编译，已有构建包不会自动替换地址。旧版客户端可能缓存了城市图片，切换资源域名时可在开发者工具中清除本地缓存再看效果。

如果移除图片后主包仍超标，应依据新构建的体积分析继续处理 JS 和页面分包，不要认为重新上传同一份旧产物能解决问题。

## 6. 后续维护

`src/utils/assets.js` 只转换 `/static/dining/` 和 `/static/posters/` 路径；HTTP(S) 图片地址、临时文件、用户上传内容及导航图标不会被重写。API 响应中的图片字段和本地城市缓存都会经过转换，因此不必修改数据库里的种子图片地址。

新增公开插图放在 `cos-assets/static/dining` 或 `cos-assets/static/posters`，页面通过 `assetUrl()` 使用。替换线上图片建议采用新文件名或新版本目录（如 `nutrilens/v2`），旧目录保留给旧版小程序使用，避免覆盖缓存或造成旧版 404。

离线开发构建只在重新编译时复制图片；单独替换 cos-assets 中的源图后应重启开发编译。归档目录和上传说明不参与打包。没有对图片重新压缩或改变视觉内容。

参考：

- [uni-app 静态资源规则](https://zh.uniapp.dcloud.io/tutorial/page-static-assets.html)
- [腾讯云 COS 控制台上传](https://cloud.tencent.com/document/product/436/14113)
- [COS 访问权限](https://cloud.tencent.cn/document/product/436/30749)
- [COS 默认域名与预览限制](https://cloud.tencent.com/document/product/436/102489)
- [腾讯云的小程序下载域名配置示例](https://cloud.tencent.com.cn/document/product/436/122402)
