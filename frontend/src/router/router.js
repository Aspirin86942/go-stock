import { createRouter, createWebHashHistory } from 'vue-router';

import stockPageView from '../pages/stock-page.vue';
import settingsPageView from '../pages/settings-page.vue';
import aboutPageView from '../pages/about-page.vue';
import fundPageView from '../pages/fund-page.vue';
import marketPageView from "../pages/market-page.vue";
import agentPageView from '../pages/agent-page.vue';
import researchPageView from "../pages/research-page.vue";
import cronTaskPageView from "../pages/cron-task-page.vue";

const routes = [
  { path: '/', component: stockPageView, name: 'stock' },
  { path: '/fund', component: fundPageView, name: 'fund' },
  { path: '/settings', component: settingsPageView, name: 'settings' },
  { path: '/about', component: aboutPageView, name: 'about' },
  { path: '/market', component: marketPageView, name: 'market' },
  { path: '/agent', component: agentPageView, name: 'agent' },
  { path: '/research', component: researchPageView, name: 'research' },
  { path: '/cron-tasks', component: cronTaskPageView, name: 'cronTasks' },
];

const router = createRouter({
  history: createWebHashHistory(),
  routes,
});

export default router;
