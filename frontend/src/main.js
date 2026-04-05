import {createApp} from 'vue'
import naive from 'naive-ui'
import App from './App.vue'
import router from './router/router'
import { emitFrontendError, installFrontendErrorHandlers, isResizeObserverNoise } from './utils/frontendLogger.mjs'
// 引入组件库的少量全局样式变量
import 'tdesign-vue-next/es/style/index.css';

const app = createApp(App)

app.config.errorHandler = (err, _instance, info) => {
  if (isResizeObserverNoise(err?.message) || isResizeObserverNoise(err?.stack)) {
    return
  }
  emitFrontendError({
    page: 'main.js',
    message: err?.message || info || 'vue error',
    error: err,
    extra: { info: info || '' },
  })
  console.error(err)
}

installFrontendErrorHandlers('main.js')

app.use(router)
app.use(naive)
app.mount('#app')
