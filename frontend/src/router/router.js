import { createRouter, createWebHashHistory } from 'vue-router';

import stockPageView from '../pages/stock-page.vue';
import settingsPageView from '../pages/settings-page.vue';
import aboutView from "../components/about.vue";
import fundView from "../components/fund.vue";
import marketPageView from "../pages/market-page.vue";
import agentChat from "../components/agent-chat.vue";
import researchPageView from "../pages/research-page.vue";
import cronTaskPageView from "../pages/cron-task-page.vue";

const routes = [
  { path: '/', component: stockPageView, name: 'stock' },
  { path: '/fund', component: fundView, name: 'fund' },
  { path: '/settings', component: settingsPageView, name: 'settings' },
  { path: '/about', component: aboutView, name: 'about' },
  { path: '/market', component: marketPageView, name: 'market' },
  { path: '/agent', component: agentChat, name: 'agent' },
  { path: '/research', component: researchPageView, name: 'research' },
  { path: '/cron-tasks', component: cronTaskPageView, name: 'cronTasks' },
];

const router = createRouter({
  history: createWebHashHistory(),
  routes,
});

export default router;
