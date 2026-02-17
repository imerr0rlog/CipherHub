# CipherHub

一个安全的命令行密码管理器。

## 功能特性

- **AES-256-GCM 加密**：所有密码使用 AES-256-GCM 算法加密
- **Argon2id 密钥派生**：主密码通过 Argon2id 安全派生加密密钥
- **本地存储**：密码库存储在本地 `~/.cipherhub/vault.json`
- **WebDAV 同步**：支持同步到任何 WebDAV 兼容的云存储
- **命令行界面**：功能完整的命令行工具
- **GUI 就绪**：公共 API 设计支持未来开发 GUI/WebUI

## 安装

### 前置条件

- Go 1.21 或更高版本

### 从源码编译

```bash
git clone https://github.com/cipherhub/cli.git
cd cli
make deps
make build
```

编译后的二进制文件位于 `bin/cipherhub`。

### 跨平台编译

编译特定平台版本：

```bash
# Windows 64位
make build-windows

# Windows ARM64
make build-windows-arm64

# Linux 64位
make build-linux

# Linux ARM64
make build-linux-arm64

# macOS 64位 (Intel)
make build-darwin

# macOS ARM64 (M1/M2/M3)
make build-darwin-arm64

# 编译所有平台
make build-all
```

或使用构建脚本：

```bash
# Linux/macOS
./build.sh windows

# Windows
build.bat windows
```

### 安装到系统路径

```bash
make install
```

## 快速开始

### 1. 初始化密码库

```bash
cipherhub init
```

系统会提示您创建主密码。

### 2. 添加密码条目

```bash
cipherhub add github
```

或使用命令行参数：

```bash
cipherhub add github -u 用户名 -p 密码 -U https://github.com -t "工作,代码"
```

### 3. 获取密码

```bash
cipherhub get github
```

显示密码：

```bash
cipherhub get github -p
```

### 4. 列出所有条目

```bash
cipherhub list
```

搜索条目：

```bash
cipherhub list -s github
```

### 5. 删除条目

```bash
cipherhub delete github
```

## WebDAV 云同步

配置 WebDAV 云存储：

```bash
cipherhub config --webdav-url https://webdav.example.com/dav \
                 --webdav-user 用户名 \
                 --webdav-pass 密码 \
                 --webdav-path /cipherhub/vault.json
```

推送到云端：

```bash
cipherhub sync
```

从云端拉取：

```bash
cipherhub sync --pull
```

## 密码生成

生成安全的随机密码：

```bash
cipherhub generate -l 24
```

## 配置管理

查看当前配置：

```bash
cipherhub config --show
```

切换到本地存储：

```bash
cipherhub config --local
```

## 项目结构

```
cipherhub/
├── cmd/cipherhub/          # 命令行入口
├── internal/
│   ├── crypto/             # AES-256-GCM + Argon2 加密模块
│   ├── storage/            # 本地 + WebDAV 存储后端
│   ├── vault/              # 密码库管理逻辑
│   └── cli/                # Cobra 命令行命令
├── pkg/
│   ├── api/                # 公共 API（供 GUI 集成）
│   └── types/              # 共享数据类型
├── go.mod                  # Go 模块定义
├── Makefile                # 构建命令
├── build.sh                # Linux/macOS 构建脚本
└── build.bat               # Windows 构建脚本
```

## 命令参考

| 命令 | 说明 |
|------|------|
| `cipherhub init` | 初始化新的密码库 |
| `cipherhub add <名称>` | 添加密码条目 |
| `cipherhub get <名称>` | 获取密码条目 |
| `cipherhub list` | 列出所有条目 |
| `cipherhub delete <名称>` | 删除条目 |
| `cipherhub config` | 管理配置 |
| `cipherhub sync` | 同步到 WebDAV |
| `cipherhub generate` | 生成随机密码 |
| `cipherhub version` | 显示版本信息 |

### 命令参数

#### add 命令

```
cipherhub add <名称> [参数]

参数:
  -u, --username string   用户名
  -p, --password string   密码（不提供则交互式输入）
  -U, --url string        网站地址
  -n, --notes string      备注
  -t, --tags string       标签（逗号分隔）
```

#### get 命令

```
cipherhub get <名称> [参数]

参数:
  -p, --password   显示密码
  -n, --notes      显示备注
  -c, --copy       复制密码到剪贴板
```

#### list 命令

```
cipherhub list [参数]

参数:
  -s, --search string   搜索条目
```

## 安全机制

- **加密算法**：AES-256-GCM 提供认证加密
- **密钥派生**：Argon2id（64MB 内存，3 次迭代，4 线程）
- **盐值**：每个密码库使用随机 16 字节盐值
- **随机数**：每次加密使用随机 12 字节 Nonce
- **完整性校验**：SHA-256 校验和确保数据完整性

## 数据存储

密码库文件格式（`vault.json`）：

```json
{
  "version": "1.0",
  "salt": "base64编码的盐值",
  "checksum": "SHA-256校验和",
  "entries": [
    {
      "id": "唯一标识",
      "name": "条目名称",
      "username": "用户名",
      "password": "AES-256-GCM加密的密码",
      "url": "网站地址",
      "notes": "加密的备注",
      "created_at": "创建时间",
      "updated_at": "更新时间"
    }
  ]
}
```

配置文件（`config.json`）：

```json
{
  "default_storage": "local",
  "vault_path": "~/.cipherhub/vault.json",
  "webdav": {
    "url": "https://webdav.example.com/dav",
    "username": "用户名",
    "password": "密码",
    "remote_path": "/cipherhub/vault.json"
  }
}
```

## 开发路线

- [ ] 桌面端 GUI 应用
- [ ] WebUI 界面
- [ ] 浏览器扩展
- [ ] TOTP 双因素认证支持
- [ ] 安全剪贴板集成
- [ ] 多密码库支持

## 许可证

MIT License
