package encoding

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"time"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// defaultConverter 实现 Converter 接口
type defaultConverter struct {
	config *ConverterConfig
	pool   *transformerPool
	mutex  sync.RWMutex
}

// transformerPool 转换器池
type transformerPool struct {
	pools map[string]*sync.Pool
	mutex sync.RWMutex
}

// NewConverter 创建新的转换器
func NewConverter(config ...*ConverterConfig) Converter {
	var cfg *ConverterConfig
	if len(config) > 0 && config[0] != nil {
		cfg = config[0]
	} else {
		cfg = GetDefaultConverterConfig()
	}

	return &defaultConverter{
		config: cfg,
		pool: &transformerPool{
			pools: make(map[string]*sync.Pool),
		},
	}
}

// Convert 在指定编码之间转换
func (c *defaultConverter) Convert(data []byte, from, to string) ([]byte, error) {
	if len(data) == 0 {
		return []byte{}, nil
	}

	// 如果源编码和目标编码相同，直接返回
	if from == to {
		return data, nil
	}

	start := time.Now()
	defer func() {
		// 这里可以记录性能指标
		_ = time.Since(start)
	}()

	// 获取源编码解码器
	fromDecoder, err := c.getDecoder(from)
	if err != nil {
		return nil, &EncodingError{
			Op:       OperationConvert,
			Encoding: from,
			Err:      fmt.Errorf("failed to get decoder for %s: %w", from, err),
		}
	}

	// 获取目标编码编码器
	toEncoder, err := c.getEncoder(to)
	if err != nil {
		return nil, &EncodingError{
			Op:       OperationConvert,
			Encoding: to,
			Err:      fmt.Errorf("failed to get encoder for %s: %w", to, err),
		}
	}

	// 创建转换管道: 源编码 -> UTF-8 -> 目标编码
	var transformer transform.Transformer
	if from == EncodingUTF8 {
		// 源编码是 UTF-8，直接编码到目标编码
		transformer = toEncoder
	} else if to == EncodingUTF8 {
		// 目标编码是 UTF-8，直接从源编码解码
		transformer = fromDecoder
	} else {
		// 两步转换：源编码 -> UTF-8 -> 目标编码
		transformer = transform.Chain(fromDecoder, toEncoder)
	}

	// 执行转换
	result, err := c.doTransform(data, transformer)
	if err != nil {
		return nil, &EncodingError{
			Op:       OperationConvert,
			Encoding: fmt.Sprintf("%s->%s", from, to),
			Err:      err,
		}
	}

	return result, nil
}

// ConvertToUTF8 转换为 UTF-8 编码
func (c *defaultConverter) ConvertToUTF8(data []byte, from string) ([]byte, error) {
	return c.Convert(data, from, EncodingUTF8)
}

// ConvertString 字符串编码转换
func (c *defaultConverter) ConvertString(text, from, to string) (string, error) {
	data, err := c.Convert([]byte(text), from, to)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// getDecoder 获取解码器
func (c *defaultConverter) getDecoder(encodingName string) (transform.Transformer, error) {
	enc, err := c.getEncoding(encodingName)
	if err != nil {
		return nil, err
	}
	return enc.NewDecoder(), nil
}

// getEncoder 获取编码器
func (c *defaultConverter) getEncoder(encodingName string) (transform.Transformer, error) {
	enc, err := c.getEncoding(encodingName)
	if err != nil {
		return nil, err
	}
	return enc.NewEncoder(), nil
}

// getEncoding 根据编码名称获取编码实例
func (c *defaultConverter) getEncoding(name string) (encoding.Encoding, error) {
	switch name {
	case EncodingUTF8:
		return unicode.UTF8, nil
	case EncodingUTF16:
		return unicode.UTF16(unicode.BigEndian, unicode.UseBOM), nil
	case EncodingUTF16LE:
		return unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM), nil
	case EncodingUTF16BE:
		return unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM), nil
	case EncodingUTF32:
		return unicode.UTF16(unicode.BigEndian, unicode.UseBOM), nil // UTF32 not directly supported, use UTF16
	case EncodingUTF32LE:
		return unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM), nil
	case EncodingUTF32BE:
		return unicode.UTF16(unicode.BigEndian, unicode.IgnoreBOM), nil

	// 中文编码
	case EncodingGBK, EncodingGB2312:
		return simplifiedchinese.GBK, nil
	case EncodingGB18030:
		return simplifiedchinese.GB18030, nil
	case EncodingBIG5:
		return traditionalchinese.Big5, nil

	// 日文编码
	case EncodingShiftJIS:
		return japanese.ShiftJIS, nil
	case EncodingEUCJP:
		return japanese.EUCJP, nil

	// 韩文编码
	case EncodingEUCKR:
		return korean.EUCKR, nil

	// 西欧编码
	case EncodingISO88591:
		return charmap.ISO8859_1, nil
	case EncodingISO88592:
		return charmap.ISO8859_2, nil
	case EncodingISO88595:
		return charmap.ISO8859_5, nil
	case EncodingISO885915:
		return charmap.ISO8859_15, nil
	case EncodingWindows1250:
		return charmap.Windows1250, nil
	case EncodingWindows1251:
		return charmap.Windows1251, nil
	case EncodingWindows1252:
		return charmap.Windows1252, nil
	case EncodingWindows1254:
		return charmap.Windows1254, nil
	case EncodingKOI8R:
		return charmap.KOI8R, nil
	case EncodingCP866:
		return charmap.CodePage866, nil
	case EncodingMacintosh:
		return charmap.Macintosh, nil

	default:
		return nil, fmt.Errorf("unsupported encoding: %s", name)
	}
}

// doTransform 执行实际的转换操作
func (c *defaultConverter) doTransform(data []byte, transformer transform.Transformer) ([]byte, error) {
	// 检查内存限制
	if c.config.MaxMemoryUsage > 0 && int64(len(data)) > c.config.MaxMemoryUsage {
		return nil, ErrInsufficientMemory
	}

	// 对于大数据，使用分块处理
	if int64(len(data)) > c.config.ChunkSize {
		return c.transformLargeData(data, transformer)
	}

	// 小数据直接转换
	return c.transformSmallData(data, transformer)
}

// transformSmallData 转换小数据
func (c *defaultConverter) transformSmallData(data []byte, transformer transform.Transformer) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(data), transformer)
	result, err := io.ReadAll(reader)
	if err != nil {
		if c.config.StrictMode {
			return nil, fmt.Errorf("conversion failed: %w", err)
		}
		// 非严格模式下，尝试忽略错误继续转换
		return c.transformWithErrorRecovery(data, transformer)
	}
	return result, nil
}

// transformLargeData 转换大数据（分块处理）
func (c *defaultConverter) transformLargeData(data []byte, transformer transform.Transformer) ([]byte, error) {
	var result bytes.Buffer
	chunkSize := int(c.config.ChunkSize)
	
	for offset := 0; offset < len(data); offset += chunkSize {
		end := offset + chunkSize
		if end > len(data) {
			end = len(data)
		}
		
		chunk := data[offset:end]
		converted, err := c.transformSmallData(chunk, transformer)
		if err != nil {
			return nil, fmt.Errorf("failed to convert chunk at offset %d: %w", offset, err)
		}
		
		result.Write(converted)
	}
	
	return result.Bytes(), nil
}

// transformWithErrorRecovery 带错误恢复的转换
func (c *defaultConverter) transformWithErrorRecovery(data []byte, transformer transform.Transformer) ([]byte, error) {
	var result bytes.Buffer
	src := bytes.NewReader(data)
	
	buf := make([]byte, c.config.BufferSize)
	for {
		n, err := src.Read(buf)
		if n == 0 {
			break
		}
		
		// 尝试转换当前块
		reader := transform.NewReader(bytes.NewReader(buf[:n]), transformer)
		converted, readErr := io.ReadAll(reader)
		
		if readErr != nil {
			// 转换失败，使用替换字符
			if c.config.InvalidCharReplacement != "" {
				result.WriteString(c.config.InvalidCharReplacement)
			}
		} else {
			result.Write(converted)
		}
		
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
	}
	
	return result.Bytes(), nil
}

// getTransformer 从池中获取转换器
func (c *defaultConverter) getTransformer(key string) transform.Transformer {
	c.pool.mutex.RLock()
	pool, exists := c.pool.pools[key]
	c.pool.mutex.RUnlock()
	
	if !exists {
		c.pool.mutex.Lock()
		// 双重检查
		if pool, exists = c.pool.pools[key]; !exists {
			pool = &sync.Pool{
				New: func() interface{} {
					// 这里应该根据 key 创建对应的转换器
					// 为了简化，这里先返回 nil
					return nil
				},
			}
			c.pool.pools[key] = pool
		}
		c.pool.mutex.Unlock()
	}
	
	if transformer := pool.Get(); transformer != nil {
		return transformer.(transform.Transformer)
	}
	
	// 如果池中没有可用的转换器，创建一个新的
	// 这里应该根据实际需求实现
	return nil
}

// putTransformer 将转换器放回池中
func (c *defaultConverter) putTransformer(key string, transformer transform.Transformer) {
	c.pool.mutex.RLock()
	pool, exists := c.pool.pools[key]
	c.pool.mutex.RUnlock()
	
	if exists {
		pool.Put(transformer)
	}
}