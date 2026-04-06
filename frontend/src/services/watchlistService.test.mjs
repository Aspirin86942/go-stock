import test from 'node:test';
import assert from 'node:assert/strict';

import {
  normalizeFollowList,
  normalizeGroupList,
  loadGroupList,
  loadFollowList,
} from './watchlistService.mjs';

test('normalizeFollowList falls back to an empty array', () => {
  assert.deepEqual(normalizeFollowList(undefined), []);
  assert.deepEqual(normalizeFollowList(null), []);
});

test('normalizeGroupList keeps arrays untouched and drops non-arrays', () => {
  assert.deepEqual(normalizeGroupList([{ ID: 1, name: '全部' }]), [{ ID: 1, name: '全部' }]);
  assert.deepEqual(normalizeGroupList('invalid'), []);
});

test('loadGroupList returns stable empty array when binding resolves null', async () => {
  const originalWindow = globalThis.window;
  globalThis.window = {
    go: {
      main: {
        App: {
          GetGroupList: async () => null,
        },
      },
    },
  };

  try {
    assert.deepEqual(await loadGroupList(), []);
  } finally {
    globalThis.window = originalWindow;
  }
});

test('loadFollowList returns stable array from binding', async () => {
  const originalWindow = globalThis.window;
  globalThis.window = {
    go: {
      main: {
        App: {
          GetFollowList: async () => [{ StockCode: 'sz000001' }],
        },
      },
    },
  };

  try {
    assert.deepEqual(await loadFollowList(0), [{ StockCode: 'sz000001' }]);
  } finally {
    globalThis.window = originalWindow;
  }
});
