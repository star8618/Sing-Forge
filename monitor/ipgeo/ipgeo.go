// Package ipgeo 提供IP地理位置查询功能
package ipgeo

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// IPInfo IP基本信息
type IPInfo struct {
	Type string `json:"type"` // ipv4/ipv6
	Text string `json:"text"` // IP地址
	CNIP bool   `json:"cnip"` // 是否中国IP
}

// IPData IP地理数据
type IPData struct {
	Info1 string `json:"info1"` // 省份/国家
	Info2 string `json:"info2"` // 城市
	Info3 string `json:"info3"` // 区县
	ISP   string `json:"isp"`   // 运营商
}

// AdCode 行政区划代码
type AdCode struct {
	O string `json:"o"` // 完整描述
	P string `json:"p"` // 省份
	C string `json:"c"` // 城市
	N string `json:"n"` // 简称
	R string `json:"r"` // 区域
	A string `json:"a"` // 行政代码
	I bool   `json:"i"` // 是否中国
}

// VoreAPIResponse VORE API响应结构
type VoreAPIResponse struct {
	Code   int    `json:"code"`
	Msg    string `json:"msg"`
	IPInfo IPInfo `json:"ipinfo"`
	IPData IPData `json:"ipdata"`
	AdCode AdCode `json:"adcode"`
	Tips   string `json:"tips"`
	Time   int64  `json:"time"`
}

// LocationInfo 地理位置信息
type LocationInfo struct {
	IP          string    `json:"ip"`           // IP地址
	Country     string    `json:"country"`      // 国家
	Province    string    `json:"province"`     // 省份
	City        string    `json:"city"`         // 城市
	District    string    `json:"district"`     // 区县
	ISP         string    `json:"isp"`          // 运营商
	Location    string    `json:"location"`     // 格式化位置 (如: 广州市-番禺区)
	IsChinaIP   bool      `json:"is_china_ip"`  // 是否中国IP
	AdminCode   string    `json:"admin_code"`   // 行政区划代码
	LastUpdated time.Time `json:"last_updated"` // 最后更新时间
}

// IPGeoService IP地理位置查询服务
type IPGeoService struct {
	// 缓存相关
	localIPCache    *LocationInfo
	proxyIPCache    *LocationInfo
	lastLocalUpdate time.Time
	lastProxyUpdate time.Time
	cacheExpireTime time.Duration

	// API配置
	localIPURL  string
	voreAPIURL  string
	httpTimeout time.Duration

	// HTTP客户端
	httpClient *http.Client
}

// NewIPGeoService 创建IP地理位置查询服务
func NewIPGeoService() *IPGeoService {
	return &IPGeoService{
		localIPURL:      "https://ip.3322.net",
		voreAPIURL:      "https://api.vore.top/api/IPdata",
		httpTimeout:     10 * time.Second,
		cacheExpireTime: 5 * time.Minute, // 缓存5分钟
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetLocalIP 获取本机外网IP地址
func (s *IPGeoService) GetLocalIP() (string, error) {
	resp, err := s.httpClient.Get(s.localIPURL)
	if err != nil {
		return "", fmt.Errorf("获取本机IP失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("获取本机IP失败: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %v", err)
	}

	ip := strings.TrimSpace(string(body))
	if ip == "" {
		return "", fmt.Errorf("获取到空的IP地址")
	}

	return ip, nil
}

// GetLocationByIP 根据IP地址获取地理位置信息
func (s *IPGeoService) GetLocationByIP(ip string) (*LocationInfo, error) {
	url := fmt.Sprintf("%s?ip=%s", s.voreAPIURL, ip)

	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("查询IP地理位置失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("查询IP地理位置失败: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var voreResp VoreAPIResponse
	if err := json.Unmarshal(body, &voreResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if voreResp.Code != 200 {
		return nil, fmt.Errorf("API返回错误: %s", voreResp.Msg)
	}

	// 转换为LocationInfo结构
	location := &LocationInfo{
		IP:          voreResp.IPInfo.Text,
		Country:     voreResp.IPData.Info1,
		Province:    voreResp.IPData.Info1,
		City:        voreResp.IPData.Info2,
		District:    voreResp.IPData.Info3,
		ISP:         voreResp.IPData.ISP,
		IsChinaIP:   voreResp.IPInfo.CNIP,
		AdminCode:   voreResp.AdCode.A,
		LastUpdated: time.Now(),
	}

	// 格式化位置信息
	location.Location = s.formatLocation(location)

	return location, nil
}

// GetLocalIPLocation 获取本机IP的地理位置信息（带缓存）
func (s *IPGeoService) GetLocalIPLocation() (*LocationInfo, error) {
	// 检查缓存
	if s.localIPCache != nil && time.Since(s.lastLocalUpdate) < s.cacheExpireTime {
		return s.localIPCache, nil
	}

	// 获取本机IP
	ip, err := s.GetLocalIP()
	if err != nil {
		return nil, err
	}

	// 获取地理位置
	location, err := s.GetLocationByIP(ip)
	if err != nil {
		return nil, err
	}

	// 更新缓存
	s.localIPCache = location
	s.lastLocalUpdate = time.Now()

	return location, nil
}

// GetProxyIPLocation 获取代理IP的地理位置信息（使用VORE API的默认IP）
func (s *IPGeoService) GetProxyIPLocation() (*LocationInfo, error) {
	// 检查缓存
	if s.proxyIPCache != nil && time.Since(s.lastProxyUpdate) < s.cacheExpireTime {
		return s.proxyIPCache, nil
	}

	// 不传IP参数，让VORE API返回代理服务器看到的IP
	url := s.voreAPIURL

	resp, err := s.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("查询代理IP地理位置失败: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("查询代理IP地理位置失败: HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var voreResp VoreAPIResponse
	if err := json.Unmarshal(body, &voreResp); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	if voreResp.Code != 200 {
		return nil, fmt.Errorf("API返回错误: %s", voreResp.Msg)
	}

	// 转换为LocationInfo结构
	location := &LocationInfo{
		IP:          voreResp.IPInfo.Text,
		Country:     voreResp.IPData.Info1,
		Province:    voreResp.IPData.Info1,
		City:        voreResp.IPData.Info2,
		District:    voreResp.IPData.Info3,
		ISP:         voreResp.IPData.ISP,
		IsChinaIP:   voreResp.IPInfo.CNIP,
		AdminCode:   voreResp.AdCode.A,
		LastUpdated: time.Now(),
	}

	// 格式化位置信息
	location.Location = s.formatLocation(location)

	// 更新缓存
	s.proxyIPCache = location
	s.lastProxyUpdate = time.Now()

	return location, nil
}

// GetBothLocations 同时获取本机IP和代理IP的地理位置信息
func (s *IPGeoService) GetBothLocations() (local *LocationInfo, proxy *LocationInfo, err error) {
	// 并发获取两个位置信息
	localChan := make(chan *LocationInfo, 1)
	proxyChan := make(chan *LocationInfo, 1)
	localErrChan := make(chan error, 1)
	proxyErrChan := make(chan error, 1)

	// 获取本机IP位置
	go func() {
		loc, err := s.GetLocalIPLocation()
		if err != nil {
			localErrChan <- err
		} else {
			localChan <- loc
		}
	}()

	// 获取代理IP位置
	go func() {
		loc, err := s.GetProxyIPLocation()
		if err != nil {
			proxyErrChan <- err
		} else {
			proxyChan <- loc
		}
	}()

	// 等待结果
	var localErr, proxyErr error
	timeout := time.After(30 * time.Second)

	for i := 0; i < 2; i++ {
		select {
		case local = <-localChan:
			// 本机IP获取成功
		case proxy = <-proxyChan:
			// 代理IP获取成功
		case localErr = <-localErrChan:
			// 本机IP获取失败
		case proxyErr = <-proxyErrChan:
			// 代理IP获取失败
		case <-timeout:
			return nil, nil, fmt.Errorf("获取IP地理位置超时")
		}
	}

	// 如果两个都失败，返回错误
	if localErr != nil && proxyErr != nil {
		return nil, nil, fmt.Errorf("获取IP地理位置失败: 本机IP(%v), 代理IP(%v)", localErr, proxyErr)
	}

	return local, proxy, nil
}

// formatLocation 格式化位置信息为"城市-区县"格式
func (s *IPGeoService) formatLocation(location *LocationInfo) string {
	if !location.IsChinaIP {
		// 非中国IP，只显示国家
		return location.Country
	}

	// 中国IP，格式化为"城市-区县"
	var parts []string

	if location.City != "" {
		city := location.City
		// 确保城市名以"市"结尾
		if !strings.HasSuffix(city, "市") && !strings.HasSuffix(city, "区") &&
			!strings.HasSuffix(city, "县") && !strings.HasSuffix(city, "盟") {
			city += "市"
		}
		parts = append(parts, city)
	}

	if location.District != "" {
		district := location.District
		// 确保区县名以适当后缀结尾
		if !strings.HasSuffix(district, "区") && !strings.HasSuffix(district, "县") &&
			!strings.HasSuffix(district, "市") && !strings.HasSuffix(district, "旗") {
			district += "区"
		}
		parts = append(parts, district)
	}

	if len(parts) == 0 {
		return location.Province
	}

	return strings.Join(parts, "-")
}

// RefreshCache 刷新缓存
func (s *IPGeoService) RefreshCache() error {
	s.localIPCache = nil
	s.proxyIPCache = nil
	s.lastLocalUpdate = time.Time{}
	s.lastProxyUpdate = time.Time{}

	// 重新获取数据
	_, _, err := s.GetBothLocations()
	return err
}

// SetCacheExpireTime 设置缓存过期时间
func (s *IPGeoService) SetCacheExpireTime(duration time.Duration) {
	s.cacheExpireTime = duration
}

// SetHTTPTimeout 设置HTTP请求超时时间
func (s *IPGeoService) SetHTTPTimeout(timeout time.Duration) {
	s.httpTimeout = timeout
	s.httpClient.Timeout = timeout
}

// GetCacheStatus 获取缓存状态
func (s *IPGeoService) GetCacheStatus() map[string]interface{} {
	status := make(map[string]interface{})

	status["local_ip_cached"] = s.localIPCache != nil
	status["proxy_ip_cached"] = s.proxyIPCache != nil
	status["cache_expire_time"] = s.cacheExpireTime.String()

	if s.localIPCache != nil {
		status["local_cache_age"] = time.Since(s.lastLocalUpdate).String()
		status["local_ip"] = s.localIPCache.IP
		status["local_location"] = s.localIPCache.Location
	}

	if s.proxyIPCache != nil {
		status["proxy_cache_age"] = time.Since(s.lastProxyUpdate).String()
		status["proxy_ip"] = s.proxyIPCache.IP
		status["proxy_location"] = s.proxyIPCache.Location
	}

	return status
}

// ValidateIP 验证IP地址格式
func ValidateIP(ip string) bool {
	// 简单的IP格式验证
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return false
	}

	for _, part := range parts {
		if len(part) == 0 || len(part) > 3 {
			return false
		}

		for _, char := range part {
			if char < '0' || char > '9' {
				return false
			}
		}
	}

	return true
}

// FormatLocationSimple 简化位置格式（仅用于显示）
func FormatLocationSimple(location *LocationInfo) string {
	if location == nil {
		return "未知位置"
	}

	if !location.IsChinaIP {
		return location.Country
	}

	return location.Location
}
