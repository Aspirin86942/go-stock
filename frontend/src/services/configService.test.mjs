import test from 'node:test';
import assert from 'node:assert/strict';

import {
  normalizeAiConfigs,
  normalizePromptTemplates,
  normalizeSettingConfig,
} from './configService.mjs';

test('normalizeAiConfigs falls back to an empty array', () => {
  assert.deepEqual(normalizeAiConfigs(null), []);
});

test('normalizePromptTemplates keeps prompt rows stable', () => {
  assert.deepEqual(
    normalizePromptTemplates([{ ID: 7, name: '系统模板', type: '模型系统Prompt', content: '先看风险' }]),
    [{ ID: 7, name: '系统模板', type: '模型系统Prompt', content: '先看风险' }],
  );
});

test('normalizeSettingConfig always returns aiConfigs array', () => {
  assert.deepEqual(normalizeSettingConfig({ darkTheme: true }), {
    darkTheme: true,
    aiConfigs: [],
  });
});
