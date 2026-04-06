import test from 'node:test';
import assert from 'node:assert/strict';

import {
  normalizeMarketFeeds,
  normalizeMarketFeed,
  normalizeMarketIndexes,
  normalizeMarketIndustryRanks,
  normalizeIndustryMoneyRanks,
  normalizeMoneyRanks,
  normalizeStockMoneyTrend,
} from './marketService.mjs';

test('normalizeMarketFeeds 缺失字段时补空数组', () => {
  assert.deepEqual(normalizeMarketFeeds({ telegraph: ['t1'] }), {
    telegraph: ['t1'],
    sina: [],
    foreign: [],
  });
});

test('normalizeMarketFeed 只保留 source 和 items', () => {
  assert.deepEqual(
    normalizeMarketFeed({
      source: 'telegraph',
      items: [{ id: 1 }],
      ignored: 'x',
    }),
    {
      source: 'telegraph',
      items: [{ id: 1 }],
    },
  );
});

test('normalizeMarketIndexes 只保留已知区域并补空数组', () => {
  assert.deepEqual(
    normalizeMarketIndexes({
      common: [{ code: '000001' }],
      europe: [{ code: 'DAX' }],
      unknown: [{ code: 'XXX' }],
    }),
    {
      common: [{ code: '000001' }],
      america: [],
      europe: [{ code: 'DAX' }],
      asia: [],
      other: [],
    },
  );
});

test('normalizeMarketIndustryRanks 对空值回退为数组', () => {
  assert.deepEqual(normalizeMarketIndustryRanks(undefined), []);
  assert.deepEqual(normalizeMarketIndustryRanks(null), []);
});

test('normalizeIndustryMoneyRanks 数值字段转 number 且缺失字段补默认值', () => {
  assert.deepEqual(
    normalizeIndustryMoneyRanks([
      {
        name: '半导体',
        avg_changeratio: '0.125',
        inamount: '1000.2',
        outamount: '800',
        netamount: '200.2',
        ratioamount: '0.3',
        ts_name: 'xx',
        ts_symbol: '000001',
        ts_changeratio: '0.06',
        ts_trade: '13.2',
        ts_ratioamount: '0.09',
      },
      {
        name: '缺省',
      },
    ]),
    [
      {
        category: '',
        name: '半导体',
        avg_changeratio: 0.125,
        inamount: 1000.2,
        outamount: 800,
        netamount: 200.2,
        ratioamount: 0.3,
        ts_name: 'xx',
        ts_symbol: '000001',
        ts_changeratio: 0.06,
        ts_trade: 13.2,
        ts_ratioamount: 0.09,
      },
      {
        category: '',
        name: '缺省',
        avg_changeratio: 0,
        inamount: 0,
        outamount: 0,
        netamount: 0,
        ratioamount: 0,
        ts_name: '',
        ts_symbol: '',
        ts_changeratio: 0,
        ts_trade: 0,
        ts_ratioamount: 0,
      },
    ],
  );
});

test('normalizeMoneyRanks 数值字段转 number 且缺失字段补默认值', () => {
  assert.deepEqual(
    normalizeMoneyRanks([
      {
        symbol: 'sz000001',
        name: '平安银行',
        trade: '11.22',
        changeratio: '0.012',
        turnover: '321',
        amount: '12345',
        outamount: '1000',
        inamount: '1200',
        netamount: '200',
        ratioamount: '0.1',
        r0_out: '11',
        r0_in: '22',
        r0_net: '11',
        r0_ratio: '0.2',
        r3_out: '33',
        r3_in: '44',
        r3_net: '11',
        r3_ratio: '0.3',
      },
      {
        symbol: 'sz000002',
      },
    ]),
    [
      {
        symbol: 'sz000001',
        name: '平安银行',
        trade: 11.22,
        changeratio: 0.012,
        turnover: 321,
        amount: 12345,
        outamount: 1000,
        inamount: 1200,
        netamount: 200,
        ratioamount: 0.1,
        r0_out: 11,
        r0_in: 22,
        r0_net: 11,
        r0_ratio: 0.2,
        r3_out: 33,
        r3_in: 44,
        r3_net: 11,
        r3_ratio: 0.3,
      },
      {
        symbol: 'sz000002',
        name: '',
        trade: 0,
        changeratio: 0,
        turnover: 0,
        amount: 0,
        outamount: 0,
        inamount: 0,
        netamount: 0,
        ratioamount: 0,
        r0_out: 0,
        r0_in: 0,
        r0_net: 0,
        r0_ratio: 0,
        r3_out: 0,
        r3_in: 0,
        r3_net: 0,
        r3_ratio: 0,
      },
    ],
  );
});

test('normalizeStockMoneyTrend 仅保留指定字段并做 number normalize', () => {
  assert.deepEqual(
    normalizeStockMoneyTrend([
      {
        opendate: '2026-04-01',
        trade: '12.33',
        netamount: '10000.8',
        r0_net: '666.6',
        ignored: 'x',
      },
      {
        opendate: '2026-04-02',
      },
    ]),
    [
      {
        opendate: '2026-04-01',
        trade: 12.33,
        netamount: 10000.8,
        r0_net: 666.6,
      },
      {
        opendate: '2026-04-02',
        trade: 0,
        netamount: 0,
        r0_net: 0,
      },
    ],
  );
});
