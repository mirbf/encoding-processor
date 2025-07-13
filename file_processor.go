package encoding

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"
)

// defaultFileProcessor 实现 FileProcessor 接口
type defaultFileProcessor struct {
	processor Processor
	config    *ProcessorConfig
}

// NewFileProcessor 创建新的文件处理器
func NewFileProcessor(config *ProcessorConfig) FileProcessor {
	if config == nil {
		config = GetDefaultProcessorConfig()
	}

	return &defaultFileProcessor{
		processor: NewProcessor(config),
		config:    config,
	}
}

// ProcessFile 处理文件（检测并转换编码）
func (fp *defaultFileProcessor) ProcessFile(inputFile, outputFile string, options *FileProcessOptions) (*FileProcessResult, error) {
	if options == nil {
		options = &FileProcessOptions{
			TargetEncoding:    EncodingUTF8,
			MinConfidence:     DefaultMinConfidence,
			CreateBackup:      true,
			BackupSuffix:      DefaultBackupSuffix,
			OverwriteExisting: false,
			BufferSize:        DefaultBufferSize,
			PreserveMode:      true,
			PreserveTime:      true,
			DryRun:            false,
		}
	}

	start := time.Now()

	// 检查输入文件
	inputInfo, err := os.Stat(inputFile)
	if err != nil {
		return nil, &FileOperationError{
			Op:   "stat",
			File: inputFile,
			Err:  err,
		}
	}

	// 检查文件大小限制
	if fp.config.MaxFileSize > 0 && inputInfo.Size() > fp.config.MaxFileSize {
		return nil, &FileOperationError{
			Op:   "size_check",
			File: inputFile,
			Err:  ErrFileTooLarge,
		}
	}

	// 检查输出文件是否存在
	if !options.OverwriteExisting {
		if _, err := os.Stat(outputFile); err == nil {
			return nil, &FileOperationError{
				Op:   "overwrite_check",
				File: outputFile,
				Err:  fmt.Errorf("output file exists and overwrite is disabled"),
			}
		}
	}

	// 如果是试运行模式，只检测编码
	if options.DryRun {
		return fp.dryRunProcess(inputFile, outputFile, options)
	}

	// 读取文件内容
	data, err := ioutil.ReadFile(inputFile)
	if err != nil {
		return nil, &FileOperationError{
			Op:   "read",
			File: inputFile,
			Err:  err,
		}
	}

	// 检测编码
	detection, err := fp.processor.DetectEncoding(data)
	if err != nil {
		return nil, err
	}

	// 检查检测置信度
	if detection.Confidence < options.MinConfidence {
		return nil, &EncodingError{
			Op:       OperationDetect,
			Encoding: detection.Encoding,
			File:     inputFile,
			Err:      fmt.Errorf("detection confidence %.2f below threshold %.2f", detection.Confidence, options.MinConfidence),
		}
	}

	// 如果源编码和目标编码相同，只需复制文件
	if detection.Encoding == options.TargetEncoding {
		return fp.copyFile(inputFile, outputFile, inputInfo, options, detection)
	}

	// 转换编码
	convertedData, err := fp.processor.Convert(data, detection.Encoding, options.TargetEncoding)
	if err != nil {
		return nil, err
	}

	// 创建备份（如果需要）
	var backupFile string
	if options.CreateBackup && inputFile == outputFile {
		backupFile, err = fp.createBackup(inputFile, options.BackupSuffix)
		if err != nil {
			return nil, err
		}
	}

	// 确保输出目录存在
	outputDir := filepath.Dir(outputFile)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, &FileOperationError{
			Op:   "mkdir",
			File: outputDir,
			Err:  err,
		}
	}

	// 写入转换后的数据
	err = fp.writeFileWithRecovery(outputFile, convertedData, inputInfo, options, backupFile)
	if err != nil {
		return nil, err
	}

	return &FileProcessResult{
		InputFile:           inputFile,
		OutputFile:          outputFile,
		BackupFile:          backupFile,
		SourceEncoding:      detection.Encoding,
		TargetEncoding:      options.TargetEncoding,
		BytesProcessed:      int64(len(data)),
		ProcessingTime:      time.Since(start),
		DetectionConfidence: detection.Confidence,
	}, nil
}

// ProcessFileInPlace 就地处理文件（直接修改源文件）
func (fp *defaultFileProcessor) ProcessFileInPlace(file string, options *FileProcessOptions) (*FileProcessResult, error) {
	return fp.ProcessFile(file, file, options)
}

// ProcessFileToBytes 读取文件并转换编码，返回字节数组
func (fp *defaultFileProcessor) ProcessFileToBytes(filename, targetEncoding string) ([]byte, error) {
	// 读取文件
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, &FileOperationError{
			Op:   "read",
			File: filename,
			Err:  err,
		}
	}

	// 智能转换
	result, err := fp.processor.SmartConvert(data, targetEncoding)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

// ProcessFileToString 读取文件并转换编码，返回字符串
func (fp *defaultFileProcessor) ProcessFileToString(filename, targetEncoding string) (string, error) {
	data, err := fp.ProcessFileToBytes(filename, targetEncoding)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// dryRunProcess 试运行处理
func (fp *defaultFileProcessor) dryRunProcess(inputFile, outputFile string, options *FileProcessOptions) (*FileProcessResult, error) {
	start := time.Now()

	// 读取文件用于检测
	data, err := ioutil.ReadFile(inputFile)
	if err != nil {
		return nil, &FileOperationError{
			Op:   "read",
			File: inputFile,
			Err:  err,
		}
	}

	// 检测编码
	detection, err := fp.processor.DetectEncoding(data)
	if err != nil {
		return nil, err
	}

	return &FileProcessResult{
		InputFile:           inputFile,
		OutputFile:          outputFile,
		BackupFile:          "",
		SourceEncoding:      detection.Encoding,
		TargetEncoding:      options.TargetEncoding,
		BytesProcessed:      int64(len(data)),
		ProcessingTime:      time.Since(start),
		DetectionConfidence: detection.Confidence,
	}, nil
}

// copyFile 复制文件（当源编码和目标编码相同时）
func (fp *defaultFileProcessor) copyFile(inputFile, outputFile string, inputInfo os.FileInfo, options *FileProcessOptions, detection *DetectionResult) (*FileProcessResult, error) {
	start := time.Now()

	data, err := ioutil.ReadFile(inputFile)
	if err != nil {
		return nil, &FileOperationError{
			Op:   "read",
			File: inputFile,
			Err:  err,
		}
	}

	// 创建备份（如果需要）
	var backupFile string
	if options.CreateBackup && inputFile == outputFile {
		backupFile, err = fp.createBackup(inputFile, options.BackupSuffix)
		if err != nil {
			return nil, err
		}
	}

	// 写入文件
	err = fp.writeFileWithRecovery(outputFile, data, inputInfo, options, backupFile)
	if err != nil {
		return nil, err
	}

	return &FileProcessResult{
		InputFile:           inputFile,
		OutputFile:          outputFile,
		BackupFile:          backupFile,
		SourceEncoding:      detection.Encoding,
		TargetEncoding:      options.TargetEncoding,
		BytesProcessed:      int64(len(data)),
		ProcessingTime:      time.Since(start),
		DetectionConfidence: detection.Confidence,
	}, nil
}

// createBackup 创建备份文件
func (fp *defaultFileProcessor) createBackup(filename, suffix string) (string, error) {
	backupFile := filename + suffix

	// 如果备份文件已存在，添加时间戳
	if _, err := os.Stat(backupFile); err == nil {
		timestamp := time.Now().Format("20060102150405")
		backupFile = fmt.Sprintf("%s.%s%s", filename, timestamp, suffix)
	}

	// 复制文件到备份位置
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", &FileOperationError{
			Op:   "read_for_backup",
			File: filename,
			Err:  err,
		}
	}

	err = ioutil.WriteFile(backupFile, data, 0644)
	if err != nil {
		return "", &FileOperationError{
			Op:   "create_backup",
			File: backupFile,
			Err:  err,
		}
	}

	return backupFile, nil
}

// writeFileWithRecovery 带恢复机制的文件写入
func (fp *defaultFileProcessor) writeFileWithRecovery(filename string, data []byte, originalInfo os.FileInfo, options *FileProcessOptions, backupFile string) error {
	// 创建临时文件
	tempFile := filename + ".tmp"

	// 写入临时文件
	err := ioutil.WriteFile(tempFile, data, 0644)
	if err != nil {
		return &FileOperationError{
			Op:   "write_temp",
			File: tempFile,
			Err:  err,
		}
	}

	// 设置文件权限
	if options.PreserveMode && originalInfo != nil {
		err = os.Chmod(tempFile, originalInfo.Mode())
		if err != nil {
			os.Remove(tempFile) // 清理临时文件
			return &FileOperationError{
				Op:   "chmod",
				File: tempFile,
				Err:  err,
			}
		}
	}

	// 原子性替换文件
	err = os.Rename(tempFile, filename)
	if err != nil {
		os.Remove(tempFile) // 清理临时文件
		// 如果有备份文件，尝试恢复
		if backupFile != "" {
			fp.restoreFromBackup(filename, backupFile)
		}
		return &FileOperationError{
			Op:   "rename",
			File: filename,
			Err:  err,
		}
	}

	// 设置文件时间戳
	if options.PreserveTime && originalInfo != nil {
		err = os.Chtimes(filename, originalInfo.ModTime(), originalInfo.ModTime())
		if err != nil {
			// 时间戳设置失败不是致命错误，只记录警告
			// 这里可以通过日志记录器记录警告
		}
	}

	return nil
}

// restoreFromBackup 从备份恢复文件
func (fp *defaultFileProcessor) restoreFromBackup(filename, backupFile string) error {
	data, err := ioutil.ReadFile(backupFile)
	if err != nil {
		return &FileOperationError{
			Op:   "read_backup",
			File: backupFile,
			Err:  err,
		}
	}

	err = ioutil.WriteFile(filename, data, 0644)
	if err != nil {
		return &FileOperationError{
			Op:   "restore_backup",
			File: filename,
			Err:  err,
		}
	}

	return nil
}