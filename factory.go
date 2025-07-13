package encoding

import "log"

// 工厂函数

// NewDefault 创建默认处理器
func NewDefault() Processor {
	return NewProcessor(GetDefaultProcessorConfig())
}

// NewDefaultWithMetrics 创建带性能监控的默认处理器
func NewDefaultWithMetrics() (Processor, MetricsCollector) {
	config := GetDefaultProcessorConfig()
	config.EnableMetrics = true
	
	processor := NewProcessor(config)
	metrics := NewMetricsCollector()
	
	return processor, metrics
}

// NewDefaultStream 创建默认流处理器
func NewDefaultStream() StreamProcessor {
	return NewStreamProcessor(GetDefaultProcessorConfig())
}

// NewDefaultFile 创建默认文件处理器
func NewDefaultFile() FileProcessor {
	return NewFileProcessor(GetDefaultProcessorConfig())
}

// NewWithLogger 创建带自定义日志的处理器
func NewWithLogger(logger Logger) Processor {
	config := GetDefaultProcessorConfig()
	config.Logger = logger
	return NewProcessor(config)
}

// NewWithConfig 使用自定义配置创建处理器
func NewWithConfig(detectorConfig *DetectorConfig, converterConfig *ConverterConfig) Processor {
	config := &ProcessorConfig{
		DetectorConfig:  detectorConfig,
		ConverterConfig: converterConfig,
		EnableMetrics:   true,
		LogLevel:        "info",
		Logger:          nil,
		TempDir:         "",
		MaxFileSize:     DefaultMaxFileSize,
	}
	return NewProcessor(config)
}

// 便捷创建函数

// NewQuick 快速创建处理器（最少配置）
func NewQuick() Processor {
	return NewDefault()
}

// NewForCLI 创建适合命令行工具的处理器
func NewForCLI() Processor {
	config := GetDefaultProcessorConfig()
	
	// 命令行工具通常需要更详细的检测
	config.DetectorConfig.SampleSize = 16384
	config.DetectorConfig.MinConfidence = 0.7
	config.DetectorConfig.EnableCache = false // 命令行工具通常不需要缓存
	
	// 更宽松的转换配置
	config.ConverterConfig.StrictMode = false
	config.ConverterConfig.BufferSize = 32768
	
	return NewProcessor(config)
}

// NewForWebService 创建适合 Web 服务的处理器
func NewForWebService() Processor {
	config := GetDefaultProcessorConfig()
	
	// Web 服务需要更快的响应
	config.DetectorConfig.SampleSize = 4096
	config.DetectorConfig.EnableCache = true
	config.DetectorConfig.CacheSize = 5000
	
	// 启用性能监控
	config.EnableMetrics = true
	
	return NewProcessor(config)
}

// NewForBatchProcessing 创建适合批量处理的处理器
func NewForBatchProcessing() Processor {
	config := GetDefaultProcessorConfig()
	
	// 批量处理可以使用更大的缓冲区
	config.ConverterConfig.BufferSize = 65536
	config.ConverterConfig.ChunkSize = 2 * 1024 * 1024 // 2MB
	
	// 更大的检测样本
	config.DetectorConfig.SampleSize = 32768
	
	// 启用缓存以提高重复文件的处理速度
	config.DetectorConfig.EnableCache = true
	config.DetectorConfig.CacheSize = 10000
	
	return NewProcessor(config)
}

// 高级工厂函数

// NewHighPerformance 创建高性能处理器
func NewHighPerformance() Processor {
	config := GetDefaultProcessorConfig()
	
	// 高性能配置
	config.DetectorConfig.SampleSize = 65536
	config.DetectorConfig.EnableCache = true
	config.DetectorConfig.CacheSize = 20000
	
	config.ConverterConfig.BufferSize = 131072 // 128KB
	config.ConverterConfig.ChunkSize = 4 * 1024 * 1024 // 4MB
	config.ConverterConfig.MaxMemoryUsage = 100 * 1024 * 1024 // 100MB
	
	config.EnableMetrics = true
	config.MaxFileSize = 1024 * 1024 * 1024 // 1GB
	
	return NewProcessor(config)
}

// NewMemoryEfficient 创建内存高效的处理器
func NewMemoryEfficient() Processor {
	config := GetDefaultProcessorConfig()
	
	// 内存高效配置
	config.DetectorConfig.SampleSize = 2048
	config.DetectorConfig.EnableCache = false // 禁用缓存以节省内存
	
	config.ConverterConfig.BufferSize = 4096
	config.ConverterConfig.ChunkSize = 256 * 1024 // 256KB
	config.ConverterConfig.MaxMemoryUsage = 10 * 1024 * 1024 // 10MB
	
	config.MaxFileSize = 50 * 1024 * 1024 // 50MB
	
	return NewProcessor(config)
}

// NewStrictMode 创建严格模式处理器（遇到错误立即失败）
func NewStrictMode() Processor {
	config := GetDefaultProcessorConfig()
	
	// 严格模式配置
	config.DetectorConfig.MinConfidence = 0.9
	config.ConverterConfig.StrictMode = true
	
	return NewProcessor(config)
}

// NewTolerantMode 创建容错模式处理器（尽量处理，忽略错误）
func NewTolerantMode() Processor {
	config := GetDefaultProcessorConfig()
	
	// 容错模式配置
	config.DetectorConfig.MinConfidence = 0.5
	config.ConverterConfig.StrictMode = false
	config.ConverterConfig.InvalidCharReplacement = "?"
	
	return NewProcessor(config)
}

// 默认日志记录器实现
type defaultLogger struct{}

func (l *defaultLogger) Debug(msg string, fields ...interface{}) {
	log.Printf("[DEBUG] "+msg, fields...)
}

func (l *defaultLogger) Info(msg string, fields ...interface{}) {
	log.Printf("[INFO] "+msg, fields...)
}

func (l *defaultLogger) Warn(msg string, fields ...interface{}) {
	log.Printf("[WARN] "+msg, fields...)
}

func (l *defaultLogger) Error(msg string, fields ...interface{}) {
	log.Printf("[ERROR] "+msg, fields...)
}

// getDefaultLogger 获取默认日志记录器
func getDefaultLogger() Logger {
	return &defaultLogger{}
}