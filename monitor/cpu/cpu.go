// Package cpu 提供跨平台CPU监控功能，专门优化Apple Silicon
package cpu

import (
	"fmt"
	"runtime"
	"time"
)

// CPUInfo CPU基本信息
type CPUInfo struct {
	Model            string    `json:"model"`             // CPU型号
	Cores            int       `json:"cores"`             // 总核心数
	PerformanceCores int       `json:"performance_cores"` // 性能核心数（Apple Silicon）
	EfficiencyCores  int       `json:"efficiency_cores"`  // 效率核心数（Apple Silicon）
	Threads          int       `json:"threads"`           // 线程数
	Frequency        float64   `json:"frequency"`         // 基础频率 (GHz)
	MaxFrequency     float64   `json:"max_frequency"`     // 最大频率 (GHz)
	Architecture     string    `json:"architecture"`      // 架构 (arm64, x86_64)
	Vendor           string    `json:"vendor"`            // 厂商
	Family           string    `json:"family"`            // CPU系列
	CacheL1          int       `json:"cache_l1"`          // L1缓存大小 (KB)
	CacheL2          int       `json:"cache_l2"`          // L2缓存大小 (KB)
	CacheL3          int       `json:"cache_l3"`          // L3缓存大小 (KB)
	Temperature      float64   `json:"temperature"`       // 温度 (℃)
	LastUpdated      time.Time `json:"last_updated"`      // 最后更新时间
}

// CPUUsage CPU使用率信息
type CPUUsage struct {
	Overall          float64   `json:"overall"`           // 总体使用率
	PerformanceCores float64   `json:"performance_cores"` // 性能核心使用率
	EfficiencyCores  float64   `json:"efficiency_cores"`  // 效率核心使用率
	PerCoreUsage     []float64 `json:"per_core_usage"`    // 每个核心使用率
	User             float64   `json:"user"`              // 用户态使用率
	System           float64   `json:"system"`            // 系统态使用率
	Idle             float64   `json:"idle"`              // 空闲率
	Nice             float64   `json:"nice"`              // Nice进程使用率
	IOWait           float64   `json:"iowait"`            // IO等待时间
	IRQ              float64   `json:"irq"`               // 硬中断时间
	SoftIRQ          float64   `json:"soft_irq"`          // 软中断时间
	LoadAvg1         float64   `json:"load_avg_1"`        // 1分钟负载平均
	LoadAvg5         float64   `json:"load_avg_5"`        // 5分钟负载平均
	LoadAvg15        float64   `json:"load_avg_15"`       // 15分钟负载平均
	LastUpdated      time.Time `json:"last_updated"`      // 最后更新时间
}

// CPUStats CPU统计信息（用于计算使用率）
type CPUStats struct {
	User    uint64
	Nice    uint64
	System  uint64
	Idle    uint64
	IOWait  uint64
	IRQ     uint64
	SoftIRQ uint64
	Steal   uint64
	Guest   uint64
	Total   uint64
}

var (
	lastCPUStats    *CPUStats
	lastUpdateTime  time.Time
	cachedCPUInfo   *CPUInfo
	cacheExpireTime time.Time
)

// GetInfo 获取CPU基本信息（带缓存）
func GetInfo() (*CPUInfo, error) {
	// 检查缓存是否有效（10分钟过期）
	if cachedCPUInfo != nil && time.Now().Before(cacheExpireTime) {
		return cachedCPUInfo, nil
	}

	info := &CPUInfo{
		Architecture: runtime.GOARCH,
		Cores:        runtime.NumCPU(),
		LastUpdated:  time.Now(),
	}

	// 根据平台获取详细信息
	var err error
	err = getPlatformCPUInfo(info)

	if err != nil {
		return nil, err
	}

	// 缓存结果
	cachedCPUInfo = info
	cacheExpireTime = time.Now().Add(10 * time.Minute)

	return info, nil
}

// GetUsage 获取CPU实时使用率
func GetUsage() (*CPUUsage, error) {
	return getPlatformCPUUsage()
}

// GetUsageWithDuration 获取指定采样时间的CPU使用率
func GetUsageWithDuration(duration time.Duration) (*CPUUsage, error) {
	// 获取当前CPU统计
	currentStats, err := getCPUStats()
	if err != nil {
		return nil, err
	}

	// 如果是第一次调用，等待一个采样周期
	if lastCPUStats == nil {
		lastCPUStats = currentStats
		lastUpdateTime = time.Now()
		time.Sleep(duration)

		currentStats, err = getCPUStats()
		if err != nil {
			return nil, err
		}
	}

	// 计算使用率
	usage := calculateCPUUsage(lastCPUStats, currentStats)
	usage.LastUpdated = time.Now()

	// 获取每个核心的使用率（如果支持）
	if perCoreUsage, err := getPerCoreCPUUsage(duration); err == nil {
		usage.PerCoreUsage = perCoreUsage
	}

	// 更新缓存
	lastCPUStats = currentStats
	lastUpdateTime = time.Now()

	return usage, nil
}

// calculateCPUUsage 计算CPU使用率
func calculateCPUUsage(last, current *CPUStats) *CPUUsage {
	// 计算时间差
	totalDiff := current.Total - last.Total
	if totalDiff == 0 {
		return &CPUUsage{}
	}

	usage := &CPUUsage{
		User:    float64(current.User-last.User) / float64(totalDiff) * 100,
		Nice:    float64(current.Nice-last.Nice) / float64(totalDiff) * 100,
		System:  float64(current.System-last.System) / float64(totalDiff) * 100,
		Idle:    float64(current.Idle-last.Idle) / float64(totalDiff) * 100,
		IOWait:  float64(current.IOWait-last.IOWait) / float64(totalDiff) * 100,
		IRQ:     float64(current.IRQ-last.IRQ) / float64(totalDiff) * 100,
		SoftIRQ: float64(current.SoftIRQ-last.SoftIRQ) / float64(totalDiff) * 100,
	}

	// 计算总体使用率
	usage.Overall = 100 - usage.Idle

	return usage
}

// GetTemperature 获取CPU温度（如果支持）
func GetTemperature() (float64, error) {
	return getPlatformCPUTemperature()
}

// GetFrequency 获取CPU当前频率
func GetFrequency() (float64, error) {
	return getPlatformCPUFrequency()
}

// IsAppleSilicon 检查是否为Apple Silicon处理器
func IsAppleSilicon() bool {
	return runtime.GOOS == "darwin" && runtime.GOARCH == "arm64"
}

// GetAppleSiliconDetails 获取Apple Silicon详细信息
func GetAppleSiliconDetails() (*AppleSiliconInfo, error) {
	if !IsAppleSilicon() {
		return nil, fmt.Errorf("not an Apple Silicon system")
	}

	return getAppleSiliconInfo()
}

// AppleSiliconInfo Apple Silicon特有信息
type AppleSiliconInfo struct {
	ChipName         string  `json:"chip_name"`         // 芯片名称 (M1, M2, M3)
	PerformanceCores int     `json:"performance_cores"` // 性能核心数
	EfficiencyCores  int     `json:"efficiency_cores"`  // 效率核心数
	GPUCores         int     `json:"gpu_cores"`         // GPU核心数
	NeuralCores      int     `json:"neural_cores"`      // Neural Engine核心数
	MemoryBandwidth  float64 `json:"memory_bandwidth"`  // 内存带宽 (GB/s)
	ProcessNode      string  `json:"process_node"`      // 制程工艺
}
