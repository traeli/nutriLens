import { defineConfig, loadEnv } from 'vite'
import { cpSync, mkdirSync, rmSync } from 'node:fs'
import { resolve, relative, isAbsolute, sep } from 'node:path'
import { fileURLToPath } from 'node:url'
import uni from '@dcloudio/vite-plugin-uni'

const projectRoot = fileURLToPath(new URL('.', import.meta.url))
const cloudAssets = resolve(projectRoot, 'cos-assets')

// 本地开发仍可离线使用图片；配置云端地址后，不将这些文件打入小程序。
function localArtworkPlugin(assetBaseURL) {
  let outputDir
  return {
    name: 'nutrilens:local-artwork',
    apply: 'build',
    enforce: 'post',
    configResolved(config) {
      outputDir = resolve(config.root, config.build.outDir)
    },
    writeBundle: {
      order: 'post',
      sequential: true,
      handler() {
        // 只操作本项目 dist 内的生成资源，绝不删除源图。
        const outputRelative = relative(resolve(projectRoot, 'dist'), outputDir)
        if (!outputRelative || outputRelative === '..' || outputRelative.startsWith('..' + sep) || isAbsolute(outputRelative)) {
          throw new Error('静态资源输出目录必须位于 miniapp/dist 的子目录中')
        }
        for (const group of ['dining', 'posters']) {
          const destination = resolve(outputDir, 'static', group)
          // 清理之前离线构建留下的图片，避免切换 COS 后仍占包体积。
          rmSync(destination, { recursive: true, force: true })
          if (!assetBaseURL) {
            mkdirSync(resolve(outputDir, 'static'), { recursive: true })
            cpSync(resolve(cloudAssets, 'static', group), destination, { recursive: true })
          }
        }
      },
    },
  }
}

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, projectRoot, 'VITE_')
  const assetBaseURL = (process.env.VITE_ASSET_BASE_URL || env.VITE_ASSET_BASE_URL || '').trim()
  if (assetBaseURL) {
    let parsed
    try { parsed = new URL(assetBaseURL) } catch { throw new Error('VITE_ASSET_BASE_URL 必须是有效的 HTTPS 资源根地址') }
    if (parsed.protocol !== 'https:' || parsed.username || parsed.password || parsed.search || parsed.hash) {
      throw new Error('VITE_ASSET_BASE_URL 必须使用 HTTPS，且不能包含凭据、查询参数或片段')
    }
  } else if (mode === 'production') {
    throw new Error('请先按 docs/cos-static-assets.md 上传图片并填写 VITE_ASSET_BASE_URL，再构建发布版本')
  }
  return {
    // H5 开发服务器也可访问外置源图，小程序开发则由上面的构建钩子复制。
    ...(process.env.UNI_PLATFORM === 'h5' && !assetBaseURL ? { publicDir: cloudAssets } : {}),
    plugins: [uni(), localArtworkPlugin(assetBaseURL)],
  }
})
