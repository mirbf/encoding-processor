package main

import (
	"fmt"
	"log"

	encoding "github.com/mirbf/encoding-processor"
)

func main() {
	// 创建处理器
	processor := encoding.NewDefault()

	// 测试编码检测
	fmt.Println("=== 编码检测测试 ===")
	
	testData := []struct {
		name string
		data []byte
	}{
		{"英文文本", []byte("Hello, World!")},
		{"中文文本", []byte("你好，世界！")},
		{"混合文本", []byte("Hello 你好 World 世界!")},
	}

	for _, test := range testData {
		result, err := processor.DetectEncoding(test.data)
		if err != nil {
			log.Printf("%s 检测失败: %v", test.name, err)
			continue
		}
		
		fmt.Printf("%s: %s (置信度: %.2f)\n", 
			test.name, result.Encoding, result.Confidence)
	}

	// 测试智能转换
	fmt.Println("\n=== 智能转换测试 ===")
	
	text := "这是一段测试文本 - This is a test text"
	result, err := processor.SmartConvertString(text, encoding.EncodingUTF8)
	if err != nil {
		log.Fatalf("智能转换失败: %v", err)
	}
	
	fmt.Printf("源编码: %s\n", result.SourceEncoding)
	fmt.Printf("目标编码: %s\n", result.TargetEncoding)
	fmt.Printf("转换结果: %s\n", result.Text)
	fmt.Printf("处理字节数: %d\n", result.BytesProcessed)
	fmt.Printf("转换耗时: %v\n", result.ConversionTime)

	// 测试工厂函数
	fmt.Println("\n=== 工厂函数测试 ===")
	
	processors := map[string]encoding.Processor{
		"默认处理器":     encoding.NewDefault(),
		"CLI处理器":     encoding.NewForCLI(),
		"Web服务处理器":   encoding.NewForWebService(),
		"高性能处理器":    encoding.NewHighPerformance(),
		"内存高效处理器":   encoding.NewMemoryEfficient(),
		"严格模式处理器":   encoding.NewStrictMode(),
		"容错模式处理器":   encoding.NewTolerantMode(),
	}

	testText := "Test text 测试文本"
	for name, proc := range processors {
		encoding_name, err := proc.DetectBestEncoding([]byte(testText))
		if err != nil {
			fmt.Printf("%s: 检测失败 - %v\n", name, err)
		} else {
			fmt.Printf("%s: %s\n", name, encoding_name)
		}
	}

	// 测试性能监控
	fmt.Println("\n=== 性能监控测试 ===")
	
	processor_with_metrics, metrics := encoding.NewDefaultWithMetrics()
	
	// 执行一些操作
	for i := 0; i < 5; i++ {
		_, err := processor_with_metrics.DetectEncoding([]byte(fmt.Sprintf("Test %d", i)))
		if err != nil {
			metrics.RecordError("detect", err)
		}
	}
	
	stats := metrics.GetStats()
	fmt.Printf("总操作数: %d\n", stats.TotalOperations)
	fmt.Printf("成功操作数: %d\n", stats.SuccessOperations)
	fmt.Printf("失败操作数: %d\n", stats.FailedOperations)

	fmt.Println("\n✅ EncodingProcessor 库实现完成!")
}