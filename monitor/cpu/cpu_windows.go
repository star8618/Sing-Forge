//go:build windows

package cpu

import (
	"fmt"
	"time"
)

// getPlatformCPUInfo 获取平台CPU信息
func getPlatformCPUInfo(info *CPUInfo) error {
	return fmt.Errorf("Windows CPU info not implemented yet")
}

// getPlatformCPUTemperature 获取平台CPU温度
func getPlatformCPUTemperature() (float64, error) {
	return 0, fmt.Errorf("Windows CPU temperature not implemented yet")
}

// getPlatformCPUFrequency 获取平台CPU频率
func getPlatformCPUFrequency() (float64, error) {
	return 0, fmt.Errorf("Windows CPU frequency not implemented yet")
}

// getPlatformCPUUsage 获取平台CPU使用率
func getPlatformCPUUsage() (*CPUUsage, error) {
	return nil, fmt.Errorf("Windows CPU usage not implemented yet")
}

// getWindowsCPUInfo 获取Windows CPU信息 (占位符实现)
func getWindowsCPUInfo(info *CPUInfo) error {
	return fmt.Errorf("Windows CPU info not implemented yet")
}

// getWindowsCPUFrequency 获取Windows CPU频率 (占位符实现)
func getWindowsCPUFrequency() (float64, error) {
	return 0, fmt.Errorf("Windows CPU frequency not implemented yet")
}

// getCPUStats 获取Windows CPU统计信息 (占位符实现)
func getCPUStats() (*CPUStats, error) {
	return &CPUStats{}, fmt.Errorf("Windows CPU stats not implemented yet")
}

// getPerCoreCPUUsage 获取Windows每个核心的CPU使用率 (占位符实现)
func getPerCoreCPUUsage(duration time.Duration) ([]float64, error) {
	return nil, fmt.Errorf("Windows per-core CPU usage not implemented yet")
}

// getAppleSiliconInfo Windows平台不支持Apple Silicon
func getAppleSiliconInfo() (*AppleSiliconInfo, error) {
	return nil, fmt.Errorf("Apple Silicon info not available on Windows")
}
