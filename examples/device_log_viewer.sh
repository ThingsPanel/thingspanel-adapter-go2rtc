#!/bin/bash

# 设备日志查看工具
# 用于快速查看和分析设备日志

DEVICE_LOG_DIR="logs/devices"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 显示帮助信息
show_help() {
    echo "设备日志查看工具"
    echo ""
    echo "使用方法:"
    echo "  $0 list                              # 列出所有设备日志文件"
    echo "  $0 tail <device_id>                  # 实时查看设备日志"
    echo "  $0 view <device_id> [lines]          # 查看设备日志(默认100行)"
    echo "  $0 errors <device_id>                # 查看设备错误日志"
    echo "  $0 connections <device_id>           # 查看设备连接日志"
    echo "  $0 data <device_id>                  # 查看设备数据交互"
    echo "  $0 commands <device_id>              # 查看设备指令日志"
    echo "  $0 search <device_id> <keyword>      # 搜索设备日志"
    echo "  $0 stats                             # 显示日志统计信息"
    echo ""
    echo "示例:"
    echo "  $0 list"
    echo "  $0 tail 00000001"
    echo "  $0 view 00000001 50"
    echo "  $0 errors 00000001"
    echo "  $0 search 00000001 temperature"
}

# 检查设备日志目录是否存在
check_log_dir() {
    if [ ! -d "$DEVICE_LOG_DIR" ]; then
        echo -e "${RED}错误: 设备日志目录不存在: $DEVICE_LOG_DIR${NC}"
        echo "请确保设备独立日志功能已启用并且有设备连接过"
        exit 1
    fi
}

# 列出所有设备日志文件
list_devices() {
    check_log_dir
    
    echo -e "${GREEN}设备日志文件列表:${NC}"
    echo ""
    
    if [ -z "$(ls -A $DEVICE_LOG_DIR 2>/dev/null)" ]; then
        echo -e "${YELLOW}暂无设备日志文件${NC}"
        return
    fi
    
    cd "$DEVICE_LOG_DIR"
    for file in *.log; do
        if [ -f "$file" ]; then
            device_id="${file%.log}"
            size=$(du -sh "$file" | cut -f1)
            modified=$(date -r "$file" "+%Y-%m-%d %H:%M:%S")
            lines=$(wc -l < "$file")
            
            echo -e "${BLUE}设备ID:${NC} $device_id"
            echo -e "  文件大小: $size"
            echo -e "  最后修改: $modified"
            echo -e "  日志行数: $lines"
            echo ""
        fi
    done
}

# 检查设备日志文件是否存在
check_device_log() {
    local device_id=$1
    local log_file="$DEVICE_LOG_DIR/${device_id}.log"
    
    if [ ! -f "$log_file" ]; then
        echo -e "${RED}错误: 设备 $device_id 的日志文件不存在${NC}"
        echo "可用的设备ID:"
        cd "$DEVICE_LOG_DIR" 2>/dev/null || return 1
        for file in *.log; do
            if [ -f "$file" ]; then
                echo "  ${file%.log}"
            fi
        done
        return 1
    fi
    return 0
}

# 实时查看设备日志
tail_device_log() {
    local device_id=$1
    check_log_dir
    check_device_log "$device_id" || return 1
    
    local log_file="$DEVICE_LOG_DIR/${device_id}.log"
    echo -e "${GREEN}实时查看设备 $device_id 的日志 (Ctrl+C 退出):${NC}"
    echo ""
    tail -f "$log_file"
}

# 查看设备日志
view_device_log() {
    local device_id=$1
    local lines=${2:-100}
    check_log_dir
    check_device_log "$device_id" || return 1
    
    local log_file="$DEVICE_LOG_DIR/${device_id}.log"
    echo -e "${GREEN}设备 $device_id 的日志 (最后 $lines 行):${NC}"
    echo ""
    tail -n "$lines" "$log_file"
}

# 查看设备错误日志
view_errors() {
    local device_id=$1
    check_log_dir
    check_device_log "$device_id" || return 1
    
    local log_file="$DEVICE_LOG_DIR/${device_id}.log"
    echo -e "${RED}设备 $device_id 的错误日志:${NC}"
    echo ""
    grep -E "error|failed|Error|Failed" "$log_file" || echo "未发现错误日志"
}

# 查看设备连接日志
view_connections() {
    local device_id=$1
    check_log_dir
    check_device_log "$device_id" || return 1
    
    local log_file="$DEVICE_LOG_DIR/${device_id}.log"
    echo -e "${BLUE}设备 $device_id 的连接日志:${NC}"
    echo ""
    grep -E "connection_established|connection_closed|connection_error|status=" "$log_file" || echo "未发现连接日志"
}

# 查看设备数据交互
view_data() {
    local device_id=$1
    check_log_dir
    check_device_log "$device_id" || return 1
    
    local log_file="$DEVICE_LOG_DIR/${device_id}.log"
    echo -e "${YELLOW}设备 $device_id 的数据交互:${NC}"
    echo ""
    grep "设备数据交互\|data_parsed" "$log_file" || echo "未发现数据交互记录"
}

# 查看设备指令日志
view_commands() {
    local device_id=$1
    check_log_dir
    check_device_log "$device_id" || return 1
    
    local log_file="$DEVICE_LOG_DIR/${device_id}.log"
    echo -e "${GREEN}设备 $device_id 的指令日志:${NC}"
    echo ""
    grep "设备指令\|direction=sent" "$log_file" || echo "未发现指令记录"
}

# 搜索设备日志
search_device_log() {
    local device_id=$1
    local keyword=$2
    check_log_dir
    check_device_log "$device_id" || return 1
    
    local log_file="$DEVICE_LOG_DIR/${device_id}.log"
    echo -e "${GREEN}在设备 $device_id 的日志中搜索 '$keyword':${NC}"
    echo ""
    grep -i "$keyword" "$log_file" || echo "未找到相关记录"
}

# 显示日志统计信息
show_stats() {
    check_log_dir
    
    echo -e "${GREEN}设备日志统计信息:${NC}"
    echo ""
    
    # 设备数量
    device_count=$(ls -1 "$DEVICE_LOG_DIR"/*.log 2>/dev/null | wc -l)
    echo -e "设备数量: ${BLUE}$device_count${NC}"
    
    # 总日志大小
    total_size=$(du -sh "$DEVICE_LOG_DIR" 2>/dev/null | cut -f1)
    echo -e "总日志大小: ${BLUE}$total_size${NC}"
    
    # 总日志行数
    total_lines=0
    if [ "$device_count" -gt 0 ]; then
        cd "$DEVICE_LOG_DIR"
        for file in *.log; do
            if [ -f "$file" ]; then
                lines=$(wc -l < "$file" 2>/dev/null || echo 0)
                total_lines=$((total_lines + lines))
            fi
        done
    fi
    echo -e "总日志行数: ${BLUE}$total_lines${NC}"
    
    echo ""
    echo -e "${YELLOW}最近活跃的设备:${NC}"
    if [ "$device_count" -gt 0 ]; then
        cd "$DEVICE_LOG_DIR"
        ls -lt *.log 2>/dev/null | head -5 | while read -r line; do
            file=$(echo "$line" | awk '{print $9}')
            if [ -n "$file" ]; then
                device_id="${file%.log}"
                modified=$(echo "$line" | awk '{print $6, $7, $8}')
                echo "  $device_id (最后修改: $modified)"
            fi
        done
    else
        echo "  暂无设备"
    fi
}

# 主程序
case "$1" in
    "list")
        list_devices
        ;;
    "tail")
        if [ -z "$2" ]; then
            echo -e "${RED}错误: 请指定设备ID${NC}"
            echo "使用方法: $0 tail <device_id>"
            exit 1
        fi
        tail_device_log "$2"
        ;;
    "view")
        if [ -z "$2" ]; then
            echo -e "${RED}错误: 请指定设备ID${NC}"
            echo "使用方法: $0 view <device_id> [lines]"
            exit 1
        fi
        view_device_log "$2" "$3"
        ;;
    "errors")
        if [ -z "$2" ]; then
            echo -e "${RED}错误: 请指定设备ID${NC}"
            echo "使用方法: $0 errors <device_id>"
            exit 1
        fi
        view_errors "$2"
        ;;
    "connections")
        if [ -z "$2" ]; then
            echo -e "${RED}错误: 请指定设备ID${NC}"
            echo "使用方法: $0 connections <device_id>"
            exit 1
        fi
        view_connections "$2"
        ;;
    "data")
        if [ -z "$2" ]; then
            echo -e "${RED}错误: 请指定设备ID${NC}"
            echo "使用方法: $0 data <device_id>"
            exit 1
        fi
        view_data "$2"
        ;;
    "commands")
        if [ -z "$2" ]; then
            echo -e "${RED}错误: 请指定设备ID${NC}"
            echo "使用方法: $0 commands <device_id>"
            exit 1
        fi
        view_commands "$2"
        ;;
    "search")
        if [ -z "$2" ] || [ -z "$3" ]; then
            echo -e "${RED}错误: 请指定设备ID和搜索关键词${NC}"
            echo "使用方法: $0 search <device_id> <keyword>"
            exit 1
        fi
        search_device_log "$2" "$3"
        ;;
    "stats")
        show_stats
        ;;
    "help"|"-h"|"--help"|"")
        show_help
        ;;
    *)
        echo -e "${RED}错误: 未知命令 '$1'${NC}"
        echo ""
        show_help
        exit 1
        ;;
esac 