<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

// ── 打字演示 ──────────────────────────────────────────────
const demoContent = `# Q3 产品路线图

> 由产品团队协作编辑 · 最后更新 2 分钟前

## 核心目标

提升用户留存率至 **85%**，优化新用户引导流程，
并完成 AI 写作辅助功能的集成上线。

## 关键里程碑

- [x] 编辑器性能优化（已完成）
- [x] 多人实时光标
- [ ] AI 写作建议  ← Carol 正在编辑...
- [ ] 评论与批注系统`

const displayed = ref('')
let idx = 0
let timerId: ReturnType<typeof setTimeout>

function tick() {
  if (idx < demoContent.length) {
    displayed.value = demoContent.slice(0, idx + 1)
    idx++
    const ch = demoContent[idx - 1]
    const delay = ch === '\n' ? 120 : idx % 8 === 0 ? 70 : 28
    timerId = setTimeout(tick, delay)
  } else {
    timerId = setTimeout(() => {
      idx = 0
      displayed.value = ''
      timerId = setTimeout(tick, 600)
    }, 3500)
  }
}

onMounted(() => { timerId = setTimeout(tick, 800) })
onUnmounted(() => clearTimeout(timerId))
</script>

<template>
  <!-- ── Hero ─────────────────────────────────────────────── -->
  <section id="demo" class="dot-grid relative overflow-hidden py-20 lg:py-28">
    <!-- 光晕装饰 -->
    <div class="pointer-events-none absolute -top-48 -right-48 size-125 rounded-full bg-primary/5 blur-3xl" />
    <div class="pointer-events-none absolute -bottom-36 -left-36 size-80 rounded-full bg-secondary/5 blur-3xl" />

    <div class="relative max-w-6xl mx-auto px-4 grid items-center gap-14 lg:grid-cols-[1.05fr_0.95fr]">
      <!-- 左侧文案 -->
      <div class="max-w-2xl">
        <div class="flex flex-wrap items-center gap-x-4 gap-y-2 border-y border-base-300/80 py-3 text-[11px] font-mono uppercase tracking-[0.22em] text-base-content/45">
          <span>Issue 03</span>
          <span>Collaborative Writing Systems</span>
          <span>2026 春季刊</span>
        </div>
        <div class="mt-8 max-w-xl">
          <div class="text-xs font-mono uppercase tracking-[0.24em] text-primary/70 mb-4">编者按</div>
          <h1 class="heading-serif text-5xl lg:text-6xl font-bold leading-[1.05] text-base-content mb-6">
            为长期写作而设计的
            <span class="block text-primary">协作编辑体验</span>
          </h1>
          <p class="text-base text-base-content/64 leading-8 mb-6 max-w-lg">
            DeepWrite 将实时同步、版本追踪与结构化编辑整合进一套稳定的写作环境。
            它更接近研究笔记、编辑部流程与论文合著，而不是一次性发布页面工具。
          </p>
          <blockquote class="border-l-2 border-secondary/60 pl-4 text-sm leading-7 text-base-content/58 italic max-w-md">
            让协作像在同一张纸面上批注与修订，保持节奏、上下文与版面秩序。
          </blockquote>
        </div>
        <div class="mt-8 flex flex-wrap gap-3">
          <a href="#pricing" class="btn btn-primary btn-lg rounded-md px-7">开始试用</a>
          <a href="#integrations" class="btn btn-ghost btn-lg rounded-md border border-base-300">阅读文档</a>
        </div>
        <div class="mt-8 flex flex-wrap gap-2">
          <span class="badge badge-ghost text-xs font-mono">Y.js</span>
          <span class="badge badge-ghost text-xs font-mono">TipTap</span>
          <span class="badge badge-ghost text-xs font-mono">WebSocket</span>
          <span class="badge badge-ghost text-xs font-mono">CRDT</span>
          <span class="badge badge-ghost text-xs font-mono">Go</span>
        </div>
      </div>

      <!-- 右侧编辑器 mockup -->
      <div class="w-full max-w-xl justify-self-end">
        <div class="editor-sheet overflow-hidden rounded-sm border border-base-300 bg-base-100 shadow-2xl ring-1 ring-primary/10">
          <!-- 标题栏 -->
          <div class="flex items-center gap-2 border-b border-base-300 bg-base-200 px-4 py-3">
            <div class="flex gap-1.5">
              <div class="w-3 h-3 rounded-full bg-error/50"></div>
              <div class="w-3 h-3 rounded-full bg-warning/50"></div>
              <div class="w-3 h-3 rounded-full bg-success/50"></div>
            </div>
            <div class="flex-1 flex justify-center">
              <div class="flex h-5 w-44 items-center justify-center rounded-sm bg-base-300/70">
                <span class="text-[10px] text-base-content/40 font-mono">q3-roadmap.md</span>
              </div>
            </div>
          </div>
          <!-- 工具栏 -->
          <div class="flex items-center gap-0.5 px-3 py-1.5 border-b border-base-300/50 bg-base-100">
            <button class="btn btn-ghost btn-xs font-bold w-7 h-7 min-h-0 p-0">B</button>
            <button class="btn btn-ghost btn-xs italic text-base-content/60 w-7 h-7 min-h-0 p-0">I</button>
            <button class="btn btn-ghost btn-xs underline w-7 h-7 min-h-0 p-0">U</button>
            <div class="w-px h-4 bg-base-300 mx-1"></div>
            <button class="btn btn-ghost btn-xs text-xs w-7 h-7 min-h-0 p-0">H1</button>
            <button class="btn btn-ghost btn-xs text-xs w-7 h-7 min-h-0 p-0">H2</button>
            <button class="btn btn-ghost btn-xs w-7 h-7 min-h-0 p-0">
              <svg viewBox="0 0 16 16" fill="none" class="size-3.5" stroke="currentColor" stroke-width="1.5" stroke-linecap="round">
                <path d="M2 4h12M2 8h10M2 12h8" />
              </svg>
            </button>
          </div>
          <!-- 协作者状态栏 -->
          <div class="flex items-center gap-2 px-4 py-2 bg-primary/4 border-b border-primary/10">
            <div class="flex -space-x-2">
              <div class="size-6 rounded-full bg-primary text-primary-content text-[10px] font-semibold flex items-center justify-center ring-2 ring-base-100">A</div>
              <div class="size-6 rounded-full bg-secondary text-secondary-content text-[10px] font-semibold flex items-center justify-center ring-2 ring-base-100">B</div>
              <div class="size-6 rounded-full bg-accent text-accent-content text-[10px] font-semibold flex items-center justify-center ring-2 ring-base-100">C</div>
            </div>
            <span class="text-xs text-base-content/50">3 人在线协作</span>
            <div class="ml-auto flex items-center gap-1.5">
              <span class="status status-success"></span>
              <span class="text-xs text-success font-medium">已同步</span>
            </div>
          </div>
          <!-- 内容 -->
          <div class="border-b border-base-300/70 bg-base-100 px-5 py-2 text-[10px] font-mono uppercase tracking-[0.22em] text-base-content/35">
            Draft / Shared Manuscript / 12 revisions
          </div>
          <div class="p-5 min-h-60 bg-base-100">
            <pre class="font-mono text-sm text-base-content/80 leading-relaxed whitespace-pre-wrap wrap-break-word">{{ displayed }}<span class="cursor-blink text-primary">▎</span></pre>
          </div>
        </div>
      </div>
    </div>
  </section>

  <!-- ── 数字亮点（主色带）────────────────────────────────── -->
  <section class="bg-primary text-primary-content py-12">
    <div class="max-w-5xl mx-auto px-4">
      <div class="grid grid-cols-2 md:grid-cols-4 divide-x divide-primary-content/20">
        <div class="text-center px-6 py-2">
          <div class="text-4xl font-bold tracking-tight font-mono">50ms</div>
          <div class="text-sm mt-1.5 opacity-65">平均同步延迟</div>
        </div>
        <div class="text-center px-6 py-2">
          <div class="text-4xl font-bold tracking-tight font-mono">∞</div>
          <div class="text-sm mt-1.5 opacity-65">协作人数上限</div>
        </div>
        <div class="text-center px-6 py-2">
          <div class="text-4xl font-bold tracking-tight font-mono">99.9%</div>
          <div class="text-sm mt-1.5 opacity-65">服务可用性</div>
        </div>
        <div class="text-center px-6 py-2">
          <div class="text-4xl font-bold tracking-tight font-mono">6+</div>
          <div class="text-sm mt-1.5 opacity-65">接入方式</div>
        </div>
      </div>
    </div>
  </section>

  <!-- ── 适用场景 ───────────────────────────────────────────── -->
  <section class="py-24 bg-base-100">
    <div class="max-w-6xl mx-auto px-4">
      <div class="mb-14">
        <div class="text-xs text-primary/50 font-mono tracking-[0.2em] uppercase mb-3">01 / Use Cases</div>
        <h2 class="heading-serif text-3xl lg:text-4xl font-bold text-base-content mb-3">适用场景</h2>
        <p class="text-base-content/55 max-w-lg">无论是团队协作还是个人写作，DeepWrite 都能成为你的写作核心</p>
      </div>
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5">

        <!-- 团队协作 -->
        <div class="card bg-base-100 border border-base-300 card-lift group">
          <div class="card-body gap-4">
            <div class="size-10 rounded-lg bg-primary/10 text-primary flex items-center justify-center transition-colors duration-200 group-hover:bg-primary group-hover:text-primary-content">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" class="size-5">
                <path d="M15 19.128a9.38 9.38 0 0 0 2.625.372 9.337 9.337 0 0 0 4.121-.952 4.125 4.125 0 0 0-7.533-2.493M15 19.128v-.003c0-1.113-.285-2.16-.786-3.07M15 19.128v.106A12.318 12.318 0 0 1 8.624 21c-2.331 0-4.512-.645-6.374-1.766l-.001-.109a6.375 6.375 0 0 1 11.964-3.07M12 6.375a3.375 3.375 0 1 1-6.75 0 3.375 3.375 0 0 1 6.75 0Zm8.25 2.25a2.625 2.625 0 1 1-5.25 0 2.625 2.625 0 0 1 5.25 0Z" />
              </svg>
            </div>
            <h3 class="font-semibold text-base-content">团队协作</h3>
            <p class="text-sm text-base-content/55 leading-relaxed">实时光标感知，多人同时在线，操作冲突自动合并，提交历史可追溯</p>
          </div>
        </div>

        <!-- 技术文档 -->
        <div class="card bg-base-100 border border-base-300 card-lift group">
          <div class="card-body gap-4">
            <div class="size-10 rounded-lg bg-secondary/10 text-secondary flex items-center justify-center transition-colors duration-200 group-hover:bg-secondary group-hover:text-secondary-content">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" class="size-5">
                <path d="M19.5 14.25v-2.625a3.375 3.375 0 0 0-3.375-3.375h-1.5A1.125 1.125 0 0 1 13.5 7.125v-1.5a3.375 3.375 0 0 0-3.375-3.375H8.25m0 12.75h7.5m-7.5 3H12M10.5 2.25H5.625c-.621 0-1.125.504-1.125 1.125v17.25c0 .621.504 1.125 1.125 1.125h12.75c.621 0 1.125-.504 1.125-1.125V11.25a9 9 0 0 0-9-9Z" />
              </svg>
            </div>
            <h3 class="font-semibold text-base-content">技术文档</h3>
            <p class="text-sm text-base-content/55 leading-relaxed">代码块高亮、表格支持、Markdown 渲染，结构化内容清晰易维护</p>
          </div>
        </div>

        <!-- 创意写作 -->
        <div class="card bg-base-100 border border-base-300 card-lift group">
          <div class="card-body gap-4">
            <div class="size-10 rounded-lg bg-accent/10 text-accent flex items-center justify-center transition-colors duration-200 group-hover:bg-accent group-hover:text-accent-content">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" class="size-5">
                <path d="m16.862 4.487 1.687-1.688a1.875 1.875 0 1 1 2.652 2.652L10.582 16.07a4.5 4.5 0 0 1-1.897 1.13L6 18l.8-2.685a4.5 4.5 0 0 1 1.13-1.897l8.932-8.931Zm0 0L19.5 7.125M18 14v4.75A2.25 2.25 0 0 1 15.75 21H5.25A2.25 2.25 0 0 1 3 18.75V8.25A2.25 2.25 0 0 1 5.25 6H10" />
              </svg>
            </div>
            <h3 class="font-semibold text-base-content">创意写作</h3>
            <p class="text-sm text-base-content/55 leading-relaxed">流畅的编辑体验，专注模式消除干扰，让创作思路不被打断</p>
          </div>
        </div>

        <!-- 学术研究 -->
        <div class="card bg-base-100 border border-base-300 card-lift group">
          <div class="card-body gap-4">
            <div class="size-10 rounded-lg bg-primary/10 text-primary flex items-center justify-center transition-colors duration-200 group-hover:bg-primary group-hover:text-primary-content">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" class="size-5">
                <path d="M4.26 10.147a60.438 60.438 0 0 0-.491 6.347A48.62 48.62 0 0 1 12 20.904a48.62 48.62 0 0 1 8.232-4.41 60.46 60.46 0 0 0-.491-6.347m-15.482 0a50.636 50.636 0 0 0-2.658-.813A59.906 59.906 0 0 1 12 3.493a59.903 59.903 0 0 1 10.399 5.84c-.896.248-1.783.52-2.658.814m-15.482 0A50.717 50.717 0 0 1 12 13.489a50.702 50.702 0 0 1 3.741-1.342M6.75 15a.75.75 0 1 0 0-1.5.75.75 0 0 0 0 1.5Zm0 0v-3.675A55.378 55.378 0 0 1 12 8.443m-7.007 11.55A5.981 5.981 0 0 0 6.75 15.75v-1.5" />
              </svg>
            </div>
            <h3 class="font-semibold text-base-content">学术研究</h3>
            <p class="text-sm text-base-content/55 leading-relaxed">批注讨论、草稿对比、版本管理，协同撰写学术论文更高效</p>
          </div>
        </div>

      </div>
    </div>
  </section>

  <!-- ── 开放集成 ───────────────────────────────────────────── -->
  <section id="integrations" class="py-24 bg-base-200">
    <div class="max-w-6xl mx-auto px-4">
      <div class="mb-14">
        <div class="text-xs text-primary/50 font-mono tracking-[0.2em] uppercase mb-3">02 / Integrations</div>
        <h2 class="heading-serif text-3xl lg:text-4xl font-bold text-base-content mb-3">开放集成</h2>
        <p class="text-base-content/55 max-w-lg">通过标准协议无缝接入你的现有工作流</p>
      </div>
      <div class="grid grid-cols-1 md:grid-cols-3 gap-6">

        <!-- REST API -->
        <div class="card bg-base-100 border border-base-300 card-lift">
          <div class="card-body gap-4">
            <div class="badge badge-primary w-fit font-mono text-xs">REST API</div>
            <h3 class="font-semibold text-base-content">HTTP 接口</h3>
            <p class="text-sm text-base-content/55">完整的 RESTful API，支持文档 CRUD、权限管理、用户鉴权，附交互式 Swagger 文档</p>
            <div class="mockup-code rounded-lg text-xs">
              <pre data-prefix="GET"><code class="text-success"> /documents/:id</code></pre>
              <pre data-prefix="POST"><code class="text-info"> /documents</code></pre>
              <pre data-prefix="PATCH"><code class="text-warning"> /documents/:id</code></pre>
              <pre data-prefix="DELETE"><code class="text-error"> /documents/:id</code></pre>
            </div>
            <div class="flex flex-wrap gap-1.5">
              <span class="badge badge-outline badge-sm">OpenAPI 3.0</span>
              <span class="badge badge-outline badge-sm">JWT</span>
            </div>
          </div>
        </div>

        <!-- WebSocket -->
        <div class="card bg-base-100 border border-base-300 card-lift">
          <div class="card-body gap-4">
            <div class="badge badge-secondary w-fit font-mono text-xs">WebSocket</div>
            <h3 class="font-semibold text-base-content">实时协作协议</h3>
            <p class="text-sm text-base-content/55">基于 Y.js + WebSocket 的 CRDT 协同协议，原生支持离线编辑与自动合并</p>
            <div class="mockup-code rounded-lg text-xs">
              <pre data-prefix="1"><code> const ydoc = new Y.Doc()</code></pre>
              <pre data-prefix="2"><code> const p = new</code></pre>
              <pre data-prefix="3"><code>   GatewayProvider(</code></pre>
              <pre data-prefix="4"><code class="text-success">     url, docId, ydoc)</code></pre>
            </div>
            <div class="flex flex-wrap gap-1.5">
              <span class="badge badge-outline badge-sm">Y.js</span>
              <span class="badge badge-outline badge-sm">CRDT</span>
            </div>
          </div>
        </div>

        <!-- Webhook -->
        <div class="card bg-base-100 border border-base-300 card-lift">
          <div class="card-body gap-4">
            <div class="badge badge-accent w-fit font-mono text-xs">Webhook</div>
            <h3 class="font-semibold text-base-content">事件通知</h3>
            <p class="text-sm text-base-content/55">文档变更、成员加入等事件实时推送，HMAC-SHA256 签名验证保障安全</p>
            <div class="mockup-code rounded-lg text-xs">
              <pre data-prefix="{"><code class="text-info"> "event": "doc.updated"</code></pre>
              <pre data-prefix=""><code>  "doc_id": "c4d1f2e0"</code></pre>
              <pre data-prefix=""><code>  "edited_by": "alice"</code></pre>
              <pre data-prefix="}"><code class="text-success"> "ts": 1741234567</code></pre>
            </div>
            <div class="flex flex-wrap gap-1.5">
              <span class="badge badge-outline badge-sm">事件驱动</span>
              <span class="badge badge-outline badge-sm">HMAC-SHA256</span>
            </div>
          </div>
        </div>

      </div>
    </div>
  </section>

  <!-- ── 订阅方案 ───────────────────────────────────────────── -->
  <section id="pricing" class="py-24 bg-base-100">
    <div class="max-w-6xl mx-auto px-4">
      <div class="mb-14 text-center">
        <div class="text-xs text-primary/50 font-mono tracking-[0.2em] uppercase mb-3">03 / Access Plans</div>
        <h2 class="heading-serif text-3xl lg:text-4xl font-bold text-base-content mb-3">订阅方案</h2>
        <p class="mx-auto max-w-2xl text-base-content/55 leading-8">
          更像期刊订阅与机构访问，而不是短期促销。根据写作规模、协作密度与部署要求选择合适方案。
        </p>
      </div>
      <div class="grid grid-cols-1 md:grid-cols-3 gap-6 max-w-5xl mx-auto items-start">

        <!-- 免费版 -->
        <div class="card border border-base-300 bg-base-100 card-lift publication-card">
          <div class="card-body gap-3">
            <div class="text-xs font-mono text-base-content/35 uppercase tracking-widest">Free</div>
            <h3 class="text-xl font-bold text-base-content">免费版</h3>
            <p class="text-xs font-mono uppercase tracking-[0.18em] text-base-content/35">个人试读权限</p>
            <div class="flex items-baseline gap-1 py-3 border-y border-base-300 my-1">
              <span class="text-4xl font-bold text-base-content">¥0</span>
              <span class="text-base-content/40 text-sm">/ 月</span>
            </div>
            <p class="text-sm text-base-content/55">适合个人用户入门体验</p>
            <ul class="space-y-3 text-sm pt-2 flex-1">
              <li class="flex items-center gap-2.5">
                <svg viewBox="0 0 16 16" fill="none" class="size-4 text-success shrink-0" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M2.5 8l4 4 7-7"/></svg>
                最多 3 个文档
              </li>
              <li class="flex items-center gap-2.5">
                <svg viewBox="0 0 16 16" fill="none" class="size-4 text-success shrink-0" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M2.5 8l4 4 7-7"/></svg>
                最多 2 人协作
              </li>
              <li class="flex items-center gap-2.5">
                <svg viewBox="0 0 16 16" fill="none" class="size-4 text-success shrink-0" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M2.5 8l4 4 7-7"/></svg>
                基础富文本编辑
              </li>
              <li class="flex items-center gap-2.5 opacity-30">
                <svg viewBox="0 0 16 16" fill="none" class="size-4 shrink-0" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"><path d="M4 4l8 8M12 4l-8 8"/></svg>
                版本历史
              </li>
              <li class="flex items-center gap-2.5 opacity-30">
                <svg viewBox="0 0 16 16" fill="none" class="size-4 shrink-0" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"><path d="M4 4l8 8M12 4l-8 8"/></svg>
                API 访问
              </li>
            </ul>
            <div class="card-actions mt-5">
              <button class="btn btn-outline w-full rounded-md">立即开始</button>
            </div>
          </div>
        </div>

        <!-- 专业版（渐变高亮） -->
        <div class="card relative overflow-hidden border border-primary/35 bg-base-100 text-base-content shadow-xl card-lift publication-card publication-card-featured">
          <div class="absolute inset-x-0 top-0 h-1 bg-secondary"></div>
          <div class="absolute -top-4 inset-x-0 flex justify-center">
            <span class="badge badge-secondary text-secondary-content font-semibold shadow-sm">编辑部推荐</span>
          </div>
          <div class="card-body gap-3">
            <div class="text-xs font-mono uppercase tracking-widest text-base-content/35">Pro</div>
            <h3 class="text-xl font-bold">专业版</h3>
            <p class="text-xs font-mono uppercase tracking-[0.18em] text-base-content/35">协作编辑席位</p>
            <div class="flex items-baseline gap-1 py-3 border-y border-base-300 my-1">
              <span class="text-4xl font-bold">¥29</span>
              <span class="text-base-content/40 text-sm">/ 月</span>
            </div>
            <p class="text-sm text-base-content/55">适合编辑团队、研究小组与内容驱动型产品</p>
            <ul class="space-y-3 text-sm pt-2 flex-1">
              <li class="flex items-center gap-2.5">
                <svg viewBox="0 0 16 16" fill="none" class="size-4 shrink-0" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M2.5 8l4 4 7-7"/></svg>
                无限文档
              </li>
              <li class="flex items-center gap-2.5">
                <svg viewBox="0 0 16 16" fill="none" class="size-4 shrink-0" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M2.5 8l4 4 7-7"/></svg>
                最多 10 人协作
              </li>
              <li class="flex items-center gap-2.5">
                <svg viewBox="0 0 16 16" fill="none" class="size-4 shrink-0" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M2.5 8l4 4 7-7"/></svg>
                完整版本历史
              </li>
              <li class="flex items-center gap-2.5">
                <svg viewBox="0 0 16 16" fill="none" class="size-4 shrink-0" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M2.5 8l4 4 7-7"/></svg>
                REST API 访问
              </li>
              <li class="flex items-center gap-2.5">
                <svg viewBox="0 0 16 16" fill="none" class="size-4 shrink-0" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M2.5 8l4 4 7-7"/></svg>
                评论与批注
              </li>
            </ul>
            <div class="card-actions mt-5">
              <button class="btn btn-primary w-full rounded-md">选择专业版</button>
            </div>
          </div>
        </div>

        <!-- 团队版 -->
        <div class="card border border-base-300 bg-base-100 card-lift publication-card">
          <div class="card-body gap-3">
            <div class="text-xs font-mono text-base-content/35 uppercase tracking-widest">Team</div>
            <h3 class="text-xl font-bold text-base-content">团队版</h3>
            <p class="text-xs font-mono uppercase tracking-[0.18em] text-base-content/35">机构访问与部署</p>
            <div class="flex items-baseline gap-1 py-3 border-y border-base-300 my-1">
              <span class="text-4xl font-bold text-base-content">¥99</span>
              <span class="text-base-content/40 text-sm">/ 月</span>
            </div>
            <p class="text-sm text-base-content/55">适合大型团队与企业级部署</p>
            <ul class="space-y-3 text-sm pt-2 flex-1">
              <li class="flex items-center gap-2.5">
                <svg viewBox="0 0 16 16" fill="none" class="size-4 text-success shrink-0" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M2.5 8l4 4 7-7"/></svg>
                无限文档与成员
              </li>
              <li class="flex items-center gap-2.5">
                <svg viewBox="0 0 16 16" fill="none" class="size-4 text-success shrink-0" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M2.5 8l4 4 7-7"/></svg>
                Webhook 集成
              </li>
              <li class="flex items-center gap-2.5">
                <svg viewBox="0 0 16 16" fill="none" class="size-4 text-success shrink-0" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M2.5 8l4 4 7-7"/></svg>
                自定义域名
              </li>
              <li class="flex items-center gap-2.5">
                <svg viewBox="0 0 16 16" fill="none" class="size-4 text-success shrink-0" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M2.5 8l4 4 7-7"/></svg>
                优先技术支持
              </li>
              <li class="flex items-center gap-2.5">
                <svg viewBox="0 0 16 16" fill="none" class="size-4 text-success shrink-0" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M2.5 8l4 4 7-7"/></svg>
                SLA 99.9% 保障
              </li>
            </ul>
            <div class="card-actions mt-5">
              <button class="btn btn-outline w-full rounded-md">联系销售</button>
            </div>
          </div>
        </div>

      </div>
    </div>
  </section>

  <!-- ── CTA ──────────────────────────────────────────────── -->
  <section class="cta-gradient py-20">
    <div class="text-center max-w-xl mx-auto px-4 text-primary-content">
      <h2 class="heading-serif text-3xl lg:text-4xl font-bold mb-4">立即体验实时协作写作</h2>
      <p class="opacity-65 mb-8 text-base">免费版永久可用，无需绑定信用卡</p>
      <a href="#pricing" class="btn btn-secondary btn-lg rounded-full px-10 shadow-lg">免费开始使用</a>
    </div>
  </section>

  <!-- ── 页脚 ─────────────────────────────────────────────── -->
  <footer class="footer footer-center bg-base-100 border-t border-base-300/50 py-8 text-base-content/35 text-sm font-mono">
    <div>
      <p>© 2026 DeepWrite · 基于 Y.js + TipTap 构建</p>
    </div>
  </footer>
</template>

<style scoped>
@import url('https://fonts.googleapis.com/css2?family=Playfair+Display:ital,wght@0,600;0,700;1,600&display=swap');

.heading-serif {
  font-family: 'Playfair Display', Georgia, serif;
}

/* Hero 点阵背景 */
.dot-grid {
  background-color: var(--color-base-100);
  background-image: radial-gradient(
    circle,
    color-mix(in srgb, var(--color-base-content) 7%, transparent) 1.5px,
    transparent 1.5px
  );
  background-size: 28px 28px;
}

/* 卡片悬浮效果 */
.card-lift {
  transition: transform 0.2s ease, box-shadow 0.2s ease, border-color 0.2s ease;
}
.card-lift:hover {
  transform: translateY(-5px);
  box-shadow: 0 20px 48px color-mix(in srgb, var(--color-primary) 14%, transparent);
  border-color: color-mix(in srgb, var(--color-primary) 35%, transparent) !important;
}

/* 专业版渐变卡片 */
/* CTA 渐变背景 */
.cta-gradient {
  background: linear-gradient(
    135deg,
    var(--color-primary) 0%,
    color-mix(in oklch, var(--color-primary) 55%, var(--color-secondary)) 100%
  );
}

.editor-sheet {
  box-shadow:
    0 30px 70px color-mix(in srgb, var(--color-neutral) 12%, transparent),
    18px 18px 0 color-mix(in srgb, var(--color-base-300) 55%, transparent);
}

.publication-card {
  position: relative;
  border-radius: 0.35rem;
}

.publication-card::before {
  content: "";
  position: absolute;
  inset: 0;
  pointer-events: none;
  border-top: 1px solid color-mix(in srgb, var(--color-primary) 20%, transparent);
}

.publication-card-featured {
  background:
    linear-gradient(to bottom, color-mix(in srgb, var(--color-secondary) 6%, var(--color-base-100)), var(--color-base-100));
}

/* 光标闪烁 */
.cursor-blink {
  animation: blink 1s step-end infinite;
}

@keyframes blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0; }
}
</style>


