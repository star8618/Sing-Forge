//go:build ignore

// IP地理位置查询示例
package main

import (
	"fmt"
	"log"
	"time"

	"native-monitor/ipgeo"
)

func main() {
	fmt.Println("🌍 IP地理位置查询示例")
	fmt.Println("====================")

	// 创建IP地理位置查询服务
	service := ipgeo.NewIPGeoService()

	fmt.Println("\n📍 1. 获取本机外网IP:")
	localIP, err := service.GetLocalIP()
	if err != nil {
		log.Printf("获取本机IP失败: %v", err)
	} else {
		fmt.Printf("  本机外网IP: %s\n", localIP)
	}

	fmt.Println("\n📍 2. 获取本机IP地理位置:")
	localLocation, err := service.GetLocalIPLocation()
	if err != nil {
		log.Printf("获取本机IP地理位置失败: %v", err)
	} else {
		printLocationInfo("本机IP", localLocation)
	}

	fmt.Println("\n📍 3. 获取代理IP地理位置:")
	proxyLocation, err := service.GetProxyIPLocation()
	if err != nil {
		log.Printf("获取代理IP地理位置失败: %v", err)
	} else {
		printLocationInfo("代理IP", proxyLocation)
	}

	fmt.Println("\n📍 4. 同时获取本机和代理IP位置:")
	local, proxy, err := service.GetBothLocations()
	if err != nil {
		log.Printf("批量获取失败: %v", err)
	} else {
		if local != nil {
			fmt.Printf("  本机IP: %s -> %s\n", local.IP, local.Location)
		}
		if proxy != nil {
			fmt.Printf("  代理IP: %s -> %s\n", proxy.IP, proxy.Location)
		}
	}

	fmt.Println("\n📍 5. 测试特定IP地理位置查询:")
	testIPs := []string{
		"113.109.24.55", // 广州市-番禺区
		"8.8.8.8",       // 谷歌DNS
		"1.1.1.1",       // Cloudflare DNS
	}

	for _, ip := range testIPs {
		fmt.Printf("\n  查询IP: %s\n", ip)
		if location, err := service.GetLocationByIP(ip); err != nil {
			fmt.Printf("    查询失败: %v\n", err)
		} else {
			fmt.Printf("    位置: %s\n", location.Location)
			fmt.Printf("    详细: %s %s %s\n", location.Country, location.City, location.District)
			fmt.Printf("    运营商: %s\n", location.ISP)
			fmt.Printf("    中国IP: %t\n", location.IsChinaIP)
		}
	}

	fmt.Println("\n📍 6. 缓存状态:")
	cacheStatus := service.GetCacheStatus()
	for key, value := range cacheStatus {
		fmt.Printf("  %s: %v\n", key, value)
	}

	fmt.Println("\n📍 7. 测试缓存功能:")
	fmt.Println("  第一次获取（从网络）:")
	start := time.Now()
	local1, err := service.GetLocalIPLocation()
	if err != nil {
		log.Printf("获取失败: %v", err)
	} else {
		fmt.Printf("    耗时: %v, 位置: %s\n", time.Since(start), local1.Location)
	}

	fmt.Println("  第二次获取（从缓存）:")
	start = time.Now()
	local2, err := service.GetLocalIPLocation()
	if err != nil {
		log.Printf("获取失败: %v", err)
	} else {
		fmt.Printf("    耗时: %v, 位置: %s\n", time.Since(start), local2.Location)
	}

	fmt.Println("\n📍 8. 刷新缓存:")
	if err := service.RefreshCache(); err != nil {
		log.Printf("刷新缓存失败: %v", err)
	} else {
		fmt.Println("  缓存刷新成功")
	}

	fmt.Println("\n✅ IP地理位置查询示例完成!")
	fmt.Println("\n🔧 功能特点:")
	fmt.Println("  - 🌐 获取本机外网IP (ip.3322.net)")
	fmt.Println("  - 📍 地理位置查询 (api.vore.top)")
	fmt.Println("  - 🇨🇳 中国IP格式化为 城市-区县")
	fmt.Println("  - 🌍 国外IP显示国家名称")
	fmt.Println("  - ⚡ 智能缓存机制")
	fmt.Println("  - 🔄 并发查询支持")
}

func printLocationInfo(label string, location *ipgeo.LocationInfo) {
	fmt.Printf("  %s:\n", label)
	fmt.Printf("    IP地址: %s\n", location.IP)
	fmt.Printf("    位置: %s\n", location.Location)
	fmt.Printf("    国家: %s\n", location.Country)
	if location.IsChinaIP {
		fmt.Printf("    省份: %s\n", location.Province)
		fmt.Printf("    城市: %s\n", location.City)
		fmt.Printf("    区县: %s\n", location.District)
		if location.AdminCode != "" {
			fmt.Printf("    行政代码: %s\n", location.AdminCode)
		}
	}
	fmt.Printf("    运营商: %s\n", location.ISP)
	fmt.Printf("    中国IP: %t\n", location.IsChinaIP)
	fmt.Printf("    更新时间: %s\n", location.LastUpdated.Format("2006-01-02 15:04:05"))
}
