import test from 'node:test';
import assert from 'node:assert/strict';

import { loadAssistantSession, normalizeAssistantSession, saveAssistantSession } from './assistantService.mjs';

test('normalizeAssistantSession 在缺失消息时回退为空消息数组', () => {
  assert.deepEqual(normalizeAssistantSession(null), { sessionId: '', messages: [] });
  assert.deepEqual(normalizeAssistantSession({ sessionId: 's1' }), { sessionId: 's1', messages: [] });
});

test('loadAssistantSession 返回空消息数组的回退', async () => {
  const originalWindow = globalThis.window;
  globalThis.window = {
    go: {
      main: {
        App: {
          GetAiAssistantSession: async () => ({ sessionId: 'latest', messages: null }),
        },
      },
    },
  };

  try {
    assert.deepEqual(await loadAssistantSession(''), {
      sessionId: 'latest',
      messages: [],
    });
  } finally {
    globalThis.window = originalWindow;
  }
});

test('saveAssistantSession 支持旧的仅消息数组调用', async () => {
  const originalWindow = globalThis.window;
  const calls = [];
  globalThis.window = {
    go: {
      main: {
        App: {
          SaveAiAssistantSession: async (sessionId, messages) => {
            calls.push({ sessionId, messages });
          },
        },
      },
    },
  };

  try {
    await saveAssistantSession([{ role: 'assistant', content: 'hello' }]);
    assert.deepEqual(calls, [
      {
        sessionId: '',
        messages: [{ role: 'assistant', content: 'hello' }],
      },
    ]);
  } finally {
    globalThis.window = originalWindow;
  }
});
