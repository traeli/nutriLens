import { readFile } from 'node:fs/promises'
import assert from 'node:assert/strict'
import { test } from 'node:test'

const original = await readFile(new URL('../src/api/request.js', import.meta.url), 'utf8')
let sequence = 0
async function setup(handle) {
  const storage = new Map([['token', 'old'], ['refresh_token', 'refresh-old'], ['token_expires_at', Date.now() + 7200000]])
  const calls = []
  globalThis.uni = {
    getStorageSync: key => storage.get(key),
    setStorageSync: (key, value) => storage.set(key, value),
    removeStorageSync: key => storage.delete(key),
    request: options => { calls.push(options); queueMicrotask(() => handle(options)) },
    uploadFile: options => { calls.push(options); queueMicrotask(() => handle(options)) },
  }
  const source = original.replace(/import \{ resolveAssetFields \}[^\n]+/, 'const resolveAssetFields = value => value').replace('import.meta.env.VITE_BASE_URL', 'undefined')
  const module = await import('data:text/javascript;base64,' + Buffer.from(source + `\n// 测试 ${sequence++}`).toString('base64'))
  return { ...module, storage, calls }
}
const pair = { token: 'new', refresh_token: 'refresh-new', expires_in: 7200 }
const unauthorized = { statusCode: 401, data: { error: { message: 'expired' } } }

test('concurrent 401 requests share one refresh and retry with the new token', async () => {
  const { api, calls, storage } = await setup(o => {
    if (o.url.endsWith('/auth/refresh')) o.success({ statusCode: 200, data: pair })
    else if (o.header.Authorization === 'Bearer old') o.success(unauthorized)
    else o.success({ statusCode: 200, data: { ok: true } })
  })
  assert.deepEqual(await Promise.all([api.getProfile(), api.getCities()]), [{ ok: true }, { ok: true }])
  assert.equal(calls.filter(o => o.url.endsWith('/auth/refresh')).length, 1)
  assert.equal(storage.get('token'), 'new')
})

test('refreshes before expiry and uploads use the refreshed token', async () => {
  const { api, storage, calls } = await setup(o => o.success({ statusCode: 200, data: o.url.endsWith('/auth/refresh') ? pair : '{"ok":true}' }))
  storage.set('token_expires_at', Date.now() + 30000)
  assert.deepEqual(await api.uploadRecordMedia(1, '/tmp/photo.jpg'), { ok: true })
  assert.ok(calls[0].url.endsWith('/auth/refresh'))
  assert.equal(calls[1].header.Authorization, 'Bearer new')
})

test('upload 401 retries the upload once', async () => {
  const { api, calls } = await setup(o => {
    if (o.url.endsWith('/auth/refresh')) o.success({ statusCode: 200, data: pair })
    else o.success(o.header.Authorization === 'Bearer old' ? { ...unauthorized, data: JSON.stringify(unauthorized.data) } : { statusCode: 200, data: '{"ok":true}' })
  })
  assert.deepEqual(await api.uploadEvidence(1, '/tmp/photo.jpg'), { ok: true })
  assert.equal(calls.filter(o => o.filePath).length, 2)
})

test('temporary refresh outage retains credentials', async () => {
  const { api, storage } = await setup(o => o.success(o.url.endsWith('/auth/refresh') ? { statusCode: 503, data: {} } : unauthorized))
  await assert.rejects(api.getProfile())
  assert.equal(storage.get('token'), 'old')
  assert.equal(storage.get('refresh_token'), 'refresh-old')
})

test('invalid refresh clears the complete session', async () => {
  const { api, storage } = await setup(o => o.success(unauthorized))
  await assert.rejects(api.getProfile())
  assert.equal(storage.size, 0)
})

test('logout during refresh cannot restore credentials', async () => {
  let pending
  const { api, clearSession, storage } = await setup(o => { pending = o })
  const result = api.refreshToken()
  await Promise.resolve()
  clearSession()
  pending.success({ statusCode: 200, data: pair })
  await assert.rejects(result)
  assert.equal(storage.size, 0)
})

test('login stores both credentials and is sent without bearer authorization', async () => {
  const { api, storage, calls } = await setup(o => o.success({ statusCode: 200, data: pair }))
  await api.wxLogin('code')
  assert.equal(calls[0].header.Authorization, undefined)
  assert.equal(storage.get('refresh_token'), 'refresh-new')
})

test('email login uses email_code without WeChat code and saves both credentials', async () => {
  const { api, storage, calls } = await setup(o => o.success({ statusCode: 200, data: pair }))
  await api.emailLogin('user@example.com', '001234')
  assert.deepEqual(calls[0].data, { email: 'user@example.com', email_code: '001234' })
  assert.equal(calls[0].header.Authorization, undefined)
  assert.equal(storage.get('refresh_token'), 'refresh-new')
})

test('send email code is public and never sends the session token', async () => {
  const { api, calls } = await setup(o => o.success({ statusCode: 200, data: { message: '验证码已发送' } }))
  await api.sendEmailCode('user@example.com')
  assert.ok(calls[0].url.endsWith('/auth/getemailcode'))
  assert.deepEqual(calls[0].data, { email: 'user@example.com' })
  assert.equal(calls[0].header.Authorization, undefined)
})
