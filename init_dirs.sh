#!/usr/bin/env bash
set -e

# 创建顶层目录
mkdir -p cmd/foodapp

# 内部逻辑层
mkdir -p internal/app                # 程序装配与启动逻辑
mkdir -p internal/domain/account     # 领域模型
mkdir -p internal/domain/food
mkdir -p internal/repository/account # 数据访问层（数据库操作）
mkdir -p internal/repository/food
mkdir -p internal/service/account    # 业务服务层
mkdir -p internal/service/food
mkdir -p internal/server/handler     # HTTP handler（Gin 控制器）
mkdir -p internal/server/middleware  # Gin 中间件
mkdir -p internal/storage/db         # 数据库初始化/迁移

# 可复用包
mkdir -p pkg/logger

# 配置与脚本
mkdir -p configs scripts

echo "✅ 目录结构已创建完成。"
