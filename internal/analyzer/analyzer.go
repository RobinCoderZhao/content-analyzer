// internal/analyzer/analyzer.go
package analyzer

import (
	"context"
	"fmt"
	"math"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/content-analyzer/internal/config"
	"github.com/content-analyzer/internal/models"
	"github.com/content-analyzer/internal/services"
)

type ContentAnalyzer struct {
	config     *config.Config
	aiService  services.AIService
	imgService services.ImageService
}

func NewContentAnalyzer(cfg *config.Config) *ContentAnalyzer {
	return &ContentAnalyzer{
		config:     cfg,
		aiService:  services.NewAIService(cfg),
		imgService: services.NewImageService(cfg),
	}
}

// Analyze 分析单个内容
func (ca *ContentAnalyzer) Analyze(content models.Content) (models.AnalysisResult, error) {
	result := models.AnalysisResult{
		ContentID: content.ID,
		Title:     content.Title,
		CreatedAt: time.Now(),
	}

	// 1. 文本分析
	textAnalysis, err := ca.analyzeText(content)
	if err != nil {
		return result, fmt.Errorf("文本分析失败: %w", err)
	}
	result.TextAnalysis = textAnalysis

	// 2. 图片分析
	if len(content.Images) > 0 {
		imageAnalyses, err := ca.analyzeImages(content.Images)
		if err != nil {
			return result, fmt.Errorf("图片分析失败: %w", err)
		}
		result.ImageAnalysis = imageAnalyses
	}

	// 3. 情感分析
	sentiment, err := ca.analyzeSentiment(content.Text + " " + content.Title)
	if err != nil {
		return result, fmt.Errorf("情感分析失败: %w", err)
	}
	result.Sentiment = sentiment

	// 4. 关键词提取
	keywords := ca.extractKeywords(content.Text)
	result.Keywords = keywords

	// 5. 可读性分析
	readability := ca.analyzeReadability(content.Text)
	result.Readability = readability

	// 6. 生成评分
	score := ca.calculateOverallScore(result)
	result.Score = score

	// 7. 生成改进建议
	suggestions := ca.generateSuggestions(result)
	result.Suggestions = suggestions

	return result, nil
}

// analyzeText 文本分析
func (ca *ContentAnalyzer) analyzeText(content models.Content) (models.TextAnalysis, error) {
	text := content.Text
	title := content.Title

	analysis := models.TextAnalysis{
		WordCount:      ca.countWords(text),
		CharCount:      utf8.RuneCountInString(text),
		ParagraphCount: ca.countParagraphs(text),
		SentenceCount:  ca.countSentences(text),
		Hashtags:       ca.extractHashtags(text),
		Mentions:       ca.extractMentions(text),
		CallToAction:   ca.extractCallToActions(text),
	}

	// 标题分析
	analysis.TitleAnalysis = models.TitleAnalysis{
		Length:         utf8.RuneCountInString(title),
		HasNumbers:     ca.hasNumbers(title),
		HasEmoji:       ca.hasEmoji(title),
		HasQuestions:   ca.hasQuestions(title),
		EmotionalWords: ca.findEmotionalWords(title),
		PowerWords:     ca.findPowerWords(title),
		ClickbaitScore: ca.calculateClickbaitScore(title),
		ClarityScore:   ca.calculateClarityScore(title),
	}

	// 内容结构分析
	analysis.ContentStructure = models.ContentStructure{
		HasIntro:        ca.hasIntroduction(text),
		HasConclusion:   ca.hasConclusion(text),
		HasBulletPoints: ca.hasBulletPoints(text),
		HasNumbers:      ca.hasNumbers(text),
		SectionCount:    ca.countSections(text),
		Structure:       ca.identifyStructure(text),
	}

	// 写作风格分析
	analysis.WritingStyle = models.WritingStyle{
		Tone:              ca.identifyTone(text),
		PersonPerspective: ca.identifyPerspective(text),
		Formality:         ca.calculateFormality(text),
		Complexity:        ca.calculateComplexity(text),
		Authenticity:      ca.calculateAuthenticity(text),
	}

	return analysis, nil
}

// analyzeImages 图片分析
func (ca *ContentAnalyzer) analyzeImages(images []models.Image) ([]models.ImageAnalysis, error) {
	var analyses []models.ImageAnalysis

	for _, img := range images {
		// 检查图片路径
		imagePath := img.Path
		if !filepath.IsAbs(imagePath) {
			imagePath = filepath.Join(ca.config.ContentDir, imagePath)
		}

		analysis, err := ca.imgService.AnalyzeImage(imagePath)
		if err != nil {
			return nil, fmt.Errorf("分析图片 %s 失败: %w", imagePath, err)
		}

		analyses = append(analyses, analysis)
	}

	return analyses, nil
}

// analyzeSentiment 情感分析
func (ca *ContentAnalyzer) analyzeSentiment(text string) (models.SentimentAnalysis, error) {
	// 使用AI服务进行情感分析
	ctx := context.Background()
	sentiment, err := ca.aiService.AnalyzeSentiment(ctx, text)
	if err != nil {
		return models.SentimentAnalysis{}, err
	}

	return sentiment, nil
}

// 文本处理工具函数
func (ca *ContentAnalyzer) countWords(text string) int {
	words := strings.Fields(text)
	return len(words)
}

func (ca *ContentAnalyzer) countParagraphs(text string) int {
	paragraphs := strings.Split(strings.TrimSpace(text), "\n\n")
	count := 0
	for _, p := range paragraphs {
		if strings.TrimSpace(p) != "" {
			count++
		}
	}
	return count
}

func (ca *ContentAnalyzer) countSentences(text string) int {
	// 简单的句子计数，基于标点符号
	re := regexp.MustCompile(`[.!?]+`)
	sentences := re.Split(text, -1)
	count := 0
	for _, s := range sentences {
		if strings.TrimSpace(s) != "" {
			count++
		}
	}
	return count
}

func (ca *ContentAnalyzer) countSections(text string) int {
	// 通过标题标记或空行分隔计算章节数
	lines := strings.Split(text, "\n")
	sections := 1

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// 检查是否是标题（以#开头或者全大写短行）
		if strings.HasPrefix(trimmed, "#") ||
			(len(trimmed) < 50 && strings.ToUpper(trimmed) == trimmed && len(trimmed) > 0) {
			sections++
		}
	}

	return sections
}

func (ca *ContentAnalyzer) extractHashtags(text string) []string {
	re := regexp.MustCompile(`#[\p{L}\p{N}_]+`)
	return re.FindAllString(text, -1)
}

func (ca *ContentAnalyzer) extractMentions(text string) []string {
	re := regexp.MustCompile(`@[\p{L}\p{N}_]+`)
	return re.FindAllString(text, -1)
}

func (ca *ContentAnalyzer) extractCallToActions(text string) []string {
	// 常见的CTA模式
	ctaPatterns := []string{
		`点击.*链接`, `立即.*`, `马上.*`, `赶快.*`, `快来.*`,
		`关注我`, `点赞.*`, `评论.*`, `分享.*`, `收藏.*`,
		`了解更多`, `查看更多`, `阅读全文`,
	}

	var ctas []string
	text = strings.ToLower(text)

	for _, pattern := range ctaPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllString(text, -1)
		ctas = append(ctas, matches...)
	}

	return ctas
}

func (ca *ContentAnalyzer) hasNumbers(text string) bool {
	re := regexp.MustCompile(`\d+`)
	return re.MatchString(text)
}

func (ca *ContentAnalyzer) hasEmoji(text string) bool {
	// 简单的emoji检测
	re := regexp.MustCompile(`[\x{1F600}-\x{1F64F}]|[\x{1F300}-\x{1F5FF}]|[\x{1F680}-\x{1F6FF}]|[\x{2600}-\x{26FF}]|[\x{2700}-\x{27BF}]`)
	return re.MatchString(text)
}

func (ca *ContentAnalyzer) hasQuestions(text string) bool {
	return strings.Contains(text, "?") || strings.Contains(text, "？")
}

func (ca *ContentAnalyzer) hasIntroduction(text string) bool {
	// 检查是否有引言性质的开头
	intro_patterns := []string{
		"大家好", "hello", "今天", "最近", "分享", "介绍",
	}

	firstSentence := strings.ToLower(text)
	if len(firstSentence) > 100 {
		firstSentence = firstSentence[:100]
	}

	for _, pattern := range intro_patterns {
		if strings.Contains(firstSentence, pattern) {
			return true
		}
	}

	return false
}

func (ca *ContentAnalyzer) hasConclusion(text string) bool {
	// 检查是否有总结性质的结尾
	conclusion_patterns := []string{
		"总结", "总之", "最后", "综上", "结论", "希望", "感谢",
	}

	lastSentence := strings.ToLower(text)
	if len(lastSentence) > 100 {
		start := len(lastSentence) - 100
		lastSentence = lastSentence[start:]
	}

	for _, pattern := range conclusion_patterns {
		if strings.Contains(lastSentence, pattern) {
			return true
		}
	}

	return false
}

func (ca *ContentAnalyzer) hasBulletPoints(text string) bool {
	// 检查是否有列表项
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "•") ||
			strings.HasPrefix(trimmed, "-") ||
			strings.HasPrefix(trimmed, "*") ||
			regexp.MustCompile(`^\d+\.`).MatchString(trimmed) {
			return true
		}
	}
	return false
}

// 更多分析函数待实现...
func (ca *ContentAnalyzer) findEmotionalWords(text string) []string {
	emotionalWords := []string{
		"惊喜", "震撼", "感动", "激动", "兴奋", "满足", "幸福", "快乐",
		"担心", "焦虑", "害怕", "紧张", "愤怒", "失望", "沮丧",
		"amazing", "wonderful", "fantastic", "incredible", "awesome",
	}

	var found []string
	lowerText := strings.ToLower(text)

	for _, word := range emotionalWords {
		if strings.Contains(lowerText, strings.ToLower(word)) {
			found = append(found, word)
		}
	}

	return found
}

func (ca *ContentAnalyzer) findPowerWords(text string) []string {
	powerWords := []string{
		"独家", "限时", "免费", "秘密", "揭秘", "内幕", "独特", "创新",
		"突破", "革命", "颠覆", "神器", "必备", "推荐", "精选",
		"exclusive", "limited", "secret", "unique", "breakthrough",
	}

	var found []string
	lowerText := strings.ToLower(text)

	for _, word := range powerWords {
		if strings.Contains(lowerText, strings.ToLower(word)) {
			found = append(found, word)
		}
	}

	return found
}

func (ca *ContentAnalyzer) calculateClickbaitScore(title string) float64 {
	score := 0.0

	// 各种clickbait特征检查
	if ca.hasNumbers(title) {
		score += 0.2
	}
	if ca.hasQuestions(title) {
		score += 0.15
	}
	if len(ca.findPowerWords(title)) > 0 {
		score += 0.3
	}
	if strings.Contains(strings.ToLower(title), "你不知道") ||
		strings.Contains(strings.ToLower(title), "震惊") {
		score += 0.4
	}

	// 限制在0-1范围内
	if score > 1.0 {
		score = 1.0
	}

	return score
}

func (ca *ContentAnalyzer) calculateClarityScore(title string) float64 {
	// 简单的清晰度评分逻辑
	score := 1.0

	length := utf8.RuneCountInString(title)
	if length > 50 {
		score -= 0.2 // 太长降分
	}
	if length < 5 {
		score -= 0.3 // 太短降分
	}

	// 检查是否有明确的主题词
	if !ca.hasNumbers(title) && len(ca.findPowerWords(title)) == 0 {
		score -= 0.1
	}

	return score
}

// 其他待实现的分析方法...
func (ca *ContentAnalyzer) identifyTone(text string) string {
	// 基于关键词识别语调
	if len(ca.findEmotionalWords(text)) > 3 {
		return "enthusiastic"
	}
	if strings.Contains(text, "。") && !ca.hasQuestions(text) {
		return "formal"
	}
	return "casual"
}

func (ca *ContentAnalyzer) identifyPerspective(text string) string {
	firstPerson := strings.Count(text, "我") + strings.Count(text, "I ")
	secondPerson := strings.Count(text, "你") + strings.Count(text, "您") + strings.Count(text, "you")

	if firstPerson > secondPerson {
		return "first"
	} else if secondPerson > 0 {
		return "second"
	}
	return "third"
}

func (ca *ContentAnalyzer) calculateFormality(text string) float64 {
	// 基于词汇和句式判断正式程度
	formalWords := []string{"因此", "然而", "此外", "综上所述", "鉴于", "据此"}
	casualWords := []string{"哈哈", "嗯", "呀", "哦", "额", "咋样"}

	formalCount := 0
	casualCount := 0

	lowerText := strings.ToLower(text)
	for _, word := range formalWords {
		formalCount += strings.Count(lowerText, word)
	}
	for _, word := range casualWords {
		casualCount += strings.Count(lowerText, word)
	}

	totalWords := ca.countWords(text)
	if totalWords == 0 {
		return 0.5
	}

	score := 0.5 + float64(formalCount-casualCount)/float64(totalWords)
	if score < 0 {
		score = 0
	}
	if score > 1 {
		score = 1
	}

	return score
}

func (ca *ContentAnalyzer) calculateComplexity(text string) float64 {
	words := strings.Fields(text)
	if len(words) == 0 {
		return 0
	}

	complexWords := 0
	totalChars := 0

	for _, word := range words {
		charCount := utf8.RuneCountInString(word)
		totalChars += charCount
		if charCount > 6 { // 认为超过6个字符的词是复杂词
			complexWords++
		}
	}

	avgWordLength := float64(totalChars) / float64(len(words))
	complexityRatio := float64(complexWords) / float64(len(words))

	// 综合平均词长和复杂词比例
	complexity := (avgWordLength/10.0)*0.6 + complexityRatio*0.4
	if complexity > 1 {
		complexity = 1
	}

	return complexity
}

func (ca *ContentAnalyzer) calculateAuthenticity(text string) float64 {
	// 基于一些指标判断真实感
	score := 0.8 // 基础分

	// 有个人经历描述加分
	personalIndicators := []string{"我觉得", "我认为", "我的经验", "亲身体验", "我发现"}
	for _, indicator := range personalIndicators {
		if strings.Contains(text, indicator) {
			score += 0.05
		}
	}

	// 过度营销词汇减分
	marketingWords := []string{"绝对", "百分百", "保证", "必定", "一定能"}
	for _, word := range marketingWords {
		if strings.Contains(text, word) {
			score -= 0.1
		}
	}

	if score > 1 {
		score = 1
	}
	if score < 0 {
		score = 0
	}

	return score
}

func (ca *ContentAnalyzer) identifyStructure(text string) string {
	// 识别内容结构类型
	if ca.hasBulletPoints(text) {
		return "list"
	}
	if strings.Contains(text, "?") || strings.Contains(text, "？") {
		return "qa"
	}
	if ca.hasIntroduction(text) && ca.hasConclusion(text) {
		return "story"
	}
	return "linear"
}

func (ca *ContentAnalyzer) extractKeywords(text string) []models.Keyword {
	// 简单的关键词提取
	words := strings.Fields(strings.ToLower(text))
	wordCount := make(map[string]int)

	// 停用词列表
	stopWords := map[string]bool{
		"的": true, "是": true, "在": true, "我": true, "你": true, "他": true,
		"了": true, "和": true, "就": true, "都": true, "而": true, "及": true,
		"与": true, "或": true, "但": true, "为": true, "也": true, "不": true,
		"可以": true, "这个": true, "那个": true, "什么": true, "怎么": true,
		"the", "a", "an", "and", "or", "but", "in", "on", "at", "to",
		"for", "of", "with", "by", "is", "are", "was", "were", "be",
	}

	for _, word := range words {
		// 清理标点符号
		word = regexp.MustCompile(`[^\p{L}\p{N}]+`).ReplaceAllString(word, "")
		if len(word) > 1 && !stopWords[word] {
			wordCount[word]++
		}
	}

	// 转换为关键词结构
	var keywords []models.Keyword
	for word, count := range wordCount {
		if count >= 2 { // 至少出现2次才算关键词
			relevance := float64(count) / float64(len(words))
			keywords = append(keywords, models.Keyword{
				Word:      word,
				Frequency: count,
				Relevance: relevance,
				Trend:     "stable", // 简化处理
				Category:  ca.categorizeKeyword(word),
			})
		}
	}

	return keywords
}

func (ca *ContentAnalyzer) categorizeKeyword(word string) string {
	// 简单的关键词分类
	emotionWords := []string{"好", "棒", "差", "爱", "恨", "喜欢", "讨厌"}
	actionWords := []string{"做", "买", "用", "看", "听", "学", "教"}

	for _, ew := range emotionWords {
		if strings.Contains(word, ew) {
			return "emotion"
		}
	}

	for _, aw := range actionWords {
		if strings.Contains(word, aw) {
			return "action"
		}
	}

	return "topic"
}

func (ca *ContentAnalyzer) analyzeReadability(text string) models.ReadabilityMetrics {
	wordCount := ca.countWords(text)
	sentenceCount := ca.countSentences(text)

	if sentenceCount == 0 {
		sentenceCount = 1
	}

	avgSentenceLength := float64(wordCount) / float64(sentenceCount)

	// 计算平均词长
	words := strings.Fields(text)
	totalChars := 0
	complexWords := 0

	for _, word := range words {
		charCount := utf8.RuneCountInString(word)
		totalChars += charCount
		if charCount > 6 {
			complexWords++
		}
	}

	var avgWordLength float64
	var complexWordRatio float64

	if len(words) > 0 {
		avgWordLength = float64(totalChars) / float64(len(words))
		complexWordRatio = float64(complexWords) / float64(len(words))
	}

	// 简化的Flesch分数计算
	fleschScore := 206.835 - 1.015*avgSentenceLength - 84.6*(avgWordLength/4.7)

	// 阅读等级判定
	grade := "中等"
	if fleschScore > 80 {
		grade = "容易"
	} else if fleschScore < 50 {
		grade = "困难"
	}

	// 预估阅读时间（按每分钟250字计算）
	readingTime := int(float64(wordCount) / 250.0 * 60)
	if readingTime < 30 {
		readingTime = 30 // 最少30秒
	}

	return models.ReadabilityMetrics{
		FleschScore:       fleschScore,
		AvgSentenceLength: avgSentenceLength,
		AvgWordLength:     avgWordLength,
		ComplexWordRatio:  complexWordRatio,
		ReadingTime:       readingTime,
		Grade:             grade,
	}
}

func (ca *ContentAnalyzer) calculateOverallScore(result models.AnalysisResult) models.OverallScore {
	breakdown := models.ScoreBreakdown{
		ContentQuality: ca.scoreContentQuality(result.TextAnalysis),
		Engagement:     ca.scoreEngagement(result.TextAnalysis),
		Visual:         ca.scoreVisual(result.ImageAnalysis),
		Title:          ca.scoreTitle(result.TextAnalysis.TitleAnalysis),
		Readability:    ca.scoreReadability(result.Readability),
		TrendRelevance: ca.scoreTrendRelevance(result.Keywords),
	}

	// 计算总分（加权平均）
	total := breakdown.ContentQuality*0.25 +
		breakdown.Engagement*0.20 +
		breakdown.Visual*0.15 +
		breakdown.Title*0.15 +
		breakdown.Readability*0.15 +
		breakdown.TrendRelevance*0.10

	level := "poor"
	if total >= 85 {
		level = "excellent"
	} else if total >= 70 {
		level = "good"
	} else if total >= 50 {
		level = "average"
	}

	reasoning := fmt.Sprintf("综合评分%.1f分，主要优势在%s，需要改进%s",
		total, ca.findStrengths(breakdown), ca.findWeaknesses(breakdown))

	return models.OverallScore{
		Total:     total,
		Breakdown: breakdown,
		Level:     level,
		Reasoning: reasoning,
	}
}

func (ca *ContentAnalyzer) scoreContentQuality(textAnalysis models.TextAnalysis) float64 {
	score := 60.0 // 基础分

	// 长度适中加分
	if textAnalysis.WordCount >= 100 && textAnalysis.WordCount <= 800 {
		score += 20
	}

	// 结构完整加分
	if textAnalysis.ContentStructure.HasIntro && textAnalysis.ContentStructure.HasConclusion {
		score += 15
	}

	// 有明确的CTA加分
	if len(textAnalysis.CallToAction) > 0 {
		score += 5
	}

	return math.Min(score, 100)
}

func (ca *ContentAnalyzer) scoreEngagement(textAnalysis models.TextAnalysis) float64 {
	score := 50.0

	// 互动元素
	if len(textAnalysis.CallToAction) > 0 {
		score += 20
	}
	if textAnalysis.TitleAnalysis.HasQuestions {
		score += 15
	}
	if len(textAnalysis.TitleAnalysis.EmotionalWords) > 0 {
		score += 10
	}
	if textAnalysis.WritingStyle.PersonPerspective == "second" {
		score += 5
	}

	return math.Min(score, 100)
}

func (ca *ContentAnalyzer) scoreVisual(imageAnalysis []models.ImageAnalysis) float64 {
	if len(imageAnalysis) == 0 {
		return 30.0 // 无图片的基础分
	}

	totalScore := 0.0
	for _, img := range imageAnalysis {
		totalScore += img.Score
	}

	return totalScore / float64(len(imageAnalysis))
}

func (ca *ContentAnalyzer) scoreTitle(titleAnalysis models.TitleAnalysis) float64 {
	score := 50.0

	// 长度适中
	if titleAnalysis.Length >= 10 && titleAnalysis.Length <= 30 {
		score += 20
	}

	// 有吸引力元素
	if titleAnalysis.HasNumbers {
		score += 10
	}
	if len(titleAnalysis.PowerWords) > 0 {
		score += 15
	}
	if titleAnalysis.ClarityScore > 0.8 {
		score += 5
	}

	return math.Min(score, 100)
}

func (ca *ContentAnalyzer) scoreReadability(readability models.ReadabilityMetrics) float64 {
	score := 50.0

	// Flesch分数越高可读性越好
	if readability.FleschScore > 70 {
		score += 30
	} else if readability.FleschScore > 50 {
		score += 20
	} else if readability.FleschScore > 30 {
		score += 10
	}

	// 句子长度适中
	if readability.AvgSentenceLength >= 10 && readability.AvgSentenceLength <= 20 {
		score += 10
	}

	// 复杂词汇不要太多
	if readability.ComplexWordRatio < 0.2 {
		score += 10
	}

	return math.Min(score, 100)
}

func (ca *ContentAnalyzer) scoreTrendRelevance(keywords []models.Keyword) float64 {
	// 简化的趋势相关性评分
	score := 60.0

	for _, keyword := range keywords {
		if keyword.Trend == "rising" {
			score += 5
		}
		if keyword.Relevance > 0.05 { // 高相关性关键词
			score += 2
		}
	}

	return math.Min(score, 100)
}

func (ca *ContentAnalyzer) findStrengths(breakdown models.ScoreBreakdown) string {
	scores := map[string]float64{
		"内容质量": breakdown.ContentQuality,
		"互动性":  breakdown.Engagement,
		"视觉效果": breakdown.Visual,
		"标题":   breakdown.Title,
		"可读性":  breakdown.Readability,
		"趋势性":  breakdown.TrendRelevance,
	}

	maxScore := 0.0
	strength := ""
	for area, score := range scores {
		if score > maxScore {
			maxScore = score
			strength = area
		}
	}

	return strength
}

func (ca *ContentAnalyzer) findWeaknesses(breakdown models.ScoreBreakdown) string {
	scores := map[string]float64{
		"内容质量": breakdown.ContentQuality,
		"互动性":  breakdown.Engagement,
		"视觉效果": breakdown.Visual,
		"标题":   breakdown.Title,
		"可读性":  breakdown.Readability,
		"趋势性":  breakdown.TrendRelevance,
	}

	minScore := 100.0
	weakness := ""
	for area, score := range scores {
		if score < minScore {
			minScore = score
			weakness = area
		}
	}

	return weakness
}

func (ca *ContentAnalyzer) generateSuggestions(result models.AnalysisResult) []models.Suggestion {
	var suggestions []models.Suggestion

	// 标题建议
	if result.Score.Breakdown.Title < 70 {
		suggestions = append(suggestions, models.Suggestion{
			Type:        "title",
			Priority:    "high",
			Current:     "当前标题吸引力不足",
			Recommended: "建议添加数字、提问或者情感词汇来增强标题吸引力",
			Reasoning:   fmt.Sprintf("标题得分仅%.1f分，低于平均水平", result.Score.Breakdown.Title),
			Impact:      "预计可提升点击率15-25%",
		})
	}

	// 内容结构建议
	if !result.TextAnalysis.ContentStructure.HasIntro {
		suggestions = append(suggestions, models.Suggestion{
			Type:        "structure",
			Priority:    "medium",
			Current:     "缺少引人入胜的开头",
			Recommended: "添加一个吸引读者注意力的开场，比如提问、故事或者数据",
			Reasoning:   "好的开头能够显著提高读者的阅读完成率",
			Impact:      "预计可提升完读率20%",
		})
	}

	// 互动性建议
	if len(result.TextAnalysis.CallToAction) == 0 {
		suggestions = append(suggestions, models.Suggestion{
			Type:        "engagement",
			Priority:    "high",
			Current:     "缺少行动召唤元素",
			Recommended: "在适当位置添加引导用户互动的内容，如'你觉得呢？'、'记得点赞收藏'",
			Reasoning:   "CTA能够显著提升用户参与度",
			Examples:    []string{"你遇到过类似情况吗？", "快来评论区分享你的经验", "觉得有用请点个赞"},
			Impact:      "预计可提升互动率30%",
		})
	}

	// 可读性建议
	if result.Readability.FleschScore < 50 {
		suggestions = append(suggestions, models.Suggestion{
			Type:        "readability",
			Priority:    "medium",
			Current:     "文本可读性偏低",
			Recommended: "尝试使用更短的句子和更简单的词汇",
			Reasoning:   fmt.Sprintf("当前可读性得分%.1f，建议提升到60以上", result.Readability.FleschScore),
			Impact:      "预计可提升用户阅读体验",
		})
	}

	// 视觉内容建议
	if len(result.ImageAnalysis) == 0 {
		suggestions = append(suggestions, models.Suggestion{
			Type:        "visual",
			Priority:    "high",
			Current:     "内容缺少视觉元素",
			Recommended: "添加相关图片、图表或者视觉元素来增强内容吸引力",
			Reasoning:   "视觉内容能够显著提升用户参与度和分享率",
			Impact:      "预计可提升参与度40-60%",
		})
	}

	return suggestions
}
