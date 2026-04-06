import * as AppBindings from '../../wailsjs/go/main/App.js';

function toArray(value) {
  return Array.isArray(value) ? value : [];
}

function toNumber(value, fallback = 0) {
  const normalized = Number(value);
  return Number.isFinite(normalized) ? normalized : fallback;
}

function toObject(value) {
  return value && typeof value === 'object' ? value : {};
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

export function normalizeHotTopics(value) {
  return toArray(value);
}

export function normalizeMinutePriceLine(value) {
  const data = toObject(value);
  return {
    stockCode: data.stockCode ?? '',
    stockName: data.stockName ?? '',
    date: data.date ?? '',
    priceData: toArray(data.priceData).map((item) => ({
      time: item?.time ?? '',
      price: toNumber(item?.price, 0),
      volume: toNumber(item?.volume, 0),
      amount: toNumber(item?.amount, 0),
    })),
  };
}

export function normalizeRealtimePrice(value) {
  const data = toObject(value);
  return {
    stockCode: data.stockCode ?? '',
    stockName: data.stockName ?? '',
    price: data.price ?? '',
    bid: data.bid ?? '',
    ask: data.ask ?? '',
    open: data.open ?? '',
    high: data.high ?? '',
    low: data.low ?? '',
    preClose: data.preClose ?? '',
    date: data.date ?? '',
    time: data.time ?? '',
  };
}

export function normalizeHotStrategy(value) {
  const data = toObject(value);
  return {
    ...data,
    code: toNumber(data.code, 0),
    data: toArray(data.data),
    message: data.message ?? '',
  };
}

export function normalizeEastMoneyKLinePageResult(value) {
  const data = toObject(value);
  return {
    ok: !!data.ok,
    data: toArray(data.data),
    message: data.message ?? '',
    errorCode: data.errorCode ?? '',
    usedCookieRetry: !!data.usedCookieRetry,
  };
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

export async function loadLongTigerRanks(date) {
  return toArray(await AppBindings.LongTigerRank(date));
}

export async function loadStockResearchReports(stockCode) {
  return toArray(await AppBindings.StockResearchReport(stockCode));
}

export async function loadStockNotices(stockCode) {
  return toArray(await AppBindings.StockNotice(stockCode));
}

export async function loadIndustryResearchReports(industryCode) {
  return toArray(await AppBindings.IndustryResearchReport(industryCode));
}

export async function loadEMDictCodes(code) {
  return toArray(await AppBindings.EMDictCode(code));
}

export async function loadHotStocks(marketType = '10') {
  return toArray(await AppBindings.HotStock(marketType));
}

export async function loadHotEvents(size = 10) {
  return toArray(await AppBindings.HotEvent(size));
}

export async function loadHotTopics(size = 10) {
  return normalizeHotTopics(await AppBindings.HotTopic(size));
}

export async function loadInvestCalendar(yearMonth) {
  return toArray(await AppBindings.InvestCalendarTimeLine(yearMonth));
}

export async function loadClsCalendar() {
  return toArray(await AppBindings.ClsCalendar());
}

export async function searchStocks(words) {
  return toObject(await AppBindings.SearchStock(words));
}

export async function loadHotStrategy() {
  return normalizeHotStrategy(await AppBindings.GetHotStrategy());
}

export async function loadStockKLine(stockCode, stockName, days = 365) {
  return toArray(await AppBindings.GetStockKLine(stockCode, stockName, days));
}

export async function loadStockCommonKLine(stockCode, stockName, days = 365) {
  return toArray(await AppBindings.GetStockCommonKLine(stockCode, stockName, days));
}

export async function loadMinutePriceLine(stockCode, stockName) {
  return normalizeMinutePriceLine(await AppBindings.GetStockMinutePriceLineData(stockCode, stockName));
}

export async function loadEastMoneyKLineResult(stockCode, stockName, klt, limit) {
  return normalizeEastMoneyKLinePageResult(
    await AppBindings.GetStockEastMoneyKLineResult(stockCode, stockName, klt, limit),
  );
}

export async function loadEastMoneyKLinePageResult(stockCode, stockName, klt, limit, end = '') {
  return normalizeEastMoneyKLinePageResult(
    await AppBindings.GetStockEastMoneyKLinePageResult(stockCode, stockName, klt, limit, end),
  );
}

export async function loadRealtimePrice(stockCode) {
  return normalizeRealtimePrice(await AppBindings.GetStockRealTimePrice(stockCode));
}
