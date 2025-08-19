// internal/report/reporter.go
package report

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/content-analyzer/internal/config"
	"github.com/content-analyzer/internal/models"
)

type Reporter struct {
	config *config.Config
}

func NewReporter(cfg *config.Config) *Reporter {
	return &Reporter{config: cfg}
}

type ReportData struct {
	GeneratedAt     time.Time               `json:"generated_at"`
	TotalContent    int                     `json:"total_content"`
	OverallScore    float64                 `json:"overall_score"`
	Results         []models.AnalysisResult `json:"results"`
	Summary         ReportSummary           `json:"summary"`
	TopKeywords     []models.Keyword        `json:"top_keywords"`
	Recommendations []GlobalRecommendation  `json:"recommendations"`
}

type ReportSummary struct {
	AverageScores   models.ScoreBreakdown `json:"average_scores"`
	BestPerforming  string                `json:"best_performing"`
	NeedImprovement string                `json:"need_improvement"`
	CommonIssues    []string              `json:"common_issues"`
	SuccessPatterns []string              `json:"success_patterns"`
}

type GlobalRecommendation struct {
	Category        string   `json:"category"`
	Priority        string   `json:"priority"`
	Description     string   `json:"description"`
	AffectedContent []string `json:"affected_content"`
	ExpectedImpact  string   `json:"expected_impact"`
}

func (r *Reporter) GenerateReport(results []models.AnalysisResult) error {
	// 创建输出目录
	if err := os.MkdirAll(r.config.OutputDir, 0755); err != nil {
		return fmt.Errorf("创建输出目录失败: %w", err)
	}

	// 生成报告数据
	reportData := r.generateReportData(results)

	// 生成JSON报告
	if err := r.generateJSONReport(reportData); err != nil {
		return fmt.Errorf("生成JSON报告失败: %w", err)
	}

	// 生成HTML报告
	if err := r.generateHTMLReport(reportData); err != nil {
		return fmt.Errorf("生成HTML报告失败: %w", err)
	}

	// 生成CSV报告
	if err := r.generateCSVReport(reportData); err != nil {
		return fmt.Errorf("生成CSV报告失败: %w", err)
	}

	return nil
}

func (r *Reporter) generateReportData(results []models.AnalysisResult) ReportData {
	data := ReportData{
		GeneratedAt:  time.Now(),
		TotalContent: len(results),
		Results:      results,
	}

	if len(results) == 0 {
		return data
	}

	// 计算总体得分
	totalScore := 0.0
	for _, result := range results {
		totalScore += result.Score.Total
	}
	data.OverallScore = totalScore / float64(len(results))

	// 生成摘要
	data.Summary = r.generateSummary(results)

	// 提取热门关键词
	data.TopKeywords = r.extractTopKeywords(results)

	// 生成全局建议
	data.Recommendations = r.generateGlobalRecommendations(results)

	return data
}

func (r *Reporter) generateSummary(results []models.AnalysisResult) ReportSummary {
	if len(results) == 0 {
		return ReportSummary{}
	}

	// 计算平均分数
	var totalBreakdown models.ScoreBreakdown
	bestScore := 0.0
	worstScore := 100.0
	bestContent := ""
	worstContent := ""

	for _, result := range results {
		totalBreakdown.ContentQuality += result.Score.Breakdown.ContentQuality
		totalBreakdown.Engagement += result.Score.Breakdown.Engagement
		totalBreakdown.Visual += result.Score.Breakdown.Visual
		totalBreakdown.Title += result.Score.Breakdown.Title
		totalBreakdown.Readability += result.Score.Breakdown.Readability
		totalBreakdown.TrendRelevance += result.Score.Breakdown.TrendRelevance

		if result.Score.Total > bestScore {
			bestScore = result.Score.Total
			bestContent = result.Title
		}

		if result.Score.Total < worstScore {
			worstScore = result.Score.Total
			worstContent = result.Title
		}
	}

	count := float64(len(results))
	averageScores := models.ScoreBreakdown{
		ContentQuality: totalBreakdown.ContentQuality / count,
		Engagement:     totalBreakdown.Engagement / count,
		Visual:         totalBreakdown.Visual / count,
		Title:          totalBreakdown.Title / count,
		Readability:    totalBreakdown.Readability / count,
		TrendRelevance: totalBreakdown.TrendRelevance / count,
	}

	// 分析常见问题
	commonIssues := r.findCommonIssues(results)
	successPatterns := r.findSuccessPatterns(results)

	return ReportSummary{
		AverageScores:   averageScores,
		BestPerforming:  bestContent,
		NeedImprovement: worstContent,
		CommonIssues:    commonIssues,
		SuccessPatterns: successPatterns,
	}
}

func (r *Reporter) findCommonIssues(results []models.AnalysisResult) []string {
	issues := make(map[string]int)

	for _, result := range results {
		if result.Score.Breakdown.Title < 60 {
			issues["标题吸引力不足"]++
		}
		if result.Score.Breakdown.Engagement < 60 {
			issues["缺乏互动元素"]++
		}
		if result.Score.Breakdown.Visual < 60 {
			issues["视觉内容质量偏低"]++
		}
		if result.Readability.FleschScore < 50 {
			issues["可读性有待提升"]++
		}
		if len(result.TextAnalysis.CallToAction) == 0 {
			issues["缺少行动召唤"]++
		}
	}

	// 转换为切片并按频率排序
	var commonIssues []string
	for issue, count := range issues {
		if count > len(results)/3 { // 超过1/3的内容有此问题才算常见问题
			commonIssues = append(commonIssues, fmt.Sprintf("%s (%d篇)", issue, count))
		}
	}

	return commonIssues
}

func (r *Reporter) findSuccessPatterns(results []models.AnalysisResult) []string {
	patterns := []string{}

	// 找出高分内容的共同特征
	highScoreResults := []models.AnalysisResult{}
	for _, result := range results {
		if result.Score.Total > 80 {
			highScoreResults = append(highScoreResults, result)
		}
	}

	if len(highScoreResults) == 0 {
		return patterns
	}

	// 分析高分内容的特征
	hasNumberInTitle := 0
	hasQuestionInTitle := 0
	hasGoodIntro := 0
	hasCallToAction := 0

	for _, result := range highScoreResults {
		if result.TextAnalysis.TitleAnalysis.HasNumbers {
			hasNumberInTitle++
		}
		if result.TextAnalysis.TitleAnalysis.HasQuestions {
			hasQuestionInTitle++
		}
		if result.TextAnalysis.ContentStructure.HasIntro {
			hasGoodIntro++
		}
		if len(result.TextAnalysis.CallToAction) > 0 {
			hasCallToAction++
		}
	}

	total := len(highScoreResults)
	if hasNumberInTitle > total/2 {
		patterns = append(patterns, "标题中使用数字")
	}
	if hasQuestionInTitle > total/2 {
		patterns = append(patterns, "标题使用疑问句")
	}
	if hasGoodIntro > total/2 {
		patterns = append(patterns, "有吸引人的开头")
	}
	if hasCallToAction > total/2 {
		patterns = append(patterns, "包含行动召唤")
	}

	return patterns
}

func (r *Reporter) extractTopKeywords(results []models.AnalysisResult) []models.Keyword {
	keywordMap := make(map[string]*models.Keyword)

	for _, result := range results {
		for _, keyword := range result.Keywords {
			if existing, exists := keywordMap[keyword.Word]; exists {
				existing.Frequency += keyword.Frequency
				existing.Relevance = (existing.Relevance + keyword.Relevance) / 2
			} else {
				keywordCopy := keyword
				keywordMap[keyword.Word] = &keywordCopy
			}
		}
	}

	// 转换为切片并排序
	var keywords []models.Keyword
	for _, keyword := range keywordMap {
		keywords = append(keywords, *keyword)
	}

	sort.Slice(keywords, func(i, j int) bool {
		return keywords[i].Frequency > keywords[j].Frequency
	})

	// 返回前20个
	if len(keywords) > 20 {
		keywords = keywords[:20]
	}

	return keywords
}

func (r *Reporter) generateGlobalRecommendations(results []models.AnalysisResult) []GlobalRecommendation {
	recommendations := []GlobalRecommendation{}

	// 统计问题频次
	titleIssues := []string{}
	engagementIssues := []string{}
	visualIssues := []string{}

	for _, result := range results {
		if result.Score.Breakdown.Title < 60 {
			titleIssues = append(titleIssues, result.Title)
		}
		if result.Score.Breakdown.Engagement < 60 {
			engagementIssues = append(engagementIssues, result.Title)
		}
		if len(result.ImageAnalysis) == 0 {
			visualIssues = append(visualIssues, result.Title)
		}
	}

	// 生成全局建议
	if len(titleIssues) > len(results)/3 {
		recommendations = append(recommendations, GlobalRecommendation{
			Category:        "标题优化",
			Priority:        "high",
			Description:     "多篇内容的标题吸引力不足，建议统一优化标题策略",
			AffectedContent: titleIssues,
			ExpectedImpact:  "预计可提升整体点击率20-30%",
		})
	}

	if len(engagementIssues) > len(results)/3 {
		recommendations = append(recommendations, GlobalRecommendation{
			Category:        "互动优化",
			Priority:        "high",
			Description:     "大部分内容缺乏互动元素，建议加强用户参与引导",
			AffectedContent: engagementIssues,
			ExpectedImpact:  "预计可提升用户参与度40-50%",
		})
	}

	if len(visualIssues) > len(results)/2 {
		recommendations = append(recommendations, GlobalRecommendation{
			Category:        "视觉内容",
			Priority:        "medium",
			Description:     "多篇内容缺少视觉元素，建议制定视觉内容策略",
			AffectedContent: visualIssues,
			ExpectedImpact:  "预计可提升内容吸引力30-40%",
		})
	}

	return recommendations
}

func (r *Reporter) generateJSONReport(data ReportData) error {
	filename := filepath.Join(r.config.OutputDir, "analysis_report.json")

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	return encoder.Encode(data)
}

func (r *Reporter) generateHTMLReport(data ReportData) error {
	tmplContent := `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>内容分析报告</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; margin: 0; padding: 20px; background: #f5f7fa; }
        .container { max-width: 1200px; margin: 0 auto; }
        .header { background: white; padding: 30px; border-radius: 10px; margin-bottom: 20px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .score-card { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 30px; border-radius: 10px; margin-bottom: 20px; }
        .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(300px, 1fr)); gap: 20px; }
        .card { background: white; padding: 20px; border-radius: 10px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        .score { font-size: 3em; font-weight: bold; }
        .metric { display: flex; justify-content: space-between; margin: 10px 0; padding: 10px; background: #f8f9fa; border-radius: 5px; }
        .content-list { max-height: 400px; overflow-y: auto; }
        .content-item { padding: 15px; border-bottom: 1px solid #eee; }
        .content-score { float: right; padding: 5px 10px; border-radius: 20px; color: white; }
        .score-excellent { background: #28a745; }
        .score-good { background: #17a2b8; }
        .score-average { background: #ffc107; color: #333; }
        .score-poor { background: #dc3545; }
        .keyword-tag { display: inline-block; background: #e9ecef; padding: 5px 10px; margin: 2px; border-radius: 15px; font-size: 0.9em; }
        .recommendation { padding: 15px; margin: 10px 0; border-left: 4px solid #007bff; background: #f8f9fa; border-radius: 5px; }
        .priority-high { border-left-color: #dc3545; }
        .priority-medium { border-left-color: #ffc107; }
        .priority-low { border-left-color: #28a745; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📊 内容分析报告</h1>
            <p>生成时间: {{.GeneratedAt.Format "2006-01-02 15:04:05"}}</p>
            <p>分析内容数量: {{.TotalContent}} 篇</p>
        </div>

        <div class="score-card">
            <div class="score">{{printf "%.1f" .OverallScore}}</div>
            <h2>总体评分</h2>
            <p>{{if ge .OverallScore 80}}优秀表现！继续保持{{else if ge .OverallScore 60}}良好水平，还有提升空间{{else}}需要重点改进{{end}}</p>
        </div>

        <div class="grid">
            <div class="card">
                <h3>📈 平均得分详情</h3>
                <div class="metric">
                    <span>内容质量</span>
                    <span>{{printf "%.1f" .Summary.AverageScores.ContentQuality}}</span>
                </div>
                <div class="metric">
                    <span>互动潜力</span>
                    <span>{{printf "%.1f" .Summary.AverageScores.Engagement}}</span>
                </div>
                <div class="metric">
                    <span>视觉吸引力</span>
                    <span>{{printf "%.1f" .Summary.AverageScores.Visual}}</span>
                </div>
                <div class="metric">
                    <span>标题质量</span>
                    <span>{{printf "%.1f" .Summary.AverageScores.Title}}</span>
                </div>
                <div class="metric">
                    <span>可读性</span>
                    <span>{{printf "%.1f" .Summary.AverageScores.Readability}}</span>
                </div>
                <div class="metric">
                    <span>趋势相关性</span>
                    <span>{{printf "%.1f" .Summary.AverageScores.TrendRelevance}}</span>
                </div>
            </div>

            <div class="card">
                <h3>🏆 表现概况</h3>
                <p><strong>最佳表现:</strong> {{.Summary.BestPerforming}}</p>
                <p><strong>需要改进:</strong> {{.Summary.NeedImprovement}}</p>
                
                <h4>常见问题:</h4>
                <ul>
                {{range .Summary.CommonIssues}}
                    <li>{{.}}</li>
                {{end}}
                </ul>

                <h4>成功模式:</h4>
                <ul>
                {{range .Summary.SuccessPatterns}}
                    <li>{{.}}</li>
                {{end}}
                </ul>
            </div>
        </div>

        <div class="card">
            <h3>📝 内容详情</h3>
            <div class="content-list">
            {{range .Results}}
                <div class="content-item">
                    <h4>{{.Title}}</h4>
                    <span class="content-score {{if ge .Score.Total 80}}score-excellent{{else if ge .Score.Total 60}}score-good{{else if ge .Score.Total 40}}score-average{{else}}score-poor{{end}}">
                        {{printf "%.1f" .Score.Total}}分
                    </span>
                    <p>{{.Score.Reasoning}}</p>
                </div>
            {{end}}
            </div>
        </div>

        <div class="grid">
            <div class="card">
                <h3>🔥 热门关键词</h3>
                {{range .TopKeywords}}
                    <span class="keyword-tag">{{.Word}} ({{.Frequency}})</span>
                {{end}}
            </div>

            <div class="card">
                <h3>💡 改进建议</h3>
                {{range .Recommendations}}
                    <div class="recommendation priority-{{.Priority}}">
                        <h4>{{.Category}}</h4>
                        <p>{{.Description}}</p>
                        <small>影响内容: {{len .AffectedContent}}篇 | {{.ExpectedImpact}}</small>
                    </div>
                {{end}}
            </div>
        </div>
    </div>
</body>
</html>`

	tmpl, err := template.New("report").Parse(tmplContent)
	if err != nil {
		return err
	}

	filename := filepath.Join(r.config.OutputDir, "analysis_report.html")
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	return tmpl.Execute(file, data)
}

func (r *Reporter) generateCSVReport(data ReportData) error {
	filename := filepath.Join(r.config.OutputDir, "analysis_report.csv")

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// CSV头部
	headers := []string{
		"标题", "总分", "内容质量", "互动性", "视觉效果", "标题质量",
		"可读性", "趋势相关性", "字数", "句子数", "段落数", "关键词数",
		"情感倾向", "阅读时间", "建议数量", "等级",
	}

	// 写入CSV头部
	file.WriteString(strings.Join(headers, ",") + "\n")

	// 写入数据
	for _, result := range data.Results {
		row := []string{
			fmt.Sprintf(`"%s"`, strings.ReplaceAll(result.Title, `"`, `""`)),
			fmt.Sprintf("%.1f", result.Score.Total),
			fmt.Sprintf("%.1f", result.Score.Breakdown.ContentQuality),
			fmt.Sprintf("%.1f", result.Score.Breakdown.Engagement),
			fmt.Sprintf("%.1f", result.Score.Breakdown.Visual),
			fmt.Sprintf("%.1f", result.Score.Breakdown.Title),
			fmt.Sprintf("%.1f", result.Score.Breakdown.Readability),
			fmt.Sprintf("%.1f", result.Score.Breakdown.TrendRelevance),
			fmt.Sprintf("%d", result.TextAnalysis.WordCount),
			fmt.Sprintf("%d", result.TextAnalysis.SentenceCount),
			fmt.Sprintf("%d", result.TextAnalysis.ParagraphCount),
			fmt.Sprintf("%d", len(result.Keywords)),
			result.Sentiment.Overall,
			fmt.Sprintf("%d", result.Readability.ReadingTime),
			fmt.Sprintf("%d", len(result.Suggestions)),
			result.Score.Level,
		}

		file.WriteString(strings.Join(row, ",") + "\n")
	}

	return nil
}
