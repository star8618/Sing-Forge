//go:build windows

package memory

import (
	"fmt"
)

// getPlatformMemoryInfo 获取平台内存信息
func getPlatformMemoryInfo(info *MemoryInfo) error {
	return fmt.Errorf("Windows memory info not implemented yet")
}

// getPlatformSwapInfo 获取平台交换分区信息
func getPlatformSwapInfo(info *SwapInfo) error {
	return fmt.Errorf("Windows swap info not implemented yet")
}

// getPlatformMemoryStats 获取平台内存统计
func getPlatformMemoryStats(stats *MemoryStats) error {
	return fmt.Errorf("Windows memory stats not implemented yet")
}

// getPlatformVirtualMemoryInfo 获取平台虚拟内存信息
func getPlatformVirtualMemoryInfo(info *VirtualMemoryInfo) error {
	return fmt.Errorf("Windows virtual memory info not implemented yet")
}

// getWindowsMemoryInfo 获取Windows内存信息 (占位符实现)
func getWindowsMemoryInfo(info *MemoryInfo) error {
	return fmt.Errorf("Windows memory info not implemented yet")
}

// getWindowsSwapInfo 获取Windows交换空间信息 (占位符实现)
func getWindowsSwapInfo(info *SwapInfo) error {
	return fmt.Errorf("Windows swap info not implemented yet")
}

// getWindowsMemoryStats 获取Windows内存详细统计 (占位符实现)
func getWindowsMemoryStats(stats *MemoryStats) error {
	return fmt.Errorf("Windows memory stats not implemented yet")
}

// getWindowsVirtualMemoryInfo 获取Windows虚拟内存信息 (占位符实现)
func getWindowsVirtualMemoryInfo(info *VirtualMemoryInfo) error {
	return fmt.Errorf("Windows virtual memory info not implemented yet")
}

// getDarwinMemoryPressure Windows平台不支持Darwin内存压力
func getDarwinMemoryPressure(pressure *MemoryPressure) error {
	return fmt.Errorf("Darwin memory pressure not available on Windows")
}
