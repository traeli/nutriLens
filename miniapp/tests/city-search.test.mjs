import assert from 'node:assert/strict'
import { test } from 'node:test'

import { citySearchGroups } from '../src/utils/city-search.js'

const cities = [
  { code: '310000', name: '上海' },
  { code: '110000', name: '北京' },
  { code: '500000', name: '重庆' },
  { code: '610100', name: '西安' },
  { code: '430100', name: '长沙' },
  { code: '440100', name: '广州' },
]

test('cities are grouped and sorted by pinyin initial', () => {
  const groups = citySearchGroups(cities)
  assert.deepEqual(groups.map(group => group.initial), ['B', 'C', 'G', 'S', 'X'])
  assert.deepEqual(groups.flatMap(group => group.items.map(city => city.name)), ['北京', '长沙', '重庆', '广州', '上海', '西安'])
})

test('fuzzy search supports Chinese, full pinyin and initials', () => {
  assert.deepEqual(citySearchGroups(cities, '海').flatMap(group => group.items.map(city => city.name)), ['上海'])
  assert.deepEqual(citySearchGroups(cities, 'chongq').flatMap(group => group.items.map(city => city.name)), ['重庆'])
  assert.deepEqual(citySearchGroups(cities, 'cq').flatMap(group => group.items.map(city => city.name)), ['重庆'])
  assert.deepEqual(citySearchGroups(cities, 'xa').flatMap(group => group.items.map(city => city.name)), ['西安'])
})
