# 全量日志模块全面重构设计

## 1. 背景

当前仓库已经存在基础日志能力，但整体处于“可用但不成体系”的状态。

已确认的现状如下：

- 仓库已有 `backend/logger`，底层使用 `zap + lumberjack`，并已落盘 `info.log` 与 `error.log`
- `main.go`、`app.go`、`app_windows.go`、`app_linux.go`、`app_darwin.go`、`backend/agent/*`、`backend/data/*` 中存在大量直接调用 `SugaredLogger` 的代码
- 仍有部分正式代码使用标准库 `log`、`fmt.Printf` 或各组件自带日志入口，日志出口并不统一
- `Wails` 当前通过 `logger.NewFileLogger(runtimePaths.WailsLogPath)` 形成独立日志源
- `backend/db` 当前将 GORM logger 设为 `Silent`，数据库慢查询、错误 SQL 和关键写入过程未纳入统一日志体系
- `ai-assistant-web` 作为本地 HTTP 子服务已经独立启动，但请求、响应和错误记录尚未统一进全链路日志
- 前端异常、Promise 未处理错误、关键 console error 没有统一回传后端落盘

因此，本次目标不是继续“补几行日志”，而是把仓库重构成真正的全链路日志系统。

## 2. 已确认需求

本次需求边界已确认如下：

- 选择“全面重构”方案，而不是最小侵入或局部扩展
- 目标范围为真正全量日志，覆盖后端、运行时、本地 HTTP 子服务、数据库、定时任务、前端异常与 AI 对话链路
- 日志默认常开，不依赖手动切换“调试模式”
- 日志以本地落盘为唯一消费方式，开发者自行查看 `logs` 目录
- 不做软件内置日志查看
- 不做一键导出
- 采用“原始内容优先”策略，排障时尽可能保留原文

需要明确说明的风险：

- AI 对话正文、请求参数、抓取结果、部分用户输入和响应原文都会落盘
- 该设计只适合本机排障，不适合直接共享日志或上传到外部系统
- 日志体积会明显增大，必须配套物理分流、滚动和清理策略

## 3. 目标

本次改动需要同时满足以下目标：

- 所有正式日志统一经过同一套日志基础设施输出
- 支持通过 `trace_id` 将一次前端操作、一次 Wails 调用、一次 AI 调用、一次 HTTP 请求、一次数据库写入串成可追踪链路
- 将启动、关闭、panic、前端错误、HTTP、AI、工具调用、数据库、定时任务、抓取与外部请求统一纳入结构化日志
- 默认常开时仍能维持可用，避免大体积原文直接打爆主日志文件
- 为后续开发建立强约束，避免仓库再次退化为散乱日志写法

## 4. 非目标

以下内容不在本次范围内：

- 不做日志 UI
- 不做一键导出
- 不做远程日志上传
- 不做告警平台集成
- 不把本次需求扩展成监控平台建设
- 不借重构日志体系之机重构业务流程本身

## 5. 方案对比

### 方案 A：最小侵入扩展

继续沿用现有 `backend/logger`，仅在缺失位置补充更多日志调用。

优点：

- 改动最小
- 落地最快

缺点：

- 仍会保留大量不统一写法
- 难以形成全链路关联
- 后续维护成本持续升高

### 方案 B：统一日志门面

保留现有 `zap` 技术栈，但升级 `backend/logger` 为统一门面，并逐步替换关键入口。

优点：

- 可维护性明显提升
- 成本可控

缺点：

- 仍存在阶段性新旧并存
- 对“真正全量”要求不够彻底

### 方案 C：全面重构日志体系（已选定）

将日志能力重构成统一基础设施层，正式代码不再允许散用标准库 `log`、`fmt.Printf`、零散 `SugaredLogger` 或各子系统自建 logger。

优点：

- 最接近真正的“全量日志模块”
- 结构清晰，后续扩展成本最低
- 能对仓库形成长期强约束

缺点：

- 改动面最大
- 如果边界不清晰，容易把业务代码一起搅乱

### 结论

采用方案 C，但边界明确为“全面重构日志体系，不重构业务功能本身”。

## 6. 总体架构

本次设计将日志能力收敛为统一基础设施层，所有正式日志都必须通过该层输出。

建议的核心文件划分如下：

- `backend/logger/bootstrap.go`
  - 负责全局初始化、配置装配、日志目录准备、输出目标构建
- `backend/logger/factory.go`
  - 负责按模块创建 logger、派生上下文 logger、附加链路字段
- `backend/logger/types.go`
  - 负责定义统一字段模型和事件常量
- `backend/logger/sinks.go`
  - 负责文件分流、滚动、写入器与原文载荷落盘策略
- `backend/logger/runtime.go`
  - 负责 panic recover、goroutine 包装、耗时记录、刷盘等运行时辅助能力
- `backend/logger/http.go`
  - 负责本地 HTTP 请求与响应日志的统一封装
- `backend/logger/wails.go`
  - 负责 Wails 生命周期、绑定调用与前端错误桥接
- `backend/logger/db.go`
  - 负责 GORM logger 适配与数据库事件标准化输出

架构原则如下：

- 业务代码只能通过统一工厂获取 logger
- 不允许正式代码继续直接依赖全局 `SugaredLogger`
- 每条日志都必须具备统一公共字段
- 大体积原文与摘要日志分离存放
- 错误和 panic 必须可从单点入口回溯完整链路

## 7. 日志文件布局

本次不再以“仅按级别”分文件，而是改为“按排障场景分流，同时保留错误聚合”。

建议产出以下文件：

- `logs/app.log`
  - 应用主日志，记录启动、关闭、模块级业务事件、Wails 生命周期
- `logs/error.log`
  - 所有错误级别日志聚合文件，作为故障排查第一入口
- `logs/http.log`
  - 本地 HTTP 服务请求、响应与中间件日志
- `logs/ai.log`
  - AI 对话、工具调用、模型请求与响应日志
- `logs/task.log`
  - 定时任务、后台任务、监控任务执行日志
- `logs/db.log`
  - 数据库慢 SQL、错误 SQL、关键写入事件
- `logs/frontend.log`
  - 前端异常、Promise 错误、关键前端事件
- `logs/panic.log`
  - panic 与堆栈日志

同时增加原文分流目录：

- `logs/payloads/YYYY-MM-DD/...`

规则如下：

- 小体积内容可直接进入主日志
- 超过阈值的请求体、响应体、AI 原文、抓取 HTML、堆栈原文改为写入 `payloads`
- 主日志仅记录摘要、字节数、哈希、`payload_file` 和关键上下文字段

这样做的原因是：在“默认常开 + 原始内容优先”前提下，如果所有原文都直接塞进主日志，日志文件会很快失去可读性和可维护性。

## 8. 统一字段模型

每条日志都至少包含以下公共字段：

- `ts`
- `level`
- `module`
- `event`
- `trace_id`
- `span_id`
- `source`
- `message`

常用扩展字段包括：

- `component`
- `path`
- `method`
- `status_code`
- `duration_ms`
- `request_id`
- `app_session_id`
- `ai_session_id`
- `task_id`
- `task_name`
- `stock_code`
- `payload_file`
- `error_class`
- `error_message`
- `error_code`
- `retryable`
- `stack`

设计原则：

- 字段优先使用结构化键值，不依赖拼接长文本表达语义
- 错误日志必须带事件名和关键上下文
- 同一链路的不同步骤必须共享同一 `trace_id`

## 9. 日志数据流设计

### 9.1 应用启动与 Wails 生命周期

`main.go` 与 `app*.go` 统一记录以下事件：

- 应用启动开始与完成
- 运行时路径初始化
- 版本信息
- 窗口尺寸、屏幕分辨率、主题与 Webview 初始化
- 应用关闭与退出前清理
- 生命周期中的 panic recover

同一次应用运行生成一个 `app_session_id`，后续日志默认带上该字段。

Wails 绑定方法需要统一包装，记录：

- 方法名
- 参数原文
- 调用开始时间
- 返回结果摘要
- 错误信息
- 调用耗时

### 9.2 前端异常回传

前端统一桥接以下异常到后端：

- `window.onerror`
- `unhandledrejection`
- 关键 `console.error`
- 当前路由
- 组件名
- 错误栈
- 附加业务上下文

该部分统一写入 `frontend.log`，并复用 `trace_id` 与后端方法调用链路关联。

### 9.3 本地 HTTP 子服务

`ai-assistant-web` 增加统一日志中间件，记录：

- 请求方法、路径、查询参数、来源地址
- 请求头
- 请求体原文或其分流文件路径
- 响应状态码
- 响应体原文或其分流文件路径
- 耗时

正常事件进入 `http.log`，错误同时汇总进 `error.log`。

### 9.4 AI 对话与工具调用

每次 AI 会话需要记录：

- `ai_session_id`
- 模型名
- `base_url`
- system prompt、user prompt、messages 原文
- 流式分片事件
- 最终聚合结果
- token 使用量
- 工具调用入参、出参、错误

该部分统一进入 `ai.log`。

保留“流式分片 + 最终汇总”双视角：

- 分片日志用于还原流式过程
- 汇总日志用于快速检索最终结论

### 9.5 数据抓取、外部请求、数据库与定时任务

`backend/data`、`backend/agent`、`backend/db` 的关键链路统一采用“开始 / 成功 / 失败”三段式事件模型。

外部请求至少记录：

- URL
- 方法
- 参数
- 响应摘要
- 重试次数
- 失败原因

Chromedp 抓取至少记录：

- 目标页面
- 关键步骤
- 抓取原文或原文文件路径
- 失败节点

数据库日志至少记录：

- SQL
- 参数
- 影响行数
- 慢查询耗时
- 错误信息

定时任务日志至少记录：

- 任务 ID
- 任务名
- 入参
- 开始时间
- 结束时间
- 结果摘要
- 错误信息

## 10. 错误处理与 panic 策略

方案 3 下，不允许仅写一句错误文本而缺乏上下文。

所有失败日志统一补齐以下字段：

- `error_class`
- `error_message`
- `error_code`
- `retryable`
- `op`
- `stack`
- `trace_id`
- `span_id`
- `payload_file`

panic 处理统一收口到日志基础设施，不再允许各处零散 `recover`：

- `main.go` 主入口统一 recover
- Wails 绑定方法统一包装 recover
- HTTP handler 统一 recovery middleware
- goroutine 启动统一通过包装器
- 定时任务执行器统一 recover 并带任务上下文

设计目标是：`panic.log` 中的任何一条记录，都可以通过 `trace_id` 反查到对应的 `app.log`、`http.log`、`ai.log` 或 `task.log` 上下文。

## 11. 日志格式、滚动与保留策略

建议统一使用 `JSON Lines` 格式，全部按 UTF-8 落盘。

原因如下：

- 结构化字段较多，纯文本格式难以稳定 grep 与关联
- 需要同时承载固定字段、扩展字段和载荷文件引用
- 后续如需脚本化分析，本地 JSON Lines 更容易处理

保留策略建议如下：

- `app.log`、`error.log`、`task.log`、`db.log`、`frontend.log`、`panic.log`
  - 保留 30 天
  - 启用滚动压缩
- `http.log`、`ai.log`
  - 保留 7 到 14 天
  - 启用滚动压缩
- `logs/payloads`
  - 保留 7 天
  - 增加目录总量上限，例如 `5 GB`
  - 超限后按最老文件优先清理

此策略用于同时控制时间维度和磁盘容量维度，避免短时间高频调用直接打满磁盘。

## 12. 平滑重构方式

虽然最终目标是全面重构，但过程不采用“单次全仓暴力替换”，而采用“目标全面重构，过程分层替换”的方式。

实施顺序如下：

1. 建立新的日志基础设施
   - `bootstrap`
   - `factory`
   - `types`
   - `sinks`
   - `runtime`
   - `http`
   - `wails`
   - `db`
2. 统一入口
   - `main.go`
   - `app.go`
   - `app_windows.go`
   - `app_linux.go`
   - `app_darwin.go`
   - `ai-assistant-web/server.go`
   - `backend/db`
3. 替换核心链路
   - `backend/agent/*`
   - `backend/data/*` 中 AI、外部请求、抓取、定时任务相关模块
4. 强约束收口
   - 清理正式代码中的标准库 `log`
   - 清理正式代码中的 `fmt.Printf`
   - 清理直接访问旧全局 `SugaredLogger`
   - 增加扫描或测试规则，阻止新代码回退到旧模式

也就是说，实施顺序是渐进的，但最终状态是硬切的，不允许新旧日志体系长期并存。

## 13. 兼容性与已知改造点

实现阶段需要重点处理以下兼容点：

- `Wails` 当前使用 `logger.NewFileLogger(runtimePaths.WailsLogPath)`，这会继续形成独立日志源
  - 方案 3 下应桥接到统一 logger，而不是维持单独 `wails.log`
- `backend/db` 当前使用 `gorm logger.Silent`
  - 方案 3 下应替换为自定义 GORM logger，统一输出慢 SQL、错误 SQL 和关键写入事件到 `db.log`

## 14. 测试策略

测试重点不在“能否打印日志”，而在以下三类风险：

- 是否漏链路
- 是否打爆磁盘
- 是否让旧调用方式继续存活

建议分四层验证：

### 14.1 单元测试

验证日志基础设施本身行为正确：

- 模块 logger 是否自动带公共字段
- `trace_id` / `span_id` 是否能正确继承
- 大 payload 是否正确分流到 `logs/payloads`
- panic wrapper 是否能稳定写入 `panic.log`
- GORM / HTTP / Wails adapter 是否写入正确日志文件

### 14.2 集成测试

直接跑关键链路，检查多个日志文件是否同步产生日志：

- 应用启动链路
- AI 对话链路
- 本地 HTTP 请求链路
- 定时任务执行链路
- 外部抓取失败链路
- 数据库错误与慢 SQL 链路

### 14.3 约束测试

防止仓库回退到旧日志写法：

- 扫描正式 Go 代码中的标准库 `log`
- 扫描正式 Go 代码中的 `fmt.Printf`
- 扫描直接使用旧全局 `SugaredLogger`
- 对测试代码或极少数兼容桥接文件做明确例外

### 14.4 手工验收

必须做真实落盘验证：

- 启动应用后确认 `logs` 目录结构正确
- 人工触发一次前端错误
- 人工触发一次 HTTP 请求
- 人工触发一次 AI 调用
- 人工触发一次任务失败
- 人工检查日志文件和 `payloads` 文件是否可读、可关联、可追踪

## 15. 验收标准

以下条件同时满足，才视为本次设计落地完成：

- 所有正式日志统一经过新 `backend/logger` 门面输出
- 正式代码不再直接使用标准库 `log` 与 `fmt.Printf` 作为业务日志
- `main`、Wails 生命周期、前端错误、HTTP 子服务、AI 对话、工具调用、数据库、定时任务都已接入统一日志
- 至少生成以下日志文件：
  - `app.log`
  - `error.log`
  - `http.log`
  - `ai.log`
  - `task.log`
  - `db.log`
  - `frontend.log`
  - `panic.log`
- 大体积原文已经物理分流到 `logs/payloads`
- 同一条关键链路可通过 `trace_id` 串起前后文
- 错误日志包含事件名、错误分类和关键上下文
- panic 能单独落盘，并附带 stack trace
- 默认常开时，单条超大内容不会破坏主日志可用性
- 存在自动化测试覆盖日志基础设施和至少两条关键集成链路

## 16. 后续实现边界

实现阶段只做以下事情：

- 建立新的日志基础设施层
- 接入 Wails、数据库、本地 HTTP 子服务、AI、任务、抓取和前端异常链路
- 清理正式代码中的旧日志出口
- 补充最小关键测试与约束测试

以下内容留到后续版本再讨论：

- 日志 UI
- 一键导出
- 远程日志上传
- 告警平台联动
- 在日志体系之外追加指标、监控或审计平台建设
