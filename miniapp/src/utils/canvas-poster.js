/**
 * Canvas 2D poster rendering — template-based share poster.
 *
 * Layers (bottom to top):
 *   1. Template background image
 *   2. Food photo (clipped to designated area)
 *   3. Food name text
 *   4. Red stamp with dynamic evaluation text
 */

// Poster dimensions (CSS pixels)
const WIDTH = 375
const HEIGHT = 562 // 2:3 ratio matching template
const DPR = 2

// Template image source (local static asset)
const TEMPLATE_SRC = '/static/share-template.jpg'

// Layout coordinates (relative to WIDTH x HEIGHT)
// Adjust these to match your template's design
const LAYOUT = {
  // Food photo area (where the food image is placed on the template)
  photo: { x: 28, y: 44, w: 319, h: 210 },
  // Food name text (below photo)
  foodName: { x: 187, y: 272, maxW: 320, fontSize: 'bold 17px', lineH: 26 },
  // Stamp (top-right corner)
  stamp: { cx: 298, cy: 102, w: 80, h: 36 },
}

// Color theme
const COLORS = {
  stampBorder: '#D32F2F',
  stampText: '#D32F2F',
  foodName: '#FFFFFF',
  foodNameShadow: 'rgba(0,0,0,0.5)',
}

// ==================== Drawing Utilities ====================

function roundRect(ctx, x, y, w, h, r) {
  ctx.beginPath()
  ctx.moveTo(x + r, y)
  ctx.lineTo(x + w - r, y)
  ctx.arcTo(x + w, y, x + w, y + r, r)
  ctx.arcTo(x + w, y + h, x + w - r, y + h, r)
  ctx.arcTo(x, y + h, x, y + h - r, r)
  ctx.arcTo(x, y, x + r, y, r)
  ctx.closePath()
}

function loadImage(canvas, src) {
  return new Promise((resolve) => {
    if (!src) return resolve(null)
    try {
      const img = canvas.createImage()
      img.onload = () => resolve(img)
      img.onerror = () => resolve(null)
      img.src = src
    } catch (e) {
      resolve(null)
    }
  })
}

// ==================== Stamp Drawing ====================

function drawStamp(ctx, cx, cy, w, h, text) {
  if (!text) return

  ctx.save()

  // Translate to center, rotate slightly for realism
  ctx.translate(cx, cy)
  ctx.rotate((-12 * Math.PI) / 180)

  const halfW = w / 2
  const halfH = h / 2
  const r = 6 // corner radius

  // Stamp border
  ctx.strokeStyle = COLORS.stampBorder
  ctx.lineWidth = 2.5
  roundRect(ctx, -halfW, -halfH, w, h, r)
  ctx.stroke()

  // Inner border (double-line stamp effect)
  ctx.strokeStyle = COLORS.stampBorder
  ctx.lineWidth = 1.2
  roundRect(ctx, -halfW + 4, -halfH + 4, w - 8, h - 8, r - 2)
  ctx.stroke()

  // Stamp text
  ctx.fillStyle = COLORS.stampText
  ctx.font = 'bold 16px "PingFang SC", "Microsoft YaHei", sans-serif'
  ctx.textAlign = 'center'
  ctx.textBaseline = 'middle'
  ctx.fillText(text, 0, 1)

  ctx.restore()
}

// ==================== Food Name Text ====================

function drawFoodNames(ctx, foods, x, y, maxW, fontSize, lineH, color) {
  if (!foods || foods.length === 0) return

  ctx.font = fontSize
  ctx.textAlign = 'center'
  ctx.textBaseline = 'top'

  const names = foods.map((f) => f.name || f).filter(Boolean)
  // Build lines that fit within maxW
  const lines = []
  let current = ''
  for (const name of names) {
    const sep = current ? '  ·  ' : ''
    const test = current + sep + name
    if (ctx.measureText(test).width > maxW && current) {
      lines.push(current)
      current = name
    } else {
      current = test
    }
  }
  if (current) lines.push(current)

  // Shadow for readability
  ctx.fillStyle = COLORS.foodNameShadow
  for (const line of lines) {
    ctx.fillText(line, x + 1, y + 1)
    y += lineH
  }
  // Reset y back to start for actual text
  y -= lines.length * lineH

  ctx.fillStyle = color
  for (const line of lines) {
    ctx.fillText(line, x, y)
    y += lineH
  }
}

// ==================== Main Entry Point ====================

/**
 * Draw the full share poster.
 *
 * @param {Object} canvas - WeChat Canvas 2D node
 * @param {Object} data
 * @param {string} data.imageUrl - food photo COS URL
 * @param {Array}  data.foods    - [{name: "宫保鸡丁"}, ...]
 * @param {string} data.stampText - stamp evaluation text, e.g. "夯爆了"
 */
export async function drawFoodPoster(canvas, data) {
  const { imageUrl, foods, stampText } = data

  canvas.width = WIDTH * DPR
  canvas.height = HEIGHT * DPR

  const ctx = canvas.getContext('2d')
  ctx.scale(DPR, DPR)

  // ---- Load images in parallel ----
  const [templateImg, foodImg] = await Promise.all([
    loadImage(canvas, TEMPLATE_SRC),
    loadImage(canvas, imageUrl),
  ])

  // ---- 1. Draw template background ----
  if (templateImg) {
    ctx.drawImage(templateImg, 0, 0, WIDTH, HEIGHT)
  } else {
    // Fallback: solid warm background
    ctx.fillStyle = '#E8D5B7'
    ctx.fillRect(0, 0, WIDTH, HEIGHT)
  }

  // ---- 2. Draw food photo ----
  if (foodImg) {
    const { x, y, w, h } = LAYOUT.photo
    ctx.save()
    roundRect(ctx, x, y, w, h, 10)
    ctx.clip()
    // Cover fill (center-crop-ish)
    const imgRatio = foodImg.width / foodImg.height
    const boxRatio = w / h
    let sx, sy, sw, sh
    if (imgRatio > boxRatio) {
      sh = foodImg.height
      sw = sh * boxRatio
      sx = (foodImg.width - sw) / 2
      sy = 0
    } else {
      sw = foodImg.width
      sh = sw / boxRatio
      sx = 0
      sy = (foodImg.height - sh) / 2
    }
    ctx.drawImage(foodImg, sx, sy, sw, sh, x, y, w, h)
    ctx.restore()
  }

  // ---- 3. Draw food names ----
  drawFoodNames(
    ctx,
    foods,
    LAYOUT.foodName.x,
    LAYOUT.foodName.y,
    LAYOUT.foodName.maxW,
    `${LAYOUT.foodName.fontSize} "PingFang SC", "Microsoft YaHei", sans-serif`,
    LAYOUT.foodName.lineH,
    COLORS.foodName,
  )

  // ---- 4. Draw red stamp ----
  if (stampText) {
    const { cx, cy, w, h } = LAYOUT.stamp
    drawStamp(ctx, cx, cy, w, h, stampText)
  }
}
