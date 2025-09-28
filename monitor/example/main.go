// Example usage of native-monitor library
package main

import (
	"fmt"
	"strings"
	"time"

	"native-monitor/cpu"
	"native-monitor/disk"
	"native-monitor/ipgeo"
	"native-monitor/memory"
	"native-monitor/network"
	"native-monitor/platform"
	"native-monitor/stats"
)

func main() {
	fmt.Println("🚀 Native Monitor - 原生系统监控库示例")
	fmt.Println(strings.Repeat("=", 50))

	// 1. 平台信息检测
	fmt.Println("\n📋 平台信息:")
	if platformInfo, err := platform.GetPlatformInfo(); err == nil {
		fmt.Printf("  操作系统: %s %s\n", platformInfo.OS, platformInfo.Version)
		fmt.Printf("  架构: %s\n", platformInfo.Architecture)
		fmt.Printf("  内核: %s\n", platformInfo.Kernel)
		fmt.Printf("  主机名: %s\n", platformInfo.Hostname)
		fmt.Printf("  运行时间: %s\n", formatUptime(platformInfo.Uptime))
	}

	// 2. 硬件平台信息
	fmt.Println("\n🔧 硬件平台:")
	if hardware, err := platform.GetHardwarePlatform(); err == nil {
		fmt.Printf("  厂商: %s\n", hardware.Vendor)
		fmt.Printf("  型号: %s\n", hardware.Model)
		fmt.Printf("  序列号: %s\n", hardware.Serial)
		fmt.Printf("  Apple Silicon: %t\n", hardware.IsAppleSilicon)
		fmt.Printf("  虚拟机: %t\n", hardware.IsVirtual)
	}

	// 3. 平台能力
	fmt.Println("\n⚡ 平台能力:")
	caps := platform.GetCapabilities()
	fmt.Printf("  CPU温度监控: %t\n", caps.CPUTemperature)
	fmt.Printf("  内存压力监控: %t\n", caps.MemoryPressure)
	fmt.Printf("  磁盘健康监控: %t\n", caps.DiskHealth)
	fmt.Printf("  GPU信息: %t\n", caps.GPUInfo)

	// 4. CPU监控
	fmt.Println("\n🔥 CPU信息:")
	if cpuInfo, err := cpu.GetInfo(); err == nil {
		fmt.Printf("  型号: %s\n", cpuInfo.Model)
		fmt.Printf("  核心数: %d\n", cpuInfo.Cores)
		if cpuInfo.PerformanceCores > 0 {
			fmt.Printf("  性能核心: %d, 效率核心: %d\n",
				cpuInfo.PerformanceCores, cpuInfo.EfficiencyCores)
		}
		fmt.Printf("  频率: %.2f GHz\n", cpuInfo.Frequency)
		fmt.Printf("  架构: %s\n", cpuInfo.Architecture)
		fmt.Printf("  厂商: %s\n", cpuInfo.Vendor)
	}

	// 5. CPU使用率
	fmt.Println("\n📊 CPU使用率:")
	if cpuUsage, err := cpu.GetUsage(); err == nil {
		fmt.Printf("  总体使用率: %.2f%%\n", cpuUsage.Overall)
		fmt.Printf("  用户态: %.2f%%\n", cpuUsage.User)
		fmt.Printf("  系统态: %.2f%%\n", cpuUsage.System)
		fmt.Printf("  空闲: %.2f%%\n", cpuUsage.Idle)
	}

	// 6. 内存信息
	fmt.Println("\n💾 内存信息:")
	if memInfo, err := memory.GetInfo(); err == nil {
		fmt.Printf("  总内存: %s\n", formatBytes(memInfo.Total))
		fmt.Printf("  已用内存: %s (%.2f%%)\n",
			formatBytes(memInfo.Used), memInfo.UsedPercent)
		fmt.Printf("  可用内存: %s\n", formatBytes(memInfo.Available))
		fmt.Printf("  活跃内存: %s\n", formatBytes(memInfo.Active))
		fmt.Printf("  非活跃内存: %s\n", formatBytes(memInfo.Inactive))
		if memInfo.Wired > 0 {
			fmt.Printf("  联动内存: %s\n", formatBytes(memInfo.Wired))
		}
		if memInfo.Compressed > 0 {
			fmt.Printf("  压缩内存: %s\n", formatBytes(memInfo.Compressed))
		}
	}

	// 7. 交换空间
	fmt.Println("\n🔄 交换空间:")
	if swapInfo, err := memory.GetSwapInfo(); err == nil {
		fmt.Printf("  总交换空间: %s\n", formatBytes(swapInfo.Total))
		fmt.Printf("  已用交换空间: %s (%.2f%%)\n",
			formatBytes(swapInfo.Used), swapInfo.UsedPercent)
		fmt.Printf("  换入次数: %d\n", swapInfo.SwapIn)
		fmt.Printf("  换出次数: %d\n", swapInfo.SwapOut)
	}

	// 8. 磁盘信息
	fmt.Println("\n💿 磁盘信息:")
	if disks, err := disk.GetDisks(); err == nil {
		for _, d := range disks {
			fmt.Printf("  %s (%s):\n", d.Mountpoint, d.FileSystem)
			fmt.Printf("    设备: %s\n", d.Device)
			fmt.Printf("    总容量: %s\n", formatBytes(d.Total))
			fmt.Printf("    已用: %s (%.2f%%)\n",
				formatBytes(d.Used), d.UsedPercent)
			fmt.Printf("    可用: %s\n", formatBytes(d.Available))
			if d.IsReadOnly {
				fmt.Printf("    只读: 是\n")
			}
			fmt.Println()
		}
	}

	// 9. 网络接口
	fmt.Println("\n🌐 网络接口:")
	if interfaces, err := network.GetInterfaces(); err == nil {
		for _, iface := range interfaces {
			if !iface.IsLoopback && iface.IsUp {
				fmt.Printf("  %s (%s):\n", iface.Name, iface.DisplayName)
				fmt.Printf("    硬件类型: %s\n", iface.Hardware)
				fmt.Printf("    MAC地址: %s\n", iface.MAC)
				if len(iface.IPv4) > 0 {
					fmt.Printf("    IPv4: %v\n", iface.IPv4)
				}
				if iface.Speed > 0 {
					fmt.Printf("    连接速度: %s\n", formatSpeed(iface.Speed))
				}
				fmt.Printf("    状态: 运行中=%t, 无线=%t\n",
					iface.IsRunning, iface.IsWireless)
				fmt.Println()
			}
		}
	}

	// 10. 实时网络速度监控
	fmt.Println("\n📡 实时网络速度 (5秒监控):")
	monitorNetworkSpeed()

	// 11. 流量统计示例
	fmt.Println("\n📈 流量统计设置:")
	setupTrafficStats()

	// 12. IP地理位置查询
	fmt.Println("\n🌍 IP地理位置信息:")
	showIPGeoInfo()

	// 13. Apple Silicon特殊信息
	if platform.IsAppleSilicon() {
		fmt.Println("\n🍎 Apple Silicon详细信息:")
		if asInfo, err := cpu.GetAppleSiliconDetails(); err == nil {
			fmt.Printf("  芯片名称: %s\n", asInfo.ChipName)
			fmt.Printf("  性能核心: %d\n", asInfo.PerformanceCores)
			fmt.Printf("  效率核心: %d\n", asInfo.EfficiencyCores)
			fmt.Printf("  GPU核心: %d\n", asInfo.GPUCores)
			fmt.Printf("  Neural核心: %d\n", asInfo.NeuralCores)
			fmt.Printf("  内存带宽: %.1f GB/s\n", asInfo.MemoryBandwidth)
			fmt.Printf("  制程工艺: %s\n", asInfo.ProcessNode)
		}
	}

	fmt.Println("\n✅ 监控库演示完成!")
}

// monitorNetworkSpeed 监控网络速度
func monitorNetworkSpeed() {
	speedChan, errorChan := network.MonitorRealTime(1 * time.Second)

	timeout := time.After(5 * time.Second)

	for {
		select {
		case speeds := <-speedChan:
			for _, speed := range speeds {
				if network.IsValidInterface(speed.Name) &&
					(speed.DownloadSpeed > 0 || speed.UploadSpeed > 0) {
					fmt.Printf("  %s: ⬇️%s ⬆️%s\n",
						speed.Name,
						formatSpeed(speed.DownloadSpeed),
						formatSpeed(speed.UploadSpeed))
				}
			}
		case err := <-errorChan:
			fmt.Printf("  错误: %v\n", err)
		case <-timeout:
			return
		}
	}
}

// setupTrafficStats 设置流量统计
func setupTrafficStats() {
	dataDir := "./traffic_data"
	collector := stats.NewTrafficCollector(dataDir, 5*time.Second, 30)

	fmt.Printf("  数据目录: %s\n", dataDir)
	fmt.Printf("  采集间隔: 5秒\n")
	fmt.Printf("  数据保留: 30天\n")

	if err := collector.Start(); err != nil {
		fmt.Printf("  启动失败: %v\n", err)
		return
	}

	fmt.Printf("  流量统计收集器已启动\n")

	// 演示完成后停止
	time.AfterFunc(10*time.Second, func() {
		collector.Stop()
		fmt.Printf("  流量统计收集器已停止\n")
	})
}

// showIPGeoInfo 显示IP地理位置信息
func showIPGeoInfo() {
	// 获取本机和代理IP位置
	local, proxy, err := ipgeo.QuickGetBothLocations()
	if err != nil {
		fmt.Printf("  获取IP地理位置失败: %v\n", err)
		return
	}

	// 显示本机IP信息
	if local != nil {
		fmt.Printf("  本机IP: %s\n", local.IP)
		fmt.Printf("  本机位置: %s\n", local.Location)
		fmt.Printf("  本机运营商: %s\n", local.ISP)
	}

	// 显示代理IP信息
	if proxy != nil {
		fmt.Printf("  代理IP: %s\n", proxy.IP)
		fmt.Printf("  代理位置: %s\n", proxy.Location)
		fmt.Printf("  代理运营商: %s\n", proxy.ISP)
	}

	// 检查是否使用代理
	if local != nil && proxy != nil {
		if local.IP == proxy.IP {
			fmt.Printf("  网络状态: 直连 (未使用代理)\n")
		} else {
			fmt.Printf("  网络状态: 代理连接\n")
		}
	}

	// 显示位置差异信息
	if diff, err := ipgeo.GetLocationDifference(); err == nil {
		if desc, exists := diff["geo_distance_desc"]; exists {
			fmt.Printf("  位置关系: %v\n", desc)
		}
	}
}

// 辅助格式化函数

func formatBytes(bytes uint64) string {
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

func formatSpeed(bps uint64) string {
	const unit = 1000
	if bps < unit {
		return fmt.Sprintf("%d B/s", bps)
	}
	div, exp := uint64(unit), 0
	for n := bps / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB/s", float64(bps)/float64(div), "KMGTPE"[exp])
}

func formatUptime(seconds uint64) string {
	if seconds < 60 {
		return fmt.Sprintf("%d秒", seconds)
	} else if seconds < 3600 {
		return fmt.Sprintf("%d分钟", seconds/60)
	} else if seconds < 86400 {
		hours := seconds / 3600
		minutes := (seconds % 3600) / 60
		return fmt.Sprintf("%d小时%d分钟", hours, minutes)
	} else {
		days := seconds / 86400
		hours := (seconds % 86400) / 3600
		return fmt.Sprintf("%d天%d小时", days, hours)
	}
}
