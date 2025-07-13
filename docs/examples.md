# EncodingProcessor 使用示例

本文档提供了 EncodingProcessor 库的详细使用示例，涵盖从基础用法到高级场景的各种应用。

## 基础使用示例

### 1. 简单的编码检测

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
    
    // 待检测的文本数据
    data := []byte("这是一段中文文本")
    
    // 检测编码
    result, err := processor.DetectEncoding(data)
    if err != nil {
        log.Fatalf("检测失败: %v", err)
    }
    
    fmt.Printf("检测结果:\n")
    fmt.Printf("  编码: %s\n", result.Encoding)
    fmt.Printf("  置信度: %.2f\n", result.Confidence)
    if result.Language != "" {
        fmt.Printf("  语言: %s\n", result.Language)
    }
}
```

### 2. 编码转换

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/mirbf/encoding-processor"
)

func main() {
    processor := encoding.NewDefault()
    
    // GBK 编码的中文文本
    gbkData := []byte{0xd5, 0xe2, 0xca, 0xc7, 0xd6, 0xd0, 0xce, 0xc4} // "这是中文"
    
    // 转换为 UTF-8
    utf8Data, err := processor.ConvertToUTF8(gbkData, encoding.EncodingGBK)
    if err != nil {
        log.Fatalf("转换失败: %v", err)
    }
    
    fmt.Printf("转换结果: %s\n", string(utf8Data))
}
```

### 3. 智能转换（自动检测+转换）

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/mirbf/encoding-processor"
)

func main() {
    processor := encoding.NewDefault()
    
    // 未知编码的数据
    unknownData := []byte{0xd5, 0xe2, 0xca, 0xc7, 0xd6, 0xd0, 0xce, 0xc4}
    
    // 智能转换为 UTF-8
    result, err := processor.SmartConvert(unknownData, encoding.EncodingUTF8)
    if err != nil {
        log.Fatalf("智能转换失败: %v", err)
    }
    
    fmt.Printf("源编码: %s\n", result.SourceEncoding)
    fmt.Printf("转换结果: %s\n", string(result.Data))
    fmt.Printf("处理字节数: %d\n", result.BytesProcessed)
}
```

## 文件处理示例

### 4. 单文件转换

```go
package main

import (
    "log"
    
    "github.com/mirbf/encoding-processor"
)

func main() {
    processor := encoding.NewDefault()
    
    // 配置处理选项
    options := &encoding.ProcessOptions{
        TargetEncoding: encoding.EncodingUTF8,
        BackupOriginal: true,  // 备份原文件
        OverwriteFile:  false, // 不覆盖已存在文件
        MinConfidence:  0.8,   // 最小置信度
    }
    
    // 处理文件
    err := processor.ProcessFile("input_gbk.txt", "output_utf8.txt", options)
    if err != nil {
        log.Fatalf("文件处理失败: %v", err)
    }
    
    log.Println("文件转换完成")
}
```

### 5. 批量文件处理

```go
package main

import (
    "fmt"
    "log"
    "path/filepath"
    "sync"
    
    "github.com/mirbf/encoding-processor"
)

func main() {
    processor := encoding.NewDefault()
    
    files := []string{
        "file1.txt",
        "file2.txt", 
        "file3.txt",
    }
    
    options := &encoding.ProcessOptions{
        TargetEncoding: encoding.EncodingUTF8,
        BackupOriginal: true,
    }
    
    // 并发处理多个文件
    var wg sync.WaitGroup
    results := make(chan error, len(files))
    
    for _, file := range files {
        wg.Add(1)
        go func(filename string) {
            defer wg.Done()
            
            outputFile := fmt.Sprintf("utf8_%s", filepath.Base(filename))
            err := processor.ProcessFile(filename, outputFile, options)
            results <- err
        }(file)
    }
    
    wg.Wait()
    close(results)
    
    // 检查结果
    successCount := 0
    for err := range results {
        if err != nil {
            log.Printf("处理失败: %v", err)
        } else {
            successCount++
        }
    }
    
    fmt.Printf("成功处理 %d/%d 个文件\n", successCount, len(files))
}
```

## 高级配置示例

### 6. 自定义检测器配置

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/mirbf/encoding-processor"
)

func main() {
    // 自定义检测器配置
    detectorConfig := &encoding.DetectorConfig{
        SampleSize:     16384,  // 增大检测样本
        MinConfidence:  0.9,    // 提高置信度要求
        EnableCache:    true,   // 启用缓存
        SupportedEncodings: []string{
            encoding.EncodingUTF8,
            encoding.EncodingGBK,
            encoding.EncodingBIG5,
        },
    }
    
    // 自定义转换器配置
    converterConfig := &encoding.ConverterConfig{
        StrictMode:             false,  // 非严格模式
        InvalidCharReplacement: "□",   // 自定义替换字符
        BufferSize:            16384,  // 增大缓冲区
        EnableParallel:        true,   // 启用并行处理
    }
    
    // 创建自定义处理器
    processor := encoding.NewProcessor(detectorConfig, converterConfig)
    
    // 使用自定义处理器
    data := []byte("测试数据")
    result, err := processor.DetectEncoding(data)
    if err != nil {
        log.Fatalf("检测失败: %v", err)
    }
    
    fmt.Printf("检测结果: %s (%.2f)\n", result.Encoding, result.Confidence)
}
```

### 7. 流式处理大文件

```go
package main

import (
    "bufio"
    "fmt"
    "io"
    "log"
    "os"
    
    "github.com/mirbf/encoding-processor"
)

func main() {
    processor := encoding.NewDefault()
    
    // 打开大文件
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
    
    // 先检测文件开头的编码
    reader := bufio.NewReader(inputFile)
    sample := make([]byte, 8192)
    n, err := reader.Read(sample)
    if err != nil && err != io.EOF {
        log.Fatal(err)
    }
    
    result, err := processor.DetectEncoding(sample[:n])
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("检测到编码: %s\n", result.Encoding)
    
    // 重置文件指针
    inputFile.Seek(0, 0)
    
    // 流式转换处理
    bufferSize := 8192
    buffer := make([]byte, bufferSize)
    
    for {
        n, err := inputFile.Read(buffer)
        if n == 0 {
            break
        }
        
        // 转换当前块
        converted, convErr := processor.Convert(buffer[:n], result.Encoding, encoding.EncodingUTF8)
        if convErr != nil {
            log.Printf("转换块失败: %v", convErr)
            continue
        }
        
        // 写入输出文件
        _, writeErr := outputFile.Write(converted)
        if writeErr != nil {
            log.Fatal(writeErr)
        }
        
        if err == io.EOF {
            break
        }
        if err != nil {
            log.Fatal(err)
        }
    }
    
    fmt.Println("大文件转换完成")
}
```

## 错误处理示例

### 8. 完整的错误处理

```go
package main

import (
    "errors"
    "fmt"
    "log"
    
    "github.com/mirbf/encoding-processor"
)

func main() {
    processor := encoding.NewDefault()
    
    data := []byte("test data")
    
    // 检测编码并处理各种错误
    result, err := processor.DetectEncoding(data)
    if err != nil {
        var encodingErr *encoding.EncodingError
        if errors.As(err, &encodingErr) {
            fmt.Printf("编码错误: 操作=%s, 编码=%s, 错误=%v\n", 
                encodingErr.Op, encodingErr.Encoding, encodingErr.Err)
        }
        
        switch {
        case errors.Is(err, encoding.ErrDetectionFailed):
            log.Println("检测失败，使用默认编码")
            result = &encoding.DetectionResult{
                Encoding:   encoding.EncodingUTF8,
                Confidence: 0.0,
            }
        case errors.Is(err, encoding.ErrInvalidInput):
            log.Fatal("输入数据无效")
        default:
            log.Fatalf("未知错误: %v", err)
        }
    }
    
    // 检查置信度
    if result.Confidence < 0.8 {
        fmt.Printf("警告: 检测置信度较低 (%.2f)\n", result.Confidence)
    }
    
    fmt.Printf("检测成功: %s\n", result.Encoding)
}
```

### 9. 文件名编码处理

```go
package main

import (
    "fmt"
    "log"
    "path/filepath"
    
    "github.com/mirbf/encoding-processor"
)

func main() {
    processor := encoding.NewDefault()
    
    // 处理可能包含非 UTF-8 编码的文件名
    corruptedFileName := "测试文件.txt" // 假设这是损坏的编码
    
    // 检测文件名编码
    result, err := processor.DetectEncoding([]byte(corruptedFileName))
    if err != nil {
        log.Printf("文件名编码检测失败: %v", err)
    }
    
    // 转换文件名为 UTF-8
    if result.Encoding != encoding.EncodingUTF8 {
        utf8FileName, err := processor.ConvertString(corruptedFileName, result.Encoding, encoding.EncodingUTF8)
        if err != nil {
            log.Printf("文件名转换失败: %v", err)
        } else {
            fmt.Printf("原文件名: %s\n", corruptedFileName)
            fmt.Printf("转换后: %s\n", utf8FileName)
        }
    }
}
```

### 10. 性能监控示例

```go
package main

import (
    "fmt"
    "log"
    "time"
    
    "github.com/mirbf/encoding-processor"
)

func main() {
    processor := encoding.NewDefault()
    
    data := make([]byte, 1024*1024) // 1MB 测试数据
    for i := range data {
        data[i] = byte('A' + (i % 26))
    }
    
    // 性能测试 - 检测
    start := time.Now()
    result, err := processor.DetectEncoding(data)
    detectTime := time.Since(start)
    
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("检测性能:\n")
    fmt.Printf("  数据大小: %d 字节\n", len(data))
    fmt.Printf("  检测时间: %v\n", detectTime)
    fmt.Printf("  检测速度: %.2f MB/s\n", float64(len(data))/detectTime.Seconds()/1024/1024)
    
    // 性能测试 - 转换
    start = time.Now()
    converted, err := processor.ConvertToUTF8(data, result.Encoding)
    convertTime := time.Since(start)
    
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("转换性能:\n")
    fmt.Printf("  转换时间: %v\n", convertTime)
    fmt.Printf("  转换速度: %.2f MB/s\n", float64(len(converted))/convertTime.Seconds()/1024/1024)
}
```

## 集成示例

### 11. Web 应用集成

```go
package main

import (
    "fmt"
    "io"
    "net/http"
    
    "github.com/mirbf/encoding-processor"
)

func uploadHandler(w http.ResponseWriter, r *http.Request) {
    processor := encoding.NewDefault()
    
    // 读取上传的文件
    file, header, err := r.FormFile("textfile")
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    defer file.Close()
    
    // 读取文件内容
    data, err := io.ReadAll(file)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    // 智能转换为 UTF-8
    result, err := processor.SmartConvert(data, encoding.EncodingUTF8)
    if err != nil {
        http.Error(w, fmt.Sprintf("转换失败: %v", err), http.StatusInternalServerError)
        return
    }
    
    // 设置响应头
    w.Header().Set("Content-Type", "text/plain; charset=utf-8")
    w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"utf8_%s\"", header.Filename))
    
    // 返回转换后的内容
    w.Write(result.Data)
}

func main() {
    http.HandleFunc("/upload", uploadHandler)
    fmt.Println("服务器启动在 :8080")
    http.ListenAndServe(":8080", nil)
}
```

## 高级应用场景

### 12. 命令行工具集成（简化版）

```go
package main

import (
    "flag"
    "fmt"
    "log"
    "os"
    "path/filepath"
    "strings"
    "sync"
    
    "github.com/mirbf/encoding-processor"
)

func main() {
    var (
        input          = flag.String("input", "", "输入文件或目录")
        output         = flag.String("output", "", "输出目录")
        targetEncoding = flag.String("encoding", "UTF-8", "目标编码")
        recursive      = flag.Bool("recursive", false, "递归处理目录")
        backup         = flag.Bool("backup", true, "创建备份文件")
        dryRun         = flag.Bool("dry-run", false, "试运行模式")
        pattern        = flag.String("pattern", "*.txt", "文件匹配模式")
        concurrency    = flag.Int("concurrency", 4, "并发处理数")
        verbose        = flag.Bool("verbose", false, "详细输出")
    )
    flag.Parse()

    if *input == "" {
        log.Fatal("必须指定输入文件或目录")
    }

    // 检查输入是文件还是目录
    info, err := os.Stat(*input)
    if err != nil {
        log.Fatalf("无法访问输入路径: %v", err)
    }

    if info.IsDir() {
        // 目录处理（应用层实现）
        processDirectory(*input, *output, *targetEncoding, *recursive, 
                        *backup, *dryRun, *pattern, *concurrency, *verbose)
    } else {
        // 单文件处理
        processSingleFile(*input, *output, *targetEncoding, *backup, 
                         *dryRun, *verbose)
    }
}

func processDirectory(inputDir, outputDir, targetEncoding string, 
                     recursive, backup, dryRun bool, pattern string, 
                     concurrency int, verbose bool) {
    
    // 收集要处理的文件（应用层职责）
    var files []string
    err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        
        if info.IsDir() {
            if !recursive && path != inputDir {
                return filepath.SkipDir
            }
            return nil
        }
        
        // 检查文件是否匹配模式
        matched, err := filepath.Match(pattern, filepath.Base(path))
        if err != nil {
            return err
        }
        
        if matched {
            files = append(files, path)
        }
        
        return nil
    })
    
    if err != nil {
        log.Fatalf("遍历目录失败: %v", err)
    }
    
    if len(files) == 0 {
        fmt.Printf("没有找到匹配模式 '%s' 的文件\n", pattern)
        return
    }
    
    fmt.Printf("找到 %d 个文件待处理\n", len(files))
    
    // 创建文件处理器
    fileProcessor := encoding.NewDefaultFile()
    
    // 并发处理文件（应用层实现）
    var wg sync.WaitGroup
    semaphore := make(chan struct{}, concurrency)
    results := make(chan error, len(files))
    
    for _, file := range files {
        wg.Add(1)
        go func(filename string) {
            defer wg.Done()
            semaphore <- struct{}{}
            defer func() { <-semaphore }()
            
            if verbose {
                fmt.Printf("处理中: %s\n", filename)
            }
            
            // 生成输出文件路径
            var outputFile string
            if outputDir == "" {
                outputFile = filename // 就地处理
            } else {
                outputFile = filepath.Join(outputDir, filepath.Base(filename))
            }
            
            options := &encoding.FileProcessOptions{
                TargetEncoding:    targetEncoding,
                CreateBackup:      backup,
                OverwriteExisting: false,
                DryRun:           dryRun,
            }
            
            var err error
            if outputDir == "" {
                // 就地处理
                _, err = fileProcessor.ProcessFileInPlace(filename, options)
            } else {
                // 输出到指定目录
                _, err = fileProcessor.ProcessFile(filename, outputFile, options)
            }
            
            results <- err
        }(file)
    }
    
    wg.Wait()
    close(results)
    
    // 统计结果
    successCount := 0
    errorCount := 0
    for err := range results {
        if err != nil {
            log.Printf("处理失败: %v", err)
            errorCount++
        } else {
            successCount++
        }
    }
    
    fmt.Printf("\n📊 处理完成:\n")
    fmt.Printf("  成功: %d 个文件\n", successCount)
    fmt.Printf("  失败: %d 个文件\n", errorCount)
    fmt.Printf("  总计: %d 个文件\n", len(files))
}

func processSingleFile(inputFile, outputDir, targetEncoding string, 
                      backup, dryRun, verbose bool) {
    
    fileProcessor := encoding.NewDefaultFile()
    
    if verbose {
        fmt.Printf("处理文件: %s\n", inputFile)
    }
    
    options := &encoding.FileProcessOptions{
        TargetEncoding:    targetEncoding,
        CreateBackup:      backup,
        OverwriteExisting: false,
        DryRun:           dryRun,
    }
    
    var result *encoding.FileProcessResult
    var err error
    
    if outputDir == "" {
        // 就地处理
        result, err = fileProcessor.ProcessFileInPlace(inputFile, options)
    } else {
        // 输出到目录
        outputFile := filepath.Join(outputDir, filepath.Base(inputFile))
        result, err = fileProcessor.ProcessFile(inputFile, outputFile, options)
    }
    
    if err != nil {
        log.Fatalf("处理失败: %v", err)
    }
    
    if dryRun {
        fmt.Printf("🔍 试运行: %s -> %s\n", result.InputFile, result.OutputFile)
    } else {
        fmt.Printf("✅ 处理完成: %s -> %s\n", result.InputFile, result.OutputFile)
        if result.BackupFile != "" {
            fmt.Printf("📦 备份文件: %s\n", result.BackupFile)
        }
    }
}
```

### 13. 数据库文本字段处理

```go
package main

import (
    "database/sql"
    "fmt"
    "log"
    
    _ "github.com/go-sql-driver/mysql"
    "github.com/mirbf/encoding-processor"
)

type TextRecord struct {
    ID      int64
    Content []byte
    Encoding string
}

func main() {
    // 连接数据库
    db, err := sql.Open("mysql", "user:password@tcp(localhost:3306)/testdb")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    processor := encoding.NewDefault()
    
    // 处理数据库中的文本字段
    err = processTextFields(db, processor)
    if err != nil {
        log.Fatalf("处理数据库文本字段失败: %v", err)
    }
}

func processTextFields(db *sql.DB, processor encoding.Processor) error {
    // 查询需要处理的记录
    rows, err := db.Query(`
        SELECT id, content, detected_encoding 
        FROM text_data 
        WHERE processed = 0 
        ORDER BY id LIMIT 1000
    `)
    if err != nil {
        return err
    }
    defer rows.Close()
    
    var records []TextRecord
    for rows.Next() {
        var record TextRecord
        err := rows.Scan(&record.ID, &record.Content, &record.Encoding)
        if err != nil {
            return err
        }
        records = append(records, record)
    }
    
    fmt.Printf("处理 %d 条记录\n", len(records))
    
    // 开始事务
    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()
    
    successCount := 0
    errorCount := 0
    
    for _, record := range records {
        err := processRecord(tx, processor, record)
        if err != nil {
            log.Printf("处理记录 %d 失败: %v", record.ID, err)
            errorCount++
            continue
        }
        successCount++
    }
    
    // 提交事务
    if err := tx.Commit(); err != nil {
        return err
    }
    
    fmt.Printf("处理完成: 成功 %d, 失败 %d\n", successCount, errorCount)
    return nil
}

func processRecord(tx *sql.Tx, processor encoding.Processor, record TextRecord) error {
    // 检测编码（如果未知）
    var sourceEncoding string
    if record.Encoding == "" || record.Encoding == "unknown" {
        result, err := processor.DetectEncoding(record.Content)
        if err != nil {
            return fmt.Errorf("编码检测失败: %w", err)
        }
        sourceEncoding = result.Encoding
    } else {
        sourceEncoding = record.Encoding
    }
    
    // 转换为 UTF-8
    utf8Content, err := processor.ConvertToUTF8(record.Content, sourceEncoding)
    if err != nil {
        return fmt.Errorf("编码转换失败: %w", err)
    }
    
    // 更新数据库记录
    _, err = tx.Exec(`
        UPDATE text_data 
        SET content = ?, 
            original_encoding = ?, 
            current_encoding = 'UTF-8',
            processed = 1,
            processed_at = NOW()
        WHERE id = ?
    `, utf8Content, sourceEncoding, record.ID)
    
    if err != nil {
        return fmt.Errorf("更新数据库失败: %w", err)
    }
    
    return nil
}
```

### 14. 日志文件编码处理

```go
package main

import (
    "bufio"
    "context"
    "fmt"
    "io"
    "log"
    "os"
    "path/filepath"
    "strings"
    "time"
    
    "github.com/mirbf/encoding-processor"
)

type LogProcessor struct {
    processor encoding.StreamProcessor
    config    *LogProcessorConfig
}

type LogProcessorConfig struct {
    InputDir       string
    OutputDir      string
    FilePattern    string
    TargetEncoding string
    RotateDaily    bool
    CompressOld    bool
    RetentionDays  int
}

func NewLogProcessor(config *LogProcessorConfig) *LogProcessor {
    return &LogProcessor{
        processor: encoding.NewDefaultStream(),
        config:    config,
    }
}

func (lp *LogProcessor) ProcessLogs(ctx context.Context) error {
    // 扫描日志文件
    logFiles, err := lp.scanLogFiles()
    if err != nil {
        return fmt.Errorf("扫描日志文件失败: %w", err)
    }
    
    log.Printf("找到 %d 个日志文件", len(logFiles))
    
    for _, logFile := range logFiles {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            if err := lp.processLogFile(ctx, logFile); err != nil {
                log.Printf("处理日志文件 %s 失败: %v", logFile, err)
                continue
            }
            log.Printf("成功处理日志文件: %s", logFile)
        }
    }
    
    // 清理旧文件
    if lp.config.RetentionDays > 0 {
        if err := lp.cleanupOldFiles(); err != nil {
            log.Printf("清理旧文件失败: %v", err)
        }
    }
    
    return nil
}

func (lp *LogProcessor) scanLogFiles() ([]string, error) {
    var files []string
    
    err := filepath.Walk(lp.config.InputDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        
        if info.IsDir() {
            return nil
        }
        
        // 检查文件模式匹配
        matched, err := filepath.Match(lp.config.FilePattern, filepath.Base(path))
        if err != nil {
            return err
        }
        
        if matched {
            files = append(files, path)
        }
        
        return nil
    })
    
    return files, err
}

func (lp *LogProcessor) processLogFile(ctx context.Context, inputFile string) error {
    // 打开输入文件
    input, err := os.Open(inputFile)
    if err != nil {
        return err
    }
    defer input.Close()
    
    // 生成输出文件路径
    outputFile := lp.generateOutputPath(inputFile)
    
    // 确保输出目录存在
    if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
        return err
    }
    
    // 创建输出文件
    output, err := os.Create(outputFile)
    if err != nil {
        return err
    }
    defer output.Close()
    
    // 流式处理日志文件
    options := &encoding.StreamOptions{
        SourceEncoding: "", // 自动检测
        TargetEncoding: lp.config.TargetEncoding,
        BufferSize:     64 * 1024, // 64KB 缓冲区
        StrictMode:     false,      // 宽松模式，跳过无效字符
    }
    
    result, err := lp.processor.ProcessReaderWriter(ctx, input, output, options)
    if err != nil {
        return err
    }
    
    log.Printf("日志处理统计: 读取 %d 字节, 写入 %d 字节, 源编码: %s",
               result.BytesRead, result.BytesWritten, result.SourceEncoding)
    
    return nil
}

func (lp *LogProcessor) generateOutputPath(inputFile string) string {
    // 获取相对于输入目录的路径
    relPath, _ := filepath.Rel(lp.config.InputDir, inputFile)
    
    // 如果需要按日期轮转
    if lp.config.RotateDaily {
        now := time.Now()
        dateDir := now.Format("2006-01-02")
        relPath = filepath.Join(dateDir, relPath)
    }
    
    // 添加编码后缀
    ext := filepath.Ext(relPath)
    base := strings.TrimSuffix(relPath, ext)
    relPath = base + "_utf8" + ext
    
    return filepath.Join(lp.config.OutputDir, relPath)
}

func (lp *LogProcessor) cleanupOldFiles() error {
    cutoff := time.Now().AddDate(0, 0, -lp.config.RetentionDays)
    
    return filepath.Walk(lp.config.OutputDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        
        if info.IsDir() {
            return nil
        }
        
        if info.ModTime().Before(cutoff) {
            log.Printf("删除过期文件: %s", path)
            return os.Remove(path)
        }
        
        return nil
    })
}

func main() {
    config := &LogProcessorConfig{
        InputDir:       "/var/log/app",
        OutputDir:      "/var/log/app_utf8",
        FilePattern:    "*.log",
        TargetEncoding: "UTF-8",
        RotateDaily:    true,
        CompressOld:    true,
        RetentionDays:  30,
    }
    
    processor := NewLogProcessor(config)
    
    ctx := context.Background()
    if err := processor.ProcessLogs(ctx); err != nil {
        log.Fatalf("日志处理失败: %v", err)
    }
    
    log.Println("日志处理完成")
}
```

### 15. 微服务编码处理

```go
package main

import (
    "context"
    "encoding/base64"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    
    "github.com/gin-gonic/gin"
    "github.com/mirbf/encoding-processor"
)

type EncodingService struct {
    processor encoding.Processor
    metrics   encoding.MetricsCollector
}

func NewEncodingService() *EncodingService {
    processor, metrics := encoding.NewDefaultWithMetrics()
    return &EncodingService{
        processor: processor,
        metrics:   metrics,
    }
}

// DetectRequest 检测请求
type DetectRequest struct {
    Data   string `json:"data" binding:"required"`
    Base64 bool   `json:"base64,omitempty"`
}

// DetectResponse 检测响应
type DetectResponse struct {
    Encoding   string  `json:"encoding"`
    Confidence float64 `json:"confidence"`
    Language   string  `json:"language,omitempty"`
}

// ConvertRequest 转换请求
type ConvertRequest struct {
    Data           string `json:"data" binding:"required"`
    SourceEncoding string `json:"source_encoding,omitempty"`
    TargetEncoding string `json:"target_encoding" binding:"required"`
    Base64         bool   `json:"base64,omitempty"`
}

// ConvertResponse 转换响应
type ConvertResponse struct {
    Data           string `json:"data"`
    SourceEncoding string `json:"source_encoding"`
    TargetEncoding string `json:"target_encoding"`
    BytesProcessed int64  `json:"bytes_processed"`
}

// MetricsResponse 指标响应
type MetricsResponse struct {
    TotalRequests      int64             `json:"total_requests"`
    SuccessRequests    int64             `json:"success_requests"`
    FailedRequests     int64             `json:"failed_requests"`
    AverageResponseTime float64          `json:"average_response_time_ms"`
    EncodingDistribution map[string]int64 `json:"encoding_distribution"`
}

func (es *EncodingService) SetupRoutes() *gin.Engine {
    r := gin.Default()
    
    // 中间件
    r.Use(es.metricsMiddleware())
    r.Use(gin.Recovery())
    
    api := r.Group("/api/v1")
    {
        api.POST("/detect", es.detectEncoding)
        api.POST("/convert", es.convertEncoding)
        api.GET("/metrics", es.getMetrics)
        api.GET("/health", es.healthCheck)
    }
    
    return r
}

func (es *EncodingService) detectEncoding(c *gin.Context) {
    var req DetectRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // 解码数据
    data, err := es.decodeData(req.Data, req.Base64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "数据解码失败"})
        return
    }
    
    // 检测编码
    result, err := es.processor.DetectEncoding(data)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "编码检测失败"})
        return
    }
    
    response := DetectResponse{
        Encoding:   result.Encoding,
        Confidence: result.Confidence,
        Language:   result.Language,
    }
    
    c.JSON(http.StatusOK, response)
}

func (es *EncodingService) convertEncoding(c *gin.Context) {
    var req ConvertRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // 解码数据
    data, err := es.decodeData(req.Data, req.Base64)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "数据解码失败"})
        return
    }
    
    var result *encoding.ConvertResult
    
    if req.SourceEncoding == "" {
        // 智能转换（自动检测源编码）
        result, err = es.processor.SmartConvert(data, req.TargetEncoding)
    } else {
        // 指定源编码转换
        convertedData, convertErr := es.processor.Convert(data, req.SourceEncoding, req.TargetEncoding)
        if convertErr != nil {
            err = convertErr
        } else {
            result = &encoding.ConvertResult{
                Data:           convertedData,
                SourceEncoding: req.SourceEncoding,
                TargetEncoding: req.TargetEncoding,
                BytesProcessed: int64(len(data)),
            }
        }
    }
    
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "编码转换失败"})
        return
    }
    
    // 编码输出数据
    outputData, err := es.encodeData(result.Data, req.Base64)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "数据编码失败"})
        return
    }
    
    response := ConvertResponse{
        Data:           outputData,
        SourceEncoding: result.SourceEncoding,
        TargetEncoding: result.TargetEncoding,
        BytesProcessed: result.BytesProcessed,
    }
    
    c.JSON(http.StatusOK, response)
}

func (es *EncodingService) getMetrics(c *gin.Context) {
    stats := es.metrics.GetStats()
    
    response := MetricsResponse{
        TotalRequests:        stats.TotalOperations,
        SuccessRequests:      stats.SuccessOperations,
        FailedRequests:       stats.FailedOperations,
        AverageResponseTime:  float64(stats.TotalProcessingTime.Milliseconds()) / float64(stats.TotalOperations),
        EncodingDistribution: stats.EncodingDistribution,
    }
    
    c.JSON(http.StatusOK, response)
}

func (es *EncodingService) healthCheck(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{
        "status":    "healthy",
        "timestamp": time.Now().Unix(),
        "version":   "1.0.0",
    })
}

func (es *EncodingService) metricsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()
        
        c.Next()
        
        duration := time.Since(start)
        es.metrics.RecordOperation(c.Request.URL.Path, duration)
    }
}

func (es *EncodingService) decodeData(data string, isBase64 bool) ([]byte, error) {
    if isBase64 {
        return base64.StdEncoding.DecodeString(data)
    }
    return []byte(data), nil
}

func (es *EncodingService) encodeData(data []byte, toBase64 bool) (string, error) {
    if toBase64 {
        return base64.StdEncoding.EncodeToString(data), nil
    }
    return string(data), nil
}

func main() {
    service := NewEncodingService()
    router := service.SetupRoutes()
    
    fmt.Println("编码处理服务启动在端口 8080")
    router.Run(":8080")
}
```

### 16. 实时文件监控和处理

```go
package main

import (
    "context"
    "log"
    "os"
    "path/filepath"
    
    "github.com/fsnotify/fsnotify"
    "github.com/mirbf/encoding-processor"
)

type FileWatcher struct {
    processor encoding.Processor
    watcher   *fsnotify.Watcher
    config    *WatcherConfig
}

type WatcherConfig struct {
    WatchDirs      []string
    OutputDir      string
    FilePatterns   []string
    TargetEncoding string
    ExcludeDirs    []string
}

func NewFileWatcher(config *WatcherConfig) (*FileWatcher, error) {
    watcher, err := fsnotify.NewWatcher()
    if err != nil {
        return nil, err
    }
    
    return &FileWatcher{
        processor: encoding.NewDefault(),
        watcher:   watcher,
        config:    config,
    }, nil
}

func (fw *FileWatcher) Start(ctx context.Context) error {
    // 添加监控目录
    for _, dir := range fw.config.WatchDirs {
        if err := fw.addWatchDir(dir); err != nil {
            return err
        }
    }
    
    log.Printf("开始监控 %d 个目录", len(fw.config.WatchDirs))
    
    for {
        select {
        case <-ctx.Done():
            return ctx.Err()
            
        case event, ok := <-fw.watcher.Events:
            if !ok {
                return nil
            }
            
            if err := fw.handleEvent(event); err != nil {
                log.Printf("处理文件事件失败: %v", err)
            }
            
        case err, ok := <-fw.watcher.Errors:
            if !ok {
                return nil
            }
            log.Printf("文件监控错误: %v", err)
        }
    }
}

func (fw *FileWatcher) addWatchDir(dir string) error {
    return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        
        if info.IsDir() {
            // 检查是否是排除目录
            for _, excludeDir := range fw.config.ExcludeDirs {
                if matched, _ := filepath.Match(excludeDir, filepath.Base(path)); matched {
                    return filepath.SkipDir
                }
            }
            
            return fw.watcher.Add(path)
        }
        
        return nil
    })
}

func (fw *FileWatcher) handleEvent(event fsnotify.Event) error {
    if event.Op&fsnotify.Create == fsnotify.Create ||
       event.Op&fsnotify.Write == fsnotify.Write {
        
        // 检查文件是否匹配模式
        if fw.shouldProcessFile(event.Name) {
            return fw.processFile(event.Name)
        }
    }
    
    return nil
}

func (fw *FileWatcher) shouldProcessFile(filename string) bool {
    for _, pattern := range fw.config.FilePatterns {
        if matched, _ := filepath.Match(pattern, filepath.Base(filename)); matched {
            return true
        }
    }
    return false
}

func (fw *FileWatcher) processFile(filename string) error {
    log.Printf("处理新文件: %s", filename)
    
    options := &encoding.FileProcessOptions{
        TargetEncoding:    fw.config.TargetEncoding,
        CreateBackup:      true,
        OverwriteExisting: false,
    }
    
    // 生成输出文件路径
    outputFile := filepath.Join(fw.config.OutputDir, filepath.Base(filename))
    
    // 确保输出目录存在
    if err := os.MkdirAll(fw.config.OutputDir, 0755); err != nil {
        return err
    }
    
    // 使用文件处理器处理文件
    fileProcessor := encoding.NewDefaultFile()
    result, err := fileProcessor.ProcessFile(filename, outputFile, options)
    if err != nil {
        return err
    }
    
    log.Printf("文件处理完成: %s -> %s", result.InputFile, result.OutputFile)
    return nil
}

func (fw *FileWatcher) Stop() error {
    return fw.watcher.Close()
}

func main() {
    config := &WatcherConfig{
        WatchDirs: []string{
            "/data/incoming",
            "/data/uploads",
        },
        OutputDir:      "/data/processed",
        FilePatterns:   []string{"*.txt", "*.log", "*.csv"},
        TargetEncoding: "UTF-8",
        ExcludeDirs:    []string{".git", ".svn", "node_modules"},
    }
    
    watcher, err := NewFileWatcher(config)
    if err != nil {
        log.Fatal(err)
    }
    defer watcher.Stop()
    
    ctx := context.Background()
    if err := watcher.Start(ctx); err != nil {
        log.Fatal(err)
    }
}
```

## 重要说明

**本示例文档展示 EncodingProcessor 库的正确用法**：
- ✅ 专注于编码检测和转换
- ✅ 处理数据流、单个文件、字节数组、字符串  
- ❌ 不包含目录遍历功能（应用层职责）

目录批量处理应该在应用层实现，而不是编码库的职责。

---

这些高级应用场景展示了 EncodingProcessor 库在实际生产环境中的正确应用：

1. **命令行工具**：应用层处理目录遍历，库负责单文件编码转换
2. **数据库集成**：处理数据库中的文本字段编码问题
3. **日志处理**：处理单个日志文件的编码转换
4. **微服务**：RESTful API 服务，提供编码检测和转换功能
5. **文件监控**：监控单个文件变化，使用库进行编码处理

这些示例正确地展示了编码处理库的职责边界，避免了职责混乱的问题。