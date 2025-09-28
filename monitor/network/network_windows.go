//go:build windows

package network

import (
	"fmt"
)

// getPlatformInterfaces 获取平台网络接口
func getPlatformInterfaces() ([]NetworkInterface, error) {
	return nil, fmt.Errorf("Windows network interfaces not implemented yet")
}

// getPlatformInterfaceStats 获取平台接口统计
func getPlatformInterfaceStats() ([]NetworkStats, error) {
	return nil, fmt.Errorf("Windows interface stats not implemented yet")
}

// getPlatformConnections 获取平台连接信息
func getPlatformConnections() ([]ConnectionInfo, error) {
	return nil, fmt.Errorf("Windows connections not implemented yet")
}

// getWindowsInterfaces 获取Windows网络接口信息 (占位符实现)
func getWindowsInterfaces() ([]NetworkInterface, error) {
	return nil, fmt.Errorf("Windows network interfaces not implemented yet")
}

// getWindowsInterfaceStats 获取Windows网络接口统计 (占位符实现)
func getWindowsInterfaceStats() ([]NetworkStats, error) {
	return nil, fmt.Errorf("Windows network stats not implemented yet")
}

// getWindowsConnections 获取Windows网络连接信息 (占位符实现)
func getWindowsConnections() ([]ConnectionInfo, error) {
	return nil, fmt.Errorf("Windows network connections not implemented yet")
}
