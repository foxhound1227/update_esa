# Update ESA Origin Rule

这是一个用于更新阿里云 ESA（边缘安全加速）回源规则配置的工具。支持 Python 脚本直接运行，也提供适用于 OpenWrt 等环境的二进制文件。

## 功能特性

*   修改 ESA 站点的回源协议（HTTP/HTTPS/Follow）和端口。
*   修改 ESA 站点的重定向规则目标端口（支持动态拼接 URL）。
*   支持按「规则名称」或「配置 ID (ConfigId)」定位规则。
*   支持列出当前站点的所有规则及配置。
*   提供 x86_64、ARM64 (aarch64) 和 ARMv7 的静态编译二进制文件，无需 Python 环境即可运行。

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
*   `update_esa_x86_64`: 普通 PC 或服务器
*   `update_esa_aarch64`: 树莓派 4、R2S、R4S 等
*   `update_esa_armv7`: 较老的 ARM 路由器

上传到设备并赋予执行权限：
```bash
chmod +x update_esa_aarch64
```

### 3. 常用命令示例

#### 查询规则列表
查看站点下所有规则的 ConfigId、名称和当前端口：
```bash
./update_esa_aarch64 --region-id cn-hangzhou --site-id <站点ID> --list --origin-scheme follow
```

#### 更新回源规则
**方式一：按 ConfigId 更新（推荐，更精准）**
```bash
./update_esa_aarch64 --region-id cn-hangzhou --site-id <站点ID> \
  --config-id <ConfigID> \
  --origin-scheme https --https-port 1235
```

**方式二：按规则名称更新**
```bash
./update_esa_aarch64 --region-id cn-hangzhou --site-id <站点ID> \
  --rule-name "规则名称" \
  --origin-scheme follow --http-port 80 --https-port 443
```
# 更新重定向规则端口
```bash
./update_esa_aarch64--region-id cn-hangzhou --site-id <站点ID> \
  --rule-name "fn" \
  --redirect-port 2222
```

### 
### 参数说明

*   `--region-id`: 必填，ESA 站点所属区域，如 `cn-hangzhou`。
*   `--site-id`: 必填，站点 ID。
*   `--config-id`: 规则配置 ID，优先使用此参数定位。
*   `--rule-name`: 规则名称，若不提供 ConfigId 则尝试按名称匹配（大小写不敏感）。
*   `--origin-scheme`: 回源协议，可选 `http`, `https`, `follow`。
*   `--http-port`: HTTP 回源端口（当协议为 http 或 follow 时需设置）。）。
*   `--redirect-port`: 重定向目标端口（仅适用于重定向规则更新
*   `--https-port`: HTTPS 回源端口（当协议为 https 或 follow 时需设置）。
*   `--list`: 列出当前规则，不执行更新。

## 编译构建

本项目使用 GitHub Actions 自动构建。如果你想本地运行或编译：

1.  安装依赖：
    ```bash
    pip install -r requirements.txt
    ```
2.  运行脚本：
    ```bash
    python3 update_esa.py --help
    ```
