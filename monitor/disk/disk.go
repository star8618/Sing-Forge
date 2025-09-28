// Package disk 提供跨平台磁盘监控功能
package disk

import (
	"fmt"
	"time"
)

// DiskInfo 磁盘基本信息
type DiskInfo struct {
	Device            string    `json:"device"`              // 设备名称
	Mountpoint        string    `json:"mountpoint"`          // 挂载点
	FileSystem        string    `json:"filesystem"`          // 文件系统类型
	Total             uint64    `json:"total"`               // 总容量 (bytes)
	Used              uint64    `json:"used"`                // 已用容量 (bytes)
	Available         uint64    `json:"available"`           // 可用容量 (bytes)
	UsedPercent       float64   `json:"used_percent"`        // 使用率百分比
	InodesTotal       uint64    `json:"inodes_total"`        // 总inode数
	InodesUsed        uint64    `json:"inodes_used"`         // 已用inode数
	InodesUsedPercent float64   `json:"inodes_used_percent"` // inode使用率
	IsReadOnly        bool      `json:"is_readonly"`         // 是否只读
	LastUpdated       time.Time `json:"last_updated"`        // 最后更新时间
}

// DiskIOStats 磁盘I/O统计信息
type DiskIOStats struct {
	Device         string    `json:"device"`           // 设备名称
	ReadCount      uint64    `json:"read_count"`       // 读取次数
	WriteCount     uint64    `json:"write_count"`      // 写入次数
	ReadBytes      uint64    `json:"read_bytes"`       // 读取字节数
	WriteBytes     uint64    `json:"write_bytes"`      // 写入字节数
	ReadTime       uint64    `json:"read_time"`        // 读取时间 (ms)
	WriteTime      uint64    `json:"write_time"`       // 写入时间 (ms)
	IOTime         uint64    `json:"io_time"`          // I/O时间 (ms)
	WeightedIOTime uint64    `json:"weighted_io_time"` // 加权I/O时间 (ms)
	IopsInProgress uint64    `json:"iops_in_progress"` // 进行中的I/O操作数
	LastUpdated    time.Time `json:"last_updated"`     // 最后更新时间
}

// DiskSpeed 磁盘速度信息
type DiskSpeed struct {
	Device          string    `json:"device"`            // 设备名称
	ReadSpeed       uint64    `json:"read_speed"`        // 读取速度 (bytes/s)
	WriteSpeed      uint64    `json:"write_speed"`       // 写入速度 (bytes/s)
	ReadIOPS        uint64    `json:"read_iops"`         // 读取IOPS
	WriteIOPS       uint64    `json:"write_iops"`        // 写入IOPS
	AvgReadLatency  float64   `json:"avg_read_latency"`  // 平均读延迟 (ms)
	AvgWriteLatency float64   `json:"avg_write_latency"` // 平均写延迟 (ms)
	Utilization     float64   `json:"utilization"`       // 利用率百分比
	LastUpdated     time.Time `json:"last_updated"`      // 最后更新时间
}

// DiskHealth 磁盘健康信息 (主要针对SSD/NVMe)
type DiskHealth struct {
	Device            string    `json:"device"`              // 设备名称
	Model             string    `json:"model"`               // 型号
	Serial            string    `json:"serial"`              // 序列号
	Firmware          string    `json:"firmware"`            // 固件版本
	Interface         string    `json:"interface"`           // 接口类型 (SATA, NVMe, etc.)
	Capacity          uint64    `json:"capacity"`            // 容量 (bytes)
	Temperature       float64   `json:"temperature"`         // 温度 (℃)
	PowerOnHours      uint64    `json:"power_on_hours"`      // 通电时间 (小时)
	PowerCycles       uint64    `json:"power_cycles"`        // 通电次数
	TotalBytesWritten uint64    `json:"total_bytes_written"` // 总写入字节数
	TotalBytesRead    uint64    `json:"total_bytes_read"`    // 总读取字节数
	WearLevelingCount uint64    `json:"wear_leveling_count"` // 磨损平衡计数
	ProgramFailCount  uint64    `json:"program_fail_count"`  // 编程失败计数
	EraseFailCount    uint64    `json:"erase_fail_count"`    // 擦除失败计数
	HealthPercentage  float64   `json:"health_percentage"`   // 健康度百分比
	RemainingLife     float64   `json:"remaining_life"`      // 剩余寿命百分比
	CriticalWarning   bool      `json:"critical_warning"`    // 严重警告
	LastUpdated       time.Time `json:"last_updated"`        // 最后更新时间
}

// PartitionInfo 分区信息
type PartitionInfo struct {
	Device        string `json:"device"`         // 设备名称
	Mountpoint    string `json:"mountpoint"`     // 挂载点
	FileSystem    string `json:"filesystem"`     // 文件系统
	Options       string `json:"options"`        // 挂载选项
	IsBootable    bool   `json:"is_bootable"`    // 是否可启动
	IsSystem      bool   `json:"is_system"`      // 是否系统分区
	PartitionType string `json:"partition_type"` // 分区类型
}

var (
	lastDiskIOStats     map[string]*DiskIOStats
	lastDiskIOStatsTime time.Time
)

// GetDisks 获取所有磁盘信息
func GetDisks() ([]DiskInfo, error) {
	var disks []DiskInfo

	// 根据平台获取磁盘信息
	var err error
	disks, err = getPlatformDisks()

	if err != nil {
		return nil, err
	}

	// 更新时间戳
	now := time.Now()
	for i := range disks {
		disks[i].LastUpdated = now
		// 计算使用率
		if disks[i].Total > 0 {
			disks[i].UsedPercent = float64(disks[i].Used) / float64(disks[i].Total) * 100
		}
		// 计算inode使用率
		if disks[i].InodesTotal > 0 {
			disks[i].InodesUsedPercent = float64(disks[i].InodesUsed) / float64(disks[i].InodesTotal) * 100
		}
	}

	return disks, nil
}

// GetDiskIOStats 获取磁盘I/O统计
func GetDiskIOStats() ([]DiskIOStats, error) {
	var stats []DiskIOStats

	// 根据平台获取I/O统计
	var err error
	stats, err = getPlatformDiskIOStats()

	if err != nil {
		return nil, err
	}

	// 更新时间戳
	now := time.Now()
	for i := range stats {
		stats[i].LastUpdated = now
	}

	return stats, nil
}

// GetDiskSpeed 获取磁盘实时速度
func GetDiskSpeed() ([]DiskSpeed, error) {
	return GetDiskSpeedWithInterval(1 * time.Second)
}

// GetDiskSpeedWithInterval 获取指定间隔的磁盘速度
func GetDiskSpeedWithInterval(interval time.Duration) ([]DiskSpeed, error) {
	// 获取当前I/O统计
	currentStats, err := GetDiskIOStats()
	if err != nil {
		return nil, err
	}

	// 转换为map以便查找
	currentStatsMap := make(map[string]*DiskIOStats)
	for i := range currentStats {
		currentStatsMap[currentStats[i].Device] = &currentStats[i]
	}

	var speeds []DiskSpeed
	now := time.Now()

	// 如果有上次的统计数据，计算速度
	if lastDiskIOStats != nil && !lastDiskIOStatsTime.IsZero() {
		timeDiff := now.Sub(lastDiskIOStatsTime).Seconds()

		if timeDiff > 0 && timeDiff < 60 { // 防止异常的时间差
			for device, currentStat := range currentStatsMap {
				if lastStat, exists := lastDiskIOStats[device]; exists {
					speed := calculateDiskSpeed(lastStat, currentStat, timeDiff)
					speeds = append(speeds, speed)
				}
			}
		}
	} else {
		// 第一次调用，等待一个间隔后再次获取
		time.Sleep(interval)
		return GetDiskSpeedWithInterval(interval)
	}

	// 更新缓存
	lastDiskIOStats = currentStatsMap
	lastDiskIOStatsTime = now

	return speeds, nil
}

// calculateDiskSpeed 计算磁盘速度
func calculateDiskSpeed(last, current *DiskIOStats, timeDiff float64) DiskSpeed {
	speed := DiskSpeed{
		Device:      current.Device,
		LastUpdated: current.LastUpdated,
	}

	// 计算读写速度
	if current.ReadBytes >= last.ReadBytes {
		speed.ReadSpeed = uint64(float64(current.ReadBytes-last.ReadBytes) / timeDiff)
	}
	if current.WriteBytes >= last.WriteBytes {
		speed.WriteSpeed = uint64(float64(current.WriteBytes-last.WriteBytes) / timeDiff)
	}

	// 计算IOPS
	if current.ReadCount >= last.ReadCount {
		speed.ReadIOPS = uint64(float64(current.ReadCount-last.ReadCount) / timeDiff)
	}
	if current.WriteCount >= last.WriteCount {
		speed.WriteIOPS = uint64(float64(current.WriteCount-last.WriteCount) / timeDiff)
	}

	// 计算平均延迟
	if speed.ReadIOPS > 0 && current.ReadTime >= last.ReadTime {
		speed.AvgReadLatency = float64(current.ReadTime-last.ReadTime) / float64(current.ReadCount-last.ReadCount)
	}
	if speed.WriteIOPS > 0 && current.WriteTime >= last.WriteTime {
		speed.AvgWriteLatency = float64(current.WriteTime-last.WriteTime) / float64(current.WriteCount-last.WriteCount)
	}

	// 计算利用率
	if current.IOTime >= last.IOTime {
		speed.Utilization = float64(current.IOTime-last.IOTime) / (timeDiff * 1000) * 100
		if speed.Utilization > 100 {
			speed.Utilization = 100
		}
	}

	return speed
}

// GetDiskHealth 获取磁盘健康信息
func GetDiskHealth() ([]DiskHealth, error) {
	var healthInfo []DiskHealth

	// 根据平台获取健康信息
	var err error
	healthInfo, err = getPlatformDiskHealth()

	if err != nil {
		return nil, err
	}

	// 更新时间戳
	now := time.Now()
	for i := range healthInfo {
		healthInfo[i].LastUpdated = now
	}

	return healthInfo, nil
}

// GetPartitions 获取分区信息
func GetPartitions() ([]PartitionInfo, error) {
	var partitions []PartitionInfo

	// 根据平台获取分区信息
	var err error
	partitions, err = getPlatformPartitions()

	return partitions, err
}

// GetSummary 获取磁盘概览信息
func GetSummary() (map[string]interface{}, error) {
	result := make(map[string]interface{})

	// 获取磁盘信息
	if disks, err := GetDisks(); err == nil {
		result["disks"] = disks

		// 计算总计
		var totalSpace, usedSpace, availableSpace uint64
		for _, disk := range disks {
			totalSpace += disk.Total
			usedSpace += disk.Used
			availableSpace += disk.Available
		}

		result["summary"] = map[string]interface{}{
			"total_space":     totalSpace,
			"used_space":      usedSpace,
			"available_space": availableSpace,
			"used_percent":    float64(usedSpace) / float64(totalSpace) * 100,
			"disk_count":      len(disks),
		}
	}

	// 获取I/O统计
	if ioStats, err := GetDiskIOStats(); err == nil {
		result["io_stats"] = ioStats
	}

	// 获取速度信息
	if speeds, err := GetDiskSpeed(); err == nil {
		result["speeds"] = speeds
	}

	// 获取健康信息
	if health, err := GetDiskHealth(); err == nil {
		result["health"] = health
	}

	// 获取分区信息
	if partitions, err := GetPartitions(); err == nil {
		result["partitions"] = partitions
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

// FormatSpeed 格式化速度为可读格式
func FormatSpeed(bytesPerSecond uint64) string {
	const unit = 1024
	if bytesPerSecond < unit {
		return fmt.Sprintf("%d B/s", bytesPerSecond)
	}
	div, exp := uint64(unit), 0
	for n := bytesPerSecond / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB/s", float64(bytesPerSecond)/float64(div), "KMGTPE"[exp])
}

// MonitorRealTime 实时监控磁盘速度 (返回channel)
func MonitorRealTime(interval time.Duration) (<-chan []DiskSpeed, <-chan error) {
	speedChan := make(chan []DiskSpeed)
	errorChan := make(chan error)

	go func() {
		defer close(speedChan)
		defer close(errorChan)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			speeds, err := GetDiskSpeed()
			if err != nil {
				errorChan <- err
				continue
			}
			speedChan <- speeds
		}
	}()

	return speedChan, errorChan
}
