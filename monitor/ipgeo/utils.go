// Package ipgeo 工具函数
package ipgeo

import (
	"fmt"
	"sync"
	"time"
)

var (
	// 全局服务实例
	globalService *IPGeoService
	globalOnce    sync.Once
)

// GetGlobalService 获取全局IP地理位置查询服务实例
func GetGlobalService() *IPGeoService {
	globalOnce.Do(func() {
		globalService = NewIPGeoService()
	})
	return globalService
}

// QuickGetLocalLocation 快速获取本机IP地理位置（使用全局服务）
func QuickGetLocalLocation() (*LocationInfo, error) {
	return GetGlobalService().GetLocalIPLocation()
}

// QuickGetProxyLocation 快速获取代理IP地理位置（使用全局服务）
func QuickGetProxyLocation() (*LocationInfo, error) {
	return GetGlobalService().GetProxyIPLocation()
}

// QuickGetBothLocations 快速获取本机和代理IP地理位置（使用全局服务）
func QuickGetBothLocations() (*LocationInfo, *LocationInfo, error) {
	return GetGlobalService().GetBothLocations()
}

// QuickGetLocationByIP 快速根据IP获取地理位置（使用全局服务）
func QuickGetLocationByIP(ip string) (*LocationInfo, error) {
	return GetGlobalService().GetLocationByIP(ip)
}

// GetLocationSummary 获取位置信息摘要
func GetLocationSummary() (map[string]interface{}, error) {
	local, proxy, err := QuickGetBothLocations()
	if err != nil {
		return nil, err
	}

	summary := make(map[string]interface{})

	if local != nil {
		summary["local"] = map[string]interface{}{
			"ip":       local.IP,
			"location": local.Location,
			"isp":      local.ISP,
			"is_china": local.IsChinaIP,
		}
	}

	if proxy != nil {
		summary["proxy"] = map[string]interface{}{
			"ip":       proxy.IP,
			"location": proxy.Location,
			"isp":      proxy.ISP,
			"is_china": proxy.IsChinaIP,
		}
	}

	summary["timestamp"] = time.Now().Format("2006-01-02 15:04:05")

	return summary, nil
}

// CompareIPs 比较本机IP和代理IP是否相同
func CompareIPs() (bool, string, string, error) {
	local, proxy, err := QuickGetBothLocations()
	if err != nil {
		return false, "", "", err
	}

	var localIP, proxyIP string
	if local != nil {
		localIP = local.IP
	}
	if proxy != nil {
		proxyIP = proxy.IP
	}

	return localIP == proxyIP, localIP, proxyIP, nil
}

// FormatLocationForDisplay 格式化位置信息用于显示
func FormatLocationForDisplay(location *LocationInfo) string {
	if location == nil {
		return "未知位置"
	}

	if !location.IsChinaIP {
		// 国外IP只显示国家
		return fmt.Sprintf("%s (%s)", location.Country, location.ISP)
	}

	// 中国IP显示详细位置
	return fmt.Sprintf("%s (%s)", location.Location, location.ISP)
}

// GetLocationDifference 获取本机和代理IP的位置差异
func GetLocationDifference() (map[string]interface{}, error) {
	local, proxy, err := QuickGetBothLocations()
	if err != nil {
		return nil, err
	}

	diff := make(map[string]interface{})

	if local != nil && proxy != nil {
		diff["same_ip"] = local.IP == proxy.IP
		diff["same_country"] = local.Country == proxy.Country
		diff["same_city"] = local.City == proxy.City
		diff["same_isp"] = local.ISP == proxy.ISP
		diff["both_china"] = local.IsChinaIP && proxy.IsChinaIP

		diff["local_location"] = local.Location
		diff["proxy_location"] = proxy.Location

		if local.IsChinaIP && proxy.IsChinaIP {
			diff["same_province"] = local.Province == proxy.Province
			diff["same_district"] = local.District == proxy.District
		}

		// 计算"距离"（简单的地理位置层级差异）
		distance := 0
		if local.Country != proxy.Country {
			distance = 4 // 不同国家
		} else if local.Province != proxy.Province {
			distance = 3 // 不同省份
		} else if local.City != proxy.City {
			distance = 2 // 不同城市
		} else if local.District != proxy.District {
			distance = 1 // 不同区县
		}
		diff["geo_distance_level"] = distance

		distanceDesc := []string{"相同位置", "不同区县", "不同城市", "不同省份", "不同国家"}
		if distance < len(distanceDesc) {
			diff["geo_distance_desc"] = distanceDesc[distance]
		}
	}

	return diff, nil
}

// MonitorLocationChanges 监控位置变化（定期检查）
func MonitorLocationChanges(interval time.Duration, callback func(local, proxy *LocationInfo)) chan struct{} {
	stopChan := make(chan struct{})

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		var lastLocal, lastProxy *LocationInfo

		for {
			select {
			case <-ticker.C:
				// 刷新缓存获取最新位置
				GetGlobalService().RefreshCache()
				local, proxy, err := QuickGetBothLocations()
				if err != nil {
					continue
				}

				// 检查是否有变化
				localChanged := (lastLocal == nil && local != nil) ||
					(lastLocal != nil && local != nil && lastLocal.IP != local.IP)
				proxyChanged := (lastProxy == nil && proxy != nil) ||
					(lastProxy != nil && proxy != nil && lastProxy.IP != proxy.IP)

				if localChanged || proxyChanged {
					callback(local, proxy)
					lastLocal = local
					lastProxy = proxy
				}

			case <-stopChan:
				return
			}
		}
	}()

	return stopChan
}

// BatchQueryIPs 批量查询多个IP的地理位置
func BatchQueryIPs(ips []string) (map[string]*LocationInfo, error) {
	service := GetGlobalService()
	results := make(map[string]*LocationInfo)

	// 使用通道进行并发查询
	type result struct {
		ip       string
		location *LocationInfo
		err      error
	}

	resultChan := make(chan result, len(ips))

	// 启动并发查询
	for _, ip := range ips {
		go func(queryIP string) {
			location, err := service.GetLocationByIP(queryIP)
			resultChan <- result{
				ip:       queryIP,
				location: location,
				err:      err,
			}
		}(ip)
	}

	// 收集结果
	for i := 0; i < len(ips); i++ {
		res := <-resultChan
		if res.err == nil {
			results[res.ip] = res.location
		} else {
			// 查询失败的IP，记录错误但继续处理其他IP
			fmt.Printf("查询IP %s 失败: %v\n", res.ip, res.err)
		}
	}

	return results, nil
}

// GetCurrentNetworkInfo 获取当前网络信息摘要
func GetCurrentNetworkInfo() map[string]interface{} {
	info := make(map[string]interface{})

	// 获取IP信息
	if local, proxy, err := QuickGetBothLocations(); err == nil {
		if local != nil {
			info["local_ip"] = local.IP
			info["local_location"] = local.Location
			info["local_isp"] = local.ISP
		}
		if proxy != nil {
			info["proxy_ip"] = proxy.IP
			info["proxy_location"] = proxy.Location
			info["proxy_isp"] = proxy.ISP
		}

		// 检查是否使用代理
		if local != nil && proxy != nil {
			info["using_proxy"] = local.IP != proxy.IP
			if local.IP != proxy.IP {
				info["proxy_type"] = "外部代理"
			} else {
				info["proxy_type"] = "直连"
			}
		}
	}

	// 获取缓存状态
	cacheStatus := GetGlobalService().GetCacheStatus()
	info["cache_status"] = cacheStatus

	info["query_time"] = time.Now().Format("2006-01-02 15:04:05")

	return info
}
