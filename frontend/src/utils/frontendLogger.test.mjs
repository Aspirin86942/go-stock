import test from 'node:test'
import assert from 'node:assert/strict'

import {
  buildFrontendErrorPayload,
  installFrontendErrorHandlers,
  __resetFrontendErrorHandlersForTest,
} from './frontendLogger.mjs'

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

test('buildFrontendErrorPayload auto fills route from window location', () => {
  const previousWindow = globalThis.window
  globalThis.window = {
    location: {
      hash: '#/settings',
      pathname: '/fallback',
    },
  }

  try {
    const payload = buildFrontendErrorPayload({
      page: 'settings.vue',
      message: 'oops',
      error: new Error('oops'),
    })
    assert.equal(payload.route, '#/settings')
  } finally {
    globalThis.window = previousWindow
  }
})

test('buildFrontendErrorPayload keeps traceId when provided', () => {
  const payload = buildFrontendErrorPayload({
    page: 'orders.vue',
    route: '/orders',
    message: 'failed',
    traceId: 'trace-123',
  })

  assert.equal(payload.traceId, 'trace-123')
})

test('installFrontendErrorHandlers is idempotent and registers listeners once', () => {
  __resetFrontendErrorHandlersForTest()
  const previousWindow = globalThis.window
  const listeners = new Map()
  globalThis.window = {
    location: { hash: '', pathname: '/stock' },
    runtime: { EventsEmit: () => null },
    addEventListener: (name, handler) => {
      if (!listeners.has(name)) {
        listeners.set(name, [])
      }
      listeners.get(name).push(handler)
    },
  }

  try {
    installFrontendErrorHandlers('main.js')
    installFrontendErrorHandlers('main.js')
    assert.equal((listeners.get('error') || []).length, 1)
    assert.equal((listeners.get('unhandledrejection') || []).length, 1)
  } finally {
    globalThis.window = previousWindow
    __resetFrontendErrorHandlersForTest()
  }
})

test('console.error proxy keeps original output and emits once for the same error', () => {
  __resetFrontendErrorHandlersForTest()
  const previousWindow = globalThis.window
  const previousConsoleError = console.error
  const listeners = new Map()
  const emitted = []
  const consoleLogs = []

  globalThis.window = {
    location: { hash: '#/stock', pathname: '/stock' },
    runtime: {
      EventsEmit: (...args) => {
        emitted.push(args)
      },
    },
    addEventListener: (name, handler) => {
      listeners.set(name, handler)
    },
  }
  console.error = (...args) => {
    consoleLogs.push(args)
  }

  try {
    installFrontendErrorHandlers('main.js')
    const repeatedError = new Error('same-console-error')

    console.error(repeatedError)
    console.error(repeatedError)

    assert.equal(consoleLogs.length, 2)
    assert.equal(emitted.length, 1)
    assert.equal(emitted[0][0], 'frontendError')
    assert.equal(emitted[0][1].message, 'same-console-error')
  } finally {
    console.error = previousConsoleError
    globalThis.window = previousWindow
    __resetFrontendErrorHandlersForTest()
  }
})

test('console.error proxy emits twice for same message with different context args', () => {
  __resetFrontendErrorHandlersForTest()
  const previousWindow = globalThis.window
  const previousConsoleError = console.error
  const listeners = new Map()
  const emitted = []
  const consoleLogs = []

  globalThis.window = {
    location: { hash: '#/stock', pathname: '/stock' },
    runtime: {
      EventsEmit: (...args) => {
        emitted.push(args)
      },
    },
    addEventListener: (name, handler) => {
      listeners.set(name, handler)
    },
  }
  console.error = (...args) => {
    consoleLogs.push(args)
  }

  try {
    installFrontendErrorHandlers('main.js')
    const logWithContext = (requestId) => {
      console.error('request failed', { requestId })
    }

    logWithContext(1)
    logWithContext(2)
    logWithContext(1)

    assert.equal(consoleLogs.length, 3)
    assert.equal(emitted.length, 2)
    assert.equal(emitted[0][0], 'frontendError')
    assert.equal(emitted[1][0], 'frontendError')
  } finally {
    console.error = previousConsoleError
    globalThis.window = previousWindow
    __resetFrontendErrorHandlersForTest()
  }
})

test('non-ResizeObserver rejection is emitted and not silently swallowed', () => {
  __resetFrontendErrorHandlersForTest()
  const previousWindow = globalThis.window
  const previousConsoleError = console.error
  const listeners = new Map()
  const emitted = []
  const consoleLogs = []
  globalThis.window = {
    location: { hash: '#/stock', pathname: '/stock' },
    runtime: {
      EventsEmit: (...args) => {
        emitted.push(args)
      },
    },
    addEventListener: (name, handler) => {
      listeners.set(name, handler)
    },
  }
  console.error = (...args) => {
    consoleLogs.push(args)
  }

  try {
    installFrontendErrorHandlers('main.js')
    const handler = listeners.get('unhandledrejection')
    assert.equal(typeof handler, 'function')

    let prevented = false
    const rejectionError = new Error('promise failed')
    handler({
      reason: rejectionError,
      preventDefault: () => {
        prevented = true
      },
    })

    assert.equal(prevented, false)
    assert.equal(emitted.length, 1)
    assert.equal(emitted[0][0], 'frontendError')
    assert.equal(consoleLogs.length, 1)
  } finally {
    console.error = previousConsoleError
    globalThis.window = previousWindow
    __resetFrontendErrorHandlersForTest()
  }
})
