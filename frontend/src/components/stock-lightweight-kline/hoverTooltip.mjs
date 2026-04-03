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
    placement: `${horizontal}-${vertical}`,
  }
}
