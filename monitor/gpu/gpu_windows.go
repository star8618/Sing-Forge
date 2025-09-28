//go:build windows

package gpu

import (
	"fmt"
)

// getPlatformGPUs 获取平台GPU信息
func getPlatformGPUs() ([]*GPUInfo, error) {
	return nil, fmt.Errorf("Windows GPU info not implemented yet")
}

// getPlatformGPUUsage 获取平台GPU使用率
func getPlatformGPUUsage() ([]*GPUUsage, error) {
	return nil, fmt.Errorf("Windows GPU usage not implemented yet")
}

// getPlatformGPUProcesses 获取平台GPU进程
func getPlatformGPUProcesses() ([]*GPUProcess, error) {
	return nil, fmt.Errorf("Windows GPU processes not implemented yet")
}

// getWindowsGPUs 获取Windows GPU信息 (占位符实现)
func getWindowsGPUs() ([]*GPUInfo, error) {
	return nil, fmt.Errorf("Windows GPU info not implemented yet")
}

// getWindowsGPUUsage 获取Windows GPU使用率 (占位符实现)
func getWindowsGPUUsage() ([]*GPUUsage, error) {
	return nil, fmt.Errorf("Windows GPU usage not implemented yet")
}

// getWindowsGPUProcesses 获取Windows GPU进程 (占位符实现)
func getWindowsGPUProcesses() ([]*GPUProcess, error) {
	return nil, fmt.Errorf("Windows GPU processes not implemented yet")
}

// getDarwinAppleGPUInfo Windows平台不支持Apple GPU
func getDarwinAppleGPUInfo() (*AppleGPUInfo, error) {
	return nil, fmt.Errorf("Apple GPU info not available on Windows")
}
