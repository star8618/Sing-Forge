//go:build linux

package platform

import (
	"fmt"
)

// getPlatformInfo 获取平台信息
func getPlatformInfo(info *PlatformInfo) error {
	return fmt.Errorf("Linux platform info not implemented yet")
}

// getPlatformHardwarePlatform 获取硬件平台
func getPlatformHardwarePlatform(hardware *HardwarePlatform) error {
	return fmt.Errorf("Linux hardware platform not implemented yet")
}

// setPlatformCapabilities 设置平台能力
func setPlatformCapabilities(caps *Capabilities) {
	// Linux占位符实现
}

// isPlatformVirtualMachine 检查是否虚拟机
func isPlatformVirtualMachine() (bool, error) {
	return false, fmt.Errorf("Linux VM detection not implemented yet")
}

// isPlatformContainer 检查是否容器
func isPlatformContainer() (bool, error) {
	return false, fmt.Errorf("Linux container detection not implemented yet")
}

// getLinuxPlatformInfo 获取Linux平台信息 (占位符实现)
func getLinuxPlatformInfo(info *PlatformInfo) error {
	return fmt.Errorf("Linux platform info not implemented yet")
}

// getLinuxHardwarePlatform 获取Linux硬件平台信息 (占位符实现)
func getLinuxHardwarePlatform(hardware *HardwarePlatform) error {
	return fmt.Errorf("Linux hardware platform not implemented yet")
}

// setLinuxCapabilities 设置Linux平台能力 (占位符实现)
func setLinuxCapabilities(caps *Capabilities) {
	// 基本设置
	caps.CPUTemperature = true
	caps.CPUFrequency = true
	caps.MemoryPressure = false
	caps.DiskHealth = true
	caps.NetworkDetails = true
	caps.ProcessDetails = true
	caps.GPUInfo = true
	caps.BatteryInfo = false
	caps.SensorInfo = true
	caps.ContainerSupport = true
	caps.VirtualizationSupport = true
}

// isLinuxVirtualMachine 检查是否为虚拟机 (占位符实现)
func isLinuxVirtualMachine() (bool, error) {
	return false, fmt.Errorf("Linux VM detection not implemented yet")
}

// isLinuxContainer 检查是否运行在容器中 (占位符实现)
func isLinuxContainer() (bool, error) {
	return false, fmt.Errorf("Linux container detection not implemented yet")
}
