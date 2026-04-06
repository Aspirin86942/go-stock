import test from 'node:test';
import assert from 'node:assert/strict';

import { normalizeFollowedFunds, loadFollowedFunds } from './fundService.mjs';

test('normalizeFollowedFunds 在空输入时回退为空数组', () => {
  assert.deepEqual(normalizeFollowedFunds(undefined), []);
  assert.deepEqual(normalizeFollowedFunds(null), []);
});

test('loadFollowedFunds 在绑定返回 null 时回退为空数组', async () => {
  const originalWindow = globalThis.window;
  globalThis.window = {
    go: {
      main: {
        App: {
          GetFollowedFund: async () => null,
        },
      },
    },
  };

  try {
    assert.deepEqual(await loadFollowedFunds(), []);
  } finally {
    globalThis.window = originalWindow;
  }
});
