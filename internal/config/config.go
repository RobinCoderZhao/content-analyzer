// internal/config/config.go
package config

import (
	"fmt"
	"os"
	"gopkg.in/yaml.v3"
)

type Config struct {
	ContentDir string    `yaml:"content_dir"`
	OutputDir  string    `yaml:"output_dir"`
	AI         AIConfig  `yaml:"ai"`
	Image      ImageConfig `yaml:"image"`
	Analysis   AnalysisConfig `yaml:"analysis"`
}

type AIConfig struct {
	Provider string `yaml:"provider"` // openai, claude, local
	APIKey   string `yaml:"api_key"`
	BaseURL  string `yaml:"base_url,omitempty"`
	Model    string `yaml:"model"`
}

type ImageConfig struct {
	MaxSize      int64    `yaml:"max_size"`      // 最大文件大小（字节）
	SupportedExt []string `yaml:"supported_ext"` // 支持的扩展名
	EnableOCR    bool     `yaml:"enable_ocr"`    // 是否启用文字识别
}

type AnalysisConfig struct {
	MinWordCount    int     `yaml:"min_word_count"`    // 最小词数要求
	MaxWordCount    int     `yaml:"max_word_count"`    // 最大词数建议
	ScoreWeights    ScoreWeights `yaml:"score_weights"`
}

type ScoreWeights struct {
	ContentQuality  float64 `yaml:"content_quality"`
	Engagement      float64 `yaml:"engagement"`
	Visual          float64 `yaml:"visual"`
	Title           float64 `yaml:"title"`
	Readability     float64 `yaml:"readability"`
	TrendRelevance  float64 `yaml:"trend_relevance"`
}

func Load(configPath string) (*Config, error) {
	// 默认配置
	config := &Config{
		ContentDir: "./content",
		OutputDir:  "./output",
		AI: AIConfig{
			Provider: "openai",
			Model:    "gpt-3.5-turbo",
		},
		Image: ImageConfig{
			MaxSize:      10 * 1024 * 1024, // 10MB
			SupportedExt: []string{".jpg", ".jpeg", ".png", ".gif", ".bmp"},
			EnableOCR:    false,
		},
		Analysis: AnalysisConfig{
			MinWordCount: 50,
			MaxWordCount: 1000,
			ScoreWeights: ScoreWeights{
				ContentQuality: 0.25,
				Engagement:     0.20,
				Visual:         0.15,
				Title:          0.15,
				Readability:    0.15,
				TrendRelevance: 0.10,
			},
		},
	}

	// 如果配置文件存在，则加载
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("读取配置文件失败: %w", err)
		}

		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("解析配置文件失败: %w", err)
		}
	}

	// 从环境变量覆盖敏感配置
	if apiKey := os.Getenv("AI_API_KEY"); apiKey != "" {
		config.AI.APIKey = apiKey
	}

	return config, nil
}

// internal/services/ai_service.go
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/RobinCoderZhao/content-analyzer/internal/config"
	"github.com/RobinCoderZhao/content-analyzer/internal/models"
)

type AIService interface {
	AnalyzeSentiment(ctx context.Context, text string) (models.SentimentAnalysis, error)
	GenerateAdvice(ctx context.Context, analysis models.AnalysisResult) (string, error)
	ExtractTopics(ctx context.Context, text string) ([]string, error)
}

type aiService struct {
	config *config.Config
}

func NewAIService(cfg *config.Config) AIService {
	return &aiService{config: cfg}
}

func (s *aiService) AnalyzeSentiment(ctx context.Context, text string) (models.SentimentAnalysis, error) {
	// 简化的情感分析实现
	// 在实际应用中，这里应该调用真正的AI API
	
	sentiment := models.SentimentAnalysis{
		Overall:    s.determineSentiment(text),
		Score:      s.calculateSentimentScore(text),
		Emotions:   s.analyzeEmotions(text),
		Confidence: 0.8,
	}
	
	return sentiment, nil
}

func (s *aiService) GenerateAdvice(ctx context.Context, analysis models.AnalysisResult) (string, error) {
	// 根据分析结果生成建议
	advice := strings.Builder{}
	
	advice.WriteString("## 内容分析建议\n\n")
	
	if analysis.Score.Total < 60 {
		advice.WriteString("📊 **总体评分偏低**，建议重点关注以下方面：\n")
	} else if analysis.Score.Total < 80 {
		advice.WriteString("📊 **总体表现良好**，还有提升空间：\n")
	} else {
		advice.WriteString("📊 **内容质量优秀**，保持当前水准：\n")
	}
	
	// 根据各项得分给出具体建议
	if analysis.Score.Breakdown.Title < 70 {
		advice.WriteString("- 🎯 **标题优化**：当前标题吸引力不足，建议使用数字、提问或情感词汇\n")
	}
	
	if analysis.Score.Breakdown.Engagement < 70 {
		advice.WriteString("- 💬 **增强互动**：添加问题引导、call-to-action等元素提升参与度\n")
	}
	
	if len(analysis.ImageAnalysis) == 0 {
		advice.WriteString("- 🖼️ **视觉内容**：建议添加相关图片或视觉元素增强吸引力\n")
	}
	
	if analysis.Readability.FleschScore < 60 {
		advice.WriteString("- 📖 **可读性优化**：使用更短的句子和简单词汇提升阅读体验\n")
	}
	
	return advice.String(), nil
}

func (s *aiService) ExtractTopics(ctx context.Context, text string) ([]string, error) {
	// 简化的主题提取
	topics := []string{}
	
	// 基于关键词匹配识别主题
	topicKeywords := map[string][]string{
		"美食": {"吃", "食物", "餐厅", "菜", "味道", "好吃"},
		"旅行": {"旅游", "景点", "酒店", "机票", "攻略", "风景"},
		"科技": {"手机", "电脑", "软件", "APP", "数码", "互联网"},
		"时尚": {"穿搭", "化妆", "护肤", "衣服", "搭配", "美妆"},
		"生活": {"日常", "分享", "经验", "感受", "生活", "日记"},
	}
	
	text = strings.ToLower(text)
	for topic, keywords := range topicKeywords {
		for _, keyword := range keywords {
			if strings.Contains(text, keyword) {
				topics = append(topics, topic)
				break
			}
		}
	}
	
	if len(topics) == 0 {
		topics = append(topics, "其他")
	}
	
	return topics, nil
}

// 辅助方法
func (s *aiService) determineSentiment(text string) string {
	positiveWords := []string{"好", "棒", "优秀", "喜欢", "爱", "开心", "满意", "推荐"}
	negativeWords := []string{"差", "坏", "糟糕", "讨厌", "恨", "失望", "不满", "后悔"}
	
	text = strings.ToLower(text)
	positiveCount := 0
	negativeCount := 0
	
	for _, word := range positiveWords {
		positiveCount += strings.Count(text, word)
	}
	
	for _, word := range negativeWords {
		negativeCount += strings.Count(text, word)
	}
	
	if positiveCount > negativeCount {
		return "positive"
	} else if negativeCount > positiveCount {
		return "negative"
	}
	
	return "neutral"
}

func (s *aiService) calculateSentimentScore(text string) float64 {
	sentiment := s.determineSentiment(text)
	switch sentiment {
	case "positive":
		return 0.7
	case "negative":
		return -0.7
	default:
		return 0.0
	}
}

func (s *aiService) analyzeEmotions(text string) map[string]float64 {
	emotions := map[string]float64{
		"joy":     0.0,
		"sadness": 0.0,
		"anger":   0.0,
		"fear":    0.0,
		"surprise": 0.0,
	}
	
	// 简化的情感检测
	text = strings.ToLower(text)
	
	if strings.Contains(text, "开心") || strings.Contains(text, "高兴") || strings.Contains(text, "快乐") {
		emotions["joy"] = 0.8
	}
	
	if strings.Contains(text, "难过") || strings.Contains(text, "伤心") || strings.Contains(text, "沮丧") {
		emotions["sadness"] = 0.8
	}
	
	if strings.Contains(text, "生气") || strings.Contains(text, "愤怒") || strings.Contains(text, "气愤") {
		emotions["anger"] = 0.8
	}
	
	if strings.Contains(text, "害怕") || strings.Contains(text, "担心") || strings.Contains(text, "恐惧") {
		emotions["fear"] = 0.8
	}
	
	if strings.Contains(text, "惊讶") || strings.Contains(text, "意外") || strings.Contains(text, "震惊") {
		emotions["surprise"] = 0.8
	}
	
	return emotions
}

// internal/services/image_service.go
package services

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	
	"github.com/RobinCoderZhao/content-analyzer/internal/config"
	"github.com/RobinCoderZhao/content-analyzer/internal/models"
)

type ImageService interface {
	AnalyzeImage(imagePath string) (models.ImageAnalysis, error)
	ValidateImage(imagePath string) error
	GetImageInfo(imagePath string) (models.Image, error)
}

type imageService struct {
	config *config.Config
}

func NewImageService(cfg *config.Config) ImageService {
	return &imageService{config: cfg}
}

func (s *imageService) AnalyzeImage(imagePath string) (models.ImageAnalysis, error) {
	// 验证图片
	if err := s.ValidateImage(imagePath); err != nil {
		return models.ImageAnalysis{}, err
	}
	
	// 获取图片基本信息
	imgInfo, err := s.GetImageInfo(imagePath)
	if err != nil {
		return models.ImageAnalysis{}, err
	}
	
	// 分析图片
	analysis := models.ImageAnalysis{
		Path: imagePath,
		VisualElements: s.analyzeVisualElements(imgInfo),
		CompositionAnalysis: s.analyzeComposition(imgInfo),
		QualityMetrics: s.analyzeQuality(imgInfo),
		StyleAnalysis: s.analyzeStyle(imgInfo),
	}
	
	// 计算综合得分
	analysis.Score = s.calculateImageScore(analysis)
	
	return analysis, nil
}

func (s *imageService) ValidateImage(imagePath string) error {
	// 检查文件是否存在
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return fmt.Errorf("图片文件不存在: %s", imagePath)
	}
	
	// 检查文件扩展名
	ext := strings.ToLower(filepath.Ext(imagePath))
	supported := false
	for _, supportedExt := range s.config.Image.SupportedExt {
		if ext == supportedExt {
			supported = true
			break
		}
	}
	
	if !supported {
		return fmt.Errorf("不支持的图片格式: %s", ext)
	}
	
	// 检查文件大小
	fileInfo, err := os.Stat(imagePath)
	if err != nil {
		return err
	}
	
	if fileInfo.Size() > s.config.Image.MaxSize {
		return fmt.Errorf("图片文件过大: %d bytes (最大: %d bytes)", 
			fileInfo.Size(), s.config.Image.MaxSize)
	}
	
	return nil
}

func (s *imageService) GetImageInfo(imagePath string) (models.Image, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return models.Image{}, err
	}
	defer file.Close()
	
	// 获取图片配置信息
	config, format, err := image.DecodeConfig(file)
	if err != nil {
		return models.Image{}, err
	}
	
	// 获取文件信息
	fileInfo, err := os.Stat(imagePath)
	if err != nil {
		return models.Image{}, err
	}
	
	return models.Image{
		Path:   imagePath,
		Width:  config.Width,
		Height: config.Height,
		Size:   fileInfo.Size(),
		Format: format,
	}, nil
}

func (s *imageService) analyzeVisualElements(imgInfo models.Image) models.VisualElements {
	// 简化的视觉元素分析
	// 在实际应用中，这里需要使用图像处理库或AI服务
	
	elements := models.VisualElements{
		DominantColors: []string{"#FF5733", "#33FF57", "#3357FF"}, // 示例颜色
		Brightness:     0.6, // 示例亮度
		Contrast:       0.7, // 示例对比度
		Saturation:     0.5, // 示例饱和度
		HasText:        false, // 是否包含文字
		HasFaces:       false, // 是否包含人脸
		ObjectCount:    3,     // 对象数量
	}
	
	return elements
}

func (s *imageService) analyzeComposition(imgInfo models.Image) models.CompositionAnalysis {
	// 简化的构图分析
	composition := models.CompositionAnalysis{
		RuleOfThirds:  s.checkRuleOfThirds(imgInfo),
		Symmetry:      s.checkSymmetry(imgInfo),
		LeadingLines:  false, // 简化处理
		FramingScore:  0.7,
		BalanceScore:  0.6,
		FocusClarity:  0.8,
	}
	
	return composition
}

func (s *imageService) analyzeQuality(imgInfo models.Image) models.QualityMetrics {
	resolution := fmt.Sprintf("%dx%d", imgInfo.Width, imgInfo.Height)
	
	// 基于分辨率判断质量
	totalPixels := imgInfo.Width * imgInfo.Height
	var overallQuality float64
	
	if totalPixels >= 2000000 { // 2MP以上
		overallQuality = 0.9
	} else if totalPixels >= 1000000 { // 1MP以上
		overallQuality = 0.7
	} else if totalPixels >= 500000 { // 0.5MP以上
		overallQuality = 0.5
	} else {
		overallQuality = 0.3
	}
	
	return models.QualityMetrics{
		Resolution:     resolution,
		Sharpness:      0.8, // 示例值
		NoiseLevel:     0.2, // 示例值
		ExposureScore:  0.7, // 示例值
		OverallQuality: overallQuality,
	}
}

func (s *imageService) analyzeStyle(imgInfo models.Image) models.StyleAnalysis {
	// 简化的风格分析
	style := models.StyleAnalysis{
		Style:       s.determineStyle(imgInfo),
		Mood:        s.determineMood(imgInfo),
		Filter:      "none",
		Consistency: 0.8,
	}
	
	return style
}

func (s *imageService) checkRuleOfThirds(imgInfo models.Image) bool {
	// 简化判断：如果图片比例接近3:2或16:9，认为符合三分法则
	ratio := float64(imgInfo.Width) / float64(imgInfo.Height)
	return ratio > 1.4 && ratio < 1.8
}

func (s *imageService) checkSymmetry(imgInfo models.Image) bool {
	// 简化判断：如果宽高比接近1:1，认为具有对称性
	ratio := float64(imgInfo.Width) / float64(imgInfo.Height)
	return ratio > 0.9 && ratio < 1.1
}

func (s *imageService) determineStyle(imgInfo models.Image) string {
	// 基于图片特征简单判断风格
	aspectRatio := float64(imgInfo.Width) / float64(imgInfo.Height)
	
	if aspectRatio > 1.5 {
		return "landscape"
	} else if aspectRatio < 0.8 {
		return "portrait"
	}
	
	return "modern"
}

func (s *imageService) determineMood(imgInfo models.Image) string {
	// 简化的情绪判断
	// 实际应用中需要使用色彩分析或AI服务
	return "neutral"
}

func (s *imageService) calculateImageScore(analysis models.ImageAnalysis) float64 {
	score := 60.0 // 基础分
	
	// 质量评分影响
	score += analysis.QualityMetrics.OverallQuality * 25
	
	// 构图评分影响
	if analysis.CompositionAnalysis.RuleOfThirds {
		score += 5
	}
	if analysis.CompositionAnalysis.Symmetry {
		score += 5
	}
	
	score += analysis.CompositionAnalysis.BalanceScore * 5
	
	// 限制在0-100范围内
	if score > 100 {
		score = 100
	}
	if score < 0 {
		score = 0
	}
	
	return score
}
