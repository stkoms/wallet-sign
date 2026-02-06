# Lotus Sign - 本地转账和签名工具

这是一个功能完整的 Lotus 钱包命令行工具，支持两种工作模式：**完整 API 模式**和**轻量级本地签名模式**。

## 两种工作模式

### 模式 1: 完整 API 模式

使用 Lotus 官方 RPC API，功能完整，适合需要完整节点功能的场景。

**特点:**

- 使用完整的 Lotus API 客户端
- 支持本地转账和矿工提现
- 依赖完整的 Lotus 库
- 编译时间较长（3-5 分钟）
- 二进制文件较大（~100MB）

### 模式 2: 轻量级本地签名模式

使用标准 HTTP JSON-RPC，避免完整 Lotus 依赖，大幅减少编译时间。

**特点:**

- 轻量级 RPC 客户端，不依赖完整 Lotus API
- 本地签名交易，无需 Lotus 钱包
- 快速编译（5-10 秒）
- 二进制文件小（~10MB）
- 依赖包少（<20 个）
- 支持连接公开 RPC 节点，无需认证

## 功能特性

- ✓ **本地转账**: 在地址之间转账 FIL
- ✓ **本地签名**: 使用私钥在本地签名交易
- ✓ **矿工提现**: 从矿工账户提取可用余额（完整 API 模式）
- ✓ **余额查询**: 查询地址余额
- ✓ **密钥生成**: 生成新的密钥对（轻量级模式）
- ✓ **无需 Token**: 支持连接公开 RPC 节点

## 依赖对比

### 完整 API 模式

```
github.com/filecoin-project/lotus v1.25.2
github.com/filecoin-project/go-address
github.com/filecoin-project/go-jsonrpc
github.com/filecoin-project/go-state-types
```

### 轻量级模式

```
github.com/filecoin-project/go-address v1.1.0      // 地址处理
github.com/filecoin-project/go-state-types v0.13.3 // 基础类型
github.com/ipfs/go-cid v0.4.1                      // CID 支持
github.com/whyrusleeping/cbor-gen v0.0.0-...       // CBOR 编码
github.com/btcsuite/btcd v0.23.4                   // 加密签名
```

编译时间从 **几分钟** 减少到 **几秒钟**。

## 安装

### 前置要求

- Go 1.21 或更高版本
- 运行中的 Lotus 节点（或公开 RPC 节点地址）

### 编译

```bash
# 安装依赖
go mod tidy

# 编译
go build -o bin/lotus-sign ./cmd/lotus-sign

### 启动 Web 控制台（Gin + React）

需要配置 MySQL、Telegram Bot 和 TOTP：

```bash
export DB_DSN="root@tcp(127.0.0.1:3306)/lotus_sign?charset=utf8mb4&parseTime=True&loc=Local" # 可用 configs/config.toml 替代
export TELEGRAM_BOT_TOKEN="your-telegram-bot-token"
export TELEGRAM_ADMIN_IDS="12345678,87654321"
export TELEGRAM_NOTIFY_CHAT_ID="12345678" # 可选
export STARTUP_AUTH_URL="https://auth.example.com/verify"
export STARTUP_AUTH_TIMEOUT="3s"
export STARTUP_AUTH_RETRIES="5"
export STARTUP_TOTP_SECRET="your-startup-totp-secret"
export TOTP_SECRET="your-google-authenticator-secret"
export JWT_SECRET="your-jwt-secret"
export BOOTSTRAP_USER="admin"
export BOOTSTRAP_PASSWORD="change-me"
export BOOTSTRAP_ROLE="admin"
export BOOTSTRAP_TOTP_SECRET="your-bootstrap-user-totp-secret"
export ADMIN_TOKEN="your-admin-token" # 可选，绕过登录
export HTTP_ADDR=":8080"

./bin/lotus-sign serve
```

浏览器访问 `http://localhost:8080`。

登录说明：

- 首次启动会创建 `BOOTSTRAP_USER` 用户（若不存在）
- 登录需要用户名、密码和用户专属 TOTP 验证码
- 登录后会获得 JWT（24h 有效），用于 Web 权限控制
- 角色：`admin`（配置/用户管理）、`operator`（发起操作）、`viewer`（只读）

启动校验：

- 启动时先进行在线授权验证（POST 到 `STARTUP_AUTH_URL`，期望 `{"ok":true}`），失败后自动切换为离线 TOTP
- 需要设置 `STARTUP_TOTP_SECRET`，在线失败时会提示输入 TOTP 验证码

### 数据库与 Redis

- MySQL: `127.0.0.1:3306`，用户 `root`，密码为空（示例见 `DB_DSN`）
- Redis: `127.0.0.1:6379`，DB `15`，密码为空（当前版本未启用，仅保留连接信息）

数据库行为：

- 启动时自动创建 `DB_DSN` 指定的数据库（若不存在）
- 自动迁移表结构（包含审批、监控、用户、私钥、交易记录等）
- 可执行 `./bin/lotus-sign db init` 单独创建数据库并同步表

使用 `configs/config.toml` 配置数据库（可替代 `DB_DSN`，环境变量优先）：

```toml
[Database]
Host = "127.0.0.1"
Port = 3306
User = "root"
Password = ""
Name = "lotus_sign"
Params = "charset=utf8mb4&parseTime=True&loc=Local"
```

其它配置也可写入 `configs/config.toml`（环境变量优先）：

```toml
[Server]
HTTPAddr = ":8080"
AdminToken = ""
MonitorInterval = "1m"

[Auth]
TOTPSecret = ""
JWTSecret = ""
BootstrapUser = "admin"
BootstrapPassword = "change-me"
BootstrapRole = "admin"
BootstrapTOTPSecret = ""

[Telegram]
Token = ""
AdminIDs = [12345678, 87654321]
NotifyChatID = 0
Disabled = false

[StartupAuth]
AuthURL = "https://auth.example.com/verify"
Timeout = "3s"
Retries = 5
TOTPSecret = ""
Disabled = false
```

临时关闭启动校验：

- 设置 `STARTUP_AUTH_DISABLED=true`，或在 `configs/config.toml` 中将 `StartupAuth.Disabled = true`

钱包私钥入库：

- `wallet import/imports/import-new` 成功后会把私钥加密保存到数据库（需设置 `DB_DSN`）

批量转账地址：

- 通过数据库 `batch_addresses` 配置与读取（Web 后台或 API），不再依赖 `configs/config.toml` 的钱包地址配置

```

## 配置

使用环境变量配置 Lotus API 连接（**可选**）：

```bash
# API 地址（可选，默认为本地节点）
export LOTUS_API_URL="http://127.0.0.1:1234/rpc/v0"

# 或连接公开 RPC 节点
export LOTUS_API_URL="https://api.node.glif.io"

# API Token（可选，公开 RPC 不需要）
export LOTUS_API_TOKEN="your-api-token"

# 本地钱包数据路径（可选，默认 data/lotus-repo）
export LOTUS_REPO="./data/lotus-repo"
```

**说明：**

- 如果不设置 `LOTUS_API_URL`，默认连接到 `http://127.0.0.1:1234/rpc/v0`
- 如果不设置 `LOTUS_API_TOKEN`，将不使用认证（适用于公开 RPC 或本地无认证节点）
- 只有需要认证的节点才需要设置 token

### 获取 API Token（如果需要）

```bash
# 从 Lotus 节点获取 token
cat ~/.lotus/token
```

## 使用方法

### 1. 生成新密钥对（轻量级模式）

```bash
./lotus-wallet gen-key
```

输出示例：

```
Generated new secp256k1 key:
Private Key: 1234567890abcdef...
Address: f1abc...
```

**重要**: 请妥善保管私钥，不要泄露！

### 2. 查询余额

```bash
./lotus-wallet balance <address>
```

**示例：**

```bash
./lotus-wallet balance f1abc...
```

### 3. 本地转账

#### 方式 A: 完整 API 模式

使用 Lotus API 自动处理签名和发送：

```bash
./lotus-wallet transfer <from_address> <to_address> <amount>
```

**示例：**

```bash
./lotus-wallet transfer f1abc... f1xyz... 10
```

这将从 `f1abc...` 转账 10 FIL 到 `f1xyz...`

**当前版本**：CLI 转账/提现/批量转账/矿工地址变更会创建审批请求，并等待 Telegram + TOTP 审批后执行。

#### 方式 B: 轻量级本地签名模式

使用私钥在本地签名交易：

```bash
./lotus-wallet transfer-local <from> <to> <amount> <private_key_hex>
```

**示例：**

```bash
./lotus-wallet transfer-local f1abc... f1xyz... 10 1234567890abcdef...
```

这将：

1. 从 RPC 获取 nonce 和 gas 估算
2. 在本地使用私钥签名交易
3. 发送已签名的交易到网络
4. 等待确认

### 4. 矿工提现（完整 API 模式）

从矿工账户提取可用余额：

```bash
./lotus-wallet withdraw <miner_address> <amount>
```

**示例：**

```bash
./lotus-wallet withdraw f01234 5
```

这将从矿工 `f01234` 提取 5 FIL 到矿工的 owner 地址

### 5. 市场提现（Storage Market）

从存储市场余额中提现：

```bash
./bin/lotus-sign market-withdraw <address> <amount>
```

**示例：**

```bash
./bin/lotus-sign market-withdraw f1abc... 5
```

## 工作原理

### 完整 API 模式转账流程

1. 连接到 Lotus RPC API
2. 验证发送和接收地址
3. 构造转账消息
4. 使用 `MpoolPushMessage` 推送消息到消息池
5. 等待消息被打包到区块中
6. 验证执行结果

### 轻量级模式转账流程

1. **创建消息**: 构造 Filecoin 消息结构
2. **获取 Nonce**: 从 RPC 获取账户 nonce
3. **Gas 估算**: 从 RPC 获取 gas 参数
4. **本地签名**: 使用私钥在本地签名消息
5. **发送交易**: 通过 RPC 发送已签名的消息
6. **等待确认**: 等待交易被打包确认

### 审批流程（Web/CLI）

1. 操作发起（转账/提现/批量/矿工地址变更）
2. 生成审批请求并通过 Telegram 推送
3. 管理员使用 Google Authenticator 生成 TOTP 码
4. Telegram 中 `/approve <id> <totp>` 批准后执行

CLI 运行时需要相同的数据库与 Telegram 配置（用于生成请求并通知审批）。

### 矿工提现流程

1. 连接到 Lotus RPC API
2. 获取矿工信息和可用余额
3. 验证可用余额是否充足
4. 构造提现消息（Method 16 - WithdrawBalance）
5. 从矿工 owner 地址发送消息
6. 等待消息执行完成

## 架构说明

### 文件结构

```
lotus-sign/
├── bin/lotus-sign             # 编译产物
├── cmd/lotus-sign/main.go     # CLI 入口
├── configs/config.toml        # 本地配置
├── data/lotus-repo/           # 本地钱包数据
├── internal/commands/         # 命令实现
├── internal/chain/types/      # 本地链类型与 CBOR
├── internal/chain/actors/     # 参数序列化
├── internal/rpc/              # JSON-RPC 客户端
├── internal/vapi/             # RPC API 封装
├── internal/wallet/           # 本地钱包与签名
├── internal/ui/tablewriter/   # 表格输出
├── internal/config/           # 配置结构
├── web/                       # Web 前端
└── go.mod                     # 依赖管理
```

## 安全注意事项

### 1. 私钥安全

- 永远不要在命令行历史中暴露私钥
- 考虑使用环境变量或配置文件
- 生产环境建议使用硬件钱包
- 妥善保管私钥，不要泄露

### 2. 网络安全

- 使用 HTTPS 连接 RPC 节点
- 验证 RPC 节点的可信度
- 公开 RPC 节点（如 Glif.io）不需要 token

### 3. Gas 费用

- 确保账户有足够余额支付 gas
- 可以手动设置 gas 参数
- 所有操作都会消耗 gas 费用

### 4. 矿工提现

- 只能提取可用余额，锁定的余额无法提取
- 矿工提现需要使用矿工的 owner 地址签名

### 5. 确认时间

- 消息需要等待 3 个区块确认

## 错误处理

常见错误及解决方法：

- **连接失败**: 检查 LOTUS_API_URL 是否正确，确保 Lotus 节点正在运行
- **认证失败**: 如果使用需要认证的节点，检查 LOTUS_API_TOKEN 是否有效（公开 RPC 无需 token）
- **余额不足**: 确保账户有足够的余额支付转账金额和 gas 费用
- **地址无效**: 确保使用正确的 Filecoin 地址格式

**提示**: 大多数公开的 Lotus RPC 节点（如 Glif.io）不需要 token，可以直接使用。

## 模式对比

| 特性    | 完整 API 模式 | 轻量级模式  |
|-------|-----------|--------|
| 编译时间  | 3-5 分钟    | 5-10 秒 |
| 二进制大小 | ~100MB    | ~10MB  |
| 依赖数量  | 200+      | <20    |
| 本地转账  | ✓         | ✓      |
| 本地签名  | ✓         | ✓      |
| 密钥生成  | -         | ✓      |
| 矿工提现  | ✓         | 待实现    |
| 钱包管理  | ✓         | 基础功能   |
| 矿工操作  | ✓         | 待实现    |
| 节点运行  | ✓         | ✗      |

## 选择建议

### 使用完整 API 模式，如果你需要：

- 矿工提现功能
- 完整的 Lotus 节点功能
- 更多高级功能

### 使用轻量级模式，如果你需要：

- 快速编译和部署
- 小的二进制文件
- 仅需要基本的转账和签名功能
- 连接公开 RPC 节点

## 许可证

MIT License
