import test from 'node:test';
import assert from 'node:assert/strict';

import {
  normalizeMarketFeeds,
  normalizeMarketFeed,
  normalizeMarketIndexes,
  normalizeMarketIndustryRanks,
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
