# EncodingProcessor 测试策略文档

## 测试概述

### 测试目标

1. **功能正确性**：确保所有编码检测和转换功能正确工作
2. **性能达标**：满足性能基准要求
3. **稳定性保证**：在各种边界条件下稳定运行
4. **兼容性验证**：在不同平台和 Go 版本下正常工作

### 测试层次

```
┌─────────────────────────────────────┐
│           E2E 测试                   │  ← 端到端集成测试
├─────────────────────────────────────┤
│          集成测试                    │  ← 组件间协作测试
├─────────────────────────────────────┤
│          单元测试                    │  ← 单个函数/方法测试
├─────────────────────────────────────┤
│        基准测试                      │  ← 性能基准测试
├─────────────────────────────────────┤
│        模糊测试                      │  ← 异常数据测试
└─────────────────────────────────────┘
```

## 单元测试策略

### 测试覆盖率目标

- **代码覆盖率**：≥ 90%
- **分支覆盖率**：≥ 85%
- **函数覆盖率**：≥ 95%

### 核心模块测试

#### 1. 编码检测模块

```go
// detector_test.go
func TestDetectEncoding(t *testing.T) {
    tests := []struct {
        name       string
        input      []byte
        expected   string
        confidence float64
    }{
        {
            name:       "UTF-8 Chinese",
            input:      []byte("这是中文测试"),
            expected:   "UTF-8",
            confidence: 0.99,
        },
        {
            name:       "GBK Chinese",
            input:      []byte{0xd5, 0xe2, 0xca, 0xc7, 0xd6, 0xd0, 0xce, 0xc4},
            expected:   "GBK",
            confidence: 0.95,
        },
        {
            name:       "Empty input",
            input:      []byte{},
            expected:   "",
            confidence: 0.0,
        },
    }
    
    detector := NewDetector()
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := detector.DetectEncoding(tt.input)
            assert.NoError(t, err)
            assert.Equal(t, tt.expected, result.Encoding)
            assert.GreaterOrEqual(t, result.Confidence, tt.confidence)
        })
    }
}

func TestDetectEncodingCache(t *testing.T) {
    config := &DetectorConfig{
        EnableCache: true,
        CacheSize:   100,
        CacheTTL:    time.Minute,
    }
    detector := NewDetector(config)
    
    input := []byte("测试缓存功能")
    
    // 第一次检测
    start := time.Now()
    result1, err := detector.DetectEncoding(input)
    duration1 := time.Since(start)
    assert.NoError(t, err)
    
    // 第二次检测（应该命中缓存）
    start = time.Now()
    result2, err := detector.DetectEncoding(input)
    duration2 := time.Since(start)
    assert.NoError(t, err)
    
    assert.Equal(t, result1.Encoding, result2.Encoding)
    assert.Less(t, duration2, duration1/2) // 缓存应该明显更快
}
```

#### 2. 编码转换模块

```go
// converter_test.go
func TestConvert(t *testing.T) {
    tests := []struct {
        name     string
        input    []byte
        from     string
        to       string
        expected string
    }{
        {
            name:     "GBK to UTF-8",
            input:    []byte{0xd5, 0xe2, 0xca, 0xc7, 0xd6, 0xd0, 0xce, 0xc4},
            from:     "GBK",
            to:       "UTF-8",
            expected: "这是中文",
        },
        {
            name:     "UTF-8 to UTF-8",
            input:    []byte("这是中文"),
            from:     "UTF-8",
            to:       "UTF-8",
            expected: "这是中文",
        },
    }
    
    converter := NewConverter()
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := converter.Convert(tt.input, tt.from, tt.to)
            assert.NoError(t, err)
            assert.Equal(t, tt.expected, string(result))
        })
    }
}

func TestConvertWithInvalidCharacters(t *testing.T) {
    converter := NewConverter(&ConverterConfig{
        StrictMode:             false,
        InvalidCharReplacement: "?",
    })
    
    // 包含无效字节的输入
    input := []byte{0xff, 0xfe, 0x41, 0x42}
    result, err := converter.ConvertToUTF8(input, "UTF-8")
    
    assert.NoError(t, err)
    assert.Contains(t, string(result), "?") // 应该包含替换字符
}
```

#### 3. 文件处理模块

```go
// file_processor_test.go
func TestProcessFileInPlace(t *testing.T) {
    // 创建临时测试文件
    tempDir := t.TempDir()
    testFile := filepath.Join(tempDir, "test.txt")
    
    // 写入 GBK 编码的测试数据
    gbkData := []byte{0xd5, 0xe2, 0xca, 0xc7, 0xd6, 0xd0, 0xce, 0xc4}
    err := os.WriteFile(testFile, gbkData, 0644)
    require.NoError(t, err)
    
    processor := NewProcessor(NewProcessorConfig())
    
    options := &ProcessOptions{
        TargetEncoding:  EncodingUTF8,
        OverwriteSource: true,
        BackupOriginal:  true,
        BackupSuffix:    ".bak",
    }
    
    backupInfo, err := processor.ProcessFileInPlace(testFile, options)
    require.NoError(t, err)
    require.NotNil(t, backupInfo)
    
    // 验证备份文件存在
    assert.FileExists(t, backupInfo.BackupFile)
    
    // 验证原文件已被转换
    convertedData, err := os.ReadFile(testFile)
    require.NoError(t, err)
    assert.Equal(t, "这是中文", string(convertedData))
    
    // 验证备份文件内容正确
    backupData, err := os.ReadFile(backupInfo.BackupFile)
    require.NoError(t, err)
    assert.Equal(t, gbkData, backupData)
}

func TestProcessFileToDir(t *testing.T) {
    tempDir := t.TempDir()
    inputFile := filepath.Join(tempDir, "input.txt")
    outputDir := filepath.Join(tempDir, "output")
    
    // 创建测试文件
    err := os.WriteFile(inputFile, []byte("测试内容"), 0644)
    require.NoError(t, err)
    
    processor := NewProcessor(NewProcessorConfig())
    
    options := &ProcessOptions{
        TargetEncoding:   EncodingUTF8,
        OverwriteSource:  false,
        OutputDir:        outputDir,
        CreateOutputDir:  true,
    }
    
    outputFile, err := processor.ProcessFileToDir(inputFile, outputDir, options)
    require.NoError(t, err)
    
    // 验证输出目录和文件存在
    assert.DirExists(t, outputDir)
    assert.FileExists(t, outputFile)
    
    // 验证文件内容
    content, err := os.ReadFile(outputFile)
    require.NoError(t, err)
    assert.Equal(t, "测试内容", string(content))
}
```

## 集成测试策略

### 组件集成测试

```go
// integration_test.go
func TestDetectorConverterIntegration(t *testing.T) {
    processor := NewDefault()
    
    testCases := []struct {
        name     string
        input    []byte
        expected string
    }{
        {"GBK", []byte{0xd5, 0xe2, 0xca, 0xc7}, "这是"},
        {"BIG5", []byte{0xa7, 0x41, 0xa6, 0x62}, "我们"},
        {"Shift_JIS", []byte{0x82, 0xb1, 0x82, 0xf1}, "こん"},
    }
    
    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // 智能转换（检测+转换）
            result, err := processor.SmartConvert(tc.input, EncodingUTF8)
            assert.NoError(t, err)
            assert.Equal(t, tc.expected, string(result.Data))
        })
    }
}

func TestBatchProcessingIntegration(t *testing.T) {
    tempDir := t.TempDir()
    batchProcessor := NewDefaultBatch()
    
    // 创建多个测试文件
    files := []string{
        filepath.Join(tempDir, "file1.txt"),
        filepath.Join(tempDir, "file2.txt"),
        filepath.Join(tempDir, "file3.txt"),
    }
    
    for i, file := range files {
        content := fmt.Sprintf("文件内容 %d", i+1)
        err := os.WriteFile(file, []byte(content), 0644)
        require.NoError(t, err)
    }
    
    options := &BatchOptions{
        ProcessOptions: ProcessOptions{
            TargetEncoding:  EncodingUTF8,
            OverwriteSource: false,
            OutputDir:       filepath.Join(tempDir, "output"),
        },
        MaxConcurrency: 2,
        StopOnError:    false,
    }
    
    ctx := context.Background()
    results := batchProcessor.ProcessFiles(ctx, files, options)
    
    successCount := 0
    for result := range results {
        if result.Success {
            successCount++
            assert.FileExists(t, result.OutputFile)
        }
    }
    
    assert.Equal(t, len(files), successCount)
}
```

## 性能测试策略

### 基准测试

```go
// benchmark_test.go
func BenchmarkDetectEncoding(b *testing.B) {
    detector := NewDetector()
    testData := []byte("这是一段用于性能测试的中文文本内容，包含足够的字符用于编码检测")
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := detector.DetectEncoding(testData)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkConvertLargeFile(b *testing.B) {
    converter := NewConverter()
    
    // 生成 1MB 测试数据
    testData := make([]byte, 1024*1024)
    for i := range testData {
        testData[i] = byte('A' + (i % 26))
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := converter.ConvertToUTF8(testData, EncodingUTF8)
        if err != nil {
            b.Fatal(err)
        }
    }
    
    // 报告处理速度
    b.SetBytes(int64(len(testData)))
}

func BenchmarkBatchProcessing(b *testing.B) {
    tempDir := b.TempDir()
    batchProcessor := NewDefaultBatch()
    
    // 创建测试文件
    files := make([]string, 100)
    for i := range files {
        file := filepath.Join(tempDir, fmt.Sprintf("file%d.txt", i))
        content := fmt.Sprintf("测试文件 %d 的内容", i)
        os.WriteFile(file, []byte(content), 0644)
        files[i] = file
    }
    
    options := &BatchOptions{
        ProcessOptions: ProcessOptions{
            TargetEncoding: EncodingUTF8,
            OutputDir:      filepath.Join(tempDir, "output"),
        },
        MaxConcurrency: runtime.NumCPU(),
    }
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        ctx := context.Background()
        results := batchProcessor.ProcessFiles(ctx, files, options)
        
        for range results {
            // 消费所有结果
        }
    }
}
```

### 性能基准指标

| 操作 | 目标性能 | 测量方法 |
|------|----------|----------|
| 编码检测 | < 1ms/KB | BenchmarkDetectEncoding |
| 编码转换 | > 10MB/s | BenchmarkConvert |
| 文件处理 | > 5MB/s | BenchmarkProcessFile |
| 批量处理 | > 100 files/s | BenchmarkBatchProcessing |

## 模糊测试策略

### 编码检测模糊测试

```go
// fuzz_test.go
func FuzzDetectEncoding(f *testing.F) {
    detector := NewDetector()
    
    // 添加种子语料
    f.Add([]byte("这是中文"))
    f.Add([]byte{0xd5, 0xe2, 0xca, 0xc7}) // GBK
    f.Add([]byte{0xff, 0xfe, 0x4f, 0x60}) // UTF-16LE BOM + 中
    
    f.Fuzz(func(t *testing.T, data []byte) {
        // 模糊测试不应该崩溃
        result, err := detector.DetectEncoding(data)
        
        if err != nil {
            // 错误是可接受的，但不应该panic
            return
        }
        
        // 如果成功检测，结果应该合理
        if result != nil {
            assert.NotEmpty(t, result.Encoding)
            assert.GreaterOrEqual(t, result.Confidence, 0.0)
            assert.LessOrEqual(t, result.Confidence, 1.0)
        }
    })
}

func FuzzConvert(f *testing.F) {
    converter := NewConverter()
    
    f.Add([]byte("测试"), "UTF-8", "GBK")
    f.Add([]byte{0xd5, 0xe2}, "GBK", "UTF-8")
    
    f.Fuzz(func(t *testing.T, data []byte, from, to string) {
        // 转换操作不应该panic
        defer func() {
            if r := recover(); r != nil {
                t.Errorf("Convert panicked: %v", r)
            }
        }()
        
        result, err := converter.Convert(data, from, to)
        
        // 如果转换成功，结果应该是有效的
        if err == nil && result != nil {
            assert.True(t, utf8.Valid(result))
        }
    })
}
```

## 兼容性测试策略

### Go 版本兼容性

```yaml
# .github/workflows/compatibility.yml
name: Compatibility Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        go-version: ['1.19', '1.20', '1.21', '1.22']
    
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: ${{ matrix.go-version }}
    
    - name: Run tests
      run: go test -v ./...
    
    - name: Run benchmarks
      run: go test -bench=. -benchmem ./...
```

### 操作系统兼容性

```yaml
# .github/workflows/cross-platform.yml
name: Cross Platform Tests

on: [push, pull_request]

jobs:
  test:
    strategy:
      matrix:
        os: [ubuntu-latest, windows-latest, macos-latest]
        go-version: ['1.21']
    
    runs-on: ${{ matrix.os }}
    
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: ${{ matrix.go-version }}
    
    - name: Run tests
      run: go test -v ./...
    
    - name: Test file operations
      run: go test -v ./... -tags=fileops
```

## 测试数据管理

### 测试文件组织

```
test/
├── testdata/
│   ├── encodings/
│   │   ├── utf8/
│   │   │   ├── chinese.txt
│   │   │   ├── japanese.txt
│   │   │   └── korean.txt
│   │   ├── gbk/
│   │   │   ├── simple.txt
│   │   │   └── complex.txt
│   │   ├── big5/
│   │   ├── shift_jis/
│   │   └── mixed/
│   ├── corrupted/
│   │   ├── incomplete.txt
│   │   ├── invalid_bytes.txt
│   │   └── mixed_encoding.txt
│   └── large/
│       ├── 1mb.txt
│       ├── 10mb.txt
│       └── 100mb.txt
├── fixtures/
│   ├── detector_test_cases.json
│   ├── converter_test_cases.json
│   └── integration_test_cases.json
└── utils/
    ├── test_helper.go
    └── data_generator.go
```

### 测试数据生成

```go
// test/utils/data_generator.go
func GenerateTestData(encoding string, size int) ([]byte, error) {
    switch encoding {
    case "UTF-8":
        return generateUTF8Data(size), nil
    case "GBK":
        return generateGBKData(size), nil
    case "BIG5":
        return generateBIG5Data(size), nil
    default:
        return nil, fmt.Errorf("unsupported encoding: %s", encoding)
    }
}

func generateUTF8Data(size int) []byte {
    chars := []rune("这是一段中文测试数据用于编码检测和转换测试")
    result := make([]byte, 0, size)
    
    for len(result) < size {
        for _, char := range chars {
            if len(result) >= size {
                break
            }
            result = append(result, []byte(string(char))...)
        }
    }
    
    return result[:size]
}
```

## 持续集成策略

### CI/CD 流水线

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.21
    
    - name: Cache Go modules
      uses: actions/cache@v3
      with:
        path: ~/go/pkg/mod
        key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
    
    - name: Download dependencies
      run: go mod download
    
    - name: Run tests with coverage
      run: go test -v -race -coverprofile=coverage.out ./...
    
    - name: Upload coverage to Codecov
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out
    
    - name: Run benchmarks
      run: go test -bench=. -benchmem ./... > benchmark.txt
    
    - name: Upload benchmark results
      uses: actions/upload-artifact@v3
      with:
        name: benchmark-results
        path: benchmark.txt
    
    - name: Run static analysis
      run: |
        go vet ./...
        go install honnef.co/go/tools/cmd/staticcheck@latest
        staticcheck ./...
    
    - name: Run security scan
      run: |
        go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
        gosec ./...
```

## 测试执行和报告

### 自动化测试脚本

```bash
#!/bin/bash
# scripts/run_tests.sh

set -e

echo "Running unit tests..."
go test -v -race -coverprofile=coverage.out ./...

echo "Running integration tests..."
go test -v -tags=integration ./test/integration/...

echo "Running benchmark tests..."
go test -bench=. -benchmem ./... | tee benchmark.txt

echo "Running fuzz tests..."
go test -fuzz=. -fuzztime=30s ./...

echo "Generating coverage report..."
go tool cover -html=coverage.out -o coverage.html

echo "Checking coverage threshold..."
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
if (( $(echo "$COVERAGE < 90" | bc -l) )); then
    echo "Coverage $COVERAGE% is below threshold 90%"
    exit 1
fi

echo "All tests passed! Coverage: $COVERAGE%"
```

### 测试报告生成

```go
// scripts/generate_report.go
package main

import (
    "encoding/json"
    "fmt"
    "os"
    "time"
)

type TestReport struct {
    Timestamp    time.Time     `json:"timestamp"`
    Coverage     float64       `json:"coverage"`
    TestResults  []TestResult  `json:"test_results"`
    Benchmarks   []Benchmark   `json:"benchmarks"`
    Issues       []Issue       `json:"issues"`
}

type TestResult struct {
    Package string  `json:"package"`
    Passed  int     `json:"passed"`
    Failed  int     `json:"failed"`
    Skipped int     `json:"skipped"`
    Time    float64 `json:"time"`
}

type Benchmark struct {
    Name         string  `json:"name"`
    Iterations   int     `json:"iterations"`
    NsPerOp      int64   `json:"ns_per_op"`
    BytesPerOp   int64   `json:"bytes_per_op"`
    AllocsPerOp  int64   `json:"allocs_per_op"`
}

func main() {
    report := generateReport()
    
    data, err := json.MarshalIndent(report, "", "  ")
    if err != nil {
        fmt.Printf("Error generating report: %v\n", err)
        os.Exit(1)
    }
    
    err = os.WriteFile("test_report.json", data, 0644)
    if err != nil {
        fmt.Printf("Error writing report: %v\n", err)
        os.Exit(1)
    }
    
    fmt.Println("Test report generated: test_report.json")
}
```

---

*本测试策略文档确保 EncodingProcessor 库的质量和可靠性，通过全面的测试覆盖保证产品质量*