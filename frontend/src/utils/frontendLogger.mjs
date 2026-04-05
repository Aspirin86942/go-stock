import { EventsEmit } from '../../wailsjs/runtime/runtime.js'

let handlersInstalled = false

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

export function buildFrontendErrorPayload({ page, route, message, source, lineno, colno, error, extra = {} }) {
  return {
    page,
    route: resolveRoute(route),
    message: resolveMessage(message, error),
    source: source || '',
    lineno: Number.isFinite(lineno) ? lineno : 0,
    colno: Number.isFinite(colno) ? colno : 0,
    error: resolveError(error),
    extra,
  }
}

export function emitFrontendError(input) {
  const payload = buildFrontendErrorPayload(input)
  if (isResizeObserverNoise(payload.message) || isResizeObserverNoise(payload.error)) {
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
}
