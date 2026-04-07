# go-stock : 基于大语言模型的AI赋能股票分析工具
## ![go-stock](./build/appicon.png)
![GitHub Release](https://img.shields.io/github/v/release/ArvinLovegood/go-stock?link=https%3A%2F%2Fgithub.com%2FArvinLovegood%2Fgo-stock%2Freleases&link=https%3A%2F%2Fgithub.com%2FArvinLovegood%2Fgo-stock%2Freleases)

###  ✨ 简介
- 本项目基于Wails和NaiveUI开发，结合AI大模型构建的股票分析工具。
- 目前已支持A股、港股、美股，并覆盖市场资讯、公告、研报、K线、AI分析等常用研究场景。
- 本项目以学习研究为主，AI 分析结果仅供参考，投资有风险，请谨慎使用。
- 开发环境主要基于Windows10+，其他平台未测试或功能受限。

### 🎯 项目定位
- 这是一个免费开源的桌面研究工具，没有收费解锁，也不是商业化产品。
- 这个仓库首先是我的自用项目，用来把股票观察、市场信息和 AI 分析放到一个本地桌面工作台里。
- 它同时也是一个持续进行中的 `vibe coding` 实验场，用来验证“需求、架构、文档、实现、测试、收口”这一整套个人开发流能不能长期跑通。
- 如果你把它当成一个个人研究型工具来理解，这个 README 会更准确；如果你期待的是成熟商业软件，它并不是按那个目标设计的。

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

### ✅ 当前能力
- 股票池覆盖 A 股、港股、美股。
- 支持市场资讯、财经日历、公告、研报、分时/K线等只读研究链路。
- 支持 AI 分析、AI 助手、本地或远程大模型接入。
- 支持定时任务、提醒、部分基金观察能力。


### 💬 支持大模型/平台
| 模型 | 状态 | 备注                                                                                                                                                                                                                                                                |
| --- | --- |-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| [OpenAI](https://platform.openai.com/) | ✅ | 可接入任何 OpenAI 接口格式模型                                                                                                                                                                                                                                               |
| [Ollama](https://ollama.com/) | ✅ | 本地大模型运行平台                                                                                                                                                                                                                                                         |
| [LMStudio](https://lmstudio.ai/) | ✅ | 本地大模型运行平台                                                                                                                                                                                                                                                         |
| [AnythingLLM](https://anythingllm.com/) | ✅ | 本地知识库                                                                                                                                                                                                                                                             |
| [DeepSeek](https://www.deepseek.com/) | ✅ | deepseek-reasoner,deepseek-chat                                                                                                                                                                                                                                   |
| 大模型聚合平台 | ✅ | 如：[硅基流动](https://cloud.siliconflow.cn/)、[火山方舟](https://www.volcengine.com/product/ark) 等兼容 OpenAI API 的平台 |

- 软件仍在持续迭代，使用时建议优先选择最新发布版本。
- 如有问题或建议，可以直接提 issue。
- 完成基础配置后即可直接使用。

## 💕 感谢以下项目
- [NaiveUI](https://www.naiveui.com/)
- [Wails](https://wails.io/)
- [Vue](https://vuejs.org/)
- [Vite](https://vitejs.dev/)
- [Tushare](https://tushare.pro/)

## License
[GNU GPLv3](LICENSE)

