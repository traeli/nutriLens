# COS 上传资源

将本目录中的 **static 文件夹**上传到 COS 的 `nutrilens/v1/` 目录，保留文件夹结构。不要上传本 README，也不要把 cos-assets 这一层放进对象键。

上传后的示例对象键：`nutrilens/v1/static/dining/noodle-shop.jpg`。

在 `.env.production` 填写 `VITE_ASSET_BASE_URL=https://你的资源域名/nutrilens/v1`。完整操作步骤见 [COS 迁移文档](../../docs/cos-static-assets.md)。

本目录不在 src/static 内，不会被 uni-app 默认复制；未配置云端地址的开发构建会由 Vite 钩子复制图片。生产构建必须配置云端地址。
