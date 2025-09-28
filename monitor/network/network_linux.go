//go:build linux

package network

import (
	"fmt"
)

// getPlatformInterfaces 获取平台网络接口
func getPlatformInterfaces() ([]NetworkInterface, error) {
	return nil, fmt.Errorf("Linux network interfaces not implemented yet")
}

// getPlatformInterfaceStats 获取平台接口统计
func getPlatformInterfaceStats() ([]NetworkStats, error) {
	return nil, fmt.Errorf("Linux interface stats not implemented yet")
}

// getPlatformConnections 获取平台连接信息
func getPlatformConnections() ([]ConnectionInfo, error) {
	return nil, fmt.Errorf("Linux connections not implemented yet")
}

// getLinuxInterfaces 获取Linux网络接口信息 (占位符实现)
func getLinuxInterfaces() ([]NetworkInterface, error) {
	return nil, fmt.Errorf("Linux network interfaces not implemented yet")
}

// getLinuxInterfaceStats 获取Linux网络接口统计 (占位符实现)
func getLinuxInterfaceStats() ([]NetworkStats, error) {
	return nil, fmt.Errorf("Linux network stats not implemented yet")
}

// getLinuxConnections 获取Linux网络连接信息 (占位符实现)
func getLinuxConnections() ([]ConnectionInfo, error) {
	return nil, fmt.Errorf("Linux network connections not implemented yet")
}
