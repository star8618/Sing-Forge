//go:build linux

package cpu

import (
	"fmt"
	"time"
)

// getPlatformCPUInfo 获取平台CPU信息
func getPlatformCPUInfo(info *CPUInfo) error {
	return fmt.Errorf("Linux CPU info not implemented yet")
}

// getPlatformCPUTemperature 获取平台CPU温度
func getPlatformCPUTemperature() (float64, error) {
	return 0, fmt.Errorf("Linux CPU temperature not implemented yet")
}

// getPlatformCPUFrequency 获取平台CPU频率
func getPlatformCPUFrequency() (float64, error) {
	return 0, fmt.Errorf("Linux CPU frequency not implemented yet")
}

// getPlatformCPUUsage 获取平台CPU使用率
func getPlatformCPUUsage() (*CPUUsage, error) {
	return nil, fmt.Errorf("Linux CPU usage not implemented yet")
}

// getLinuxCPUInfo 获取Linux CPU信息 (占位符实现)
func getLinuxCPUInfo(info *CPUInfo) error {
	return fmt.Errorf("Linux CPU info not implemented yet")
}

// getLinuxCPUTemperature 获取Linux CPU温度 (占位符实现)
func getLinuxCPUTemperature() (float64, error) {
	return 0, fmt.Errorf("Linux CPU temperature not implemented yet")
}

// getLinuxCPUFrequency 获取Linux CPU频率 (占位符实现)
func getLinuxCPUFrequency() (float64, error) {
	return 0, fmt.Errorf("Linux CPU frequency not implemented yet")
}

// getCPUStats 获取Linux CPU统计信息 (占位符实现)
func getCPUStats() (*CPUStats, error) {
	return &CPUStats{}, fmt.Errorf("Linux CPU stats not implemented yet")
}

// getPerCoreCPUUsage 获取Linux每个核心的CPU使用率 (占位符实现)
func getPerCoreCPUUsage(duration time.Duration) ([]float64, error) {
	return nil, fmt.Errorf("Linux per-core CPU usage not implemented yet")
}

// getAppleSiliconInfo Linux平台不支持Apple Silicon
func getAppleSiliconInfo() (*AppleSiliconInfo, error) {
	return nil, fmt.Errorf("Apple Silicon info not available on Linux")
}
