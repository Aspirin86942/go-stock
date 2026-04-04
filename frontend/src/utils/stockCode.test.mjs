import test from 'node:test';
import assert from 'node:assert/strict';

import { normalizeFollowStockCode, resolveFollowStockCode } from './stockCode.mjs';

const sampleStocks = [
  { name: '贵州茅台', ts_code: '600519.SH' },
  { name: '平安银行', ts_code: '000001.SZ' },
  { name: '柏楚电子', ts_code: '688188.SH' },
  { name: '腾讯控股', ts_code: '00700.HK' },
  { name: '苹果', ts_code: 'aapl.us' },
];

test('normalizeFollowStockCode converts tushare SH/SZ code to market code', () => {
  assert.equal(normalizeFollowStockCode('600519.SH', sampleStocks), 'sh600519');
  assert.equal(normalizeFollowStockCode('000001.SZ', sampleStocks), 'sz000001');
});

test('normalizeFollowStockCode resolves exact stock name from stock list', () => {
  assert.equal(normalizeFollowStockCode('贵州茅台', sampleStocks), 'sh600519');
});

test('normalizeFollowStockCode keeps already normalized market code', () => {
  assert.equal(normalizeFollowStockCode('sh600519', sampleStocks), 'sh600519');
  assert.equal(normalizeFollowStockCode('hk00700', sampleStocks), 'hk00700');
  assert.equal(normalizeFollowStockCode('usaapl', sampleStocks), 'usaapl');
});

test('resolveFollowStockCode falls back to selected input value when code state is stale', () => {
  assert.equal(resolveFollowStockCode([null, '688188.SH'], sampleStocks), 'sh688188');
  assert.equal(resolveFollowStockCode(['', '柏楚电子'], sampleStocks), 'sh688188');
  assert.equal(resolveFollowStockCode([null, '柏楚电子 - 688188.SH'], sampleStocks), 'sh688188');
});

test('normalizeFollowStockCode returns null for partial or unknown input', () => {
  assert.equal(normalizeFollowStockCode('600519', sampleStocks), null);
  assert.equal(normalizeFollowStockCode('贵州', sampleStocks), null);
  assert.equal(normalizeFollowStockCode('不存在的股票', sampleStocks), null);
  assert.equal(normalizeFollowStockCode('', sampleStocks), null);
});
