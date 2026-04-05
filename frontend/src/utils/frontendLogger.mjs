import { EventsEmit } from '../../wailsjs/runtime/runtime.js'

let handlersInstalled = false
let originalConsoleError = null
let recentErrorObjects = new WeakMap()
const recentFingerprints = new Map()

const ERROR_DEDUPE_WINDOW_MS = 1200

export function isResizeObserverNoise(value) {
  if (!value) {
    return false
  }
  return String(value).includes('ResizeObserver')
}

function resolveRoute(route) {
  if (route) {
    return route
  }
  if (typeof window === 'undefined') {
    return ''
  }
  return window.location.hash || window.location.pathname || ''
}

function resolveMessage(message, error) {
  if (typeof message === 'string') {
    return message
  }
  if (message != null) {
    return String(message)
  }
  if (error?.message) {
    return error.message
  }
  return 'unknown frontend error'
}

function resolveError(error) {
  if (!error) {
    return null
  }
  if (error?.stack) {
    return error.stack
  }
  if (error?.message) {
    return error.message
  }
  return String(error)
}

function cleanupExpiredFingerprints(nowMs) {
  for (const [fingerprint, lastAt] of recentFingerprints.entries()) {
    if (nowMs - lastAt > ERROR_DEDUPE_WINDOW_MS) {
      recentFingerprints.delete(fingerprint)
    }
  }
}

function buildErrorFingerprint(payload) {
  return [
    payload.page || '',
    payload.route || '',
    payload.message || '',
    payload.source || '',
    payload.lineno || 0,
    payload.colno || 0,
    payload.error || '',
    payload.traceId || '',
  ].join('|')
}

function shouldDedupeFrontendError(error, payload) {
  const nowMs = Date.now()
  cleanupExpiredFingerprints(nowMs)

  if (error && (typeof error === 'object' || typeof error === 'function')) {
    const lastAt = recentErrorObjects.get(error)
    recentErrorObjects.set(error, nowMs)
    if (typeof lastAt === 'number' && nowMs - lastAt <= ERROR_DEDUPE_WINDOW_MS) {
      return true
    }
  }

  const fingerprint = buildErrorFingerprint(payload)
  const lastFingerprintAt = recentFingerprints.get(fingerprint)
  recentFingerprints.set(fingerprint, nowMs)
  return typeof lastFingerprintAt === 'number' && nowMs - lastFingerprintAt <= ERROR_DEDUPE_WINDOW_MS
}

function buildConsoleErrorInput(page, args) {
  if (!Array.isArray(args) || args.length === 0) {
    return null
  }

  const error = args.find((item) => item instanceof Error) || null
  let message = null

  if (error?.message) {
    message = error.message
  } else if (typeof args[0] === 'string') {
    message = args[0]
  } else if (args[0] != null) {
    message = String(args[0])
  }

  return {
    page,
    message: message || 'console.error',
    error: error || new Error((message || 'console.error').toString()),
  }
}

function installConsoleErrorProxy(page) {
  if (typeof console === 'undefined' || typeof console.error !== 'function' || originalConsoleError) {
    return
  }

  originalConsoleError = console.error
  console.error = (...args) => {
    originalConsoleError.apply(console, args)

    const input = buildConsoleErrorInput(page, args)
    if (!input) {
      return
    }
    emitFrontendError(input)
  }
}

export function buildFrontendErrorPayload({ page, route, message, source, lineno, colno, error, traceId, extra = {} }) {
  return {
    page,
    route: resolveRoute(route),
    message: resolveMessage(message, error),
    source: source || '',
    lineno: Number.isFinite(lineno) ? lineno : 0,
    colno: Number.isFinite(colno) ? colno : 0,
    error: resolveError(error),
    traceId: traceId || '',
    extra,
  }
}

export function emitFrontendError(input) {
  const payload = buildFrontendErrorPayload(input)
  if (isResizeObserverNoise(payload.message) || isResizeObserverNoise(payload.error)) {
    return false
  }
  if (shouldDedupeFrontendError(input?.error, payload)) {
    return false
  }
  EventsEmit('frontendError', payload)
  return true
}

export function installFrontendErrorHandlers(page) {
  if (typeof window === 'undefined' || handlersInstalled) {
    return
  }
  handlersInstalled = true
  installConsoleErrorProxy(page)

  window.addEventListener('error', (event) => {
    const message = event?.message
    const error = event?.error
    if (isResizeObserverNoise(message) || isResizeObserverNoise(error?.message) || isResizeObserverNoise(error?.stack)) {
      event.preventDefault?.()
      return
    }
    emitFrontendError({
      page,
      message,
      source: event?.filename,
      lineno: event?.lineno,
      colno: event?.colno,
      error,
    })
  })

  window.addEventListener('unhandledrejection', (event) => {
    const reason = event?.reason
    const reasonMessage = reason?.message || reason
    if (isResizeObserverNoise(reasonMessage)) {
      event.preventDefault()
      return
    }
    const emitted = emitFrontendError({
      page,
      message: reason?.message || 'unhandledrejection',
      error: reason instanceof Error ? reason : new Error(String(reason)),
    })
    if (emitted) {
      console.error('Unhandled promise rejection:', reason)
    }
  })
}

export function __resetFrontendErrorHandlersForTest() {
  handlersInstalled = false
  if (originalConsoleError) {
    console.error = originalConsoleError
    originalConsoleError = null
  }
  recentErrorObjects = new WeakMap()
  recentFingerprints.clear()
}
