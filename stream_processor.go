package encoding

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"golang.org/x/text/transform"
)

// defaultStreamProcessor 实现 StreamProcessor 接口
type defaultStreamProcessor struct {
	processor Processor
	config    *ProcessorConfig
	bufferPool sync.Pool
}

// streamReader 包装转换后的读取器
type streamReader struct {
	reader      io.Reader
	transformer transform.Transformer
	buf         []byte
	err         error
}

// streamWriter 包装转换后的写入器
type streamWriter struct {
	writer      io.Writer
	transformer transform.Transformer
	buf         []byte
}

// NewStreamProcessor 创建新的流处理器
func NewStreamProcessor(config *ProcessorConfig) StreamProcessor {
	if config == nil {
		config = GetDefaultProcessorConfig()
	}

	sp := &defaultStreamProcessor{
		processor: NewProcessor(config),
		config:    config,
	}

	// 初始化缓冲区池
	sp.bufferPool = sync.Pool{
		New: func() interface{} {
			return make([]byte, config.ConverterConfig.BufferSize)
		},
	}

	return sp
}

// ProcessReader 处理输入流
func (sp *defaultStreamProcessor) ProcessReader(ctx context.Context, r io.Reader, sourceEncoding, targetEncoding string) (io.Reader, error) {
	if sourceEncoding == "" {
		// 需要自动检测编码，先读取样本
		return sp.processReaderWithDetection(ctx, r, targetEncoding)
	}

	// 直接创建转换读取器
	return sp.createTransformReader(r, sourceEncoding, targetEncoding)
}

// ProcessWriter 创建转换写入器
func (sp *defaultStreamProcessor) ProcessWriter(ctx context.Context, w io.Writer, sourceEncoding, targetEncoding string) (io.Writer, error) {
	return sp.createTransformWriter(w, sourceEncoding, targetEncoding)
}

// ProcessReaderWriter 处理读写流
func (sp *defaultStreamProcessor) ProcessReaderWriter(ctx context.Context, r io.Reader, w io.Writer, options *StreamOptions) (*StreamResult, error) {
	if options == nil {
		options = &StreamOptions{
			TargetEncoding:      EncodingUTF8,
			BufferSize:          DefaultBufferSize,
			DetectionSampleSize: DefaultSampleSize,
			StrictMode:          false,
		}
	}

	start := time.Now()
	var bytesRead, bytesWritten int64
	var sourceEncoding string
	var errorCount int

	// 如果需要自动检测编码
	if options.SourceEncoding == "" {
		detected, sample, err := sp.detectEncodingFromStream(r, options.DetectionSampleSize)
		if err != nil {
			return nil, fmt.Errorf("failed to detect encoding from stream: %w", err)
		}
		sourceEncoding = detected
		
		// 先写入检测样本
		if len(sample) > 0 {
			convertedSample, err := sp.processor.Convert(sample, sourceEncoding, options.TargetEncoding)
			if err != nil {
				if !options.StrictMode {
					errorCount++
				} else {
					return nil, fmt.Errorf("failed to convert detection sample: %w", err)
				}
			} else {
				n, err := w.Write(convertedSample)
				if err != nil {
					return nil, fmt.Errorf("failed to write converted sample: %w", err)
				}
				bytesWritten += int64(n)
			}
			bytesRead += int64(len(sample))
		}
	} else {
		sourceEncoding = options.SourceEncoding
	}

	// 处理剩余数据
	buffer := make([]byte, options.BufferSize)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		n, err := r.Read(buffer)
		if n > 0 {
			bytesRead += int64(n)
			
			// 转换数据
			converted, convertErr := sp.processor.Convert(buffer[:n], sourceEncoding, options.TargetEncoding)
			if convertErr != nil {
				if options.StrictMode {
					return nil, fmt.Errorf("conversion failed at byte %d: %w", bytesRead, convertErr)
				}
				errorCount++
				// 非严格模式下跳过错误数据
				continue
			}

			// 写入转换后的数据
			written, writeErr := w.Write(converted)
			if writeErr != nil {
				return nil, fmt.Errorf("write failed: %w", writeErr)
			}
			bytesWritten += int64(written)
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read failed: %w", err)
		}
	}

	return &StreamResult{
		BytesRead:      bytesRead,
		BytesWritten:   bytesWritten,
		SourceEncoding: sourceEncoding,
		TargetEncoding: options.TargetEncoding,
		ProcessingTime: time.Since(start),
		ErrorCount:     errorCount,
	}, nil
}

// processReaderWithDetection 处理需要检测编码的读取器
func (sp *defaultStreamProcessor) processReaderWithDetection(ctx context.Context, r io.Reader, targetEncoding string) (io.Reader, error) {
	// 创建缓冲读取器
	bufReader := bufio.NewReader(r)
	
	// 预读样本用于检测编码
	sample := make([]byte, DefaultSampleSize)
	n, err := bufReader.Read(sample)
	if err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read sample for detection: %w", err)
	}
	
	if n == 0 {
		// 空数据，返回空读取器
		return &streamReader{
			reader: bufReader,
			buf:    []byte{},
		}, nil
	}

	// 检测编码
	result, err := sp.processor.DetectEncoding(sample[:n])
	if err != nil {
		return nil, fmt.Errorf("failed to detect encoding: %w", err)
	}

	// 创建多读取器，将样本和剩余数据合并
	multiReader := io.MultiReader(
		io.NewSectionReader(
			&bytesReaderAt{data: sample[:n]}, 
			0, 
			int64(n),
		),
		bufReader,
	)

	return sp.createTransformReader(multiReader, result.Encoding, targetEncoding)
}

// detectEncodingFromStream 从流中检测编码
func (sp *defaultStreamProcessor) detectEncodingFromStream(r io.Reader, sampleSize int) (string, []byte, error) {
	sample := make([]byte, sampleSize)
	n, err := r.Read(sample)
	if err != nil && err != io.EOF {
		return "", nil, err
	}

	if n == 0 {
		return EncodingUTF8, []byte{}, nil
	}

	result, err := sp.processor.DetectEncoding(sample[:n])
	if err != nil {
		return "", nil, err
	}

	return result.Encoding, sample[:n], nil
}

// createTransformReader 创建转换读取器
func (sp *defaultStreamProcessor) createTransformReader(r io.Reader, sourceEncoding, targetEncoding string) (io.Reader, error) {
	if sourceEncoding == targetEncoding {
		return r, nil
	}

	// 获取转换器
	converter, ok := sp.processor.(*defaultProcessor)
	if !ok {
		return nil, fmt.Errorf("invalid processor type")
	}

	// 获取源编码解码器
	decoder, err := converter.converter.(*defaultConverter).getDecoder(sourceEncoding)
	if err != nil {
		return nil, fmt.Errorf("failed to get decoder for %s: %w", sourceEncoding, err)
	}

	// 获取目标编码编码器
	encoder, err := converter.converter.(*defaultConverter).getEncoder(targetEncoding)
	if err != nil {
		return nil, fmt.Errorf("failed to get encoder for %s: %w", targetEncoding, err)
	}

	// 创建转换链
	var transformer transform.Transformer
	if sourceEncoding == EncodingUTF8 {
		transformer = encoder
	} else if targetEncoding == EncodingUTF8 {
		transformer = decoder
	} else {
		transformer = transform.Chain(decoder, encoder)
	}

	return transform.NewReader(r, transformer), nil
}

// createTransformWriter 创建转换写入器
func (sp *defaultStreamProcessor) createTransformWriter(w io.Writer, sourceEncoding, targetEncoding string) (io.Writer, error) {
	if sourceEncoding == targetEncoding {
		return w, nil
	}

	// 获取转换器
	converter, ok := sp.processor.(*defaultProcessor)
	if !ok {
		return nil, fmt.Errorf("invalid processor type")
	}

	// 获取源编码解码器
	decoder, err := converter.converter.(*defaultConverter).getDecoder(sourceEncoding)
	if err != nil {
		return nil, fmt.Errorf("failed to get decoder for %s: %w", sourceEncoding, err)
	}

	// 获取目标编码编码器
	encoder, err := converter.converter.(*defaultConverter).getEncoder(targetEncoding)
	if err != nil {
		return nil, fmt.Errorf("failed to get encoder for %s: %w", targetEncoding, err)
	}

	// 创建转换链
	var transformer transform.Transformer
	if sourceEncoding == EncodingUTF8 {
		transformer = encoder
	} else if targetEncoding == EncodingUTF8 {
		transformer = decoder
	} else {
		transformer = transform.Chain(decoder, encoder)
	}

	return transform.NewWriter(w, transformer), nil
}

// bytesReaderAt 实现 io.ReaderAt 接口
type bytesReaderAt struct {
	data []byte
}

func (r *bytesReaderAt) ReadAt(p []byte, off int64) (n int, err error) {
	if off < 0 || off >= int64(len(r.data)) {
		return 0, io.EOF
	}
	n = copy(p, r.data[off:])
	if n < len(p) {
		err = io.EOF
	}
	return
}

// Read 实现 streamReader 的 Read 方法
func (sr *streamReader) Read(p []byte) (n int, err error) {
	if sr.err != nil {
		return 0, sr.err
	}

	// 如果有缓冲数据，先返回缓冲数据
	if len(sr.buf) > 0 {
		n = copy(p, sr.buf)
		sr.buf = sr.buf[n:]
		return n, nil
	}

	// 从底层读取器读取数据
	return sr.reader.Read(p)
}

// Write 实现 streamWriter 的 Write 方法
func (sw *streamWriter) Write(p []byte) (n int, err error) {
	if sw.transformer == nil {
		return sw.writer.Write(p)
	}

	// 使用转换器转换数据
	reader := transform.NewReader(io.NopCloser(bytes.NewReader(p)), sw.transformer)
	converted, err := io.ReadAll(reader)
	if err != nil {
		return 0, err
	}

	// 写入转换后的数据
	_, err = sw.writer.Write(converted)
	if err != nil {
		return 0, err
	}

	return len(p), nil
}