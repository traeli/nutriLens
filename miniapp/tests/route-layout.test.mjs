import assert from 'node:assert/strict'
import { readFile } from 'node:fs/promises'
import { test } from 'node:test'

import { projectRouteStops, routeSketchLayout } from '../src/utils/route-layout.js'

const routeDetailSource = await readFile(new URL('../src/pages/route-detail/route-detail.vue', import.meta.url), 'utf8')

test('west-to-east coordinates render from left to right', () => {
  const points = projectRouteStops([
    { longitude: 121.42, latitude: 31.23 },
    { longitude: 121.43, latitude: 31.23 },
    { longitude: 121.44, latitude: 31.23 },
  ])
  assert.ok(points[0].x < points[1].x && points[1].x < points[2].x)
  assert.ok(points.every(point => Math.abs(point.y - 50) < 0.1))
})

test('southwest-to-northeast coordinates render from lower-left to upper-right', () => {
  const points = projectRouteStops([
    { longitude: 121.42, latitude: 31.22 },
    { longitude: 121.43, latitude: 31.23 },
    { longitude: 121.44, latitude: 31.24 },
  ])
  assert.ok(points[0].x < points[2].x)
  assert.ok(points[0].y > points[2].y)
})

test('identical or unavailable coordinates use a readable diagonal fallback', () => {
  const points = projectRouteStops([{ name: '一' }, { name: '二' }, { name: '三' }])
  assert.equal(points.length, 3)
  assert.ok(points[0].x < points[2].x)
  assert.ok(points[0].y > points[2].y)
})

test('layout creates one continuous quadratic curve chain through every stop', () => {
  const stops = [
    { longitude: 121.42, latitude: 31.22 },
    { longitude: 121.43, latitude: 31.23 },
    { longitude: 121.44, latitude: 31.24 },
  ]
  const layout = routeSketchLayout(stops)
  assert.equal(layout.points.length, 3)
  assert.equal(layout.curves.length, 2)
  assert.deepEqual(layout.curves[0].end, layout.curves[1].start)
})

test('horizontal coordinates receive a visible alternating bend instead of a straight connector', () => {
  const layout = routeSketchLayout([
    { longitude: 121.42, latitude: 31.23 },
    { longitude: 121.43, latitude: 31.23 },
    { longitude: 121.44, latitude: 31.23 },
  ])
  assert.ok(layout.curves[0].control.y > 53)
  assert.ok(layout.curves[1].control.y < 47)
})

test('route detail draws the complete curve chain in one canvas stroke', () => {
  assert.match(routeDetailSource, /<canvas[^>]+canvas-id="route-line-canvas"/)
  assert.equal((routeDetailSource.match(/context\.beginPath\(\)/g) || []).length, 1)
  assert.equal((routeDetailSource.match(/context\.stroke\(\)/g) || []).length, 1)
  assert.equal((routeDetailSource.match(/context\.draw\(false\)/g) || []).length, 1)
  assert.doesNotMatch(routeDetailSource, /sketchLayout\.strokes/)
})
