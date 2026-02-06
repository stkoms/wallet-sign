# Lotus Sign

Filecoin 钱包签名工具，用于管理密钥和签署区块链交易。

## 功能特性

### 钱包管理
- 生成新密钥（支持 BLS 和 secp256k1 类型）
- 导入/导出钱包密钥（支持 hex-lotus、json-lotus、gfc-json 格式）
- 查看钱包列表及余额
- 删除钱包密钥

### 交易操作
- FIL 转账
- 批量转账
- 矿工提现（从矿工账户提取余额）
- 市场提现（从存储市场托管账户提取）

### 矿工管理
- 查看矿工信息（owner、worker、control 地址及余额）
- 更改矿工 owner 地址
- 更改 worker 地址

## 安装

### 前置要求
- Go 1.25.1 或更高版本
- Lotus 节点访问权限（公共或私有）

### 编译
```bash
go build -o lotus-sign main.go
```

## 配置

编辑 `config.toml` 文件：

```toml
[Lotus]
Host = "https://api.node.glif.io/rpc/v0"  # Lotus 节点 RPC 地址
Token = ""                                 # API Token（可选）

[Security]
Seed = "your-encryption-seed"              # 加密种子

[Database]
Path = "~/.lotus-sign/wallet.db"           # 数据库路径
```

## 使用方法

### 钱包操作

```bash
# 生成新钱包
./lotus-sign wallet new secp256k1
./lotus-sign wallet new bls

# 查看钱包列表
./lotus-sign wallet list

# 查看钱包余额
./lotus-sign wallet balance <address>

# 导出钱包
./lotus-sign wallet export <address>

# 导入钱包
./lotus-sign wallet import <private-key>

# 删除钱包
./lotus-sign wallet delete <address>
```

### 转账操作

```bash
# 发送 FIL
./lotus-sign send --from <from-address> <to-address> <amount>

# 批量转账
./lotus-sign send --from <from-address> --batch <file>
```

### 矿工操作

```bash
# 查看矿工信息
./lotus-sign actor info <miner-id>

# 矿工提现
./lotus-sign withdraw <miner-id> <amount>

# 更改 owner
./lotus-sign actor set-owner <miner-id> <new-owner> <from-address>

# 更改 worker
./lotus-sign actor set-worker <miner-id> <new-worker>
./lotus-sign actor confirm-worker <miner-id>
```

### 市场操作

```bash
# 市场提现
./lotus-sign market withdraw <address> <amount>
```

### 消息推送

```bash
# 推送已签名消息
./lotus-sign push <signed-message>
```

## 项目结构

```
lotus-sign/
├── main.go                 # 程序入口
├── config.toml             # 配置文件
├── cli/                    # CLI 命令实现
│   ├── commands.go         # 命令注册
│   ├── wallet.go           # 钱包命令
│   ├── send.go             # 转账命令
│   ├── actor.go            # 矿工命令
│   ├── withdraw.go         # 提现命令
│   ├── market.go           # 市场命令
│   └── push.go             # 消息推送
├── internal/
│   ├── config/             # 配置加载
│   ├── crypto/             # 加密工具
│   ├── repository/         # 数据持久化
│   ├── wallet/             # 钱包加密操作
│   ├── service/            # 业务逻辑
│   ├── chain/              # Filecoin 链类型
│   ├── rpc/                # JSON-RPC 客户端
│   ├── vapi/               # 节点 API 封装
│   ├── models/             # 数据库模型
│   └── ui/                 # UI 工具
└── lib/
    └── signlog/            # 日志配置
```

## 安全特性

- 私钥加密存储于 SQLite 数据库
- 支持 BLS 和 secp256k1 签名算法
- 危险操作需用户确认
- 完整的操作日志记录

## 许可证

Apache 2.0