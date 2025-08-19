// internal/services/services.go
package services

import (
	"context"
	"fmt"
	"log"

	"github.com/content-analyzer/internal/config"
	"github.com/content-analyzer/internal/models"
)

// ServiceManager 服务管理器
type ServiceManager struct {
	AIService    AIService
	ImageService ImageService
	config       *config.Config
}

// NewServiceManager 创建服务管理器
func NewServiceManager(cfg *config.Config) *ServiceManager {
	return &ServiceManager{
		AIService:    NewAIService(cfg),
		ImageService: NewImageService(cfg),
		config:       cfg,
	}
}

// HealthCheck 健康检查
func (sm *ServiceManager) HealthCheck(ctx context.Context) error {
	log.Println("开始服务健康检查...")

	// 检查AI服务
	if err := sm.checkAIService(ctx); err != nil {
		log.Printf("AI服务检查失败: %v", err)
	} else {
		log.Println("✅ AI服务正常")
	}

	// 检查图片服务
	if err := sm.checkImageService(); err != nil {
		log.Printf("图片服务检查失败: %v", err)
		return fmt.Errorf("图片服务不可用: %w", err)
	}
	log.Println("✅ 图片服务正常")

	log.Println("服务健康检查完成")
	return nil
}

func (sm *ServiceManager) checkAIService(ctx context.Context) error {
	// 如果没有配置API密钥，跳过检查
	if sm.config.AI.APIKey == "" {
		log.Println("⚠️  AI API密钥未配置，将使用简化版本")
		return nil
	}

	// 简单的连通性测试
	testText := "这是一个测试文本"
	_, err := sm.AIService.AnalyzeSentiment(ctx, testText)
	if err != nil {
		return fmt.Errorf("AI服务连接失败: %w", err)
	}

	return nil
}

func (sm *ServiceManager) checkImageService() error {
	// 检查图片服务配置
	if len(sm.config.Image.SupportedExt) == 0 {
		return fmt.Errorf("未配置支持的图片格式")
	}

	if sm.config.Image.MaxSize <= 0 {
		return fmt.Errorf("图片大小限制配置错误")
	}

	return nil
}

// GetServiceInfo 获取服务信息
func (sm *ServiceManager) GetServiceInfo() map[string]interface{} {
	info := map[string]interface{}{
		"ai_service": map[string]interface{}{
			"provider":    sm.config.AI.Provider,
			"model":       sm.config.AI.Model,
			"has_api_key": sm.config.AI.APIKey != "",
			"base_url":    sm.config.AI.BaseURL,
		},
		"image_service": map[string]interface{}{
			"max_size":      sm.config.Image.MaxSize,
			"supported_ext": sm.config.Image.SupportedExt,
			"enable_ocr":    sm.config.Image.EnableOCR,
		},
		"analysis_config": map[string]interface{}{
			"min_word_count": sm.config.Analysis.MinWordCount,
			"max_word_count": sm.config.Analysis.MaxWordCount,
			"score_weights":  sm.config.Analysis.ScoreWeights,
		},
	}

	return info
}

// BatchProcessImages 批量处理图片
func (sm *ServiceManager) BatchProcessImages(imagePaths []string) ([]models.ImageAnalysis, error) {
	return sm.ImageService.BatchAnalyze(imagePaths)
}

// GenerateContentAdvice 生成内容建议
func (sm *ServiceManager) GenerateContentAdvice(ctx context.Context, analysis models.AnalysisResult) (string, error) {
	return sm.AIService.GenerateAdvice(ctx, analysis)
}

// ExtractContentTopics 提取内容主题
func (sm *ServiceManager) ExtractContentTopics(ctx context.Context, text string) ([]string, error) {
	return sm.AIService.ExtractTopics(ctx, text)
}

// ValidateContent 验证内容质量
func (sm *ServiceManager) ValidateContent(content models.Content) []string {
	var issues []string

	// 检查内容长度
	wordCount := len(content.Text)
	if wordCount < sm.config.Analysis.MinWordCount {
		issues = append(issues, fmt.Sprintf("内容过短，当前%d字，建议至少%d字",
			wordCount, sm.config.Analysis.MinWordCount))
	}

	if wordCount > sm.config.Analysis.MaxWordCount {
		issues = append(issues, fmt.Sprintf("内容过长，当前%d字，建议不超过%d字",
			wordCount, sm.config.Analysis.MaxWordCount))
	}

	// 检查标题
	if len(content.Title) == 0 {
		issues = append(issues, "缺少标题")
	} else if len(content.Title) > 100 {
		issues = append(issues, "标题过长，建议控制在50字以内")
	}

	// 检查图片
	for i, img := range content.Images {
		if err := sm.ImageService.ValidateImage(img.Path); err != nil {
			issues = append(issues, fmt.Sprintf("图片%d验证失败: %v", i+1, err))
		}
	}

	return issues
}
