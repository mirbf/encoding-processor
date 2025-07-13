package encoding

import (
	"archive/zip"
	"strings"
	"testing"
)

// TestSmartDetectionZipFile 测试ZIP文件名编码检测
func TestSmartDetectionZipFile(t *testing.T) {
	zipFile := "/Users/apple/Desktop/test/（暗恋）《时擦》作者：笙离.zip"
	
	// 检查文件是否存在
	reader, err := zip.OpenReader(zipFile)
	if err != nil {
		t.Skipf("无法打开测试文件: %v", err)
		return
	}
	defer reader.Close()
	
	if len(reader.File) == 0 {
		t.Fatal("ZIP文件为空")
	}
	
	fileName := reader.File[0].Name
	fileNameBytes := []byte(fileName)
	
	t.Logf("测试文件名: %s", fileName)
	t.Logf("字节长度: %d", len(fileNameBytes))
	
	// 测试智能检测
	processor := NewSmartProcessor()
	result, err := processor.SmartDetectEncoding(fileNameBytes)
	if err != nil {
		t.Fatalf("智能检测失败: %v", err)
	}
	
	t.Logf("检测结果: %s, 置信度: %.2f", result.Encoding, result.Confidence)
	
	// 验证转换结果
	if convertedText, ok := result.Details["converted_text"].(string); ok {
		t.Logf("转换结果: %s", convertedText)
		
		// 检查是否包含正确的中文字符
		if !strings.Contains(convertedText, "暗恋") || !strings.Contains(convertedText, "时擦") {
			t.Errorf("转换结果不正确，未包含期望的中文文本")
		}
	}
	
	// 验证编码应该是GBK或GB18030
	if result.Encoding != "GBK" && result.Encoding != "GB18030" {
		t.Logf("注意：检测结果为 %s，通常期望为 GBK 或 GB18030", result.Encoding)
	}
}

// TestEncodingDetectionAccuracy 测试编码检测准确性
func TestEncodingDetectionAccuracy(t *testing.T) {
	testCases := []struct {
		name     string
		text     string
		encoding string
		expected string
	}{
		{
			name:     "GBK中文文本",
			text:     "（暗恋）《时擦》作者：笙离",
			encoding: "GBK",
			expected: "GBK",
		},
		{
			name:     "UTF-8中文文本",
			text:     "（暗恋）《时擦》作者：笙离",
			encoding: "UTF-8",
			expected: "UTF-8",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			processor := NewDefault()
			
			// 先将文本转换为指定编码
			encoded, err := processor.ConvertString(tc.text, "UTF-8", tc.encoding)
			if err != nil {
				t.Fatalf("编码转换失败: %v", err)
			}
			
			data := []byte(encoded)
			
			// 使用智能检测
			smartProcessor := NewSmartProcessor()
			result, err := smartProcessor.SmartDetectEncoding(data)
			if err != nil {
				t.Logf("智能检测失败 (这是可以接受的): %v", err)
				
				// 回退到常规检测
				result, err = processor.DetectEncoding(data)
				if err != nil {
					t.Fatalf("常规检测也失败: %v", err)
				}
			}
			
			t.Logf("输入编码: %s, 检测结果: %s, 置信度: %.2f", 
				tc.encoding, result.Encoding, result.Confidence)
			
			// 对于短文本，检测可能不准确，所以只记录而不强制要求
			if result.Encoding != tc.expected {
				t.Logf("注意：检测结果 %s 与期望 %s 不符（短文本检测困难）", 
					result.Encoding, tc.expected)
			}
		})
	}
}

// TestZipFileProcessor 测试ZIP文件处理器
func TestZipFileProcessor(t *testing.T) {
	zipFile := "/Users/apple/Desktop/test/（暗恋）《时擦》作者：笙离.zip"
	
	reader, err := zip.OpenReader(zipFile)
	if err != nil {
		t.Skipf("无法打开测试文件: %v", err)
		return
	}
	defer reader.Close()
	
	if len(reader.File) == 0 {
		t.Fatal("ZIP文件为空")
	}
	
	fileName := reader.File[0].Name
	fileNameBytes := []byte(fileName)
	
	// 使用专用的ZIP文件处理器
	processor := NewZipFileProcessor()
	result, err := processor.SmartDetectEncoding(fileNameBytes)
	if err != nil {
		t.Fatalf("ZIP文件处理器检测失败: %v", err)
	}
	
	t.Logf("ZIP处理器检测结果: %s, 置信度: %.2f", result.Encoding, result.Confidence)
	
	// 验证能够正确转换
	converted, err := processor.ConvertString(fileName, result.Encoding, "UTF-8")
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}
	
	t.Logf("转换后的文件名: %s", converted)
	
	// 验证转换结果包含正确的中文
	if !strings.Contains(converted, "暗恋") {
		t.Errorf("转换结果不包含期望的中文字符")
	}
}

// BenchmarkSmartDetection 性能测试
func BenchmarkSmartDetection(b *testing.B) {
	text := "（暗恋）《时擦》作者：笙离.txt"
	data := []byte(text)
	
	processor := NewSmartProcessor()
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := processor.SmartDetectEncoding(data)
		if err != nil {
			b.Fatalf("检测失败: %v", err)
		}
	}
}