//go:build darwin

package disk

import (
	"bufio"
	"bytes"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

// getPlatformDisks 获取平台磁盘信息
func getPlatformDisks() ([]DiskInfo, error) {
	return getDarwinDisks()
}

// getPlatformDiskIOStats 获取平台磁盘I/O统计
func getPlatformDiskIOStats() ([]DiskIOStats, error) {
	return getDarwinDiskIOStats()
}

// getPlatformDiskHealth 获取平台磁盘健康信息
func getPlatformDiskHealth() ([]DiskHealth, error) {
	return getDarwinDiskHealth()
}

// getPlatformPartitions 获取平台分区信息
func getPlatformPartitions() ([]PartitionInfo, error) {
	return getDarwinPartitions()
}

// getDarwinDisks 获取macOS磁盘信息
func getDarwinDisks() ([]DiskInfo, error) {
	var disks []DiskInfo

	// 使用df命令获取磁盘使用情况
	cmd := exec.Command("df", "-k")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	// 跳过标题行
	if scanner.Scan() {
		// 标题行: Filesystem     1K-blocks      Used Available Use% Mounted on
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// 解析df输出
		if disk := parseDfLine(line); disk != nil {
			// 获取更多详细信息
			if err := getDarwinDiskDetails(disk); err == nil {
				disks = append(disks, *disk)
			}
		}
	}

	return disks, nil
}

// parseDfLine 解析df命令输出行
func parseDfLine(line string) *DiskInfo {
	fields := strings.Fields(line)
	if len(fields) < 6 {
		return nil
	}

	// 跳过不需要的文件系统
	filesystem := fields[0]
	if strings.HasPrefix(filesystem, "map ") ||
		strings.HasPrefix(filesystem, "devfs") ||
		strings.HasPrefix(filesystem, "fdesc") ||
		strings.Contains(filesystem, "com.apple") {
		return nil
	}

	// 解析数值 (df -k 输出的是1K blocks)
	total, _ := strconv.ParseUint(fields[1], 10, 64)
	used, _ := strconv.ParseUint(fields[2], 10, 64)
	available, _ := strconv.ParseUint(fields[3], 10, 64)

	// 转换为字节
	total *= 1024
	used *= 1024
	available *= 1024

	mountpoint := fields[5]

	return &DiskInfo{
		Device:     filesystem,
		Mountpoint: mountpoint,
		Total:      total,
		Used:       used,
		Available:  available,
	}
}

// getDarwinDiskDetails 获取macOS磁盘详细信息
func getDarwinDiskDetails(disk *DiskInfo) error {
	// 使用mount命令获取文件系统类型
	if err := getMountInfo(disk); err != nil {
		// 如果mount命令失败，使用diskutil作为备选
		getDiskutilInfo(disk)
	}

	// 使用stat系统调用获取更精确的信息
	getStatfsInfo(disk)

	return nil
}

// getMountInfo 从mount命令获取挂载信息
func getMountInfo(disk *DiskInfo) error {
	cmd := exec.Command("mount")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()

		// 查找匹配的挂载点
		if strings.Contains(line, disk.Mountpoint) {
			// 解析mount输出: /dev/disk1s1 on / (apfs, local, read-only, journaled, noatime)
			parts := strings.Split(line, " on ")
			if len(parts) >= 2 {
				device := parts[0]
				remaining := parts[1]

				// 提取文件系统类型和选项
				if idx := strings.Index(remaining, " ("); idx != -1 {
					options := remaining[idx+2:]
					if idx := strings.Index(options, ")"); idx != -1 {
						options = options[:idx]
						parts := strings.Split(options, ", ")
						if len(parts) > 0 {
							disk.FileSystem = parts[0]
							disk.IsReadOnly = strings.Contains(options, "read-only")
						}
					}
				}

				// 更新设备名称
				if device != "" {
					disk.Device = device
				}
			}
			break
		}
	}

	return nil
}

// getDiskutilInfo 使用diskutil获取磁盘信息
func getDiskutilInfo(disk *DiskInfo) error {
	// 使用diskutil info获取详细信息
	cmd := exec.Command("diskutil", "info", disk.Mountpoint)
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "File System Personality:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				disk.FileSystem = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "Device / Media Name:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				disk.Device = strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(line, "Read-Only") {
			disk.IsReadOnly = strings.Contains(line, "Yes")
		}
	}

	return nil
}

// getStatfsInfo 使用statfs系统调用获取更精确的信息
func getStatfsInfo(disk *DiskInfo) error {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(disk.Mountpoint, &stat); err != nil {
		return err
	}

	// 计算更精确的大小信息
	blockSize := uint64(stat.Bsize)
	disk.Total = stat.Blocks * blockSize
	disk.Available = stat.Bavail * blockSize
	disk.Used = disk.Total - disk.Available

	// inode信息
	disk.InodesTotal = stat.Files
	disk.InodesUsed = stat.Files - stat.Ffree

	return nil
}

// getDarwinDiskIOStats 获取macOS磁盘I/O统计
func getDarwinDiskIOStats() ([]DiskIOStats, error) {
	var stats []DiskIOStats

	// 使用iostat命令获取I/O统计
	cmd := exec.Command("iostat", "-d", "1", "1")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// 解析iostat输出
	if ioStats := parseIostatOutput(string(output)); ioStats != nil {
		stats = append(stats, ioStats...)
	}

	return stats, nil
}

// parseIostatOutput 解析iostat输出
func parseIostatOutput(output string) []DiskIOStats {
	var stats []DiskIOStats

	lines := strings.Split(output, "\n")
	headerFound := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 查找数据行标题
		if strings.Contains(line, "device") && strings.Contains(line, "r/s") {
			headerFound = true
			continue
		}

		if headerFound && !strings.HasPrefix(line, "disk") {
			continue
		}

		if headerFound {
			if stat := parseIostatLine(line); stat != nil {
				stats = append(stats, *stat)
			}
		}
	}

	return stats
}

// parseIostatLine 解析iostat数据行
func parseIostatLine(line string) *DiskIOStats {
	fields := strings.Fields(line)
	if len(fields) < 6 {
		return nil
	}

	// iostat输出格式: device  r/s   w/s    KB/r   KB/w  wait svc_t %busy
	device := fields[0]

	// 解析数值
	readOps, _ := strconv.ParseFloat(fields[1], 64)
	writeOps, _ := strconv.ParseFloat(fields[2], 64)
	readKB, _ := strconv.ParseFloat(fields[3], 64)
	writeKB, _ := strconv.ParseFloat(fields[4], 64)

	return &DiskIOStats{
		Device:     device,
		ReadCount:  uint64(readOps),
		WriteCount: uint64(writeOps),
		ReadBytes:  uint64(readKB * 1024),
		WriteBytes: uint64(writeKB * 1024),
	}
}

// getDarwinDiskHealth 获取macOS磁盘健康信息
func getDarwinDiskHealth() ([]DiskHealth, error) {
	// 使用system_profiler获取存储设备信息
	cmd := exec.Command("system_profiler", "SPStorageDataType", "-json")
	output, err := cmd.Output()
	if err != nil {
		// 如果system_profiler失败，尝试其他方法
		return getDiskHealthFromDiskutil()
	}

	// 解析system_profiler JSON输出
	// 这里需要json解析，为简化实现，使用文本解析
	return parseDiskHealthFromSystemProfiler(string(output))
}

// getDiskHealthFromDiskutil 使用diskutil获取健康信息
func getDiskHealthFromDiskutil() ([]DiskHealth, error) {
	var healthInfo []DiskHealth

	// 获取所有磁盘列表
	cmd := exec.Command("diskutil", "list")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// 解析diskutil list输出，提取磁盘设备
	devices := extractDiskDevices(string(output))

	// 为每个设备获取详细信息
	for _, device := range devices {
		if health := getDiskHealthForDevice(device); health != nil {
			healthInfo = append(healthInfo, *health)
		}
	}

	return healthInfo, nil
}

// extractDiskDevices 从diskutil list输出中提取磁盘设备
func extractDiskDevices(output string) []string {
	var devices []string

	re := regexp.MustCompile(`/dev/(disk\d+)`)
	matches := re.FindAllStringSubmatch(output, -1)

	seen := make(map[string]bool)
	for _, match := range matches {
		if len(match) >= 2 {
			device := match[1]
			if !seen[device] {
				devices = append(devices, device)
				seen[device] = true
			}
		}
	}

	return devices
}

// getDiskHealthForDevice 获取指定设备的健康信息
func getDiskHealthForDevice(device string) *DiskHealth {
	health := &DiskHealth{
		Device: device,
	}

	// 使用diskutil info获取基本信息
	cmd := exec.Command("diskutil", "info", device)
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "Device / Media Name:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				health.Model = strings.TrimSpace(parts[1])
			}
		} else if strings.HasPrefix(line, "Disk Size:") {
			if size := extractSizeFromDiskutilLine(line); size > 0 {
				health.Capacity = size
			}
		} else if strings.HasPrefix(line, "Protocol:") {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				health.Interface = strings.TrimSpace(parts[1])
			}
		}
	}

	// 尝试使用smartctl获取SMART信息（如果安装了）
	if smartInfo := getSMARTInfo(device); smartInfo != nil {
		mergeSMARTInfo(health, smartInfo)
	}

	// 设置默认健康度
	if health.HealthPercentage == 0 {
		health.HealthPercentage = 100 // 默认100%健康
		health.RemainingLife = 100
	}

	return health
}

// getSMARTInfo 获取SMART信息
func getSMARTInfo(device string) map[string]string {
	// 尝试使用smartctl（需要安装smartmontools）
	cmd := exec.Command("smartctl", "-a", "/dev/"+device)
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	smartInfo := make(map[string]string)
	scanner := bufio.NewScanner(bytes.NewReader(output))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 解析SMART属性
		if strings.Contains(line, "Temperature_Celsius") {
			if temp := extractSMARTValue(line); temp != "" {
				smartInfo["temperature"] = temp
			}
		} else if strings.Contains(line, "Power_On_Hours") {
			if hours := extractSMARTValue(line); hours != "" {
				smartInfo["power_on_hours"] = hours
			}
		} else if strings.Contains(line, "Power_Cycle_Count") {
			if cycles := extractSMARTValue(line); cycles != "" {
				smartInfo["power_cycles"] = cycles
			}
		}
	}

	return smartInfo
}

// mergeSMARTInfo 合并SMART信息到健康信息中
func mergeSMARTInfo(health *DiskHealth, smartInfo map[string]string) {
	if temp, exists := smartInfo["temperature"]; exists {
		if t, err := strconv.ParseFloat(temp, 64); err == nil {
			health.Temperature = t
		}
	}

	if hours, exists := smartInfo["power_on_hours"]; exists {
		if h, err := strconv.ParseUint(hours, 10, 64); err == nil {
			health.PowerOnHours = h
		}
	}

	if cycles, exists := smartInfo["power_cycles"]; exists {
		if c, err := strconv.ParseUint(cycles, 10, 64); err == nil {
			health.PowerCycles = c
		}
	}
}

// getDarwinPartitions 获取macOS分区信息
func getDarwinPartitions() ([]PartitionInfo, error) {
	var partitions []PartitionInfo

	// 使用diskutil list获取分区信息
	cmd := exec.Command("diskutil", "list")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	// 解析diskutil list输出
	partitions = parseDiskutilList(string(output))

	return partitions, nil
}

// parseDiskutilList 解析diskutil list输出
func parseDiskutilList(output string) []PartitionInfo {
	var partitions []PartitionInfo

	lines := strings.Split(output, "\n")
	var currentDisk string

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// 检查是否是磁盘标题行
		if strings.HasPrefix(line, "/dev/disk") {
			currentDisk = line
			continue
		}

		// 解析分区行
		if strings.Contains(line, ":") && currentDisk != "" {
			if partition := parsePartitionLine(line, currentDisk); partition != nil {
				partitions = append(partitions, *partition)
			}
		}
	}

	return partitions
}

// parsePartitionLine 解析分区行
func parsePartitionLine(line, disk string) *PartitionInfo {
	// diskutil list分区行格式: 1:                        GUID_partition_scheme                     *500.3 GB   disk0
	parts := strings.Split(line, ":")
	if len(parts) < 2 {
		return nil
	}

	remaining := strings.TrimSpace(parts[1])
	fields := strings.Fields(remaining)
	if len(fields) < 2 {
		return nil
	}

	partition := &PartitionInfo{
		Device:        disk + "s" + strings.TrimSpace(parts[0]),
		PartitionType: fields[0],
	}

	// 检查是否是系统分区
	if strings.Contains(remaining, "Apple_Boot") || strings.Contains(remaining, "EFI") {
		partition.IsBootable = true
	}

	if strings.Contains(remaining, "Apple_APFS") || strings.Contains(remaining, "Apple_HFS") {
		partition.IsSystem = true
	}

	return partition
}

// 辅助函数

// extractSizeFromDiskutilLine 从diskutil输出行中提取大小
func extractSizeFromDiskutilLine(line string) uint64 {
	re := regexp.MustCompile(`(\d+(?:\.\d+)?)\s*([KMGTPE]?B)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) >= 3 {
		if size, err := strconv.ParseFloat(matches[1], 64); err == nil {
			multiplier := uint64(1)
			switch matches[2] {
			case "KB":
				multiplier = 1000
			case "MB":
				multiplier = 1000 * 1000
			case "GB":
				multiplier = 1000 * 1000 * 1000
			case "TB":
				multiplier = 1000 * 1000 * 1000 * 1000
			}
			return uint64(size * float64(multiplier))
		}
	}
	return 0
}

// extractSMARTValue 从SMART行中提取数值
func extractSMARTValue(line string) string {
	// SMART行格式通常是: ID# ATTRIBUTE_NAME          FLAG     VALUE WORST THRESH TYPE      UPDATED  WHEN_FAILED RAW_VALUE
	fields := strings.Fields(line)
	if len(fields) >= 10 {
		return fields[9] // RAW_VALUE通常在最后一列
	}
	return ""
}

// parseDiskHealthFromSystemProfiler 从system_profiler输出解析健康信息
func parseDiskHealthFromSystemProfiler(output string) ([]DiskHealth, error) {
	// 这里需要完整的JSON解析实现
	// 为简化，返回空结果
	return []DiskHealth{}, nil
}
