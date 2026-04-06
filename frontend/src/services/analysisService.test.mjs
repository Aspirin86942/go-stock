import test from 'node:test';
import assert from 'node:assert/strict';

import {
  normalizeAnalysisResult,
  normalizeAnalysisResultPage,
  normalizePromptTemplates,
  normalizePromptTemplatePage,
} from './analysisService.mjs';

test('normalizeAnalysisResult fills missing fields with stable defaults', () => {
  assert.deepEqual(
    normalizeAnalysisResult({ stockCode: '000001.SZ', content: '分析完成' }),
    {
      ID: 0,
      chatId: '',
      modelName: '',
      stockCode: '000001.SZ',
      stockName: '',
      question: '',
      content: '分析完成',
      CreatedAt: '',
      UpdatedAt: '',
    },
  );
});

test('normalizeAnalysisResultPage maps list and pagination defaults', () => {
  assert.deepEqual(
    normalizeAnalysisResultPage({
      list: [{ stockCode: '000001.SZ', content: '分析完成' }],
      total: 1,
      totalPages: 1,
    }),
    {
      list: [
        {
          ID: 0,
          chatId: '',
          modelName: '',
          stockCode: '000001.SZ',
          stockName: '',
          question: '',
          content: '分析完成',
          CreatedAt: '',
          UpdatedAt: '',
        },
      ],
      total: 1,
      page: 1,
      pageSize: 10,
      totalPages: 1,
    },
  );
});

test('normalizePromptTemplates keeps only normalized prompt rows', () => {
  assert.deepEqual(
    normalizePromptTemplates([{ ID: 3, name: '系统模板', type: '模型系统Prompt', content: '请先分析风险' }]),
    [{ ID: 3, name: '系统模板', type: '模型系统Prompt', content: '请先分析风险', CreatedAt: '', UpdatedAt: '' }],
  );
});

test('normalizePromptTemplatePage falls back to an empty page', () => {
  assert.deepEqual(normalizePromptTemplatePage(null), {
    list: [],
    total: 0,
    page: 1,
    pageSize: 10,
    totalPages: 0,
  });
});
