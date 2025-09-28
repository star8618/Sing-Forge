// Package memory 提供跨平台内存监控功能
package memory

import (
	"fmt"
	"runtime"
	"time"
)

// MemoryInfo 内存基本信息
type MemoryInfo struct {
	Total       uint64    `json:"total"`        // 总内存 (bytes)
	Available   uint64    `json:"available"`    // 可用内存 (bytes)
	Used        uint64    `json:"used"`         // 已用内存 (bytes)
	Free        uint64    `json:"free"`         // 空闲内存 (bytes)
	UsedPercent float64   `json:"used_percent"` // 使用率百分比
	Cached      uint64    `json:"cached"`       // 缓存内存 (bytes)
	Buffers     uint64    `json:"buffers"`      // 缓冲区内存 (bytes)
	Shared      uint64    `json:"shared"`       // 共享内存 (bytes)
	Active      uint64    `json:"active"`       // 活跃内存 (bytes)
	Inactive    uint64    `json:"inactive"`     // 非活跃内存 (bytes)
	Wired       uint64    `json:"wired"`        // 联动内存 (bytes) - macOS特有
	Compressed  uint64    `json:"compressed"`   // 压缩内存 (bytes) - macOS特有
	LastUpdated time.Time `json:"last_updated"` // 最后更新时间
}

// SwapInfo 交换空间信息
type SwapInfo struct {
	Total       uint64    `json:"total"`        // 总交换空间 (bytes)
	Used        uint64    `json:"used"`         // 已用交换空间 (bytes)
	Free        uint64    `json:"free"`         // 空闲交换空间 (bytes)
	UsedPercent float64   `json:"used_percent"` // 使用率百分比
	SwapIn      uint64    `json:"swap_in"`      // 换入次数
	SwapOut     uint64    `json:"swap_out"`     // 换出次数
	LastUpdated time.Time `json:"last_updated"` // 最后更新时间
}

// MemoryPressure 内存压力信息 (macOS特有)
type MemoryPressure struct {
	Level            string    `json:"level"`             // 压力级别: normal, warn, urgent, critical
	Percentage       float64   `json:"percentage"`        // 压力百分比
	PagesFreed       uint64    `json:"pages_freed"`       // 释放的页面数
	PagesPurged      uint64    `json:"pages_purged"`      // 清除的页面数
	PagesSpeculative uint64    `json:"pages_speculative"` // 推测页面数
	LastUpdated      time.Time `json:"last_updated"`      // 最后更新时间
}

// MemoryStats 内存详细统计
type MemoryStats struct {
	PageSize      uint64    `json:"page_size"`      // 页面大小 (bytes)
	TotalPages    uint64    `json:"total_pages"`    // 总页面数
	FreePages     uint64    `json:"free_pages"`     // 空闲页面数
	ActivePages   uint64    `json:"active_pages"`   // 活跃页面数
	InactivePages uint64    `json:"inactive_pages"` // 非活跃页面数
	WiredPages    uint64    `json:"wired_pages"`    // 联动页面数
	Faults        uint64    `json:"faults"`         // 页面错误次数
	Lookups       uint64    `json:"lookups"`        // 查找次数
	Hits          uint64    `json:"hits"`           // 命中次数
	Purges        uint64    `json:"purges"`         // 清除次数
	LastUpdated   time.Time `json:"last_updated"`   // 最后更新时间
}

// VirtualMemoryInfo 虚拟内存信息
type VirtualMemoryInfo struct {
	Total       uint64    `json:"total"`        // 总虚拟内存 (bytes)
	Used        uint64    `json:"used"`         // 已用虚拟内存 (bytes)
	Free        uint64    `json:"free"`         // 空闲虚拟内存 (bytes)
	UsedPercent float64   `json:"used_percent"` // 使用率百分比
	LastUpdated time.Time `json:"last_updated"` // 最后更新时间
}

var (
	lastMemoryStats *MemoryStats
	lastStatsTime   time.Time
)

// GetInfo 获取内存基本信息
func GetInfo() (*MemoryInfo, error) {
	info := &MemoryInfo{
		LastUpdated: time.Now(),
	}

	// 根据平台获取内存信息
	err := getPlatformMemoryInfo(info)

	if err != nil {
		return nil, err
	}

	// 计算使用率
	if info.Total > 0 {
		info.UsedPercent = float64(info.Used) / float64(info.Total) * 100
	}

	return info, nil
}

// GetSwapInfo 获取交换空间信息
func GetSwapInfo() (*SwapInfo, error) {
	info := &SwapInfo{
		LastUpdated: time.Now(),
	}

	err := getPlatformSwapInfo(info)

	if err != nil {
		return nil, err
	}

	// 计算使用率
	if info.Total > 0 {
		info.UsedPercent = float64(info.Used) / float64(info.Total) * 100
	}

	return info, nil
}

// GetStats 获取内存详细统计
func GetStats() (*MemoryStats, error) {
	stats := &MemoryStats{
		LastUpdated: time.Now(),
	}

	err := getPlatformMemoryStats(stats)

	return stats, err
}

// GetVirtualMemoryInfo 获取虚拟内存信息
func GetVirtualMemoryInfo() (*VirtualMemoryInfo, error) {
	info := &VirtualMemoryInfo{
		LastUpdated: time.Now(),
	}

	err := getPlatformVirtualMemoryInfo(info)

	if err != nil {
		return nil, err
	}

	// 计算使用率
	if info.Total > 0 {
		info.UsedPercent = float64(info.Used) / float64(info.Total) * 100
	}

	return info, nil
}

// GetMemoryPressure 获取内存压力信息 (仅macOS)
func GetMemoryPressure() (*MemoryPressure, error) {
	if runtime.GOOS != "darwin" {
		return nil, fmt.Errorf("memory pressure monitoring only supported on macOS")
	}

	pressure := &MemoryPressure{
		LastUpdated: time.Now(),
	}

	err := getDarwinMemoryPressure(pressure)
	return pressure, err
}

// GetDetailedInfo 获取完整的内存信息
func GetDetailedInfo() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 基本内存信息
	if info, err := GetInfo(); err == nil {
		result["memory"] = info
	}

	// 交换空间信息
	if swap, err := GetSwapInfo(); err == nil {
		result["swap"] = swap
	}

	// 虚拟内存信息
	if virtual, err := GetVirtualMemoryInfo(); err == nil {
		result["virtual"] = virtual
	}

	// 详细统计
	if stats, err := GetStats(); err == nil {
		result["stats"] = stats
	}

	// macOS特有的内存压力信息
	if runtime.GOOS == "darwin" {
		if pressure, err := GetMemoryPressure(); err == nil {
			result["pressure"] = pressure
		}
	}

	return result, nil
}

// FormatBytes 格式化字节数为可读格式
func FormatBytes(bytes uint64) string {
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

// GetUsageHistory 获取内存使用历史 (需要持续调用来建立历史数据)
func GetUsageHistory(duration time.Duration, interval time.Duration) ([]MemoryInfo, error) {
	var history []MemoryInfo

	start := time.Now()
	for time.Since(start) < duration {
		if info, err := GetInfo(); err == nil {
			history = append(history, *info)
		}
		time.Sleep(interval)
	}

	return history, nil
}
