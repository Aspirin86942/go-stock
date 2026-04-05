# 全量日志闭环手工验收

## 1. 日志目录

- Windows 桌面应用默认日志目录：`%LOCALAPPDATA%\go-stock\logs`
- 先确认存在：`app.log`、`error.log`、`http.log`、`ai.log`、`task.log`、`db.log`、`frontend.log`、`panic.log`
- 再确认存在：`payloads\YYYY-MM-DD\`

## 2. 前端错误

1. 打开任一页面，在浏览器控制台执行 `console.error(new Error('manual-frontend-probe'))`
2. 打开 `frontend.log`
3. 确认最新记录包含 `event=frontend.error`、`error_class=frontend_error`、`trace_id`

## 3. HTTP 请求

1. 发送一次本地请求：`Invoke-WebRequest http://127.0.0.1:18888/api/health`
2. 打开 `http.log`
3. 如请求或响应体超过阈值，继续检查 `payloads` 目录下是否出现新文件
4. 确认 `http.log` 中的 `trace_id` 可与同次链路的下游日志对应

## 4. AI 调用

1. 在应用里发起一次真实 AI 分析请求，例如输入 `测试日志链路`
2. 打开 `ai.log`
3. 确认存在本次请求的 `trace_id`、模型信息和关键事件

## 5. 任务失败

1. 在任务配置中执行一个 `stock_analysis` 任务，并故意将 `params` 写成非法 JSON
2. 打开 `task.log`
3. 确认最新失败记录包含 `error_class=task_error`、`error_message`、`trace_id`

## 6. 最终核对

- 同一条关键链路至少能在两个不同 sink 里看到同一个 `trace_id`
- 有大体积原文时，主日志只保留摘要字段，真实内容在 `payload_file`
- 开发者只需要查看 `logs` 和 `payloads`，不依赖额外 UI 或导出工具
