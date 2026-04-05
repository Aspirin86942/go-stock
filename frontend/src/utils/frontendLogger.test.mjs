import test from 'node:test'
import assert from 'node:assert/strict'

import { buildFrontendErrorPayload } from './frontendLogger.mjs'

test('buildFrontendErrorPayload normalizes page route and stack', () => {
  const payload = buildFrontendErrorPayload({
    page: 'stock.vue',
    route: '/stock',
    message: 'boom',
    error: new Error('boom'),
    extra: { panel: 'watchlist' },
  })

  assert.equal(payload.page, 'stock.vue')
  assert.equal(payload.route, '/stock')
  assert.match(payload.error, /boom/)
  assert.deepEqual(payload.extra, { panel: 'watchlist' })
})
