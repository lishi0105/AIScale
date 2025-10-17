# syntax=docker/dockerfile:1
FROM ubuntu:22.04

ENV DEBIAN_FRONTEND=noninteractive

# 先装 CA，再切换到清华镜像
RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates && \
    rm -rf /var/lib/apt/lists/*

# 切换到清华镜像（含 security）
RUN sed -i 's|http://\(.*\).ubuntu.com|https://mirrors.tuna.tsinghua.edu.cn|g' /etc/apt/sources.list && \
    sed -i 's|http://security.ubuntu.com|https://mirrors.tuna.tsinghua.edu.cn/ubuntu|g' /etc/apt/sources.list

# 常用构建工具 & 依赖
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
      build-essential ninja-build cmake make \
      wget curl git unzip vim \
      python3 python3-pip \
      automake autoconf libtool m4 pkg-config \
      gnupg clang-format libssl-dev sudo && \
    rm -rf /var/lib/apt/lists/*

# pip 国内镜像
RUN mkdir -p /root/.pip && \
    printf "[global]\nindex-url = https://pypi.tuna.tsinghua.edu.cn/simple\n" > /root/.pip/pip.conf

# ===== 安装 Go 1.24.9 (amd64 固定) =====
ENV GO_VERSION=1.24.9
COPY go1.24.9.linux-amd64.tar.gz /tmp/go.tgz
RUN rm -rf /usr/local/go && \
    tar -C /usr/local -xzf /tmp/go.tgz && \
    rm -f /tmp/go.tgz

# Go 环境变量（含国内代理、自动工具链）
ENV PATH="/usr/local/go/bin:${PATH}" \
    GOPATH="/go" \
    GOCACHE="/go/.cache" \
    GOPROXY="https://goproxy.cn,direct" \
    GO111MODULE="on" \
    GOTOOLCHAIN="auto"

# ===== 安装 Node.js 18 LTS =====
RUN apt-get update && \
    curl -fsSL https://deb.nodesource.com/setup_20.x | bash - && \
    apt-get install -y --no-install-recommends nodejs && \
    rm -rf /var/lib/apt/lists/*
RUN corepack enable && \
    corepack prepare pnpm@latest --activate
# npm 国内镜像（可按需修改/删除）
RUN npm config set registry https://registry.npmmirror.com

# 工作目录
RUN mkdir -p /workspace "${GOPATH}" "${GOCACHE}"
WORKDIR /workspace

# 版本自检
RUN bash -lc "go version && node -v && npm -v"

# 默认进入 bash
CMD ["/bin/bash"]
