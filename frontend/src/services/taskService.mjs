import * as AppBindings from '../../wailsjs/go/main/App.js';

function toArray(value) {
  return Array.isArray(value) ? value : [];
}

function toNumber(value, fallback = 0) {
  return Number.isFinite(Number(value)) ? Number(value) : fallback;
}

export function normalizeCronTask(value) {
  const data = value ?? {};
  return {
    id: toNumber(data.id ?? data.ID, 0),
    name: data.name ?? data.Name ?? '',
    cronExpr: data.cronExpr ?? data.CronExpr ?? '',
    taskType: data.taskType ?? data.TaskType ?? '',
    target: data.target ?? data.Target ?? '',
    params: data.params ?? data.Params ?? '',
    enable: Boolean(data.enable ?? data.Enable),
    status: data.status ?? data.Status ?? '',
    description: data.description ?? data.Description ?? '',
    lastRunAt: data.lastRunAt ?? data.LastRunAt ?? '',
    nextRunAt: data.nextRunAt ?? data.NextRunAt ?? '',
    runCount: toNumber(data.runCount ?? data.RunCount, 0),
    lastRunResult: data.lastRunResult ?? data.LastRunResult ?? '',
  };
}

export function normalizeCronTaskPage(value) {
  const data = value ?? {};
  return {
    total: toNumber(data.total ?? data.Total, 0),
    data: toArray(data.data ?? data.Data).map(normalizeCronTask),
  };
}

export async function loadCronTask(id) {
  const result = await AppBindings.GetCronTaskByID(id);
  return result ? normalizeCronTask(result) : null;
}

export async function loadCronTaskList(query) {
  return normalizeCronTaskPage(await AppBindings.GetCronTaskList(query));
}

export async function createCronTask(task) {
  return AppBindings.CreateCronTask(task);
}

export async function updateCronTask(task) {
  return AppBindings.UpdateCronTask(task);
}

export async function deleteCronTask(id) {
  return AppBindings.DeleteCronTask(id);
}

export async function toggleCronTask(id, enable) {
  return AppBindings.EnableCronTask(id, enable);
}

export async function executeCronTaskNow(id) {
  return AppBindings.ExecuteCronTaskNow(id);
}

export async function loadCronTaskTypes() {
  return toArray(await AppBindings.GetCronTaskTypes());
}

export async function validateCronExpression(expr) {
  return AppBindings.ValidateCronExpr(expr);
}

export async function calculateNextRunTime(expr) {
  return AppBindings.CalculateNextRunTime(expr);
}

export async function calculateNextRunTimes(expr, count) {
  return toArray(await AppBindings.CalculateNextRunTimes(expr, count));
}

export async function searchCronTasks(keyword) {
  return toArray(await AppBindings.SearchCronTasks(keyword)).map(normalizeCronTask);
}
