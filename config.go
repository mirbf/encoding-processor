package encoding

import "time"

// DetectorConfig 检测器配置
type DetectorConfig struct {
	// SampleSize 检测样本大小（字节，默认 8192）
	SampleSize int `json:"sample_size"`

	// MinConfidence 最小置信度（默认 0.6）
	MinConfidence float64 `json:"min_confidence"`

	// SupportedEncodings 支持的编码列表
	SupportedEncodings []string `json:"supported_encodings"`

	// EnableCache 是否启用检测结果缓存
	EnableCache bool `json:"enable_cache"`

	// CacheSize 缓存大小（默认 1000）
	CacheSize int `json:"cache_size"`

	// CacheTTL 缓存过期时间（默认 1 小时）
	CacheTTL time.Duration `json:"cache_ttl"`

	// EnableLanguageDetection 是否启用语言检测
	EnableLanguageDetection bool `json:"enable_language_detection"`

	// PreferredEncodings 优先编码列表（检测时优先考虑）
	PreferredEncodings []string `json:"preferred_encodings"`
}

// ConverterConfig 转换器配置
type ConverterConfig struct {
	// StrictMode 严格模式（遇到无法转换字符时报错）
	StrictMode bool `json:"strict_mode"`

	// InvalidCharReplacement 无效字符替换字符
	InvalidCharReplacement string `json:"invalid_char_replacement"`

	// BufferSize 转换缓冲区大小
	BufferSize int `json:"buffer_size"`

	// MaxMemoryUsage 最大内存使用量（字节，0 表示无限制）
	MaxMemoryUsage int64 `json:"max_memory_usage"`

	// ChunkSize 大文件分块大小（字节，默认 1MB）
	ChunkSize int64 `json:"chunk_size"`

	// PreserveBOM 是否保留 BOM
	PreserveBOM bool `json:"preserve_bom"`

	// NormalizeLineEndings 是否规范化换行符
	NormalizeLineEndings bool `json:"normalize_line_endings"`

	// TargetLineEnding 目标换行符（LF, CRLF, CR）
	TargetLineEnding string `json:"target_line_ending"`
}

// ProcessorConfig 处理器配置（集成配置）
type ProcessorConfig struct {
	// DetectorConfig 检测器配置
	DetectorConfig *DetectorConfig `json:"detector_config"`

	// ConverterConfig 转换器配置
	ConverterConfig *ConverterConfig `json:"converter_config"`

	// EnableMetrics 是否启用性能监控
	EnableMetrics bool `json:"enable_metrics"`

	// LogLevel 日志级别
	LogLevel string `json:"log_level"`

	// Logger 自定义日志记录器
	Logger Logger `json:"-"`

	// TempDir 临时文件目录
	TempDir string `json:"temp_dir"`

	// MaxFileSize 最大文件大小（字节，0 表示无限制）
	MaxFileSize int64 `json:"max_file_size"`
}

// GetDefaultDetectorConfig 获取默认检测器配置
func GetDefaultDetectorConfig() *DetectorConfig {
	return &DetectorConfig{
		SampleSize:              DefaultSampleSize,
		MinConfidence:           DefaultMinConfidence,
		EnableCache:             true,
		CacheSize:               DefaultCacheSize,
		CacheTTL:                DefaultCacheTTL,
		EnableLanguageDetection: false,
		SupportedEncodings: []string{
			EncodingUTF8,
			EncodingUTF16,
			EncodingUTF16LE,
			EncodingUTF16BE,
			EncodingGBK,
			EncodingGB2312,
			EncodingGB18030,
			EncodingBIG5,
			EncodingShiftJIS,
			EncodingEUCJP,
			EncodingEUCKR,
			EncodingISO88591,
			EncodingWindows1252,
		},
		PreferredEncodings: []string{
			EncodingUTF8,
			EncodingGBK,
			EncodingBIG5,
		},
	}
}

// GetDefaultConverterConfig 获取默认转换器配置
func GetDefaultConverterConfig() *ConverterConfig {
	return &ConverterConfig{
		StrictMode:             false,
		InvalidCharReplacement: DefaultInvalidChar,
		BufferSize:             DefaultBufferSize,
		MaxMemoryUsage:         0, // 无限制
		ChunkSize:              DefaultChunkSize,
		PreserveBOM:            false,
		NormalizeLineEndings:   false,
		TargetLineEnding:       LineEndingLF,
	}
}

// GetDefaultProcessorConfig 获取默认处理器配置
func GetDefaultProcessorConfig() *ProcessorConfig {
	return &ProcessorConfig{
		DetectorConfig:  GetDefaultDetectorConfig(),
		ConverterConfig: GetDefaultConverterConfig(),
		EnableMetrics:   true,
		LogLevel:        "info",
		Logger:          nil, // 使用默认日志记录器
		TempDir:         "",  // 使用系统临时目录
		MaxFileSize:     DefaultMaxFileSize,
	}
}