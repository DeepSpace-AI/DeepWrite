# 对话标题生成故障排查方案

## 诊断步骤

### 1. 检查日志输出

重启 gateway 服务后，发送一条新消息，观察日志中的以下关键点：

```bash
# 查看 gateway 日志
# 应该看到类似以下的日志序列：

[TitleTask] Submitting task for session xxx, model: xxx
[TitleTask] Task submitted successfully for session xxx (queue size: 0)
[TitleWorker] Starting title generation for session xxx
[TitleWorker] Found X messages for session xxx
[SystemConfig] Using configured model 'xxx' (provider: xxx)
[TitleWorker] Using model: xxx (provider: xxx, base_url: xxx)
[GenerateTitle] Request URL: xxx
[GenerateTitle] Request Body: {...}
[GenerateTitle] Response Status: 200
[GenerateTitle] Response Body: {...}
[GenerateTitle] Generated title: xxx
[TitleWorker] SUCCESS - Updated title for session xxx: xxx
```

### 2. 检查关键日志标识

#### 2.1 任务提交失败
```
[TitleTask] Task queue full (100), skipping title generation
```
**原因**：队列已满  
**解决**：重启 gateway 服务清空队列

#### 2.2 没有消息
```
[TitleWorker] No messages found for session xxx
```
**原因**：消息未保存到数据库  
**解决**：检查消息保存逻辑

#### 2.3 模型未配置
```
[SystemConfig] No default model configured, using fallback model
```
**原因**：系统配置表中没有配置默认模型  
**解决**：
- 访问 Admin 端 `/system` 页面
- 在 "AI 模型配置" 区域配置 `default_ai_model`
- 点击 "保存 AI 配置"

#### 2.4 模型 API 错误
```
[GenerateTitle] AI service error: 400 - {...}
```
**原因**：模型 API 返回错误  
**解决**：
- 检查 `[GenerateTitle] Request Body` 日志，查看发送的请求
- 检查 `[GenerateTitle] Response Body` 日志，查看错误详情
- 根据错误信息调整请求格式或更换模型

#### 2.5 数据库更新失败
```
[TitleWorker] Failed to update title for session xxx: xxx
```
**原因**：数据库操作失败  
**解决**：检查数据库连接和权限

### 3. 数据库检查

#### 3.1 检查系统配置表
```sql
SELECT * FROM system_configs;
```

期望结果：
```
key                   | value
----------------------+------------------
default_ai_model      | gpt-4
title_generation_model| (可为空)
```

如果没有记录，访问 Admin 端配置页面添加。

#### 3.2 检查会话标题
```sql
SELECT id, title, created_at FROM sessions ORDER BY created_at DESC LIMIT 5;
```

#### 3.3 检查消息是否存在
```sql
SELECT session_id, role, substring(content, 1, 50) as content_preview 
FROM messages 
WHERE session_id = 'your-session-id' 
ORDER BY created_at;
```

### 4. 模型配置检查

#### 4.1 检查可用模型
```sql
SELECT model, provider, base_url, enabled 
FROM dw_ai_models 
WHERE enabled = true;
```

#### 4.2 确认模型配置正确
- `model`: 模型标识（如 `gpt-4`）
- `request_model`: 实际请求使用的模型名
- `base_url`: API 端点
- `enabled`: 必须为 `true`

### 5. 手动测试模型 API

使用 curl 测试模型是否正常工作：

```bash
curl -X POST https://your-api-endpoint/chat/completions \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -d '{
    "model": "your-model",
    "messages": [
      {"role": "user", "content": "Say hello"}
    ],
    "max_tokens": 10
  }'
```

### 6. 完整排查流程图

```
开始
  ↓
检查日志是否有 [TitleTask] 提交记录
  ├─ 无 → 检查 handleAgentStreamResponse 是否调用 submitTitleGenerationTask
  └─ 有 ↓
检查日志是否有 [TitleWorker] 开始处理记录
  ├─ 无 → Worker 未启动，检查 init() 函数
  └─ 有 ↓
检查日志是否有 "Found X messages"
  ├─ 无 → 检查消息保存逻辑
  └─ 有 ↓
检查日志是否有模型配置信息
  ├─ 无 → 检查 system_configs 表
  └─ 有 ↓
检查 [GenerateTitle] Response Status
  ├─ 4xx/5xx → 检查请求格式和模型配置
  └─ 200 ↓
检查数据库 sessions 表的 title 字段
  ├─ 为空 → 检查 UpdateSessionTitle 函数
  └─ 有值 → 成功！
```

### 7. 常见问题解决方案

#### 问题 1：chat_template 错误
```
"message":"As of transformers v4.44, default chat template is no longer allowed"
```
**解决**：代码已添加 chat_template 参数，重启服务即可

#### 问题 2：使用回退标题
```
[TitleWorker] Using simple title: xxx
```
**说明**：AI 生成失败，使用了用户消息的前 50 字符作为标题

#### 问题 3：队列满
```
[TitleTask] Task queue full
```
**解决**：
- 重启服务
- 或增加队列大小（修改代码中的 `titleTaskChan` 容量）

### 8. 增强调试模式

如果需要更详细的调试信息，可以临时修改日志级别：

```go
// 在 generateSessionTitle 函数开头添加
log.Printf("[GenerateTitle] DEBUG - Model: %+v", model)
log.Printf("[GenerateTitle] DEBUG - Messages count: %d", len(messages))
log.Printf("[GenerateTitle] DEBUG - Conversation text length: %d", conversationText.Len())
```

## 检查清单

- [ ] Gateway 服务已重启
- [ ] Admin 端已配置默认 AI 模型
- [ ] 日志中有 `[TitleTask]` 提交记录
- [ ] 日志中有 `[TitleWorker]` 处理记录
- [ ] 日志中有 `[GenerateTitle]` 请求/响应记录
- [ ] 数据库 sessions 表有 title 更新
- [ ] 前端会话列表显示标题