# go-stock 项目级 AGENTS

## 1. 优先级与沟通

- 用户直接指令优先级最高。
- 外层系统、流程技能和平台约束高于本文件。
- 本文件用于约束本仓库内 Go 后端、Wails 桥接、前端 JS/Vue 的默认开发方式。
- 默认使用简体中文；结论优先；表达简洁。
- 事实不确定时必须明确写出“无确切信息”或“需补充材料”，禁止把猜测写成事实。

## 2. 核心原则

- 决策顺序固定为：`Correctness > Maintainability > Observability`。
- “第一性原理 / 如非必要勿增实体”是本仓库硬约束，不是口号。
- 在新增文件、类型、接口、配置、日志字段、测试夹具、封装层之前，必须先回答三个问题：
  1. 为什么现有实现不能承载？
  2. 这个新增项独占什么职责？
  3. 删除它会失去什么能力？
- 上述三个问题答不清，就不要新增。
- 禁止为了“看起来更整洁”“先抽一层”“以后可能复用”而引入额外抽象、包装器、helper 层、配置层。
- 默认优先复用现有模块、现有命名、现有测试入口和最短可验证路径。

## 3. 日志范式

- 正式日志范围包括：非测试 Go 代码、Wails 生命周期与桥接、前端异常回传、HTTP、数据库、任务、AI 链路日志。
- 新增正式日志必须优先走 `backend/logger` 统一门面。默认从 `logger.Default().ForSink(...)` 或现有 helper 获取 logger，不在业务代码里自建并行 logger。
- 非测试 Go 代码禁止新增以下业务日志写法：
  - `fmt.Printf`
  - `log.Print` / `log.Println` / `log.Printf` / `log.Fatal*`
  - `logger.SugaredLogger`
- 现有历史残留不是新代码继续沿用的依据；以 `backend/logger/legacy_usage_test.go` 的守卫方向为准。
- 复用现有 sink 语义，不擅自新增并行日志出口。当前统一 sink 包括：
  - `SinkApp`
  - `SinkError`
  - `SinkHTTP`
  - `SinkAI`
  - `SinkTask`
  - `SinkDB`
  - `SinkFrontend`
  - `SinkPanic`
- 需要链路关联时，沿用已有字段语义：
  - `trace_id`
  - `span_id`
  - `app_session_id`
  - `source`
- 禁止重新发明并行字段名，例如自造 `requestTrace`、`sessionTrace`、`traceId2` 一类变体。
- 日志事件名统一采用可检索格式，推荐 `<domain>.<action>.<result>`，例如：
  - `http.request.completed`
  - `frontend.error`
  - `db.query.failed`
- 关键链路默认采用“开始 / 成功 / 失败”三段式事件模型；不要只写一条模糊的自然语言日志。
- 失败日志必须带足够上下文，至少包含：
  - 事件名
  - 错误对象或错误信息
  - 关键业务键，例如 `path`、`task_name`、`stock_code`、`status_code`
- 大体积载荷、前端异常、Wails 桥接错误、HTTP/DB/任务日志，优先复用现有 payload、trace 和 sink 机制，不新增散落文件或自定义落盘路径。
- 前端运行时错误应复用现有前端错误桥接约定，不得只停留在 `console.error` 而不回传后端链路。
- 测试代码默认不要用生产 logger 验证普通业务行为。普通测试调试优先使用 `t.Log` / `t.Logf`。
- 只有在“测试日志系统本身”时，才直接断言日志内容、payload 文件、trace 传播或 guard 行为；此类测试优先放在 `backend/logger` 附近。

## 4. 测试范式

- 默认先写“最小可证明”的测试，再扩展实现；不要先堆大而全的测试矩阵。
- Go 纯逻辑优先使用表驱动测试，测试文件靠近被测包放置。
- IO、文件、HTTP、数据库相关测试优先使用：
  - `t.TempDir()`
  - `httptest`
  - 临时数据库或临时文件
- 默认避免真实外网依赖、真实 GUI 依赖和不可控系统状态；能隔离就先隔离。
- 需要真实集成链路时，沿用仓库已有门槛和隔离方式。当前集成测试默认通过 `requireIntegrationTest(t)` 控制，并要求显式设置 `GO_STOCK_RUN_INTEGRATION_TESTS=1`。
- 日志相关改动优先补到以下测试方向，而不是只看控制台输出：
  - `backend/logger` 守卫测试
  - trace 传播测试
  - payload 行为测试
  - closeout / integration 测试
- 前端纯逻辑和桥接逻辑优先使用现有 `node --test` 风格的 `.mjs` 测试，沿用当前仓库已有模式，例如：
  - `frontend/src/utils/*.test.mjs`
  - `frontend/src/components/**/hoverTooltip.test.mjs`
- UI 变更如果不适合自动化，至少补充最小手工验证步骤，说明要看什么行为、在哪个页面复现。
- `npm run build`、`go build`、`wails build` 只属于附加验证，不替代行为测试；“能编译”不等于“已验证正确”。
- 测试代码同样遵守“如非必要勿增实体”：
  - 不新增巨型 fixture
  - 不默认引入全局 mock 框架
  - 不创建无边界测试 helper
- 只有当重复成本已经被证明存在时，才允许抽测试工具层。

## 5. 工具与构建范式

- 本仓库是 Go/Wails 项目，不套用 Python、conda 或其他无关工具链的默认流程。
- 日常验证优先直接使用：
  - `go`
  - `node`
  - `npm`
  - Wails CLI
- `C:\Users\Aspir\go\bin` 下当前可用工具包括：
  - `wails.exe`：构建与打包
  - `dlv.exe`：调试
  - `gopls.exe`：代码智能与跳转
  - `gotests.exe`：测试脚手架起点
  - `impl.exe`：接口实现脚手架起点
  - `goplay.exe`：Go 片段快速验证
- 对 `gotests.exe`、`impl.exe` 生成的内容必须人工收敛、删冗余、补语义；禁止机械生成后直接提交。
- 不要假设 `C:\Users\Aspir\go\bin` 已经在 PATH 中。需要调用这些工具时，优先使用绝对路径，或在命令里显式补 PATH。
- 需要 Windows 发布产物或安装包时，唯一发布入口是仓库根目录 `publish_process.txt` 所表达的流程。
- 产出 Windows 可运行程序或安装包时，统一执行与下述内容等价的 PowerShell 流程，不允许改走裸 `wails build`，也不允许把 `scripts/build-windows.sh` 当成发布规范源：

```powershell
$env:PATH = (Join-Path $env:LOCALAPPDATA 'NSIS3\Bin') + ';' + $env:PATH
& 'C:\Users\Aspir\go\bin\wails.exe' build --clean --platform windows/amd64 --nsis
```

- 开发期快速验证允许使用轻量命令，但这些命令不替代正式发布打包：
  - 定向 `go test`
  - 必要时 `go test ./...`
  - 现有 `node --test`
  - `frontend` 目录下的 `npm run build`
- 前端当前没有统一 `npm test` 脚本时，不要杜撰一个新的测试入口；直接使用仓库里已经存在的 `node --test` 命令。

## 6. 提交前自查

- 我这次改动是否真的需要新增实体？如果不需要，先删掉多余抽象。
- 新增日志是否复用了现有 `backend/logger` 门面、sink 和 trace 语义？
- 失败日志是否带了错误对象与关键业务上下文？
- 我补的是“最小可证明”测试，还是只是把构建命令跑通？
- 我使用的是开发验证命令，还是误把发布命令当日常校验？
- 如果本次需要发布 Windows 产物，是否严格按 `publish_process` 等价流程执行？
