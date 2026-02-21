#!/bin/bash

# Jira Tracker 停止脚本

echo "======================================"
echo "   Jira Tracker 停止脚本"
echo "======================================"
echo ""

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# 查找并停止后端服务
stop_backend() {
    echo -e "${YELLOW}正在停止后端服务...${NC}"

    # 查找后端进程
    BACKEND_PID=$(ps aux | grep "[j]ira-tracker-backend/main" | awk '{print $2}')

    if [ -n "$BACKEND_PID" ]; then
        kill $BACKEND_PID
        echo -e "${GREEN}✓ 后端服务已停止 (PID: $BACKEND_PID)${NC}"
    else
        # 尝试通过端口查找
        if lsof -Pi :8080 -sTCP:LISTEN -t >/dev/null 2>&1; then
            PORT_PID=$(lsof -Pi :8080 -sTCP:LISTEN -t)
            kill $PORT_PID
            echo -e "${GREEN}✓ 后端服务已停止 (PID: $PORT_PID)${NC}"
        else
            echo -e "${YELLOW}✗ 后端服务未运行${NC}"
        fi
    fi
}

# 查找并停止前端服务
stop_frontend() {
    echo -e "${YELLOW}正在停止前端服务...${NC}"

    # 查找前端进程 (vite dev server)
    FRONTEND_PID=$(ps aux | grep "[v]ite.*jira-tracker-frontend" | awk '{print $2}')

    if [ -n "$FRONTEND_PID" ]; then
        kill $FRONTEND_PID
        echo -e "${GREEN}✓ 前端服务已停止 (PID: $FRONTEND_PID)${NC}"
    else
        # 尝试通过端口查找
        if lsof -Pi :5173 -sTCP:LISTEN -t >/dev/null 2>&1; then
            PORT_PID=$(lsof -Pi :5173 -sTCP:LISTEN -t)
            kill $PORT_PID
            echo -e "${GREEN}✓ 前端服务已停止 (PID: $PORT_PID)${NC}"
        else
            echo -e "${YELLOW}✗ 前端服务未运行${NC}"
        fi
    fi
}

# 停止所有服务
stop_all() {
    stop_backend
    stop_frontend

    echo ""
    echo -e "${GREEN}======================================"
    echo "   所有服务已停止"
    echo "======================================${NC}"
}

echo "======================================"
echo "   选择操作"
echo "======================================"
echo "1. 停止后端服务"
echo "2. 停止前端服务"
echo "3. 停止所有服务"
echo "4. 退出"
echo ""
read -p "请输入选项 (1-4): " choice

case $choice in
    1)
        stop_backend
        ;;
    2)
        stop_frontend
        ;;
    3)
        stop_all
        ;;
    4)
        echo "退出"
        exit 0
        ;;
    *)
        echo -e "${RED}无效的选项${NC}"
        exit 1
        ;;
esac
