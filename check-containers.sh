#!/bin/bash
# 检查实际的容器名称

echo "=== 检查Docker容器 ==="
docker ps --format "table {{.Names}}\t{{.Status}}"
