// internal/services/image_service.go
package services

import (
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/content-analyzer/internal/config"
	"github.com/content-analyzer/internal/models"
)

type ImageService interface {
	AnalyzeImage(imagePath string) (models.ImageAnalysis, error)
	ValidateImage(imagePath string) error
	GetImageInfo(imagePath string) (models.Image, error)
	BatchAnalyze(imagePaths []string) ([]models.ImageAnalysis, error)
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

	// 加载图片数据进行分析
	img, err := s.loadImage(imagePath)
	if err != nil {
		return models.ImageAnalysis{}, fmt.Errorf("加载图片失败: %w", err)
	}

	// 分析图片
	analysis := models.ImageAnalysis{
		Path:                imagePath,
		VisualElements:      s.analyzeVisualElements(img, imgInfo),
		CompositionAnalysis: s.analyzeComposition(img, imgInfo),
		QualityMetrics:      s.analyzeQuality(img, imgInfo),
		StyleAnalysis:       s.analyzeStyle(img, imgInfo),
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

func (s *imageService) BatchAnalyze(imagePaths []string) ([]models.ImageAnalysis, error) {
	var analyses []models.ImageAnalysis

	for _, path := range imagePaths {
		analysis, err := s.AnalyzeImage(path)
		if err != nil {
			return nil, fmt.Errorf("分析图片 %s 失败: %w", path, err)
		}
		analyses = append(analyses, analysis)
	}

	return analyses, nil
}

func (s *imageService) loadImage(imagePath string) (image.Image, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	return img, nil
}

func (s *imageService) analyzeVisualElements(img image.Image, imgInfo models.Image) models.VisualElements {
	bounds := img.Bounds()
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y

	// 分析主要颜色
	dominantColors := s.extractDominantColors(img)

	// 分析亮度、对比度、饱和度
	brightness, contrast, saturation := s.analyzeColorMetrics(img)

	// 检测对象和特征
	hasText := s.detectText(img)
	hasFaces := s.detectFaces(img) 
	objectCount := s.countObjects(img)

	return models.VisualElements{
		DominantColors: dominantColors,
		Brightness:     brightness,
		Contrast:       contrast,
		Saturation:     saturation,
		HasText:        hasText,
		HasFaces:       hasFaces,
		ObjectCount:    objectCount,
	}
}

func (s *imageService) analyzeComposition(img image.Image, imgInfo models.Image) models.CompositionAnalysis {
	bounds := img.Bounds()

	return models.CompositionAnalysis{
		RuleOfThirds:  s.checkRuleOfThirds(img),
		Symmetry:      s.checkSymmetry(img),
		LeadingLines:  s.detectLeadingLines(img),
		FramingScore:  s.calculateFramingScore(img),
		BalanceScore:  s.calculateBalanceScore(img),
		FocusClarity:  s.calculateFocusClarity(img),
	}
}

func (s *imageService) analyzeQuality(img image.Image, imgInfo models.Image) models.QualityMetrics {
	resolution := fmt.Sprintf("%dx%d", imgInfo.Width, imgInfo.Height)

	// 基于分辨率判断质量
	totalPixels := imgInfo.Width * imgInfo.Height
	var resolutionScore float64

	if totalPixels >= 2000000 { // 2MP以上
		resolutionScore = 0.9
	} else if totalPixels >= 1000000 { // 1MP以上
		resolutionScore = 0.7
	} else if totalPixels >= 500000 { // 0.5MP以上
		resolutionScore = 0.5
	} else {
		resolutionScore = 0.3
	}

	// 计算清晰度
	sharpness := s.calculateSharpness(img)

	// 计算噪点水平
	noiseLevel := s.calculateNoiseLevel(img)

	// 计算曝光评分
	exposureScore := s.calculateExposureScore(img)

	// 综合质量评分
	overallQuality := (resolutionScore*0.3 + sharpness*0.3 + (1-noiseLevel)*0.2 + exposureScore*0.2)

	return models.QualityMetrics{
		Resolution:     resolution,
		Sharpness:      sharpness,
		NoiseLevel:     noiseLevel,
		ExposureScore:  exposureScore,
		OverallQuality: overallQuality,
	}
}

func (s *imageService) analyzeStyle(img image.Image, imgInfo models.Image) models.StyleAnalysis {
	style := s.determineStyle(img, imgInfo)
	mood := s.determineMood(img)
	filter := s.detectFilter(img)
	consistency := s.calculateConsistency(img)

	return models.StyleAnalysis{
		Style:       style,
		Mood:        mood,
		Filter:      filter,
		Consistency: consistency,
	}
}

// 颜色分析相关方法
func (s *imageService) extractDominantColors(img image.Image) []string {
	colorMap := make(map[string]int)
	bounds := img.Bounds()
	
	// 采样分析（每10个像素采样一次以提高性能）
	for y := bounds.Min.Y; y < bounds.Max.Y; y += 10 {
		for x := bounds.Min.X; x < bounds.Max.X; x += 10 {
			r, g, b, _ := img.At(x, y).RGBA()
			// 转换为8位颜色
			r8 := uint8(r >> 8)
			g8 := uint8(g >> 8)
			b8 := uint8(b >> 8)
			
			// 量化颜色以减少计算复杂度
			colorKey := fmt.Sprintf("#%02X%02X%02X", r8&0xF0, g8&0xF0, b8&0xF0)
			colorMap[colorKey]++
		}
	}
	
	// 找出出现频率最高的颜色
	type colorFreq struct {
		color string
		freq  int
	}
	
	var colors []colorFreq
	for color, freq := range colorMap {
		colors = append(colors, colorFreq{color, freq})
	}
	
	// 排序并返回前5个主要颜色
	if len(colors) > 5 {
		// 简单冒泡排序，取前5个
		for i := 0; i < 5; i++ {
			for j := i + 1; j < len(colors); j++ {
				if colors[j].freq > colors[i].freq {
					colors[i], colors[j] = colors[j], colors[i]
				}
			}
		}
		colors = colors[:5]
	}
	
	var dominantColors []string
	for _, c := range colors {
		dominantColors = append(dominantColors, c.color)
	}
	
	return dominantColors
}

func (s *imageService) analyzeColorMetrics(img image.Image) (brightness, contrast, saturation float64) {
	bounds := img.Bounds()
	var totalR, totalG, totalB float64
	var minLum, maxLum float64 = 1.0, 0.0
	var totalSat float64
	pixelCount := 0
	
	// 采样分析
	for y := bounds.Min.Y; y < bounds.Max.Y; y += 5 {
		for x := bounds.Min.X; x < bounds.Max.X; x += 5 {
			r, g, b, _ := img.At(x, y).RGBA()
			
			// 转换为0-1范围
			rf := float64(r) / 65535.0
			gf := float64(g) / 65535.0
			bf := float64(b) / 65535.0
			
			totalR += rf
			totalG += gf
			totalB += bf
			
			// 计算亮度 (相对亮度公式)
			luminance := 0.299*rf + 0.587*gf + 0.114*bf
			if luminance > maxL
