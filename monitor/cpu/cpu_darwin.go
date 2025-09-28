//go:build darwin

package cpu

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

// macOS系统调用相关常量
const (
	CTL_HW          = 6
	CTL_KERN        = 1
	HW_CPU_FREQ     = 15
	HW_TB_FREQ      = 16
	KERN_BOOTTIME   = 21
	HW_MEMSIZE      = 24
	HW_MACHINE      = 1
	HW_MODEL        = 2
	HW_NCPU         = 3
	HW_BYTEORDER    = 4
	HW_PHYSMEM      = 5
	HW_USERMEM      = 6
	HW_PAGESIZE     = 7
	HW_DISKNAMES    = 8
	HW_DISKSTATS    = 9
	HW_FLOATINGPT   = 10
	HW_MACHINE_ARCH = 11
	HW_VECTORUNIT   = 12
	HW_BUS_FREQ     = 13
	HW_CPU_FREQ_MAX = 14
)

// SystemProfilerHardware system_profiler硬件信息结构
type SystemProfilerHardware struct {
	SPHardwareDataType []struct {
		ChipType         string `json:"chip_type"`
		MachineModel     string `json:"machine_model"`
		MachineName      string `json:"machine_name"`
		NumberProcessors string `json:"number_processors"`
		ProcessorName    string `json:"processor_name"`
		ProcessorSpeed   string `json:"processor_speed"`
		TotalNumberCores string `json:"total_number_of_cores"`
		L2CacheCore      string `json:"l2_cache_per_core"`
		L3Cache          string `json:"l3_cache"`
		Memory           string `json:"memory"`
	} `json:"SPHardwareDataType"`
}

// getPlatformCPUInfo 获取平台CPU信息
func getPlatformCPUInfo(info *CPUInfo) error {
	return getDarwinCPUInfo(info)
}

// getPlatformCPUTemperature 获取平台CPU温度
func getPlatformCPUTemperature() (float64, error) {
	return getDarwinCPUTemperature()
}

// getPlatformCPUFrequency 获取平台CPU频率
func getPlatformCPUFrequency() (float64, error) {
	return getDarwinCPUFrequency()
}

// getPlatformCPUUsage 获取平台CPU使用率
func getPlatformCPUUsage() (*CPUUsage, error) {
	return getCPUStatsFromHostInfo()
}

// getDarwinCPUInfo 获取macOS CPU信息
func getDarwinCPUInfo(info *CPUInfo) error {
	// 1. 使用system_profiler获取详细硬件信息
	if err := getSystemProfilerInfo(info); err != nil {
		// 如果system_profiler失败，使用sysctl作为备选
		if err := getSysctlCPUInfo(info); err != nil {
			return fmt.Errorf("failed to get CPU info: %v", err)
		}
	}

	// 2. 获取Apple Silicon特有信息
	if IsAppleSilicon() {
		if err := getAppleSiliconCPUInfo(info); err != nil {
			// Apple Silicon信息获取失败不影响基本信息
			fmt.Printf("Warning: failed to get Apple Silicon info: %v\n", err)
		}
	}

	// 3. 获取缓存信息
	getCacheInfo(info)

	// 4. 获取频率信息
	if freq, err := getDarwinCPUFrequency(); err == nil {
		info.Frequency = freq
	}

	return nil
}

// getSystemProfilerInfo 使用system_profiler获取硬件信息
func getSystemProfilerInfo(info *CPUInfo) error {
	cmd := exec.Command("system_profiler", "SPHardwareDataType", "-json")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	var hardware SystemProfilerHardware
	if err := json.Unmarshal(output, &hardware); err != nil {
		return err
	}

	if len(hardware.SPHardwareDataType) == 0 {
		return fmt.Errorf("no hardware data found")
	}

	hw := hardware.SPHardwareDataType[0]

	// 设置CPU信息
	info.Model = hw.ProcessorName
	if hw.ChipType != "" {
		info.Model = hw.ChipType // 对于Apple Silicon，使用chip_type
	}

	// 解析核心数
	if cores, err := strconv.Atoi(hw.TotalNumberCores); err == nil {
		info.Cores = cores
	}

	// 解析频率
	if speedStr := hw.ProcessorSpeed; speedStr != "" {
		if freq, err := parseFrequency(speedStr); err == nil {
			info.Frequency = freq
		}
	}

	// 设置架构信息
	info.Architecture = "arm64"
	if strings.Contains(hw.ProcessorName, "Intel") {
		info.Architecture = "x86_64"
		info.Vendor = "Intel"
	} else if strings.Contains(hw.ChipType, "Apple") {
		info.Vendor = "Apple"
	}

	return nil
}

// getSysctlCPUInfo 使用sysctl获取CPU信息
func getSysctlCPUInfo(info *CPUInfo) error {
	// 获取CPU品牌字符串
	if brand, err := sysctlString("machdep.cpu.brand_string"); err == nil {
		info.Model = brand
	}

	// 获取CPU厂商
	if vendor, err := sysctlString("machdep.cpu.vendor"); err == nil {
		info.Vendor = vendor
	}

	// 获取CPU系列
	if family, err := sysctlString("machdep.cpu.family"); err == nil {
		info.Family = family
	}

	// 获取线程数
	if threads, err := sysctlUint64("machdep.cpu.thread_count"); err == nil {
		info.Threads = int(threads)
	}

	// 获取最大频率
	if maxFreq, err := sysctlUint64("hw.cpufrequency_max"); err == nil {
		info.MaxFrequency = float64(maxFreq) / 1000000000 // 转换为GHz
	}

	return nil
}

// getAppleSiliconCPUInfo 获取Apple Silicon特有信息
func getAppleSiliconCPUInfo(info *CPUInfo) error {
	// 使用sysctl获取性能核心和效率核心数量
	if pCores, err := sysctlUint64("hw.perflevel0.physicalcpu"); err == nil {
		info.PerformanceCores = int(pCores)
	}

	if eCores, err := sysctlUint64("hw.perflevel1.physicalcpu"); err == nil {
		info.EfficiencyCores = int(eCores)
	}

	// 验证核心数总和
	if info.PerformanceCores+info.EfficiencyCores != info.Cores {
		// 如果不匹配，尝试其他方法
		if err := getAppleSiliconCoreInfoAlternative(info); err != nil {
			return err
		}
	}

	return nil
}

// getAppleSiliconCoreInfoAlternative Apple Silicon核心信息的备选方法
func getAppleSiliconCoreInfoAlternative(info *CPUInfo) error {
	// 使用powermetrics获取详细信息（需要sudo权限）
	cmd := exec.Command("powermetrics", "--samplers", "cpu_power", "-n", "1", "--show-process-coalition")
	output, err := cmd.Output()
	if err != nil {
		// 如果powermetrics失败，使用默认估算
		return estimateAppleSiliconCores(info)
	}

	// 解析powermetrics输出
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "P-Cluster") {
			// 提取性能核心信息
			if count := extractCoreCount(line); count > 0 {
				info.PerformanceCores = count
			}
		} else if strings.Contains(line, "E-Cluster") {
			// 提取效率核心信息
			if count := extractCoreCount(line); count > 0 {
				info.EfficiencyCores = count
			}
		}
	}

	return nil
}

// estimateAppleSiliconCores 估算Apple Silicon核心配置
func estimateAppleSiliconCores(info *CPUInfo) error {
	// 根据总核心数估算P核心和E核心配置
	switch info.Cores {
	case 8: // M1
		info.PerformanceCores = 4
		info.EfficiencyCores = 4
	case 10: // M1 Pro
		info.PerformanceCores = 6
		info.EfficiencyCores = 4
	case 12: // M1 Max, M2 Pro
		info.PerformanceCores = 8
		info.EfficiencyCores = 4
	case 16: // M1 Ultra (2x M1 Max)
		info.PerformanceCores = 16
		info.EfficiencyCores = 0
	case 20: // M1 Ultra
		info.PerformanceCores = 16
		info.EfficiencyCores = 4
	default:
		// 对于未知配置，假设一半是性能核心
		info.PerformanceCores = info.Cores / 2
		info.EfficiencyCores = info.Cores - info.PerformanceCores
	}

	return nil
}

// getCacheInfo 获取缓存信息
func getCacheInfo(info *CPUInfo) {
	// L1缓存
	if l1i, err := sysctlUint64("hw.l1icachesize"); err == nil {
		if l1d, err := sysctlUint64("hw.l1dcachesize"); err == nil {
			info.CacheL1 = int((l1i + l1d) / 1024) // 转换为KB
		}
	}

	// L2缓存
	if l2, err := sysctlUint64("hw.l2cachesize"); err == nil {
		info.CacheL2 = int(l2 / 1024) // 转换为KB
	}

	// L3缓存
	if l3, err := sysctlUint64("hw.l3cachesize"); err == nil {
		info.CacheL3 = int(l3 / 1024) // 转换为KB
	}
}

// getDarwinCPUFrequency 获取CPU当前频率
func getDarwinCPUFrequency() (float64, error) {
	// 尝试从sysctl获取
	if freq, err := sysctlUint64("hw.cpufrequency"); err == nil {
		return float64(freq) / 1000000000, nil // 转换为GHz
	}

	// 备选方法：从system_profiler解析
	cmd := exec.Command("system_profiler", "SPHardwareDataType")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	// 查找处理器速度行
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "Processor Speed") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				speedStr := strings.TrimSpace(parts[1])
				return parseFrequency(speedStr)
			}
		}
	}

	return 0, fmt.Errorf("CPU frequency not found")
}

// getDarwinCPUTemperature 获取CPU温度
func getDarwinCPUTemperature() (float64, error) {
	// 尝试使用istats命令（如果安装了）
	cmd := exec.Command("istats", "cpu", "temp", "--value-only")
	output, err := cmd.Output()
	if err == nil {
		tempStr := strings.TrimSpace(string(output))
		if temp, err := strconv.ParseFloat(tempStr, 64); err == nil {
			return temp, nil
		}
	}

	// 尝试使用sensors命令（如果安装了）
	cmd = exec.Command("sensors")
	output, err = cmd.Output()
	if err == nil {
		// 解析sensors输出
		scanner := bufio.NewScanner(bytes.NewReader(output))
		for scanner.Scan() {
			line := scanner.Text()
			if strings.Contains(line, "CPU") && strings.Contains(line, "°C") {
				re := regexp.MustCompile(`(\d+\.?\d*)\s*°C`)
				matches := re.FindStringSubmatch(line)
				if len(matches) >= 2 {
					if temp, err := strconv.ParseFloat(matches[1], 64); err == nil {
						return temp, nil
					}
				}
			}
		}
	}

	return 0, fmt.Errorf("CPU temperature monitoring not available")
}

// getCPUStats 获取CPU统计信息
func getCPUStats() (*CPUStats, error) {
	// 这个函数用于计算差值，暂时返回空实现
	return &CPUStats{}, fmt.Errorf("use getPlatformCPUUsage instead")
}

// getPerCoreCPUUsage 获取每个核心的CPU使用率
func getPerCoreCPUUsage(duration time.Duration) ([]float64, error) {
	// macOS的每核心使用率需要更复杂的实现
	// 这里先返回空切片，表示不支持
	return nil, fmt.Errorf("per-core CPU usage not implemented for macOS")
}

// getAppleSiliconInfo 获取Apple Silicon详细信息
func getAppleSiliconInfo() (*AppleSiliconInfo, error) {
	info := &AppleSiliconInfo{}

	// 从system_profiler获取芯片信息
	cmd := exec.Command("system_profiler", "SPHardwareDataType", "-json")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var hardware SystemProfilerHardware
	if err := json.Unmarshal(output, &hardware); err != nil {
		return nil, err
	}

	if len(hardware.SPHardwareDataType) == 0 {
		return nil, fmt.Errorf("no hardware data found")
	}

	hw := hardware.SPHardwareDataType[0]
	info.ChipName = extractChipName(hw.ChipType)

	// 获取核心信息
	if pCores, err := sysctlUint64("hw.perflevel0.physicalcpu"); err == nil {
		info.PerformanceCores = int(pCores)
	}
	if eCores, err := sysctlUint64("hw.perflevel1.physicalcpu"); err == nil {
		info.EfficiencyCores = int(eCores)
	}

	// 估算其他信息
	info.GPUCores = estimateGPUCores(info.ChipName)
	info.NeuralCores = estimateNeuralCores(info.ChipName)
	info.MemoryBandwidth = estimateMemoryBandwidth(info.ChipName)
	info.ProcessNode = estimateProcessNode(info.ChipName)

	return info, nil
}

// 辅助函数

// sysctlString 获取字符串类型的sysctl值
func sysctlString(name string) (string, error) {
	// 实现sysctl系统调用
	return "", fmt.Errorf("not implemented")
}

// sysctlUint64 获取uint64类型的sysctl值
func sysctlUint64(name string) (uint64, error) {
	// 实现sysctl系统调用
	nameBytes := []byte(name + "\x00")

	// 先获取需要的缓冲区大小
	var size uintptr
	_, _, errno := syscall.Syscall6(
		syscall.SYS___SYSCTL,
		uintptr(unsafe.Pointer(&nameBytes[0])),
		uintptr(len(nameBytes)-1),
		0, // oldp
		uintptr(unsafe.Pointer(&size)),
		0, // newp
		0, // newlen
	)

	if errno != 0 {
		return 0, errno
	}

	// 分配缓冲区并获取实际值
	buf := make([]byte, size)
	_, _, errno = syscall.Syscall6(
		syscall.SYS___SYSCTL,
		uintptr(unsafe.Pointer(&nameBytes[0])),
		uintptr(len(nameBytes)-1),
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
		0, // newp
		0, // newlen
	)

	if errno != 0 {
		return 0, errno
	}

	// 转换为uint64
	if size == 8 {
		return *(*uint64)(unsafe.Pointer(&buf[0])), nil
	} else if size == 4 {
		return uint64(*(*uint32)(unsafe.Pointer(&buf[0]))), nil
	}

	return 0, fmt.Errorf("unexpected size: %d", size)
}

// parseFrequency 解析频率字符串
func parseFrequency(freqStr string) (float64, error) {
	// 移除单位并解析
	freqStr = strings.TrimSpace(freqStr)
	freqStr = strings.Replace(freqStr, " GHz", "", -1)
	freqStr = strings.Replace(freqStr, " MHz", "", -1)

	freq, err := strconv.ParseFloat(freqStr, 64)
	if err != nil {
		return 0, err
	}

	// 如果原字符串包含MHz，转换为GHz
	if strings.Contains(freqStr, "MHz") {
		freq = freq / 1000
	}

	return freq, nil
}

// extractCoreCount 从字符串中提取核心数量
func extractCoreCount(line string) int {
	re := regexp.MustCompile(`(\d+)\s*core`)
	matches := re.FindStringSubmatch(line)
	if len(matches) >= 2 {
		if count, err := strconv.Atoi(matches[1]); err == nil {
			return count
		}
	}
	return 0
}

// extractChipName 提取芯片名称
func extractChipName(chipType string) string {
	if strings.Contains(chipType, "M1") {
		return "M1"
	} else if strings.Contains(chipType, "M2") {
		return "M2"
	} else if strings.Contains(chipType, "M3") {
		return "M3"
	}
	return chipType
}

// 估算函数
func estimateGPUCores(chipName string) int {
	switch chipName {
	case "M1":
		return 7 // M1: 7核或8核GPU
	case "M1 Pro":
		return 14 // M1 Pro: 14核或16核GPU
	case "M1 Max":
		return 24 // M1 Max: 24核或32核GPU
	case "M2":
		return 8 // M2: 8核或10核GPU
	case "M2 Pro":
		return 16 // M2 Pro: 16核或19核GPU
	case "M2 Max":
		return 30 // M2 Max: 30核或38核GPU
	default:
		return 8
	}
}

func estimateNeuralCores(chipName string) int {
	// 大多数Apple Silicon都有16核Neural Engine
	return 16
}

func estimateMemoryBandwidth(chipName string) float64 {
	switch chipName {
	case "M1":
		return 68.25 // M1: 68.25 GB/s
	case "M1 Pro":
		return 200 // M1 Pro: 200 GB/s
	case "M1 Max":
		return 400 // M1 Max: 400 GB/s
	case "M2":
		return 100 // M2: 100 GB/s
	case "M2 Pro":
		return 200 // M2 Pro: 200 GB/s
	case "M2 Max":
		return 400 // M2 Max: 400 GB/s
	default:
		return 100
	}
}

func estimateProcessNode(chipName string) string {
	switch {
	case strings.HasPrefix(chipName, "M1"):
		return "5nm"
	case strings.HasPrefix(chipName, "M2"):
		return "5nm" // M2是改进的5nm工艺
	case strings.HasPrefix(chipName, "M3"):
		return "3nm"
	default:
		return "5nm"
	}
}

// getCPUStatsFromHostInfo 从host_processor_info获取CPU统计
func getCPUStatsFromHostInfo() (*CPUUsage, error) {
	// 使用top命令获取CPU使用率
	cmd := exec.Command("top", "-l", "1", "-n", "0")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	usage := &CPUUsage{
		LastUpdated: time.Now(),
	}

	// 解析top命令输出
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()

		// 查找CPU使用率行
		if strings.Contains(line, "CPU usage:") {
			// 解析 "CPU usage: 12.5% user, 6.25% sys, 81.25% idle"
			parts := strings.Split(line, ",")
			for _, part := range parts {
				part = strings.TrimSpace(part)
				if strings.Contains(part, "user") {
					if percent := extractPercentFromString(part); percent >= 0 {
						usage.User = percent
					}
				} else if strings.Contains(part, "sys") {
					if percent := extractPercentFromString(part); percent >= 0 {
						usage.System = percent
					}
				} else if strings.Contains(part, "idle") {
					if percent := extractPercentFromString(part); percent >= 0 {
						usage.Idle = percent
					}
				}
			}
		}

		// 查找负载平均值行
		if strings.Contains(line, "Load Avg:") {
			// 解析 "Load Avg: 1.23, 1.45, 1.67"
			if start := strings.Index(line, ":"); start != -1 {
				loadStr := strings.TrimSpace(line[start+1:])
				loads := strings.Split(loadStr, ",")
				if len(loads) >= 3 {
					if load1, err := strconv.ParseFloat(strings.TrimSpace(loads[0]), 64); err == nil {
						usage.LoadAvg1 = load1
					}
					if load5, err := strconv.ParseFloat(strings.TrimSpace(loads[1]), 64); err == nil {
						usage.LoadAvg5 = load5
					}
					if load15, err := strconv.ParseFloat(strings.TrimSpace(loads[2]), 64); err == nil {
						usage.LoadAvg15 = load15
					}
				}
			}
		}
	}

	// 计算总使用率
	usage.Overall = usage.User + usage.System
	if usage.Overall > 100 {
		usage.Overall = 100
	}

	return usage, nil
}

// extractPercentFromString 从字符串中提取百分比
func extractPercentFromString(s string) float64 {
	re := regexp.MustCompile(`(\d+\.?\d*)%`)
	matches := re.FindStringSubmatch(s)
	if len(matches) >= 2 {
		if percent, err := strconv.ParseFloat(matches[1], 64); err == nil {
			return percent
		}
	}
	return -1
}
