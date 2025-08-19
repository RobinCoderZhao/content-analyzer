// internal/services/ai_service.go
package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/content-analyzer/internal/config"
	"github.com/content-analyzer/internal/models"
)

type AIService interface {
	AnalyzeSentiment(ctx context.Context, text string) (models.SentimentAnalysis, error)
	GenerateAdvice(ctx context.Context, analysis models.AnalysisResult) (string, error)
	ExtractTopics(ctx context.Context, text string) ([]string, error)
	ImproveContent(ctx context.Context, content string, suggestions []models.Suggestion) (string, error)
}

type aiService struct {
	config     *config.Config
	httpClient *http.Client
}

type OpenAIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIResponse struct {
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

type Choice struct {
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

func NewAIService(cfg *config.Config) AIService {
	return &aiService{
		config: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (s *aiService) AnalyzeSentiment(ctx context.Context, text string) (models.SentimentAnalysis, error) {
	// 如果没有配置API密钥，使用简化版本
	if s.config.AI.APIKey == "" {
		return s.simpleSentimentAnalysis(text), nil
	}

	prompt := fmt.Sprintf(`请分析以下文本的情感倾向，返回JSON格式：
{
  "overall": "positive/negative/neutral",
  "score": -1到1之间的数字,
  "emotions": {
    "joy": 0-1之间的数字,
    "sadness": 0-1之间的数字,
    "anger": 0-1之间的数字,
    "fear": 0-1之间的数字,
    "surprise": 0-1之间的数字
  },
  "confidence": 0-1之间的数字
}

文本内容：
%s`, text)

	response, err := s.callAI(ctx, prompt)
	if err != nil {
		// 如果AI调用失败，降级到简单版本
		return s.simpleSentimentAnalysis(text), nil
	}

	var sentiment models.SentimentAnalysis
	if err := json.Unmarshal([]byte(response), &sentiment); err != nil {
		// 解析失败，使用简单版本
		return s.simpleSentimentAnalysis(text), nil
	}

	return sentiment, nil
}

func (s *aiService) GenerateAdvice(ctx context.Context, analysis models.AnalysisResult) (string, error) {
	if s.config.AI.APIKey == "" {
		return s.simpleAdviceGeneration(analysis), nil
	}

	prompt := fmt.Sprintf(`基于以下内容分析结果，生成详细的改进建议：

标题：%s
总分：%.1f
各项得分：
- 内容质量：%.1f
- 互动性：%.1f  
- 视觉效果：%.1f
- 标题质量：%.1f
- 可读性：%.1f
- 趋势相关性：%.1f

文本分析：
- 字数：%d
- 句子数：%d
- 段落数：%d
- 是否有引言：%t
- 是否有结论：%t
- 行动召唤数量：%d

请生成具体、可执行的改进建议，包括：
1. 优先级最高的3个改进点
2. 每个改进点的具体操作建议
3. 预期效果说明`,
		analysis.Title,
		analysis.Score.Total,
		analysis.Score.Breakdown.ContentQuality,
		analysis.Score.Breakdown.Engagement,
		analysis.Score.Breakdown.Visual,
		analysis.Score.Breakdown.Title,
		analysis.Score.Breakdown.Readability,
		analysis.Score.Breakdown.TrendRelevance,
		analysis.TextAnalysis.WordCount,
		analysis.TextAnalysis.SentenceCount,
		analysis.TextAnalysis.ParagraphCount,
		analysis.TextAnalysis.ContentStructure.HasIntro,
		analysis.TextAnalysis.ContentStructure.HasConclusion,
		len(analysis.TextAnalysis.CallToAction),
	)

	response, err := s.callAI(ctx, prompt)
	if err != nil {
		return s.simpleAdviceGeneration(analysis), nil
	}

	return response, nil
}

func (s *aiService) ExtractTopics(ctx context.Context, text string) ([]string, error) {
	if s.config.AI.APIKey == "" {
		return s.simpleTopicExtraction(text), nil
	}

	prompt := fmt.Sprintf(`从以下文本中提取主要话题标签，返回JSON数组格式：
["话题1", "话题2", "话题3"]

要求：
1. 最多返回5个最相关的话题
2. 话题应该简洁明了
3. 优先选择热门话题标签

文本内容：
%s`, text)

	response, err := s.callAI(ctx, prompt)
	if err != nil {
		return s.simpleTopicExtraction(text), nil
	}

	var topics []string
	if err := json.Unmarshal([]byte(response), &topics); err != nil {
		return s.simpleTopicExtraction(text), nil
	}

	return topics, nil
}

func (s *aiService) ImproveContent(ctx context.Context, content string, suggestions []models.Suggestion) (string, error) {
	if s.config.AI.APIKey == "" {
		return content, fmt.Errorf("AI service not configured")
	}

	suggestionText := ""
	for _, suggestion := range suggestions {
		suggestionText += fmt.Sprintf("- %s: %s\n", suggestion.Type, suggestion.Recommended)
	}

	prompt := fmt.Sprintf(`请根据以下改进建议优化内容：

改进建议：
%s

原内容：
%s

请返回优化后的内容，保持原有风格的同时应用改进建议。`, suggestionText, content)

	return s.callAI(ctx, prompt)
}

func (s *aiService) callAI(ctx context.Context, prompt string) (string, error) {
	switch s.config.AI.Provider {
	case "openai":
		return s.callOpenAI(ctx, prompt)
	case "claude":
		return s.callClaude(ctx, prompt)
	default:
		return "", fmt.Errorf("unsupported AI provider: %s", s.config.AI.Provider)
	}
}

func (s *aiService) callOpenAI(ctx context.Context, prompt string) (string, error) {
	url := "https://api.openai.com/v1/chat/completions"
	if s.config.AI.BaseURL != "" {
		url = s.config.AI.BaseURL + "/chat/completions"
	}

	reqBody := OpenAIRequest{
		Model: s.config.AI.Model,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
		Temperature: 0.7,
		MaxTokens:   1000,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.config.AI.APIKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var response OpenAIResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return response.Choices[0].Message.Content, nil
}

func (s *aiService) callClaude(ctx context.Context, prompt string) (string, error) {
	// Claude API调用实现
	// 这里可以实现Claude API的调用逻辑
	return "", fmt.Errorf("Claude API not implemented yet")
}

// 简化版本的分析方法，不依赖AI API
func (s *aiService) simpleSentimentAnalysis(text string) models.SentimentAnalysis {
	positiveWords := []string{"好", "棒", "优秀", "喜欢", "爱", "开心", "满意", "推荐", "amazing", "great", "excellent", "wonderful", "fantastic"}
	negativeWords := []string{"差", "坏", "糟糕", "讨厌", "恨", "失望", "不满", "后悔", "terrible", "awful", "horrible", "disappointing"}

	text = strings.ToLower(text)
	positiveCount := 0
	negativeCount := 0

	for _, word := range positiveWords {
		positiveCount += strings.Count(text, strings.ToLower(word))
	}

	for _, word := range negativeWords {
		negativeCount += strings.Count(text, strings.ToLower(word))
	}

	overall := "neutral"
	score := 0.0

	if positiveCount > negativeCount {
		overall = "positive"
		score = 0.6
	} else if negativeCount > positiveCount {
		overall = "negative"
		score = -0.6
	}

	// 简单的情感分析
	emotions := map[string]float64{
		"joy":      0.0,
		"sadness":  0.0,
		"anger":    0.0,
		"fear":     0.0,
		"surprise": 0.0,
	}

	if strings.Contains(text, "开心") || strings.Contains(text, "高兴") || strings.Contains(text, "快乐") {
		emotions["joy"] = 0.7
	}
	if strings.Contains(text, "难过") || strings.Contains(text, "伤心") {
		emotions["sadness"] = 0.7
	}
	if strings.Contains(text, "生气") || strings.Contains(text, "愤怒") {
		emotions["anger"] = 0.7
	}

	return models.SentimentAnalysis{
		Overall:    overall,
		Score:      score,
		Emotions:   emotions,
		Confidence: 0.6,
	}
}

func (s *aiService) simpleAdviceGeneration(analysis models.AnalysisResult) string {
	advice := strings.Builder{}
	advice.WriteString("## 内容分析建议\n\n")

	if analysis.Score.Total < 60 {
		advice.WriteString("📊 **总体评分偏低**，建议重点关注以下方面：\n\n")
	} else if analysis.Score.Total < 80 {
		advice.WriteString("📊 **总体表现良好**，还有提升空间：\n\n")
	} else {
		advice.WriteString("📊 **内容质量优秀**，保持当前水准：\n\n")
	}

	// 根据各项得分给出具体建议
	if analysis.Score.Breakdown.Title < 70 {
		advice.WriteString("### 🎯 标题优化（优先级：高）\n")
		advice.WriteString("- **现状**：当前标题吸引力不足\n")
		advice.WriteString("- **建议**：使用数字、提问或情感词汇增强吸引力\n")
		advice.WriteString("- **示例**：添加\"5个\"、\"如何\"、\"超实用\"等词汇\n")
		advice.WriteString("- **预期效果**：提升点击率15-25%\n\n")
	}

	if analysis.Score.Breakdown.Engagement < 70 {
		advice.WriteString("### 💬 增强互动（优先级：高）\n")
		advice.WriteString("- **现状**：缺乏用户互动引导\n")
		advice.WriteString("- **建议**：添加问题引导、call-to-action等元素\n")
		advice.WriteString("- **示例**：\"你觉得呢？\"、\"快来评论区分享\"\n")
		advice.WriteString("- **预期效果**：提升互动率30-40%\n\n")
	}

	if len(analysis.ImageAnalysis) == 0 {
		advice.WriteString("### 🖼️ 视觉内容（优先级：中）\n")
		advice.WriteString("- **现状**：缺少视觉元素\n")
		advice.WriteString("- **建议**：添加相关图片或视觉元素\n")
		advice.WriteString("- **要点**：确保图片清晰、相关性强\n")
		advice.WriteString("- **预期效果**：提升参与度40-60%\n\n")
	}

	if analysis.Readability.FleschScore < 60 {
		advice.WriteString("### 📖 可读性优化（优先级：中）\n")
		advice.WriteString("- **现状**：文本可读性偏低\n")
		advice.WriteString("- **建议**：使用更短的句子和简单词汇\n")
		advice.WriteString("- **要点**：平均句长控制在15-20字\n")
		advice.WriteString("- **预期效果**：提升用户阅读体验\n\n")
	}

	advice.WriteString("### 📈 执行建议\n")
	advice.WriteString("1. **立即执行**：标题和互动优化（1-2天内完成）\n")
	advice.WriteString("2. **短期规划**：增加视觉内容（1周内完成）\n")
	advice.WriteString("3. **持续改进**：定期检查可读性和用户反馈\n")

	return advice.String()
}

func (s *aiService) simpleTopicExtraction(text string) []string {
	topics := []string{}

	// 基于关键词匹配识别主题
	topicKeywords := map[string][]string{
		"美食": {"吃", "食物", "餐厅", "菜", "味道", "好吃", "料理", "烹饪"},
		"旅行": {"旅游", "景点", "酒店", "机票", "攻略", "风景", "旅行", "度假"},
		"科技": {"手机", "电脑", "软件", "APP", "数码", "互联网", "科技", "技术"},
		"时尚": {"穿搭", "化妆", "护肤", "衣服", "搭配", "美妆", "时尚", "潮流"},
		"生活": {"日常", "分享", "经验", "感受", "生活", "日记", "心情", "随感"},
		"健康": {"健身", "运动", "养生", "保健", "医疗", "健康", "锻炼", "营养"},
		"教育": {"学习", "教程", "知识", "技能", "培训", "教育", "课程", "学校"},
		"娱乐": {"电影", "音乐", "游戏", "娱乐", "明星", "综艺", "动漫", "小说"},
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

	// 最多返回5个话题
	if len(topics) > 5 {
		topics = topics[:5]
	}

	return topics
}
