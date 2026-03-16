#!/bin/bash

# Jira Tracker 启动脚本

echo "======================================"
echo "   Jira Tracker 启动脚本"
echo "======================================"
echo ""

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 检查是否安装了必要的工具
check_command() {
    # 使用type命令检测，不输出路径
    if type "$1" >/dev/null 2>&1; then
        echo -e "${GREEN}✓ $1 已安装${NC}"
        return 0
    fi

    echo -e "${RED}错误: 未找到 $1，请先安装${NC}"
    return 1
}

# 检查服务是否运行
check_service() {
    local service=$1
    local port=$2

    # 方法1: 使用nc命令检测端口
    if type nc >/dev/null 2>&1; then
        if nc -z localhost "$port" 2>/dev/null; then
            echo -e "${GREEN}✓ $service 已在端口 $port 运行${NC}"
            return 0
        fi
    fi

    # 方法2: 使用lsof检测端口（需要root权限）
    if type lsof >/dev/null 2>&1; then
        if lsof -Pi :"$port" -sTCP:LISTEN -t >/dev/null 2>&1 || \
           lsof -i :"$port" -sTCP:LISTEN -t >/dev/null 2>&1; then
            echo -e "${GREEN}✓ $service 已在端口 $port 运行${NC}"
            return 0
        fi
    fi

    # 方法3: 使用netstat检测端口
    if type netstat >/dev/null 2>&1; then
        if netstat -an 2>/dev/null | grep -q "LISTEN.*:$port "; then
            echo -e "${GREEN}✓ $service 已在端口 $port 运行${NC}"
            return 0
        fi
    fi

    echo -e "${YELLOW}✗ $service 未在端口 $port 运行${NC}"
    return 1
}

echo "检查必要的工具..."
check_command go
check_command node
check_command npm

echo ""
echo "检查服务状态..."
check_service "MySQL" 3306
check_service "Redis" 6379

echo ""
echo "======================================"
echo "   选择操作"
echo "======================================"
echo "1. 初始化数据库"
echo "2. 启动后端服务"
echo "3. 启动前端服务"
echo "4. 同时启动前后端服务"
echo "5. 退出"
echo ""
read -p "请输入选项 (1-5): " choice

case $choice in
    1)
        echo ""
        echo -e "${YELLOW}正在初始化数据库...${NC}"
        cd jira-tracker-backend
        go run scripts/init_data.go
        echo -e "${GREEN}数据库初始化完成！${NC}"
        ;;
    2)
        echo ""
        echo -e "${YELLOW}正在启动后端服务...${NC}"
        cd jira-tracker-backend
        go run main.go
        ;;
    3)
        echo ""
        echo -e "${YELLOW}正在启动前端服务...${NC}"
        cd jira-tracker-frontend
        npm install
        npm run dev
        ;;
    4)
        echo ""
        echo -e "${YELLOW}正在启动前后端服务...${NC}"
        
        # 获取后端端口配置
        BACKEND_PORT=$(cd jira-tracker-backend && grep -E "^\s*port:" config/config.yaml | head -1 | awk '{print $2}')
        BACKEND_PORT=${BACKEND_PORT:-8080}
        
        # 启动后端
        cd jira-tracker-backend
        go run main.go > /tmp/jira-backend.log 2>&1 &
        BACKEND_PID=$!
        
        # 等待后端启动并检查是否成功
        echo -e "${YELLOW}等待后端服务启动...${NC}"
        MAX_WAIT=30
        WAIT_COUNT=0
        BACKEND_READY=false
        
        while [ $WAIT_COUNT -lt $MAX_WAIT ]; do
            # 检查进程是否还在运行
            if ! kill -0 $BACKEND_PID 2>/dev/null; then
                echo -e "${RED}后端服务启动失败！${NC}"
                echo -e "${YELLOW}查看日志: tail -f /tmp/jira-backend.log${NC}"
                exit 1
            fi
            
            # 检查端口是否被监听
            if nc -z localhost $BACKEND_PORT 2>/dev/null; then
                BACKEND_READY=true
                break
            fi
            
            sleep 1
            WAIT_COUNT=$((WAIT_COUNT + 1))
            echo -n "."
        done
        echo ""
        
        if [ "$BACKEND_READY" = false ]; then
            echo -e "${RED}后端服务启动超时（${MAX_WAIT}秒）！${NC}"
            echo -e "${YELLOW}查看日志: tail -f /tmp/jira-backend.log${NC}"
            kill $BACKEND_PID 2>/dev/null
            exit 1
        fi
        
        echo -e "${GREEN}✓ 后端服务启动成功 (PID: $BACKEND_PID)${NC}"
        
        # 启动前端
        cd ../jira-tracker-frontend
        npm install
        npm run dev > /tmp/jira-frontend.log 2>&1 &
        FRONTEND_PID=$!
        
        echo ""
        echo -e "${GREEN}服务已启动！${NC}"
        echo -e "后端地址: ${GREEN}http://localhost:$BACKEND_PORT${NC}"
        echo -e "前端地址: ${GREEN}http://localhost:3000${NC}"
        echo -e "后端日志: ${YELLOW}/tmp/jira-backend.log${NC}"
        echo -e "前端日志: ${YELLOW}/tmp/jira-frontend.log${NC}"
        echo ""
        echo "按 Ctrl+C 停止服务"
        
        # 等待中断信号
        trap "echo ''; echo -e '${YELLOW}正在停止服务...${NC}'; kill $BACKEND_PID $FRONTEND_PID 2>/dev/null; exit" INT TERM
        wait
        ;;
    5)
        echo "退出"
        exit 0
        ;;
    *)
        echo -e "${RED}无效的选项${NC}"
        exit 1
        ;;
esac
