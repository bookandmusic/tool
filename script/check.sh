#!/bin/bash
set -e

# ----------------- 检查工具 -----------------
command -v goimports-reviser >/dev/null 2>&1 || {
    echo "安装 goimports-reviser..."
    go install github.com/incu6us/goimports-reviser/v3@latest
}

command -v golangci-lint >/dev/null 2>&1 || {
    echo "安装 golangci-lint..."
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
}

command -v gosec >/dev/null 2>&1 || {
    echo "安装 gosec..."
    go install github.com/securego/gosec/v2/cmd/gosec@latest
}

command -v gocyclo >/dev/null 2>&1 || {
    echo "安装 gocyclo..."
    go install github.com/fzipp/gocyclo/cmd/gocyclo@latest
}

# ----------------- 格式化 -----------------
echo "正在格式化代码..."
goimports-reviser -rm-unused -set-alias -format ./...

# ----------------- 静态检查 -----------------
echo "正在运行 golangci-lint..."
golangci-lint run ./...

# ----------------- 复杂度检查 -----------------
echo "正在检查函数复杂度（阈值 10）..."
gocyclo -over 10 .

# ----------------- 安全漏洞扫描 -----------------
echo "正在进行 gosec 安全扫描..."
gosec ./...

echo "完成！"
