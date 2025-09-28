// Package network 提供跨平台实时网络监控功能
package network

import (
	"fmt"
	"sort"
	"time"
)

// NetworkInterface 网络接口信息
type NetworkInterface struct {
	Name        string    `json:"name"`         // 接口名称
	DisplayName string    `json:"display_name"` // 显示名称
	Hardware    string    `json:"hardware"`     // 硬件类型 (ethernet, wifi, etc.)
	MAC         string    `json:"mac"`          // MAC地址
	MTU         int       `json:"mtu"`          // 最大传输单元
	Speed       uint64    `json:"speed"`        // 连接速度 (bps)
	Duplex      string    `json:"duplex"`       // 双工模式 (full, half)
	IsUp        bool      `json:"is_up"`        // 是否启用
	IsRunning   bool      `json:"is_running"`   // 是否运行中
	IsLoopback  bool      `json:"is_loopback"`  // 是否回环接口
	IsWireless  bool      `json:"is_wireless"`  // 是否无线接口
	IPv4        []string  `json:"ipv4"`         // IPv4地址列表
	IPv6        []string  `json:"ipv6"`         // IPv6地址列表
	LastUpdated time.Time `json:"last_updated"` // 最后更新时间
}

// NetworkStats 网络接口统计信息
type NetworkStats struct {
	Name            string    `json:"name"`             // 接口名称
	BytesReceived   uint64    `json:"bytes_received"`   // 接收字节数
	BytesSent       uint64    `json:"bytes_sent"`       // 发送字节数
	PacketsReceived uint64    `json:"packets_received"` // 接收包数
	PacketsSent     uint64    `json:"packets_sent"`     // 发送包数
	ErrorsReceived  uint64    `json:"errors_received"`  // 接收错误数
	ErrorsSent      uint64    `json:"errors_sent"`      // 发送错误数
	DropsReceived   uint64    `json:"drops_received"`   // 接收丢包数
	DropsSent       uint64    `json:"drops_sent"`       // 发送丢包数
	LastUpdated     time.Time `json:"last_updated"`     // 最后更新时间
}

// NetworkSpeed 网络速度信息
type NetworkSpeed struct {
	Name          string    `json:"name"`           // 接口名称
	DownloadSpeed uint64    `json:"download_speed"` // 下载速度 (bytes/s)
	UploadSpeed   uint64    `json:"upload_speed"`   // 上传速度 (bytes/s)
	DownloadTotal uint64    `json:"download_total"` // 累计下载量 (bytes)
	UploadTotal   uint64    `json:"upload_total"`   // 累计上传量 (bytes)
	LastUpdated   time.Time `json:"last_updated"`   // 最后更新时间
}

// ConnectionInfo 网络连接信息
type ConnectionInfo struct {
	Protocol    string `json:"protocol"`     // 协议 (tcp, udp)
	LocalAddr   string `json:"local_addr"`   // 本地地址
	LocalPort   uint16 `json:"local_port"`   // 本地端口
	RemoteAddr  string `json:"remote_addr"`  // 远程地址
	RemotePort  uint16 `json:"remote_port"`  // 远程端口
	State       string `json:"state"`        // 连接状态
	ProcessName string `json:"process_name"` // 进程名称
	ProcessID   uint32 `json:"process_id"`   // 进程ID
}

// NetworkSummary 网络概览信息
type NetworkSummary struct {
	TotalInterfaces  int                `json:"total_interfaces"`  // 总接口数
	ActiveInterfaces int                `json:"active_interfaces"` // 活跃接口数
	TotalDownload    uint64             `json:"total_download"`    // 总下载量
	TotalUpload      uint64             `json:"total_upload"`      // 总上传量
	CurrentDownload  uint64             `json:"current_download"`  // 当前下载速度
	CurrentUpload    uint64             `json:"current_upload"`    // 当前上传速度
	PrimaryInterface *NetworkInterface  `json:"primary_interface"` // 主要接口
	Interfaces       []NetworkInterface `json:"interfaces"`        // 所有接口
	LastUpdated      time.Time          `json:"last_updated"`      // 最后更新时间
}

var (
	lastNetworkStats         map[string]*NetworkStats
	lastNetworkStatsTime     time.Time
	speedCalculationInterval = 1 * time.Second
)

// GetInterfaces 获取所有网络接口信息
func GetInterfaces() ([]NetworkInterface, error) {
	var interfaces []NetworkInterface

	// 根据平台获取接口信息
	var err error
	interfaces, err = getPlatformInterfaces()

	if err != nil {
		return nil, err
	}

	// 更新时间戳
	now := time.Now()
	for i := range interfaces {
		interfaces[i].LastUpdated = now
	}

	return interfaces, nil
}

// GetInterfaceStats 获取所有网络接口统计信息
func GetInterfaceStats() ([]NetworkStats, error) {
	var stats []NetworkStats

	// 根据平台获取统计信息
	var err error
	stats, err = getPlatformInterfaceStats()

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

// GetRealTimeSpeed 获取实时网络速度
func GetRealTimeSpeed() ([]NetworkSpeed, error) {
	return GetRealTimeSpeedWithInterval(speedCalculationInterval)
}

// GetRealTimeSpeedWithInterval 获取指定间隔的实时网络速度
func GetRealTimeSpeedWithInterval(interval time.Duration) ([]NetworkSpeed, error) {
	// 获取当前统计
	currentStats, err := GetInterfaceStats()
	if err != nil {
		return nil, err
	}

	// 转换为map以便查找
	currentStatsMap := make(map[string]*NetworkStats)
	for i := range currentStats {
		currentStatsMap[currentStats[i].Name] = &currentStats[i]
	}

	var speeds []NetworkSpeed
	now := time.Now()

	// 如果有上次的统计数据，计算速度
	if lastNetworkStats != nil && !lastNetworkStatsTime.IsZero() {
		timeDiff := now.Sub(lastNetworkStatsTime).Seconds()

		if timeDiff > 0 && timeDiff < 60 { // 防止异常的时间差
			for name, currentStat := range currentStatsMap {
				if lastStat, exists := lastNetworkStats[name]; exists {
					speed := NetworkSpeed{
						Name:          name,
						DownloadTotal: currentStat.BytesReceived,
						UploadTotal:   currentStat.BytesSent,
						LastUpdated:   now,
					}

					// 计算速度 (bytes/s)
					if currentStat.BytesReceived >= lastStat.BytesReceived {
						downloadDiff := currentStat.BytesReceived - lastStat.BytesReceived
						speed.DownloadSpeed = uint64(float64(downloadDiff) / timeDiff)
					}

					if currentStat.BytesSent >= lastStat.BytesSent {
						uploadDiff := currentStat.BytesSent - lastStat.BytesSent
						speed.UploadSpeed = uint64(float64(uploadDiff) / timeDiff)
					}

					speeds = append(speeds, speed)
				}
			}
		}
	} else {
		// 第一次调用，只初始化缓存，返回空结果
		for name, currentStat := range currentStatsMap {
			speeds = append(speeds, NetworkSpeed{
				Name:          name,
				DownloadTotal: currentStat.BytesReceived,
				UploadTotal:   currentStat.BytesSent,
				DownloadSpeed: 0, // 第一次调用速度为0
				UploadSpeed:   0, // 第一次调用速度为0
				LastUpdated:   now,
			})
		}
	}

	// 更新缓存
	lastNetworkStats = currentStatsMap
	lastNetworkStatsTime = now

	// 按接口名称排序
	sort.Slice(speeds, func(i, j int) bool {
		return speeds[i].Name < speeds[j].Name
	})

	return speeds, nil
}

// GetConnections 获取网络连接信息
func GetConnections() ([]ConnectionInfo, error) {
	var connections []ConnectionInfo

	// 根据平台获取连接信息
	var err error
	connections, err = getPlatformConnections()

	return connections, err
}

// GetSummary 获取网络概览信息
func GetSummary() (*NetworkSummary, error) {
	summary := &NetworkSummary{
		LastUpdated: time.Now(),
	}

	// 获取接口信息
	interfaces, err := GetInterfaces()
	if err != nil {
		return nil, err
	}

	summary.Interfaces = interfaces
	summary.TotalInterfaces = len(interfaces)

	// 统计活跃接口和主要接口
	var primaryInterface *NetworkInterface
	for i := range interfaces {
		if interfaces[i].IsUp && interfaces[i].IsRunning {
			summary.ActiveInterfaces++

			// 选择主要接口 (非回环、有IP地址的接口)
			if !interfaces[i].IsLoopback && len(interfaces[i].IPv4) > 0 {
				if primaryInterface == nil ||
					(!interfaces[i].IsWireless && primaryInterface.IsWireless) {
					primaryInterface = &interfaces[i]
				}
			}
		}
	}

	summary.PrimaryInterface = primaryInterface

	// 获取速度信息
	speeds, err := GetRealTimeSpeed()
	if err == nil {
		for _, speed := range speeds {
			summary.TotalDownload += speed.DownloadTotal
			summary.TotalUpload += speed.UploadTotal
			summary.CurrentDownload += speed.DownloadSpeed
			summary.CurrentUpload += speed.UploadSpeed
		}
	}

	return summary, nil
}

// GetActiveInterfaceSpeed 获取活跃接口的网络速度
func GetActiveInterfaceSpeed() (*NetworkSpeed, error) {
	// 获取所有速度信息
	speeds, err := GetRealTimeSpeed()
	if err != nil {
		return nil, err
	}

	// 获取接口信息以确定哪些是活跃的
	interfaces, err := GetInterfaces()
	if err != nil {
		return nil, err
	}

	// 创建活跃接口映射
	activeInterfaces := make(map[string]bool)
	for _, iface := range interfaces {
		if iface.IsUp && iface.IsRunning && !iface.IsLoopback {
			activeInterfaces[iface.Name] = true
		}
	}

	// 合并活跃接口的速度
	totalSpeed := &NetworkSpeed{
		Name:        "total",
		LastUpdated: time.Now(),
	}

	for _, speed := range speeds {
		if activeInterfaces[speed.Name] {
			totalSpeed.DownloadSpeed += speed.DownloadSpeed
			totalSpeed.UploadSpeed += speed.UploadSpeed
			totalSpeed.DownloadTotal += speed.DownloadTotal
			totalSpeed.UploadTotal += speed.UploadTotal
		}
	}

	return totalSpeed, nil
}

// MonitorRealTime 实时监控网络速度 (返回channel)
func MonitorRealTime(interval time.Duration) (<-chan []NetworkSpeed, <-chan error) {
	speedChan := make(chan []NetworkSpeed)
	errorChan := make(chan error)

	go func() {
		defer close(speedChan)
		defer close(errorChan)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			speeds, err := GetRealTimeSpeed()
			if err != nil {
				errorChan <- err
				continue
			}
			speedChan <- speeds
		}
	}()

	return speedChan, errorChan
}

// FormatSpeed 格式化网络速度为可读格式
func FormatSpeed(bytesPerSecond uint64) string {
	const unit = 1000 // 网络速度通常使用1000而不是1024
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

// IsValidInterface 检查接口是否为有效的监控目标
func IsValidInterface(name string) bool {
	// 排除回环和虚拟接口
	excludePrefixes := []string{"lo", "docker", "veth", "br-", "virbr", "tap", "tun"}

	for _, prefix := range excludePrefixes {
		if len(name) >= len(prefix) && name[:len(prefix)] == prefix {
			return false
		}
	}

	return true
}
