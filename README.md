# EncodingProcessor

一个专注于编码检测和转换的 Go 语言库，提供简单易用且功能强大的编码处理能力。

## 特性

- 🔍 **智能编码检测**: 支持多种编码格式的自动检测，包括 UTF-8、GBK、BIG5、Shift_JIS 等
- 🔄 **高效编码转换**: 在不同编码格式之间进行快速转换
- 📁 **文件处理**: 支持单个文件的编码检测和转换，包含安全的备份机制
- 🌊 **流式处理**: 支持大文件的流式处理，内存友好
- 📊 **性能监控**: 内置性能指标收集和统计功能
- ⚙️ **高度可配置**: 丰富的配置选项满足不同场景需求
- 🛡️ **错误恢复**: 完善的错误处理和恢复机制

## 安装

```bash
go get github.com/mirbf/encoding-processor
```

## 快速开始

### 基础使用

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
        log.Fatal(err)
    }
    
    fmt.Printf("检测到编码: %s (置信度: %.2f)\n", result.Encoding, result.Confidence)
    
    // 智能转换
    convertResult, err := processor.SmartConvertString("测试文本", encoding.EncodingUTF8)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("转换结果: %s\n", convertResult.Text)
}
```

### 文件处理

```go
// 创建文件处理器
fileProcessor := encoding.NewDefaultFile()

// 配置处理选项
options := &encoding.FileProcessOptions{
    TargetEncoding:    encoding.EncodingUTF8,
    CreateBackup:      true,
    OverwriteExisting: false,
    PreserveMode:      true,
    PreserveTime:      true,
}

// 处理文件
result, err := fileProcessor.ProcessFile("input.txt", "output.txt", options)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("处理完成: %s -> %s\n", result.InputFile, result.OutputFile)
```

### 流式处理

```go
streamProcessor := encoding.NewDefaultStream()

options := &encoding.StreamOptions{
    SourceEncoding: "", // 自动检测
    TargetEncoding: encoding.EncodingUTF8,
    BufferSize:     16384,
    StrictMode:     false,
}

ctx := context.Background()
result, err := streamProcessor.ProcessReaderWriter(ctx, inputReader, outputWriter, options)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("处理完成: 读取 %d 字节, 写入 %d 字节\n", result.BytesRead, result.BytesWritten)
```

## 支持的编码

- **Unicode**: UTF-8, UTF-16, UTF-16LE, UTF-16BE, UTF-32*, UTF-32LE*, UTF-32BE*
- **中文**: GBK, GB2312, GB18030, BIG5
- **日文**: Shift_JIS, EUC-JP
- **韩文**: EUC-KR
- **西欧**: ISO-8859-1, ISO-8859-2, ISO-8859-5, ISO-8859-15
- **Windows**: Windows-1250, Windows-1251, Windows-1252, Windows-1254
- **其他**: KOI8-R, CP866, Macintosh

*注: UTF-32 系列编码目前映射到 UTF-16 实现

## 工厂函数

库提供了多种预配置的工厂函数：

```go
// 基础工厂函数
processor := encoding.NewDefault()                    // 默认配置
processor := encoding.NewQuick()                      // 快速配置（与默认相同）
processor := encoding.NewForCLI()                     // 命令行工具优化
processor := encoding.NewForWebService()              // Web 服务优化
processor := encoding.NewForBatchProcessing()         // 批量处理优化

// 性能优化
processor := encoding.NewHighPerformance()            // 高性能配置
processor := encoding.NewMemoryEfficient()            // 内存高效配置

// 错误处理模式
processor := encoding.NewStrictMode()                 // 严格模式
processor := encoding.NewTolerantMode()               // 容错模式

// 自定义配置
processor := encoding.NewWithLogger(customLogger)     // 带自定义日志
processor := encoding.NewWithConfig(detCfg, convCfg)  // 完全自定义配置

// 专用处理器
streamProcessor := encoding.NewDefaultStream()        // 流处理器
fileProcessor := encoding.NewDefaultFile()            // 文件处理器

// 带监控
processor, metrics := encoding.NewDefaultWithMetrics() // 带性能监控
```

## 性能监控

```go
processor, metrics := encoding.NewDefaultWithMetrics()

// 执行一些操作...

stats := metrics.GetStats()
fmt.Printf("总操作数: %d\n", stats.TotalOperations)
fmt.Printf("成功率: %.2f%%\n", float64(stats.SuccessOperations)/float64(stats.TotalOperations)*100)
fmt.Printf("平均处理速度: %.2f MB/s\n", stats.AverageProcessingSpeed/1024/1024)
```

## 错误处理

库提供了结构化的错误类型：

```go
result, err := processor.DetectEncoding(data)
if err != nil {
    var encodingErr *encoding.EncodingError
    if errors.As(err, &encodingErr) {
        fmt.Printf("编码错误: 操作=%s, 编码=%s, 错误=%v\n", 
            encodingErr.Op, encodingErr.Encoding, encodingErr.Err)
    }
    
    switch {
    case errors.Is(err, encoding.ErrDetectionFailed):
        // 处理检测失败
    case errors.Is(err, encoding.ErrUnsupportedEncoding):
        // 处理不支持的编码
    default:
        // 处理其他错误
    }
}
```

## 核心接口

### 主要接口

- `Detector`: 编码检测功能
- `Converter`: 编码转换功能  
- `Processor`: 集成检测和转换功能
- `StreamProcessor`: 流式处理功能
- `FileProcessor`: 文件处理功能
- `MetricsCollector`: 性能监控功能

### 数据结构

- `DetectionResult`: 检测结果，包含编码名称、置信度等
- `ConvertResult`: 转换结果，包含转换后数据和元信息
- `FileProcessResult`: 文件处理结果
- `StreamResult`: 流处理结果
- `ProcessingStats`: 性能统计信息

## 设计原则

- **单一职责**: 专注于编码检测和转换，不涉及文件系统操作
- **接口设计**: 清晰的接口分离，便于测试和扩展
- **内存安全**: 大文件分块处理，避免内存溢出
- **并发安全**: 所有公共接口都是线程安全的
- **错误恢复**: 提供完善的错误处理和恢复机制

## 职责边界

✅ **库的职责**：
- 编码检测（字节数组、字符串、单个文件）
- 编码转换（字节数组、字符串、数据流）
- 单个文件的编码处理

❌ **不是库的职责**：
- 目录遍历和批量文件管理
- 文件系统监控
- 并发文件处理调度

这些功能应该在应用层实现。

## 性能特性

- **智能检测**: BOM 检测 → UTF-8 验证 → chardet 库检测
- **缓存机制**: 检测结果缓存，避免重复检测
- **内存优化**: 大文件分块处理，可配置内存限制
- **流式处理**: 支持无限大小文件的流式转换
- **并发安全**: 所有接口支持并发调用

## 示例

查看 [example/main.go](./example/main.go) 了解完整的使用示例。

## 文档

- [API 文档](./docs/api.md) - 完整的 API 参考
- [使用示例](./docs/examples.md) - 详细的使用场景
- [实施规划](./docs/implementation-plan.md) - 项目实施详情
- [测试策略](./docs/test-strategy.md) - 测试方法和策略

## 测试

```bash
# 运行所有测试
go test -v

# 运行示例
cd example
go run main.go

# 检查代码构建
go build ./...
```

## 依赖

- `github.com/saintfish/chardet` - 编码检测
- `golang.org/x/text/encoding` - 编码转换
- `golang.org/x/text/transform` - 转换框架

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！