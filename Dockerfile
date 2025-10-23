# syntax=docker/dockerfile:1

########################################
# Stage 0: build 
########################################
FROM ubuntu:22.04 AS build

ENV DEBIAN_FRONTEND=noninteractive

# 先装 CA，再切清华镜像
RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates && \
    rm -rf /var/lib/apt/lists/*
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

# ===== 安装 Node.js 20 + pnpm =====
RUN apt-get update && \
    curl -fsSL https://deb.nodesource.com/setup_20.x | bash - && \
    apt-get install -y --no-install-recommends nodejs && \
    rm -rf /var/lib/apt/lists/*
RUN corepack enable && corepack prepare pnpm@latest --activate
RUN npm config set registry https://registry.npmmirror.com

# 工作目录
WORKDIR /workspace/foodapp

# 复制 Go 源码
COPY pkg /workspace/foodapp/pkg
COPY configs /workspace/foodapp/configs
COPY cmd /workspace/foodapp/cmd
COPY internal /workspace/foodapp/internal
COPY go.mod go.sum /workspace/foodapp/
COPY build.sh /workspace/foodapp/
# 运行构建脚本并保证可执行
RUN chmod +x build.sh
RUN ./build.sh

# 复制构建产物
RUN mkdir -p /out/app && cp -r /workspace/foodapp/foodapp /out/app/

#  ===== 构建 Web 应用 ======
WORKDIR /workspace/foodapp-web
# 只复制 package.json + package-lock.json，先安装依赖（用缓存加速）
COPY foodapp-web/package.json /workspace/foodapp-web/package.json
COPY foodapp-web/package-lock.json /workspace/foodapp-web/package-lock.json
RUN --mount=type=cache,target=/root/.npm npm ci

# 再复制其余“干净源码”
COPY foodapp-web/vite.config.ts /workspace/foodapp-web/vite.config.ts
COPY foodapp-web/tsconfig.json /workspace/foodapp-web/tsconfig.json
COPY foodapp-web/tsconfig.app.json /workspace/foodapp-web/tsconfig.app.json
COPY foodapp-web/tsconfig.node.json /workspace/foodapp-web/tsconfig.node.json
COPY foodapp-web/index.html /workspace/foodapp-web/index.html
COPY foodapp-web/public /workspace/foodapp-web/public
COPY foodapp-web/src /workspace/foodapp-web/src

RUN npm run build
# 把前端产物单独打包到 /out/static
RUN mkdir -p /out/static && cp -r /workspace/foodapp-web/dist/* /out/static/


########################################
# Stage 1: runtime（精简运行环境）
########################################
FROM debian:12-slim AS runtime

ENV DEBIAN_FRONTEND=noninteractive
# 时区 & CA 证书（避免 HTTPS 问题）
RUN apt-get update && apt-get install -y --no-install-recommends \
      ca-certificates tzdata && \
    rm -rf /var/lib/apt/lists/*
ENV TZ=Asia/Shanghai

# 放置可执行文件与静态资源
WORKDIR /app
COPY --from=build /out/app/foodapp /app/foodapp
COPY --from=build /out/static /app/web
COPY .env* /app/
RUN chmod +x /app/foodapp

ENTRYPOINT ["/app/foodapp"]
