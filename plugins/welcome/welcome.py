#!/usr/bin/env python3
import argparse
import platform
import os

parser = argparse.ArgumentParser()
parser.add_argument("--lang", default="cn")
parser.add_argument("--name", default="User")
args = parser.parse_args()

# 标签映射
labels = {
    "en": ["System", "Version", "Arch", "User"],
    "cn": ["系统", "版本", "架构", "用户"],
}

# 选择语言（非 en 一律用 cn）
lang = "en" if args.lang.lower() == "en" else "cn"

# 系统信息拼接
sys_info = f"""{labels[lang][0]}: {platform.system()} {platform.release()}
{labels[lang][1]}: {platform.version()}
{labels[lang][2]}: {platform.machine()}
{labels[lang][3]}: {os.getlogin()}"""

# 欢迎信息
welcome_map = {
    "en": f"Welcome to Tool, {args.name}!\nMake your development smoother.",
    "cn": f"欢迎使用开发工具 Tool\n让你的开发更顺心, {args.name}!"
}

ascii_art = r"""
 _______           _       
|__   __|         | |      
   | | ___   ___  | |_ ___ 
   | |/ _ \ / _ \ | __/ _ \
   | | (_) |  __/ | ||  __/
   |_|\___/ \___|  \__\___|
"""

print(sys_info)
print(ascii_art)
print(welcome_map[lang])
