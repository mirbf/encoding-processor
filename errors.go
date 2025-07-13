package encoding

import (
	"errors"
	"fmt"
)

// 预定义错误
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