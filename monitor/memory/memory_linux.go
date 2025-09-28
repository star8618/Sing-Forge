//go:build linux

package memory

import (
	"fmt"
)

// getPlatformMemoryInfo 获取平台内存信息
func getPlatformMemoryInfo(info *MemoryInfo) error {
	return fmt.Errorf("Linux memory info not implemented yet")
}

// getPlatformSwapInfo 获取平台交换分区信息
func getPlatformSwapInfo(info *SwapInfo) error {
	return fmt.Errorf("Linux swap info not implemented yet")
}

// getPlatformMemoryStats 获取平台内存统计
func getPlatformMemoryStats(stats *MemoryStats) error {
	return fmt.Errorf("Linux memory stats not implemented yet")
}

// getPlatformVirtualMemoryInfo 获取平台虚拟内存信息
func getPlatformVirtualMemoryInfo(info *VirtualMemoryInfo) error {
	return fmt.Errorf("Linux virtual memory info not implemented yet")
}

// getLinuxMemoryInfo 获取Linux内存信息 (占位符实现)
func getLinuxMemoryInfo(info *MemoryInfo) error {
	return fmt.Errorf("Linux memory info not implemented yet")
}

// getLinuxSwapInfo 获取Linux交换空间信息 (占位符实现)
func getLinuxSwapInfo(info *SwapInfo) error {
	return fmt.Errorf("Linux swap info not implemented yet")
}

// getLinuxMemoryStats 获取Linux内存详细统计 (占位符实现)
func getLinuxMemoryStats(stats *MemoryStats) error {
	return fmt.Errorf("Linux memory stats not implemented yet")
}

// getLinuxVirtualMemoryInfo 获取Linux虚拟内存信息 (占位符实现)
func getLinuxVirtualMemoryInfo(info *VirtualMemoryInfo) error {
	return fmt.Errorf("Linux virtual memory info not implemented yet")
}

// getDarwinMemoryPressure Linux平台不支持Darwin内存压力
func getDarwinMemoryPressure(pressure *MemoryPressure) error {
	return fmt.Errorf("Darwin memory pressure not available on Linux")
}
