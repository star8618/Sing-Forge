# Native Monitor API 文档 📚

本文档详细描述了Native Monitor库中所有模块的API接口和使用方法。

## 📋 目录

- [CPU模块](#cpu模块)
- [内存模块](#内存模块)
- [GPU模块](#gpu模块)
- [磁盘模块](#磁盘模块)
- [网络模块](#网络模块)
- [流量统计模块](#流量统计模块)
- [平台检测模块](#平台检测模块)
- [IP地理位置模块](#ip地理位置模块)

---

## 🖥️ CPU模块

### 导入
```go
import "native-monitor/cpu"
```

### 数据结构

#### CPUInfo
```go
type CPUInfo struct {
    ModelName        string    `json:"model_name"`        // CPU型号名称
    Architecture     string    `json:"architecture"`     // 架构 (arm64, x86_64)
    Cores           int       `json:"cores"`            // 物理核心数
    Threads         int       `json:"threads"`          // 逻辑线程数
    PerformanceCores int       `json:"performance_cores"` // 性能核心数 (Apple Silicon)
    EfficiencyCores  int       `json:"efficiency_cores"`  // 效率核心数 (Apple Silicon)
    BaseFrequency   float64   `json:"base_frequency"`   // 基础频率 (MHz)
    MaxFrequency    float64   `json:"max_frequency"`    // 最大频率 (MHz)
    CacheL1         uint64    `json:"cache_l1"`         // L1缓存大小
    CacheL2         uint64    `json:"cache_l2"`         // L2缓存大小
    CacheL3         uint64    `json:"cache_l3"`         // L3缓存大小
    Vendor          string    `json:"vendor"`           // 厂商
    Family          string    `json:"family"`           // 家族
    Model           string    `json:"model"`            // 型号
    Stepping        string    `json:"stepping"`         // 步进
    LastUpdated     time.Time `json:"last_updated"`     // 最后更新时间
}
```

#### CPUUsage
```go
type CPUUsage struct {
    Overall    float64   `json:"overall"`     // 总体使用率
    PerCore    []float64 `json:"per_core"`    // 每核心使用率
    User       float64   `json:"user"`        // 用户态使用率
    System     float64   `json:"system"`      // 系统态使用率
    Idle       float64   `json:"idle"`        // 空闲率
    IOWait     float64   `json:"iowait"`      // IO等待率
    LoadAvg1   float64   `json:"load_avg_1"`  // 1分钟负载
    LoadAvg5   float64   `json:"load_avg_5"`  // 5分钟负载
    LoadAvg15  float64   `json:"load_avg_15"` // 15分钟负载
    LastUpdated time.Time `json:"last_updated"` // 最后更新时间
}
```

### 主要函数

#### GetInfo() (*CPUInfo, error)
获取CPU基本信息。

**返回值:**
- `*CPUInfo`: CPU信息结构体
- `error`: 错误信息

**示例:**
```go
info, err := cpu.GetInfo()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("CPU: %s (%d核)\n", info.ModelName, info.Cores)
```

#### GetUsage() (*CPUUsage, error)
获取CPU实时使用率。

**返回值:**
- `*CPUUsage`: CPU使用率结构体
- `error`: 错误信息

**示例:**
```go
usage, err := cpu.GetUsage()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("CPU使用率: %.2f%%\n", usage.Overall)
```

#### GetTemperature() (float64, error)
获取CPU温度（如果支持）。

**返回值:**
- `float64`: 温度值（摄氏度）
- `error`: 错误信息

#### GetFrequency() (float64, error)
获取当前CPU频率。

**返回值:**
- `float64`: 频率值（MHz）
- `error`: 错误信息

#### IsAppleSilicon() bool
检查是否为Apple Silicon。

**返回值:**
- `bool`: 是否为Apple Silicon

---

## 💾 内存模块

### 导入
```go
import "native-monitor/memory"
```

### 数据结构

#### MemoryInfo
```go
type MemoryInfo struct {
    Total       uint64    `json:"total"`        // 总内存
    Available   uint64    `json:"available"`    // 可用内存
    Used        uint64    `json:"used"`         // 已用内存
    Free        uint64    `json:"free"`         // 空闲内存
    Cached      uint64    `json:"cached"`       // 缓存内存
    Buffers     uint64    `json:"buffers"`      // 缓冲区内存
    Shared      uint64    `json:"shared"`       // 共享内存
    UsedPercent float64   `json:"used_percent"` // 使用率百分比
    LastUpdated time.Time `json:"last_updated"` // 最后更新时间
}
```

#### SwapInfo
```go
type SwapInfo struct {
    Total       uint64    `json:"total"`        // 总交换空间
    Used        uint64    `json:"used"`         // 已用交换空间
    Free        uint64    `json:"free"`         // 空闲交换空间
    UsedPercent float64   `json:"used_percent"` // 使用率百分比
    LastUpdated time.Time `json:"last_updated"` // 最后更新时间
}
```

### 主要函数

#### GetInfo() (*MemoryInfo, error)
获取内存基本信息。

#### GetSwapInfo() (*SwapInfo, error)
获取交换分区信息。

#### GetStats() (map[string]interface{}, error)
获取详细内存统计信息。

#### GetVirtualMemoryInfo() (map[string]interface{}, error)
获取虚拟内存信息（Apple Silicon统一内存架构）。

---

## 🎮 GPU模块

### 导入
```go
import "native-monitor/gpu"
```

### 数据结构

#### GPUInfo
```go
type GPUInfo struct {
    Name            string    `json:"name"`              // GPU名称
    Vendor          string    `json:"vendor"`            // 厂商
    Model           string    `json:"model"`             // 型号
    Architecture    string    `json:"architecture"`      // 架构
    Cores           int       `json:"cores"`             // 核心数
    ComputeUnits    int       `json:"compute_units"`     // 计算单元数
    Memory          uint64    `json:"memory"`            // 显存大小
    MemoryType      string    `json:"memory_type"`       // 显存类型
    MemoryBandwidth float64   `json:"memory_bandwidth"`  // 内存带宽
    ClockSpeed      float64   `json:"clock_speed"`       // 基础时钟频率
    BoostClock      float64   `json:"boost_clock"`       // 加速时钟频率
    PowerDraw       float64   `json:"power_draw"`        // 功耗
    Temperature     float64   `json:"temperature"`       // 温度
    DriverVersion   string    `json:"driver_version"`    // 驱动版本
    IsIntegrated    bool      `json:"is_integrated"`     // 是否集成显卡
    IsDiscrete      bool      `json:"is_discrete"`       // 是否独立显卡
    LastUpdated     time.Time `json:"last_updated"`      // 最后更新时间
}
```

#### GPUUsage
```go
type GPUUsage struct {
    GPUPercent      float64   `json:"gpu_percent"`      // GPU使用率
    MemoryPercent   float64   `json:"memory_percent"`   // 显存使用率
    MemoryUsed      uint64    `json:"memory_used"`      // 已用显存
    MemoryFree      uint64    `json:"memory_free"`      // 空闲显存
    PowerUsage      float64   `json:"power_usage"`      // 当前功耗
    Temperature     float64   `json:"temperature"`      // 当前温度
    FanSpeed        float64   `json:"fan_speed"`        // 风扇转速
    ClockSpeed      float64   `json:"clock_speed"`      // 当前时钟频率
    MemoryClock     float64   `json:"memory_clock"`     // 显存时钟频率
    LastUpdated     time.Time `json:"last_updated"`     // 最后更新时间
}
```

#### AppleGPUInfo
```go
type AppleGPUInfo struct {
    ChipName        string  `json:"chip_name"`        // 芯片名称
    GPUCores        int     `json:"gpu_cores"`        // GPU核心数
    TileMemory      uint64  `json:"tile_memory"`      // Tile内存
    UnifiedMemory   uint64  `json:"unified_memory"`   // 统一内存
    MemoryBandwidth float64 `json:"memory_bandwidth"` // 内存带宽
    MetalVersion    string  `json:"metal_version"`    // Metal版本
    TBDRCapable     bool    `json:"tbdr_capable"`     // 是否支持TBDR
}
```

### 主要函数

#### GetGPUs() ([]*GPUInfo, error)
获取所有GPU信息。

**示例:**
```go
gpus, err := gpu.GetGPUs()
if err != nil {
    log.Fatal(err)
}
for _, g := range gpus {
    fmt.Printf("GPU: %s (%d核)\n", g.Name, g.Cores)
}
```

#### GetPrimaryGPU() (*GPUInfo, error)
获取主GPU信息。

#### GetGPUUsage() ([]*GPUUsage, error)
获取GPU使用率信息。采用多重检测策略：
1. IORegistry检测
2. Activity Monitor分析
3. System Profiler检测
4. 进程分析法
5. PowerMetrics备选

#### GetGPUProcesses() ([]*GPUProcess, error)
获取使用GPU的进程列表。

#### GetAppleGPUInfo() (*AppleGPUInfo, error)
获取Apple GPU特有信息（仅Apple Silicon）。

#### IsAppleGPU() bool
检查是否为Apple GPU。

#### GetGPUSummary() (map[string]interface{}, error)
获取GPU概览信息。

#### FormatMemory(bytes uint64) string
格式化显存大小为可读格式。

---

## 💿 磁盘模块

### 导入
```go
import "native-monitor/disk"
```

### 数据结构

#### DiskUsage
```go
type DiskUsage struct {
    Path        string    `json:"path"`         // 挂载路径
    Filesystem  string    `json:"filesystem"`   // 文件系统类型
    Total       uint64    `json:"total"`        // 总空间
    Used        uint64    `json:"used"`         // 已用空间
    Free        uint64    `json:"free"`         // 空闲空间
    UsedPercent float64   `json:"used_percent"` // 使用率百分比
    LastUpdated time.Time `json:"last_updated"` // 最后更新时间
}
```

#### DiskIOStats
```go
type DiskIOStats struct {
    Device      string    `json:"device"`       // 设备名称
    ReadBytes   uint64    `json:"read_bytes"`   // 读取字节数
    WriteBytes  uint64    `json:"write_bytes"`  // 写入字节数
    ReadOps     uint64    `json:"read_ops"`     // 读操作次数
    WriteOps    uint64    `json:"write_ops"`    // 写操作次数
    ReadTime    uint64    `json:"read_time"`    // 读取时间
    WriteTime   uint64    `json:"write_time"`   // 写入时间
    LastUpdated time.Time `json:"last_updated"` // 最后更新时间
}
```

### 主要函数

#### GetUsage(path string) (*DiskUsage, error)
获取指定路径的磁盘使用情况。

#### GetIOStats() ([]*DiskIOStats, error)
获取磁盘I/O统计信息。

#### GetHealth(device string) (map[string]interface{}, error)
获取磁盘健康状态。

#### GetPartitions() ([]PartitionInfo, error)
获取所有分区信息。

---

## 🌐 网络模块

### 导入
```go
import "native-monitor/network"
```

### 数据结构

#### NetworkInterface
```go
type NetworkInterface struct {
    Name        string   `json:"name"`         // 接口名称
    Index       int      `json:"index"`        // 接口索引
    MTU         int      `json:"mtu"`          // 最大传输单元
    HardwareAddr string  `json:"hardware_addr"` // MAC地址
    Flags       []string `json:"flags"`        // 接口标志
    Addresses   []string `json:"addresses"`    // IP地址列表
    IsUp        bool     `json:"is_up"`        // 是否启用
    IsLoopback  bool     `json:"is_loopback"`  // 是否回环接口
}
```

#### NetworkSpeed
```go
type NetworkSpeed struct {
    Interface    string    `json:"interface"`     // 接口名称
    Upload       uint64    `json:"upload"`        // 上传速度 (bytes/s)
    Download     uint64    `json:"download"`      // 下载速度 (bytes/s)
    UploadTotal  uint64    `json:"upload_total"`  // 总上传量
    DownloadTotal uint64   `json:"download_total"` // 总下载量
    LastUpdated  time.Time `json:"last_updated"`  // 最后更新时间
}
```

### 主要函数

#### GetInterfaces() ([]*NetworkInterface, error)
获取所有网络接口信息。

#### MonitorRealTime(interval time.Duration) (*RealTimeMonitor, error)
创建实时网络速度监控器。

**示例:**
```go
monitor, err := network.MonitorRealTime(time.Second)
if err != nil {
    log.Fatal(err)
}

go monitor.Start()
defer monitor.Stop()

// 获取实时速度
speed := monitor.GetSpeed("en0")
fmt.Printf("上传: %s/s, 下载: %s/s\n", 
    formatSpeed(speed.Upload), 
    formatSpeed(speed.Download))
```

#### IsValidInterface(name string) bool
检查接口名称是否有效。

---

## 📊 流量统计模块

### 导入
```go
import "native-monitor/stats"
```

### 数据结构

#### TrafficRecord
```go
type TrafficRecord struct {
    Timestamp   time.Time `json:"timestamp"`    // 时间戳
    Interface   string    `json:"interface"`    // 接口名称
    Upload      uint64    `json:"upload"`       // 上传量
    Download    uint64    `json:"download"`     // 下载量
    TotalUpload uint64    `json:"total_upload"` // 累计上传
    TotalDownload uint64  `json:"total_download"` // 累计下载
}
```

#### TrafficSummary
```go
type TrafficSummary struct {
    Period        string    `json:"period"`         // 统计周期
    StartTime     time.Time `json:"start_time"`     // 开始时间
    EndTime       time.Time `json:"end_time"`       // 结束时间
    TotalUpload   uint64    `json:"total_upload"`   // 总上传量
    TotalDownload uint64    `json:"total_download"` // 总下载量
    PeakUpload    uint64    `json:"peak_upload"`    // 峰值上传速度
    PeakDownload  uint64    `json:"peak_download"`  // 峰值下载速度
    AvgUpload     uint64    `json:"avg_upload"`     // 平均上传速度
    AvgDownload   uint64    `json:"avg_download"`   // 平均下载速度
}
```

### 主要函数

#### NewTrafficCollector() *TrafficCollector
创建新的流量收集器。

#### (tc *TrafficCollector) Start()
开始收集流量数据。

#### (tc *TrafficCollector) Stop()
停止收集流量数据。

#### (tc *TrafficCollector) GetDailyStats() []*TrafficSummary
获取日流量统计。

#### (tc *TrafficCollector) GetWeeklyStats() []*TrafficSummary
获取周流量统计。

#### (tc *TrafficCollector) GetMonthlyStats() []*TrafficSummary
获取月流量统计。

**示例:**
```go
collector := stats.NewTrafficCollector()
go collector.Start()
defer collector.Stop()

// 获取今日流量
dailyStats := collector.GetDailyStats()
for _, stat := range dailyStats {
    fmt.Printf("日期: %s, 上传: %s, 下载: %s\n",
        stat.StartTime.Format("2006-01-02"),
        formatBytes(stat.TotalUpload),
        formatBytes(stat.TotalDownload))
}
```

---

## 🔍 平台检测模块

### 导入
```go
import "native-monitor/platform"
```

### 数据结构

#### PlatformInfo
```go
type PlatformInfo struct {
    OS           string `json:"os"`            // 操作系统
    Architecture string `json:"architecture"`  // 架构
    Hostname     string `json:"hostname"`      // 主机名
    Platform     string `json:"platform"`      // 平台
    Family       string `json:"family"`        // 系统家族
    Version      string `json:"version"`       // 系统版本
    KernelVersion string `json:"kernel_version"` // 内核版本
}
```

### 主要函数

#### GetPlatformInfo() (*PlatformInfo, error)
获取平台基本信息。

#### GetHardwarePlatform() (string, error)
获取硬件平台信息。

#### GetCapabilities() map[string]bool
获取平台功能支持情况。

#### IsVirtualMachine() bool
检查是否运行在虚拟机中。

#### IsContainer() bool
检查是否运行在容器中。

#### IsAppleSilicon() bool
检查是否为Apple Silicon平台。

---

## 🌍 IP地理位置模块

### 导入
```go
import "native-monitor/ipgeo"
```

### 数据结构

#### LocationInfo
```go
type LocationInfo struct {
    IP          string    `json:"ip"`           // IP地址
    Country     string    `json:"country"`      // 国家
    Region      string    `json:"region"`       // 地区/省份
    City        string    `json:"city"`         // 城市
    District    string    `json:"district"`     // 区县
    ISP         string    `json:"isp"`          // 网络服务提供商
    Organization string   `json:"organization"` // 组织
    Timezone    string    `json:"timezone"`     // 时区
    Latitude    float64   `json:"latitude"`     // 纬度
    Longitude   float64   `json:"longitude"`    // 经度
    IsChinaIP   bool      `json:"is_china_ip"`  // 是否中国IP
    LastUpdated time.Time `json:"last_updated"` // 最后更新时间
}
```

#### VoreIPResponse
```go
type VoreIPResponse struct {
    Code   int    `json:"code"`
    Msg    string `json:"msg"`
    IPInfo struct {
        Type  string `json:"type"`
        Text  string `json:"text"`
        CnIP  bool   `json:"cnip"`
    } `json:"ipinfo"`
    IPData struct {
        Info1 string `json:"info1"` // 省份
        Info2 string `json:"info2"` // 城市  
        Info3 string `json:"info3"` // 区县
        ISP   string `json:"isp"`   // 运营商
    } `json:"ipdata"`
    AdCode struct {
        O string `json:"o"` // 完整位置
        P string `json:"p"` // 省份
        C string `json:"c"` // 城市
        N string `json:"n"` // 简化名称
        R string `json:"r"` // 地区名称
        A string `json:"a"` // 行政代码
        I bool   `json:"i"` // 是否中国
    } `json:"adcode"`
    Tips string `json:"tips"`
    Time int64  `json:"time"`
}
```

### 主要函数

#### GetLocalIP() (string, error)
从ip.3322.net获取本地外网IP。

**示例:**
```go
localIP, err := ipgeo.GetLocalIP()
if err != nil {
    log.Fatal(err)
}
fmt.Printf("本地IP: %s\n", localIP)
```

#### GetLocationByIP(ip string) (*LocationInfo, error)
根据IP地址获取地理位置信息。

#### GetProxyIPLocation() (*LocationInfo, error)
获取代理IP和地理位置信息。

#### QuickGetBothLocations() (local, proxy *LocationInfo, err error)
并发获取本地IP和代理IP的地理位置信息。

**示例:**
```go
local, proxy, err := ipgeo.QuickGetBothLocations()
if err != nil {
    log.Fatal(err)
}

localLoc := ipgeo.FormatLocation(local)
proxyLoc := ipgeo.FormatLocation(proxy)

fmt.Printf("本地IP: %s / %s\n", local.IP, localLoc)
fmt.Printf("代理IP: %s / %s\n", proxy.IP, proxyLoc)
```

#### FormatLocation(info *LocationInfo) string
格式化地理位置为可读字符串。
- 中国IP: "广州市-番禺区"
- 国外IP: "United States"

#### GetLocationDifference(loc1, loc2 *LocationInfo) map[string]interface{}
比较两个位置的差异。

#### GetNetworkSummary() (map[string]interface{}, error)
获取网络连接概览。

#### BatchGetLocations(ips []string) ([]*LocationInfo, error)
批量获取多个IP的地理位置。

---

## 🛠️ 工具函数

### 格式化函数

#### formatBytes(bytes uint64) string
格式化字节数为可读格式。
```go
fmt.Println(formatBytes(1024))      // "1.0 KB"
fmt.Println(formatBytes(1048576))   // "1.0 MB"
```

#### formatSpeed(bytesPerSec uint64) string
格式化网络速度为可读格式。
```go
fmt.Println(formatSpeed(1048576))   // "1.0 MB/s"
```

#### formatDuration(seconds uint64) string
格式化时间长度为可读格式。
```go
fmt.Println(formatDuration(3661))   // "1小时1分钟"
```

### 检测函数

#### isAppleSilicon() bool
检查当前平台是否为Apple Silicon。

#### isValidInterface(name string) bool
检查网络接口名称是否有效。

#### isGPUProcess(processName string) bool
检查进程是否为GPU相关进程。

---

## 📝 使用注意事项

### 权限要求
- **macOS**: 某些功能可能需要管理员权限（如powermetrics）
- **Linux**: 读取/proc和/sys需要适当权限
- **Windows**: WMI查询可能需要管理员权限

### 性能考虑
- 使用缓存减少重复系统调用
- 避免在高频循环中调用昂贵的操作
- 并发获取多个信息源时注意资源限制

### 错误处理
- 所有函数都返回错误，请妥善处理
- 网络相关功能可能因为网络问题失败
- 平台特定功能在不支持的平台上会返回错误

### 最佳实践
- 使用适当的刷新间隔（建议1-5秒）
- 缓存静态信息（CPU型号、内存大小等）
- 使用协程处理长时间运行的监控任务
- 及时释放监控器资源（调用Stop()方法）
