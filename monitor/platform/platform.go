// Package platform 提供跨平台系统信息检测和平台特定优化
package platform

import (
	"fmt"
	"runtime"
	"time"
)

// PlatformInfo 平台信息
type PlatformInfo struct {
	OS           string    `json:"os"`           // 操作系统
	Architecture string    `json:"architecture"` // 架构
	Kernel       string    `json:"kernel"`       // 内核版本
	Distribution string    `json:"distribution"` // 发行版 (仅Linux)
	Version      string    `json:"version"`      // 系统版本
	BuildNumber  string    `json:"build_number"` // 构建号
	Hostname     string    `json:"hostname"`     // 主机名
	Uptime       uint64    `json:"uptime"`       // 运行时间 (秒)
	BootTime     time.Time `json:"boot_time"`    // 启动时间
	LastUpdated  time.Time `json:"last_updated"` // 最后更新时间
}

// HardwarePlatform 硬件平台信息
type HardwarePlatform struct {
	Vendor         string `json:"vendor"`           // 厂商
	Model          string `json:"model"`            // 型号
	Serial         string `json:"serial"`           // 序列号
	UUID           string `json:"uuid"`             // UUID
	Chassis        string `json:"chassis"`          // 机箱类型
	IsVirtual      bool   `json:"is_virtual"`       // 是否虚拟机
	IsContainer    bool   `json:"is_container"`     // 是否容器
	IsAppleSilicon bool   `json:"is_apple_silicon"` // 是否Apple Silicon
}

// Capabilities 平台能力
type Capabilities struct {
	// 监控能力
	CPUTemperature bool `json:"cpu_temperature"` // CPU温度监控
	CPUFrequency   bool `json:"cpu_frequency"`   // CPU频率监控
	PerCoreUsage   bool `json:"per_core_usage"`  // 每核心使用率
	MemoryPressure bool `json:"memory_pressure"` // 内存压力监控
	DiskHealth     bool `json:"disk_health"`     // 磁盘健康监控
	NetworkDetails bool `json:"network_details"` // 详细网络信息
	ProcessDetails bool `json:"process_details"` // 详细进程信息

	// 硬件信息能力
	GPUInfo     bool `json:"gpu_info"`     // GPU信息
	BatteryInfo bool `json:"battery_info"` // 电池信息
	SensorInfo  bool `json:"sensor_info"`  // 传感器信息

	// 系统能力
	ContainerSupport      bool `json:"container_support"`      // 容器支持
	VirtualizationSupport bool `json:"virtualization_support"` // 虚拟化支持
}

// GetPlatformInfo 获取平台信息
func GetPlatformInfo() (*PlatformInfo, error) {
	info := &PlatformInfo{
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
		LastUpdated:  time.Now(),
	}

	// 获取平台特定信息
	err := getPlatformInfo(info)
	return info, err
}

// GetHardwarePlatform 获取硬件平台信息
func GetHardwarePlatform() (*HardwarePlatform, error) {
	hardware := &HardwarePlatform{
		IsAppleSilicon: IsAppleSilicon(),
	}

	// 根据平台获取硬件信息
	err := getPlatformHardwarePlatform(hardware)

	return hardware, err
}

// GetCapabilities 获取平台监控能力
func GetCapabilities() *Capabilities {
	caps := &Capabilities{}

	setPlatformCapabilities(caps)

	return caps
}

// IsAppleSilicon 检查是否为Apple Silicon
func IsAppleSilicon() bool {
	return runtime.GOOS == "darwin" && runtime.GOARCH == "arm64"
}

// IsVirtualMachine 检查是否运行在虚拟机中
func IsVirtualMachine() (bool, error) {
	return isPlatformVirtualMachine()
}

// IsContainer 检查是否运行在容器中
func IsContainer() (bool, error) {
	return isPlatformContainer()
}

// GetOptimalSampleInterval 获取平台优化的采样间隔
func GetOptimalSampleInterval() time.Duration {
	switch runtime.GOOS {
	case "darwin":
		if IsAppleSilicon() {
			return 100 * time.Millisecond // Apple Silicon优化
		}
		return 200 * time.Millisecond
	case "linux":
		return 250 * time.Millisecond
	case "windows":
		return 500 * time.Millisecond
	default:
		return 1 * time.Second
	}
}

// GetOptimalConcurrency 获取平台优化的并发数
func GetOptimalConcurrency() int {
	cores := runtime.NumCPU()

	switch runtime.GOOS {
	case "darwin":
		if IsAppleSilicon() {
			// Apple Silicon优化：使用更多并发
			return cores * 2
		}
		return cores
	case "linux":
		return cores
	case "windows":
		// Windows较保守
		return cores / 2
	default:
		return 1
	}
}

// SupportsFeature 检查平台是否支持特定功能
func SupportsFeature(feature string) bool {
	caps := GetCapabilities()

	switch feature {
	case "cpu_temperature":
		return caps.CPUTemperature
	case "cpu_frequency":
		return caps.CPUFrequency
	case "memory_pressure":
		return caps.MemoryPressure
	case "disk_health":
		return caps.DiskHealth
	case "gpu_info":
		return caps.GPUInfo
	case "battery_info":
		return caps.BatteryInfo
	default:
		return false
	}
}

// GetSystemCallInterface 获取系统调用接口优化配置
func GetSystemCallInterface() map[string]interface{} {
	config := make(map[string]interface{})

	switch runtime.GOOS {
	case "darwin":
		config["use_sysctl"] = true
		config["use_mach_calls"] = true
		config["use_iokit"] = IsAppleSilicon()
		config["cpu_sample_method"] = "host_processor_info"
		config["memory_sample_method"] = "vm_statistics64"
	case "linux":
		config["use_procfs"] = true
		config["use_sysfs"] = true
		config["cpu_sample_method"] = "/proc/stat"
		config["memory_sample_method"] = "/proc/meminfo"
	case "windows":
		config["use_wmi"] = true
		config["use_perfcounters"] = true
		config["cpu_sample_method"] = "GetSystemTimes"
		config["memory_sample_method"] = "GlobalMemoryStatusEx"
	}

	return config
}

// GetRecommendedBufferSizes 获取推荐的缓冲区大小
func GetRecommendedBufferSizes() map[string]int {
	sizes := make(map[string]int)

	switch runtime.GOOS {
	case "darwin":
		sizes["cpu_history"] = 60     // 1分钟历史
		sizes["memory_history"] = 60  // 1分钟历史
		sizes["network_history"] = 30 // 30秒历史
		sizes["disk_history"] = 30    // 30秒历史
	case "linux":
		sizes["cpu_history"] = 120    // 2分钟历史
		sizes["memory_history"] = 120 // 2分钟历史
		sizes["network_history"] = 60 // 1分钟历史
		sizes["disk_history"] = 60    // 1分钟历史
	case "windows":
		sizes["cpu_history"] = 180     // 3分钟历史
		sizes["memory_history"] = 180  // 3分钟历史
		sizes["network_history"] = 120 // 2分钟历史
		sizes["disk_history"] = 120    // 2分钟历史
	default:
		sizes["cpu_history"] = 60
		sizes["memory_history"] = 60
		sizes["network_history"] = 30
		sizes["disk_history"] = 30
	}

	return sizes
}

// GetPlatformSpecificConfig 获取平台特定配置
func GetPlatformSpecificConfig() map[string]interface{} {
	config := make(map[string]interface{})

	// 基础配置
	config["os"] = runtime.GOOS
	config["arch"] = runtime.GOARCH
	config["cpu_count"] = runtime.NumCPU()
	config["is_apple_silicon"] = IsAppleSilicon()

	// 平台特定配置
	switch runtime.GOOS {
	case "darwin":
		config["use_system_profiler"] = true
		config["use_powermetrics"] = IsAppleSilicon()
		config["temperature_sensors"] = []string{"TC0P", "TC0H", "TC0D"}
		config["network_interfaces_cmd"] = "networksetup"
		config["disk_utility_cmd"] = "diskutil"
	case "linux":
		config["use_systemd"] = true
		config["hwmon_path"] = "/sys/class/hwmon"
		config["thermal_path"] = "/sys/class/thermal"
		config["network_path"] = "/sys/class/net"
		config["block_path"] = "/sys/block"
	case "windows":
		config["use_wmic"] = true
		config["use_powershell"] = true
		config["perfmon_counters"] = true
		config["wmi_namespace"] = "root/cimv2"
	}

	// 监控优化配置
	config["sample_interval"] = GetOptimalSampleInterval()
	config["max_concurrency"] = GetOptimalConcurrency()
	config["buffer_sizes"] = GetRecommendedBufferSizes()
	config["capabilities"] = GetCapabilities()

	return config
}

// ValidatePlatformRequirements 验证平台要求
func ValidatePlatformRequirements() error {
	// 检查Go版本要求
	if !isGoVersionSupported() {
		return fmt.Errorf("unsupported Go version, requires Go 1.19+")
	}

	// 检查操作系统支持
	if !isPlatformSupported() {
		return fmt.Errorf("unsupported platform: %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	// 检查权限要求
	if err := checkPermissions(); err != nil {
		return fmt.Errorf("insufficient permissions: %v", err)
	}

	return nil
}

// 辅助函数

func isGoVersionSupported() bool {
	// 简化实现，假设支持
	return true
}

func isPlatformSupported() bool {
	supportedPlatforms := map[string][]string{
		"darwin":  {"amd64", "arm64"},
		"linux":   {"amd64", "arm64", "386", "arm"},
		"windows": {"amd64", "386"},
	}

	if archs, exists := supportedPlatforms[runtime.GOOS]; exists {
		for _, arch := range archs {
			if arch == runtime.GOARCH {
				return true
			}
		}
	}

	return false
}

func checkPermissions() error {
	// 检查基本的读取权限
	// 这里可以添加具体的权限检查逻辑
	return nil
}
