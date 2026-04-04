function convertTsCodeToMarketCode(tsCode) {
  if (typeof tsCode !== 'string') {
    return null;
  }

  const normalized = tsCode.trim();
  if (!normalized) {
    return null;
  }

  const match = normalized.match(/^([0-9A-Za-z_]+)\.([A-Za-z]+)$/);
  if (!match) {
    return null;
  }

  const [, code, market] = match;
  return `${market.toLowerCase()}${code.toLowerCase()}`;
}

export function normalizeFollowStockCode(value, stockList = []) {
  if (typeof value !== 'string') {
    return null;
  }

  const normalized = value.trim();
  if (!normalized) {
    return null;
  }

  const labelMatch = normalized.match(/^(.+?)\s+-\s+([0-9A-Za-z_.-]+\.[A-Za-z]+)$/);
  if (labelMatch) {
    const [, , tsCode] = labelMatch;
    const normalizedFromLabel = convertTsCodeToMarketCode(tsCode);
    if (normalizedFromLabel) {
      return normalizedFromLabel;
    }
  }

  if (/^(sh|sz|bj|hk)\d+$/i.test(normalized) || /^gb_[0-9a-z._-]+$/i.test(normalized) || /^us[0-9a-z._-]+$/i.test(normalized)) {
    return normalized.toLowerCase();
  }

  const tsCodeAsMarketCode = convertTsCodeToMarketCode(normalized);
  if (tsCodeAsMarketCode) {
    return tsCodeAsMarketCode;
  }

  const exactMatch = Array.isArray(stockList)
    ? stockList.find((item) => item?.name === normalized || item?.ts_code === normalized)
    : null;
  if (!exactMatch) {
    return null;
  }

  return convertTsCodeToMarketCode(exactMatch.ts_code);
}

export function resolveFollowStockCode(values, stockList = []) {
  if (!Array.isArray(values)) {
    return null;
  }

  for (const value of values) {
    const normalized = normalizeFollowStockCode(value, stockList);
    if (normalized) {
      return normalized;
    }
  }

  return null;
}
