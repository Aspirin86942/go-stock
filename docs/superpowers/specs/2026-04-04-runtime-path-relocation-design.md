# 运行时路径收口设计

## 1. 背景

当前 go-stock 的数据库、日志和部分运行时目录使用相对路径，默认落在 `exe` 当前目录：

- SQLite 默认使用 `data/stock.db`
- Wails 日志写入 `./logs/wails.log`
- 业务日志写入 `./logs/info.log` 与 `./logs/error.log`
- 情感分析用户词典读取 `data/dict/user.txt`
- `ai-assistant-web` 启动时也会创建 `data` 和 `logs`

这对绿色版勉强可用，但对安装包场景不合适。若安装目录位于 `Program Files`，继续写相对路径会带来权限风险，也会让发布目录产生额外文件。

## 2. 目标

- 发布目录默认只保留主程序 `exe`
- 运行时数据库、日志、WebView 用户数据统一写入用户目录
- Windows 下统一写入 `%LOCALAPPDATA%\\go-stock`
- 保持主程序与 `ai-assistant-web` 使用同一套路径规则
- 尽量减少对现有功能和数据模型的影响

## 3. 非目标

- 不做旧版同目录数据库的自动迁移
- 不改动配置导出行为，用户手动导出的 `config.json` 仍按用户选择路径保存
- 不顺带重构与路径无关的业务逻辑

## 4. 方案对比

### 方案 A：继续使用相对路径

优点：

- 改动最小

缺点：

- 安装包场景容易出现权限问题
- 发布目录会持续出现数据库、日志和 WAL 文件

### 方案 B：统一收口到应用数据根目录（推荐）

新增统一路径入口，Windows 下将运行时文件统一写入 `%LOCALAPPDATA%\\go-stock`，其他平台保留可兼容的用户目录回退逻辑。

优点：

- 安装包和绿色版都能共用同一套路径策略
- 发布目录可以保持干净
- 后续若增加缓存、导入临时文件、WebView 数据，也有统一归宿

缺点：

- 需要修改多个模块的默认路径来源

### 方案 C：数据库进内存、日志禁用

优点：

- 目录最干净

缺点：

- 数据无法持久化
- 丢失审计与排障能力
- 不符合桌面应用实际需求

### 结论

采用方案 B。

## 5. 总体设计

新增一个独立的运行时路径包，负责：

- 解析应用根目录
- 创建 `data`、`logs`、`webview` 等子目录
- 返回数据库、日志、用户词典、WebView 数据路径

路径规则：

- Windows：`%LOCALAPPDATA%\\go-stock`
- 非 Windows：使用系统用户配置目录或等价回退目录，保持跨平台兼容

## 6. 路径映射

- 数据库：`<app-root>\\data\\stock.db`
- 用户词典：`<app-root>\\data\\dict\\user.txt`
- Wails 日志：`<app-root>\\logs\\wails.log`
- 业务信息日志：`<app-root>\\logs\\info.log`
- 业务错误日志：`<app-root>\\logs\\error.log`
- WebView 用户数据：`<app-root>\\webview`

## 7. 模块改造范围

### 7.1 新增统一路径入口

新增独立包，封装：

- 应用根目录解析
- 子目录创建
- 具体文件路径拼装

设计要求：

- 不依赖 `logger`，避免初始化循环
- 默认返回绝对路径
- 目录创建失败时返回明确错误

### 7.2 数据库

`backend/db/db.go` 不再写死 `data/stock.db`，而是调用统一路径入口生成默认 SQLite DSN。

### 7.3 日志

`backend/logger/lgo.go` 与 `main.go` 中的 Wails logger 统一使用新路径入口，避免把日志写到 `exe` 同目录。

### 7.4 词典

`backend/data/stock_sentiment_analysis.go` 读取用户词典时改用新路径，避免仍查找旧相对路径。

### 7.5 启动初始化

`main.go` 与 `ai-assistant-web/server.go` 不再手工创建相对 `data`/`logs` 目录，而是在启动早期调用统一目录初始化逻辑。

### 7.6 WebView

`main.go` 中 `WebviewUserDataPath` 显式指向新路径下的 `webview` 目录，避免 WebView2 运行时数据散落到默认位置或当前目录。

## 8. 兼容性

- 本次不迁移旧目录中的 `data/stock.db`
- 新版本首次启动时若用户此前只使用同目录数据库，会被视为新环境
- 由于应用内置了基础股票搜索数据，空库首次启动仍可初始化基本搜索能力

## 9. 风险与应对

### 风险 1：日志在 `logger` 包初始化阶段提前落盘到旧目录

应对：

- 让 `logger` 包直接依赖统一路径入口，并在内部主动确保日志目录存在

### 风险 2：主程序与 `ai-assistant-web` 使用不同目录

应对：

- 两侧都通过同一套路径包取值，不再各自拼接目录

### 风险 3：仍有遗漏的相对路径

应对：

- 改动前后都全文检索 `./logs`、`data/stock.db`、`data/dict/user.txt`
- 通过构建与针对性测试确认默认路径全部切换

## 10. 测试与验收

至少验证以下内容：

- 路径解析单元测试覆盖 Windows 场景
- 默认数据库 DSN 指向应用数据目录
- 默认日志文件路径指向应用数据目录
- 主程序构建成功
- 运行后 `exe` 同目录不再新增 `data`、`logs` 等运行时目录
