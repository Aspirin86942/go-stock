# go-stock 整体架构重构 spec 闭环验收

## 1. 结论

截至 2026-04-07，`factor` 分支上的累计实现已经满足 [2026-04-06-go-stock-overall-architecture-refactor-design.md](D:/codex_work/go-stock/docs/superpowers/specs/2026-04-06-go-stock-overall-architecture-refactor-design.md) 的闭环目标。

这里的“闭环”指的是：

- 整体重构要求的阶段性迁移已经全部落到 `factor`
- Phase 5 所要求的旧路径收缩和 guard 机制已经完成
- `factor` 当前状态不再需要“从 final-closure plan 开头重新做一遍”

本文件记录的是 **`factor` 当前累计状态的验收结论**，不是某一条单独提交的变更说明。

## 2. 分阶段对照

### 2.1 阶段 0：日志与测试底座

状态：已闭环。

已满足的目标：

- 日志统一与测试分层底座已在当前分支基线中存在
- 默认测试门禁、`external` / `release-smoke` 分层与 logger closeout 已落地
- 架构迁移已经建立在可验证底座之上，而不是裸改业务路径

当前分支中的直接证据包括：

- `9683bef test: finish test layering and logging closure`
- `backend/logger/*`
- `internal/testenv`
- `scripts/testing`

### 2.2 阶段 1：建立新边界

状态：已闭环。

已满足的目标：

- `backend/source` 已建立并承接真实来源适配
- `backend/service` 已建立并承接稳定业务用例
- `frontend/src/pages` 与 `frontend/src/services` 已建立并承接真实链路

当前分支中的直接证据包括：

- `backend/source/marketnews`
- `backend/source/analysis`
- `backend/service/market`
- `backend/service/analysis`
- `frontend/src/pages`
- `frontend/src/services`

### 2.3 阶段 2：市场只读链路迁移

状态：已闭环。

已满足的目标：

- 市场只读链路已经走 `bridge -> service -> source`
- 前端市场读链路已经走 `router -> pages -> services`
- 后续 residual market reads 也已在 Phase 5 中完成收口

当前分支中的直接证据包括：

- `289f182 refactor: add market read service contracts`
- `1556c37 refactor: bridge market read service through app`
- `b6b3ddf refactor: add frontend market service wrappers`
- `7eb4eec refactor: migrate market read path to page and service layers`
- `08de9b2 refactor: close residual market read boundaries`

### 2.4 阶段 3：AI 分析链路迁移

状态：已闭环。

已满足的目标：

- 分析型 AI 链路已经从旧混合路径收回到 analysis boundary
- AI 结果读取、历史结果访问与前端页面壳已经走新边界
- 整体 spec 明确排除的多轮 assistant/agent 深改没有被错误纳入这轮收口

当前分支中的直接证据包括：

- `e85e90a refactor: add analysis service boundary`
- `7c9e6e4 refactor: bridge analysis service through app`
- `d1e2882 refactor: add frontend analysis service wrappers`
- `4bb8095 refactor: migrate analysis screens to page and service layers`

### 2.5 阶段 4：任务、预警与设置迁移

状态：已闭环。

已满足的目标：

- config / task / notification / watchlist 相关副作用链路已经被 service boundary 吸收
- 设置页和 cron task 页已经通过 `pages/**` 与 `services/**` 收口
- 任务、通知、watchlist 与 app bridge 的分工已经明确

当前分支中的直接证据包括：

- `bd4325c refactor: add shared config service boundary`
- `c33e263 refactor: add cron task service boundary`
- `73b8fef refactor: add analysis artifact and config export services`
- `db3bd49 refactor: close out app bridge notifications and watchlist`
- `7bf8800 fix: preserve cron start signal and cache fallback`

### 2.6 阶段 5：清理旧路径

状态：已闭环。

这一步的最终收口已经并入 `factor`，关键提交为：

- `08de9b2 refactor: close residual market read boundaries`
- `fa0ff16 fix: restore legacy market read compatibility`
- `5a404d7 fix: preserve legacy realtime price failure contract`
- `6ea6393 fix: restore realtime price compatibility`
- `8955763 fix: guard app shell group list reads`
- `3e81031 refactor: close watchlist and research boundaries`
- `b9b624d refactor: finish route pages and assistant compatibility fence`
- `ed3d29e test: add architecture closeout guards`

## 3. Phase 5 验收矩阵

### 3.1 `app.go` / `app_common.go` 收敛为 bridge adapter

状态：通过。

验收点：

- bridge 不再直接持有被禁止的 `backend/data` 构造器
- 残余 business orchestration 已收回 `market` / `watchlist` / `research` / `fund` service

直接证据：

- `architecture_closeout_test.go`
- `backend/service/market`
- `backend/service/watchlist`
- `backend/service/research`
- `backend/service/fund`

### 3.2 前端 direct Wails import 仅保留在 `services/**`

状态：通过。

验收点：

- 组件层和页面层不再直接 import `wailsjs/go/main/App`
- 剩余 direct import 只保留在 `frontend/src/services/**`

直接证据：

- `architecture_closeout_test.go`
- `frontend/src/services/*.mjs`

### 3.3 路由统一走 `pages/**`

状态：通过。

验收点：

- `frontend/src/router/router.js` 不再直接引用 `../components/**`
- 路由入口统一转为 `frontend/src/pages/**`

直接证据：

- `frontend/src/router/router.js`
- `frontend/src/pages/about-page.vue`
- `frontend/src/pages/agent-page.vue`
- `frontend/src/pages/fund-page.vue`

### 3.4 assistant compatibility island 已显式围栏

状态：通过。

验收点：

- 多轮 assistant / agent bridge 没有继续扩张为新改造面
- 保留项被视为明确 allowlist，而不是漏网旧路径

直接证据：

- `app.go` 中的 compatibility allowlist 注释
- `frontend/src/services/assistantService.mjs`

### 3.5 仓库具备自动 guard，防止 Phase 5 回归

状态：通过。

验收点：

- bridge direct-data regression 有自动 guard
- frontend direct Wails import regression 有自动 guard
- 路由回指 `components/**` 有自动 guard

直接证据：

- `architecture_closeout_test.go`

## 4. 2026-04-07 新鲜验证证据

以下命令已在 `factor` 分支当前状态重新执行并通过：

### 4.1 Go guard

```powershell
& 'C:\Program Files\Go\bin\go.exe' test . -run 'Test(AppBridgeNoLongerOwnsBusinessDataApis|FrontendDirectWailsImportsAreLimitedToServices|RouterUsesPagesForAllRouteComponents)' -count=1
```

结果：`PASS`

### 4.2 Go closeout 验证

```powershell
& 'C:\Program Files\Go\bin\go.exe' test ./backend/service/market ./backend/service/watchlist ./backend/service/research ./backend/service/fund ./backend/service/config ./backend/service/analysis ./backend/service/notification ./backend/service/task . -count=1
```

结果：`PASS`

### 4.3 前端 service 验证

```powershell
node --test frontend/src/services/marketService.test.mjs frontend/src/services/appShellService.test.mjs frontend/src/services/watchlistService.test.mjs frontend/src/services/researchService.test.mjs frontend/src/services/fundService.test.mjs frontend/src/services/assistantService.test.mjs
```

结果：`PASS`

说明：

- Node 运行时存在 `MODULE_TYPELESS_PACKAGE_JSON` warning
- 该 warning 不影响当前收口结论，也不是这轮 architecture closure 的目标范围

### 4.4 前端构建验证

工作目录：

```text
D:\codex_work\go-stock\frontend
```

命令：

```powershell
npm run build
```

结果：`PASS`

说明：

- Vite 仍会提示大 chunk warning
- 该 warning 不构成本轮闭环失败条件

## 5. 显式 allowlist 与非目标

以下内容仍然保留在 bridge 或兼容层中，但这是 spec 明确允许的保留项，不属于未完成缺口。

### 5.1 生命周期与运行时 helper

- `startup`
- `domReady`
- `shutdown`
- `OpenURL`
- `SaveImage`
- `SaveWordFile`
- `CheckUpdate`
- `GetVersionInfo`
- `GetSponsorInfo`
- `GetEffectiveSponsorVip`
- `CheckSponsorCode`
- `GetTimezone`
- `FetchAiModels`

### 5.2 assistant compatibility bridge

- `ChatWithAgent`
- `AbortChatWithAgent`
- `GetAiAssistantSession`
- `SaveAiAssistantSession`

保留理由：

- 总体 spec 明确排除了对多轮 assistant / agent 子系统的深改
- 本轮目标是显式围栏，而不是借收口扩大改造面

### 5.3 非 guard 黑名单项

`app_common.go` 中仍然存在 `data.NewsAnalyze(...)`。

这不是 `data.New*` 构造器，也不在 final closure guard 的禁止 token 列表中，因此不构成这轮闭环的破坏。

## 6. 后续边界

从整体 spec 的闭环角度，当前已经没有需要继续回到 `final-closure` 计划开头重新施工的内容。

如果后续继续推进，只应作为新的独立工作处理，而不是再挂在这轮闭环名下。例如：

- assistant / agent 深层重构
- UI 视觉重设计
- 大块模型或 provider 体系重排
- 非闭环必需的抽象重写

## 7. 结论复述

`factor` 当前状态可以作为 `2026-04-06-go-stock-overall-architecture-refactor-design.md` 的闭环结果使用：

- 阶段 0 到阶段 5 的累计迁移已经闭环
- Phase 5 的旧路径收缩已经完成
- 自动 guard 与 fresh verification 已具备
- 显式 allowlist 与非目标边界清晰

因此，后续不应再把这轮工作描述为“architecture refactor 仍未闭环”，而应描述为“整体架构重构闭环已完成，后续若继续演进属于新任务”。
