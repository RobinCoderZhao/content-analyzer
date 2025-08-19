# content-analyzer
# 内容分析框架 (Content Analyzer)

🚀 基于 Go 语言开发的智能内容分析工具，帮助分析和提升小红书等平台的内容质量。

## ✨ 功能特点

- 📝 **文本分析**: 词数统计、可读性分析、情感分析、关键词提取
- 🎯 **标题分析**: 吸引力评分、点击率预测、优化建议
- 🖼️ **图片分析**: 视觉质量评估、构图分析、风格识别
- 📊 **综合评分**: 多维度评分体系，量化内容质量
- 💡 **改进建议**: AI 智能生成具体的优化建议
- 📈 **趋势分析**: 关键词热度、内容趋势识别
- 📋 **多格式报告**: JSON、HTML、CSV 格式输出

## 🏗️ 项目结构

```
content-analyzer/
├── cmd/
│   └── main.go                 # 主程序入口
├── internal/
│   ├── analyzer/
│   │   └── analyzer.go         # 核心分析逻辑
│   ├── config/
│   │   └── config.go          # 配置管理
│   ├── models/
│   │   └── models.go          # 数据模型
│   ├── report/
│   │   └── report.go          # 报告生成
│   └── services/
│       ├── ai_service.go      # AI服务接口
│       └── image_service.go   # 图片分析服务
├── content/                   # 内容文件目录
│   ├── examples/             # 示例文件
│   └── images/               # 图片文件
├── output/                   # 分析结果输出
├── scripts/                  # 脚本文件
├── config.yaml              # 配置文件
├── go.mod                   # Go模块定义
├── Makefile                 # 构建脚本
└── README.md               # 项目说明
```

## 🚀 快速开始

### 1. 环境准备

确保已安装 Go 1.21+ 版本：

```bash
go version  # 检查Go版本
```

### 2. 克隆项目

```bash
git clone <repository-url>
cd content-analyzer
```

### 3. 一键初始化

```bash
make quickstart
```

此命令会：
- 创建项目目录结构
- 安装依赖包
- 生成配置文件模板
- 创建示例内容文件

### 4. 配置设置

编辑 `config.yaml` 文件：

```yaml
content_dir: "./content"
output_dir: "./output"

ai:
  provider: "openai"
  api_key: ""  # 设置你的API密钥
  model: "gpt-3.5-turbo"
```

设置环境变量（推荐方式）：

```bash
# 编辑 .env 文件
AI_API_KEY=your_actual_api_key_here
```

### 5. 准备内容文件

在 `content/` 目录下放置要分析的文件：

**JSON 格式示例：**

```json
{
  "id": "post1",
  "title": "5个超实用的护肤小技巧，让你的皮肤水嫩如初！",
  "text": "大家好！今天想和大家分享一些我亲身试验过的护肤心得...",
  "images": [
    {
      "path": "images/skincare1.jpg",
      "caption": "护肤产品展示"
    }
  ],
  "tags": ["护肤", "美妆", "生活"],
  "published_at": "2024-01-15T10:00:00Z",
  "engagement": {
    "likes": 1250,
    "comments": 89,
    "shares": 45
  }
}
```

**Markdown 格式示例：**

```markdown
# 我的减肥日记：30天瘦了8斤的真实经历

大家好！我是一个普通的上班族...

## 第一周：适应期

刚开始真的很难！以前习惯了大吃大喝...

#减肥 #健康生活 #减肥日记
```

### 6. 运行分析

```bash
make analyze
```

### 7. 查看结果

```bash
make report  # 自动打开HTML报告
```

或手动查看：
- `output/analysis_report.html` - 可视化HTML报告
- `output/analysis_report.json` - 详细JSON数据
- `output/analysis_report.csv` - 电子表格格式

## 📊 分析维度

### 文本分析
- **基础指标**: 字数、句子数、段落数
- **结构分析**: 是否有引言、结论、列表等
- **写作风格**: 语调、人称、正式程度
- **互动元素**: CTA、提问、话题标签

### 标题分析
- **长度适中**: 10-30字符最佳
- **吸引元素**: 数字、疑问句、情感词汇
- **清晰度**: 主题明确程度
- **点击率预测**: 基于历史数据的预测

### 图片分析
- **质量指标**: 分辨率、清晰度、噪点
- **构图分析**: 三分法则、对称性、平衡感
- **视觉元素**: 色彩、亮度、对比度
- **风格识别**: 现代、复古、简约等

### 综合评分
- **内容质量** (25%): 原创性、信息价值、结构完整性
- **互动潜力** (20%): 引导互动、情感共鸣、话题性
- **视觉吸引** (15%): 图片质量、视觉冲击力
- **标题质量** (15%): 吸引力、清晰度、优化程度
- **可读性** (15%): 语言难度、句子长度、逻辑性
- **趋势相关** (10%): 热门话题、关键词热度

## 🛠️ 高级功能

### AI 服务配置

支持多种 AI 服务提供商：

```yaml
ai:
  provider: "openai"  # 可选: openai, claude
  api_key: "your-key"
  base_url: "https://api.openai.com/v1"  # 自定义API地址
  model: "gpt-3.5-turbo"
```

### 批量分析

```bash
# 分析指定目录下的所有文件
make analyze

# 检查项目状态
make status

# 验证配置
make validate
```

### 自定义评分权重

修改 `config.yaml` 中的权重配置：

```yaml
analysis:
  score_weights:
    content_quality: 0.30    # 提高内容质量权重
    engagement: 0.25         # 提高互动性权重
    visual: 0.10            # 降低视觉权重
    title: 0.15
    readability: 0.15
    trend_relevance: 0.05
```

## 📋 可用命令

### 快速命令
```bash
make quickstart    # 一键初始化项目
make analyze       # 分析内容文件
make report        # 打开分析报告
make status        # 查看项目状态
```

### 项目管理
```bash
make build         # 构建项目
make run           # 运行项目
make test          # 运行测试
make clean         # 清理构建文件
make install       # 安装依赖
```

### 配置管理
```bash
make setup         # 设置项目
make validate      # 验证配置
make examples      # 创建示例内容
```

### 获取帮助
```bash
make help          # 显示所有可用命令
```

## 🔧 开发指南

### 添加新的分析维度

1. 在 `internal/models/models.go` 中定义新的数据结构
2. 在 `internal/analyzer/analyzer.go` 中实现分析逻辑
3. 更新 `internal/report/report.go` 中的报告生成逻辑

### 扩展 AI 服务

1. 在 `internal/services/ai_service.go` 中添加新的提供商
2. 实现对应的 API 调用方法
3. 在配置文件中添加相关配置项

### 添加新的报告格式

1. 在 `internal/report/report.go` 中添加新的生成方法
2. 实现对应的格式转换逻辑
3. 更新主程序调用

## 📈 使用场景

### 内容创作者
- 分析已发布内容的表现
- 获取优化建议提升质量
- 跟踪内容趋势和热点

### 营销团队
- 批量分析竞品内容
- 制定内容策略
- 优化投放效果

### 个人博主
- 提升写作水平
- 增加粉丝互动
- 优化内容规划

## 🤝 贡献指南

欢迎贡献代码！请遵循以下步骤：

1. Fork 本项目
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交改动 (`git commit -m 'Add amazing feature'`)
4. 推送分支 (`git push origin feature/amazing-feature`)
5. 创建 Pull Request

## 📝 更新日志

### v1.0.0 (2024-01-20)
- ✨ 初始版本发布
- 📝 支持文本和图片分析
- 📊 多格式报告生成
- 🤖 AI 智能建议生成

## ❓ 常见问题

### Q: 如何设置 API 密钥？
A: 推荐在 `.env` 文件中设置 `AI_API_KEY=your_key`，或直接在 `config.yaml` 中配置。

### Q: 支持哪些文件格式？
A: 目前支持 JSON 和 Markdown 格式的内容文件，以及 JPG、PNG、GIF 等图片格式。

### Q: 可以不使用 AI 服务吗？
A: 可以！如果不设置 API 密钥，系统会使用简化版本的分析算法。

### Q: 如何批量分析大量文件？
A: 将所有文件放在 `content/` 目录下，运行 `make analyze` 即可批量处理。

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

## 🌟 致谢

感谢所有贡献者和开源社区的支持！

---

如果这个项目对你有帮助，请给个 ⭐️ Star！
