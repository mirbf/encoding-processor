package encoding

import "time"

// DetectionResult 编码检测结果结构
type DetectionResult struct {
	// Encoding 检测到的编码名称
	Encoding string `json:"encoding"`

	// Confidence 检测置信度 (0.0-1.0)
	Confidence float64 `json:"confidence"`

	// Language 检测到的语言（可选）
	Language string `json:"language,omitempty"`

	// Details 详细信息（可选）
	Details map[string]interface{} `json:"details,omitempty"`
}

// ConvertResult 编码转换结果结构
type ConvertResult struct {
	// Data 转换后的数据
	Data []byte `json:"-"`

	// SourceEncoding 源编码
	SourceEncoding string `json:"source_encoding"`

	// TargetEncoding 目标编码
	TargetEncoding string `json:"target_encoding"`

	// BytesProcessed 处理的字节数
	BytesProcessed int64 `json:"bytes_processed"`

	// ConversionTime 转换耗时
	ConversionTime time.Duration `json:"conversion_time"`
}

// StringConvertResult 字符串转换结果
type StringConvertResult struct {
	// Text 转换后的字符串
	Text string `json:"text"`

	// SourceEncoding 源编码
	SourceEncoding string `json:"source_encoding"`

	// TargetEncoding 目标编码
	TargetEncoding string `json:"target_encoding"`

	// BytesProcessed 处理的字节数
	BytesProcessed int64 `json:"bytes_processed"`

	// ConversionTime 转换耗时
	ConversionTime time.Duration `json:"conversion_time"`
}

// StreamOptions 流处理选项
type StreamOptions struct {
	// SourceEncoding 源编码（空值表示自动检测）
	SourceEncoding string `json:"source_encoding"`

	// TargetEncoding 目标编码（默认 UTF-8）
	TargetEncoding string `json:"target_encoding"`

	// BufferSize 缓冲区大小（默认 8192）
	BufferSize int `json:"buffer_size"`

	// DetectionSampleSize 编码检测样本大小（默认 8192）
	DetectionSampleSize int `json:"detection_sample_size"`

	// SkipBOM 是否跳过 BOM（默认 false）
	SkipBOM bool `json:"skip_bom"`

	// StrictMode 严格模式（遇到无法转换字符时报错，默认 false）
	StrictMode bool `json:"strict_mode"`
}

// StreamResult 流处理结果
type StreamResult struct {
	// BytesRead 读取的字节数
	BytesRead int64 `json:"bytes_read"`

	// BytesWritten 写入的字节数
	BytesWritten int64 `json:"bytes_written"`

	// SourceEncoding 检测到的源编码
	SourceEncoding string `json:"source_encoding"`

	// TargetEncoding 目标编码
	TargetEncoding string `json:"target_encoding"`

	// ProcessingTime 处理耗时
	ProcessingTime time.Duration `json:"processing_time"`

	// ErrorCount 转换错误次数
	ErrorCount int `json:"error_count"`
}

// FileProcessOptions 文件处理选项
type FileProcessOptions struct {
	// TargetEncoding 目标编码（默认 UTF-8）
	TargetEncoding string `json:"target_encoding"`

	// MinConfidence 最小置信度阈值（默认 0.8）
	MinConfidence float64 `json:"min_confidence"`

	// CreateBackup 是否创建备份文件（默认 true）
	CreateBackup bool `json:"create_backup"`

	// BackupSuffix 备份文件后缀（默认 ".bak"）
	BackupSuffix string `json:"backup_suffix"`

	// OverwriteExisting 是否覆盖已存在的输出文件（默认 false）
	OverwriteExisting bool `json:"overwrite_existing"`

	// BufferSize 缓冲区大小（字节，默认 8192）
	BufferSize int `json:"buffer_size"`

	// PreserveMode 是否保持文件权限（默认 true）
	PreserveMode bool `json:"preserve_mode"`

	// PreserveTime 是否保持文件时间戳（默认 true）
	PreserveTime bool `json:"preserve_time"`

	// DryRun 试运行模式，不实际修改文件（默认 false）
	DryRun bool `json:"dry_run"`
}

// FileProcessResult 文件处理结果
type FileProcessResult struct {
	// InputFile 输入文件路径
	InputFile string `json:"input_file"`

	// OutputFile 输出文件路径
	OutputFile string `json:"output_file"`

	// BackupFile 备份文件路径（如果创建了备份）
	BackupFile string `json:"backup_file,omitempty"`

	// SourceEncoding 检测到的源编码
	SourceEncoding string `json:"source_encoding"`

	// TargetEncoding 目标编码
	TargetEncoding string `json:"target_encoding"`

	// BytesProcessed 处理的字节数
	BytesProcessed int64 `json:"bytes_processed"`

	// ProcessingTime 处理耗时
	ProcessingTime time.Duration `json:"processing_time"`

	// DetectionConfidence 编码检测置信度
	DetectionConfidence float64 `json:"detection_confidence"`
}

// ProcessingStats 处理统计信息
type ProcessingStats struct {
	// TotalOperations 总操作数
	TotalOperations int64 `json:"total_operations"`

	// SuccessOperations 成功操作数
	SuccessOperations int64 `json:"success_operations"`

	// FailedOperations 失败操作数
	FailedOperations int64 `json:"failed_operations"`

	// TotalBytes 总字节数
	TotalBytes int64 `json:"total_bytes"`

	// TotalProcessingTime 总处理时间
	TotalProcessingTime time.Duration `json:"total_processing_time"`

	// AverageProcessingSpeed 平均处理速度（字节/秒）
	AverageProcessingSpeed float64 `json:"average_processing_speed"`

	// EncodingDistribution 编码分布统计
	EncodingDistribution map[string]int64 `json:"encoding_distribution"`

	// StartTime 统计开始时间
	StartTime time.Time `json:"start_time"`

	// LastUpdateTime 最后更新时间
	LastUpdateTime time.Time `json:"last_update_time"`
}