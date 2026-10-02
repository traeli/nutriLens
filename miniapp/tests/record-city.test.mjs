import { readFile } from 'node:fs/promises'
import assert from 'node:assert/strict'
import { test } from 'node:test'

const vue = await readFile(new URL('../src/pages/record/record.vue', import.meta.url), 'utf8')
const script = vue.match(/<script>([\s\S]*?)<\/script>/)[1].replace(/^import .+$/gm, '')
const makeComponent = new Function('api', 'getSelectedCity', 'refreshCityOptions', script.replace('export default', 'return'))

async function submit({ initialCity, loadedCity, loadError }) {
  let city = initialCity
  let refreshes = 0
  const requests = []
  const messages = []
  globalThis.uni = {
    getStorageSync: () => 'token',
    showToast: options => messages.push(options.title),
    showModal: () => {},
  }
  const api = {
    createPlace: async data => { requests.push(['place', data]); return { id: 1 } },
    createVisitRecord: async data => { requests.push(['record', data]); return { record: { id: 2 } } },
    submitVisitRecord: async () => ({}),
  }
  const component = makeComponent(api, () => city, async () => {
    refreshes++
    if (loadError) throw new Error('城市加载失败')
    city = loadedCity
  })
  const vm = {
    submitting: false, validateRequiredFields: () => true,
    placeId: 0, place: '测试餐厅', selectedPlaceLocation: { address: '测试地址', longitude: 121, latitude: 31 },
    recommendedDish: '', today: () => '2026-10-02', conclusion: 'recommend', averageCost: '', waitMinutes: '',
    content: '', selectedTags: [], recordRequestKey: 'test', receiptImage: '', experiencePhotos: [],
    resetVisitForm: () => {}, loadRecentRecords: async () => {},
  }
  vm.ensureRecordCity = component.methods.ensureRecordCity.bind(vm)
  await component.methods.finishRecord.call(vm)
  return { requests, refreshes, messages, vm }
}

test('cold city cache is loaded before /places and code is included in both writes', async () => {
  const result = await submit({ initialCity: { code: '' }, loadedCity: { code: '310000', name: '上海' } })
  assert.equal(result.refreshes, 1)
  assert.equal(result.requests.length, 2)
  assert.equal(result.requests[0][1].city_code, '310000')
  assert.equal(result.requests[1][1].city_code, '310000')
})

test('selected city is preserved without loading a fallback', async () => {
  const result = await submit({ initialCity: { code: '330100', name: '杭州' } })
  assert.equal(result.refreshes, 0)
  assert.equal(result.requests[0][1].city_code, '330100')
})

test('empty city list prevents /places request and releases submitting state', async () => {
  const result = await submit({ initialCity: { code: '' }, loadedCity: { code: '' } })
  assert.equal(result.requests.length, 0)
  assert.match(result.messages[0], /暂无可用城市/)
  assert.equal(result.vm.submitting, false)
})

test('city loading failure prevents writes', async () => {
  const result = await submit({ initialCity: { code: '' }, loadError: true })
  assert.equal(result.requests.length, 0)
  assert.equal(result.messages[0], '城市加载失败')
})
