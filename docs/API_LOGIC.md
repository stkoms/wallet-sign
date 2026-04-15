# Lotus-Sign API 和逻辑文档

## 项目概述

Lotus-Sign 是一个 Filecoin 钱包本地签名工具，支持：

- 钱包管理（生成、导入、导出、查询余额）
- 转账操作（单笔、批量）
- 矿工管理（查看信息、变更所有者、变更Worker）
- 矿工提现和市场提现

## 架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                        CLI Layer                             │
│  (send, wallet, actor, withdraw, market-withdraw, push)     │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                     Service Layer                            │
│  (Executor: transfer, minerWithdraw, marketWithdraw...)     │
└─────────────────────────────────────────────────────────────┘
                              │
              ┌───────────────┼───────────────┐
              ▼               ▼               ▼
┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
│   Wallet Layer  │ │    RPC Layer    │ │  Repository     │
│ (签名、密钥管理) │ │ (Lotus节点通信) │ │  (数据库访问)   │
└─────────────────┘ └─────────────────┘ └─────────────────┘
              │               │               │
              ▼               ▼               ▼
┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
│  Crypto Layer   │ │  Chain Types    │ │    Database     │
│  (加密/解密)    │ │  (消息、签名)   │ │    (MySQL)      │
└─────────────────┘ └─────────────────┘ └─────────────────┘
```

---

## 数据模型

### WalletKey 钱包密钥表

| 字段            | 类型          | 说明                   |
|---------------|-------------|----------------------|
| id            | uint        | 主键，自增                |
| address       | string(128) | 钱包地址，唯一索引            |
| key_type      | string(32)  | 密钥类型：secp256k1 或 bls |
| encrypted_key | blob        | 加密后的私钥               |
| created_at    | datetime    | 创建时间                 |
| updated_at    | datetime    | 更新时间                 |

---

## CLI 命令接口

### 1. wallet - 钱包管理

#### wallet new

生成新钱包

```bash
lotus-sign wallet new [--type secp256k1|bls]
```

- `--type`: 密钥类型，默认 secp256k1

#### wallet list

列出所有钱包

```bash
lotus-sign wallet list
```

#### wallet balance

查询钱包余额

```bash
lotus-sign wallet balance <address>
```

#### wallet export

导出钱包私钥

```bash
lotus-sign wallet export <address>
```

#### wallet import

导入钱包（仅存储到数据库）

```bash
lotus-sign wallet import <hex_private_key>
```

#### wallet importnew

导入钱包（存储并返回地址）

```bash
lotus-sign wallet importnew <hex_private_key>
```

---

### 2. send - 转账

```bash
lotus-sign send --from <address> --to <address> --amount <fil_amount>
```

参数：

- `--from`: 发送方地址
- `--to`: 接收方地址
- `--amount`: 转账金额（FIL）

---

### 3. actor - 矿工管理

#### actor info

查看矿工信息

```bash
lotus-sign actor info <miner_id>
```

#### actor set-owner

变更矿工所有者

```bash
lotus-sign actor set-owner --miner <miner_id> --new-owner <address> --from <current_owner>
```

#### actor propose-change-worker

提议变更Worker

```bash
lotus-sign actor propose-change-worker --miner <miner_id> --worker <new_worker> --from <owner>
```

#### actor confirm-change-worker

确认变更Worker

```bash
lotus-sign actor confirm-change-worker --miner <miner_id> --from <owner>
```

---

### 4. withdraw - 矿工提现

```bash
lotus-sign withdraw --miner <miner_id> --amount <fil_amount> --from <owner>
```

---

### 5. market-withdraw - 市场提现

```bash
lotus-sign market-withdraw --address <address> --amount <fil_amount> --from <address>
```

---

### 6. mpool-push - 推送消息

```bash
lotus-sign mpool-push <signed_message_json>
```

---

## 核心业务逻辑

### 1. 转账流程 (transfer)

```
1. 获取发送方 nonce (MpoolGetNonce)
2. 构造 Message 结构
3. 估算 Gas 参数 (GasEstimateMessageGas)
4. 检查本地是否有私钥 (WalletHas)
5. 使用私钥签名消息 (WalletSign)
6. 推送已签名消息到内存池 (MpoolPush)
7. 等待消息确认 (StateWaitMsg)
```

### 2. 矿工提现流程 (minerWithdraw)

```
1. 获取矿工信息 (StateMinerInfo)
2. 验证 from 地址是否为 Owner
3. 构造 WithdrawBalance 方法调用
4. 序列化参数 (CBOR)
5. 执行转账流程
```

### 3. 市场提现流程 (marketWithdraw)

```
1. 查询市场余额 (StateMarketBalance)
2. 构造 MarketWithdraw 方法调用
3. 序列化参数 (CBOR)
4. 执行转账流程
```

### 4. 变更矿工所有者 (changeMinerOwner)

```
1. 获取矿工信息
2. 验证当前所有者
3. 构造 ChangeOwnerAddress 方法调用
4. 执行转账流程
```

### 5. 变更Worker流程

```
提议阶段 (proposeChangeWorker):
1. 构造 ChangeWorkerAddress 方法调用
2. 执行转账流程

确认阶段 (confirmChangeWorker):
1. 构造 ConfirmChangeWorker 方法调用
2. 执行转账流程
```

---

## RPC 接口

### Lotus JSON-RPC 方法

| 方法                                    | 说明      | 参数                    |
|---------------------------------------|---------|-----------------------|
| `Filecoin.WalletBalance`              | 查询余额    | address               |
| `Filecoin.MpoolPush`                  | 推送消息    | SignedMessage         |
| `Filecoin.MpoolGetNonce`              | 获取nonce | address               |
| `Filecoin.StateWaitMsg`               | 等待确认    | cid, confidence       |
| `Filecoin.GasEstimateMessageGas`      | 估算Gas   | Message, spec, tipset |
| `Filecoin.StateMinerInfo`             | 矿工信息    | miner_id, tipset      |
| `Filecoin.StateMinerAvailableBalance` | 可用余额    | miner_id, tipset      |
| `Filecoin.StateMarketBalance`         | 市场余额    | address, tipset       |
| `Filecoin.ChainHead`                  | 链头      | -                     |

---

## 签名算法

### secp256k1

- 椭圆曲线数字签名
- 地址前缀: f1 (主网) / t1 (测试网)

### BLS

- Boneh-Lynn-Shacham 签名
- 地址前缀: f3 (主网) / t3 (测试网)

---

## 配置说明

```toml
[Lotus]
Host = "https://api.node.glif.io/rpc/v0"  # RPC节点地址
Token = ""                                  # 认证Token

[Database]
Host = "127.0.0.1"
Port = 3306
User = "root"
Password = ""
Name = "lotus_sign"
```

