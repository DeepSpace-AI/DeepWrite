## Auth & User Store 使用指南

### 概述
新增两个 Pinia stores 用于集中管理认证和用户信息：
- `useAuthStore()` - 管理 access token、refresh token 和认证状态
- `useUserStore()` - 管理当前登录用户信息

### Auth Store (`stores/auth.ts`)

```typescript
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()

// 读取状态
authStore.accessToken        // string - 当前访问令牌
authStore.refreshToken       // string - 当前刷新令牌
authStore.isAuthenticated    // boolean - 是否已登录

// 方法
authStore.setTokens(access, refresh)       // 设置令牌对
authStore.setAccessToken(token)            // 仅更新访问令牌
authStore.clearTokens()                    // 清除所有令牌
authStore.initializeFromStorage()          // 从 localStorage 恢复令牌（应用启动时自动调用）
```

#### 示例：检查登录状态
```vue
<script setup>
import { useAuthStore } from '@/stores/auth'

const authStore = useAuthStore()
</script>

<template>
  <div v-if="authStore.isAuthenticated" class="user-authenticated">
    已登录
  </div>
  <div v-else class="user-not-authenticated">
    未登录
  </div>
</template>
```

### User Store (`stores/user.ts`)

```typescript
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()

// 读取状态
userStore.user    // User | null - 当前用户对象

// 方法
userStore.setUser(userData)         // 设置用户信息
userStore.clearUser()               // 清除用户信息
userStore.initializeFromStorage()   // 从 localStorage 恢复用户信息（应用启动时自动调用）
```

#### User 数据结构
```typescript
interface User {
  id: string
  email: string
  displayName: string
  role: string
  status: string
}
```

#### 示例：显示用户信息
```vue
<script setup>
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
</script>

<template>
  <div v-if="userStore.user" class="user-card">
    <p>欢迎，{{ userStore.user.displayName }}</p>
    <p>邮箱：{{ userStore.user.email }}</p>
    <p>角色：{{ userStore.user.role }}</p>
  </div>
</template>
```

### 认证流程与 Stores 集成

#### 登录后（使用 api/auth.ts 的 saveSession）
```typescript
import { login, saveSession } from '@/api/auth'

const data = await login({ email, password })
// saveSession 自动调用 authStore.setTokens() 和 userStore.setUser()
saveSession(data)
```

#### 登出（使用 api/auth.ts 的 logout）
```typescript
import { logout } from '@/api/auth'

function handleLogout() {
  logout()  // 自动清除 authStore 和 userStore 中的所有信息
  router.push('/login')
}
```

### HTTP 拦截器中的自动集成

`api/http.ts` 已配置为自动使用 stores：

1. **请求拦截器** - 自动从 `authStore.accessToken` 读取令牌并添加到请求头
2. **响应拦截器** - 当遇到 401 错误时：
   - 自动刷新令牌（调用 `authStore.setAccessToken()`）
   - 重试原始请求
   - 如果认证完全失败，调用 `userStore.clearUser()` 和 `clearTokens()`

### 本地存储同步

所有 store 变更自动同步到浏览器存储：
- Token 存储在 `localStorage`（持久化）
- 用户信息存储在 `localStorage`（持久化）
- 应用启动时自动从存储恢复，保证页面刷新后状态保留

### 在受保护的路由中使用

```typescript
// router/index.ts
import { useAuthStore } from '@/stores/auth'

router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()
  
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next({ name: 'login', query: { redirect: to.path } })
  } else {
    next()
  }
})
```

### 最佳实践

1. **使用 stores 读取数据** - 而不是直接访问 localStorage
2. **通过 api/auth.ts 管理会话** - 而不是直接调用 store 方法
3. **响应式更新** - stores 是响应式的，任何变更都会自动触发组件重新渲染
4. **避免重复调用** - API 拦截器已处理令牌刷新和存储，无需手动干预
