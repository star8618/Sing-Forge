//go:build darwin

package memory

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// getPlatformMemoryInfo 获取平台内存信息
func getPlatformMemoryInfo(info *MemoryInfo) error {
	return getDarwinMemoryInfo(info)
}

// getPlatformSwapInfo 获取平台交换分区信息
func getPlatformSwapInfo(info *SwapInfo) error {
	return getDarwinSwapInfo(info)
}

// getPlatformMemoryStats 获取平台内存统计
func getPlatformMemoryStats(stats *MemoryStats) error {
	return getDarwinMemoryStats(stats)
}

// getPlatformVirtualMemoryInfo 获取平台虚拟内存信息
func getPlatformVirtualMemoryInfo(info *VirtualMemoryInfo) error {
	return getDarwinVirtualMemoryInfo(info)
}

// macOS系统调用常量
const (
	CTL_VM       = 2
	CTL_HW       = 6
	VM_LOADAVG   = 2
	VM_SWAPUSAGE = 5
	HW_MEMSIZE   = 24
	HW_PAGESIZE  = 7

	// vm_stat相关常量
	HOST_VM_INFO = 2
)

// vm_statistics64_data_t 结构体 (对应 mach/vm_statistics.h)
type vmStatistics64 struct {
	FreeCount                          uint32
	ActiveCount                        uint32
	InactiveCount                      uint32
	WireCount                          uint32
	ZeroFillCount                      uint64
	Reactivations                      uint64
	Pageins                            uint64
	Pageouts                           uint64
	Faults                             uint64
	CowFaults                          uint64
	Lookups                            uint64
	Hits                               uint64
	Purges                             uint64
	PurgeableCount                     uint32
	SpeculativeCount                   uint32
	Decompressions                     uint64
	Compressions                       uint64
	Swapins                            uint64
	Swapouts                           uint64
	CompressorPageCount                uint32
	ThrottledCount                     uint32
	ExternalPageCount                  uint32
	InternalPageCount                  uint32
	TotalUncompressedPagesInCompressor uint64
}

// xsw_usage 结构体 (交换空间使用情况)
type xswUsage struct {
	Total uint64
	Used  uint64
	Avail uint64
}

// getDarwinMemoryInfo 获取macOS内存信息
func getDarwinMemoryInfo(info *MemoryInfo) error {
	// 1. 获取总内存大小
	totalMem, err := sysctlUint64("hw.memsize")
	if err != nil {
		return fmt.Errorf("failed to get total memory: %v", err)
	}
	info.Total = totalMem

	// 2. 获取页面大小
	pageSize, err := sysctlUint64("hw.pagesize")
	if err != nil {
		return fmt.Errorf("failed to get page size: %v", err)
	}

	// 3. 获取vm_stat信息
	vmStats, err := getVMStatistics()
	if err != nil {
		return fmt.Errorf("failed to get vm statistics: %v", err)
	}

	// 4. 计算各种内存使用量
	info.Free = uint64(vmStats.FreeCount) * pageSize
	info.Active = uint64(vmStats.ActiveCount) * pageSize
	info.Inactive = uint64(vmStats.InactiveCount) * pageSize
	info.Wired = uint64(vmStats.WireCount) * pageSize
	info.Compressed = uint64(vmStats.CompressorPageCount) * pageSize

	// 5. 计算可用内存 (Free + Inactive + Cached + Speculative)
	speculative := uint64(vmStats.SpeculativeCount) * pageSize
	purgeable := uint64(vmStats.PurgeableCount) * pageSize
	info.Available = info.Free + info.Inactive + purgeable + speculative

	// 6. 计算已用内存
	info.Used = info.Total - info.Available

	// 7. 设置缓存和缓冲区 (macOS没有明确区分，使用purgeable作为cached)
	info.Cached = purgeable
	info.Buffers = 0 // macOS没有buffers概念

	return nil
}

// getDarwinSwapInfo 获取macOS交换空间信息
func getDarwinSwapInfo(info *SwapInfo) error {
	// 使用sysctl获取交换空间使用情况
	usage, err := getSwapUsage()
	if err != nil {
		return err
	}

	info.Total = usage.Total
	info.Used = usage.Used
	info.Free = usage.Avail

	// 获取换入换出次数
	vmStats, err := getVMStatistics()
	if err == nil {
		info.SwapIn = vmStats.Swapins
		info.SwapOut = vmStats.Swapouts
	}

	return nil
}

// getDarwinMemoryStats 获取macOS内存详细统计
func getDarwinMemoryStats(stats *MemoryStats) error {
	// 获取页面大小
	pageSize, err := sysctlUint64("hw.pagesize")
	if err != nil {
		return err
	}
	stats.PageSize = pageSize

	// 获取vm统计
	vmStats, err := getVMStatistics()
	if err != nil {
		return err
	}

	// 设置页面统计
	stats.FreePages = uint64(vmStats.FreeCount)
	stats.ActivePages = uint64(vmStats.ActiveCount)
	stats.InactivePages = uint64(vmStats.InactiveCount)
	stats.WiredPages = uint64(vmStats.WireCount)
	stats.TotalPages = (stats.FreePages + stats.ActivePages +
		stats.InactivePages + stats.WiredPages)

	// 设置操作统计
	stats.Faults = vmStats.Faults
	stats.Lookups = vmStats.Lookups
	stats.Hits = vmStats.Hits
	stats.Purges = vmStats.Purges

	return nil
}

// getDarwinVirtualMemoryInfo 获取macOS虚拟内存信息
func getDarwinVirtualMemoryInfo(info *VirtualMemoryInfo) error {
	// macOS的虚拟内存信息较难直接获取
	// 这里使用vm_stat命令的输出作为参考
	cmd := exec.Command("vm_stat")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	// 解析vm_stat输出
	return parseVMStatOutput(string(output), info)
}

// getDarwinMemoryPressure 获取macOS内存压力信息
func getDarwinMemoryPressure(pressure *MemoryPressure) error {
	// 使用memory_pressure命令
	cmd := exec.Command("memory_pressure")
	output, err := cmd.Output()
	if err != nil {
		// 如果memory_pressure命令不可用，尝试通过vm_stat计算
		return calculateMemoryPressureFromVMStat(pressure)
	}

	// 解析memory_pressure输出
	return parseMemoryPressureOutput(string(output), pressure)
}

// getVMStatistics 获取VM统计信息
func getVMStatistics() (*vmStatistics64, error) {
	// 这里需要调用host_statistics64系统调用
	// 由于Go语言限制，我们使用vm_stat命令作为替代
	return getVMStatisticsFromCommand()
}

// getVMStatisticsFromCommand 通过vm_stat命令获取统计信息
func getVMStatisticsFromCommand() (*vmStatistics64, error) {
	cmd := exec.Command("vm_stat")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	stats := &vmStatistics64{}
	scanner := bufio.NewScanner(bytes.NewReader(output))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "Pages free:") {
			if count, err := extractNumber(line); err == nil {
				stats.FreeCount = uint32(count)
			}
		} else if strings.HasPrefix(line, "Pages active:") {
			if count, err := extractNumber(line); err == nil {
				stats.ActiveCount = uint32(count)
			}
		} else if strings.HasPrefix(line, "Pages inactive:") {
			if count, err := extractNumber(line); err == nil {
				stats.InactiveCount = uint32(count)
			}
		} else if strings.HasPrefix(line, "Pages wired down:") {
			if count, err := extractNumber(line); err == nil {
				stats.WireCount = uint32(count)
			}
		} else if strings.HasPrefix(line, "Pages stored in compressor:") {
			if count, err := extractNumber(line); err == nil {
				stats.CompressorPageCount = uint32(count)
			}
		} else if strings.HasPrefix(line, "Pages occupied by compressor:") {
			// 这个值在某些版本的macOS中可用
		} else if strings.HasPrefix(line, "\"Swapins\":") {
			if count, err := extractNumber(line); err == nil {
				stats.Swapins = count
			}
		} else if strings.HasPrefix(line, "\"Swapouts\":") {
			if count, err := extractNumber(line); err == nil {
				stats.Swapouts = count
			}
		}
	}

	return stats, nil
}

// getSwapUsage 获取交换空间使用情况
func getSwapUsage() (*xswUsage, error) {
	// 使用sysctl VM_SWAPUSAGE
	usage := &xswUsage{}

	// 由于直接调用sysctl比较复杂，我们使用sysctl命令
	cmd := exec.Command("sysctl", "vm.swapusage")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// 解析输出: vm.swapusage: total = 1024.00M  used = 512.00M  free = 512.00M  (encrypted)
	line := string(output)

	// 提取total
	if match := regexp.MustCompile(`total = ([\d.]+)([KMGT]?)`).FindStringSubmatch(line); len(match) >= 3 {
		if size, err := parseSize(match[1], match[2]); err == nil {
			usage.Total = size
		}
	}

	// 提取used
	if match := regexp.MustCompile(`used = ([\d.]+)([KMGT]?)`).FindStringSubmatch(line); len(match) >= 3 {
		if size, err := parseSize(match[1], match[2]); err == nil {
			usage.Used = size
		}
	}

	// 提取free
	if match := regexp.MustCompile(`free = ([\d.]+)([KMGT]?)`).FindStringSubmatch(line); len(match) >= 3 {
		if size, err := parseSize(match[1], match[2]); err == nil {
			usage.Avail = size
		}
	}

	return usage, nil
}

// parseVMStatOutput 解析vm_stat输出获取虚拟内存信息
func parseVMStatOutput(output string, info *VirtualMemoryInfo) error {
	// vm_stat输出主要是物理内存信息，虚拟内存需要其他方式获取
	// 这里提供一个基本实现
	scanner := bufio.NewScanner(strings.NewReader(output))

	var totalPages, usedPages uint64
	pageSize := uint64(4096) // 默认页面大小

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.Contains(line, "page size of") {
			if size, err := extractNumber(line); err == nil {
				pageSize = size
			}
		} else if strings.HasPrefix(line, "Pages") {
			if count, err := extractNumber(line); err == nil {
				totalPages += count
				if !strings.Contains(line, "free") {
					usedPages += count
				}
			}
		}
	}

	info.Total = totalPages * pageSize
	info.Used = usedPages * pageSize
	info.Free = info.Total - info.Used

	return nil
}

// parseMemoryPressureOutput 解析memory_pressure输出
func parseMemoryPressureOutput(output string, pressure *MemoryPressure) error {
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.Contains(line, "System-wide memory free percentage:") {
			if percent, err := extractPercentage(line); err == nil {
				pressure.Percentage = 100 - percent // 转换为压力百分比
			}
		} else if strings.Contains(line, "normal") {
			pressure.Level = "normal"
		} else if strings.Contains(line, "warn") {
			pressure.Level = "warn"
		} else if strings.Contains(line, "urgent") {
			pressure.Level = "urgent"
		} else if strings.Contains(line, "critical") {
			pressure.Level = "critical"
		}
	}

	if pressure.Level == "" {
		pressure.Level = "normal"
	}

	return nil
}

// calculateMemoryPressureFromVMStat 从vm_stat计算内存压力
func calculateMemoryPressureFromVMStat(pressure *MemoryPressure) error {
	vmStats, err := getVMStatistics()
	if err != nil {
		return err
	}

	// 获取页面大小
	pageSize, err := sysctlUint64("hw.pagesize")
	if err != nil {
		pageSize = 4096
	}
	_ = pageSize // 避免未使用变量警告

	// 计算内存压力
	totalPages := vmStats.FreeCount + vmStats.ActiveCount + vmStats.InactiveCount + vmStats.WireCount
	freePages := vmStats.FreeCount + vmStats.InactiveCount + vmStats.PurgeableCount

	if totalPages > 0 {
		freePercentage := float64(freePages) / float64(totalPages) * 100
		pressure.Percentage = 100 - freePercentage

		// 根据可用内存百分比设置压力级别
		if freePercentage > 80 {
			pressure.Level = "normal"
		} else if freePercentage > 60 {
			pressure.Level = "warn"
		} else if freePercentage > 40 {
			pressure.Level = "urgent"
		} else {
			pressure.Level = "critical"
		}
	}

	pressure.PagesFreed = vmStats.Purges
	pressure.PagesPurged = vmStats.Purges
	pressure.PagesSpeculative = uint64(vmStats.SpeculativeCount)

	return nil
}

// 辅助函数

// sysctlUint64 获取sysctl的uint64值
func sysctlUint64(name string) (uint64, error) {
	// 使用sysctl命令作为简化实现
	cmd := exec.Command("sysctl", "-n", name)
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	valueStr := strings.TrimSpace(string(output))
	return strconv.ParseUint(valueStr, 10, 64)
}

// extractNumber 从字符串中提取数字
func extractNumber(line string) (uint64, error) {
	re := regexp.MustCompile(`(\d+)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) >= 2 {
		return strconv.ParseUint(matches[1], 10, 64)
	}
	return 0, fmt.Errorf("no number found in line: %s", line)
}

// extractPercentage 从字符串中提取百分比
func extractPercentage(line string) (float64, error) {
	re := regexp.MustCompile(`([\d.]+)%`)
	matches := re.FindStringSubmatch(line)
	if len(matches) >= 2 {
		return strconv.ParseFloat(matches[1], 64)
	}
	return 0, fmt.Errorf("no percentage found in line: %s", line)
}

// parseSize 解析大小字符串 (如 "512.00M")
func parseSize(sizeStr, unit string) (uint64, error) {
	size, err := strconv.ParseFloat(sizeStr, 64)
	if err != nil {
		return 0, err
	}

	multiplier := uint64(1)
	switch strings.ToUpper(unit) {
	case "K":
		multiplier = 1024
	case "M":
		multiplier = 1024 * 1024
	case "G":
		multiplier = 1024 * 1024 * 1024
	case "T":
		multiplier = 1024 * 1024 * 1024 * 1024
	}

	return uint64(size * float64(multiplier)), nil
}
