import * as AppBindings from '../../wailsjs/go/main/App.js';

function toArray(value) {
  return Array.isArray(value) ? value : [];
}

function toNumber(value, fallback = 0) {
  return Number.isFinite(Number(value)) ? Number(value) : fallback;
}

export function normalizeAiConfigs(value) {
  return toArray(value).map((item) => ({
    ...(item ?? {}),
    ID: toNumber(item?.ID, 0),
    name: item?.name ?? '',
    modelName: item?.modelName ?? '',
  }));
}

export function normalizePromptTemplates(value) {
  return toArray(value).map((item) => ({
    ID: toNumber(item?.ID, 0),
    name: item?.name ?? '',
    type: item?.type ?? '',
    content: item?.content ?? '',
  }));
}

export function normalizeSettingConfig(value) {
  const data = value ?? {};
  return {
    ...data,
    aiConfigs: normalizeAiConfigs(data.aiConfigs),
  };
}

export async function loadAppConfig() {
  return normalizeSettingConfig(await AppBindings.GetConfig());
}

export async function saveAppConfig(config) {
  return AppBindings.UpdateConfig(config);
}

export async function exportAppConfig() {
  return AppBindings.ExportConfig();
}

export async function fetchAiModels(baseUrl, apiKey) {
  return toArray(await AppBindings.FetchAiModels(baseUrl, apiKey)).map(String);
}

export async function loadAiConfigs() {
  return normalizeAiConfigs(await AppBindings.GetAiConfigs());
}

export async function loadPromptTemplates(name = '', type = '') {
  return normalizePromptTemplates(await AppBindings.GetPromptTemplates(name, type));
}

export async function saveLegacyPrompt(prompt) {
  return AppBindings.AddPrompt(prompt);
}

export async function deleteLegacyPrompt(id) {
  return AppBindings.DelPrompt(id);
}

export async function sendTypedNotification(message, stockCode, msgType) {
  return AppBindings.SendDingDingMessageByType(message, stockCode, msgType);
}
