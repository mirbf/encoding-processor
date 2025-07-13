# EncodingProcessor API 设计文档

## 概述

EncodingProcessor 是一个专注于编码检测和转换的 Go 语言库，提供简单易用且功能强大的编码处理能力。

**核心原则**：
- 专注于编码检测和转换
- 支持数据流、单个文件、字节数组、字符串的处理
- 不负责目录遍历和批量文件管理（应用层职责）

## 核心接口

### Detector 接口

编码检测器接口，用于检测文本的编码格式。

```go
type Detector interface {
    // DetectEncoding 检测数据的编码格式
    // 参数: data - 待检测的字节数据
    // 返回: 检测结果和错误信息
    DetectEncoding(data []byte) (*DetectionResult, error)
    
    // DetectFileEncoding 检测文件的编码格式
    // 参数: filename - 文件路径
    // 返回: 检测结果和错误信息
    DetectFileEncoding(filename string) (*DetectionResult, error)
    
    // DetectBestEncoding 检测最可能的编码格式（简化版本）
    // 参数: data - 待检测的字节数据
    // 返回: 编码名称和错误信息
    DetectBestEncoding(data []byte) (string, error)
}
```

### Converter 接口

编码转换器接口，用于在不同编码格式之间转换。

```go
type Converter interface {
    // Convert 在指定编码之间转换
    // 参数: data - 源数据, from - 源编码, to - 目标编码
    // 返回: 转换后的数据和错误信息
    Convert(data []byte, from, to string) ([]byte, error)
    
    // ConvertToUTF8 转换为 UTF-8 编码
    // 参数: data - 源数据, from - 源编码
    // 返回: UTF-8 编码的数据和错误信息
    ConvertToUTF8(data []byte, from string) ([]byte, error)
    
    // ConvertString 字符串编码转换
    // 参数: text - 源字符串, from - 源编码, to - 目标编码
    // 返回: 转换后的字符串和错误信息
    ConvertString(text, from, to string) (string, error)
}
```

### Processor 接口

编码处理器接口，集成检测和转换功能。

```go
type Processor interface {
    Detector
    Converter
    
    // SmartConvert 智能转换（自动检测源编码）
    // 参数: data - 源数据, target - 目标编码
    // 返回: 转换结果和错误信息
    SmartConvert(data []byte, target string) (*ConvertResult, error)
    
    // SmartConvertString 智能字符串转换（自动检测源编码）
    // 参数: text - 源字符串, target - 目标编码
    // 返回: 转换结果和错误信息
    SmartConvertString(text, target string) (*StringConvertResult, error)
}
```

### StreamProcessor 接口

流式处理接口，用于处理大文件或实时数据流。

```go
type StreamProcessor interface {
    // ProcessReader 处理输入流
    // 参数: ctx - 上下文, r - 输入流, sourceEncoding - 源编码, targetEncoding - 目标编码
    // 返回: 转换后的读取器和错误信息
    ProcessReader(ctx context.Context, r io.Reader, sourceEncoding, targetEncoding string) (io.Reader, error)
    
    // ProcessWriter 创建转换写入器
    // 参数: ctx - 上下文, w - 目标写入器, sourceEncoding - 源编码, targetEncoding - 目标编码
    // 返回: 转换写入器和错误信息
    ProcessWriter(ctx context.Context, w io.Writer, sourceEncoding, targetEncoding string) (io.Writer, error)
    
    // ProcessReaderWriter 处理读写流
    // 参数: ctx - 上下文, r - 输入流, w - 输出流, options - 流处理选项
    // 返回: 处理统计和错误信息
    ProcessReaderWriter(ctx context.Context, r io.Reader, w io.Writer, options *StreamOptions) (*StreamResult, error)
}
```

### FileProcessor 接口

文件处理接口，专门处理单个文件的编码转换。

```go
type FileProcessor interface {
    // ProcessFile 处理文件（检测并转换编码）
    // 参数: inputFile - 输入文件, outputFile - 输出文件, options - 处理选项
    // 返回: 处理结果和错误信息
    ProcessFile(inputFile, outputFile string, options *FileProcessOptions) (*FileProcessResult, error)
    
    // ProcessFileInPlace 就地处理文件（直接修改源文件）
    // 参数: file - 文件路径, options - 处理选项
    // 返回: 处理结果和错误信息
    ProcessFileInPlace(file string, options *FileProcessOptions) (*FileProcessResult, error)
    
    // ProcessFileToBytes 读取文件并转换编码，返回字节数组
    // 参数: filename - 文件路径, targetEncoding - 目标编码
    // 返回: 转换后的数据和错误信息
    ProcessFileToBytes(filename, targetEncoding string) ([]byte, error)
    
    // ProcessFileToString 读取文件并转换编码，返回字符串
    // 参数: filename - 文件路径, targetEncoding - 目标编码
    // 返回: 转换后的字符串和错误信息
    ProcessFileToString(filename, targetEncoding string) (string, error)
}
```

### MetricsCollector 接口

性能监控和统计接口。

```go
type MetricsCollector interface {
    // GetStats 获取处理统计信息
    // 返回: 统计信息
    GetStats() *ProcessingStats
    
    // ResetStats 重置统计信息
    ResetStats()
    
    // RecordOperation 记录操作
    // 参数: operation - 操作类型, duration - 耗时
    RecordOperation(operation string, duration time.Duration)
    
    // RecordError 记录错误
    // 参数: operation - 操作类型, err - 错误信息
    RecordError(operation string, err error)
}
```

## 数据结构

### DetectionResult

编码检测结果结构。

```go
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
```

### ConvertResult

编码转换结果结构。

```go
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
```

### StringConvertResult

字符串转换结果。

```go
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
```

### StreamOptions

流处理选项。

```go
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
```

### StreamResult

流处理结果。

```go
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
```

### FileProcessOptions

文件处理选项。

```go
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
```

### FileProcessResult

文件处理结果。

```go
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
```

### ProcessingStats

处理统计信息。

```go
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
```

## 配置选项

### DetectorConfig

检测器配置。

```go
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
```

### ConverterConfig

转换器配置。

```go
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
```

### ProcessorConfig

处理器配置（集成配置）。

```go
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
```

### Logger

日志记录器接口。

```go
type Logger interface {
    Debug(msg string, fields ...interface{})
    Info(msg string, fields ...interface{})
    Warn(msg string, fields ...interface{})
    Error(msg string, fields ...interface{})
}
```

## 错误类型

### 预定义错误

```go
var (
    // ErrUnsupportedEncoding 不支持的编码
    ErrUnsupportedEncoding = errors.New("unsupported encoding")
    
    // ErrDetectionFailed 检测失败
    ErrDetectionFailed = errors.New("encoding detection failed")
    
    // ErrConversionFailed 转换失败
    ErrConversionFailed = errors.New("encoding conversion failed")
    
    // ErrInvalidInput 无效输入
    ErrInvalidInput = errors.New("invalid input data")
    
    // ErrFileTooLarge 文件过大
    ErrFileTooLarge = errors.New("file too large")
    
    // ErrInsufficientMemory 内存不足
    ErrInsufficientMemory = errors.New("insufficient memory")
    
    // ErrFileNotFound 文件不存在
    ErrFileNotFound = errors.New("file not found")
    
    // ErrPermissionDenied 权限不足
    ErrPermissionDenied = errors.New("permission denied")
    
    // ErrInvalidConfiguration 无效配置
    ErrInvalidConfiguration = errors.New("invalid configuration")
)
```

### 结构化错误

```go
// EncodingError 编码相关错误
type EncodingError struct {
    Op       string // 操作名称
    Encoding string // 相关编码
    File     string // 相关文件（可选）
    Err      error  // 原始错误
}

func (e *EncodingError) Error() string {
    if e.File != "" {
        return fmt.Sprintf("encoding %s in %s for file %s: %v", e.Encoding, e.Op, e.File, e.Err)
    }
    return fmt.Sprintf("encoding %s in %s: %v", e.Encoding, e.Op, e.Err)
}

func (e *EncodingError) Unwrap() error {
    return e.Err
}

// FileOperationError 文件操作错误
type FileOperationError struct {
    Op   string // 操作名称
    File string // 文件路径
    Err  error  // 原始错误
}

func (e *FileOperationError) Error() string {
    return fmt.Sprintf("file operation %s on %s: %v", e.Op, e.File, e.Err)
}

func (e *FileOperationError) Unwrap() error {
    return e.Err
}
```

## 工厂函数

### 基础创建函数

```go
// NewDetector 创建编码检测器
func NewDetector(config ...*DetectorConfig) Detector

// NewConverter 创建编码转换器  
func NewConverter(config ...*ConverterConfig) Converter

// NewProcessor 创建编码处理器
func NewProcessor(config *ProcessorConfig) Processor

// NewStreamProcessor 创建流处理器
func NewStreamProcessor(config *ProcessorConfig) StreamProcessor

// NewFileProcessor 创建文件处理器
func NewFileProcessor(config *ProcessorConfig) FileProcessor

// NewMetricsCollector 创建性能监控器
func NewMetricsCollector() MetricsCollector
```

### 便捷创建函数

```go
// NewDefault 创建默认处理器
func NewDefault() Processor

// NewDefaultWithMetrics 创建带性能监控的默认处理器
func NewDefaultWithMetrics() (Processor, MetricsCollector)

// NewDefaultStream 创建默认流处理器
func NewDefaultStream() StreamProcessor

// NewDefaultFile 创建默认文件处理器
func NewDefaultFile() FileProcessor

// NewWithLogger 创建带自定义日志的处理器
func NewWithLogger(logger Logger) Processor

// NewWithConfig 使用自定义配置创建处理器
func NewWithConfig(detectorConfig *DetectorConfig, converterConfig *ConverterConfig) Processor

// NewQuick 快速创建处理器（最少配置）
func NewQuick() Processor

// NewForCLI 创建适合命令行工具的处理器
func NewForCLI() Processor

// NewForWebService 创建适合 Web 服务的处理器
func NewForWebService() Processor

// NewForBatchProcessing 创建适合批量处理的处理器
func NewForBatchProcessing() Processor

// NewHighPerformance 创建高性能处理器
func NewHighPerformance() Processor

// NewMemoryEfficient 创建内存高效的处理器
func NewMemoryEfficient() Processor

// NewStrictMode 创建严格模式处理器（遇到错误立即失败）
func NewStrictMode() Processor

// NewTolerantMode 创建容错模式处理器（尽量处理，忽略错误）
func NewTolerantMode() Processor
```

## 常量定义

### 支持的编码

```go
const (
    EncodingUTF8        = "UTF-8"
    EncodingUTF16       = "UTF-16"
    EncodingUTF16LE     = "UTF-16LE"
    EncodingUTF16BE     = "UTF-16BE"
    EncodingUTF32       = "UTF-32"
    EncodingUTF32LE     = "UTF-32LE"
    EncodingUTF32BE     = "UTF-32BE"
    EncodingGBK         = "GBK"
    EncodingGB2312      = "GB2312"
    EncodingGB18030     = "GB18030"
    EncodingBIG5        = "BIG5"
    EncodingShiftJIS    = "SHIFT_JIS"
    EncodingEUCJP       = "EUC-JP"
    EncodingEUCKR       = "EUC-KR"
    EncodingISO88591    = "ISO-8859-1"
    EncodingISO88592    = "ISO-8859-2"
    EncodingISO88595    = "ISO-8859-5"
    EncodingISO885915   = "ISO-8859-15"
    EncodingWindows1250 = "WINDOWS-1250"
    EncodingWindows1251 = "WINDOWS-1251"
    EncodingWindows1252 = "WINDOWS-1252"
    EncodingWindows1254 = "WINDOWS-1254"
    EncodingKOI8R       = "KOI8-R"
    EncodingCP866       = "CP866"
    EncodingMacintosh   = "MACINTOSH"
)
```

### 操作类型

```go
const (
    OperationDetect    = "detect"
    OperationConvert   = "convert"
    OperationProcess   = "process"
    OperationValidate  = "validate"
)
```

### 默认配置值

```go
const (
    DefaultSampleSize         = 8192          // 默认检测样本大小
    DefaultMinConfidence      = 0.8           // 默认最小置信度
    DefaultBufferSize         = 8192          // 默认缓冲区大小
    DefaultInvalidChar        = "?"           // 默认无效字符替换
    DefaultBackupSuffix       = ".bak"        // 默认备份后缀
    DefaultChunkSize          = 1024 * 1024   // 默认分块大小 (1MB)
    DefaultMaxFileSize        = 100 << 20     // 默认最大文件大小 (100MB)
    DefaultCacheSize          = 1000          // 默认缓存大小
    DefaultCacheTTL           = time.Hour     // 默认缓存过期时间
)
```

### 换行符常量

```go
const (
    LineEndingLF   = "\n"      // Unix/Linux 换行符
    LineEndingCRLF = "\r\n"    // Windows 换行符
    LineEndingCR   = "\r"      // Classic Mac 换行符
)
```

## 使用示例

### 1. 基础编码检测和转换

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/mirbf/encoding-processor"
)

func main() {
    // 创建处理器
    processor := encoding.NewDefault()
    
    // 检测编码
    data := []byte("这是一段中文文本")
    result, err := processor.DetectEncoding(data)
    if err != nil {
        log.Fatalf("检测失败: %v", err)
    }
    
    fmt.Printf("检测到编码: %s (置信度: %.2f)\n", result.Encoding, result.Confidence)
    
    // 智能转换
    convertResult, err := processor.SmartConvert(data, encoding.EncodingUTF8)
    if err != nil {
        log.Fatalf("转换失败: %v", err)
    }
    
    fmt.Printf("转换结果: %s\n", string(convertResult.Data))
}
```

### 2. 字符串处理

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/mirbf/encoding-processor"
)

func main() {
    processor := encoding.NewDefault()
    
    // 智能字符串转换
    text := "这是测试文本"
    result, err := processor.SmartConvertString(text, encoding.EncodingGBK)
    if err != nil {
        log.Fatalf("转换失败: %v", err)
    }
    
    fmt.Printf("转换结果: %s\n", result.Text)
    fmt.Printf("源编码: %s\n", result.SourceEncoding)
}
```

### 3. 文件处理

```go
package main

import (
    "log"
    
    "github.com/mirbf/encoding-processor"
)

func main() {
    fileProcessor := encoding.NewDefaultFile()
    
    options := &encoding.FileProcessOptions{
        TargetEncoding:   encoding.EncodingUTF8,
        CreateBackup:     true,
        OverwriteExisting: false,
        PreserveMode:     true,
        PreserveTime:     true,
    }
    
    result, err := fileProcessor.ProcessFile("input.txt", "output.txt", options)
    if err != nil {
        log.Fatalf("文件处理失败: %v", err)
    }
    
    log.Printf("处理完成: %s -> %s", result.InputFile, result.OutputFile)
    if result.BackupFile != "" {
        log.Printf("备份文件: %s", result.BackupFile)
    }
}
```

### 4. 流式处理

```go
package main

import (
    "context"
    "log"
    "os"
    
    "github.com/mirbf/encoding-processor"
)

func main() {
    streamProcessor := encoding.NewDefaultStream()
    
    // 打开输入文件
    inputFile, err := os.Open("large_file.txt")
    if err != nil {
        log.Fatal(err)
    }
    defer inputFile.Close()
    
    // 创建输出文件
    outputFile, err := os.Create("large_file_utf8.txt")
    if err != nil {
        log.Fatal(err)
    }
    defer outputFile.Close()
    
    options := &encoding.StreamOptions{
        SourceEncoding: "", // 自动检测
        TargetEncoding: encoding.EncodingUTF8,
        BufferSize:     16384,
        StrictMode:     false,
    }
    
    ctx := context.Background()
    result, err := streamProcessor.ProcessReaderWriter(ctx, inputFile, outputFile, options)
    if err != nil {
        log.Fatalf("流处理失败: %v", err)
    }
    
    log.Printf("处理完成: 读取 %d 字节, 写入 %d 字节", result.BytesRead, result.BytesWritten)
}
```

---

*本API文档专注于编码检测和转换的核心功能，目录遍历和批量文件管理应在应用层实现*