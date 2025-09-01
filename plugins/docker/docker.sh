#!/bin/bash
set -eu

# ------------------ 配置变量 ------------------
CONTAINERD_VERSION="2.1.4"
ROOTLESSKIT_VERSION="2.3.5"
NERDCTL_VERSION="2.1.3"
CNI_VERSION="1.7.1"
CONTAINERD_DATA_ROOT="/data/virt-home/virt/containerd"

DOCKER_VERSION="28.3.3"
DOCKER_COMPOSE_VERSION="v2.39.2"
DOCKER_BUILDX_VERSION="v0.27.0"
DOCKER_DATA_ROOT="/data/virt-home/virt/docker"

# GitHub 镜像代理
GITHUB_MIRROR="https://gh-proxy.net/"

# 国内镜像加速
APT_MIRROR="http://mirrors.aliyun.com/ubuntu/"
DOCKER_REGISTRY_MIRRORS=( "https://docker.1ms.run" "https://docker.m.daocloud.io" )

# 工作目录
WORKDIR=~/docker_install
DOWNLOAD_DIR="${WORKDIR}/download"
LOGDIR="${WORKDIR}/log"
SHDIR="${WORKDIR}/sh"
LOGFILE="${LOGDIR}/docker_install.log"

# ------------------ 日志函数 ------------------
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info()    { echo -e "${BLUE}[INFO] $*${NC}"; }
log_success() { echo -e "${GREEN}[SUCCESS] $*${NC}"; }
log_warn()    { echo -e "${YELLOW}[WARN] $*${NC}"; }
log_error()   { echo -e "${RED}[ERROR] $*${NC}"; }

# ------------------ 系统检测函数 ------------------
check_system() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        if [[ "$ID" != "ubuntu" && "$ID" != "debian" ]]; then
            log_error "不支持的系统: $ID"
            log_error "本脚本仅支持 Ubuntu 和 Debian 系统"
            exit 1
        fi
        log_info "检测到系统: $NAME $VERSION"
    else
        log_error "无法检测系统类型"
        exit 1
    fi
}

# ------------------ 参数解析函数 ------------------
parse_arguments() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            --*=*)
                key="${1%%=*}"
                key="${key:2}"
                value="${1#*=}"
                case $key in
                    work-dir) 
                        WORKDIR="$value"
                        DOWNLOAD_DIR="${WORKDIR}/download"
                        LOGDIR="${WORKDIR}/log"
                        SHDIR="${WORKDIR}/sh"
                        LOGFILE="${LOGDIR}/docker_install.log"
                        ;;
                    containerd-version) CONTAINERD_VERSION="$value" ;;
                    rootlesskit-version) ROOTLESSKIT_VERSION="$value" ;;
                    nerdctl-version) NERDCTL_VERSION="$value" ;;
                    cni-version) CNI_VERSION="$value" ;;
                    containerd-data-root) CONTAINERD_DATA_ROOT="$value" ;;
                    docker-version) DOCKER_VERSION="$value" ;;
                    docker-compose-version) DOCKER_COMPOSE_VERSION="$value" ;;
                    docker-buildx-version) DOCKER_BUILDX_VERSION="$value" ;;
                    docker-data-root) DOCKER_DATA_ROOT="$value" ;;
                    github-mirror) GITHUB_MIRROR="$value" ;;
                    apt-mirror) APT_MIRROR="$value" ;;
                    docker-registry-mirrors)
                        IFS=',' read -ra MIRRORS <<< "$value"
                        DOCKER_REGISTRY_MIRRORS=("${MIRRORS[@]}")
                        ;;
                    *) log_warn "未知参数: $key" ;;
                esac
                shift
                ;;
            *)
                log_warn "忽略未知选项: $1"
                shift
                ;;
        esac
    done
}

download_file() {
    local url="$1"
    local filename="$2"

    local filepath="${DOWNLOAD_DIR}/${filename}"
    if [ -f "$filepath" ]; then
        log_info "已存在: $filepath, 跳过下载"
    else
        log_info "下载 $url -> $filepath"
        curl -L -o "$filepath" "$url"
    fi
}


# ------------------ 安装函数 ------------------
install_docker() {
    log_info "开始安装 containerd v$CONTAINERD_VERSION + Docker v$DOCKER_VERSION ($(date))"
    
    # 检查系统类型
    check_system

    mkdir -p "$WORKDIR" "$LOGDIR" "$SHDIR" "$DOWNLOAD_DIR"
    cd "$WORKDIR"
    
    # 日志同时输出到文件
    exec > >(tee -a "$LOGFILE") 2>&1

    # ------------------ 系统架构检测 ------------------
    ARCH=$(uname -m)
    if [ "$ARCH" = "x86_64" ]; then
        ARCH2="amd64"
    elif [ "$ARCH" = "aarch64" ]; then
        ARCH2="arm64"
    else
        log_error "不支持的架构: $ARCH"
        exit 1
    fi
    log_info "系统架构: $ARCH"

    # ------------------ 0. 替换 APT 镜像 ------------------
    log_info "替换 APT 镜像为阿里云"
    sudo tee /etc/apt/sources.list.d/ubuntu.sources <<EOF
Types: deb
URIs: $APT_MIRROR
Suites: noble noble-updates noble-backports
Components: main universe restricted multiverse
Signed-By: /usr/share/keyrings/ubuntu-archive-keyring.gpg

Types: deb
URIs: $APT_MIRROR
Suites: noble-security
Components: main universe restricted multiverse
Signed-By: /usr/share/keyrings/ubuntu-archive-keyring.gpg
EOF

    sudo apt-get update
    sudo apt-get install -y curl tar iptables socat conntrack ebtables jq file xz-utils uidmap slirp4netns ethtool
    log_success "APT 镜像替换完成并安装依赖包"

    # ------------------ 1. 安装 containerd ------------------
    log_info "下载 containerd v$CONTAINERD_VERSION"
    CONTAINERD_URL="${GITHUB_MIRROR}https://github.com/containerd/containerd/releases/download/v${CONTAINERD_VERSION}/containerd-${CONTAINERD_VERSION}-linux-${ARCH2}.tar.gz"
    download_file "$CONTAINERD_URL" "containerd-${CONTAINERD_VERSION}-linux-${ARCH2}.tar.gz"
    CONTAINERD_TGZ="${DOWNLOAD_DIR}/containerd-${CONTAINERD_VERSION}-linux-${ARCH2}.tar.gz"
    sudo tar -C /usr/local/bin --strip-components=1 -xzf "$CONTAINERD_TGZ"
    log_success "containerd 解压完成"

    # containerd systemd 服务
    log_info "创建 containerd systemd 服务文件"
    sudo tee /etc/systemd/system/containerd.service <<'EOF'
[Unit]
Description=containerd container runtime
After=network.target local-fs.target
Wants=network.target

[Service]
ExecStart=/usr/local/bin/containerd
Restart=always
RestartSec=5
Delegate=yes
KillMode=process
OOMScoreAdjust=-999
LimitNOFILE=1048576
LimitNPROC=infinity
LimitCORE=infinity
TasksMax=infinity
Slice=machine.slice
Environment=PATH=/usr/local/bin:/usr/local/sbin:/usr/sbin:/usr/bin:/sbin:/bin

[Install]
WantedBy=multi-user.target
EOF
    log_success "containerd systemd 服务文件创建完成"

    # containerd 配置文件
    log_info "生成 containerd 配置文件并修改 systemd cgroup 与 pause 镜像"
    sudo mkdir -p /etc/containerd
    containerd config default | sudo tee /etc/containerd/config.toml > /dev/null
    sudo sed -i 's/SystemdCgroup = false/SystemdCgroup = true/' /etc/containerd/config.toml
    sudo sed -i "s|^root = .*|root = \"${CONTAINERD_DATA_ROOT}\"|" /etc/containerd/config.toml

    log_success "containerd 配置文件生成完成"

    # containerd 镜像加速
    log_info "配置 containerd 镜像加速器"
    sudo mkdir -p /etc/containerd/certs.d/docker.io
    {
        echo 'server = "https://docker.io"'
        for mirror in "${DOCKER_REGISTRY_MIRRORS[@]}"; do
            echo "[host.\"$mirror\"]"
            echo '  capabilities = ["pull", "resolve"]'
        done
    } | sudo tee /etc/containerd/certs.d/docker.io/hosts.toml >/dev/null
    log_success "containerd 镜像加速配置完成"


    sudo systemctl daemon-reload
    sudo systemctl enable --now containerd
    log_success "containerd 服务启动完成"

    # ------------------ 2. 安装 rootlesskit ------------------
    log_info "下载 rootlesskit v$ROOTLESSKIT_VERSION"
    ROOTLESSKIT_URL="${GITHUB_MIRROR}https://github.com/rootless-containers/rootlesskit/releases/download/v${ROOTLESSKIT_VERSION}/rootlesskit-${ARCH}.tar.gz"
    download_file "$ROOTLESSKIT_URL" "rootlesskit-${ROOTLESSKIT_VERSION}-${ARCH}.tar.gz"
    ROOTLESSKIT_TGZ="${DOWNLOAD_DIR}/rootlesskit-${ROOTLESSKIT_VERSION}-${ARCH}.tar.gz"
    sudo tar -C /usr/local/bin -xzf "$ROOTLESSKIT_TGZ"

    log_success "rootlesskit 安装完成"
    log_info "普通用户可执行: containerd-rootless-setuptool.sh install 初始化 rootless containerd"

    # ------------------ 3. 安装 CNI ------------------
    log_info "下载 CNI v$CNI_VERSION"

    CNI_URL="${GITHUB_MIRROR}https://github.com/containernetworking/plugins/releases/download/v${CNI_VERSION}/cni-plugins-linux-${ARCH2}-v${CNI_VERSION}.tgz"
    download_file "$CNI_URL" "cni-plugins-linux-${ARCH2}-v${CNI_VERSION}.tgz"
    CNI_TGZ="${DOWNLOAD_DIR}/cni-plugins-linux-${ARCH2}-v${CNI_VERSION}.tgz"

    sudo mkdir -p /opt/cni/bin
    sudo tar -C /opt/cni/bin -xzf "$CNI_TGZ"
    log_success "CNI 安装完成"

    # ------------------ 4. 安装 nerdctl ------------------
    log_info "下载 nerdctl v$NERDCTL_VERSION"
    NERDCTL_URL="${GITHUB_MIRROR}https://github.com/containerd/nerdctl/releases/download/v${NERDCTL_VERSION}/nerdctl-${NERDCTL_VERSION}-linux-${ARCH2}.tar.gz"
    download_file "$NERDCTL_URL" "nerdctl-${NERDCTL_VERSION}-linux-${ARCH2}.tar.gz"
    NERDCTL_TGZ="${DOWNLOAD_DIR}/nerdctl-${NERDCTL_VERSION}-linux-${ARCH2}.tar.gz"

    sudo tar -C /usr/local/bin -xzf "$NERDCTL_TGZ"
    log_success "nerdctl & rootless containerd 工具安装完成"

    # ------------------ 5. 安装 Docker ------------------
    log_info "下载 Docker v$DOCKER_VERSION"
    DOCKER_URL="https://mirrors.aliyun.com/docker-ce/linux/static/stable/${ARCH}/docker-${DOCKER_VERSION}.tgz"
    download_file "$DOCKER_URL" "docker-${DOCKER_VERSION}-linux-${ARCH}.tgz"
    DOCKER_TGZ="${DOWNLOAD_DIR}/docker-${DOCKER_VERSION}-linux-${ARCH}.tgz"
    sudo tar -C /usr/local/bin --strip-components=1 \
        --exclude='containerd' \
        --exclude='containerd-shim-runc-v2' \
        --exclude='ctr' \
        -xzf "$DOCKER_TGZ"
    log_success "Docker 安装完成"

    # ------------------ 6. 安装 Docker Compose & Buildx ------------------
    log_info "下载 Docker Compose & Buildx"
    COMPOSE_URL="${GITHUB_MIRROR}https://github.com/docker/compose/releases/download/$DOCKER_COMPOSE_VERSION/docker-compose-linux-$ARCH"
    BUILDX_URL="${GITHUB_MIRROR}https://github.com/docker/buildx/releases/download/v${DOCKER_BUILDX_VERSION}/buildx-v${DOCKER_BUILDX_VERSION}.linux-$ARCH2"

    download_file "$COMPOSE_URL" "docker-compose-linux-$ARCH-${DOCKER_COMPOSE_VERSION}"
    COMPOSE_BIN="${DOWNLOAD_DIR}/docker-compose-linux-${ARCH}-${DOCKER_COMPOSE_VERSION}"
    download_file "$BUILDX_URL" "docker-buildx-${DOCKER_BUILDX_VERSION}.linux-$ARCH2"
    BUILDX_BIN="${DOWNLOAD_DIR}/docker-buildx-${DOCKER_BUILDX_VERSION}.linux-$ARCH2"

    sudo chmod +x "$COMPOSE_BIN" "$BUILDX_BIN"
    sudo mkdir -p /usr/lib/docker/cli-plugins

    # 使用 cp 而不是 mv，保留下载目录文件
    sudo cp "$COMPOSE_BIN" /usr/lib/docker/cli-plugins/docker-compose
    sudo cp "$BUILDX_BIN" /usr/lib/docker/cli-plugins/docker-buildx


    log_success "Docker Compose & Buildx 安装完成"

    # Docker 配置文件
    log_info "生成 Docker 配置文件并设置 registry 镜像加速"
    sudo mkdir -p /etc/docker
    {
        echo '{'
        echo '  "registry-mirrors": ['
        for i in "${!DOCKER_REGISTRY_MIRRORS[@]}"; do
            mirror="${DOCKER_REGISTRY_MIRRORS[$i]}"
            if [ "$i" -lt "$(( ${#DOCKER_REGISTRY_MIRRORS[@]} - 1 ))" ]; then
                echo "    \"$mirror\","
            else
                echo "    \"$mirror\""
            fi
        done
        echo '  ],'
        echo "  \"data-root\": \"${DOCKER_DATA_ROOT}\""
        echo '}'
    } | sudo tee /etc/docker/daemon.json >/dev/null
    log_success "Docker 配置文件生成完成"


    # Docker systemd 服务
    log_info "创建 Docker systemd 服务文件"
    sudo tee /etc/systemd/system/docker.service <<'EOF'
[Unit]
Description=Docker Application Container Engine
After=network.target containerd.service
Requires=containerd.service

[Service]
ExecStart=/usr/local/bin/dockerd --containerd=/run/containerd/containerd.sock
ExecReload=/bin/kill -s HUP $MAINPID
Restart=always
RestartSec=5
LimitNOFILE=1048576
LimitNPROC=infinity
TasksMax=infinity
Delegate=yes
Environment=PATH=/usr/local/bin:/usr/local/sbin:/usr/sbin:/usr/bin:/sbin:/bin

[Install]
WantedBy=multi-user.target
EOF
    log_success "Docker systemd 服务文件创建完成"

    # 重载 systemd 并启动服务
    log_info "重载 systemd 并启动 Docker 服务"
    sudo systemctl daemon-reload
    sudo systemctl enable --now docker
    log_success "Docker 服务启动完成"
}

# ------------------ 卸载函数 ------------------
uninstall_docker() {
    log_info "开始卸载 Docker 和 Containerd"
    
    # 检查系统类型
    check_system
    
    # 停止服务
    sudo systemctl stop docker containerd || true
    sudo systemctl disable docker containerd || true
    
    # 删除服务文件
    sudo rm -f /etc/systemd/system/docker.service /etc/systemd/system/containerd.service
    sudo systemctl daemon-reload
    
    # 删除二进制文件
    sudo rm -f /usr/local/bin/containerd \
        /usr/local/bin/containerd-shim \
        /usr/local/bin/containerd-shim-runc-v2 \
        /usr/local/bin/ctr \
        /usr/local/bin/docker \
        /usr/local/bin/dockerd \
        /usr/local/bin/docker-init \
        /usr/local/bin/docker-proxy \
        /usr/local/bin/rootlesskit \
        /usr/local/bin/rootlesskit-docker-proxy \
        /usr/local/bin/nerdctl
    
    # 删除 CLI 插件
    sudo rm -rf /usr/lib/docker/cli-plugins
    
    # 删除配置文件
    sudo rm -rf /etc/containerd /etc/docker
    
    # 删除 CNI 插件
    sudo rm -rf /opt/cni
    
    # 删除工作目录
    rm -rf "$WORKDIR"
    
    log_success "Docker 和 Containerd 卸载完成"
}

# ------------------ 帮助函数 ------------------
show_help() {
    cat << EOF
使用方式: $0 [command] [options]

命令:
  install     安装 Docker 和 Containerd
  uninstall   卸载 Docker 和 Containerd
  help        显示帮助信息

选项:
  --work-dir                           指定插件工作目录
  --containerd-version=<version>       指定 containerd 版本
  --rootlesskit-version=<version>      指定 rootlesskit 版本
  --nerdctl-version=<version>          指定 nerdctl 版本
  --cni-version=<version>              指定 CNI 版本
  --containerd-data-root=<value>       指定 containerd 数据目录
  --docker-version=<version>           指定 Docker 版本
  --docker-compose-version=<version>   指定 Docker Compose 版本
  --docker-buildx-version=<version>    指定 Docker Buildx 版本
  --docker-data-root=<value>           指定 Docker 数据目录
  --github-mirror=<url>                指定 GitHub 镜像代理
  --apt-mirror=<url>                   指定 APT 镜像源
  --docker-registry-mirrors=<json>     指定 Docker 镜像加速器

示例:
  $0 install --docker-version=28.3.3 --containerd-version=2.1.4
  $0 uninstall
EOF
}

# ------------------ 主程序 ------------------
main() {
    local command="$1"
    shift
    
    case "$command" in
        install)
            parse_arguments "$@"
            install_docker
            ;;
        uninstall)
            uninstall_docker
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            log_error "未知命令: $command"
            show_help
            exit 1
            ;;
    esac
}

# 检查是否传递了参数
if [ $# -eq 0 ]; then
    show_help
    exit 1
fi

# 执行主程序
main "$@"