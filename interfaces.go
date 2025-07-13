package encoding

import (
	"context"
	"io"
	"time"
)

// Detector 编码检测器接口
type Detector interface {
	// DetectEncoding 检测数据的编码格式
	DetectEncoding(data []byte) (*DetectionResult, error)

	// DetectFileEncoding 检测文件的编码格式
	DetectFileEncoding(filename string) (*DetectionResult, error)

	// DetectBestEncoding 检测最可能的编码格式（简化版本）
	DetectBestEncoding(data []byte) (string, error)
}

// Converter 编码转换器接口
type Converter interface {
	// Convert 在指定编码之间转换
	Convert(data []byte, from, to string) ([]byte, error)

	// ConvertToUTF8 转换为 UTF-8 编码
	ConvertToUTF8(data []byte, from string) ([]byte, error)

	// ConvertString 字符串编码转换
	ConvertString(text, from, to string) (string, error)
}

// Processor 编码处理器接口，集成检测和转换功能
type Processor interface {
	Detector
	Converter

	// SmartConvert 智能转换（自动检测源编码）
	SmartConvert(data []byte, target string) (*ConvertResult, error)

	// SmartConvertString 智能字符串转换（自动检测源编码）
	SmartConvertString(text, target string) (*StringConvertResult, error)
}

// StreamProcessor 流式处理接口
type StreamProcessor interface {
	// ProcessReader 处理输入流
	ProcessReader(ctx context.Context, r io.Reader, sourceEncoding, targetEncoding string) (io.Reader, error)

	// ProcessWriter 创建转换写入器
	ProcessWriter(ctx context.Context, w io.Writer, sourceEncoding, targetEncoding string) (io.Writer, error)

	// ProcessReaderWriter 处理读写流
	ProcessReaderWriter(ctx context.Context, r io.Reader, w io.Writer, options *StreamOptions) (*StreamResult, error)
}

// FileProcessor 文件处理接口
type FileProcessor interface {
	// ProcessFile 处理文件（检测并转换编码）
	ProcessFile(inputFile, outputFile string, options *FileProcessOptions) (*FileProcessResult, error)

	// ProcessFileInPlace 就地处理文件（直接修改源文件）
	ProcessFileInPlace(file string, options *FileProcessOptions) (*FileProcessResult, error)

	// ProcessFileToBytes 读取文件并转换编码，返回字节数组
	ProcessFileToBytes(filename, targetEncoding string) ([]byte, error)

	// ProcessFileToString 读取文件并转换编码，返回字符串
	ProcessFileToString(filename, targetEncoding string) (string, error)
}

// MetricsCollector 性能监控和统计接口
type MetricsCollector interface {
	// GetStats 获取处理统计信息
	GetStats() *ProcessingStats

	// ResetStats 重置统计信息
	ResetStats()

	// RecordOperation 记录操作
	RecordOperation(operation string, duration time.Duration)

	// RecordError 记录错误
	RecordError(operation string, err error)
}

// Logger 日志记录器接口
type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
}