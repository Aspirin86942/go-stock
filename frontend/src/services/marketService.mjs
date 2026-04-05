import * as AppBindings from '../../wailsjs/go/main/App.js';

function toArray(value) {
  return Array.isArray(value) ? value : [];
}

export function normalizeMarketFeeds(value) {
  const data = value ?? {};
  return {
    telegraph: toArray(data.telegraph),
    sina: toArray(data.sina),
    foreign: toArray(data.foreign),
  };
}

export function normalizeMarketFeed(value) {
  const data = value ?? {};
  return {
    source: data.source ?? '',
    items: toArray(data.items),
  };
}

export function normalizeMarketIndexes(value) {
  const data = value ?? {};
  return {
    common: toArray(data.common),
    america: toArray(data.america),
    europe: toArray(data.europe),
    asia: toArray(data.asia),
    other: toArray(data.other),
  };
}

export function normalizeMarketIndustryRanks(value) {
  return toArray(value);
}

export async function loadMarketFeeds() {
  const result = await AppBindings.GetMarketFeeds();
  return normalizeMarketFeeds(result);
}

export async function loadMarketGlobalIndexes() {
  const result = await AppBindings.GetMarketGlobalIndexes();
  return normalizeMarketIndexes(result);
}

export async function loadMarketIndustryRanks(sort = '0', count = 150) {
  const result = await AppBindings.GetMarketIndustryRanks(sort, count);
  return normalizeMarketIndustryRanks(result);
}

export async function refreshMarketFeed(source) {
  const result = await AppBindings.RefreshMarketFeed(source);
  return normalizeMarketFeed(result);
}
