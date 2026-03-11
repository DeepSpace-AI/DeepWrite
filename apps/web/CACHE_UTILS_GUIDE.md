## Cache Utils 使用指南

### 概述

`utils/cache.ts` 提供了一个 `CacheManager` 类，支持以下特性：

- **过期时间控制** - 可为每项缓存设置 TTL（Time To Live）
- **开发环境加密** - 在 `import.meta.env.DEV` 为 true 时自动使用 Base64 加密存储
- **自动过期检查** - 读取时自动检查是否过期，过期数据自动删除
- **错误处理** - 解密或解析失败时安全降级处理

### 核心 API

```typescript
import { cacheManager, CACHE_TTL } from '@/utils/cache'

// 设置缓存（支持 TTL）
cacheManager.set(key, value, ttl)

// 获取缓存（自动检查过期）
const data = cacheManager.get<T>(key)

// 删除缓存
cacheManager.remove(key)

// 批量删除
cacheManager.removeMultiple([key1, key2])

// 清空所有缓存
cacheManager.clear()
```

### 预定义 TTL 常量

```typescript
// 所有 TTL 单位为毫秒
CACHE_TTL.FIVE_MINUTES      // 5 分钟
CACHE_TTL.FIFTEEN_MINUTES   // 15 分钟
CACHE_TTL.ONE_HOUR          // 1 小时
CACHE_TTL.SIX_HOURS         // 6 小时
CACHE_TTL.ONE_DAY           // 24 小时（默认用于 Token）
CACHE_TTL.SEVEN_DAYS        // 7 天
```

### 在 Auth Store 中的使用

auth store 已配置为使用 cache utils 存储令牌：

```typescript
// stores/auth.ts
import { cacheManager, CACHE_TTL } from '@/utils/cache'

const TOKEN_TTL = CACHE_TTL.ONE_DAY  // Token 有效期：24 小时

// setTokens() 方法自动使用 cache utils
authStore.setTokens(access, refresh)  // 自动加密保存（开发环境）

// initializeFromStorage() 自动从缓存恢复令牌
authStore.initializeFromStorage()  // 检查过期时间
```

#### Token 缓存的生命周期

1. **登录成功** → `saveSession()` → `authStore.setTokens()` → 写入缓存（TTL: 24h）
2. **页面刷新** → `authStore.initializeFromStorage()` → 从缓存读取令牌
3. **Token 过期** → 自动刷新 → `authStore.setAccessToken()` → 更新缓存
4. **登出** → `authStore.clearTokens()` → 删除缓存

### 在 User Store 中的使用

user store 已配置为使用 cache utils 存储用户信息：

```typescript
// stores/user.ts
const USER_TTL = CACHE_TTL.ONE_DAY  // 用户信息有效期：24 小时

// setUser() 方法自动加密保存用户对象
userStore.setUser(userData)

// initializeFromStorage() 自动从缓存恢复用户信息
userStore.initializeFromStorage()  // 检查过期和数据完整性
```

#### User 缓存的生命周期

1. **登录成功** → `saveSession()` → `userStore.setUser()` → 写入缓存（TTL: 24h）
2. **页面刷新** → `userStore.initializeFromStorage()` → 从缓存读取用户信息
3. **登出** → `userStore.clearUser()` → 删除缓存

### 在 HTTP 拦截器中的使用

http 客户端已配置自动从 cache utils 读取令牌：

```typescript
// api/http.ts 已集成 cache utils

// 请求拦截器
// 从 authStore 或 cache utils 读取 token，自动添加到请求头

// 响应拦截器
// Token 刷新时自动调用 cacheManager.set() 更新缓存
```

### 开发环境与生产环境的行为

#### 开发环境（`import.meta.env.DEV === true`）

- 缓存数据使用 Base64 加密存储
- localStorage 中的内容为加密后的 JSON 字符串
- **示例**：
  ```
  // 原始数据
  { "value": "token_abc123...", "expiresAt": 1234567890 }
  
  // 存储后（加密）
  "eyJ2YWx1ZSI6InRva2VuX2FiYzEyMyIsImV4cGlyZXNBdCI6MTIzNDU2Nzg5MH0="
  ```

#### 生产环境（`import.meta.env.DEV === false`）

- 缓存数据直接以未加密的 JSON 存储
- localStorage 中的内容为明文 JSON
- **注意**：生产环境中 token 仍然受到浏览器的同源策略和 HTTPS 保护

### 常见场景

#### 场景 1：用户登录

```typescript
// LoginView.vue
const { login, saveSession } = '@/api/auth'

async function onSubmit() {
  const data = await login({ email, password })
  saveSession(data)  // 自动保存 token 和用户信息到 cache utils
  router.push({ name: 'home' })
}
```

#### 场景 2：页面刷新后保持登录状态

```typescript
// main.ts 中自动调用
const authStore = useAuthStore()
const userStore = useUserStore()
authStore.initializeFromStorage()  // 从缓存恢复 token
userStore.initializeFromStorage()  // 从缓存恢复用户信息
// 如果 token 未过期，用户状态自动恢复
```

#### 场景 3：自定义缓存管理

```typescript
// 如需在其他地方直接使用 cache utils
import { cacheManager, CACHE_TTL } from '@/utils/cache'

// 缓存用户偏好（1 小时有效期）
cacheManager.set('user_prefs', { theme: 'dark' }, CACHE_TTL.ONE_HOUR)

// 读取（自动检查过期）
const prefs = cacheManager.get<{ theme: string }>('user_prefs')
```

#### 场景 4：登出时清理缓存

```typescript
// api/auth.ts
import { logout } from '@/api/auth'

function handleLogout() {
  logout()  // 自动清除 auth 和 user cache
  router.push('/login')
}
```

### 性能考虑

- **加密开销**: 在开发环境中，Base64 加密/解密是轻量级操作，对性能影响可忽略
- **缓存大小**: localStorage 限制通常为 5-10MB，当前使用场景远低于此限制
- **读取速度**: 本地 localStorage 读取速度 < 1ms，无需担忧

### 最佳实践

1. **使用预定义常量** - 而不是手动计算 TTL
   ```typescript
   // ✅ 好
   cacheManager.set(key, value, CACHE_TTL.ONE_DAY)
   
   // ❌ 避免
   cacheManager.set(key, value, 24 * 60 * 60 * 1000)
   ```

2. **通过 Store 访问** - 而不是直接调用 cache utils
   ```typescript
   // ✅ 推荐
   authStore.setTokens(access, refresh)
   userStore.setUser(userData)
   
   // ❌ 避免
   cacheManager.set('token', value)
   ```

3. **处理过期数据** - cache utils 自动删除过期项
   ```typescript
   const data = cacheManager.get<T>(key)
   if (!data) {
     // 数据已过期或不存在
     // 自动重新认证或刷新
   }
   ```

4. **不缓存敏感数据** - 除非在 HTTPS 环境
   ```typescript
   // ❌ 不要这样做（信用卡号码等）
   cacheManager.set('card_number', '4111-1111-1111-1111')
   
   // 仅缓存 Token 和非敏感用户信息
   ```

### 调试和监控

#### 在浏览器 DevTools 中检查缓存

```javascript
// 开发环境（加密后）
localStorage.getItem('deepwrite_access_token')
// Output: "eyJ2YWx1ZSI6InRva2VuXzEyMyIsImV4cGlyZXNBdCI6MTIzNDU2Nzg5MH0="

// 解密（仅开发环境）
atob("eyJ2YWx1ZSI6InRva2VuXzEyMyIsImV4cGlyZXNBdCI6MTIzNDU2Nzg5MH0=")
// Output: {"value":"token_123","expiresAt":1234567890}
```

#### 清除缓存用于测试

```typescript
// 清除所有缓存
cacheManager.clear()

// 或清除特定缓存
cacheManager.removeMultiple(['deepwrite_access_token', 'deepwrite_refresh_token', 'deepwrite_user'])
```

### 故障排查

| 问题 | 原因 | 解决方案 |
|------|------|--------|
| Token 意外被清除 | 已过期（24h）| 无需干预，自动刷新机制会处理 |
| 登出后仍能访问资源 | 缓存未清除或路由未保护 | 检查 `logout()` 是否调用，路由是否添加守卫 |
| 页面刷新后未保持登录 | Token 已过期或缓存损坏 | 检查浏览器 storage，清除数据重新登录 |
| 开发环境缓存显示乱码 | 正常，这是 Base64 编码 | 使用 DevTools 的 atob() 解码查看 |
