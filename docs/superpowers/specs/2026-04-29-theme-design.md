# DeepWrite 主题设计方案

## 概述

DeepWrite 是一款专为科研人员设计的 AI 辅助科研写作平台。本设计方案旨在为平台建立统一、专业的视觉风格，提升用户体验和品牌辨识度。

## 设计决策

| 决策项 | 选择 | 理由 |
|--------|------|------|
| 风格方向 | 学术简约风 | 类似 Notion/Typora，干净专注，适合长时间写作 |
| 主色调 | 紫色 (#7C3AED) | 代表创造力和智慧，有学术气质，辨识度高 |
| 动效风格 | 流畅自然 | 200-300ms 过渡，spring 弹性曲线，活泼不花哨 |
| 暗色模式 | 支持 | 减少视觉疲劳，适应不同环境 |
| 切换按钮 | 顶部导航栏 | 全局可见，符合主流设计习惯 |

## 色彩系统

### 亮色模式

| 用途 | 色值 | 说明 |
|------|------|------|
| 主色 | `#7C3AED` | 紫色 600，用于按钮、链接、强调 |
| 主色悬停 | `#6D28D9` | 紫色 700，交互状态 |
| 主色背景 | `#F5F3FF` | 紫色 50，浅色背景 |
| 页面背景 | `#FFFFFF` | 纯白 |
| 卡片背景 | `#FFFFFF` | 纯白 |
| 次要背景 | `#F9FAFB` | 灰色 50 |
| 边框 | `#E5E7EB` | 灰色 200 |
| 主要文字 | `#111827` | 灰色 900 |
| 次要文字 | `#6B7280` | 灰色 500 |
| 静音文字 | `#9CA3AF` | 灰色 400 |

### 暗色模式

| 用途 | 色值 | 说明 |
|------|------|------|
| 主色 | `#A78BFA` | 紫色 400，柔和版本 |
| 主色悬停 | `#C4B5FD` | 紫色 300 |
| 主色背景 | `#1E1B2E` | 深紫背景 |
| 页面背景 | `#0F0F11` | 近黑 |
| 卡片背景 | `#1C1C1E` | 深灰 |
| 次要背景 | `#18181B` | 灰色 900 |
| 边框 | `#27272A` | 灰色 800 |
| 主要文字 | `#F9FAFB` | 灰色 50 |
| 次要文字 | `#A1A1AA` | 灰色 400 |
| 静音文字 | `#71717A` | 灰色 500 |

## 组件规范

### 圆角

| 组件 | 圆角值 |
|------|--------|
| 按钮 | `8px` |
| 卡片 | `12px` |
| 输入框 | `8px` |
| 头像 | `50%` |
| 徽章 | `20px` |

### 阴影

| 用途 | 值 |
|------|-----|
| 卡片 | `0 1px 3px rgba(0,0,0,0.1)` |
| 悬浮 | `0 4px 12px rgba(124,58,237,0.4)` |
| 下拉菜单 | `0 10px 25px rgba(0,0,0,0.15)` |

### 动效

| 类型 | 时长 | 缓动函数 |
|------|------|----------|
| 颜色过渡 | `150ms` | `ease` |
| 位移动画 | `200ms` | `ease-out` |
| 弹性动画 | `300ms` | `cubic-bezier(0.34,1.56,0.64,1)` |
| 页面切换 | `250ms` | `ease-in-out` |

## 组件样式

### 按钮

```css
/* 主要按钮 */
.btn-primary {
  background: var(--accent);
  color: white;
  border-radius: 8px;
  padding: 10px 20px;
  font-weight: 500;
  transition: all 200ms ease;
}

.btn-primary:hover {
  background: var(--accent-hover);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(124,58,237,0.4);
}

/* 次要按钮 */
.btn-secondary {
  background: var(--bg-secondary);
  color: var(--text-primary);
  border: 1px solid var(--border);
}

.btn-secondary:hover {
  background: var(--accent-light);
  border-color: var(--accent);
  color: var(--accent);
}
```

### 输入框

```css
.input-field {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 10px 12px;
  transition: all 150ms ease;
}

.input-field:focus {
  outline: none;
  border-color: var(--accent);
  box-shadow: 0 0 0 3px var(--accent-lighter);
}
```

### 卡片

```css
.card {
  background: var(--bg-card);
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 20px;
  transition: all 200ms ease;
}

.card:hover {
  border-color: var(--accent);
  box-shadow: 0 4px 12px rgba(0,0,0,0.1);
}
```

## 布局规范

### 侧边栏

- 宽度：`240px`（展开）/ `64px`（收起）
- 背景：`var(--bg-sidebar)`
- 边框：`1px solid var(--border)` 右侧
- 过渡：`300ms ease`

### 导航栏

- 高度：`56px`
- 背景：`var(--bg-card)`
- 边框：`1px solid var(--border)` 底部
- 固定：`sticky top-0 z-30`

### 主内容区

- 内边距：`24px`
- 最大宽度：无限制（响应式）
- 背景：`var(--bg-secondary)`

## 主题切换实现

### 技术方案

- 使用 `next-themes` 库管理主题状态
- CSS 变量定义所有主题 token
- TailwindCSS v4 的 `@theme` 指令集成

### 切换按钮

- 位置：导航栏右侧
- 图标：太阳（亮色）/ 月亮（暗色）
- 样式：`36px` 圆角按钮

### 系统偏好

- 自动检测系统主题偏好
- 首次访问时跟随系统
- 用户手动切换后记住选择

## 文件结构

```
frontend/
├── app/
│   └── globals.css          # CSS 变量定义
├── components/
│   ├── layout/
│   │   ├── sidebar.tsx      # 侧边栏组件
│   │   ├── navbar.tsx       # 导航栏组件
│   │   └── main-layout.tsx  # 主布局
│   └── theme-toggle.tsx     # 主题切换按钮
├── lib/
│   └── utils.ts             # 工具函数
└── stores/
    └── auth.ts              # 认证状态（已不用于主题）
```

## 验收标准

1. **亮色模式**：所有页面正确显示紫色主题
2. **暗色模式**：所有页面正确显示暗色主题
3. **切换功能**：点击按钮即时切换，无闪烁
4. **持久化**：刷新后保持用户选择
5. **系统偏好**：首次访问跟随系统主题
6. **响应式**：移动端和桌面端均正常显示
7. **动画效果**：过渡流畅自然，无卡顿

## 实施计划

### Wave 1: 基础设施
- 安装 `next-themes`
- 配置 CSS 变量
- 创建主题切换组件

### Wave 2: 组件更新
- 更新侧边栏样式
- 更新导航栏样式
- 更新所有页面组件

### Wave 3: 动效优化
- 添加过渡动画
- 优化交互反馈
- 测试和调试

### Wave 4: 验证
- 手动测试所有页面
- 验证暗色模式
- 检查响应式布局
