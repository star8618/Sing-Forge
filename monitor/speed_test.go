// 简单网络速度测试
package main

import (
	"fmt"
	"time"

	"native-monitor/network"
)

func main() {
	fmt.Println("🌐 简单网络速度测试")
	fmt.Println("==================")

	// 第一次获取 - 初始化
	fmt.Print("🔄 初始化...")
	speeds1, err := network.GetRealTimeSpeed()
	if err != nil {
		fmt.Printf("❌ 失败: %v\n", err)
		return
	}
	fmt.Printf("✅ 完成 (获取到 %d 个接口)\n", len(speeds1))

	// 等待2秒
	fmt.Print("⏰ 等待2秒...")
	time.Sleep(2 * time.Second)
	fmt.Println("✅")

	// 第二次获取 - 计算速度
	fmt.Print("📊 获取实时速度...")
	speeds2, err := network.GetRealTimeSpeed()
	if err != nil {
		fmt.Printf("❌ 失败: %v\n", err)
		return
	}
	fmt.Printf("✅ 完成\n")

	// 显示结果
	fmt.Println("\n📈 网络速度结果:")
	fmt.Println("================")

	hasTraffic := false
	for _, speed := range speeds2 {
		if speed.DownloadSpeed > 0 || speed.UploadSpeed > 0 {
			hasTraffic = true
			fmt.Printf("📡 %s:\n", speed.Name)
			fmt.Printf("  ⬇️  下载: %s\n", network.FormatSpeed(speed.DownloadSpeed))
			fmt.Printf("  ⬆️  上传: %s\n", network.FormatSpeed(speed.UploadSpeed))
			fmt.Printf("  📊 累计: ⬇️%s / ⬆️%s\n",
				network.FormatBytes(speed.DownloadTotal),
				network.FormatBytes(speed.UploadTotal))
		}
	}

	if !hasTraffic {
		fmt.Println("💤 当前没有检测到网络流量")

		// 显示总的累计流量
		fmt.Println("\n📊 累计网络流量:")
		for i, speed := range speeds2 {
			if i >= 5 { // 只显示前5个
				break
			}
			if speed.DownloadTotal > 0 || speed.UploadTotal > 0 {
				fmt.Printf("📡 %s: ⬇️%s / ⬆️%s\n",
					speed.Name,
					network.FormatBytes(speed.DownloadTotal),
					network.FormatBytes(speed.UploadTotal))
			}
		}
	}

	fmt.Println("\n✅ 测试完成!")
}
