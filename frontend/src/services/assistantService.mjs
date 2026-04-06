import * as AppBindings from '../../wailsjs/go/main/App.js';

function toArray(value) {
  return Array.isArray(value) ? value : [];
}

function toObject(value) {
  return value && typeof value === 'object' ? value : {};
}

export function normalizeAssistantSession(value, fallbackSessionId = '') {
  const data = toObject(value);
  return {
    sessionId: data.sessionId ?? data.SessionId ?? fallbackSessionId,
    messages: toArray(data.messages ?? data.Messages),
  };
}

export async function loadAssistantSession(sessionId = '') {
  return normalizeAssistantSession(await AppBindings.GetAiAssistantSession(sessionId), sessionId);
}

export async function saveAssistantSession(sessionIdOrMessages, messagesMaybe) {
  if (Array.isArray(sessionIdOrMessages) && messagesMaybe === undefined) {
    return AppBindings.SaveAiAssistantSession('', sessionIdOrMessages);
  }
  return AppBindings.SaveAiAssistantSession(sessionIdOrMessages ?? '', toArray(messagesMaybe));
}

export async function startAgentChat(question, aiConfigId, sysPromptId, memoryMode, memoryCount, thinkingMode) {
  return AppBindings.ChatWithAgent(question, aiConfigId, sysPromptId, memoryMode, memoryCount, thinkingMode);
}

export async function abortAgentChat() {
  return AppBindings.AbortChatWithAgent();
}

export async function startAssistantSummary(question, aiConfigId, sysPromptId, enableTools, thinkingMode, eventName, historyJSON) {
  return AppBindings.SummaryStockNews(question, aiConfigId, sysPromptId, enableTools, thinkingMode, eventName, historyJSON);
}

export async function abortAssistantSummary() {
  return AppBindings.AbortSummaryStockNews();
}

export async function shareAssistantText(text, title) {
  return AppBindings.ShareText(text, title);
}
