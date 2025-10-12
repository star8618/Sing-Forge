#!/bin/bash

# SingForge 自动安装脚本
# 从 GitHub Releases 下载最新版本并自动安装到 /opt/singforge

set -e  # 遇到错误立即退出

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# 配置
REPO="star8618/Sing-Forge"
INSTALL_DIR="/opt/singforge"
TEMP_DIR="/tmp/singforge-install-$$"
GITHUB_API="https://api.github.com/repos/${REPO}/releases/latest"

# 全局变量
LATEST_VERSION=""
PACKAGE_NAME=""
DOWNLOAD_URL=""

echo -e "${BLUE}╔════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║   SingForge 自动安装脚本                   ║${NC}"
echo -e "${BLUE}║   从 GitHub Releases 自动下载安装          ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════╝${NC}"
echo ""

# 1. 系统检查
check_system() {
    echo -e "${CYAN}[1/8] 检查系统环境...${NC}"
    
    # 检测操作系统
    if [[ "$OSTYPE" != "linux-gnu"* ]]; then
        echo -e "${RED}❌ 错误: 此脚本仅支持 Linux 系统${NC}"
        echo -e "${YELLOW}💡 当前系统: $OSTYPE${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ 检测到 Linux 系统${NC}"
    
    # 检查 root 权限
    if [ "$EUID" -ne 0 ]; then 
        echo -e "${RED}❌ 错误: 需要 root 权限${NC}"
        echo -e "${YELLOW}💡 请使用 sudo 运行此脚本:${NC}"
        echo -e "${YELLOW}   sudo bash install-from-github.sh${NC}"
        exit 1
    fi
    echo -e "${GREEN}✓ 已获取 root 权限${NC}"
    
    # 检测系统架构
    ARCH=$(uname -m)
    case "$ARCH" in
        x86_64|amd64)
            PACKAGE_NAME="linux-amd64.tar.gz"
            echo -e "${GREEN}✓ 架构: x86_64 (amd64)${NC}"
            ;;
        aarch64|arm64)
            PACKAGE_NAME="linux-arm64.tar.gz"
            echo -e "${GREEN}✓ 架构: ARM64${NC}"
            ;;
        *)
            echo -e "${RED}❌ 不支持的架构: $ARCH${NC}"
            echo -e "${YELLOW}💡 支持的架构: x86_64, aarch64${NC}"
            exit 1
            ;;
    esac
    
    # 检查必要命令
    for cmd in curl tar; do
        if ! command -v $cmd &> /dev/null; then
            echo -e "${RED}❌ 缺少必要命令: $cmd${NC}"
            echo -e "${YELLOW}💡 请安装: $cmd${NC}"
            exit 1
        fi
    done
    echo -e "${GREEN}✓ 必要命令已安装${NC}"
    echo ""
}

# 2. 检测已安装版本
check_installed() {
    echo -e "${CYAN}[2/8] 检测已安装版本...${NC}"
    
    if [ -d "$INSTALL_DIR" ] && [ -f "$INSTALL_DIR/singforge-backend" ]; then
        CURRENT_VERSION="已安装"
        if [ -f "$INSTALL_DIR/VERSION" ]; then
            CURRENT_VERSION=$(cat "$INSTALL_DIR/VERSION")
        fi
        
        echo -e "${YELLOW}⚠️  检测到已安装 SingForge${NC}"
        echo -e "${YELLOW}   当前版本: ${CURRENT_VERSION}${NC}"
        echo -e "${YELLOW}   安装路径: ${INSTALL_DIR}${NC}"
        echo ""
        echo -e "${BLUE}是否卸载并重新安装最新版本? (y/n)${NC}"
        read -r response
        
        if [[ ! "$response" =~ ^[Yy]$ ]]; then
            echo -e "${YELLOW}已取消安装${NC}"
            exit 0
        fi
        
        echo -e "${YELLOW}开始卸载旧版本...${NC}"
        uninstall_old
    else
        echo -e "${GREEN}✓ 未检测到已安装版本，准备全新安装${NC}"
    fi
    echo ""
}

# 卸载旧版本
uninstall_old() {
    # 停止运行的服务（兼容多种 Linux 发行版）
    if command -v pgrep >/dev/null 2>&1; then
        if pgrep -f "singforge-backend" > /dev/null; then
            echo -e "${YELLOW}停止运行中的服务...${NC}"
            pkill -f "singforge-backend" || true
            sleep 2
        fi
    else
        # Alpine Linux 等精简系统的备用方案
        PID=$(ps aux | grep 'singforge-backend' | grep -v grep | awk '{print $2}' | head -1)
        if [ ! -z "$PID" ]; then
            echo -e "${YELLOW}停止运行中的服务 (PID: $PID)...${NC}"
            kill -15 "$PID" 2>/dev/null || true
            sleep 2
        fi
    fi
    
    # 删除旧文件
    echo -e "${YELLOW}删除旧文件: ${INSTALL_DIR}${NC}"
    rm -rf "$INSTALL_DIR"
    
    # 删除符号链接
    if [ -L "/usr/local/bin/singforge" ]; then
        rm -f "/usr/local/bin/singforge"
    fi
    
    echo -e "${GREEN}✓ 旧版本已卸载${NC}"
}

# 3. 获取最新版本
get_latest_version() {
    echo -e "${CYAN}[3/8] 获取最新版本信息...${NC}"
    
    # 从 GitHub API 获取最新版本
    echo -e "${YELLOW}正在查询 GitHub API...${NC}"
    
    # 尝试使用 jq 解析（如果可用）
    if command -v jq &> /dev/null; then
        LATEST_VERSION=$(curl -s "$GITHUB_API" | jq -r '.tag_name' 2>/dev/null)
    else
        # 备用方案：使用 grep 和 sed
        LATEST_VERSION=$(curl -s "$GITHUB_API" | grep '"tag_name"' | sed -E 's/.*"tag_name": "([^"]+)".*/\1/')
    fi
    
    if [ -z "$LATEST_VERSION" ] || [ "$LATEST_VERSION" == "null" ]; then
        echo -e "${RED}❌ 无法获取最新版本信息${NC}"
        echo -e "${YELLOW}💡 请检查网络连接或稍后重试${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✓ 最新版本: ${LATEST_VERSION}${NC}"
    echo ""
}

# 4. 下载安装包
download_package() {
    echo -e "${CYAN}[4/8] 下载安装包...${NC}"
    
    # 创建临时目录
    mkdir -p "$TEMP_DIR"
    cd "$TEMP_DIR"
    
    # 构建下载地址
    CDN_URL="https://cdn.jsdelivr.net/gh/${REPO}@${LATEST_VERSION}/dist/${PACKAGE_NAME}"
    GITHUB_URL="https://github.com/${REPO}/releases/download/${LATEST_VERSION}/${PACKAGE_NAME}"
    
    DOWNLOAD_FILE="${TEMP_DIR}/${PACKAGE_NAME}"
    
    # 方法1: 优先使用 jsdelivr CDN
    echo -e "${BLUE}尝试从 CDN 下载 (jsdelivr)...${NC}"
    echo -e "${YELLOW}地址: ${CDN_URL}${NC}"
    
    if curl -L --progress-bar --connect-timeout 10 --max-time 300 -o "$DOWNLOAD_FILE" "$CDN_URL" 2>&1; then
        if [ -f "$DOWNLOAD_FILE" ] && [ -s "$DOWNLOAD_FILE" ]; then
            echo -e "${GREEN}✓ CDN 下载成功${NC}"
            FILE_SIZE=$(du -h "$DOWNLOAD_FILE" | cut -f1)
            echo -e "${GREEN}✓ 文件大小: ${FILE_SIZE}${NC}"
            return 0
        fi
    fi
    
    # 方法2: 备用 GitHub 官方地址
    echo -e "${YELLOW}CDN 下载失败，切换到 GitHub 官方地址...${NC}"
    echo -e "${YELLOW}地址: ${GITHUB_URL}${NC}"
    
    rm -f "$DOWNLOAD_FILE"
    
    if curl -L --progress-bar --connect-timeout 10 --max-time 600 -o "$DOWNLOAD_FILE" "$GITHUB_URL" 2>&1; then
        if [ -f "$DOWNLOAD_FILE" ] && [ -s "$DOWNLOAD_FILE" ]; then
            echo -e "${GREEN}✓ GitHub 下载成功${NC}"
            FILE_SIZE=$(du -h "$DOWNLOAD_FILE" | cut -f1)
            echo -e "${GREEN}✓ 文件大小: ${FILE_SIZE}${NC}"
            return 0
        fi
    fi
    
    # 下载失败
    echo -e "${RED}❌ 下载失败${NC}"
    echo -e "${YELLOW}💡 请检查网络连接或手动下载${NC}"
    echo -e "${YELLOW}💡 手动下载地址: ${GITHUB_URL}${NC}"
    cleanup
    exit 1
}

# 5. 解压和安装
install_package() {
    echo -e "${CYAN}[5/8] 解压和安装...${NC}"
    
    DOWNLOAD_FILE="${TEMP_DIR}/${PACKAGE_NAME}"
    
    # 解压到临时目录
    echo -e "${YELLOW}正在解压文件...${NC}"
    tar -xzf "$DOWNLOAD_FILE" -C "$TEMP_DIR"
    
    # 查找解压后的目录
    EXTRACT_DIR=$(find "$TEMP_DIR" -mindepth 1 -maxdepth 1 -type d | head -1)
    
    if [ -z "$EXTRACT_DIR" ] || [ ! -d "$EXTRACT_DIR" ]; then
        echo -e "${RED}❌ 解压失败或找不到解压目录${NC}"
        cleanup
        exit 1
    fi
    
    echo -e "${GREEN}✓ 解压完成: ${EXTRACT_DIR}${NC}"
    
    # 检查必要文件
    if [ ! -f "$EXTRACT_DIR/singforge-backend" ]; then
        echo -e "${RED}❌ 未找到 singforge-backend 可执行文件${NC}"
        cleanup
        exit 1
    fi
    
    # 创建安装目录
    echo -e "${YELLOW}创建安装目录: ${INSTALL_DIR}${NC}"
    mkdir -p "$INSTALL_DIR"
    
    # 复制所有文件
    echo -e "${YELLOW}复制文件到安装目录...${NC}"
    cp -r "$EXTRACT_DIR"/* "$INSTALL_DIR/"
    
    # 保存版本信息
    echo "$LATEST_VERSION" > "$INSTALL_DIR/VERSION"
    
    echo -e "${GREEN}✓ 文件安装完成${NC}"
    echo ""
}

# 6. 设置权限
set_permissions() {
    echo -e "${CYAN}[6/8] 设置文件权限...${NC}"
    
    # 后端程序
    chmod 755 "$INSTALL_DIR/singforge-backend"
    echo -e "${GREEN}✓ singforge-backend (755)${NC}"
    
    # sing-box 核心
    if [ -f "$INSTALL_DIR/data/cores/sing-box" ]; then
        chmod 755 "$INSTALL_DIR/data/cores/sing-box"
        echo -e "${GREEN}✓ sing-box 核心 (755)${NC}"
    fi
    
    # 脚本文件
    for script in start.sh stop.sh test-singbox.sh auto-update.sh install-linux.sh install-from-github.sh; do
        if [ -f "$INSTALL_DIR/$script" ]; then
            chmod 755 "$INSTALL_DIR/$script"
            echo -e "${GREEN}✓ $script (755)${NC}"
        fi
    done
    
    # 创建符号链接
    ln -sf "$INSTALL_DIR/start.sh" /usr/local/bin/singforge
    echo -e "${GREEN}✓ 已创建命令: singforge${NC}"
    echo ""
}

# 7. 切换到安装目录
change_directory() {
    echo -e "${CYAN}[7/8] 切换到安装目录...${NC}"
    cd "$INSTALL_DIR"
    echo -e "${GREEN}✓ 当前目录: $(pwd)${NC}"
    echo ""
}

# 8. 启动服务
start_service() {
    echo -e "${CYAN}[8/8] 启动服务...${NC}"
    
    # 确保日志目录存在
    mkdir -p data/logs
    
    # 后台启动服务
    nohup ./singforge-backend > data/logs/backend.log 2>&1 &
    BACKEND_PID=$!
    
    echo -e "${GREEN}✓ 服务已启动 (PID: ${BACKEND_PID})${NC}"
    
    # 等待服务就绪
    echo -e "${YELLOW}等待服务就绪...${NC}"
    for i in {1..15}; do
        sleep 1
        if curl -s http://localhost:8383/api/health > /dev/null 2>&1; then
            echo -e "${GREEN}✓ 服务已就绪！${NC}"
            break
        fi
        echo -n "."
    done
    echo ""
}

# 9. 清理临时文件
cleanup() {
    if [ -d "$TEMP_DIR" ]; then
        echo -e "${YELLOW}清理临时文件...${NC}"
        rm -rf "$TEMP_DIR"
        echo -e "${GREEN}✓ 临时文件已清理${NC}"
    fi
}

# 显示安装完成信息
show_success() {
    # 获取本机IP
    get_local_ip() {
        if command -v ip &> /dev/null; then
            ip addr show | grep 'inet ' | grep -v '127.0.0.1' | awk '{print $2}' | cut -d/ -f1 | head -n1
        elif command -v ifconfig &> /dev/null; then
            ifconfig | grep 'inet ' | grep -v '127.0.0.1' | awk '{print $2}' | head -n1
        else
            echo "unknown"
        fi
    }
    
    LOCAL_IP=$(get_local_ip)
    
    echo ""
    echo -e "${GREEN}╔════════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║           ✅ 安装完成！                    ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${CYAN}📦 安装信息:${NC}"
    echo -e "   版本: ${GREEN}${LATEST_VERSION}${NC}"
    echo -e "   路径: ${YELLOW}${INSTALL_DIR}${NC}"
    echo -e "   数据: ${YELLOW}${INSTALL_DIR}/data${NC}"
    echo -e "   日志: ${YELLOW}${INSTALL_DIR}/data/logs${NC}"
    echo ""
    echo -e "${CYAN}🌐 访问地址:${NC}"
    echo -e "   本地: ${GREEN}http://localhost:8383${NC}"
    if [ "$LOCAL_IP" != "unknown" ] && [ ! -z "$LOCAL_IP" ]; then
        echo -e "   局域网: ${GREEN}http://${LOCAL_IP}:8383${NC}"
    fi
    echo ""
    echo -e "${CYAN}🔐 默认账号:${NC}"
    echo -e "   用户名: ${YELLOW}admin${NC}"
    echo -e "   密码:   ${YELLOW}admin123${NC}"
    echo ""
    echo -e "${CYAN}🎮 管理命令:${NC}"
    echo -e "   启动: ${YELLOW}sudo singforge${NC}"
    echo -e "   停止: ${YELLOW}cd ${INSTALL_DIR} && sudo ./stop.sh${NC}"
    echo -e "   日志: ${YELLOW}tail -f ${INSTALL_DIR}/data/logs/backend.log${NC}"
    echo ""
    echo -e "${GREEN}🎉 开始使用 SingForge 吧！${NC}"
    echo ""
}

# 错误处理
trap 'echo -e "\n${RED}❌ 安装过程中出现错误${NC}"; cleanup; exit 1' ERR

# 主流程
main() {
    check_system
    check_installed
    get_latest_version
    download_package
    install_package
    set_permissions
    change_directory
    start_service
    cleanup
    show_success
}

# 执行主函数
main

