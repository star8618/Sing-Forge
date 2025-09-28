// Package stats 提供网络流量统计功能
package stats

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"native-monitor/network"
)

// TrafficRecord 流量记录
type TrafficRecord struct {
	Timestamp  time.Time `json:"timestamp"`   // 时间戳
	Interface  string    `json:"interface"`   // 接口名称
	BytesIn    uint64    `json:"bytes_in"`    // 入站字节数
	BytesOut   uint64    `json:"bytes_out"`   // 出站字节数
	PacketsIn  uint64    `json:"packets_in"`  // 入站包数
	PacketsOut uint64    `json:"packets_out"` // 出站包数
	SpeedIn    uint64    `json:"speed_in"`    // 入站速度 (bytes/s)
	SpeedOut   uint64    `json:"speed_out"`   // 出站速度 (bytes/s)
}

// DailyTrafficStats 每日流量统计
type DailyTrafficStats struct {
	Date          string            `json:"date"`            // 日期 (YYYY-MM-DD)
	TotalBytesIn  uint64            `json:"total_bytes_in"`  // 总入站字节数
	TotalBytesOut uint64            `json:"total_bytes_out"` // 总出站字节数
	PeakSpeedIn   uint64            `json:"peak_speed_in"`   // 峰值入站速度
	PeakSpeedOut  uint64            `json:"peak_speed_out"`  // 峰值出站速度
	AvgSpeedIn    uint64            `json:"avg_speed_in"`    // 平均入站速度
	AvgSpeedOut   uint64            `json:"avg_speed_out"`   // 平均出站速度
	Records       []TrafficRecord   `json:"records"`         // 详细记录
	Summary       map[string]uint64 `json:"summary"`         // 按接口汇总
}

// WeeklyTrafficStats 每周流量统计
type WeeklyTrafficStats struct {
	Week          string              `json:"week"`            // 周 (YYYY-WW)
	StartDate     string              `json:"start_date"`      // 开始日期
	EndDate       string              `json:"end_date"`        // 结束日期
	TotalBytesIn  uint64              `json:"total_bytes_in"`  // 总入站字节数
	TotalBytesOut uint64              `json:"total_bytes_out"` // 总出站字节数
	DailyStats    []DailyTrafficStats `json:"daily_stats"`     // 每日统计
	Summary       map[string]uint64   `json:"summary"`         // 按接口汇总
}

// MonthlyTrafficStats 每月流量统计
type MonthlyTrafficStats struct {
	Month         string               `json:"month"`           // 月份 (YYYY-MM)
	TotalBytesIn  uint64               `json:"total_bytes_in"`  // 总入站字节数
	TotalBytesOut uint64               `json:"total_bytes_out"` // 总出站字节数
	WeeklyStats   []WeeklyTrafficStats `json:"weekly_stats"`    // 每周统计
	DailyStats    []DailyTrafficStats  `json:"daily_stats"`     // 每日统计
	Summary       map[string]uint64    `json:"summary"`         // 按接口汇总
}

// TrafficCollector 流量收集器
type TrafficCollector struct {
	dataDir         string
	collectInterval time.Duration
	retentionDays   int
	isCollecting    bool
	stopChan        chan struct{}
}

// NewTrafficCollector 创建流量收集器
func NewTrafficCollector(dataDir string, collectInterval time.Duration, retentionDays int) *TrafficCollector {
	return &TrafficCollector{
		dataDir:         dataDir,
		collectInterval: collectInterval,
		retentionDays:   retentionDays,
		stopChan:        make(chan struct{}),
	}
}

// Start 开始收集流量数据
func (tc *TrafficCollector) Start() error {
	if tc.isCollecting {
		return fmt.Errorf("traffic collector is already running")
	}

	// 确保数据目录存在
	if err := os.MkdirAll(tc.dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %v", err)
	}

	tc.isCollecting = true

	go tc.collectLoop()
	go tc.cleanupLoop()

	return nil
}

// Stop 停止收集流量数据
func (tc *TrafficCollector) Stop() {
	if !tc.isCollecting {
		return
	}

	close(tc.stopChan)
	tc.isCollecting = false
}

// collectLoop 收集循环
func (tc *TrafficCollector) collectLoop() {
	ticker := time.NewTicker(tc.collectInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := tc.collectOnce(); err != nil {
				fmt.Printf("Error collecting traffic data: %v\n", err)
			}
		case <-tc.stopChan:
			return
		}
	}
}

// cleanupLoop 清理循环
func (tc *TrafficCollector) cleanupLoop() {
	ticker := time.NewTicker(24 * time.Hour) // 每天清理一次
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := tc.cleanup(); err != nil {
				fmt.Printf("Error cleaning up old data: %v\n", err)
			}
		case <-tc.stopChan:
			return
		}
	}
}

// collectOnce 执行一次流量收集
func (tc *TrafficCollector) collectOnce() error {
	// 获取网络速度信息
	speeds, err := network.GetRealTimeSpeed()
	if err != nil {
		return err
	}

	// 获取网络统计信息
	stats, err := network.GetInterfaceStats()
	if err != nil {
		return err
	}

	// 创建统计映射
	statsMap := make(map[string]*network.NetworkStats)
	for i := range stats {
		statsMap[stats[i].Name] = &stats[i]
	}

	now := time.Now()
	var records []TrafficRecord

	// 为每个接口创建记录
	for _, speed := range speeds {
		if !network.IsValidInterface(speed.Name) {
			continue
		}

		record := TrafficRecord{
			Timestamp: now,
			Interface: speed.Name,
			SpeedIn:   speed.DownloadSpeed,
			SpeedOut:  speed.UploadSpeed,
		}

		// 添加累计统计
		if stat, exists := statsMap[speed.Name]; exists {
			record.BytesIn = stat.BytesReceived
			record.BytesOut = stat.BytesSent
			record.PacketsIn = stat.PacketsReceived
			record.PacketsOut = stat.PacketsSent
		}

		records = append(records, record)
	}

	// 保存记录
	return tc.saveRecords(now, records)
}

// saveRecords 保存流量记录
func (tc *TrafficCollector) saveRecords(timestamp time.Time, records []TrafficRecord) error {
	dateStr := timestamp.Format("2006-01-02")
	filename := filepath.Join(tc.dataDir, fmt.Sprintf("traffic_%s.json", dateStr))

	// 读取现有数据
	var dailyStats DailyTrafficStats
	if data, err := os.ReadFile(filename); err == nil {
		json.Unmarshal(data, &dailyStats)
	}

	// 初始化如果为空
	if dailyStats.Date == "" {
		dailyStats.Date = dateStr
		dailyStats.Summary = make(map[string]uint64)
		dailyStats.Records = []TrafficRecord{}
	}

	// 添加新记录
	dailyStats.Records = append(dailyStats.Records, records...)

	// 更新统计信息
	tc.updateDailyStats(&dailyStats)

	// 保存数据
	data, err := json.MarshalIndent(dailyStats, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

// updateDailyStats 更新每日统计信息
func (tc *TrafficCollector) updateDailyStats(stats *DailyTrafficStats) {
	interfaceTraffic := make(map[string]struct {
		bytesIn, bytesOut       uint64
		maxSpeedIn, maxSpeedOut uint64
		speedSum                int
		speedCount              int
	})

	// 计算每个接口的统计
	for _, record := range stats.Records {
		iface := interfaceTraffic[record.Interface]

		// 更新最大值
		if record.BytesIn > iface.bytesIn {
			iface.bytesIn = record.BytesIn
		}
		if record.BytesOut > iface.bytesOut {
			iface.bytesOut = record.BytesOut
		}
		if record.SpeedIn > iface.maxSpeedIn {
			iface.maxSpeedIn = record.SpeedIn
		}
		if record.SpeedOut > iface.maxSpeedOut {
			iface.maxSpeedOut = record.SpeedOut
		}

		interfaceTraffic[record.Interface] = iface
	}

	// 计算总值
	stats.TotalBytesIn = 0
	stats.TotalBytesOut = 0
	stats.PeakSpeedIn = 0
	stats.PeakSpeedOut = 0

	for ifaceName, iface := range interfaceTraffic {
		stats.TotalBytesIn += iface.bytesIn
		stats.TotalBytesOut += iface.bytesOut

		if iface.maxSpeedIn > stats.PeakSpeedIn {
			stats.PeakSpeedIn = iface.maxSpeedIn
		}
		if iface.maxSpeedOut > stats.PeakSpeedOut {
			stats.PeakSpeedOut = iface.maxSpeedOut
		}

		// 更新接口汇总
		stats.Summary[ifaceName+"_in"] = iface.bytesIn
		stats.Summary[ifaceName+"_out"] = iface.bytesOut
	}
}

// cleanup 清理过期数据
func (tc *TrafficCollector) cleanup() error {
	cutoff := time.Now().AddDate(0, 0, -tc.retentionDays)

	entries, err := os.ReadDir(tc.dataDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "traffic_") {
			// 提取日期
			dateStr := strings.TrimPrefix(entry.Name(), "traffic_")
			dateStr = strings.TrimSuffix(dateStr, ".json")

			if date, err := time.Parse("2006-01-02", dateStr); err == nil {
				if date.Before(cutoff) {
					filename := filepath.Join(tc.dataDir, entry.Name())
					if err := os.Remove(filename); err != nil {
						fmt.Printf("Failed to remove old file %s: %v\n", filename, err)
					}
				}
			}
		}
	}

	return nil
}

// GetDailyStats 获取指定日期的流量统计
func (tc *TrafficCollector) GetDailyStats(date time.Time) (*DailyTrafficStats, error) {
	dateStr := date.Format("2006-01-02")
	filename := filepath.Join(tc.dataDir, fmt.Sprintf("traffic_%s.json", dateStr))

	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var stats DailyTrafficStats
	if err := json.Unmarshal(data, &stats); err != nil {
		return nil, err
	}

	return &stats, nil
}

// GetWeeklyStats 获取指定周的流量统计
func (tc *TrafficCollector) GetWeeklyStats(year, week int) (*WeeklyTrafficStats, error) {
	// 计算该周的开始和结束日期
	startDate := getWeekStartDate(year, week)
	endDate := startDate.AddDate(0, 0, 6)

	weekStats := &WeeklyTrafficStats{
		Week:      fmt.Sprintf("%d-W%02d", year, week),
		StartDate: startDate.Format("2006-01-02"),
		EndDate:   endDate.Format("2006-01-02"),
		Summary:   make(map[string]uint64),
	}

	// 收集该周每天的统计
	for d := 0; d < 7; d++ {
		date := startDate.AddDate(0, 0, d)
		if dailyStats, err := tc.GetDailyStats(date); err == nil {
			weekStats.DailyStats = append(weekStats.DailyStats, *dailyStats)
			weekStats.TotalBytesIn += dailyStats.TotalBytesIn
			weekStats.TotalBytesOut += dailyStats.TotalBytesOut

			// 合并接口汇总
			for key, value := range dailyStats.Summary {
				weekStats.Summary[key] += value
			}
		}
	}

	return weekStats, nil
}

// GetMonthlyStats 获取指定月份的流量统计
func (tc *TrafficCollector) GetMonthlyStats(year, month int) (*MonthlyTrafficStats, error) {
	monthStats := &MonthlyTrafficStats{
		Month:   fmt.Sprintf("%d-%02d", year, month),
		Summary: make(map[string]uint64),
	}

	// 获取该月的所有日期
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1)

	// 收集每天的统计
	for date := startDate; !date.After(endDate); date = date.AddDate(0, 0, 1) {
		if dailyStats, err := tc.GetDailyStats(date); err == nil {
			monthStats.DailyStats = append(monthStats.DailyStats, *dailyStats)
			monthStats.TotalBytesIn += dailyStats.TotalBytesIn
			monthStats.TotalBytesOut += dailyStats.TotalBytesOut

			// 合并接口汇总
			for key, value := range dailyStats.Summary {
				monthStats.Summary[key] += value
			}
		}
	}

	// 计算周统计
	monthStats.WeeklyStats = tc.calculateWeeklyStatsForMonth(year, month)

	return monthStats, nil
}

// calculateWeeklyStatsForMonth 计算月份内的周统计
func (tc *TrafficCollector) calculateWeeklyStatsForMonth(year, month int) []WeeklyTrafficStats {
	var weeklyStats []WeeklyTrafficStats

	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1)

	// 找到包含该月的所有周
	_, startWeek := startDate.ISOWeek()
	_, endWeek := endDate.ISOWeek()

	for week := startWeek; week <= endWeek; week++ {
		if weekStats, err := tc.GetWeeklyStats(year, week); err == nil {
			weeklyStats = append(weeklyStats, *weekStats)
		}
	}

	return weeklyStats
}

// GetRecentStats 获取最近N天的流量统计
func (tc *TrafficCollector) GetRecentStats(days int) ([]DailyTrafficStats, error) {
	var allStats []DailyTrafficStats

	for i := 0; i < days; i++ {
		date := time.Now().AddDate(0, 0, -i)
		if dailyStats, err := tc.GetDailyStats(date); err == nil {
			allStats = append(allStats, *dailyStats)
		}
	}

	// 按日期排序
	sort.Slice(allStats, func(i, j int) bool {
		return allStats[i].Date < allStats[j].Date
	})

	return allStats, nil
}

// 辅助函数

// getWeekStartDate 获取指定年份和周数的开始日期
func getWeekStartDate(year, week int) time.Time {
	// ISO 8601周日期计算
	jan1 := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)

	// 找到第一个周一
	daysFromMonday := int(jan1.Weekday()) - 1
	if daysFromMonday < 0 {
		daysFromMonday = 6
	}

	firstMonday := jan1.AddDate(0, 0, -daysFromMonday)

	// 计算指定周的开始日期
	return firstMonday.AddDate(0, 0, (week-1)*7)
}

// FormatTrafficSize 格式化流量大小
func FormatTrafficSize(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatSpeed 格式化速度
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
