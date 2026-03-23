<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import IconFormatListBulleted from '~icons/mdi/format-list-bulleted'
import IconAccountGroupOutline from '~icons/mdi/account-group-outline'
import IconFileDocumentOutline from '~icons/mdi/file-document-outline'
import IconPencilOutline from '~icons/mdi/pencil-outline'
import IconSchoolOutline from '~icons/mdi/school-outline'
import IconCheck from '~icons/mdi/check'
import IconClose from '~icons/mdi/close'

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
  <section id="demo" class="dot-grid relative overflow-hidden py-20 lg:py-28">
    <div class="glow-blob -top-48 -right-48 size-[30rem] bg-[var(--glow-primary)] opacity-20" />
    <div class="glow-blob -bottom-36 -left-36 size-80 bg-[var(--glow-secondary)] opacity-15" />

    <div class="relative max-w-6xl mx-auto px-4 grid items-center gap-14 lg:grid-cols-[1.05fr_0.95fr]">
      <div class="max-w-2xl">
        <div class="label-sm py-3">
          <span>Issue 03</span>
          <span class="mx-3 opacity-50">·</span>
          <span>Collaborative Writing Systems</span>
          <span class="mx-3 opacity-50">·</span>
          <span>2026 春季刊</span>
        </div>
        <div class="mt-8 max-w-xl">
          <p class="label-sm mb-4">编者按</p>
          <h1 class="text-editorial text-5xl lg:text-6xl font-bold leading-[1.05] mb-6">
            为长期写作而设计的
            <span class="text-pretty-secondary">协作编辑体验</span>
          </h1>
          <p class="body-lg text-pretty-secondary leading-8 mb-6 max-w-lg">
            DeepWrite 将实时同步、版本追踪与结构化编辑整合进一套稳定的写作环境。
            它更接近研究笔记、编辑部流程与论文合著，而不是一次性发布页面工具。
          </p>
          <blockquote class="paper-quote body-md max-w-md">
            让协作像在同一张纸面上批注与修订，保持节奏、上下文与版面秩序。
          </blockquote>
        </div>
        <div class="mt-8 flex flex-wrap gap-3">
          <a href="#pricing" class="btn-primary-vellum btn-lg rounded-lg px-8">开始试用</a>
          <a href="#integrations" class="btn-tertiary btn-lg rounded-lg px-6">阅读文档</a>
        </div>
        <div class="mt-8 flex flex-wrap gap-2">
          <span class="label-sm px-3 py-1.5 rounded-lg bg-[var(--surface-raised)]">Y.js</span>
          <span class="label-sm px-3 py-1.5 rounded-lg bg-[var(--surface-raised)]">TipTap</span>
          <span class="label-sm px-3 py-1.5 rounded-lg bg-[var(--surface-raised)]">WebSocket</span>
          <span class="label-sm px-3 py-1.5 rounded-lg bg-[var(--surface-raised)]">CRDT</span>
          <span class="label-sm px-3 py-1.5 rounded-lg bg-[var(--surface-raised)]">Go</span>
        </div>
      </div>

      <div class="w-full max-w-xl justify-self-end">
        <div class="paper-card overflow-hidden rounded-lg">
          <div class="flex items-center gap-2 bg-[var(--surface-raised)] px-4 py-3">
            <div class="flex gap-1.5">
              <div class="w-3 h-3 rounded-full bg-red-400/40"></div>
              <div class="w-3 h-3 rounded-full bg-yellow-400/40"></div>
              <div class="w-3 h-3 rounded-full bg-emerald-500/40"></div>
            </div>
            <div class="flex-1 flex justify-center">
              <div class="flex h-5 w-44 items-center justify-center rounded-md bg-[var(--surface-base)]/70">
                <span class="text-[10px] text-pretty-muted">q3-roadmap.md</span>
              </div>
            </div>
          </div>
          <div class="flex items-center gap-0.5 px-3 py-1.5 bg-[var(--surface-base)]/50">
            <button class="btn-tertiary h-8 w-8 rounded-md text-sm font-bold">B</button>
            <button class="btn-tertiary h-8 w-8 rounded-md text-sm italic text-pretty-muted">I</button>
            <button class="btn-tertiary h-8 w-8 rounded-md text-sm underline">U</button>
            <div class="mx-2 h-5 w-3 rounded-sm bg-[var(--surface-overlay)]"></div>
            <button class="btn-tertiary h-8 w-8 rounded-md text-sm">H1</button>
            <button class="btn-tertiary h-8 w-8 rounded-md text-sm">H2</button>
            <button class="btn-tertiary h-8 w-8 rounded-md text-sm">
              <IconFormatListBulleted class="size-4" />
            </button>
          </div>
          <div class="flex items-center gap-2 px-4 py-2 bg-[var(--glow-primary)]/5">
            <div class="flex -space-x-2">
              <div class="size-6 rounded-full bg-[var(--color-primary)] text-white text-[10px] font-semibold flex items-center justify-center ring-2 ring-[var(--surface-base)]">A</div>
              <div class="size-6 rounded-full bg-[var(--color-secondary)] text-white text-[10px] font-semibold flex items-center justify-center ring-2 ring-[var(--surface-base)]">B</div>
              <div class="size-6 rounded-full bg-[var(--color-primary)]/70 text-white text-[10px] font-semibold flex items-center justify-center ring-2 ring-[var(--surface-base)]">C</div>
            </div>
            <span class="text-xs text-pretty-muted">3 人在线协作</span>
            <div class="ml-auto flex items-center gap-1.5">
              <span class="h-2 w-2 rounded-full bg-emerald-500"></span>
              <span class="text-xs text-emerald-600 font-medium">已同步</span>
            </div>
          </div>
          <div class="bg-[var(--surface-base)]/50 px-5 py-2">
            <span class="label-sm opacity-60">Draft / Shared Manuscript / 12 revisions</span>
          </div>
          <div class="p-5 min-h-60 bg-[var(--surface-base)]">
            <pre class="font-serif text-sm text-pretty-secondary leading-relaxed whitespace-pre-wrap">{{ displayed }}<span class="cursor-blink text-pretty-secondary">▎</span></pre>
          </div>
        </div>
      </div>
    </div>
  </section>

  <section class="bg-[var(--surface-raised)] py-12">
    <div class="max-w-5xl mx-auto px-4">
      <div class="grid grid-cols-2 md:grid-cols-4">
        <div class="text-center px-6 py-2">
          <div class="text-4xl font-bold tracking-tight text-editorial">50ms</div>
          <div class="text-sm mt-1.5 text-pretty-muted">平均同步延迟</div>
        </div>
        <div class="text-center px-6 py-2">
          <div class="text-4xl font-bold tracking-tight text-editorial">∞</div>
          <div class="text-sm mt-1.5 text-pretty-muted">协作人数上限</div>
        </div>
        <div class="text-center px-6 py-2">
          <div class="text-4xl font-bold tracking-tight text-editorial">99.9%</div>
          <div class="text-sm mt-1.5 text-pretty-muted">服务可用性</div>
        </div>
        <div class="text-center px-6 py-2">
          <div class="text-4xl font-bold tracking-tight text-editorial">6+</div>
          <div class="text-sm mt-1.5 text-pretty-muted">接入方式</div>
        </div>
      </div>
    </div>
  </section>

  <section class="py-24 bg-[var(--surface-base)]">
    <div class="max-w-6xl mx-auto px-4">
      <div class="mb-14">
        <p class="label-sm mb-3">01 / Use Cases</p>
        <h2 class="headline-lg text-pretty mb-3">适用场景</h2>
        <p class="body-lg text-pretty-secondary max-w-lg">无论是团队协作还是个人写作，DeepWrite 都能成为你的写作核心</p>
      </div>
      <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-5">
        <div class="surface-card group p-6 transition-all duration-300 hover:-translate-y-1">
          <div class="size-12 rounded-xl bg-[var(--glow-primary)]/20 text-pretty flex items-center justify-center mb-4 transition-colors duration-200 group-hover:bg-[var(--color-primary)] group-hover:text-white">
            <IconAccountGroupOutline class="size-6" />
          </div>
          <h3 class="font-semibold text-pretty mb-2 headline-md">团队协作</h3>
          <p class="body-md text-pretty-secondary">实时光标感知，多人同时在线，操作冲突自动合并，提交历史可追溯</p>
        </div>

        <div class="surface-card group p-6 transition-all duration-300 hover:-translate-y-1">
          <div class="size-12 rounded-xl bg-[var(--glow-secondary)]/20 text-pretty flex items-center justify-center mb-4 transition-colors duration-200 group-hover:bg-[var(--color-secondary)] group-hover:text-white">
            <IconFileDocumentOutline class="size-6" />
          </div>
          <h3 class="font-semibold text-pretty mb-2 headline-md">技术文档</h3>
          <p class="body-md text-pretty-secondary">代码块高亮、表格支持、Markdown 渲染，结构化内容清晰易维护</p>
        </div>

        <div class="surface-card group p-6 transition-all duration-300 hover:-translate-y-1">
          <div class="size-12 rounded-xl bg-[var(--glow-primary)]/20 text-pretty flex items-center justify-center mb-4 transition-colors duration-200 group-hover:bg-[var(--color-primary)] group-hover:text-white">
            <IconPencilOutline class="size-6" />
          </div>
          <h3 class="font-semibold text-pretty mb-2 headline-md">创意写作</h3>
          <p class="body-md text-pretty-secondary">流畅的编辑体验，专注模式消除干扰，让创作思路不被打断</p>
        </div>

        <div class="surface-card group p-6 transition-all duration-300 hover:-translate-y-1">
          <div class="size-12 rounded-xl bg-[var(--glow-secondary)]/20 text-pretty flex items-center justify-center mb-4 transition-colors duration-200 group-hover:bg-[var(--color-secondary)] group-hover:text-white">
            <IconSchoolOutline class="size-6" />
          </div>
          <h3 class="font-semibold text-pretty mb-2 headline-md">学术研究</h3>
          <p class="body-md text-pretty-secondary">批注讨论、草稿对比、版本管理，协同撰写学术论文更高效</p>
        </div>
      </div>
    </div>
  </section>

  <section id="integrations" class="py-24 bg-[var(--surface-raised)]">
    <div class="max-w-6xl mx-auto px-4">
      <div class="mb-14">
        <p class="label-sm mb-3">02 / Integrations</p>
        <h2 class="headline-lg text-pretty mb-3">开放集成</h2>
        <p class="body-lg text-pretty-secondary max-w-lg">通过标准协议无缝接入你的现有工作流</p>
      </div>
      <div class="grid grid-cols-1 md:grid-cols-3 gap-6">
        <div class="paper-card rounded-lg p-6">
          <span class="label-sm px-3 py-1.5 rounded-lg bg-[var(--color-primary)] text-white mb-4 inline-block">REST API</span>
          <h3 class="font-semibold text-pretty mb-2 headline-md">HTTP 接口</h3>
          <p class="body-md text-pretty-secondary mb-4">完整的 RESTful API，支持文档 CRUD、权限管理、用户鉴权，附交互式 Swagger 文档</p>
          <div class="rounded-md bg-[var(--surface-overlay)] p-3 font-mono text-xs">
            <pre data-prefix="GET" class="text-emerald-600"> /documents/:id</pre>
            <pre data-prefix="POST" class="text-blue-500"> /documents</pre>
            <pre data-prefix="PATCH" class="text-amber-500"> /documents/:id</pre>
            <pre data-prefix="DELETE" class="text-red-500"> /documents/:id</pre>
          </div>
          <div class="flex flex-wrap gap-1.5 mt-4">
            <span class="label-sm rounded-md bg-[var(--surface-overlay)] px-2 py-1">OpenAPI 3.0</span>
            <span class="label-sm rounded-md bg-[var(--surface-overlay)] px-2 py-1">JWT</span>
          </div>
        </div>

        <div class="paper-card rounded-lg p-6">
          <span class="label-sm px-3 py-1.5 rounded-lg bg-[var(--color-secondary)] text-white mb-4 inline-block">WebSocket</span>
          <h3 class="font-semibold text-pretty mb-2 headline-md">实时协作协议</h3>
          <p class="body-md text-pretty-secondary mb-4">基于 Y.js + WebSocket 的 CRDT 协同协议，原生支持离线编辑与自动合并</p>
          <div class="rounded-md bg-[var(--surface-overlay)] p-3 font-mono text-xs">
            <pre data-prefix="1" class="text-pretty-secondary"> const ydoc = new Y.Doc()</pre>
            <pre data-prefix="2" class="text-pretty-secondary"> const p = new</pre>
            <pre data-prefix="3" class="text-pretty-secondary">   GatewayProvider(</pre>
            <pre data-prefix="4" class="text-emerald-600">     url, docId, ydoc)</pre>
          </div>
          <div class="flex flex-wrap gap-1.5 mt-4">
            <span class="label-sm rounded-md bg-[var(--surface-overlay)] px-2 py-1">Y.js</span>
            <span class="label-sm rounded-md bg-[var(--surface-overlay)] px-2 py-1">CRDT</span>
          </div>
        </div>

        <div class="paper-card rounded-lg p-6">
          <span class="label-sm px-3 py-1.5 rounded-lg bg-[var(--color-primary)]/70 text-white mb-4 inline-block">Webhook</span>
          <h3 class="font-semibold text-pretty mb-2 headline-md">事件通知</h3>
          <p class="body-md text-pretty-secondary mb-4">文档变更、成员加入等事件实时推送，HMAC-SHA256 签名验证保障安全</p>
          <div class="rounded-md bg-[var(--surface-overlay)] p-3 font-mono text-xs">
            <pre data-prefix="{" class="text-blue-500"> "event": "doc.updated"</pre>
            <pre data-prefix="" class="text-pretty-secondary">  "doc_id": "c4d1f2e0"</pre>
            <pre data-prefix="" class="text-pretty-secondary">  "edited_by": "alice"</pre>
            <pre data-prefix="}" class="text-emerald-600"> "ts": 1741234567</pre>
          </div>
          <div class="flex flex-wrap gap-1.5 mt-4">
            <span class="label-sm rounded-md bg-[var(--surface-overlay)] px-2 py-1">事件驱动</span>
            <span class="label-sm rounded-md bg-[var(--surface-overlay)] px-2 py-1">HMAC-SHA256</span>
          </div>
        </div>
      </div>
    </div>
  </section>

  <section id="pricing" class="py-24 bg-[var(--surface-base)]">
    <div class="max-w-6xl mx-auto px-4">
      <div class="mb-14 text-center">
        <p class="label-sm mb-3">03 / Access Plans</p>
        <h2 class="headline-lg text-pretty mb-3">订阅方案</h2>
        <p class="mx-auto max-w-2xl body-lg text-pretty-secondary leading-8">
          更像期刊订阅与机构访问，而不是短期促销。根据写作规模、协作密度与部署要求选择合适方案。
        </p>
      </div>
      <div class="grid grid-cols-1 md:grid-cols-3 gap-6 max-w-5xl mx-auto items-start">
        <div class="paper-card rounded-lg p-6">
          <div class="label-sm">Free</div>
          <h3 class="text-xl font-bold text-pretty mt-2">免费版</h3>
          <p class="label-sm mt-1 opacity-60">个人试读权限</p>
          <div class="flex items-baseline gap-1 py-3 my-4">
            <span class="text-4xl font-bold text-pretty">¥0</span>
            <span class="text-pretty-muted text-sm">/ 月</span>
          </div>
          <p class="body-md text-pretty-secondary">适合个人用户入门体验</p>
          <ul class="space-y-3 text-sm pt-4 flex-1">
            <li class="flex items-center gap-2.5 text-pretty-secondary">
              <IconCheck class="size-4 text-[var(--color-secondary)] shrink-0" />
              最多 3 个文档
            </li>
            <li class="flex items-center gap-2.5 text-pretty-secondary">
              <IconCheck class="size-4 text-[var(--color-secondary)] shrink-0" />
              最多 2 人协作
            </li>
            <li class="flex items-center gap-2.5 text-pretty-secondary">
              <IconCheck class="size-4 text-[var(--color-secondary)] shrink-0" />
              基础富文本编辑
            </li>
            <li class="flex items-center gap-2.5 text-pretty-muted opacity-50">
              <IconClose class="size-4 shrink-0" />
              版本历史
            </li>
            <li class="flex items-center gap-2.5 text-pretty-muted opacity-50">
              <IconClose class="size-4 shrink-0" />
              API 访问
            </li>
          </ul>
          <button class="btn-tertiary w-full rounded-lg mt-6">立即开始</button>
        </div>

        <div class="paper-card relative rounded-lg p-6">
          <div class="absolute inset-x-0 top-0 h-1 rounded-t-lg bg-[var(--color-secondary)]"></div>
          <div class="absolute -top-4 inset-x-0 flex justify-center">
            <span class="label-sm px-3 py-1.5 rounded-lg bg-[var(--color-secondary)] text-white shadow-sm">编辑部推荐</span>
          </div>
          <div class="label-sm mt-2">Pro</div>
          <h3 class="text-xl font-bold text-pretty mt-2">专业版</h3>
          <p class="label-sm mt-1 opacity-60">协作编辑席位</p>
          <div class="flex items-baseline gap-1 py-3 my-4">
            <span class="text-4xl font-bold text-pretty">¥29</span>
            <span class="text-pretty-muted text-sm">/ 月</span>
          </div>
          <p class="body-md text-pretty-secondary">适合编辑团队、研究小组与内容驱动型产品</p>
          <ul class="space-y-3 text-sm pt-4 flex-1">
            <li class="flex items-center gap-2.5 text-pretty-secondary">
              <IconCheck class="size-4 shrink-0 text-[var(--color-secondary)]" />
              无限文档
            </li>
            <li class="flex items-center gap-2.5 text-pretty-secondary">
              <IconCheck class="size-4 shrink-0 text-[var(--color-secondary)]" />
              最多 10 人协作
            </li>
            <li class="flex items-center gap-2.5 text-pretty-secondary">
              <IconCheck class="size-4 shrink-0 text-[var(--color-secondary)]" />
              完整版本历史
            </li>
            <li class="flex items-center gap-2.5 text-pretty-secondary">
              <IconCheck class="size-4 shrink-0 text-[var(--color-secondary)]" />
              REST API 访问
            </li>
            <li class="flex items-center gap-2.5 text-pretty-secondary">
              <IconCheck class="size-4 shrink-0 text-[var(--color-secondary)]" />
              评论与批注
            </li>
          </ul>
          <button class="btn-primary-vellum w-full rounded-lg mt-6">选择专业版</button>
        </div>

        <div class="paper-card rounded-lg p-6">
          <div class="label-sm">Team</div>
          <h3 class="text-xl font-bold text-pretty mt-2">团队版</h3>
          <p class="label-sm mt-1 opacity-60">机构访问与部署</p>
          <div class="flex items-baseline gap-1 py-3 my-4">
            <span class="text-4xl font-bold text-pretty">¥99</span>
            <span class="text-pretty-muted text-sm">/ 月</span>
          </div>
          <p class="body-md text-pretty-secondary">适合大型团队与企业级部署</p>
          <ul class="space-y-3 text-sm pt-4 flex-1">
            <li class="flex items-center gap-2.5 text-pretty-secondary">
              <IconCheck class="size-4 text-[var(--color-secondary)] shrink-0" />
              无限文档与成员
            </li>
            <li class="flex items-center gap-2.5 text-pretty-secondary">
              <IconCheck class="size-4 text-[var(--color-secondary)] shrink-0" />
              Webhook 集成
            </li>
            <li class="flex items-center gap-2.5 text-pretty-secondary">
              <IconCheck class="size-4 text-[var(--color-secondary)] shrink-0" />
              自定义域名
            </li>
            <li class="flex items-center gap-2.5 text-pretty-secondary">
              <IconCheck class="size-4 text-[var(--color-secondary)] shrink-0" />
              优先技术支持
            </li>
            <li class="flex items-center gap-2.5 text-pretty-secondary">
              <IconCheck class="size-4 text-[var(--color-secondary)] shrink-0" />
              SLA 99.9% 保障
            </li>
          </ul>
          <button class="btn-tertiary w-full rounded-lg mt-6">联系销售</button>
        </div>
      </div>
    </div>
  </section>

  <section class="py-20 bg-[var(--surface-raised)]">
    <div class="text-center max-w-xl mx-auto px-4">
      <h2 class="headline-lg text-pretty mb-4">立即体验实时协作写作</h2>
      <p class="body-lg text-pretty-muted mb-8">免费版永久可用，无需绑定信用卡</p>
      <a href="#pricing" class="btn-secondary-vellum btn-lg rounded-full px-10">免费开始使用</a>
    </div>
  </section>

  <footer class="footer footer-center bg-[var(--surface-base)] py-8 text-pretty-muted text-sm">
    <div>
      <p>© 2026 DeepWrite · 基于 Y.js + TipTap 构建</p>
    </div>
  </footer>
</template>

<style scoped>
.cursor-blink {
  animation: blink 1s step-end infinite;
}

@keyframes blink {
  0%, 100% { opacity: 1; }
  50% { opacity: 0; }
}
</style>
