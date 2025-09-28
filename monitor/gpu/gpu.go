// Package gpu 提供跨平台GPU监控功能
package gpu

import (
	"fmt"
	"runtime"
	"time"
)

// GPUInfo GPU基本信息
type GPUInfo struct {
	Name            string    `json:"name"`             // GPU名称
	Vendor          string    `json:"vendor"`           // 厂商 (Apple, NVIDIA, AMD, Intel)
	Model           string    `json:"model"`            // 型号
	Architecture    string    `json:"architecture"`     // 架构
	Cores           int       `json:"cores"`            // 核心数
	ComputeUnits    int       `json:"compute_units"`    // 计算单元数
	Memory          uint64    `json:"memory"`           // 显存大小 (bytes)
	MemoryType      string    `json:"memory_type"`      // 显存类型
	MemoryBandwidth float64   `json:"memory_bandwidth"` // 内存带宽 (GB/s)
	ClockSpeed      float64   `json:"clock_speed"`      // 基础时钟频率 (MHz)
	BoostClock      float64   `json:"boost_clock"`      // 加速时钟频率 (MHz)
	PowerDraw       float64   `json:"power_draw"`       // 功耗 (W)
	Temperature     float64   `json:"temperature"`      // 温度 (°C)
	DriverVersion   string    `json:"driver_version"`   // 驱动版本
	IsIntegrated    bool      `json:"is_integrated"`    // 是否集成显卡
	IsDiscrete      bool      `json:"is_discrete"`      // 是否独立显卡
	LastUpdated     time.Time `json:"last_updated"`     // 最后更新时间
}

// GPUUsage GPU使用率信息
type GPUUsage struct {
	GPUPercent    float64   `json:"gpu_percent"`    // GPU使用率
	MemoryPercent float64   `json:"memory_percent"` // 显存使用率
	MemoryUsed    uint64    `json:"memory_used"`    // 已用显存 (bytes)
	MemoryFree    uint64    `json:"memory_free"`    // 空闲显存 (bytes)
	PowerUsage    float64   `json:"power_usage"`    // 当前功耗 (W)
	Temperature   float64   `json:"temperature"`    // 当前温度 (°C)
	FanSpeed      float64   `json:"fan_speed"`      // 风扇转速 (%)
	ClockSpeed    float64   `json:"clock_speed"`    // 当前时钟频率 (MHz)
	MemoryClock   float64   `json:"memory_clock"`   // 显存时钟频率 (MHz)
	LastUpdated   time.Time `json:"last_updated"`   // 最后更新时间
}

// GPUProcess GPU进程信息
type GPUProcess struct {
	PID         uint32  `json:"pid"`          // 进程ID
	ProcessName string  `json:"process_name"` // 进程名称
	MemoryUsed  uint64  `json:"memory_used"`  // 使用的显存 (bytes)
	GPUPercent  float64 `json:"gpu_percent"`  // GPU使用率
}

// AppleGPUInfo Apple GPU特有信息
type AppleGPUInfo struct {
	ChipName        string  `json:"chip_name"`        // 芯片名称 (M1, M2, M3)
	GPUCores        int     `json:"gpu_cores"`        // GPU核心数
	TileMemory      uint64  `json:"tile_memory"`      // Tile内存
	UnifiedMemory   uint64  `json:"unified_memory"`   // 统一内存
	MemoryBandwidth float64 `json:"memory_bandwidth"` // 内存带宽
	MetalVersion    string  `json:"metal_version"`    // Metal版本
	TBDRCapable     bool    `json:"tbdr_capable"`     // 是否支持TBDR
}

var (
	cachedGPUInfo   []*GPUInfo
	cacheExpireTime time.Time
)

// GetGPUs 获取所有GPU信息
func GetGPUs() ([]*GPUInfo, error) {
	// 检查缓存
	if cachedGPUInfo != nil && time.Now().Before(cacheExpireTime) {
		return cachedGPUInfo, nil
	}

	var gpus []*GPUInfo
	var err error

	// 根据平台获取GPU信息
	gpus, err = getPlatformGPUs()

	if err != nil {
		return nil, err
	}

	// 更新时间戳
	now := time.Now()
	for _, gpu := range gpus {
		gpu.LastUpdated = now
	}

	// 缓存结果（10分钟）
	cachedGPUInfo = gpus
	cacheExpireTime = time.Now().Add(10 * time.Minute)

	return gpus, nil
}

// GetPrimaryGPU 获取主GPU信息
func GetPrimaryGPU() (*GPUInfo, error) {
	gpus, err := GetGPUs()
	if err != nil {
		return nil, err
	}

	if len(gpus) == 0 {
		return nil, fmt.Errorf("no GPU found")
	}

	// 返回第一个GPU（通常是主GPU）
	return gpus[0], nil
}

// GetGPUUsage 获取GPU使用率信息
func GetGPUUsage() ([]*GPUUsage, error) {
	var usage []*GPUUsage
	var err error

	// 根据平台获取GPU使用率
	usage, err = getPlatformGPUUsage()

	if err != nil {
		return nil, err
	}

	// 更新时间戳
	now := time.Now()
	for _, u := range usage {
		u.LastUpdated = now
	}

	return usage, nil
}

// GetGPUProcesses 获取使用GPU的进程列表
func GetGPUProcesses() ([]*GPUProcess, error) {
	var processes []*GPUProcess
	var err error

	// 根据平台获取GPU进程信息
	processes, err = getPlatformGPUProcesses()

	return processes, err
}

// GetAppleGPUInfo 获取Apple GPU特有信息
func GetAppleGPUInfo() (*AppleGPUInfo, error) {
	if runtime.GOOS != "darwin" || runtime.GOARCH != "arm64" {
		return nil, fmt.Errorf("Apple GPU info only available on Apple Silicon")
	}

	return getDarwinAppleGPUInfo()
}

// IsAppleGPU 检查是否为Apple GPU
func IsAppleGPU() bool {
	return runtime.GOOS == "darwin" && runtime.GOARCH == "arm64"
}

// GetGPUSummary 获取GPU概览信息
func GetGPUSummary() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 获取GPU基本信息
	if gpus, err := GetGPUs(); err == nil {
		result["gpus"] = gpus
		result["gpu_count"] = len(gpus)

		if len(gpus) > 0 {
			result["primary_gpu"] = gpus[0]
		}
	}

	// 获取GPU使用率
	if usage, err := GetGPUUsage(); err == nil {
		result["usage"] = usage
	}

	// 获取Apple GPU特有信息
	if IsAppleGPU() {
		if appleInfo, err := GetAppleGPUInfo(); err == nil {
			result["apple_gpu"] = appleInfo
		}
	}

	// 获取GPU进程
	if processes, err := GetGPUProcesses(); err == nil {
		result["processes"] = processes
		result["process_count"] = len(processes)
	}

	return result, nil
}

// RefreshCache 刷新GPU信息缓存
func RefreshCache() {
	cachedGPUInfo = nil
	cacheExpireTime = time.Time{}
}

// FormatMemory 格式化显存大小
func FormatMemory(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// GetGPUCapabilities 获取GPU能力
func GetGPUCapabilities() map[string]bool {
	caps := make(map[string]bool)

	switch runtime.GOOS {
	case "darwin":
		caps["temperature_monitoring"] = IsAppleGPU()
		caps["usage_monitoring"] = true
		caps["memory_monitoring"] = true
		caps["process_monitoring"] = true
		caps["power_monitoring"] = IsAppleGPU()
		caps["metal_support"] = true
		caps["opencl_support"] = true
		caps["compute_support"] = true
	case "linux":
		caps["temperature_monitoring"] = true
		caps["usage_monitoring"] = true
		caps["memory_monitoring"] = true
		caps["process_monitoring"] = true
		caps["power_monitoring"] = true
		caps["vulkan_support"] = true
		caps["opencl_support"] = true
		caps["cuda_support"] = true // 对NVIDIA GPU
	case "windows":
		caps["temperature_monitoring"] = true
		caps["usage_monitoring"] = true
		caps["memory_monitoring"] = true
		caps["process_monitoring"] = true
		caps["power_monitoring"] = true
		caps["directx_support"] = true
		caps["vulkan_support"] = true
		caps["opencl_support"] = true
		caps["cuda_support"] = true // 对NVIDIA GPU
	}

	return caps
}
