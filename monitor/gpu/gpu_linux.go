//go:build linux

package gpu

import (
	"fmt"
)

// getPlatformGPUs 获取平台GPU信息
func getPlatformGPUs() ([]*GPUInfo, error) {
	return nil, fmt.Errorf("Linux GPU info not implemented yet")
}

// getPlatformGPUUsage 获取平台GPU使用率
func getPlatformGPUUsage() ([]*GPUUsage, error) {
	return nil, fmt.Errorf("Linux GPU usage not implemented yet")
}

// getPlatformGPUProcesses 获取平台GPU进程
func getPlatformGPUProcesses() ([]*GPUProcess, error) {
	return nil, fmt.Errorf("Linux GPU processes not implemented yet")
}

// getLinuxGPUs 获取Linux GPU信息 (占位符实现)
func getLinuxGPUs() ([]*GPUInfo, error) {
	return nil, fmt.Errorf("Linux GPU info not implemented yet")
}

// getLinuxGPUUsage 获取Linux GPU使用率 (占位符实现)
func getLinuxGPUUsage() ([]*GPUUsage, error) {
	return nil, fmt.Errorf("Linux GPU usage not implemented yet")
}

// getLinuxGPUProcesses 获取Linux GPU进程 (占位符实现)
func getLinuxGPUProcesses() ([]*GPUProcess, error) {
	return nil, fmt.Errorf("Linux GPU processes not implemented yet")
}

// getDarwinAppleGPUInfo Linux平台不支持Apple GPU
func getDarwinAppleGPUInfo() (*AppleGPUInfo, error) {
	return nil, fmt.Errorf("Apple GPU info not available on Linux")
}
