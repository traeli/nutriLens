const DEFAULT_PADDING = { left: 14, right: 14, top: 17, bottom: 17 }

function validCoordinate(stop) {
  const longitude = Number(stop && stop.longitude)
  const latitude = Number(stop && stop.latitude)
  return Number.isFinite(longitude) && Number.isFinite(latitude) && Boolean(longitude || latitude)
}

function fallbackPoints(length, padding) {
  if (length <= 1) return [{ x: 50, y: 50 }].slice(0, length)
  return Array.from({ length }, (_, index) => {
    const ratio = index / (length - 1)
    return {
      x: padding.left + ratio * (100 - padding.left - padding.right),
      y: 100 - padding.bottom - ratio * (100 - padding.top - padding.bottom),
    }
  })
}

// 将 GCJ-02 坐标等比投影到近似方形的手绘画布；纬度反转后，东北方向自然落在右上方。
export function projectRouteStops(stops, options = {}) {
  const items = Array.isArray(stops) ? stops : []
  const padding = { ...DEFAULT_PADDING, ...(options.padding || {}) }
  if (!items.length) return []
  if (!items.every(validCoordinate)) return fallbackPoints(items.length, padding)

  const averageLatitude = items.reduce((sum, stop) => sum + Number(stop.latitude), 0) / items.length
  const longitudeScale = Math.max(0.2, Math.cos(averageLatitude * Math.PI / 180))
  const raw = items.map(stop => ({
    x: Number(stop.longitude) * longitudeScale,
    y: Number(stop.latitude),
  }))
  const minX = Math.min(...raw.map(point => point.x))
  const maxX = Math.max(...raw.map(point => point.x))
  const minY = Math.min(...raw.map(point => point.y))
  const maxY = Math.max(...raw.map(point => point.y))
  const spanX = maxX - minX
  const spanY = maxY - minY
  if (spanX < 1e-7 && spanY < 1e-7) return fallbackPoints(items.length, padding)

  const drawableWidth = 100 - padding.left - padding.right
  const drawableHeight = 100 - padding.top - padding.bottom
  const scale = Math.min(
    spanX > 0 ? drawableWidth / spanX : Number.POSITIVE_INFINITY,
    spanY > 0 ? drawableHeight / spanY : Number.POSITIVE_INFINITY,
  )
  const usedWidth = spanX * scale
  const usedHeight = spanY * scale
  const offsetX = padding.left + (drawableWidth - usedWidth) / 2
  const offsetY = padding.top + (drawableHeight - usedHeight) / 2

  return raw.map(point => ({
    x: Number((offsetX + (point.x - minX) * scale).toFixed(2)),
    y: Number((offsetY + (maxY - point.y) * scale).toFixed(2)),
  }))
}

export function routeSketchLayout(stops) {
  const points = projectRouteStops(stops)
  const curves = []
  points.slice(0, -1).forEach((point, index) => {
    const next = points[index + 1]
    const dx = next.x - point.x
    const dy = next.y - point.y
    const distance = Math.sqrt(dx * dx + dy * dy)
    if (!distance) return

    const normalX = -dy / distance
    const normalY = dx / distance
    const bend = Math.min(14, Math.max(7, distance * 0.2)) * (index % 2 ? -1 : 1)
    const control = {
      x: (point.x + next.x) / 2 + normalX * bend,
      y: (point.y + next.y) / 2 + normalY * bend,
    }
    curves.push({
      start: { ...point },
      control: { x: Number(control.x.toFixed(2)), y: Number(control.y.toFixed(2)) },
      end: { ...next },
      segment: index,
    })
  })
  return {
    points: points.map((point, index) => ({ ...point, stop: stops[index], index })),
    curves,
  }
}
