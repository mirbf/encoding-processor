package encoding

import (
	"time"
)

// defaultProcessor 实现 Processor 接口
type defaultProcessor struct {
	detector  Detector
	converter Converter
	config    *ProcessorConfig
}

// NewProcessor 创建新的处理器
func NewProcessor(config *ProcessorConfig) Processor {
	if config == nil {
		config = GetDefaultProcessorConfig()
	}

	return &defaultProcessor{
		detector:  NewDetector(config.DetectorConfig),
		converter: NewConverter(config.ConverterConfig),
		config:    config,
	}
}

// DetectEncoding 检测数据的编码格式
func (p *defaultProcessor) DetectEncoding(data []byte) (*DetectionResult, error) {
	return p.detector.DetectEncoding(data)
}

// DetectFileEncoding 检测文件的编码格式
func (p *defaultProcessor) DetectFileEncoding(filename string) (*DetectionResult, error) {
	return p.detector.DetectFileEncoding(filename)
}

// DetectBestEncoding 检测最可能的编码格式
func (p *defaultProcessor) DetectBestEncoding(data []byte) (string, error) {
	return p.detector.DetectBestEncoding(data)
}

// SmartDetectEncoding 智能编码检测
func (p *defaultProcessor) SmartDetectEncoding(data []byte) (*DetectionResult, error) {
	return p.detector.SmartDetectEncoding(data)
}

// Convert 在指定编码之间转换
func (p *defaultProcessor) Convert(data []byte, from, to string) ([]byte, error) {
	return p.converter.Convert(data, from, to)
}

// ConvertToUTF8 转换为 UTF-8 编码
func (p *defaultProcessor) ConvertToUTF8(data []byte, from string) ([]byte, error) {
	return p.converter.ConvertToUTF8(data, from)
}

// ConvertString 字符串编码转换
func (p *defaultProcessor) ConvertString(text, from, to string) (string, error) {
	return p.converter.ConvertString(text, from, to)
}

// SmartConvert 智能转换（自动检测源编码）
func (p *defaultProcessor) SmartConvert(data []byte, target string) (*ConvertResult, error) {
	if len(data) == 0 {
		return &ConvertResult{
			Data:           []byte{},
			SourceEncoding: EncodingUTF8,
			TargetEncoding: target,
			BytesProcessed: 0,
			ConversionTime: 0,
		}, nil
	}

	start := time.Now()

	// 检测源编码
	detection, err := p.detector.DetectEncoding(data)
	if err != nil {
		return nil, err
	}

	// 转换编码
	convertedData, err := p.converter.Convert(data, detection.Encoding, target)
	if err != nil {
		return nil, err
	}

	return &ConvertResult{
		Data:           convertedData,
		SourceEncoding: detection.Encoding,
		TargetEncoding: target,
		BytesProcessed: int64(len(data)),
		ConversionTime: time.Since(start),
	}, nil
}

// SmartConvertString 智能字符串转换（自动检测源编码）
func (p *defaultProcessor) SmartConvertString(text, target string) (*StringConvertResult, error) {
	if text == "" {
		return &StringConvertResult{
			Text:           "",
			SourceEncoding: EncodingUTF8,
			TargetEncoding: target,
			BytesProcessed: 0,
			ConversionTime: 0,
		}, nil
	}

	start := time.Now()
	data := []byte(text)

	// 检测源编码
	detection, err := p.detector.DetectEncoding(data)
	if err != nil {
		return nil, err
	}

	// 转换编码
	convertedText, err := p.converter.ConvertString(text, detection.Encoding, target)
	if err != nil {
		return nil, err
	}

	return &StringConvertResult{
		Text:           convertedText,
		SourceEncoding: detection.Encoding,
		TargetEncoding: target,
		BytesProcessed: int64(len(data)),
		ConversionTime: time.Since(start),
	}, nil
}