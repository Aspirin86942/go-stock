import test from 'node:test';
import assert from 'node:assert/strict';

import { resolveFirstAiConfigId } from './aiConfig.mjs';

test('resolveFirstAiConfigId returns null when config list is empty', () => {
  assert.equal(resolveFirstAiConfigId([]), null);
});

test('resolveFirstAiConfigId returns the first config ID', () => {
  assert.equal(resolveFirstAiConfigId([{ ID: 7 }, { ID: 9 }]), 7);
});

test('resolveFirstAiConfigId falls back to lowercase id', () => {
  assert.equal(resolveFirstAiConfigId([{ id: 11 }]), 11);
});
