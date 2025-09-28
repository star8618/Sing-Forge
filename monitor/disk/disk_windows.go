//go:build windows

package disk

import (
	"fmt"
)

// getPlatformDisks 获取平台磁盘信息
func getPlatformDisks() ([]DiskInfo, error) {
	return nil, fmt.Errorf("Windows disk info not implemented yet")
}

// getPlatformDiskIOStats 获取平台磁盘I/O统计
func getPlatformDiskIOStats() ([]DiskIOStats, error) {
	return nil, fmt.Errorf("Windows disk IO stats not implemented yet")
}

// getPlatformDiskHealth 获取平台磁盘健康信息
func getPlatformDiskHealth() ([]DiskHealth, error) {
	return nil, fmt.Errorf("Windows disk health not implemented yet")
}

// getPlatformPartitions 获取平台分区信息
func getPlatformPartitions() ([]PartitionInfo, error) {
	return nil, fmt.Errorf("Windows partitions not implemented yet")
}

// getWindowsDisks 获取Windows磁盘信息 (占位符实现)
func getWindowsDisks() ([]DiskInfo, error) {
	return nil, fmt.Errorf("Windows disk info not implemented yet")
}

// getWindowsDiskIOStats 获取Windows磁盘I/O统计 (占位符实现)
func getWindowsDiskIOStats() ([]DiskIOStats, error) {
	return nil, fmt.Errorf("Windows disk IO stats not implemented yet")
}

// getWindowsDiskHealth 获取Windows磁盘健康信息 (占位符实现)
func getWindowsDiskHealth() ([]DiskHealth, error) {
	return nil, fmt.Errorf("Windows disk health not implemented yet")
}

// getWindowsPartitions 获取Windows分区信息 (占位符实现)
func getWindowsPartitions() ([]PartitionInfo, error) {
	return nil, fmt.Errorf("Windows partition info not implemented yet")
}
