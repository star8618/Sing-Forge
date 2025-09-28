//go:build darwin

package network

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// getPlatformInterfaces 获取平台网络接口
func getPlatformInterfaces() ([]NetworkInterface, error) {
	return getDarwinInterfaces()
}

// getPlatformInterfaceStats 获取平台接口统计
func getPlatformInterfaceStats() ([]NetworkStats, error) {
	return getDarwinInterfaceStats()
}

// getPlatformConnections 获取平台连接信息
func getPlatformConnections() ([]ConnectionInfo, error) {
	return getDarwinConnections()
}

// getDarwinInterfaces 获取macOS网络接口信息
func getDarwinInterfaces() ([]NetworkInterface, error) {
	var interfaces []NetworkInterface

	// 1. 获取基本接口信息
	netInterfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	// 2. 为每个接口获取详细信息
	for _, iface := range netInterfaces {
		netIface := NetworkInterface{
			Name:        iface.Name,
			DisplayName: iface.Name,
			MAC:         iface.HardwareAddr.String(),
			MTU:         iface.MTU,
			IsUp:        iface.Flags&net.FlagUp != 0,
			IsRunning:   iface.Flags&net.FlagRunning != 0,
			IsLoopback:  iface.Flags&net.FlagLoopback != 0,
		}

		// 3. 获取IP地址
		if addrs, err := iface.Addrs(); err == nil {
			for _, addr := range addrs {
				if ipnet, ok := addr.(*net.IPNet); ok {
					if ipnet.IP.To4() != nil {
						netIface.IPv4 = append(netIface.IPv4, ipnet.IP.String())
					} else if ipnet.IP.To16() != nil {
						netIface.IPv6 = append(netIface.IPv6, ipnet.IP.String())
					}
				}
			}
		}

		// 4. 获取硬件类型和速度信息
		if err := getDarwinInterfaceDetails(&netIface); err == nil {
			// 详细信息获取成功
		}

		interfaces = append(interfaces, netIface)
	}

	return interfaces, nil
}

// getDarwinInterfaceDetails 获取macOS接口详细信息
func getDarwinInterfaceDetails(iface *NetworkInterface) error {
	// 使用networksetup命令获取详细信息
	cmd := exec.Command("networksetup", "-listallhardwareports")
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	// 解析networksetup输出
	scanner := bufio.NewScanner(bytes.NewReader(output))
	var currentPort string
	var currentDevice string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if strings.HasPrefix(line, "Hardware Port:") {
			currentPort = strings.TrimSpace(strings.TrimPrefix(line, "Hardware Port:"))
		} else if strings.HasPrefix(line, "Device:") {
			currentDevice = strings.TrimSpace(strings.TrimPrefix(line, "Device:"))

			// 如果设备名匹配，设置硬件信息
			if currentDevice == iface.Name {
				iface.DisplayName = currentPort
				iface.Hardware = determineHardwareType(currentPort)
				iface.IsWireless = strings.Contains(strings.ToLower(currentPort), "wi-fi") ||
					strings.Contains(strings.ToLower(currentPort), "wireless")
				break
			}
		}
	}

	// 使用ifconfig获取更多详细信息
	return getIfconfigDetails(iface)
}

// getIfconfigDetails 使用ifconfig获取接口详细信息
func getIfconfigDetails(iface *NetworkInterface) error {
	cmd := exec.Command("ifconfig", iface.Name)
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 解析状态信息
		if strings.Contains(line, "status:") {
			status := strings.TrimSpace(strings.Split(line, ":")[1])
			iface.IsRunning = status == "active"
		}

		// 解析速度信息 (如果可用)
		if strings.Contains(line, "media:") {
			if speed := extractSpeedFromMedia(line); speed > 0 {
				iface.Speed = speed
			}
		}
	}

	return nil
}

// getDarwinInterfaceStats 获取macOS网络接口统计信息
func getDarwinInterfaceStats() ([]NetworkStats, error) {
	var stats []NetworkStats

	// 使用netstat命令获取统计信息
	cmd := exec.Command("netstat", "-i", "-b")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(bytes.NewReader(output))
	// 跳过标题行
	if scanner.Scan() {
		// 第一行是标题
	}
	if scanner.Scan() {
		// 第二行也是标题的一部分
	}

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// 解析netstat输出
		if stat := parseNetstatLine(line); stat != nil {
			stats = append(stats, *stat)
		}
	}

	return stats, nil
}

// parseNetstatLine 解析netstat输出行
func parseNetstatLine(line string) *NetworkStats {
	fields := strings.Fields(line)
	if len(fields) < 10 {
		return nil
	}

	// netstat -i -b 的输出格式:
	// Name  Mtu   Network       Address            Ipkts Ierrs Ibytes    Opkts Oerrs Obytes  Coll
	name := fields[0]

	// 解析数值字段
	ipkts, _ := strconv.ParseUint(fields[4], 10, 64)
	ierrs, _ := strconv.ParseUint(fields[5], 10, 64)
	ibytes, _ := strconv.ParseUint(fields[6], 10, 64)
	opkts, _ := strconv.ParseUint(fields[7], 10, 64)
	oerrs, _ := strconv.ParseUint(fields[8], 10, 64)
	obytes, _ := strconv.ParseUint(fields[9], 10, 64)

	return &NetworkStats{
		Name:            name,
		BytesReceived:   ibytes,
		BytesSent:       obytes,
		PacketsReceived: ipkts,
		PacketsSent:     opkts,
		ErrorsReceived:  ierrs,
		ErrorsSent:      oerrs,
	}
}

// getDarwinConnections 获取macOS网络连接信息
func getDarwinConnections() ([]ConnectionInfo, error) {
	var connections []ConnectionInfo

	// 获取TCP连接
	tcpConns, err := getDarwinTCPConnections()
	if err == nil {
		connections = append(connections, tcpConns...)
	}

	// 获取UDP连接
	udpConns, err := getDarwinUDPConnections()
	if err == nil {
		connections = append(connections, udpConns...)
	}

	return connections, nil
}

// getDarwinTCPConnections 获取TCP连接
func getDarwinTCPConnections() ([]ConnectionInfo, error) {
	cmd := exec.Command("netstat", "-an", "-p", "tcp")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var connections []ConnectionInfo
	scanner := bufio.NewScanner(bytes.NewReader(output))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "tcp") {
			continue
		}

		if conn := parseTCPConnection(line); conn != nil {
			connections = append(connections, *conn)
		}
	}

	return connections, nil
}

// getDarwinUDPConnections 获取UDP连接
func getDarwinUDPConnections() ([]ConnectionInfo, error) {
	cmd := exec.Command("netstat", "-an", "-p", "udp")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var connections []ConnectionInfo
	scanner := bufio.NewScanner(bytes.NewReader(output))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "udp") {
			continue
		}

		if conn := parseUDPConnection(line); conn != nil {
			connections = append(connections, *conn)
		}
	}

	return connections, nil
}

// parseTCPConnection 解析TCP连接行
func parseTCPConnection(line string) *ConnectionInfo {
	fields := strings.Fields(line)
	if len(fields) < 6 {
		return nil
	}

	protocol := fields[0]
	localAddr := fields[3]
	remoteAddr := fields[4]
	state := fields[5]

	conn := &ConnectionInfo{
		Protocol: protocol,
		State:    state,
	}

	// 解析本地地址和端口
	if localHost, localPort, err := parseAddress(localAddr); err == nil {
		conn.LocalAddr = localHost
		if port, err := strconv.ParseUint(localPort, 10, 16); err == nil {
			conn.LocalPort = uint16(port)
		}
	}

	// 解析远程地址和端口
	if remoteHost, remotePort, err := parseAddress(remoteAddr); err == nil {
		conn.RemoteAddr = remoteHost
		if port, err := strconv.ParseUint(remotePort, 10, 16); err == nil {
			conn.RemotePort = uint16(port)
		}
	}

	return conn
}

// parseUDPConnection 解析UDP连接行
func parseUDPConnection(line string) *ConnectionInfo {
	fields := strings.Fields(line)
	if len(fields) < 4 {
		return nil
	}

	protocol := fields[0]
	localAddr := fields[3]

	conn := &ConnectionInfo{
		Protocol: protocol,
		State:    "LISTEN", // UDP没有连接状态，标记为LISTEN
	}

	// 解析本地地址和端口
	if localHost, localPort, err := parseAddress(localAddr); err == nil {
		conn.LocalAddr = localHost
		if port, err := strconv.ParseUint(localPort, 10, 16); err == nil {
			conn.LocalPort = uint16(port)
		}
	}

	return conn
}

// 辅助函数

// determineHardwareType 确定硬件类型
func determineHardwareType(portName string) string {
	portLower := strings.ToLower(portName)

	if strings.Contains(portLower, "ethernet") {
		return "ethernet"
	} else if strings.Contains(portLower, "wi-fi") || strings.Contains(portLower, "wireless") {
		return "wifi"
	} else if strings.Contains(portLower, "bluetooth") {
		return "bluetooth"
	} else if strings.Contains(portLower, "thunderbolt") {
		return "thunderbolt"
	} else if strings.Contains(portLower, "usb") {
		return "usb"
	}

	return "unknown"
}

// extractSpeedFromMedia 从media字符串中提取速度
func extractSpeedFromMedia(mediaLine string) uint64 {
	// 查找类似 "1000baseT" 的模式
	re := regexp.MustCompile(`(\d+)base`)
	matches := re.FindStringSubmatch(mediaLine)
	if len(matches) >= 2 {
		if speed, err := strconv.ParseUint(matches[1], 10, 64); err == nil {
			return speed * 1000000 // 转换为bps (Mbps -> bps)
		}
	}

	// 查找类似 "100Mb/s" 的模式
	re = regexp.MustCompile(`(\d+)Mb/s`)
	matches = re.FindStringSubmatch(mediaLine)
	if len(matches) >= 2 {
		if speed, err := strconv.ParseUint(matches[1], 10, 64); err == nil {
			return speed * 1000000 // 转换为bps
		}
	}

	return 0
}

// parseAddress 解析地址:端口格式
func parseAddress(addr string) (host, port string, err error) {
	// 处理IPv6地址
	if strings.HasPrefix(addr, "[") {
		// IPv6格式: [::1]:80
		if idx := strings.LastIndex(addr, "]:"); idx != -1 {
			host = addr[1:idx]
			port = addr[idx+2:]
			return host, port, nil
		}
		return "", "", fmt.Errorf("invalid IPv6 address format: %s", addr)
	}

	// IPv4格式: 127.0.0.1:80
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		host = addr[:idx]
		port = addr[idx+1:]
		return host, port, nil
	}

	return "", "", fmt.Errorf("invalid address format: %s", addr)
}
