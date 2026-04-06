import * as AppBindings from '../../wailsjs/go/main/App.js';

function toArray(value) {
  return Array.isArray(value) ? value : [];
}

export function normalizeFollowList(value) {
  return toArray(value);
}

export function normalizeGroupList(value) {
  return toArray(value);
}

export async function loadGroupList() {
  return normalizeGroupList(await AppBindings.GetGroupList());
}

export async function loadFollowList(groupId = 0) {
  return normalizeFollowList(await AppBindings.GetFollowList(groupId));
}

export async function followStock(stockCode) {
  return AppBindings.Follow(stockCode);
}

export async function unfollowStock(stockCode) {
  return AppBindings.UnFollow(stockCode);
}

export async function addGroup(group) {
  return AppBindings.AddGroup(group);
}

export async function updateGroupSort(id, newSort) {
  return AppBindings.UpdateGroupSort(id, newSort);
}

export async function initializeGroupSort() {
  return AppBindings.InitializeGroupSort();
}

export async function addGroupStock(groupId, stockCode) {
  return AppBindings.AddStockGroup(groupId, stockCode);
}

export async function removeGroupStock(stockCode, name, groupId) {
  return AppBindings.RemoveStockGroup(stockCode, name, groupId);
}

export async function removeGroup(groupId) {
  return AppBindings.RemoveGroup(groupId);
}

export async function setCostPriceAndVolume(stockCode, price, volume) {
  return AppBindings.SetCostPriceAndVolume(stockCode, price, volume);
}

export async function setTradingPrice(stockCode, entryPrice, takeProfitPrice, stopLossPrice, costPrice) {
  return AppBindings.SetTradingPrice(stockCode, entryPrice, takeProfitPrice, stopLossPrice, costPrice);
}

export async function setAlarmChangePercent(val, alarmPrice, stockCode) {
  return AppBindings.SetAlarmChangePercent(val, alarmPrice, stockCode);
}

export async function setStockSort(sort, stockCode) {
  return AppBindings.SetStockSort(sort, stockCode);
}

export async function saveStockAICron(cronText, stockCode) {
  return AppBindings.SetStockAICron(cronText, stockCode);
}
