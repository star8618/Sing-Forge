//go:build linux

package disk

import (
	"fmt"
)

// getPlatformDisks 获取平台磁盘信息
func getPlatformDisks() ([]DiskInfo, error) {
	return nil, fmt.Errorf("Linux disk info not implemented yet")
}

// getPlatformDiskIOStats 获取平台磁盘I/O统计
func getPlatformDiskIOStats() ([]DiskIOStats, error) {
	return nil, fmt.Errorf("Linux disk IO stats not implemented yet")
}

// getPlatformDiskHealth 获取平台磁盘健康信息
func getPlatformDiskHealth() ([]DiskHealth, error) {
	return nil, fmt.Errorf("Linux disk health not implemented yet")
}

// getPlatformPartitions 获取平台分区信息
func getPlatformPartitions() ([]PartitionInfo, error) {
	return nil, fmt.Errorf("Linux partitions not implemented yet")
}

// getLinuxDisks 获取Linux磁盘信息 (占位符实现)
func getLinuxDisks() ([]DiskInfo, error) {
	return nil, fmt.Errorf("Linux disk info not implemented yet")
}

// getLinuxDiskIOStats 获取Linux磁盘I/O统计 (占位符实现)
func getLinuxDiskIOStats() ([]DiskIOStats, error) {
	return nil, fmt.Errorf("Linux disk IO stats not implemented yet")
}

// getLinuxDiskHealth 获取Linux磁盘健康信息 (占位符实现)
func getLinuxDiskHealth() ([]DiskHealth, error) {
	return nil, fmt.Errorf("Linux disk health not implemented yet")
}

// getLinuxPartitions 获取Linux分区信息 (占位符实现)
func getLinuxPartitions() ([]PartitionInfo, error) {
	return nil, fmt.Errorf("Linux partition info not implemented yet")
}
