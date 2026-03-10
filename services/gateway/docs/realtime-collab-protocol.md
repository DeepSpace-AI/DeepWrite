# DeepWrite 实时协作协议（Collab v1）

本文档描述 DeepWrite 网关与 Web 客户端之间的实时协作二进制协议，用于同步文档内容（Yjs Update）与在线状态（Awareness）。

## 1. 传输层

- 协议：WebSocket（二进制帧）
- 路由：`/api/v1/documents/:id/collab/ws?token=...&name=...`
- 鉴权：`collab-token`（短期 token，支持单次使用）
- 心跳：服务端定时 ping，客户端需保持连接可读写

## 2. 编码规则

### 2.1 基础编码

- 整数：`varuint`（无符号可变长整数）
- 帧结构：
  - `messageType(varuint)`
  - 可选 `subtype(varuint)`
  - `payload(bytes)`

### 2.2 messageType

- `0`：Sync（文档状态同步）
- `1`：Awareness（在线用户/光标/选区状态）

### 2.3 Sync subtype

- `0`：`SyncStep1`
- `1`：`SyncStep2`
- `2`：`SyncUpdate`

## 3. 同步流程

### 3.1 连接建立

1. 客户端建立 WebSocket。
2. 服务端创建会话并将客户端加入文档房间。
3. 服务端发送 `SyncStep1`（请求客户端状态）。
4. 客户端收到 `SyncStep1` 后返回 `SyncStep2(全量状态 update)`。
5. 双方后续通过 `SyncUpdate` 增量同步。

### 3.2 房间回放

- 新客户端加入时，服务端会回放历史 `SyncUpdate`。
- 新客户端加入时，服务端会回放当前房间内已缓存的 Awareness 状态，确保在线用户与光标可见。

## 4. Awareness（在线状态）

### 4.1 载荷说明

- `payload` 使用 y-protocols awareness update 二进制格式。
- 典型字段：
  - `user.name`：显示名
  - `selection.anchor/head`：选区范围（用于光标/选区显示）

### 4.2 发送时机

- 客户端本地 awareness 变化时立即发送。
- 客户端定时发送 awareness 心跳（当前实现 15 秒）以保持在线状态与光标新鲜度。

### 4.3 服务端行为

- 服务端按“会话”维度缓存 awareness（不是按 userID 覆盖）。
- 服务端接收到 awareness 后向同房间其他客户端广播。

## 5. 权限约束

- 只读角色：可接收 Sync/Awareness，不可提交 `SyncStep2`/`SyncUpdate`。
- 可编辑角色：可发送文档更新与 awareness。

## 6. 落盘与审计

- `SyncUpdate` 会进入批量落盘队列。
- 达到阈值或定时器触发时写入 `collab_updates`。
- 会话结束时写入审计数据（更新量、输入字节、延迟 P95 等）。

## 7. 兼容与演进

- 当前版本：Collab v1（隐式版本，以 messageType/subtype 约定）
- 若后续扩展：
  - 新增 `messageType` 时保持旧类型语义不变。
  - 新增字段应向后兼容（旧客户端可忽略未知内容）。

## 8. 维护清单

每次修改协作协议时，请同时更新：

- 本文档（字段、流程、时序变化）
- `services/gateway/pkg/collab/protocol.go`
- `services/gateway/handler/collab_handler.go`
- `apps/web/src/collab/gatewayProvider.ts`
- 协作页面联调验证记录（至少双端在线、光标、断线重连）
