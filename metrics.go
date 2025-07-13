package encoding

import (
	"sync"
	"sync/atomic"
	"time"
)

// defaultMetricsCollector 实现 MetricsCollector 接口
type defaultMetricsCollector struct {
	stats *ProcessingStats
	mutex sync.RWMutex
}

// NewMetricsCollector 创建新的性能监控器
func NewMetricsCollector() MetricsCollector {
	return &defaultMetricsCollector{
		stats: &ProcessingStats{
			EncodingDistribution: make(map[string]int64),
			StartTime:            time.Now(),
			LastUpdateTime:       time.Now(),
		},
	}
}

// GetStats 获取处理统计信息
func (mc *defaultMetricsCollector) GetStats() *ProcessingStats {
	mc.mutex.RLock()
	defer mc.mutex.RUnlock()

	// 创建副本以避免并发修改
	statsCopy := &ProcessingStats{
		TotalOperations:      atomic.LoadInt64(&mc.stats.TotalOperations),
		SuccessOperations:    atomic.LoadInt64(&mc.stats.SuccessOperations),
		FailedOperations:     atomic.LoadInt64(&mc.stats.FailedOperations),
		TotalBytes:           atomic.LoadInt64(&mc.stats.TotalBytes),
		TotalProcessingTime:  mc.stats.TotalProcessingTime,
		StartTime:            mc.stats.StartTime,
		LastUpdateTime:       mc.stats.LastUpdateTime,
		EncodingDistribution: make(map[string]int64),
	}

	// 复制编码分布
	for encoding, count := range mc.stats.EncodingDistribution {
		statsCopy.EncodingDistribution[encoding] = count
	}

	// 计算平均处理速度
	if statsCopy.TotalProcessingTime > 0 {
		statsCopy.AverageProcessingSpeed = float64(statsCopy.TotalBytes) / statsCopy.TotalProcessingTime.Seconds()
	}

	return statsCopy
}

// ResetStats 重置统计信息
func (mc *defaultMetricsCollector) ResetStats() {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()

	atomic.StoreInt64(&mc.stats.TotalOperations, 0)
	atomic.StoreInt64(&mc.stats.SuccessOperations, 0)
	atomic.StoreInt64(&mc.stats.FailedOperations, 0)
	atomic.StoreInt64(&mc.stats.TotalBytes, 0)
	mc.stats.TotalProcessingTime = 0
	mc.stats.StartTime = time.Now()
	mc.stats.LastUpdateTime = time.Now()
	mc.stats.EncodingDistribution = make(map[string]int64)
}

// RecordOperation 记录操作
func (mc *defaultMetricsCollector) RecordOperation(operation string, duration time.Duration) {
	atomic.AddInt64(&mc.stats.TotalOperations, 1)
	atomic.AddInt64(&mc.stats.SuccessOperations, 1)

	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	mc.stats.TotalProcessingTime += duration
	mc.stats.LastUpdateTime = time.Now()
}

// RecordError 记录错误
func (mc *defaultMetricsCollector) RecordError(operation string, err error) {
	atomic.AddInt64(&mc.stats.TotalOperations, 1)
	atomic.AddInt64(&mc.stats.FailedOperations, 1)

	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	mc.stats.LastUpdateTime = time.Now()
}

// RecordBytes 记录处理的字节数
func (mc *defaultMetricsCollector) RecordBytes(bytes int64) {
	atomic.AddInt64(&mc.stats.TotalBytes, bytes)
}

// RecordEncoding 记录编码类型
func (mc *defaultMetricsCollector) RecordEncoding(encoding string) {
	mc.mutex.Lock()
	defer mc.mutex.Unlock()
	
	mc.stats.EncodingDistribution[encoding]++
	mc.stats.LastUpdateTime = time.Now()
}