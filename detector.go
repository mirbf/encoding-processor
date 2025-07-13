package encoding

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"regexp"
	"sort"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/saintfish/chardet"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/transform"
)

// defaultDetector 实现 Detector 接口
type defaultDetector struct {
	config *DetectorConfig
	cache  *detectionCache
	mutex  sync.RWMutex
}

// detectionCache 检测结果缓存
type detectionCache struct {
	cache map[string]*cacheEntry
	mutex sync.RWMutex
}

type cacheEntry struct {
	result    *DetectionResult
	timestamp time.Time
}

// DetectionCandidate 检测候选结果
type DetectionCandidate struct {
	Encoding      string
	Confidence    float64
	Method        string
	ConvertedText string
	Score         float64
}

// NewDetector 创建新的检测器
func NewDetector(config ...*DetectorConfig) Detector {
	var cfg *DetectorConfig
	if len(config) > 0 && config[0] != nil {
		cfg = config[0]
	} else {
		cfg = GetDefaultDetectorConfig()
	}

	detector := &defaultDetector{
		config: cfg,
	}

	if cfg.EnableCache {
		detector.cache = &detectionCache{
			cache: make(map[string]*cacheEntry),
		}
	}

	return detector
}

// SmartDetectEncoding 智能编码检测
func (d *defaultDetector) SmartDetectEncoding(data []byte) (*DetectionResult, error) {
	if len(data) == 0 {
		return nil, &EncodingError{
			Op:  OperationDetect,
			Err: ErrInvalidInput,
		}
	}

	// 1. 使用传统方法检测
	traditionalResult, _ := d.DetectEncoding(data)
	
	// 2. 获取所有候选编码
	candidates := d.getAllCandidates(data)
	
	// 3. 对候选编码进行评分
	scoredCandidates := d.scoreCandidates(data, candidates)
	
	// 4. 选择最佳结果
	bestCandidate := d.selectBestCandidate(scoredCandidates, traditionalResult)
	
	if bestCandidate == nil {
		return nil, &EncodingError{
			Op:       OperationDetect,
			Encoding: "unknown",
			Err:      ErrDetectionFailed,
		}
	}

	return &DetectionResult{
		Encoding:   bestCandidate.Encoding,
		Confidence: bestCandidate.Confidence,
		Details: map[string]interface{}{
			"method": bestCandidate.Method,
			"score": bestCandidate.Score,
			"converted_text": bestCandidate.ConvertedText,
		},
	}, nil
}

// DetectEncoding 检测数据的编码格式
func (d *defaultDetector) DetectEncoding(data []byte) (*DetectionResult, error) {
	if len(data) == 0 {
		return nil, &EncodingError{
			Op:  OperationDetect,
			Err: ErrInvalidInput,
		}
	}

	// 检查缓存
	if d.cache != nil {
		if cached := d.getCachedResult(data); cached != nil {
			return cached, nil
		}
	}

	// 限制检测样本大小
	sampleSize := d.config.SampleSize
	if len(data) > sampleSize {
		data = data[:sampleSize]
	}

	// 首先尝试检测 BOM
	if bomResult := d.detectBOM(data); bomResult != nil {
		d.cacheResult(data, bomResult)
		return bomResult, nil
	}

	// 检查是否是有效的 UTF-8
	if utf8Result := d.detectUTF8(data); utf8Result != nil {
		d.cacheResult(data, utf8Result)
		return utf8Result, nil
	}

	// 使用 chardet 进行检测
	detector := chardet.NewTextDetector()
	results, err := detector.DetectAll(data)
	if err != nil {
		return nil, &EncodingError{
			Op:       OperationDetect,
			Encoding: "unknown",
			Err:      fmt.Errorf("chardet detection failed: %w", err),
		}
	}

	if len(results) == 0 {
		return nil, &EncodingError{
			Op:       OperationDetect,
			Encoding: "unknown",
			Err:      ErrDetectionFailed,
		}
	}

	// 选择最佳结果
	bestResult := d.selectBestResult(results)
	if bestResult == nil {
		return nil, &EncodingError{
			Op:       OperationDetect,
			Encoding: "unknown",
			Err:      ErrDetectionFailed,
		}
	}

	// 验证编码是否支持
	if !d.isEncodingSupported(bestResult.Encoding) {
		return nil, &EncodingError{
			Op:       OperationDetect,
			Encoding: bestResult.Encoding,
			Err:      ErrUnsupportedEncoding,
		}
	}

	// 检查置信度
	if bestResult.Confidence < d.config.MinConfidence {
		return nil, &EncodingError{
			Op:       OperationDetect,
			Encoding: bestResult.Encoding,
			Err:      fmt.Errorf("confidence too low: %.2f < %.2f", bestResult.Confidence, d.config.MinConfidence),
		}
	}

	// 缓存结果
	d.cacheResult(data, bestResult)

	return bestResult, nil
}

// DetectFileEncoding 检测文件的编码格式
func (d *defaultDetector) DetectFileEncoding(filename string) (*DetectionResult, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, &FileOperationError{
			Op:   OperationDetect,
			File: filename,
			Err:  err,
		}
	}

	result, err := d.DetectEncoding(data)
	if err != nil {
		if encErr, ok := err.(*EncodingError); ok {
			encErr.File = filename
		}
		return nil, err
	}

	return result, nil
}

// DetectBestEncoding 检测最可能的编码格式（简化版本）
func (d *defaultDetector) DetectBestEncoding(data []byte) (string, error) {
	result, err := d.DetectEncoding(data)
	if err != nil {
		return "", err
	}
	return result.Encoding, nil
}

// detectUTF8 检测是否是有效的 UTF-8
func (d *defaultDetector) detectUTF8(data []byte) *DetectionResult {
	if len(data) == 0 {
		return nil
	}

	// 检查是否是有效的 UTF-8
	if utf8.Valid(data) {
		// 计算置信度
		confidence := 0.95 // 基础置信度

		// 如果包含非 ASCII 字符，提高置信度
		hasNonASCII := false
		for _, b := range data {
			if b > 127 {
				hasNonASCII = true
				break
			}
		}

		if hasNonASCII {
			confidence = 0.99
		} else {
			// 对于纯 ASCII 文本，置信度稍低
			confidence = 0.85
		}

		return &DetectionResult{
			Encoding:   EncodingUTF8,
			Confidence: confidence,
			Details: map[string]interface{}{
				"method": "utf8_validation",
				"has_non_ascii": hasNonASCII,
			},
		}
	}

	return nil
}

// detectBOM 检测字节顺序标记
func (d *defaultDetector) detectBOM(data []byte) *DetectionResult {
	if len(data) < 2 {
		return nil
	}

	// UTF-8 BOM: EF BB BF
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		return &DetectionResult{
			Encoding:   EncodingUTF8,
			Confidence: 1.0,
			Details: map[string]interface{}{
				"bom": true,
				"method": "bom_detection",
			},
		}
	}

	// UTF-16 LE BOM: FF FE
	if data[0] == 0xFF && data[1] == 0xFE {
		// 检查是否是 UTF-32 LE
		if len(data) >= 4 && data[2] == 0x00 && data[3] == 0x00 {
			return &DetectionResult{
				Encoding:   EncodingUTF32LE,
				Confidence: 1.0,
				Details: map[string]interface{}{
					"bom": true,
					"method": "bom_detection",
				},
			}
		}
		return &DetectionResult{
			Encoding:   EncodingUTF16LE,
			Confidence: 1.0,
			Details: map[string]interface{}{
				"bom": true,
				"method": "bom_detection",
			},
		}
	}

	// UTF-16 BE BOM: FE FF
	if data[0] == 0xFE && data[1] == 0xFF {
		return &DetectionResult{
			Encoding:   EncodingUTF16BE,
			Confidence: 1.0,
			Details: map[string]interface{}{
				"bom": true,
				"method": "bom_detection",
			},
		}
	}

	// UTF-32 BE BOM: 00 00 FE FF
	if len(data) >= 4 && data[0] == 0x00 && data[1] == 0x00 && data[2] == 0xFE && data[3] == 0xFF {
		return &DetectionResult{
			Encoding:   EncodingUTF32BE,
			Confidence: 1.0,
			Details: map[string]interface{}{
				"bom": true,
				"method": "bom_detection",
			},
		}
	}

	return nil
}

// selectBestResult 选择最佳检测结果
func (d *defaultDetector) selectBestResult(results []chardet.Result) *DetectionResult {
	if len(results) == 0 {
		return nil
	}

	// 首先检查优先编码列表
	for _, preferred := range d.config.PreferredEncodings {
		for _, result := range results {
			if d.normalizeEncodingName(result.Charset) == preferred {
				return &DetectionResult{
					Encoding:   preferred,
					Confidence: float64(result.Confidence) / 100.0,
					Language:   result.Language,
					Details: map[string]interface{}{
						"method": "chardet",
						"charset": result.Charset,
					},
				}
			}
		}
	}

	// 选择置信度最高的结果
	best := results[0]
	encoding := d.normalizeEncodingName(best.Charset)

	return &DetectionResult{
		Encoding:   encoding,
		Confidence: float64(best.Confidence) / 100.0,
		Language:   best.Language,
		Details: map[string]interface{}{
			"method": "chardet",
			"charset": best.Charset,
		},
	}
}

// normalizeEncodingName 规范化编码名称
func (d *defaultDetector) normalizeEncodingName(charset string) string {
	// 映射 chardet 的编码名称到我们的标准名称
	mapping := map[string]string{
		"UTF-8":        EncodingUTF8,
		"UTF-16":       EncodingUTF16,
		"UTF-16LE":     EncodingUTF16LE,
		"UTF-16BE":     EncodingUTF16BE,
		"UTF-32":       EncodingUTF32,
		"UTF-32LE":     EncodingUTF32LE,
		"UTF-32BE":     EncodingUTF32BE,
		"GB2312":       EncodingGBK, // 将 GB2312 映射为 GBK
		"GBK":          EncodingGBK,
		"GB18030":      EncodingGB18030,
		"Big5":         EncodingBIG5,
		"Shift_JIS":    EncodingShiftJIS,
		"EUC-JP":       EncodingEUCJP,
		"EUC-KR":       EncodingEUCKR,
		"ISO-8859-1":   EncodingISO88591,
		"windows-1252": EncodingWindows1252,
		"KOI8-R":       EncodingKOI8R,
	}

	if normalized, exists := mapping[charset]; exists {
		return normalized
	}

	return charset
}

// isEncodingSupported 检查编码是否在支持列表中
func (d *defaultDetector) isEncodingSupported(encoding string) bool {
	if len(d.config.SupportedEncodings) == 0 {
		return true // 如果没有限制，支持所有编码
	}

	for _, supported := range d.config.SupportedEncodings {
		if encoding == supported {
			return true
		}
	}
	return false
}

// getCachedResult 获取缓存的检测结果
func (d *defaultDetector) getCachedResult(data []byte) *DetectionResult {
	if d.cache == nil {
		return nil
	}

	key := d.generateCacheKey(data)
	d.cache.mutex.RLock()
	defer d.cache.mutex.RUnlock()

	entry, exists := d.cache.cache[key]
	if !exists {
		return nil
	}

	// 检查是否过期
	if time.Since(entry.timestamp) > d.config.CacheTTL {
		// 异步删除过期项
		go d.removeExpiredCacheEntry(key)
		return nil
	}

	return entry.result
}

// cacheResult 缓存检测结果
func (d *defaultDetector) cacheResult(data []byte, result *DetectionResult) {
	if d.cache == nil {
		return
	}

	key := d.generateCacheKey(data)
	d.cache.mutex.Lock()
	defer d.cache.mutex.Unlock()

	// 如果缓存已满，删除最旧的条目
	if len(d.cache.cache) >= d.config.CacheSize {
		d.evictOldestEntry()
	}

	d.cache.cache[key] = &cacheEntry{
		result:    result,
		timestamp: time.Now(),
	}
}

// generateCacheKey 生成缓存键
func (d *defaultDetector) generateCacheKey(data []byte) string {
	// 使用数据的哈希值作为缓存键
	hash := sha256.Sum256(data)
	return fmt.Sprintf("%x", hash)
}

// removeExpiredCacheEntry 删除过期的缓存项
func (d *defaultDetector) removeExpiredCacheEntry(key string) {
	d.cache.mutex.Lock()
	defer d.cache.mutex.Unlock()
	delete(d.cache.cache, key)
}

// evictOldestEntry 删除最旧的缓存项
func (d *defaultDetector) evictOldestEntry() {
	var oldestKey string
	var oldestTime time.Time

	for key, entry := range d.cache.cache {
		if oldestKey == "" || entry.timestamp.Before(oldestTime) {
			oldestKey = key
			oldestTime = entry.timestamp
		}
	}

	if oldestKey != "" {
		delete(d.cache.cache, oldestKey)
	}
}

// getAllCandidates 获取所有候选编码
func (d *defaultDetector) getAllCandidates(data []byte) []*DetectionCandidate {
	var candidates []*DetectionCandidate
	
	// 1. chardet检测结果
	detector := chardet.NewTextDetector()
	if results, err := detector.DetectAll(data); err == nil {
		for _, result := range results {
			encoding := d.normalizeEncodingName(result.Charset)
			candidates = append(candidates, &DetectionCandidate{
				Encoding:   encoding,
				Confidence: float64(result.Confidence) / 100.0,
				Method:     "chardet",
			})
		}
	}
	
	// 2. 为中文编码增加额外候选
	if d.containsChineseBytes(data) {
		chineseEncodings := []string{EncodingGBK, EncodingGB18030, EncodingBIG5}
		for _, enc := range chineseEncodings {
			found := false
			for _, candidate := range candidates {
				if candidate.Encoding == enc {
					found = true
					break
				}
			}
			if !found {
				candidates = append(candidates, &DetectionCandidate{
					Encoding:   enc,
					Confidence: 0.05, // 低置信度候选
					Method:     "chinese_heuristic",
				})
			}
		}
	}
	
	return candidates
}

// scoreCandidates 对候选编码进行评分
func (d *defaultDetector) scoreCandidates(data []byte, candidates []*DetectionCandidate) []*DetectionCandidate {
	for _, candidate := range candidates {
		// 尝试转换为UTF-8
		convertedText := d.tryConvert(data, candidate.Encoding)
		candidate.ConvertedText = convertedText
		
		// 计算综合得分
		score := d.calculateScore(data, candidate, convertedText)
		candidate.Score = score
	}
	
	// 按得分排序
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})
	
	return candidates
}

// calculateScore 计算候选编码的综合得分
func (d *defaultDetector) calculateScore(data []byte, candidate *DetectionCandidate, convertedText string) float64 {
	score := candidate.Confidence * 0.4 // 基础置信度权重40%
	
	if convertedText == "" {
		return score * 0.1 // 转换失败大幅降低得分
	}
	
	// 1. 检查是否包含有效的中文字符
	chineseScore := d.scoreChineseCharacters(convertedText)
	score += chineseScore * 0.3 // 中文字符得分权重30%
	
	// 2. 检查字符合理性
	validityScore := d.scoreCharacterValidity(convertedText)
	score += validityScore * 0.2 // 字符有效性权重20%
	
	// 3. 检查是否有乱码特征
	garbledScore := d.scoreGarbledText(convertedText)
	score += garbledScore * 0.1 // 乱码检测权重10%
	
	return score
}

// scoreChineseCharacters 评分中文字符质量
func (d *defaultDetector) scoreChineseCharacters(text string) float64 {
	if text == "" {
		return 0
	}
	
	totalRunes := 0
	chineseRunes := 0
	commonChineseRunes := 0
	
	// 常见中文字符范围
	commonChineseChars := map[rune]bool{
		'的': true, '一': true, '是': true, '在': true, '不': true,
		'了': true, '有': true, '和': true, '人': true, '这': true,
		'中': true, '大': true, '为': true, '上': true, '个': true,
		'文': true, '件': true, '作': true, '者': true, '时': true,
	}
	
	for _, r := range text {
		totalRunes++
		if r >= 0x4e00 && r <= 0x9fff {
			chineseRunes++
			if commonChineseChars[r] {
				commonChineseRunes++
			}
		}
	}
	
	if totalRunes == 0 {
		return 0
	}
	
	chineseRatio := float64(chineseRunes) / float64(totalRunes)
	commonRatio := float64(commonChineseRunes) / float64(totalRunes)
	
	return chineseRatio*0.7 + commonRatio*0.3
}

// scoreCharacterValidity 评分字符有效性
func (d *defaultDetector) scoreCharacterValidity(text string) float64 {
	if text == "" {
		return 0
	}
	
	validChars := 0
	totalChars := 0
	
	for _, r := range text {
		totalChars++
		
		// 检查是否是有效字符
		if d.isValidCharacter(r) {
			validChars++
		}
	}
	
	if totalChars == 0 {
		return 0
	}
	
	return float64(validChars) / float64(totalChars)
}

// isValidCharacter 检查字符是否有效
func (d *defaultDetector) isValidCharacter(r rune) bool {
	// ASCII字符
	if r >= 32 && r <= 126 {
		return true
	}
	
	// 中文字符
	if r >= 0x4e00 && r <= 0x9fff {
		return true
	}
	
	// 中文标点符号
	if (r >= 0x3000 && r <= 0x303f) || // CJK符号和标点
		(r >= 0xff00 && r <= 0xffef) {  // 全角ASCII
		return true
	}
	
	// 控制字符（换行等）
	if r == '\n' || r == '\r' || r == '\t' {
		return true
	}
	
	return false
}

// scoreGarbledText 检测乱码特征
func (d *defaultDetector) scoreGarbledText(text string) float64 {
	if text == "" {
		return 0
	}
	
	// 乱码特征检测
	garbledPatterns := []*regexp.Regexp{
		regexp.MustCompile(`[��]+`),           // 替换字符
		regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]+`), // 控制字符
		regexp.MustCompile(`[ÿþ]+`),          // 常见乱码字符
	}
	
	garbledCount := 0
	for _, pattern := range garbledPatterns {
		if pattern.MatchString(text) {
			garbledCount++
		}
	}
	
	// 返回0-1之间的得分，乱码越少得分越高
	return 1.0 - float64(garbledCount)/float64(len(garbledPatterns))
}

// selectBestCandidate 选择最佳候选结果
func (d *defaultDetector) selectBestCandidate(candidates []*DetectionCandidate, traditionalResult *DetectionResult) *DetectionCandidate {
	if len(candidates) == 0 {
		return nil
	}
	
	// 如果传统方法有高置信度结果，优先考虑
	if traditionalResult != nil && traditionalResult.Confidence >= 0.8 {
		for _, candidate := range candidates {
			if candidate.Encoding == traditionalResult.Encoding {
				candidate.Score += 0.2 // 加分
				break
			}
		}
		
		// 重新排序
		sort.Slice(candidates, func(i, j int) bool {
			return candidates[i].Score > candidates[j].Score
		})
	}
	
	return candidates[0]
}

// containsChineseBytes 检查是否包含中文字节特征
func (d *defaultDetector) containsChineseBytes(data []byte) bool {
	chineseByteCount := 0
	for _, b := range data {
		// GBK/GB2312: A1-FE
		// BIG5: A1-FE
		if b >= 0xA1 && b <= 0xFE {
			chineseByteCount++
		}
	}
	
	// 如果超过30%的字节在中文范围内
	return float64(chineseByteCount)/float64(len(data)) > 0.3
}

// tryConvert 尝试转换编码
func (d *defaultDetector) tryConvert(data []byte, encoding string) string {
	var decoder transform.Transformer
	
	switch encoding {
	case EncodingGBK, "GB2312":
		decoder = simplifiedchinese.GBK.NewDecoder()
	case EncodingGB18030:
		decoder = simplifiedchinese.GB18030.NewDecoder()
	case EncodingBIG5:
		decoder = traditionalchinese.Big5.NewDecoder()
	default:
		return ""
	}
	
	if decoder == nil {
		return ""
	}
	
	result, _, err := transform.Bytes(decoder, data)
	if err != nil {
		return ""
	}
	
	// 检查结果是否是有效的UTF-8
	if !utf8.Valid(result) {
		return ""
	}
	
	return string(result)
}