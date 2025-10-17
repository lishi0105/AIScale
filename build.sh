#!/bin/bash
set -e  # 任一命令失败则退出

# 解析命令行参数
REBUILD=false
while [[ "$#" -gt 0 ]]; do
    case $1 in
        -r|--rebuild) REBUILD=true ;;
        *) echo "未知参数: $1"; exit 1 ;;
    esac
    shift
done

# 设置环境变量
export GOPROXY=https://goproxy.cn,direct
export GOWORK=off

GO_MOD="go.mod"
GO_SUM="go.sum"

if [[ "$REBUILD" == true ]]; then
    echo "🔄 启用重建模式：删除现有模块文件..."
    [[ -f "$GO_MOD" ]] && rm "$GO_MOD" && echo "🗑️ 已删除 $GO_MOD"
    [[ -f "$GO_SUM" ]] && rm "$GO_SUM" && echo "🗑️ 已删除 $GO_SUM"

    echo "📦 初始化模块..."
    go mod init hdzk.cn/foodapp
    go mod tidy
else
    if [[ ! -f "$GO_MOD" ]] || [[ ! -f "$GO_SUM" ]]; then
        echo "⚠️ 检测到模块文件缺失，正在初始化..."
        go mod init hdzk.cn/foodapp
        go mod tidy
    else
        echo "✅ 检测到 go.mod 和 go.sum，跳过模块初始化"
    fi
fi

echo "🔨 开始构建..."
go build -buildvcs=false -o foodapp ./cmd/foodapp

echo "✅ 编译成功！可执行文件: ./foodapp"