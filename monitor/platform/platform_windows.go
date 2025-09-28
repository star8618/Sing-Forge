//go:build windows

package platform

import (
	"fmt"
)

// getPlatformInfo 获取平台信息
func getPlatformInfo(info *PlatformInfo) error {
	return fmt.Errorf("Windows platform info not implemented yet")
}

// getPlatformHardwarePlatform 获取硬件平台
func getPlatformHardwarePlatform(hardware *HardwarePlatform) error {
	return fmt.Errorf("Windows hardware platform not implemented yet")
}

// setPlatformCapabilities 设置平台能力
func setPlatformCapabilities(caps *Capabilities) {
	// Windows占位符实现
}

// isPlatformVirtualMachine 检查是否虚拟机
func isPlatformVirtualMachine() (bool, error) {
	return false, fmt.Errorf("Windows VM detection not implemented yet")
}

// isPlatformContainer 检查是否容器
func isPlatformContainer() (bool, error) {
	return false, fmt.Errorf("Windows container detection not implemented yet")
}

// getWindowsPlatformInfo 获取Windows平台信息 (占位符实现)
func getWindowsPlatformInfo(info *PlatformInfo) error {
	return fmt.Errorf("Windows platform info not implemented yet")
}

// getWindowsHardwarePlatform 获取Windows硬件平台信息 (占位符实现)
func getWindowsHardwarePlatform(hardware *HardwarePlatform) error {
	return fmt.Errorf("Windows hardware platform not implemented yet")
}

// setWindowsCapabilities 设置Windows平台能力 (占位符实现)
func setWindowsCapabilities(caps *Capabilities) {
	// 基本设置
	caps.CPUTemperature = false
	caps.CPUFrequency = true
	caps.MemoryPressure = false
	caps.DiskHealth = true
	caps.NetworkDetails = true
	caps.ProcessDetails = true
	caps.GPUInfo = true
	caps.BatteryInfo = true
	caps.SensorInfo = false
	caps.ContainerSupport = true
	caps.VirtualizationSupport = true
}

// isWindowsVirtualMachine 检查是否为虚拟机 (占位符实现)
func isWindowsVirtualMachine() (bool, error) {
	return false, fmt.Errorf("Windows VM detection not implemented yet")
}

// isWindowsContainer 检查是否运行在容器中 (占位符实现)
func isWindowsContainer() (bool, error) {
	return false, fmt.Errorf("Windows container detection not implemented yet")
}
