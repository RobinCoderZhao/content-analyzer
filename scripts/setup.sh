#!/bin/bash
# scripts/setup.sh - 项目初始化脚本

echo "🚀 初始化内容分析框架..."

# 创建目录结构
echo "📁 创建目录结构..."
mkdir -p content/{images,examples}
mkdir -p output
mkdir -p internal/{analyzer,config,models,report,services}
mkdir -p cmd
mkdir -p scripts

# 检查Go环境
echo "🔍 检查Go环境..."
if ! command -v go &> /dev/null; then
    echo "❌ Go未安装，请先安装Go 1.21+"
    exit 1
fi

echo "✅ Go版本: $(go version)"

# 初始化Go模块
echo "📦 初始化Go模块..."
if [ ! -f "go.mod" ]; then
    go mod init github.com/content-analyzer
    echo "✅ Go模块初始化完成"
else
    echo "✅ Go模块已存在"
fi

# 安装依赖
echo "📥 安装依赖包..."
go mod tidy

# 创建配置文件
echo "⚙️ 创建配置文件..."
if [ ! -f "config.yaml" ]; then
    cp config.yaml.example config.yaml 2>/dev/null || echo "请手动创建config.yaml文件"
fi

if [ ! -f ".env" ]; then
    cp .env.example .env 2>/dev/null || echo "请手动创建.env文件"
fi

# 设置执行权限
chmod +x scripts/*.sh

echo "✅ 项目初始化完成！"
echo ""
echo "📋 下一步:"
echo "1. 编辑 config.yaml 和 .env 文件"
echo "2. 在 content/ 目录放置要分析的文件"
echo "3. 运行: make build && make run"
echo ""
echo "🎯 快速开始:"
echo "make install  # 安装依赖"
echo "make build    # 构建项目"
echo "make run      # 运行分析"

