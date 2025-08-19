#!/bin/bash
# scripts/run_analysis.sh - 分析执行脚本

#!/bin/bash
set -e

echo "📊 开始内容分析..."

# 检查配置文件
if [ ! -f "config.yaml" ]; then
    echo "❌ 找不到配置文件 config.yaml"
    echo "请先运行: ./scripts/setup.sh"
    exit 1
fi

# 检查内容目录
if [ ! -d "content" ] || [ -z "$(ls -A content 2>/dev/null)" ]; then
    echo "❌ content 目录为空或不存在"
    echo "请在 content/ 目录下放置要分析的文件"
    exit 1
fi

echo "✅ 配置检查通过"

# 构建项目
echo "🔨 构建项目..."
if [ ! -f "bin/content-analyzer" ] || [ "cmd/main.go" -nt "bin/content-analyzer" ]; then
    make build
    echo "✅ 构建完成"
else
    echo "✅ 使用已有构建文件"
fi

# 显示分析概况
content_count=$(find content -name "*.json" -o -name "*.md" | wc -l)
echo "📝 发现 $content_count 个内容文件"

# 执行分析
echo "🚀 执行分析..."
./bin/content-analyzer

# 检查结果
if [ $? -eq 0 ]; then
    echo ""
    echo "✅ 分析完成！"
    echo ""
    echo "📄 生成的报告文件:"
    ls -la output/
    echo ""
    echo "🌐 打开HTML报告: open output/analysis_report.html"
    echo "📊 查看CSV数据: open output/analysis_report.csv"
    echo "🔍 详细JSON数据: output/analysis_report.json"
else
    echo "❌ 分析失败"
    exit 1
fi
