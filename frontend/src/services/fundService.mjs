import * as AppBindings from '../../wailsjs/go/main/App.js';

export function normalizeFundList(value) {
  return Array.isArray(value) ? value : [];
}

export function normalizeFollowedFunds(value) {
  return Array.isArray(value) ? value : [];
}

export async function loadFundList(key = '') {
  return normalizeFundList(await AppBindings.GetfundList(key));
}

export async function loadFollowedFunds() {
  return normalizeFollowedFunds(await AppBindings.GetFollowedFund());
}

export async function followFund(fundCode) {
  return AppBindings.FollowFund(fundCode);
}

export async function unfollowFund(fundCode) {
  return AppBindings.UnFollowFund(fundCode);
}
