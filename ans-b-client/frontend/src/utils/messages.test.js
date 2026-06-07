import assert from 'node:assert/strict'
import { describe, it } from 'node:test'

import { replaceMessage } from './messages.js'

describe('replaceMessage', () => {
  it('replaces a message object so Vue sees a new array value', () => {
    const pending = { id: 2, content: '正在检索', loading: true }
    const messages = [
      { id: 1, content: '你好' },
      pending,
    ]

    const next = replaceMessage(messages, 2, {
      content: '二食堂晚餐到 21:00。',
      loading: false,
    })

    assert.notEqual(next, messages)
    assert.notEqual(next[1], pending)
    assert.deepEqual(next[1], {
      id: 2,
      content: '二食堂晚餐到 21:00。',
      loading: false,
    })
    assert.equal(next[0], messages[0])
  })
})
