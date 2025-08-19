package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/content-analyzer/internal/analyzer"
	"github.com/content-analyzer/internal/config"
	"github.com/content-analyzer/internal/models"
	"github.com/content-analyzer/internal/report"
)

func main() {
	// 初始化配置
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatal("加载配置失败:", err)
	}

	// 创建分析器
	contentAnalyzer := analyzer.NewContentAnalyzer(cfg)

	// 扫描内容目录
	fmt.Println("开始扫描内容目录...")
	contents, err := scanContentDirectory(cfg.ContentDir)
	if err != nil {
		log.Fatal("扫描目录失败:", err)
	}

	fmt.Printf("发现 %d 个内容文件\n", len(contents))

	// 分析内容
	var results []models.AnalysisResult
	for i, content := range contents {
		fmt.Printf("分析进度: %d/%d - %s\n", i+1, len(contents), content.Title)

		result, err := contentAnalyzer.Analyze(content)
		if err != nil {
			log.Printf("分析失败 %s: %v", content.Title, err)
			continue
		}

		results = append(results, result)

		// 避免API调用过快
		time.Sleep(time.Second * 2)
	}

	// 生成报告
	fmt.Println("\n生成分析报告...")
	reporter := report.NewReporter(cfg)

	if err := reporter.GenerateReport(results); err != nil {
		log.Fatal("生成报告失败:", err)
	}

	fmt.Printf("分析完成！报告已保存到: %s\n", cfg.OutputDir)
}

// scanContentDirectory 扫描内容目录
func scanContentDirectory(dir string) ([]models.Content, error) {
	var contents []models.Content

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 解析内容文件
		content, err := parseContentFile(path)
		if err != nil {
			log.Printf("解析文件失败 %s: %v", path, err)
			return nil // 继续处理其他文件
		}

		if content != nil {
			contents = append(contents, *content)
		}

		return nil
	})

	return contents, err
}

// parseContentFile 解析内容文件
func parseContentFile(filePath string) (*models.Content, error) {
	ext := filepath.Ext(filePath)

	switch ext {
	case ".json":
		return parseJSONContent(filePath)
	case ".md":
		return parseMarkdownContent(filePath)
	default:
		// 跳过不支持的文件类型
		return nil, nil
	}
}

// parseJSONContent 解析JSON格式的内容文件
func parseJSONContent(filePath string) (*models.Content, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var content models.Content
	if err := json.Unmarshal(data, &content); err != nil {
		return nil, err
	}

	content.FilePath = filePath
	return &content, nil
}

// parseMarkdownContent 解析Markdown格式的内容文件
func parseMarkdownContent(filePath string) (*models.Content, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// 简单的Markdown解析（可以后续优化）
	content := models.Content{
		FilePath: filePath,
		Title:    filepath.Base(filePath),
		Text:     string(data),
		Type:     "markdown",
	}

	return &content, nil
}
