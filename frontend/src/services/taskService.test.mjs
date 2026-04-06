import test from 'node:test';
import assert from 'node:assert/strict';

import {
  normalizeCronTask,
  normalizeCronTaskPage,
} from './taskService.mjs';

test('normalizeCronTask keeps stable camel-case fields', () => {
  assert.deepEqual(
    normalizeCronTask({ id: 7, name: '市场分析', cronExpr: '0 */5 * * * *', enable: true }),
    {
      id: 7,
      name: '市场分析',
      cronExpr: '0 */5 * * * *',
      enable: true,
      taskType: '',
      params: '',
      status: '',
      description: '',
      target: '',
      lastRunAt: '',
      nextRunAt: '',
      runCount: 0,
      lastRunResult: '',
    },
  );
});

test('normalizeCronTaskPage falls back to empty page', () => {
  assert.deepEqual(normalizeCronTaskPage(null), { total: 0, data: [] });
});
