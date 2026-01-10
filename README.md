# Update ESA Origin Rule

这是一个用于更新阿里云 ESA（边缘安全加速）回源规则配置的工具。已从 Python 迁移至 Go，以支持编译为单文件在嵌入式设备（OpenWrt 等）上运行。

## 功能特性

*   修改 ESA 站点的回源协议（HTTP/HTTPS/Follow）和端口。
*   修改 ESA 站点的重定向规则目标端口（支持动态拼接 URL）。
*   支持按「规则名称」或「配置 ID (ConfigId)」定位规则。
*   支持列出当前站点的所有规则及配置。
*   提供 Windows, Linux (amd64, arm64) 的静态编译二进制文件。

## 使用方法

### 1. 准备工作

你需要阿里云的 AccessKey ID 和 AccessKey Secret，并确保该账号拥有 ESA 的相关权限（如 `esa:ListOriginRules`, `esa:UpdateOriginRule`）。

**推荐使用环境变量设置凭证：**

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

也可以通过命令行参数 `--access-key-id` 和 `--access-key-secret` 传入（不推荐，易泄露）。

### 2. 获取二进制文件

前往 GitHub Actions 的 Artifacts 下载对应架构的二进制文件：
*   `update_esa-linux-amd64`: 普通 Linux PC 或服务器
*   `update_esa-linux-arm64`: 树莓派、OpenWrt (ARM64)
*   `update_esa-windows-amd64.exe`: Windows PC

上传到设备并赋予执行权限（Linux/macOS）：
```bash
chmod +x update_esa-linux-arm64
```

### 3. 常用命令示例

#### 查询规则列表
查看站点下所有规则的 ConfigId、名称和当前端口：
```bash
./update_esa-linux-arm64 --region-id cn-hangzhou --site-id <站点ID> --list
```

#### 更新回源规则
**方式一：按 ConfigId 更新（推荐，更精准）**
```bash
./update_esa-linux-arm64 --region-id cn-hangzhou --site-id <站点ID> \
  --config-id <ConfigID> \
  --origin-scheme https --https-port 1235
```

**方式二：按规则名称更新**
```bash
./update_esa-linux-arm64 --region-id cn-hangzhou --site-id <站点ID> \
  --rule-name "规则名称" \
  --origin-scheme follow --http-port 80 --https-port 443
```

#### 更新重定向规则端口
```bash
./update_esa-linux-arm64 --region-id cn-hangzhou --site-id <站点ID> \
  --rule-name "fn" \
  --redirect-port 2222
```

### 参数说明

*   `--region-id`: 必填，ESA 站点所属区域，如 `cn-hangzhou`。
*   `--site-id`: 必填，站点 ID。
*   `--config-id`: 规则配置 ID，优先使用此参数定位。
*   `--rule-name`: 规则名称，若不提供 ConfigId 则尝试按名称匹配（大小写不敏感）。
*   `--origin-scheme`: 回源协议，可选 `http`, `https`, `follow`。
*   `--http-port`: HTTP 回源端口（当协议为 http 或 follow 时需设置）。
*   `--redirect-port`: 重定向目标端口（仅适用于重定向规则更新）。
*   `--https-port`: HTTPS 回源端口（当协议为 https 或 follow 时需设置）。
*   `--list`: 列出当前规则，不执行更新。

## 本地编译

1.  安装 Go (1.20+): https://go.dev/dl/
2.  编译：
    ```bash
    go mod tidy
    go build -o update_esa main.go
    ```
