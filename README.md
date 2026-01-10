# Update ESA Origin Rule

[![Go Version](https://img.shields.io/badge/Go-1.20+-00ADD8?style=flat&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

一个用于更新阿里云 ESA（边缘安全加速）回源规则配置的命令行工具。

## 📋 目录

- [项目简介](#项目简介)
- [功能特性](#功能特性)
- [快速开始](#快速开始)
  - [下载安装](#下载安装)
  - [配置凭证](#配置凭证)
- [使用指南](#使用指南)
  - [查询规则列表](#查询规则列表)
  - [更新回源规则](#更新回源规则)
  - [更新重定向规则](#更新重定向规则)
- [参数说明](#参数说明)
- [使用场景](#使用场景)
- [常见问题](#常见问题)
- [本地开发](#本地开发)
- [许可证](#许可证)

## 项目简介

本工具最初使用 Python 开发，为了能在嵌入式设备（如 OpenWrt 路由器、树莓派等）上更方便地运行，已迁移至 Go 语言实现。Go 版本可以编译为单个静态二进制文件，无需安装运行时环境，特别适合资源受限的设备。

**适用场景：**
- 🏠 家庭网络环境中动态更新 CDN 回源端口
- 🖥️ 嵌入式设备（OpenWrt、树莓派）上自动化配置
- 🔄 需要频繁切换回源配置的场景
- 📦 CI/CD 流程中的自动化部署

## 功能特性

- ✅ 修改 ESA 站点的回源协议（HTTP/HTTPS/Follow）和端口
- ✅ 修改 ESA 站点的重定向规则目标端口（支持动态拼接 URL）
- ✅ 支持按「规则名称」或「配置 ID (ConfigId)」精准定位规则
- ✅ 列出当前站点的所有规则及配置信息
- ✅ 提供多平台静态编译二进制文件（Windows, Linux amd64/arm64）
- ✅ 零依赖，单文件运行

## 快速开始

### 下载安装

#### 方式一：从 GitHub Actions 下载

前往 [GitHub Actions](../../actions) 的 Artifacts 下载对应架构的二进制文件：

| 平台 | 文件名 | 适用设备 |
|------|--------|----------|
| Linux (x86_64) | `update_esa-linux-amd64` | 普通 Linux PC、服务器 |
| Linux (ARM64) | `update_esa-linux-arm64` | 树莓派、OpenWrt (ARM64)、ARM 服务器 |
| Windows (x86_64) | `update_esa-windows-amd64.exe` | Windows PC |

#### 方式二：本地编译

```bash
# 克隆仓库
git clone <repository-url>
cd update_esa

# 编译
go mod tidy
go build -o update_esa main.go
```

#### 安装到系统

**Linux / macOS / OpenWrt:**
```bash
# 赋予执行权限
chmod +x update_esa-linux-arm64

# 移动到系统路径（可选）
sudo mv update_esa-linux-arm64 /usr/local/bin/update_esa

# 验证安装
update_esa --help
```

**Windows:**
```powershell
# 将 exe 文件移动到合适的位置，并添加到 PATH 环境变量
```

### 配置凭证

你需要阿里云的 **AccessKey ID** 和 **AccessKey Secret**，并确保该账号拥有 ESA 的相关权限：
- `esa:ListOriginRules`
- `esa:UpdateOriginRule`
- `esa:ListRedirectRules`
- `esa:UpdateRedirectRule`

> [!IMPORTANT]
> 推荐使用 RAM 用户并授予最小权限，避免使用主账号 AccessKey。

**推荐方式：使用环境变量**

Linux / macOS / OpenWrt:
```bash
export ALIBABA_CLOUD_ACCESS_KEY_ID="你的AccessKeyId"
export ALIBABA_CLOUD_ACCESS_KEY_SECRET="你的AccessKeySecret"
```

Windows PowerShell:
```powershell
$env:ALIBABA_CLOUD_ACCESS_KEY_ID = "你的AccessKeyId"
$env:ALIBABA_CLOUD_ACCESS_KEY_SECRET = "你的AccessKeySecret"
```

> [!WARNING]
> 不推荐通过命令行参数 `--access-key-id` 和 `--access-key-secret` 传入凭证，容易在命令历史中泄露。

## 使用指南

### 查询规则列表

查看站点下所有回源规则和重定向规则的详细信息：

```bash
./update_esa --region-id cn-hangzhou --site-id 123456789 --list
```

**输出示例：**
```
=== 回源规则 (Origin Rules) ===
ID: 1001    Name: default-origin    Scheme: follow    HTTP: 80    HTTPS: 443
ID: 1002    Name: api-origin        Scheme: https     HTTP:       HTTPS: 8443

=== 重定向规则 (Redirect Rules) ===
ID: 2001    Name: redirect-rule-1   Type: 301    Target: https://example.com:8080
```

### 更新回源规则

#### 方式一：按 ConfigId 更新（推荐）

ConfigId 是唯一标识，更精准且不会误匹配：

```bash
./update_esa --region-id cn-hangzhou --site-id 123456789 \
  --config-id 1001 \
  --origin-scheme https --https-port 8443
```

#### 方式二：按规则名称更新

规则名称匹配不区分大小写：

```bash
./update_esa --region-id cn-hangzhou --site-id 123456789 \
  --rule-name "default-origin" \
  --origin-scheme follow --http-port 80 --https-port 443
```

#### 仅更新协议

```bash
./update_esa --region-id cn-hangzhou --site-id 123456789 \
  --config-id 1001 \
  --origin-scheme http
```

### 更新重定向规则

修改重定向规则的目标端口（自动更新 URL 中的端口号）：

```bash
./update_esa --region-id cn-hangzhou --site-id 123456789 \
  --rule-name "redirect-rule-1" \
  --redirect-port 9090
```

**处理逻辑：**
- 如果目标 URL 已包含端口（如 `https://example.com:8080`），则替换为新端口
- 如果目标 URL 不含端口（如 `https://example.com`），则添加端口

## 参数说明

| 参数 | 必填 | 类型 | 说明 | 示例 |
|------|------|------|------|------|
| `--region-id` | ✅ | string | ESA 站点所属区域 | `cn-hangzhou` |
| `--site-id` | ✅ | int64 | 站点 ID | `123456789` |
| `--config-id` | ❌ | int64 | 规则配置 ID，优先使用此参数定位 | `1001` |
| `--rule-name` | ❌ | string | 规则名称，不区分大小写 | `"default-origin"` |
| `--origin-scheme` | ❌ | string | 回源协议：`http`、`https`、`follow` | `https` |
| `--http-port` | ❌ | int | HTTP 回源端口 | `80` |
| `--https-port` | ❌ | int | HTTPS 回源端口 | `443` |
| `--redirect-port` | ❌ | int | 重定向目标端口 | `8080` |
| `--access-key-id` | ❌ | string | AccessKey ID（不推荐） | - |
| `--access-key-secret` | ❌ | string | AccessKey Secret（不推荐） | - |
| `--list` | ❌ | bool | 列出所有规则，不执行更新 | - |

> [!NOTE]
> - 必须提供 `--config-id` 或 `--rule-name` 之一来定位规则
> - 使用 `--origin-scheme` 时，根据协议类型需要设置对应的端口参数
> - `--redirect-port` 仅用于更新重定向规则

## 使用场景

### 场景 1：家庭动态端口映射

在家庭网络中，内网服务端口可能会变化，需要自动更新 CDN 回源配置：

```bash
#!/bin/bash
# 脚本：update_cdn_port.sh

NEW_PORT=$(cat /tmp/current_port)  # 从文件读取当前端口

./update_esa --region-id cn-hangzhou --site-id 123456789 \
  --rule-name "home-server" \
  --origin-scheme https --https-port $NEW_PORT

echo "已更新回源端口为: $NEW_PORT"
```

### 场景 2：OpenWrt 定时任务

在 OpenWrt 路由器上配置定时任务，每小时检查并更新：

```bash
# 添加到 crontab
0 * * * * /usr/local/bin/update_esa --region-id cn-hangzhou --site-id 123456789 --rule-name "router-service" --origin-scheme https --https-port 8443
```

### 场景 3：CI/CD 自动部署

在部署流程中自动更新 CDN 配置：

```yaml
# .github/workflows/deploy.yml
- name: Update ESA Origin Rule
  env:
    ALIBABA_CLOUD_ACCESS_KEY_ID: ${{ secrets.ALIYUN_AK_ID }}
    ALIBABA_CLOUD_ACCESS_KEY_SECRET: ${{ secrets.ALIYUN_AK_SECRET }}
  run: |
    ./update_esa --region-id cn-hangzhou --site-id 123456789 \
      --config-id 1001 \
      --origin-scheme https --https-port 443
```

## 常见问题

### Q1: 提示"未找到指定回源规则名称"怎么办？

**原因：** 规则名称不匹配或站点下没有该规则。

**解决方法：**
1. 使用 `--list` 参数查看所有可用规则名称
2. 检查规则名称拼写是否正确
3. 尝试使用 `--config-id` 参数精准定位

### Q2: 如何获取站点 ID (site-id)？

登录阿里云 ESA 控制台，在站点列表中可以看到每个站点的 ID。

### Q3: 权限不足错误怎么处理？

**错误示例：** `User not authorized to operate on the specified resource`

**解决方法：**
1. 确认 RAM 用户已授予 ESA 相关权限
2. 检查 AccessKey 是否正确
3. 验证 `--region-id` 和 `--site-id` 是否匹配

### Q4: 可以同时更新多个规则吗？

当前版本不支持批量更新，需要多次调用命令。可以编写 Shell 脚本循环处理：

```bash
for rule_id in 1001 1002 1003; do
  ./update_esa --region-id cn-hangzhou --site-id 123456789 \
    --config-id $rule_id \
    --origin-scheme https --https-port 443
done
```

### Q5: 在 OpenWrt 上运行提示"Permission denied"？

需要赋予执行权限：
```bash
chmod +x update_esa-linux-arm64
```

## 本地开发

### 环境要求

- Go 1.20 或更高版本
- Git

### 编译说明

```bash
# 克隆仓库
git clone <repository-url>
cd update_esa

# 下载依赖
go mod tidy

# 本地编译
go build -o update_esa main.go

# 交叉编译（示例：编译 ARM64 版本）
GOOS=linux GOARCH=arm64 go build -o update_esa-linux-arm64 main.go

# 编译 Windows 版本
GOOS=windows GOARCH=amd64 go build -o update_esa-windows-amd64.exe main.go
```

### 项目结构

```
.
├── main.go              # 主程序入口
├── go.mod               # Go 模块依赖
├── README.md            # 项目文档
├── .github/
│   └── workflows/
│       └── build.yml    # GitHub Actions 构建配置
└── requirements.txt     # Python 版本依赖（已废弃）
```

## 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件
