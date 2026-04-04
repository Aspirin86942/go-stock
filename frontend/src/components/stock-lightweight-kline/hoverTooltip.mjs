export const HOVER_TOOLTIP_OFFSET_X = 16
export const HOVER_TOOLTIP_OFFSET_Y = 16
export const HOVER_TOOLTIP_MARGIN = 8
export const HOVER_TOOLTIP_WIDTH = 180
export const HOVER_TOOLTIP_HEIGHT = 132

export function resolveMainPaneVerticalBounds({
  containerHeight,
  mainPaneHeight,
}) {
  if (Number.isFinite(mainPaneHeight) && mainPaneHeight > 0) {
    const bottom = Number.isFinite(containerHeight)
      ? Math.min(mainPaneHeight, containerHeight)
      : mainPaneHeight
    return {
      mainPaneTop: 0,
      mainPaneBottom: bottom,
    }
  }

  if (Number.isFinite(containerHeight) && containerHeight > 0) {
    return {
      mainPaneTop: 0,
      mainPaneBottom: containerHeight,
    }
  }

  return {
    mainPaneTop: null,
    mainPaneBottom: null,
  }
}

export function shouldShowHoverTooltip({
  point,
  time,
  paneIndex,
  hasBar,
  containerWidth,
  containerHeight,
  mainPaneTop,
  mainPaneBottom,
}) {
  if (!point || typeof point.x !== 'number' || typeof point.y !== 'number') {
    return false
  }
  if (time === undefined || time === null) {
    return false
  }
  if (!hasBar) {
    return false
  }
  if (
    Number.isFinite(containerWidth) &&
    (point.x < 0 || point.x >= containerWidth)
  ) {
    return false
  }
  if (
    Number.isFinite(containerHeight) &&
    (point.y < 0 || point.y >= containerHeight)
  ) {
    return false
  }
  if (
    Number.isFinite(mainPaneTop) &&
    Number.isFinite(mainPaneBottom) &&
    (point.y < mainPaneTop || point.y >= mainPaneBottom)
  ) {
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
    /** horizontal-vertical values reflect the flip direction chosen before the margin clamp */
    placement: `${horizontal}-${vertical}`,
  }
}
