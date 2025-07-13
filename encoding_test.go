package encoding

import (
	"strings"
	"testing"
)

func TestBasicDetection(t *testing.T) {
	processor := NewDefault()

	tests := []struct {
		name     string
		data     []byte
		expected string
	}{
		{
			name:     "UTF-8 English text",
			data:     []byte("Hello, World!"),
			expected: EncodingUTF8,
		},
		{
			name:     "UTF-8 Chinese text",
			data:     []byte("你好，世界！"),
			expected: EncodingUTF8,
		},
		{
			name:     "Empty data",
			data:     []byte{},
			expected: "", // Should return error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processor.DetectEncoding(tt.data)
			
			if tt.expected == "" {
				// Expecting an error
				if err == nil {
					t.Errorf("Expected error for %s, but got none", tt.name)
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error for %s: %v", tt.name, err)
				return
			}
			
			if result.Encoding != tt.expected {
				t.Errorf("Expected encoding %s for %s, got %s", tt.expected, tt.name, result.Encoding)
			}
			
			if result.Confidence <= 0 {
				t.Errorf("Expected positive confidence for %s, got %f", tt.name, result.Confidence)
			}
		})
	}
}

func TestBasicConversion(t *testing.T) {
	processor := NewDefault()

	tests := []struct {
		name     string
		input    string
		from     string
		to       string
		expected string
	}{
		{
			name:     "UTF-8 to UTF-8 (no change)",
			input:    "Hello, World!",
			from:     EncodingUTF8,
			to:       EncodingUTF8,
			expected: "Hello, World!",
		},
		{
			name:     "UTF-8 Chinese",
			input:    "你好",
			from:     EncodingUTF8,
			to:       EncodingUTF8,
			expected: "你好",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processor.ConvertString(tt.input, tt.from, tt.to)
			if err != nil {
				t.Errorf("Unexpected error for %s: %v", tt.name, err)
				return
			}
			
			if result != tt.expected {
				t.Errorf("Expected %s for %s, got %s", tt.expected, tt.name, result)
			}
		})
	}
}

func TestSmartConversion(t *testing.T) {
	processor := NewDefault()

	tests := []struct {
		name     string
		input    string
		target   string
		expected string
	}{
		{
			name:     "Smart convert English",
			input:    "Hello, World!",
			target:   EncodingUTF8,
			expected: "Hello, World!",
		},
		{
			name:     "Smart convert Chinese",
			input:    "你好，世界！",
			target:   EncodingUTF8,
			expected: "你好，世界！",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processor.SmartConvertString(tt.input, tt.target)
			if err != nil {
				t.Errorf("Unexpected error for %s: %v", tt.name, err)
				return
			}
			
			if result.Text != tt.expected {
				t.Errorf("Expected %s for %s, got %s", tt.expected, tt.name, result.Text)
			}
			
			if result.TargetEncoding != tt.target {
				t.Errorf("Expected target encoding %s for %s, got %s", tt.target, tt.name, result.TargetEncoding)
			}
		})
	}
}

func TestFactoryFunctions(t *testing.T) {
	tests := []struct {
		name    string
		factory func() Processor
	}{
		{"NewDefault", NewDefault},
		{"NewQuick", NewQuick},
		{"NewForCLI", NewForCLI},
		{"NewForWebService", NewForWebService},
		{"NewForBatchProcessing", NewForBatchProcessing},
		{"NewHighPerformance", NewHighPerformance},
		{"NewMemoryEfficient", NewMemoryEfficient},
		{"NewStrictMode", NewStrictMode},
		{"NewTolerantMode", NewTolerantMode},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			processor := tt.factory()
			if processor == nil {
				t.Errorf("Factory function %s returned nil", tt.name)
				return
			}

			// Test basic functionality
			result, err := processor.DetectBestEncoding([]byte("Hello"))
			if err != nil {
				t.Errorf("Basic detection failed for %s: %v", tt.name, err)
				return
			}

			if result == "" {
				t.Errorf("Expected non-empty encoding for %s", tt.name)
			}
		})
	}
}

func TestDefaultWithMetrics(t *testing.T) {
	processor, metrics := NewDefaultWithMetrics()
	
	if processor == nil {
		t.Error("NewDefaultWithMetrics returned nil processor")
		return
	}
	
	if metrics == nil {
		t.Error("NewDefaultWithMetrics returned nil metrics")
		return
	}

	// Test that metrics collector works
	stats := metrics.GetStats()
	if stats == nil {
		t.Error("GetStats returned nil")
		return
	}

	// Initial stats should be zero
	if stats.TotalOperations != 0 {
		t.Errorf("Expected 0 total operations, got %d", stats.TotalOperations)
	}
}

func TestErrorHandling(t *testing.T) {
	processor := NewDefault()

	// Test with invalid encoding
	_, err := processor.Convert([]byte("test"), "INVALID_ENCODING", EncodingUTF8)
	if err == nil {
		t.Error("Expected error for invalid encoding, got none")
	}

	// Check if it's the correct error type
	if !strings.Contains(err.Error(), "unsupported encoding") {
		t.Errorf("Expected unsupported encoding error, got: %v", err)
	}
}

func TestStreamProcessor(t *testing.T) {
	streamProcessor := NewDefaultStream()
	if streamProcessor == nil {
		t.Error("NewDefaultStream returned nil")
	}

	// Test that the interface is implemented correctly
	var _ StreamProcessor = streamProcessor
}

func TestFileProcessor(t *testing.T) {
	fileProcessor := NewDefaultFile()
	if fileProcessor == nil {
		t.Error("NewDefaultFile returned nil")
	}

	// Test that the interface is implemented correctly
	var _ FileProcessor = fileProcessor
}

func TestMetricsCollector(t *testing.T) {
	metrics := NewMetricsCollector()
	if metrics == nil {
		t.Error("NewMetricsCollector returned nil")
		return
	}

	// Test recording operations
	metrics.RecordOperation("test", 100)
	
	stats := metrics.GetStats()
	if stats.TotalOperations != 1 {
		t.Errorf("Expected 1 total operation, got %d", stats.TotalOperations)
	}
	
	if stats.SuccessOperations != 1 {
		t.Errorf("Expected 1 success operation, got %d", stats.SuccessOperations)
	}

	// Test reset
	metrics.ResetStats()
	stats = metrics.GetStats()
	if stats.TotalOperations != 0 {
		t.Errorf("Expected 0 total operations after reset, got %d", stats.TotalOperations)
	}
}