# 全量日志模块闭环补充设计

## 1. 文档定位

本文档是对 [2026-04-05-full-logging-rebuild-design.md](./2026-04-05-full-logging-rebuild-design.md) 的闭环补充，不替代原始重构设计。

原始设计已经定义了全量日志体系的目标、架构、文件布局和验收标准；本补充设计只解决当前实现与第 15 节验收标准之间仍未闭合的差距。

当前基线实现位于提交：

- `1d8e381 refactor: complete full logging rebuild`

## 2. 当前未闭环点

基于当前实现审查，以下问题会阻止本次需求被判定为“已按设计完成”：

- `trace_id` 已普遍存在，但很多链路是在不同模块现场重新生成，尚不能稳定证明“同一条关键链路共享同一个 `trace_id`”
- 前端异常已接入 `window error` 与 `unhandledrejection`，但尚未将关键 `console.error` 统一桥接到后端落盘
- 自动化测试已覆盖日志基础设施和部分接线，但尚未直接证明至少两条关键链路可以带着同一个 `trace_id` 贯穿多个 sink
- 设计文档第 `14.4` 节要求的真实落盘手工验收尚未沉淀为可重复执行的验证资产
- 普通错误日志普遍具备事件名和上下文，但错误分类字段尚未形成统一补齐策略

因此，本次补充设计的目标不是再次重构日志架构，而是把现有实现补到可以按原设计第 `15` 节口径宣称“验收完成”。

## 3. 本次补齐目标

本次补齐完成后，需要同时满足以下条件：

- 至少两条关键链路能通过同一个 `trace_id` 跨 sink 串联
- 前端关键 `console.error` 会进入统一前端错误桥接链路，同时不破坏开发态控制台输出
- 自动化测试能够直接证明关键链路的真实落盘与 `trace_id` 复用
- 手工验收步骤被文档化，且可由开发者按步骤检查 `logs` 与 `payloads`
- 普通错误日志至少具备统一的 `error_class` 字段，满足“错误分类”验收口径

## 4. 非目标

以下内容仍不在本次范围内：

- 不重做全量日志架构
- 不新增日志 UI
- 不新增远程日志上传、导出、告警或监控
- 不借本次工作重构 AI、任务、数据库或前端业务流程
- 不强制把所有业务函数签名都改造成显式传递 `context.Context`

## 5. 设计原则

本次补齐遵循以下原则：

- 优先复用现有 `backend/logger` 基础设施，而不是引入第二套上下文模型
- 优先在关键入口与公共 helper 层补足 trace 透传能力，而不是在业务层做大面积签名改造
- 测试必须直接证明验收目标，而不是只证明“代码执行过”
- 手工验收必须沉淀为可复跑步骤，避免最终仅靠口头说明

## 6. Trace 闭环方案

### 6.1 目标

将 `trace_id` 的语义从“每条日志都有字段”提升为“同一条关键链路复用同一个链路标识”。

### 6.2 补充能力

在 `backend/logger` 中补充轻量上下文传递能力：

- 支持将 `TraceContext` 写入 `context.Context`
- 支持从 `context.Context` 中读取已有 `TraceContext`
- 支持“有 trace 就复用，无 trace 才新建”的统一 helper

该能力只负责 trace 的携带与复用，不引入新的复杂 request scope 容器。

### 6.3 入口规则

#### HTTP 链路

- HTTP middleware 在请求进入时生成一次 `TraceContext`
- 将该 `TraceContext` 绑定到 `request context`
- middleware 内部的请求体读取告警、完成日志、响应日志都必须复用这一份 trace
- handler、AI 调用、工具调用、DB 日志如果拿得到该 request context，则必须优先复用该 trace

#### Wails / 前端错误链路

- Wails 生命周期入口仍可按当前模式新建 trace
- 前端错误回传 payload 增加可选 `traceId`
- 前端在可关联的情况下将当前 trace 一并上报
- 后端收到 `frontendError` 事件后，若 payload 自带 `traceId`，则必须复用；若没有，才兜底生成新的 trace

#### 定时任务链路

- `ExecuteTask` 开始时生成一次任务级 trace
- 同一次任务执行过程中产生的任务日志、AI 日志、工具日志、数据库日志优先复用这份 trace
- 任务内调用下游 helper 时，如果 helper 支持从 `context.Context` 取 trace，则统一透传

### 6.4 覆盖范围

本次必须覆盖以下三类关键链路：

- `HTTP -> middleware -> AI / tool / DB`
- `task -> AI / tool / DB`
- `frontend error emit -> Wails event -> frontend.log`

不要求一次性覆盖仓库内所有边缘链路，但上述三类链路必须成为可验证的标准样板。

## 7. 前端关键错误桥接补齐

### 7.1 当前问题

现有实现已覆盖：

- `window error`
- `unhandledrejection`
- Vue `app.config.errorHandler`

但仍有大量组件直接使用 `console.error(...)`，这些日志并不会自动进入统一桥接链路。

### 7.2 方案

在前端错误桥接模块中增加轻量 `console.error` 代理：

- 保留原始 `console.error` 行为，控制台仍然照常输出
- 对 `Error`、普通对象、字符串错误进行统一归一化
- 归一化后通过现有 `frontendError` 事件补充上报
- 对同一轮 `error` / `unhandledrejection` 已上报过的错误对象做短窗口去重，避免重复落盘

### 7.3 去重要求

以下情况要避免双报：

- 同一个 `Error` 对象先被 `unhandledrejection` 捕获，又被代理的 `console.error` 输出
- 已经由 Vue `errorHandler` 上报的异常，再被组件内部 `console.error(err)` 追加上报

本次只做短生命周期去重，不追求跨页面或长时间窗口去重。

## 8. 错误分类补齐策略

原设计要求错误日志包含“错误分类”。当前 panic 已具备 `error_class=panic`，但普通错误还没有统一策略。

本次采用最小必要收口：

- panic recover 继续使用 `error_class=panic`
- 前端错误统一使用 `error_class=frontend_error`
- HTTP 中间件错误统一使用 `error_class=http_error`
- 数据库日志错误统一使用 `error_class=db_error`
- 定时任务执行错误统一使用 `error_class=task_error`
- AI / 工具调用错误统一使用 `error_class=ai_error` 或 `error_class=tool_error`

本次不强制所有普通错误都补齐 `error_code` 与 `retryable`，但要求关键错误事件至少具备：

- `event`
- `error_class`
- `error_message`
- `trace_id`
- 关键上下文字段

## 9. 自动化测试补齐

### 9.1 单元测试

新增或补强以下测试：

- 同一次 HTTP 请求中的多条中间件日志共享同一 `trace_id`
- 前端错误 payload 自带 `traceId` 时，后端落盘保留该值
- 前端错误 payload 不带 `traceId` 时，后端兜底生成 trace
- 给定已有 trace 的 helper 在生成 logger 时不会重新 `NewTrace()`
- `console.error` 代理已安装、原始 `console.error` 仍被调用、归一化上报符合预期

### 9.2 集成测试

至少新增两条直接检查真实落盘内容的关键链路测试：

- `HTTP -> log file / payload spill` 链路
  - 检查 `http.log`
  - 检查 `payloads`
  - 检查同链路事件的 `trace_id` 一致性
- `task -> downstream sink` 链路
  - 至少覆盖 `task.log` 与下游 `ai.log` 或 `db.log`
  - 检查同一任务执行过程中的 `trace_id` 一致性

若现有业务链路过重，可通过测试专用 handler / stub 执行器实现，但必须真实走日志基础设施并落盘。

### 9.3 约束测试

继续保留并通过现有 guard：

- 正式 Go 代码中禁止回退到标准库 `log`
- 正式 Go 代码中禁止回退到 `fmt.Printf`
- 正式 Go 代码中禁止直接使用旧 `SugaredLogger`

## 10. 手工验收资产

为满足原设计第 `14.4` 节，本次需要新增一份简短的手工验收文档，内容只包含：

- 启动应用后的 `logs` 目录检查
- 触发一次前端错误并检查 `frontend.log`
- 触发一次 HTTP 请求并检查 `http.log` 与 `payloads`
- 触发一次 AI 调用并检查 `ai.log`
- 触发一次任务失败并检查 `task.log`
- 检查 `trace_id`、`payload_file`、`error_class` 是否满足预期

该文档的目标是让开发者打开 `logs` 目录就能照着查，而不是提供额外工具或脚本。

## 11. 实现边界与顺序

本次实现严格按以下顺序推进：

1. 先写失败测试，锁定 `trace` 复用、前端 `console.error` 桥接、关键链路落盘行为
2. 再补 `backend/logger` 的 trace 透传 helper
3. 再补 HTTP / Wails / 任务入口的 trace 复用接线
4. 再补前端 `console.error` 代理与后端前端错误 trace 复用
5. 再补关键链路集成测试与手工验收文档
6. 最后重新按原始设计第 `15` 节逐项验收

如果实现过程中发现必须对大量业务函数签名做系统性改造，说明当前补齐方案边界失效，需要暂停并重新评估，不允许无提示扩 scope。

## 12. 完成定义

以下条件同时满足，才可宣称“全量日志模块已按设计闭环完成”：

- 原始设计文档第 `15` 节中除非目标外的验收项全部有直接证据
- 至少两条关键链路经自动化测试证明可复用同一个 `trace_id`
- 前端关键 `console.error` 已进入统一桥接链路，且不破坏原始控制台输出
- 普通错误日志已具备统一 `error_class`
- 手工验收文档存在，且开发者可按文档检查真实落盘文件

在此之前，只能表述为“基础设施已基本完成”，不能表述为“验收闭环完成”。
