.PHONY: build run test clean install init analyze help

# 默认目标
default: help

# 构建项目
build:
	@echo "🔨 构建项目..."
	@mkdir -p bin
	go build -ldflags "-s -w" -o bin/content-analyzer cmd/main.go
	@echo "✅ 构建完成: bin/content-analyzer"

# 运行项目
run:
	@echo "🚀 运行内容分析..."
	go run cmd/main.go

# 运行测试
test:
	@echo "🧪 运行测试..."
	go test ./...

# 清理构建文件
clean:
	@echo "🧹 清理构建文件..."
	rm -rf bin/
	rm -rf output/
	@echo "✅ 清理完成"

# 安装依赖
install:
	@echo "📦 安装依赖..."
	go mod tidy
	go mod download
	@echo "✅ 依赖安装完成"

# 初始化项目结构
init:
	@echo "📁 初始化项目结构..."
	mkdir -p content/{images,examples}
	mkdir -p output
	mkdir -p internal/{analyzer,config,models,report,services}
	mkdir -p cmd scripts
	@echo "✅ 项目结构初始化完成"

# 分析指定目录
analyze: build
	@echo "📊 开始内容分析..."
	@if [ ! -d "content" ] || [ -z "$$(ls -A content 2>/dev/null)" ]; then \
		echo "❌ content 目录为空或不存在"; \
		echo "请在 content/ 目录下放置要分析的文件"; \
		exit 1; \
	fi
	./bin/content-analyzer
	@echo "✅ 分析完成，结果保存在 output/ 目录"

# 创建示例内容
examples:
	@echo "📝 创建示例内容文件..."
	@mkdir -p content/examples
	@cat > content/examples/example_post.json << 'EOF'
{
  "id": "example1",
  "title": "5个提升工作效率的小技巧",
  "text": "大家好！今天分享几个我在工作中总结的效率技巧。\n\n1. 使用番茄工作法\n25分钟专注工作，5分钟休息，这样能保持高效状态。\n\n2. 整理桌面和电脑文件\n一个整洁的工作环境能提升专注力。\n\n3. 制定每日计划\n列出当天要完成的任务，按优先级排序。\n\n4. 学会说不\n不要什么都答应，专注于重要的事情。\n\n5. 定期复盘总结\n每周花点时间总结得失，持续改进。\n\n你们有什么提升效率的小妙招吗？评论区分享一下吧！",
  "tags": ["效率", "工作", "技巧", "生活"],
  "published_at": "2024-01-20T09:00:00Z",
  "author": "效率达人",
  "type": "tips",
  "engagement": {
    "likes": 328,
    "comments": 45,
    "shares": 23,
    "views": 1200
  }
}
EOF
	@echo "✅ 示例文件创建完成: content/examples/example_post.json"

# 验证配置文件
validate:
	@echo "🔍 验证配置文件..."
	@if [ ! -f "config.yaml" ]; then \
		echo "❌ 找不到 config.yaml 文件"; \
		echo "请运行: make setup"; \
		exit 1; \
	fi
	@echo "✅ 配置文件存在"
	@if [ -z "$$AI_API_KEY" ]; then \
		echo "⚠️  警告: 未设置 AI_API_KEY 环境变量"; \
		echo "AI 功能将使用简化版本"; \
	else \
		echo "✅ AI_API_KEY 已设置"; \
	fi

# 项目设置
setup: init install
	@echo "⚙️  设置项目配置..."
	@if [ ! -f "config.yaml" ]; then \
		cp config.yaml.example config.yaml 2>/dev/null || echo "请手动创建 config.yaml 文件"; \
	fi
	@if [ ! -f ".env" ]; then \
		cp .env.example .env 2>/dev/null || echo "请手动创建 .env 文件"; \
	fi
	@chmod +x scripts/*.sh 2>/dev/null || true
	@echo "✅ 项目设置完成"

# 生成配置文件模板
config-template:
	@echo "📄 生成配置文件模板..."
	@cat > config.yaml << 'EOF'
# 内容分析配置文件
content_dir: "./content"      # 内容文件目录
output_dir: "./output"        # 分析结果输出目录

# AI服务配置
ai:
  provider: "openai"          # 可选: openai, claude, local
  api_key: ""                 # API密钥，建议通过环境变量 AI_API_KEY 设置
  base_url: ""                # 自定义API地址（可选）
  model: "gpt-3.5-turbo"      # 使用的模型

# 图片分析配置
image:
  max_size: 10485760          # 最大文件大小 10MB
  supported_ext:              # 支持的图片格式
    - ".jpg"
    - ".jpeg"
    - ".png"
    - ".gif"
    - ".bmp"
    - ".webp"
  enable_ocr: false           # 是否启用OCR文字识别

# 分析配置
analysis:
  min_word_count: 50          # 最小字数要求
  max_word_count: 1000        # 推荐最大字数
  score_weights:              # 评分权重
    content_quality: 0.25     # 内容质量权重
    engagement: 0.20          # 互动性权重
    visual: 0.15              # 视觉效果权重
    title: 0.15               # 标题质量权重
    readability: 0.15         # 可读性权重
    trend_relevance: 0.10     # 趋势相关性权重
EOF
	@echo "✅ 配置文件模板生成完成: config.yaml"

# 生成环境变量模板
env-template:
	@echo "📄 生成环境变量模板..."
	@cat > .env << 'EOF'
# 环境变量配置文件
# 复制此文件并填入真实值

# AI API配置
AI_API_KEY=your_api_key_here
AI_BASE_URL=https://api.openai.com/v1

# 其他配置
CONTENT_DIR=./content
OUTPUT_DIR=./output
EOF
	@echo "✅ 环境变量模板生成完成: .env"

# 快速开始（一键设置）
quickstart: clean setup config-template env-template examples
	@echo ""
	@echo "🎉 项目初始化完成！"
	@echo ""
	@echo "📋 下一步操作："
	@echo "1. 编辑 config.yaml 配置文件"
	@echo "2. 编辑 .env 文件，设置 API 密钥"
	@echo "3. 在 content/ 目录放置要分析的文件"
	@echo "4. 运行: make analyze"
	@echo ""
	@echo "💡 快速测试："
	@echo "make analyze  # 使用示例文件进行测试"

# 查看项目状态
status:
	@echo "📊 项目状态检查"
	@echo "=================="
	@echo "Go版本: $$(go version 2>/dev/null || echo '未安装')"
	@echo "项目根目录: $$(pwd)"
	@echo "配置文件: $$([ -f config.yaml ] && echo '✅ 存在' || echo '❌ 不存在')"
	@echo "环境变量文件: $$([ -f .env ] && echo '✅ 存在' || echo '❌ 不存在')"
	@echo "构建文件: $$([ -f bin/content-analyzer ] && echo '✅ 存在' || echo '❌ 不存在')"
	@echo "内容目录: $$([ -d content ] && echo '✅ 存在' || echo '❌ 不存在')"
	@echo "输出目录: $$([ -d output ] && echo '✅ 存在' || echo '❌ 不存在')"
	@echo ""
	@if [ -d content ]; then \
		echo "内容文件数量: $$(find content -name '*.json' -o -name '*.md' | wc -l)"; \
	fi
	@if [ -d output ]; then \
		echo "分析报告: $$([ -f output/analysis_report.html ] && echo '✅ 存在' || echo '❌ 不存在')"; \
	fi

# 打开报告
report:
	@if [ -f "output/analysis_report.html" ]; then \
		echo "📊 打开分析报告..."; \
		open output/analysis_report.html 2>/dev/null || \
		xdg-open output/analysis_report.html 2>/dev/null || \
		echo "请手动打开 output/analysis_report.html"; \
	else \
		echo "❌ 报告文件不存在，请先运行 make analyze"; \
	fi

# 部署到服务器（示例）
deploy:
	@echo "🚀 部署到服务器..."
	@echo "此功能需要根据具体部署环境进行配置"

# 帮助信息
help:
	@echo "📚 内容分析框架 - 可用命令"
	@echo "================================"
	@echo ""
	@echo "🚀 快速开始："
	@echo "  quickstart    一键初始化项目"
	@echo "  analyze       分析内容文件"
	@echo "  report        打开分析报告"
	@echo ""
	@echo "🔧 项目管理："
	@echo "  build         构建项目"
	@echo "  run           运行项目"
	@echo "  test          运行测试"
	@echo "  clean         清理构建文件"
	@echo "  install       安装依赖"
	@echo ""
	@echo "⚙️  配置管理："
	@echo "  setup         设置项目"
	@echo "  config-template  生成配置模板"
	@echo "  env-template     生成环境变量模板"
	@echo "  validate      验证配置"
	@echo ""
	@echo "📁 内容管理："
	@echo "  examples      创建示例内容"
	@echo "  status        查看项目状态"
	@echo ""
	@echo "💡 使用示例："
	@echo "  make quickstart  # 初始化项目"
	@echo "  make analyze     # 分析内容"
	@echo "  make report      # 查看报告"
