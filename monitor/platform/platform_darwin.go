//go:build darwin

package platform

import (
	"bufio"
	"bytes"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// getPlatformInfo 获取平台信息
func getPlatformInfo(info *PlatformInfo) error {
	return getDarwinPlatformInfo(info)
}

// getPlatformHardwarePlatform 获取硬件平台
func getPlatformHardwarePlatform(hardware *HardwarePlatform) error {
	return getDarwinHardwarePlatform(hardware)
}

// setPlatformCapabilities 设置平台能力
func setPlatformCapabilities(caps *Capabilities) {
	setDarwinCapabilities(caps)
}

// isPlatformVirtualMachine 检查是否虚拟机
func isPlatformVirtualMachine() (bool, error) {
	vm, err := isDarwinVirtualMachine()
	return vm, err
}

// isPlatformContainer 检查是否容器
func isPlatformContainer() (bool, error) {
	container, err := isDarwinContainer()
	return container, err
}

// getDarwinPlatformInfo 获取macOS平台信息
func getDarwinPlatformInfo(info *PlatformInfo) error {
	// 获取内核版本
	if kernel, err := getKernelVersion(); err == nil {
		info.Kernel = kernel
	}

	// 获取系统版本
	if version, err := getSystemVersion(); err == nil {
		info.Version = version
	}

	// 获取构建号
	if build, err := getBuildNumber(); err == nil {
		info.BuildNumber = build
	}

	// 获取主机名
	if hostname, err := os.Hostname(); err == nil {
		info.Hostname = hostname
	}

	// 获取运行时间
	if uptime, err := getUptime(); err == nil {
		info.Uptime = uptime
		info.BootTime = time.Now().Add(-time.Duration(uptime) * time.Second)
	}

	return nil
}

// getDarwinHardwarePlatform 获取macOS硬件平台信息
func getDarwinHardwarePlatform(hardware *HardwarePlatform) error {
	// 使用system_profiler获取硬件信息
	if err := getSystemProfilerHardware(hardware); err != nil {
		// 如果system_profiler失败，使用sysctl作为备选
		getSysctlHardware(hardware)
	}

	// 检查是否为虚拟机
	if isVM, err := isDarwinVirtualMachine(); err == nil {
		hardware.IsVirtual = isVM
	}

	// 检查是否为容器
	if isContainer, err := isDarwinContainer(); err == nil {
		hardware.IsContainer = isContainer
	}

	return nil
}

// setDarwinCapabilities 设置macOS平台能力
func setDarwinCapabilities(caps *Capabilities) {
	// macOS的监控能力
	caps.CPUTemperature = true // 支持温度监控
	caps.CPUFrequency = true   // 支持频率监控
	caps.PerCoreUsage = false  // 暂不支持每核心使用率
	caps.MemoryPressure = true // 支持内存压力监控
	caps.DiskHealth = true     // 支持磁盘健康监控
	caps.NetworkDetails = true // 支持详细网络信息
	caps.ProcessDetails = true // 支持详细进程信息

	// 硬件信息能力
	caps.GPUInfo = true     // 支持GPU信息
	caps.BatteryInfo = true // 支持电池信息
	caps.SensorInfo = true  // 支持传感器信息

	// 系统能力
	caps.ContainerSupport = true      // 支持容器
	caps.VirtualizationSupport = true // 支持虚拟化

	// Apple Silicon特殊优化
	if IsAppleSilicon() {
		caps.PerCoreUsage = true // Apple Silicon支持P/E核心监控
	}
}

// isDarwinVirtualMachine 检查是否为虚拟机
func isDarwinVirtualMachine() (bool, error) {
	// 检查系统信息中是否包含虚拟机标识
	cmd := exec.Command("system_profiler", "SPHardwareDataType")
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	outputStr := string(output)

	// 检查常见的虚拟机标识
	vmIndicators := []string{
		"VMware",
		"VirtualBox",
		"Parallels",
		"QEMU",
		"Virtual Machine",
		"VM",
	}

	for _, indicator := range vmIndicators {
		if strings.Contains(outputStr, indicator) {
			return true, nil
		}
	}

	// 检查CPU型号是否包含虚拟机标识
	if brand, err := sysctlString("machdep.cpu.brand_string"); err == nil {
		for _, indicator := range vmIndicators {
			if strings.Contains(brand, indicator) {
				return true, nil
			}
		}
	}

	return false, nil
}

// isDarwinContainer 检查是否运行在容器中
func isDarwinContainer() (bool, error) {
	// 检查是否存在容器相关的环境变量
	containerEnvs := []string{
		"DOCKER_CONTAINER",
		"container",
		"KUBERNETES_SERVICE_HOST",
		"K8S_POD_NAME",
	}

	for _, env := range containerEnvs {
		if _, exists := os.LookupEnv(env); exists {
			return true, nil
		}
	}

	// 检查是否存在容器相关的文件
	containerFiles := []string{
		"/.dockerenv",
		"/run/.containerenv",
	}

	for _, file := range containerFiles {
		if _, err := os.Stat(file); err == nil {
			return true, nil
		}
	}

	return false, nil
}

// 辅助函数

// getKernelVersion 获取内核版本
func getKernelVersion() (string, error) {
	cmd := exec.Command("uname", "-r")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// getSystemVersion 获取系统版本
func getSystemVersion() (string, error) {
	cmd := exec.Command("sw_vers", "-productVersion")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// getBuildNumber 获取构建号
func getBuildNumber() (string, error) {
	cmd := exec.Command("sw_vers", "-buildVersion")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// getUptime 获取系统运行时间
func getUptime() (uint64, error) {
	// 使用sysctl获取启动时间
	var boottime syscall.Timeval
	mib := []int32{1, 21} // CTL_KERN, KERN_BOOTTIME

	// 避免未使用变量警告
	_ = boottime
	_ = mib

	// 这里需要系统调用实现，简化版本使用uptime命令
	cmd := exec.Command("uptime")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	// 解析uptime输出
	return parseUptimeOutput(string(output))
}

// parseUptimeOutput 解析uptime命令输出
func parseUptimeOutput(output string) (uint64, error) {
	// uptime输出格式: "10:30  up 2 days,  3:45, 2 users, load averages: 1.23 1.45 1.67"

	if strings.Contains(output, " day") {
		// 包含天数
		parts := strings.Split(output, " up ")
		if len(parts) >= 2 {
			uptimePart := strings.Split(parts[1], ",")[0]
			return parseDaysUptime(uptimePart)
		}
	} else if strings.Contains(output, ":") {
		// 只有小时:分钟
		parts := strings.Split(output, " up ")
		if len(parts) >= 2 {
			uptimePart := strings.Split(parts[1], ",")[0]
			return parseHoursUptime(uptimePart)
		}
	}

	return 0, nil
}

// parseDaysUptime 解析包含天数的运行时间
func parseDaysUptime(uptime string) (uint64, error) {
	// 格式: "2 days, 3:45" 或 "2 days"
	parts := strings.Fields(uptime)
	if len(parts) >= 2 {
		days, err := strconv.ParseUint(parts[0], 10, 64)
		if err != nil {
			return 0, err
		}

		totalSeconds := days * 24 * 3600

		// 检查是否还有小时:分钟部分
		if len(parts) >= 3 && strings.Contains(parts[2], ":") {
			timePart := strings.TrimSuffix(parts[2], ",")
			if hoursMinutes, err := parseTimeString(timePart); err == nil {
				totalSeconds += hoursMinutes
			}
		}

		return totalSeconds, nil
	}

	return 0, nil
}

// parseHoursUptime 解析只有小时的运行时间
func parseHoursUptime(uptime string) (uint64, error) {
	uptime = strings.TrimSpace(uptime)
	return parseTimeString(uptime)
}

// parseTimeString 解析时间字符串 "3:45"
func parseTimeString(timeStr string) (uint64, error) {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return 0, nil
	}

	hours, err := strconv.ParseUint(parts[0], 10, 64)
	if err != nil {
		return 0, err
	}

	minutes, err := strconv.ParseUint(parts[1], 10, 64)
	if err != nil {
		return 0, err
	}

	return hours*3600 + minutes*60, nil
}

// getSystemProfilerHardware 使用system_profiler获取硬件信息
func getSystemProfilerHardware(hardware *HardwarePlatform) error {
	cmd := exec.Command("system_profiler", "SPHardwareDataType")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "Model Name:") {
			hardware.Model = extractValue(line)
		} else if strings.HasPrefix(line, "Model Identifier:") {
			if hardware.Model == "" {
				hardware.Model = extractValue(line)
			}
		} else if strings.HasPrefix(line, "Serial Number:") {
			hardware.Serial = extractValue(line)
		} else if strings.HasPrefix(line, "Hardware UUID:") {
			hardware.UUID = extractValue(line)
		} else if strings.Contains(line, "Chip") || strings.Contains(line, "Processor") {
			if hardware.Vendor == "" {
				if strings.Contains(line, "Apple") {
					hardware.Vendor = "Apple"
				} else if strings.Contains(line, "Intel") {
					hardware.Vendor = "Intel"
				}
			}
		}
	}

	return nil
}

// getSysctlHardware 使用sysctl获取硬件信息
func getSysctlHardware(hardware *HardwarePlatform) error {
	// 获取机器型号
	if model, err := sysctlString("hw.model"); err == nil {
		hardware.Model = model
	}

	// 获取CPU品牌
	if brand, err := sysctlString("machdep.cpu.brand_string"); err == nil {
		if strings.Contains(brand, "Apple") {
			hardware.Vendor = "Apple"
		} else if strings.Contains(brand, "Intel") {
			hardware.Vendor = "Intel"
		}
	}

	return nil
}

// sysctlString 获取字符串类型的sysctl值
func sysctlString(name string) (string, error) {
	cmd := exec.Command("sysctl", "-n", name)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// extractValue 从"Key: Value"格式的行中提取值
func extractValue(line string) string {
	parts := strings.Split(line, ":")
	if len(parts) >= 2 {
		return strings.TrimSpace(parts[1])
	}
	return ""
}
