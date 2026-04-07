# go-stock : 基于大语言模型的AI赋能股票分析工具
## ![go-stock](./build/appicon.png)
![GitHub Release](https://img.shields.io/github/v/release/ArvinLovegood/go-stock?link=https%3A%2F%2Fgithub.com%2FArvinLovegood%2Fgo-stock%2Freleases&link=https%3A%2F%2Fgithub.com%2FArvinLovegood%2Fgo-stock%2Freleases)

###  ✨ 简介
- 本项目基于Wails和NaiveUI开发，结合AI大模型构建的股票分析工具。
- 目前已支持A股，港股，美股，未来计划加入基金，ETF等支持。
- 支持市场整体/个股情绪分析，K线技术指标分析等功能。
- 本项目以学习研究为主，AI 分析结果仅供参考，投资有风险，请谨慎使用。
- 开发环境主要基于Windows10+，其他平台未测试或功能受限。

### 🧭 架构与维护说明
- 这个项目主要用于自用研究、学习 AI 能力和持续练习架构设计，不以商业化为目标。
- 当前技术栈固定为 `Go + Wails + Vue`，本轮重构的重点不是换栈，而是把自然生长的代码边界收紧。
- 后端现在以 `bridge -> service -> source` 为目标边界，前端现在以 `router -> pages -> components/services` 为目标边界，尽量避免继续把业务逻辑堆回桥接层和重量级组件。
- 这次重构的目标是把项目收敛成一个可长期维护的桌面研究工作台，而不是为了“更优雅”继续增加抽象层。
- 更完整的背景、阶段目标、非目标和收口结果见：
  - [整体架构重构 spec](docs/superpowers/specs/2026-04-06-go-stock-overall-architecture-refactor-design.md)
  - [整体架构重构 closeout](docs/superpowers/checklists/2026-04-07-go-stock-overall-architecture-refactor-closeout.md)

### 📃 使用手册
[go-stock使用手册](docs/go-stock使用手册.md)

### 📦 立即体验
[//]: # (- 安装版：[go-stock-amd64-installer.exe]&#40;https://github.com/ArvinLovegood/go-stock/releases&#41;)
- 绿色版：[go-stock-windows-amd64.exe](https://github.com/ArvinLovegood/go-stock/releases)
- MACOS绿色版：[go-stock-darwin-universal](https://github.com/ArvinLovegood/go-stock/releases)

[//]: # (- MACOS安装版：[go-stock-darwin-universal.pkg]&#40;https://github.com/ArvinLovegood/go-stock/releases&#41;)


### 💬 支持大模型/平台
| 模型 | 状态 | 备注                                                                                                                                                                                                                                                                |
| --- | --- |-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| [OpenAI](https://platform.openai.com/) | ✅ | 可接入任何 OpenAI 接口格式模型                                                                                                                                                                                                                                               |
| [Ollama](https://ollama.com/) | ✅ | 本地大模型运行平台                                                                                                                                                                                                                                                         |
| [LMStudio](https://lmstudio.ai/) | ✅ | 本地大模型运行平台                                                                                                                                                                                                                                                         |
| [AnythingLLM](https://anythingllm.com/) | ✅ | 本地知识库                                                                                                                                                                                                                                                             |
| [DeepSeek](https://www.deepseek.com/) | ✅ | deepseek-reasoner,deepseek-chat                                                                                                                                                                                                                                   |
| 大模型聚合平台 | ✅ | 如：[硅基流动](https://cloud.siliconflow.cn/)、[火山方舟](https://www.volcengine.com/product/ark) 等兼容 OpenAI API 的平台 |

- 软件快速迭代开发中,请大家优先测试和使用最新发布的版本。
- 欢迎大家提出宝贵的建议，欢迎提 issue、PR。
- 完成基础配置后即可直接使用。

## 🧩 重大功能开发计划
| 功能说明            | 状态 | 备注                                                                                                       |
|-----------------|----|----------------------------------------------------------------------------------------------------------|
| 股票分析知识库         | 🚧 | 未来计划                                                                                                     |
| Ai智能选股          | ✅  | Ai智能选股功能(市场行情-》AI总结/AI智能体功能)                                                                             |
| ETF支持           | 🚧 | ETF数据支持 (目前可以查看净值和估值)                                                                                    |
| 美股支持            | ✅  | 美股数据支持                                                                                                   |
| 港股支持            | ✅  | 港股数据支持                                                                                                   |
| 多轮对话            | ✅  | AI分析后可继续对话提问                                                                                             |
| 自定义AI分析提问模板     | ✅  | 可配置的提问模板 [v2025.2.12.7-alpha](https://github.com/ArvinLovegood/go-stock/releases/tag/v2025.2.12.7-alpha) |
| 不再强制依赖Chrome浏览器 | ✅  | 默认使用edge浏览器抓取新闻资讯                                                                                        |

## 👀 更新日志
### 2026.03.10 新增AI助手功能
### 2026.02.08 完善AI推荐股票功能和数据查询功能
### 2026.01.25 添加AI分析报告和ai推荐股票的历史数据查看功能
### 2025.12.16 新增AI思考模式与热门选股策略功能
### 2025.11.21 新增带频率权重的情感分析功能
### 2025.10.30 添加AI智能体功能开关(默认关闭，因为使用体验不理想)，移除页面水印
### 2025.09.27 添加机构/券商的研究报告AI工具函数
### 2025.08.09 添加AI智能体聊天功能
### 2025.07.08 实现软件自动更新功能
### 2025.07.07 卡片添加迷你分时图
### 2025.07.05 MacOs支持
### 2025.07.01 AI分析集成工具函数，AI分析将更加智能
### 2025.06.30 添加指标选股功能
### 2025.06.27 添加财经日历和重大事件时间轴功能
### 2025.06.25 添加热门股票、事件和话题功能
### 2025.06.18 更新内置股票基础数据,软件内实时市场资讯信息提醒，添加行业研究功能
### 2025.06.15 添加公司公告信息搜索/查看功能
### 2025.06.15 添加个股研报到弹出菜单
### 2025.06.13 添加个股研报功能
### 2025.06.12 添加龙虎榜功能，新增行业排名分类
### 2025.05.30 优化股票分时图显示
### 2025.05.20 修复财联社电报获取问题
### 2025.05.16 优化资金趋势图表组件
### 2025.05.15 重构应用加载和数据初始化逻辑，添加股票资金趋势功能，资金趋势图表增加主力当日净流入数据并优化展示效果
### 2025.05.14 添加个股资金流向功能，排行榜增加股票行情K线图弹窗
### 2025.05.13 添加行业排名功能
### 2025.05.09 添加A股盘口数据解析和展示功能
### 2025.05.07 优化分时图的展示
### 2025.04.29 补全港股/美股基础数据，优化港股股价延迟问题，优化初始化逻辑
### 2025.04.25 市场资讯支持AI分析和总结：让AI帮你读市场！
### 2025.04.24 新增市场行情模块：即时掌握全球市场行情资讯/动态，从此再也不用偷摸去各大财经网站啦。go-stock一键帮你搞定！
### 2025.04.22 优化K线图展示，支持拉伸放大，看得更舒服啦！
### 2025.04.21 港股，美股K线数据获取优化
### 2025.04.01 优化部分设置选项，避免重启软件
### 2025.03.31 优化数据爬取
### 2025.03.30 AI自动定时分析功能
### 2025.03.29 多提示词模板管理，AI分析时支持选择不同提示词模板
### 2025.03.28 AI分析结果保存为markdown文件时，支持保存位置目录选择
### 2025.03.15 自定义爬虫使用的浏览器路径配置
### 2025.03.14 优化编译构建，大幅减少编译后的程序文件大小
### 2025.03.09 基金估值和净值监控查看
### 2025.03.06 项目社区分享功能
### 2025.02.28 美股数据支持
### 2025.02.23 弹幕功能，盯盘不再孤单，无聊划个水！😎
### 2025.02.22 港股数据支持(目前有延迟)

### 2025.02.16 AI分析后可继续对话提问
- [v2025.2.16.1-alpha](https://github.com/ArvinLovegood/go-stock/releases/tag/v2025.2.16.1-alpha)

### 2025.02.12 可配置的提问模板
- [v2025.2.12.7-alpha](https://github.com/ArvinLovegood/go-stock/releases/tag/v2025.2.12.7-alpha)


## 🦄 重大更新
### BIG NEWS !!! 重大更新！！！
- 2026.03.10 新增AI助手功能
![img_1.png](build/screenshot/img16.png)
- 2025.11.21 新增带频率权重的情感分析功能
![img_1.png](build/screenshot/img15.png)
- 2025.04.25 市场资讯支持AI分析和总结：让AI帮你读市场！
![img.png](img.png)
- 2025.04.24 新增市场行情模块：即时掌握全球市场行情资讯/动态，从此再也不用偷摸去各大财经网站啦。go-stock一键帮你搞定！
![img.png](build/screenshot/img13.png)
![img_13.png](build/screenshot/img_13.png)
- ![img_14.png](build/screenshot/img_14.png)
- 2025.01.17 新增AI大模型分析股票功能
  ![img_5.png](build/screenshot/img.png)
## 📸 功能截图
![img_1.png](build/screenshot/img_6.png)
### 设置
![img_12.png](build/screenshot/img_4.png)
### 成本设置
![img.png](build/screenshot/img_7.png)
### 日K
![img_12.png](build/screenshot/img_12.png)
### 分时
![img_3.png](build/screenshot/img_9.png)
### 钉钉报警通知
![img_4.png](build/screenshot/img_5.png)
### AI分析股票
![img_5.png](build/screenshot/img.png)
### 版本信息提示
![img_11.png](build/screenshot/img_11.png)

## 💕 感谢以下项目
- [NaiveUI](https://www.naiveui.com/)
- [Wails](https://wails.io/)
- [Vue](https://vuejs.org/)
- [Vite](https://vitejs.dev/)
- [Tushare](https://tushare.pro/)

[//]: # (<img src="./build/wx.jpg" width="301px" height="402px" alt="ArvinLovegood">)

## License
[GNU GPLv3](LICENSE)

