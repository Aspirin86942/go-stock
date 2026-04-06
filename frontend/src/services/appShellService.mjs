import * as AppBindings from '../../wailsjs/go/main/App.js';
import { BrowserOpenURL, Environment } from '../../wailsjs/runtime/runtime.js';

export function buildCenteredWindowFeatures({
  width,
  height,
  screenWidth,
  screenHeight,
}) {
  const left = Math.round((screenWidth - width) / 2);
  const top = Math.round((screenHeight - height) / 2);
  return `width=${width},height=${height},left=${left},top=${top},location=no,menubar=no,toolbar=no,display=standalone`;
}

export async function getVersionInfo() {
  return AppBindings.GetVersionInfo();
}

export async function getGroupList() {
  const result = await AppBindings.GetGroupList();
  return Array.isArray(result) ? result : [];
}

export async function saveImageFile(name, base64) {
  return AppBindings.SaveImage(name, base64);
}

export async function saveWordFile(name, base64) {
  return AppBindings.SaveWordFile(name, base64);
}

export async function openBrowserUrl(url, { browserOpener = BrowserOpenURL } = {}) {
  await browserOpener(url);
  return 'browser';
}

export async function openExternalUrl(
  url,
  {
    width = 1000,
    height = 600,
    environmentResolver = Environment,
    wailsOpener = AppBindings.OpenURL,
    windowObject = globalThis.window,
  } = {},
) {
  const env = await environmentResolver();
  if (env?.platform === 'windows' && windowObject?.open && windowObject?.screen) {
    windowObject.open(
      url,
      'centeredWindow',
      buildCenteredWindowFeatures({
        width,
        height,
        screenWidth: windowObject.screen.width,
        screenHeight: windowObject.screen.height,
      }),
    );
    return 'window';
  }
  await wailsOpener(url);
  return 'wails';
}
