import assert from 'node:assert/strict'
import { describe, it } from 'node:test'

import { errorMessage, isAuthExpiredError } from './errors.js'

describe('errorMessage', () => {
  it('uses string errors returned by Wails', () => {
    assert.equal(errorMessage('JWT_SECRET is required'), 'JWT_SECRET is required')
  })

  it('uses standard Error messages', () => {
    assert.equal(errorMessage(new Error('invalid username or password')), 'invalid username or password')
  })

  it('falls back when the error shape is empty', () => {
    assert.equal(errorMessage(null), '请求失败，请稍后重试')
  })
})

describe('isAuthExpiredError', () => {
  it('matches normalized auth expiry errors', () => {
    assert.equal(isAuthExpiredError(new Error('登录已过期，请重新登录')), true)
  })

  it('does not match generic errors', () => {
    assert.equal(isAuthExpiredError(new Error('服务内部错误')), false)
  })
})
