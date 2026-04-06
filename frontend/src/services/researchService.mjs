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

export function normalizeStockChangePage(value) {
  const data = toObject(value);
  return {
    totalCount: toNumber(data.totalCount, 0),
    data: toArray(data.data),
  };
}

export function normalizeAiRecommendPage(value, fallback = {}) {
  const data = toObject(value);
  const query = toObject(fallback);
  return {
    list: toArray(data.list),
    total: toNumber(data.total, 0),
    page: toNumber(data.page, query.page ?? 1),
    pageSize: toNumber(data.pageSize, query.pageSize ?? 10),
    totalPages: toNumber(data.totalPages, 0),
  };
}

export function normalizeTradingRecordPage(value, fallback = {}) {
  const data = toObject(value);
  const query = toObject(fallback);
  return {
    list: toArray(data.list),
    total: toNumber(data.total, 0),
    page: toNumber(data.page, query.page ?? 1),
    pageSize: toNumber(data.pageSize, query.pageSize ?? 20),
    totalPages: toNumber(data.totalPages, 0),
  };
}

export function normalizeStockChangeHistoryPage(value, fallback = {}) {
  const data = toObject(value);
  const query = toObject(fallback);
  return {
    list: toArray(data.list),
    total: toNumber(data.total, 0),
    page: toNumber(data.page, query.page ?? 1),
    pageSize: toNumber(data.pageSize, query.pageSize ?? 20),
    totalPages: toNumber(data.totalPages, 0),
  };
}

export function normalizeAllStockInfoPage(value, fallback = {}) {
  const data = toObject(value);
  const query = toObject(fallback);
  return {
    list: toArray(data.list),
    total: toNumber(data.total, 0),
    page: toNumber(data.page, query.page ?? 1),
    pageSize: toNumber(data.pageSize, query.pageSize ?? 20),
    totalPages: toNumber(data.totalPages, 0),
  };
}

export function normalizeTradingRecordStatistics(value) {
  const data = toObject(value);
  return {
    totalBuyAmount: toNumber(data.totalBuyAmount, 0),
    totalSellAmount: toNumber(data.totalSellAmount, 0),
    totalProfit: toNumber(data.totalProfit, 0),
    profitRate: toNumber(data.profitRate, 0),
    holdingsAmount: toNumber(data.holdingsAmount, 0),
    currentValue: toNumber(data.currentValue, 0),
    stockCount: toNumber(data.stockCount, 0),
  };
}

export function normalizeFrequentTradingCheck(value) {
  const data = toObject(value);
  return {
    canTrade: data.canTrade ?? true,
    msg: data.msg ?? '',
  };
}

export async function loadStockChanges(changeTypes, pageIndex, pageSize) {
  return normalizeStockChangePage(await AppBindings.GetStockChanges(changeTypes, pageIndex, pageSize));
}

export async function loadAllStockChangesWithPaging(pageSize) {
  return normalizeStockChangePage(await AppBindings.GetAllStockChangesWithPaging(pageSize));
}

export async function loadStockChangeHistory(query) {
  return normalizeStockChangeHistoryPage(await AppBindings.GetStockChangeHistory(query), query);
}

export async function saveStockChangesToHistory(changeTypes) {
  return AppBindings.SaveStockChangesToHistory(changeTypes);
}

export async function deleteStockChangeHistory(days) {
  return AppBindings.DeleteStockChangeHistory(days);
}

export async function loadAiRecommendPage(query) {
  return normalizeAiRecommendPage(await AppBindings.GetAiRecommendStocksList(query), query);
}

export async function deleteAiRecommendStock(id) {
  return AppBindings.DeleteAiRecommendStocks(id);
}

export async function updateAiRecommendStockAlert(id, enableAlert) {
  return AppBindings.UpdateAiRecommendStocksAlert(id, enableAlert);
}

export async function loadAllStockInfoPage(query) {
  return normalizeAllStockInfoPage(await AppBindings.GetAllStockInfoList(query), query);
}

export async function loadAllStockInfoById(id) {
  return toObject(await AppBindings.GetAllStockInfoById(id));
}

export async function loadAllMarkets() {
  return toArray(await AppBindings.GetAllMarkets());
}

export async function loadAllIndustries() {
  return toArray(await AppBindings.GetAllIndustries());
}

export async function loadAllConcepts() {
  return toArray(await AppBindings.GetAllConcepts());
}

export async function loadTradingRecordPage(query) {
  return normalizeTradingRecordPage(await AppBindings.GetTradingRecordList(query), query);
}

export async function addTradingRecord(record) {
  return AppBindings.AddTradingRecord(record);
}

export async function loadTradingRecordById(id) {
  return toObject(await AppBindings.GetTradingRecordById(id));
}

export async function loadTradingRecordStatistics() {
  return normalizeTradingRecordStatistics(await AppBindings.GetTradingRecordStatistics());
}

export async function updateTradingRecord(record) {
  return AppBindings.UpdateTradingRecord(record);
}

export async function deleteTradingRecord(id) {
  return AppBindings.DeleteTradingRecord(id);
}

export async function checkFrequentTrading(stockCode) {
  return normalizeFrequentTradingCheck(await AppBindings.CheckFrequentTrading(stockCode));
}
