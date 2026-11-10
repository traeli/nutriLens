import assert from 'node:assert/strict'
import { test } from 'node:test'

import { foodIconSource, foodThemeAccent } from '../src/utils/food-icons.js'

const themes = [
  'dumpling_bun', 'noodles', 'rice_meal', 'local_dishes', 'hotpot',
  'bbq_skewers', 'street_snack', 'breakfast', 'coffee_tea', 'dessert',
  'bakery', 'seafood', 'vegetarian', 'bar_drinks', 'general',
]

test('every supported restaurant theme resolves to a unique COS asset path and accent', async () => {
  const sources = themes.map(foodIconSource)
  assert.equal(new Set(sources).size, themes.length)
  for (const source of sources) assert.match(source, /^\/static\/route-theme-icons\/[a-z_]+\.png$/)
  for (const theme of themes) assert.match(foodThemeAccent(theme), /^#[0-9A-F]{6}$/i)
})
