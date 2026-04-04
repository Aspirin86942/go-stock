import test from 'node:test'
import assert from 'node:assert/strict'

import {
  resolveHoverTooltipLayout,
  resolveMainPaneVerticalBounds,
  shouldShowHoverTooltip,
} from './hoverTooltip.mjs'

test('shouldShowHoverTooltip returns true only for the main K-line pane', () => {
  assert.equal(
    shouldShowHoverTooltip({
      point: { x: 120, y: 90 },
      time: 1712219400,
      paneIndex: 0,
      hasBar: true,
    }),
    true,
  )

  assert.equal(
    shouldShowHoverTooltip({
      point: { x: 120, y: 90 },
      time: 1712219400,
      paneIndex: 1,
      hasBar: true,
    }),
    false,
  )

  assert.equal(
    shouldShowHoverTooltip({
      point: { x: 120, y: 90 },
      time: undefined,
      paneIndex: 0,
      hasBar: true,
    }),
    false,
  )
})

test('shouldShowHoverTooltip returns false when point is outside chart shell bounds', () => {
  assert.equal(
    shouldShowHoverTooltip({
      point: { x: -1, y: 40 },
      time: 1712219400,
      paneIndex: 0,
      hasBar: true,
      containerWidth: 520,
      containerHeight: 360,
      mainPaneTop: 0,
      mainPaneBottom: 320,
    }),
    false,
  )

  assert.equal(
    shouldShowHoverTooltip({
      point: { x: 520, y: 40 },
      time: 1712219400,
      paneIndex: 0,
      hasBar: true,
      containerWidth: 520,
      containerHeight: 360,
      mainPaneTop: 0,
      mainPaneBottom: 320,
    }),
    false,
  )
})

test('shouldShowHoverTooltip returns false when point is outside main pane vertical region', () => {
  assert.equal(
    shouldShowHoverTooltip({
      point: { x: 120, y: 18 },
      time: 1712219400,
      paneIndex: 0,
      hasBar: true,
      containerWidth: 520,
      containerHeight: 360,
      mainPaneTop: 20,
      mainPaneBottom: 280,
    }),
    false,
  )

  assert.equal(
    shouldShowHoverTooltip({
      point: { x: 120, y: 281 },
      time: 1712219400,
      paneIndex: 0,
      hasBar: true,
      containerWidth: 520,
      containerHeight: 360,
      mainPaneTop: 20,
      mainPaneBottom: 280,
    }),
    false,
  )
})

test('shouldShowHoverTooltip accepts point at main pane top boundary', () => {
  assert.equal(
    shouldShowHoverTooltip({
      point: { x: 120, y: 20 },
      time: 1712219400,
      paneIndex: 0,
      hasBar: true,
      containerWidth: 520,
      containerHeight: 360,
      mainPaneTop: 20,
      mainPaneBottom: 280,
    }),
    true,
  )
})

test('shouldShowHoverTooltip accepts point at main pane bottom minus one boundary', () => {
  assert.equal(
    shouldShowHoverTooltip({
      point: { x: 120, y: 279 },
      time: 1712219400,
      paneIndex: 0,
      hasBar: true,
      containerWidth: 520,
      containerHeight: 360,
      mainPaneTop: 20,
      mainPaneBottom: 280,
    }),
    true,
  )
})

test('resolveMainPaneVerticalBounds returns content-area aligned bounds from pane height', () => {
  assert.deepEqual(
    resolveMainPaneVerticalBounds({
      containerHeight: 360,
      mainPaneHeight: 320,
    }),
    {
      mainPaneTop: 0,
      mainPaneBottom: 320,
    },
  )
})

test('resolveHoverTooltipLayout keeps the tooltip on the default side when there is room', () => {
  assert.deepEqual(
    resolveHoverTooltipLayout({
      pointX: 120,
      pointY: 140,
      containerWidth: 520,
      containerHeight: 360,
      tooltipWidth: 180,
      tooltipHeight: 132,
    }),
    {
      left: 136,
      top: 156,
      placement: 'right-bottom',
    },
  )
})

test('resolveHoverTooltipLayout flips left near the right edge', () => {
  assert.deepEqual(
    resolveHoverTooltipLayout({
      pointX: 470,
      pointY: 160,
      containerWidth: 520,
      containerHeight: 360,
      tooltipWidth: 180,
      tooltipHeight: 132,
    }),
    {
      left: 274,
      top: 176,
      placement: 'left-bottom',
    },
  )
})

test('resolveHoverTooltipLayout flips downward near the top edge and clamps into the chart area', () => {
  assert.deepEqual(
    resolveHoverTooltipLayout({
      pointX: 40,
      pointY: 12,
      containerWidth: 500,
      containerHeight: 150,
      tooltipWidth: 180,
      tooltipHeight: 132,
    }),
    {
      left: 56,
      top: 8,
      placement: 'right-top',
    },
  )
})

test('resolveHoverTooltipLayout keeps placement hints even when margin clamp shifts left', () => {
  assert.deepEqual(
    resolveHoverTooltipLayout({
      pointX: 40,
      pointY: 100,
      containerWidth: 200,
      containerHeight: 300,
      tooltipWidth: 180,
      tooltipHeight: 132,
    }),
    {
      left: 8,
      top: 116,
      placement: 'left-bottom',
    },
  )
})
