package encoding

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/saintfish/chardet"
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