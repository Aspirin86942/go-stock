import test from 'node:test';
import assert from 'node:assert/strict';

import {
  buildCenteredWindowFeatures,
  openExternalUrl,
} from './appShellService.mjs';

test('buildCenteredWindowFeatures 生成稳定的窗口参数字符串', () => {
  assert.equal(
    buildCenteredWindowFeatures({ width: 1000, height: 600, screenWidth: 1920, screenHeight: 1080 }),
    'width=1000,height=600,left=460,top=240,location=no,menubar=no,toolbar=no,display=standalone',
  );
});

test('openExternalUrl 在 windows 优先打开居中弹窗', async () => {
  const calls = [];
  const result = await openExternalUrl('https://example.com', {
    width: 900,
    height: 500,
    environmentResolver: async () => ({ platform: 'windows' }),
    wailsOpener: async (url) => {
      calls.push({ type: 'wails', url });
    },
    windowObject: {
      screen: { width: 1600, height: 900 },
      open(url, target, features) {
        calls.push({ type: 'window', url, target, features });
      },
    },
  });

  assert.equal(result, 'window');
  assert.deepEqual(calls, [
    {
      type: 'window',
      url: 'https://example.com',
      target: 'centeredWindow',
      features: 'width=900,height=500,left=350,top=200,location=no,menubar=no,toolbar=no,display=standalone',
    },
  ]);
});

test('openExternalUrl 在非 windows 回退到 Wails OpenURL', async () => {
  const calls = [];
  const result = await openExternalUrl('https://example.com', {
    environmentResolver: async () => ({ platform: 'linux' }),
    wailsOpener: async (url) => {
      calls.push(url);
    },
    windowObject: {
      screen: { width: 1600, height: 900 },
      open() {
        throw new Error('should not open popup');
      },
    },
  });

  assert.equal(result, 'wails');
  assert.deepEqual(calls, ['https://example.com']);
});
