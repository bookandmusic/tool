#!/bin/bash

VERSION="Tool v0.1.0"

# 判断是否传入 --verbose
VERBOSE=false
if [ "$1" = "--verbose" ]; then
    VERBOSE=true
fi

# 默认显示版本
echo "$VERSION"

# verbose 模式显示更多工具信息
if [ "$VERBOSE" = true ]; then
    echo
    echo "工具功能模块:"
    echo " - 日志系统: 支持 debug/info/error 输出"
    echo " - 配置管理: YAML 配置文件支持"
    echo " - 插件管理: 支持动态插件(shell/python)"
    echo " - 软件管理: 支持启用/禁用/安装/卸载/批量操作"
fi
