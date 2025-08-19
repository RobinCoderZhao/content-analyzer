// internal/models/content.go
package models

import "time"

// Content 内容结构体
type Content struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Text        string     `json:"text"`
	Images      []Image    `json:"images,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
	PublishedAt time.Time  `json:"published_at,omitempty"`
	Author      string     `json:"author,omitempty"`
	FilePath    string     `json:"file_path,omitempty"`
	Type        string     `json:"type"` // post, story, video等
	Engagement  Engagement `json:"engagement,omitempty"`
}

// Image 图片信息
type Image struct {
	Path    string `json:"path"`
	URL     string `json:"url,omitempty"`
	Caption string `json:"caption,omitempty"`
	Width   int    `json:"width,omitempty"`
	Height  int    `json:"height,omitempty"`
	Size    int64  `json:"size,omitempty"`
	Format  string `json:"format,omitempty"`
}

// Engagement 互动数据
type Engagement struct {
	Likes    int `json:"likes,omitempty"`
	Comments int `json:"comments,omitempty"`
	Shares   int `json:"shares,omitempty"`
	Views    int `json:"views,omitempty"`
}

// AnalysisResult 分析结果
type AnalysisResult struct {
	ContentID     string             `json:"content_id"`
	Title         string             `json:"title"`
	Score         OverallScore       `json:"score"`
	TextAnalysis  TextAnalysis       `json:"text_analysis"`
	ImageAnalysis []ImageAnalysis    `json:"image_analysis,omitempty"`
	Suggestions   []Suggestion       `json:"suggestions"`
	Keywords      []Keyword          `json:"keywords"`
	Sentiment     SentimentAnalysis  `json:"sentiment"`
	Readability   ReadabilityMetrics `json:"readability"`
	CreatedAt     time.Time          `json:"created_at"`
}

// OverallScore 总体评分
type OverallScore struct {
	Total     float64        `json:"total"`     // 总分 0-100
	Breakdown ScoreBreakdown `json:"breakdown"` // 分项得分
	Level     string         `json:"level"`     // 等级: excellent, good, average, poor
	Reasoning string         `json:"reasoning"` // 评分理由
}

// ScoreBreakdown 分项评分
type ScoreBreakdown struct {
	ContentQuality float64 `json:"content_quality"` // 内容质量
	Engagement     float64 `json:"engagement"`      // 互动潜力
	Visual         float64 `json:"visual"`          // 视觉吸引力
	Title          float64 `json:"title"`           // 标题吸引力
	Readability    float64 `json:"readability"`     // 可读性
	TrendRelevance float64 `json:"trend_relevance"` // 趋势相关性
}

// TextAnalysis 文本分析结果
type TextAnalysis struct {
	WordCount        int              `json:"word_count"`
	CharCount        int              `json:"char_count"`
	ParagraphCount   int              `json:"paragraph_count"`
	SentenceCount    int              `json:"sentence_count"`
	TitleAnalysis    TitleAnalysis    `json:"title_analysis"`
	ContentStructure ContentStructure `json:"content_structure"`
	WritingStyle     WritingStyle     `json:"writing_style"`
	CallToAction     []string         `json:"call_to_action"`
	Hashtags         []string         `json:"hashtags"`
	Mentions         []string         `json:"mentions"`
}

// TitleAnalysis 标题分析
type TitleAnalysis struct {
	Length         int      `json:"length"`
	HasNumbers     bool     `json:"has_numbers"`
	HasEmoji       bool     `json:"has_emoji"`
	HasQuestions   bool     `json:"has_questions"`
	EmotionalWords []string `json:"emotional_words"`
	PowerWords     []string `json:"power_words"`
	ClickbaitScore float64  `json:"clickbait_score"`
	ClarityScore   float64  `json:"clarity_score"`
}

// ContentStructure 内容结构分析
type ContentStructure struct {
	HasIntro        bool   `json:"has_intro"`
	HasConclusion   bool   `json:"has_conclusion"`
	HasBulletPoints bool   `json:"has_bullet_points"`
	HasNumbers      bool   `json:"has_numbers"`
	SectionCount    int    `json:"section_count"`
	Structure       string `json:"structure"` // linear, story, list, qa等
}

// WritingStyle 写作风格分析
type WritingStyle struct {
	Tone              string  `json:"tone"`               // casual, formal, enthusiastic等
	PersonPerspective string  `json:"person_perspective"` // first, second, third
	Formality         float64 `json:"formality"`          // 0-1 正式程度
	Complexity        float64 `json:"complexity"`         // 0-1 复杂程度
	Authenticity      float64 `json:"authenticity"`       // 0-1 真实感
}

// ImageAnalysis 图片分析结果
type ImageAnalysis struct {
	Path                string              `json:"path"`
	VisualElements      VisualElements      `json:"visual_elements"`
	CompositionAnalysis CompositionAnalysis `json:"composition"`
	QualityMetrics      QualityMetrics      `json:"quality"`
	StyleAnalysis       StyleAnalysis       `json:"style"`
	Score               float64             `json:"score"`
}

// VisualElements 视觉元素分析
type VisualElements struct {
	DominantColors []string `json:"dominant_colors"`
	Brightness     float64  `json:"brightness"`
	Contrast       float64  `json:"contrast"`
	Saturation     float64  `json:"saturation"`
	HasText        bool     `json:"has_text"`
	HasFaces       bool     `json:"has_faces"`
	ObjectCount    int      `json:"object_count"`
}

// CompositionAnalysis 构图分析
type CompositionAnalysis struct {
	RuleOfThirds bool    `json:"rule_of_thirds"`
	Symmetry     bool    `json:"symmetry"`
	LeadingLines bool    `json:"leading_lines"`
	FramingScore float64 `json:"framing_score"`
	BalanceScore float64 `json:"balance_score"`
	FocusClarity float64 `json:"focus_clarity"`
}

// QualityMetrics 质量指标
type QualityMetrics struct {
	Resolution     string  `json:"resolution"`
	Sharpness      float64 `json:"sharpness"`
	NoiseLevel     float64 `json:"noise_level"`
	ExposureScore  float64 `json:"exposure_score"`
	OverallQuality float64 `json:"overall_quality"`
}

// StyleAnalysis 风格分析
type StyleAnalysis struct {
	Style       string  `json:"style"` // minimalist, vintage, modern等
	Mood        string  `json:"mood"`  // happy, calm, energetic等
	Filter      string  `json:"filter,omitempty"`
	Consistency float64 `json:"consistency"` // 与其他图片的一致性
}

// Suggestion 改进建议
type Suggestion struct {
	Type        string   `json:"type"`               // title, content, image, structure等
	Priority    string   `json:"priority"`           // high, medium, low
	Current     string   `json:"current"`            // 当前情况描述
	Recommended string   `json:"recommended"`        // 建议改进
	Reasoning   string   `json:"reasoning"`          // 建议理由
	Examples    []string `json:"examples,omitempty"` // 示例
	Impact      string   `json:"impact"`             // 预期影响
}

// Keyword 关键词分析
type Keyword struct {
	Word      string  `json:"word"`
	Frequency int     `json:"frequency"`
	Relevance float64 `json:"relevance"`
	Trend     string  `json:"trend"`    // rising, stable, declining
	Category  string  `json:"category"` // topic, emotion, action等
}

// SentimentAnalysis 情感分析
type SentimentAnalysis struct {
	Overall    string             `json:"overall"`    // positive, negative, neutral
	Score      float64            `json:"score"`      // -1 到 1
	Emotions   map[string]float64 `json:"emotions"`   // joy, anger, fear等情感得分
	Confidence float64            `json:"confidence"` // 置信度
}

// ReadabilityMetrics 可读性指标
type ReadabilityMetrics struct {
	FleschScore       float64 `json:"flesch_score"` // Flesch阅读难度
	AvgSentenceLength float64 `json:"avg_sentence_length"`
	AvgWordLength     float64 `json:"avg_word_length"`
	ComplexWordRatio  float64 `json:"complex_word_ratio"`
	ReadingTime       int     `json:"reading_time"` // 预估阅读时间（秒）
	Grade             string  `json:"grade"`        // 阅读等级
}
