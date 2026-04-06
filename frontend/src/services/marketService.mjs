import * as AppBindings from '../../wailsjs/go/main/App.js';

function toArray(value) {
  return Array.isArray(value) ? value : [];
}

function toNumber(value, fallback = 0) {
  const normalized = Number(value);
  return Number.isFinite(normalized) ? normalized : fallback;
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

export function normalizeIndustryMoneyRanks(value) {
  return toArray(value).map((item) => ({
    category: item?.category ?? '',
    name: item?.name ?? '',
    avg_changeratio: toNumber(item?.avg_changeratio, 0),
    inamount: toNumber(item?.inamount, 0),
    outamount: toNumber(item?.outamount, 0),
    netamount: toNumber(item?.netamount, 0),
    ratioamount: toNumber(item?.ratioamount, 0),
    ts_name: item?.ts_name ?? '',
    ts_symbol: item?.ts_symbol ?? '',
    ts_changeratio: toNumber(item?.ts_changeratio, 0),
    ts_trade: toNumber(item?.ts_trade, 0),
    ts_ratioamount: toNumber(item?.ts_ratioamount, 0),
  }));
}

export function normalizeMoneyRanks(value) {
  return toArray(value).map((item) => ({
    symbol: item?.symbol ?? '',
    name: item?.name ?? '',
    trade: toNumber(item?.trade, 0),
    changeratio: toNumber(item?.changeratio, 0),
    turnover: toNumber(item?.turnover, 0),
    amount: toNumber(item?.amount, 0),
    outamount: toNumber(item?.outamount, 0),
    inamount: toNumber(item?.inamount, 0),
    netamount: toNumber(item?.netamount, 0),
    ratioamount: toNumber(item?.ratioamount, 0),
    r0_out: toNumber(item?.r0_out, 0),
    r0_in: toNumber(item?.r0_in, 0),
    r0_net: toNumber(item?.r0_net, 0),
    r0_ratio: toNumber(item?.r0_ratio, 0),
    r3_out: toNumber(item?.r3_out, 0),
    r3_in: toNumber(item?.r3_in, 0),
    r3_net: toNumber(item?.r3_net, 0),
    r3_ratio: toNumber(item?.r3_ratio, 0),
  }));
}

export function normalizeStockMoneyTrend(value) {
  return toArray(value).map((item) => ({
    opendate: item?.opendate ?? '',
    trade: toNumber(item?.trade, 0),
    netamount: toNumber(item?.netamount, 0),
    r0_net: toNumber(item?.r0_net, 0),
  }));
}

export async function analyzeMarketSentiment(keyword = '') {
  return AppBindings.AnalyzeSentimentWithFreqWeight(keyword);
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

export async function loadIndustryMoneyRanks(fenlei = '0', sort = 'netamount') {
  const result = await AppBindings.GetIndustryMoneyRankSina(fenlei, sort);
  return normalizeIndustryMoneyRanks(result);
}

export async function loadMoneyRanks(sort = 'netamount') {
  const result = await AppBindings.GetMoneyRankSina(sort);
  return normalizeMoneyRanks(result);
}

export async function loadStockMoneyTrend(stockCode, days = 20) {
  const result = await AppBindings.GetStockMoneyTrendByDay(stockCode, days);
  return normalizeStockMoneyTrend(result);
}

export async function refreshMarketFeed(source) {
  const result = await AppBindings.RefreshMarketFeed(source);
  return normalizeMarketFeed(result);
}
