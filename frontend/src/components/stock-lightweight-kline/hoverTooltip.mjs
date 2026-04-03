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

  const maxRight = containerWidth - tooltipWidth - margin
  const overflowRight = left + tooltipWidth + margin - containerWidth
  const flipLeft = pointX - tooltipWidth - offsetX

  if (overflowRight > 0) {
    if (flipLeft >= margin) {
      left = flipLeft
      horizontal = 'left'
    } else {
      const shift = Math.max(margin, overflowRight)
      left = left - shift
    }
  }

  if (maxRight >= margin && left > maxRight) {
    left = maxRight
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
