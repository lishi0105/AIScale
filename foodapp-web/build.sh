#!/bin/bash

# 默认不重建依赖
REBUILD=false

# 解析参数
case "$1" in
  -r|--rebuild)
    REBUILD=true
    ;;
  "")
    # 无参数，保持默认（只构建）
    ;;
  *)
    echo "未知参数: $1"
    echo "用法: $0 [-r|--rebuild]"
    exit 1
    ;;
esac

# 如果需要重建，则更新依赖并强制安装
if [ "$REBUILD" = true ]; then
  echo "🔄 执行依赖更新与完整构建..."
  pnpm add element-plus@latest
  npm config set registry https://registry.npmmirror.com/
  npm install --force
  npm run build
else
  echo "🚀 仅执行构建..."
  npm run build
fi

# 无论是否 rebuild，都复制构建产物
echo "📦 部署前端资源到后端目录..."
mkdir -p /home/lishi/FoodInspection/foodapp/web/
rm -rf /home/lishi/FoodInspection/foodapp/web/*
cp -rf dist/* /home/lishi/FoodInspection/foodapp/web/
echo "✅ 部署完成！"