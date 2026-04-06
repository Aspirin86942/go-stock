import test from 'node:test';
import assert from 'node:assert/strict';

import {
  normalizeAiRecommendPage,
  normalizeStockChangePage,
  normalizeTradingRecordPage,
  loadAiRecommendPage,
} from './researchService.mjs';

test('normalizeStockChangePage falls back to a stable empty page', () => {
  assert.deepEqual(normalizeStockChangePage(null), {
    totalCount: 0,
    data: [],
  });
});

test('normalizeAiRecommendPage keeps page metadata and list defaults', () => {
  assert.deepEqual(normalizeAiRecommendPage({ total: 2, list: null, page: 3 }), {
    list: [],
    total: 2,
    page: 3,
    pageSize: 10,
    totalPages: 0,
  });
});

test('normalizeTradingRecordPage falls back to a stable empty page', () => {
  assert.deepEqual(normalizeTradingRecordPage(undefined), {
    list: [],
    total: 0,
    page: 1,
    pageSize: 20,
    totalPages: 0,
  });
});

test('loadAiRecommendPage returns empty page when binding resolves undefined', async () => {
  const originalWindow = globalThis.window;
  globalThis.window = {
    go: {
      main: {
        App: {
          GetAiRecommendStocksList: async () => undefined,
        },
      },
    },
  };

  try {
    assert.deepEqual(await loadAiRecommendPage({ page: 2, pageSize: 50 }), {
      list: [],
      total: 0,
      page: 2,
      pageSize: 50,
      totalPages: 0,
    });
  } finally {
    globalThis.window = originalWindow;
  }
});
