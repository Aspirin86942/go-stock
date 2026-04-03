# Multi-KLine Hover Tooltip Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add explicit dashed crosshair lines, axis labels, and a mouse-adjacent compact tooltip to the stock-detail multi-period K-line modal while keeping the existing top information strip intact.

**Architecture:** Keep the backend and EastMoney data flow unchanged. Extract only the hover-tooltip visibility and placement math into a small ESM helper module covered by `node --test`, then wire that helper into `frontend/src/components/StockLightweightKlineChart.vue` so the chart gains explicit crosshair styling, tooltip state reset points, and a dedicated overlay wrapper that does not interfere with the `lightweight-charts` container.

**Tech Stack:** Vue 3, lightweight-charts 5.1.0, Naive UI, Node 22 `node:test`, Vite, Wails

---

## File Map

- Create: `D:\codex_work\go-stock\frontend\src\components\stock-lightweight-kline\hoverTooltip.mjs`
  - Responsibility: pure helper functions for tooltip visibility gating and edge-aware placement math
- Create: `D:\codex_work\go-stock\frontend\src\components\stock-lightweight-kline\hoverTooltip.test.mjs`
  - Responsibility: lightweight Node tests for main-pane gating and tooltip edge flipping
- Modify: `D:\codex_work\go-stock\frontend\src\components\StockLightweightKlineChart.vue`
  - Responsibility: explicit crosshair styling, tooltip state, crosshair handler integration, overlay template, theme-aware styles

### Task 1: Add pure hover-tooltip helpers with executable tests

**Files:**
- Create: `D:\codex_work\go-stock\frontend\src\components\stock-lightweight-kline\hoverTooltip.test.mjs`
- Create: `D:\codex_work\go-stock\frontend\src\components\stock-lightweight-kline\hoverTooltip.mjs`

- [ ] **Step 1: Write the failing Node tests first**

```js
import test from 'node:test'
import assert from 'node:assert/strict'

import {
  resolveHoverTooltipLayout,
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
```

- [ ] **Step 2: Run the tests to confirm the helper module does not exist yet**

Run:

```powershell
node --test frontend/src/components/stock-lightweight-kline/hoverTooltip.test.mjs
```

Expected:

- test run fails with `ERR_MODULE_NOT_FOUND` for `hoverTooltip.mjs`

- [ ] **Step 3: Create the helper module with only the placement and visibility logic**

```js
export const HOVER_TOOLTIP_OFFSET_X = 16
export const HOVER_TOOLTIP_OFFSET_Y = 16
export const HOVER_TOOLTIP_MARGIN = 8
export const HOVER_TOOLTIP_WIDTH = 180
export const HOVER_TOOLTIP_HEIGHT = 132

export function shouldShowHoverTooltip({ point, time, paneIndex, hasBar }) {
  if (!point || typeof point.x !== 'number' || typeof point.y !== 'number') {
    return false
  }
  if (time === undefined || time === null) {
    return false
  }
  if (!hasBar) {
    return false
  }
  return paneIndex == null || paneIndex === 0
}

export function resolveHoverTooltipLayout({
  pointX,
  pointY,
  containerWidth,
  containerHeight,
  tooltipWidth = HOVER_TOOLTIP_WIDTH,
  tooltipHeight = HOVER_TOOLTIP_HEIGHT,
  offsetX = HOVER_TOOLTIP_OFFSET_X,
  offsetY = HOVER_TOOLTIP_OFFSET_Y,
  margin = HOVER_TOOLTIP_MARGIN,
}) {
  let left = pointX + offsetX
  let top = pointY + offsetY
  let horizontal = 'right'
  let vertical = 'bottom'

  if (left + tooltipWidth + margin > containerWidth) {
    left = pointX - tooltipWidth - offsetX
    horizontal = 'left'
  }
  if (left < margin) {
    left = margin
  }

  if (top + tooltipHeight + margin > containerHeight) {
    top = pointY - tooltipHeight - offsetY
    vertical = 'top'
  }
  if (top < margin) {
    top = margin
  }

  return {
    left,
    top,
    placement: `${horizontal}-${vertical}`, // indicates the flip direction chosen before the margin clamp ensures the tooltip stays inside the chart
  }
}
```

- [ ] **Step 4: Re-run the helper tests and verify they pass**

Run:

```powershell
node --test frontend/src/components/stock-lightweight-kline/hoverTooltip.test.mjs
```

Expected:

- all 5 tests pass
- no snapshot files or extra config files are created

- [ ] **Step 5: Commit the helper slice**

```bash
git add frontend/src/components/stock-lightweight-kline/hoverTooltip.mjs frontend/src/components/stock-lightweight-kline/hoverTooltip.test.mjs
git commit -m "test: add multikline hover tooltip helpers"
```

### Task 2: Wire explicit crosshair styling and tooltip state into the chart component

**Files:**
- Modify: `D:\codex_work\go-stock\frontend\src\components\StockLightweightKlineChart.vue`
- Test: `D:\codex_work\go-stock\frontend\src\components\stock-lightweight-kline\hoverTooltip.test.mjs`

- [ ] **Step 1: Import the helper module and add tooltip state with a clear-reset helper**

```js
import {
  CrosshairMode,
  CandlestickSeries,
  createChart,
  HistogramSeries,
  LineSeries,
  LineStyle,
  TickMarkType,
} from 'lightweight-charts'

import {
  HOVER_TOOLTIP_HEIGHT,
  HOVER_TOOLTIP_WIDTH,
  resolveHoverTooltipLayout,
  shouldShowHoverTooltip,
} from './stock-lightweight-kline/hoverTooltip.mjs'

const hoverTooltipVisible = ref(false)
const hoverTooltipLeft = ref(0)
const hoverTooltipTop = ref(0)
const hoverTooltipPlacement = ref('right-bottom')

const hoverTooltipPanel = computed(() => {
  if (!hoverRawRow.value) return null
  const panel = crosshairPanel.value
  if (!panel) return null
  return {
    title: panel.title,
    open: panel.open,
    close: panel.close,
    high: panel.high,
    low: panel.low,
    changePercent: panel.changePercent,
    volume: panel.volume,
    cOpenClose: panel.cOpenClose,
    cHigh: panel.cHigh,
    cLow: panel.cLow,
    cChg: panel.cChg,
    cNeu: panel.cNeu,
  }
})

function clearHoverTooltip() {
  hoverTooltipVisible.value = false
  hoverTooltipLeft.value = 0
  hoverTooltipTop.value = 0
  hoverTooltipPlacement.value = 'right-bottom'
}
```

- [ ] **Step 2: Make the chart crosshair explicitly use dashed lines and axis labels**

```js
function chartThemeOptions(isDark) {
  const minuteLike = !DAILY_LIKE_KLT.has(activeKlt.value)
  const crosshairColor = isDark ? '#64748b' : '#94a3b8'
  const labelBg = isDark ? '#334155' : '#e2e8f0'

  return {
    layout: {
      background: { type: 'solid', color: isDark ? '#141414' : '#ffffff' },
      textColor: isDark ? '#cbd5e1' : '#334155',
    },
    grid: {
      vertLines: { color: isDark ? '#27272a' : '#f1f5f9' },
      horzLines: { color: isDark ? '#27272a' : '#f1f5f9' },
    },
    crosshair: {
      mode: CrosshairMode.Normal,
      vertLine: {
        visible: true,
        labelVisible: true,
        width: 1,
        style: LineStyle.LargeDashed,
        color: crosshairColor,
        labelBackgroundColor: labelBg,
      },
      horzLine: {
        visible: true,
        labelVisible: true,
        width: 1,
        style: LineStyle.LargeDashed,
        color: crosshairColor,
        labelBackgroundColor: labelBg,
      },
    },
    rightPriceScale: { borderColor: isDark ? '#3f3f46' : '#e2e8f0' },
    localization: {
      locale: 'zh-CN',
      dateFormat: 'yyyy-MM-dd',
      timeFormatter: (t) => formatCrosshairTime(t),
    },
    timeScale: {
      borderColor: isDark ? '#3f3f46' : '#e2e8f0',
      timeVisible: minuteLike,
      secondsVisible: false,
      tickMarkFormatter: (t, tickMarkType) => formatTickTime(t, tickMarkType),
    },
  }
}
```

- [ ] **Step 3: Update the crosshair move handler so the tooltip only appears on the main K-line pane and resets on every invalid state**

```js
crosshairMoveHandler = (param) => {
  if (param.point === undefined) {
    clearHoverTooltip()
    hoverRawRow.value = null
    clearLongPriceLinePaneCursor()
    return
  }

  refreshLongPriceLineCursorFromCrosshair(param)

  const bar = param.time === undefined ? null : param.seriesData.get(candleSeries)
  if (
    !shouldShowHoverTooltip({
      point: param.point,
      time: param.time,
      paneIndex: param.paneIndex,
      hasBar: !!bar,
    })
  ) {
    clearHoverTooltip()
    hoverRawRow.value = null
    return
  }

  const rawRow = findRawRowByChartTime(param.time)
  hoverRawRow.value = rawRow
  if (!rawRow || !chartContainerRef.value) {
    clearHoverTooltip()
    return
  }

  const layout = resolveHoverTooltipLayout({
    pointX: param.point.x,
    pointY: param.point.y,
    containerWidth: chartContainerRef.value.clientWidth,
    containerHeight: chartContainerRef.value.clientHeight,
    tooltipWidth: HOVER_TOOLTIP_WIDTH,
    tooltipHeight: HOVER_TOOLTIP_HEIGHT,
  })

  hoverTooltipLeft.value = layout.left
  hoverTooltipTop.value = layout.top
  hoverTooltipPlacement.value = layout.placement
  hoverTooltipVisible.value = true
}
```

- [ ] **Step 4: Clear the tooltip whenever the component reloads or tears down chart state**

```js
function disposeChart() {
  clearPoll()
  clearHoverTooltip()
  // existing cleanup remains here
}

async function loadData() {
  clearHoverTooltip()
  hoverRawRow.value = null
  // existing loadData body remains here
}

watch(
  () => props.code,
  () => {
    clearHoverTooltip()
    hoverRawRow.value = null
    loadData()
    setupPoll()
  },
)

watch(activeKlt, () => {
  clearHoverTooltip()
  hoverRawRow.value = null
  chart?.applyOptions(chartThemeOptions(props.darkTheme))
  loadData()
  setupPoll()
})
```

- [ ] **Step 5: Re-run the helper tests and the frontend build**

Run:

```powershell
node --test frontend/src/components/stock-lightweight-kline/hoverTooltip.test.mjs
npm --prefix frontend run build
```

Expected:

- Node helper tests still pass
- Vite build exits `0`
- no import or template compile error is introduced by the new tooltip state

- [ ] **Step 6: Commit the component logic slice**

```bash
git add frontend/src/components/StockLightweightKlineChart.vue
git commit -m "feat: wire multikline hover tooltip state"
```

### Task 3: Render the overlay tooltip and verify the full interaction in Wails

**Files:**
- Modify: `D:\codex_work\go-stock\frontend\src\components\StockLightweightKlineChart.vue`
- Verify only: `D:\codex_work\go-stock\frontend\src\components\stock-lightweight-kline\hoverTooltip.test.mjs`

- [ ] **Step 1: Wrap the chart container so the tooltip overlay is rendered outside the library-owned canvas container**

```vue
<div
  class="lw-kline-chart-shell"
  :style="{ height: chartHeight + 'px', minHeight: chartHeight + 'px' }"
>
  <div
    ref="chartContainerRef"
    class="lw-kline-chart"
    :style="{ height: '100%', minHeight: '100%' }"
  />

  <div
    v-if="hoverTooltipVisible && hoverTooltipPanel"
    class="lw-kline-hover-tooltip"
    :class="[
      darkTheme ? 'lw-kline-hover-tooltip--dark' : '',
      `lw-kline-hover-tooltip--${hoverTooltipPlacement}`,
    ]"
    :style="{
      left: hoverTooltipLeft + 'px',
      top: hoverTooltipTop + 'px',
    }"
  >
    <div class="lw-kline-hover-tooltip__title">{{ hoverTooltipPanel.title }}</div>
    <div class="lw-kline-hover-tooltip__grid">
      <span class="lw-kline-kv">
        <span class="lw-kline-hover-tooltip__k">开</span>
        <span class="lw-kline-hover-tooltip__v" :style="{ color: hoverTooltipPanel.cOpenClose }">{{ hoverTooltipPanel.open }}</span>
      </span>
      <span class="lw-kline-kv">
        <span class="lw-kline-hover-tooltip__k">高</span>
        <span class="lw-kline-hover-tooltip__v" :style="{ color: hoverTooltipPanel.cHigh }">{{ hoverTooltipPanel.high }}</span>
      </span>
      <span class="lw-kline-kv">
        <span class="lw-kline-hover-tooltip__k">低</span>
        <span class="lw-kline-hover-tooltip__v" :style="{ color: hoverTooltipPanel.cLow }">{{ hoverTooltipPanel.low }}</span>
      </span>
      <span class="lw-kline-kv">
        <span class="lw-kline-hover-tooltip__k">收</span>
        <span class="lw-kline-hover-tooltip__v" :style="{ color: hoverTooltipPanel.cOpenClose }">{{ hoverTooltipPanel.close }}</span>
      </span>
      <span class="lw-kline-kv">
        <span class="lw-kline-hover-tooltip__k">涨跌幅</span>
        <span class="lw-kline-hover-tooltip__v" :style="{ color: hoverTooltipPanel.cChg }">{{ hoverTooltipPanel.changePercent }}</span>
      </span>
      <span class="lw-kline-kv">
        <span class="lw-kline-hover-tooltip__k">成交量</span>
        <span class="lw-kline-hover-tooltip__v" :style="{ color: hoverTooltipPanel.cNeu }">{{ hoverTooltipPanel.volume }}</span>
      </span>
    </div>
  </div>
</div>
```

- [ ] **Step 2: Add dedicated tooltip and wrapper styles without breaking the existing chart sizing**

```css
.lw-kline-chart-shell {
  width: 100%;
  max-width: 100%;
  min-width: 0;
  position: relative;
}

.lw-kline-chart {
  width: 100%;
  max-width: 100%;
  min-width: 0;
  height: 100%;
  position: relative;
  touch-action: none;
  box-sizing: border-box;
}

.lw-kline-hover-tooltip {
  position: absolute;
  z-index: 5;
  width: 180px;
  padding: 8px 10px;
  border-radius: 8px;
  border: 1px solid #cbd5e1;
  background: rgba(255, 255, 255, 0.96);
  box-shadow: 0 10px 30px rgba(15, 23, 42, 0.18);
  pointer-events: none;
  backdrop-filter: blur(6px);
}

.lw-kline-hover-tooltip--dark {
  border-color: #475569;
  background: rgba(15, 23, 42, 0.94);
  box-shadow: 0 10px 30px rgba(2, 6, 23, 0.45);
}

.lw-kline-hover-tooltip__title {
  margin-bottom: 6px;
  font-size: 12px;
  font-weight: 700;
  color: #0f172a;
}

.lw-kline-hover-tooltip--dark .lw-kline-hover-tooltip__title {
  color: #f8fafc;
}

.lw-kline-hover-tooltip__grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 4px 10px;
  font-size: 11px;
}

.lw-kline-hover-tooltip__k {
  color: #64748b;
}

.lw-kline-hover-tooltip--dark .lw-kline-hover-tooltip__k {
  color: #94a3b8;
}

.lw-kline-hover-tooltip__v {
  font-variant-numeric: tabular-nums;
}
```

- [ ] **Step 3: Re-run the helper tests and frontend build after the template/style changes**

Run:

```powershell
node --test frontend/src/components/stock-lightweight-kline/hoverTooltip.test.mjs
npm --prefix frontend run build
```

Expected:

- helper tests still pass
- frontend build exits `0`
- Vue template compiles cleanly with the new chart wrapper and overlay DOM

- [ ] **Step 4: Run the desktop app and manually verify the hover interaction**

Run:

```powershell
wails dev
```

Manual verification checklist:

- open a stock detail window and click `多周期K线`
- hover the main candlestick pane and confirm dashed crosshair lines are visible
- confirm the price axis and time axis both show crosshair labels
- confirm a compact tooltip appears near the cursor with `时间 / 开 / 高 / 低 / 收 / 涨跌幅 / 成交量`
- move the cursor onto the volume pane or any sub-indicator pane and confirm the tooltip hides
- move the cursor near the chart right edge and top edge and confirm the tooltip flips but stays inside the chart shell
- switch `1分 / 5分 / 日K / 周K` and repeat the hover check
- drag left to load more history, then hover older candles and confirm the tooltip still follows correctly
- toggle dark theme and verify tooltip contrast remains readable

Expected:

- all checks pass
- no console error appears for the tooltip logic
- existing long-position lines, indicators, and top information strip still work

- [ ] **Step 5: Commit the rendered UI slice**

```bash
git add frontend/src/components/StockLightweightKlineChart.vue
git commit -m "feat: add multikline hover tooltip UI"
```

## Self-Review Notes

- Spec coverage:
  - explicit dashed crosshair and axis labels are covered in Task 2
  - mouse-adjacent compact tooltip is covered in Task 2 and Task 3
  - main-pane-only visibility and edge flipping are covered by Task 1 tests and Task 2 integration
  - keeping the top strip intact is preserved because the plan only adds overlay state and DOM
- Placeholder scan:
  - no placeholder wording or shorthand cross-references remain
  - every code-changing step includes concrete code
- Type consistency:
  - helper names stay consistent as `shouldShowHoverTooltip`, `resolveHoverTooltipLayout`, and `clearHoverTooltip`
  - tooltip state names stay consistent as `hoverTooltipVisible`, `hoverTooltipLeft`, `hoverTooltipTop`, `hoverTooltipPlacement`, and `hoverTooltipPanel`
