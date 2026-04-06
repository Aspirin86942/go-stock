import * as AppBindings from '../../wailsjs/go/main/App.js';

function toArray(value) {
  return Array.isArray(value) ? value : [];
}

function toNumber(value, fallback = 0) {
  return Number.isFinite(Number(value)) ? Number(value) : fallback;
}

export function normalizeAnalysisResult(value) {
  const data = value ?? {};
  return {
    ID: toNumber(data.ID, 0),
    chatId: data.chatId ?? '',
    modelName: data.modelName ?? '',
    stockCode: data.stockCode ?? '',
    stockName: data.stockName ?? '',
    question: data.question ?? '',
    content: data.content ?? '',
    CreatedAt: data.CreatedAt ?? '',
    UpdatedAt: data.UpdatedAt ?? '',
  };
}

export function normalizeAnalysisResultPage(value) {
  const data = value ?? {};
  return {
    list: toArray(data.list).map(normalizeAnalysisResult),
    total: toNumber(data.total, 0),
    page: toNumber(data.page, 1),
    pageSize: toNumber(data.pageSize, 10),
    totalPages: toNumber(data.totalPages, 0),
  };
}

function normalizePromptTemplate(value) {
  const data = value ?? {};
  return {
    ID: toNumber(data.ID, 0),
    name: data.name ?? '',
    type: data.type ?? '',
    content: data.content ?? '',
    CreatedAt: data.CreatedAt ?? '',
    UpdatedAt: data.UpdatedAt ?? '',
  };
}

export function normalizePromptTemplates(value) {
  return toArray(value).map(normalizePromptTemplate);
}

export function normalizePromptTemplatePage(value) {
  const data = value ?? {};
  return {
    list: toArray(data.list).map(normalizePromptTemplate),
    total: toNumber(data.total, 0),
    page: toNumber(data.page, 1),
    pageSize: toNumber(data.pageSize, 10),
    totalPages: toNumber(data.totalPages, 0),
  };
}

export async function startStockAnalysis({
  stockName,
  stockCode,
  question,
  aiConfigId,
  sysPromptId,
  enableTools = true,
  think = false,
}) {
  return AppBindings.NewChatStream(stockName, stockCode, question, aiConfigId, sysPromptId, enableTools, think);
}

export async function startMarketSummary({
  question,
  aiConfigId,
  sysPromptId,
  enableTools = true,
  think = false,
  eventName = 'summaryStockNews',
  historyJSON = '',
}) {
  return AppBindings.SummaryStockNews(question, aiConfigId, sysPromptId, enableTools, think, eventName, historyJSON);
}

export async function saveAnalysisResult({
  stockCode,
  stockName,
  content,
  chatId,
  question,
  aiConfigId,
}) {
  return AppBindings.SaveAIResponseResult(stockCode, stockName, content, chatId, question, aiConfigId);
}

export async function loadLatestAnalysisResult(stockCode) {
  return normalizeAnalysisResult(await AppBindings.GetAIResponseResult(stockCode));
}

export async function loadAnalysisResultPage(query) {
  return normalizeAnalysisResultPage(await AppBindings.GetAIResponseResultList(query));
}

export async function deleteAnalysisResult(id) {
  return AppBindings.DeleteAIResponseResult(id);
}

export async function loadPromptTemplates(name = '', type = '') {
  return normalizePromptTemplates(await AppBindings.GetPromptTemplates(name, type));
}

export async function loadPromptTemplatePage(query) {
  return normalizePromptTemplatePage(await AppBindings.GetPromptTemplateList(query));
}

export async function savePromptTemplate(template, { edit = false } = {}) {
  if (edit) {
    return AppBindings.UpdatePromptTemplate(template);
  }
  return AppBindings.AddPromptTemplate(template);
}

export async function deletePromptTemplate(id) {
  return AppBindings.DeletePromptTemplate(id);
}

export async function shareAnalysis(stockCode, stockName) {
  return AppBindings.ShareAnalysis(stockCode, stockName);
}

export async function saveAnalysisMarkdown(stockCode, stockName) {
  return AppBindings.SaveAsMarkdown(stockCode, stockName);
}
