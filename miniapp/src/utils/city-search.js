import { pinyin } from 'pinyin-pro'

function normalized(value) {
  return String(value || '').trim().toLowerCase().replace(/[\s'-]+/g, '')
}

function isSubsequence(query, target) {
  if (!query) return true
  let queryIndex = 0
  for (let index = 0; index < target.length && queryIndex < query.length; index++) {
    if (target[index] === query[queryIndex]) queryIndex++
  }
  return queryIndex === query.length
}

export function citySearchMeta(city) {
  const syllables = pinyin(String(city && city.name || ''), { toneType: 'none', type: 'array' })
    .map(normalized)
    .filter(Boolean)
  const pinyinText = syllables.join('')
  const initials = syllables.map(item => item[0]).join('')
  const initial = (initials[0] || '#').toUpperCase()
  return { city, pinyin: pinyinText, initials, initial: /^[A-Z]$/.test(initial) ? initial : '#' }
}

export function citySearchGroups(cities, query = '') {
  const keyword = normalized(query)
  const items = (Array.isArray(cities) ? cities : [])
    .map(citySearchMeta)
    .filter(item => {
      if (!keyword) return true
      const name = normalized(item.city.name)
      const code = normalized(item.city.code)
      return name.includes(keyword)
        || code.includes(keyword)
        || item.pinyin.includes(keyword)
        || item.initials.includes(keyword)
        || isSubsequence(keyword, item.pinyin)
        || isSubsequence(keyword, item.initials)
    })
    .sort((left, right) => left.initial.localeCompare(right.initial)
      || left.pinyin.localeCompare(right.pinyin)
      || String(left.city.code || '').localeCompare(String(right.city.code || '')))

  const groups = []
  for (const item of items) {
    let group = groups[groups.length - 1]
    if (!group || group.initial !== item.initial) {
      group = { initial: item.initial, items: [] }
      groups.push(group)
    }
    group.items.push(item.city)
  }
  return groups
}
