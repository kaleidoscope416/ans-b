import assert from 'node:assert/strict'
import { describe, it } from 'node:test'

import { formatDateTime, parseTags, submissionStatusText } from './student.js'

describe('parseTags', () => {
  it('splits tags by commas and line breaks', () => {
    assert.deepEqual(parseTags('图书馆, 时间，借阅\n夜间'), ['图书馆', '时间', '借阅', '夜间'])
  })
})

describe('submissionStatusText', () => {
  it('maps known statuses to Chinese labels', () => {
    assert.equal(submissionStatusText('approved'), '已通过')
  })
})

describe('formatDateTime', () => {
  it('returns fallback for empty values', () => {
    assert.equal(formatDateTime(''), '-')
  })
})
