# CipherHub

一个安全、便携的命令行密码管理器。

## 目录

- [功能特性](#功能特性)
- [安全机制](#安全机制)
- [快速开始](#快速开始)
- [安装](#安装)
- [使用指南](#使用指南)
- [高级功能](#高级功能)
- [公共 API](#公共-api)
- [文件存储](#文件存储)
- [项目结构](#项目结构)
- [开发路线](#开发路线)
- [代码质量](#代码质量)
- [许可证](#许可证)

---

## 功能特性

| 特性 | 说明 |
|------|------|
| **便携存储** | 数据文件默认存储在程序同目录，U盘即插即用 |
| **自定义路径** | 支持 `--config` 和 `--vault` 参数指定任意存储位置 |
| **AES-256-GCM** | 所有密码使用 AES-256-GCM 算法加密 |
| **Argon2id** | 主密码通过 Argon2id 安全派生加密密钥 |
| **WebDAV 同步** | 支持同步到任何 WebDAV 兼容的云存储 |
| **密码隐藏** | 交互式输入密码时不显示明文 |

---

## 安全机制

| 机制 | 实现 |
|------|------|
| 加密算法 | AES-256-GCM 认证加密 |
| 密钥派生 | Argon2id（64MB 内存，3 次迭代，4 线程）|
| 盐值 | 每个密码库随机 16 字节 |
| Nonce | 每次加密随机 12 字节 |
| 完整性 | SHA-256 校验和 |

---

## 快速开始

### 1. 初始化密码库

```bash
cipherhub init
```

系统会提示创建主密码（输入时隐藏显示）。

### 2. 添加密码

```bash
cipherhub add github
```

或直接指定参数：

```bash
cipherhub add github -u myuser -p mypass -U https://github.com -t "工作,代码"
```

### 3. 获取密码

```bash
cipherhub get github
```

显示密码明文：

```bash
cipherhub get github -p
```

### 4. 列出条目

```bash
cipherhub list
cipherhub list -s github
```

### 5. 更新密码条目

```bash
cipherhub update github
```

或直接指定参数：

```bash
cipherhub update github -u newuser -p newpass -U https://new.github.com
```

### 6. 删除条目

```bash
cipherhub delete github
```

---

## 安装

### 从源码编译

```bash
git clone https://github.com/imerr0rlog/CipherHub.git
cd CipherHub
go mod tidy
go build -o cipherhub ./cmd/cipherhub
```

### 跨平台编译

```bash
# Windows 64位
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o cipherhub-windows-amd64.exe ./cmd/cipherhub

# Linux 64位
GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o cipherhub-linux-amd64 ./cmd/cipherhub

# macOS ARM64 (M1/M2/M3)
GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w" -o cipherhub-darwin-arm64 ./cmd/cipherhub
```

---

## 使用指南

### 命令参考

| 命令 | 说明 |
|------|------|
| `init` | 初始化密码库 |
| `add <名称>` | 添加密码条目 |
| `get <名称>` | 获取密码条目 |
| `update <名称>` | 更新密码条目 |
| `list` | 列出所有条目 |
| `delete <名称>` | 删除条目 |
| `config` | 管理配置 |
| `sync` | WebDAV 同步 |
| `generate` | 生成随机密码 |
| `version` | 显示版本 |

### 全局参数

| 参数 | 说明 | 默认值 |
|------|------|--------|
| `--config` | 配置文件路径 | 程序同目录 `config.json` |
| `--vault` | 密码库文件路径 | 程序同目录 `vault.json` |

#### 使用示例

**默认模式**（文件存储在程序同目录）：
```bash
cipherhub init
cipherhub add github
```

**自定义路径模式**：
```bash
# 指定数据目录
cipherhub --config D:\mydata\config.json --vault D:\mydata\vault.json init

# 仅指定密码库路径
cipherhub --vault D:\mydata\vault.json get github
```

### 命令详细说明

#### add 参数

```
-u, --username   用户名
-p, --password   密码（不提供则交互输入）
-U, --url        网站地址
-n, --notes      备注
-t, --tags       标签（逗号分隔）
```

#### get 参数

```
-p, --password   显示密码明文
-n, --notes      显示备注
-c, --copy       复制密码到剪贴板
```

#### update 参数

```
-u, --username   新用户名
-p, --password   新密码（留空保持现有密码）
-U, --url        新 URL
-n, --notes      新备注
```

#### list 参数

```
-s, --search     搜索条目
```

#### 密码生成

```bash
cipherhub generate -l 24
```

#### 配置管理

```bash
# 查看配置
cipherhub config --show

# 配置 WebDAV
cipherhub config --webdav-url https://webdav.example.com/dav
cipherhub config --webdav-user 用户名
cipherhub config --webdav-pass 密码
cipherhub config --webdav-path /cipherhub/vault.json
cipherhub config --webdav-config-path /cipherhub/config.json
```

---

## 高级功能

### WebDAV 云同步

#### 同步流程

```
本地文件                WebDAV 云端
─────────              ────────────
vault.json    <--->    /cipherhub/vault.json
config.json   <--->    /cipherhub/config.json
```

#### 配置步骤

```bash
# 1. 初始化本地密码库
cipherhub init

# 2. 配置 WebDAV 连接信息
cipherhub config --webdav-url https://webdav.example.com/dav \
                 --webdav-user 用户名 \
                 --webdav-pass 密码 \
                 --webdav-path /cipherhub/vault.json \
                 --webdav-config-path /cipherhub/config.json

# 3. 推送到云端（默认同时推送 vault 和 config）
cipherhub sync

# 4. 从云端拉取
cipherhub sync --pull
cipherhub sync --pull --force  # 跳过确认
```

#### sync 参数

| 参数 | 说明 |
|------|------|
| `--pull` | 从云端拉取到本地 |
| `-f, --force` | 拉取时跳过确认 |
| `--vault-only` | 仅同步 vault.json 文件 |
| `--config-only` | 仅同步 config.json 文件 |

#### 单独同步示例

```bash
# 仅推送 vault
cipherhub sync --vault-only

# 仅拉取 config
cipherhub sync --pull --config-only

# 强制覆盖拉取 vault
cipherhub sync --pull --vault-only --force
```

---

## 公共 API

CipherHub 提供 `pkg/api` 包供外部程序集成使用。

### 快速开始

```go
package main

import (
    "fmt"
    "github.com/imerr0rlog/CipherHub/pkg/api"
)

func main() {
    // 创建客户端
    client, err := api.NewClientWithOptions(&api.ClientOptions{
        VaultPath:  "./mydata/vault.json",
        ConfigPath: "./mydata/config.json",
    })
    if err != nil {
        panic(err)
    }

    // 初始化密码库
    if err := client.InitVault("master-password"); err != nil {
        panic(err)
    }

    // 添加条目
    entry, err := client.AddEntry("github", "user", "pass", "https://github.com", "", nil)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Added: %s\n", entry.Name)

    // 获取解密后的密码
    password, _ := client.GetDecryptedPassword("github")
    fmt.Printf("Password: %s\n", password)
}
```

### API 参考

| 方法 | 说明 |
|------|------|
| **客户端管理** | |
| `NewClient(cfg)` | 使用配置创建客户端 |
| `NewClientWithOptions(opts)` | 使用选项创建客户端 |
| `LoadConfig(path)` | 加载配置文件 |
| `SaveConfig(path, cfg)` | 保存配置文件 |
| `DefaultConfig()` | 获取默认配置 |
| **密码库操作** | |
| `InitVault(password)` | 初始化密码库 |
| `OpenVault(password)` | 打开密码库 |
| `CloseVault()` | 关闭密码库 |
| `IsVaultOpen()` | 检查密码库是否打开 |
| `VaultExists()` | 检查密码库是否存在 |
| **条目管理** | |
| `AddEntry(...)` | 添加条目 |
| `GetEntry(name)` | 获取条目 |
| `GetDecryptedPassword(name)` | 获取解密密码 |
| `GetDecryptedNotes(name)` | 获取解密备注 |
| `ListEntries()` | 列出所有条目 |
| `SearchEntries(query)` | 搜索条目 |
| `UpdateEntry(name, updates)` | 更新条目 |
| `DeleteEntry(name)` | 删除条目 |
| **WebDAV 同步** | |
| `SyncToWebDAV(opts)` | 推送到 WebDAV |
| `PullFromWebDAV(opts)` | 从 WebDAV 拉取 |
| **工具函数** | |
| `GeneratePassword(length)` | 生成随机密码 |
| `Encrypt(password, salt, plaintext)` | 加密字符串 |
| `Decrypt(password, salt, ciphertext)` | 解密字符串 |
| `GenerateSalt()` | 生成盐值 |

### WebDAV 同步示例

```go
// 配置 WebDAV
cfg := client.Config()
cfg.WebDAV = &types.WebDAVConfig{
    URL:              "https://webdav.example.com/dav",
    Username:         "user",
    Password:         "pass",
    RemotePath:       "/cipherhub/vault.json",
    ConfigRemotePath: "/cipherhub/config.json",
}

// 推送 vault 和 config
client.SyncToWebDAV(nil)

// 仅推送 vault
client.SyncToWebDAV(&api.SyncOptions{SyncVault: true})

// 拉取
client.PullFromWebDAV(nil)
```

---

## 文件存储

### 默认位置

程序所在目录

```
cipherhub.exe
config.json      # 配置文件
vault.json       # 密码库
```

### vault.json 结构

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

### config.json 结构

```json
{
  "default_storage": "local",
  "vault_path": "./vault.json",
  "webdav": {
    "url": "https://webdav.example.com/dav",
    "username": "用户名",
    "password": "密码",
    "remote_path": "/cipherhub/vault.json",
    "config_remote_path": "/cipherhub/config.json"
  }
}
```

---

## 项目结构

```
CipherHub/
├── cmd/cipherhub/          # 程序入口
├── internal/
│   ├── cli/                # 命令行处理
│   ├── crypto/             # 加密模块
│   ├── storage/            # 存储后端
│   └── vault/              # 密码库管理
├── pkg/
│   ├── api/                # 公共 API
│   └── types/              # 数据类型
├── go.mod
├── Makefile
└── README.md
```

---

## 开发路线

- [ ] 桌面端 GUI 应用
- [ ] WebUI 界面
- [ ] 浏览器扩展
- [ ] TOTP 双因素认证
- [ ] 多密码库支持

---

## 代码质量

CipherHub 项目注重代码质量和安全性，包括：

- ✅ 完整的错误处理，所有随机数生成操作都有正确的错误检查
- ✅ 正确的 SHA-256 校验和计算，确保数据完整性
- ✅ 搜索结果去重，避免重复条目显示
- ✅ 清晰的代码注释，便于维护和理解
- ✅ 安全的内存清理，防止密钥泄露

---

## 许可证

MIT License
