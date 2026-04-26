# DeepWrite 贡献指南

感谢你对 DeepWrite 的关注！我们欢迎并鼓励任何形式的贡献，包括修复 Bug、增加新功能、改进文档或提出建议。

## 开发流程

1. **Fork 仓库**: 将项目 Fork 到你个人的 GitHub 账号。
2. **创建分支**: 从 `main` 分支切出一个功能分支。
   - `feat/feature-name`
   - `fix/bug-description`
   - `docs/doc-update`
3. **本地开发**: 在本地进行代码编写和测试。
4. **提交代码**: 遵循 [Commit 规范](#commit-规范)。
5. **发起 PR**: 提交 Pull Request 到主仓库的 `main` 分支。
6. **代码审查**: 等待维护者审查代码并根据反馈进行调整。

## 代码规范

### 前端 (React/Next.js)
- 使用 TypeScript 编写所有逻辑，避免使用 `any`。
- 遵循 React 19 的最新实践（如 Server Components）。
- 使用 TailwindCSS 进行样式开发，遵循一致的 UI 风格。

### 后端 (Go/Python)
- Go: 遵循 [Effective Go](https://golang.org/doc/effective_go.html) 规范。
- Python: 遵循 PEP 8 规范。
- 确保所有 API 都有明确的错误处理。

## Commit 规范

我们采用 [Conventional Commits](https://www.conventionalcommits.org/) 规范。提交消息格式如下：

`<type>(<scope>): <description>`

常见的 `type`：
- `feat`: 新功能
- `fix`: 修复 Bug
- `docs`: 文档变更
- `style`: 代码格式调整（不影响逻辑）
- `refactor`: 重构代码
- `test`: 增加或修改测试
- `chore`: 构建过程或辅助工具的变动

示例：`feat(auth): 增加 GitHub OAuth 登录支持`

## PR 流程与要求

- **关联 Issue**: 每个 PR 建议关联一个已有的 Issue。
- **原子性**: 一个 PR 尽量只做一件事。
- **测试**: 确保新代码通过了相关测试，并且没有引入新的 Regression。
- **文档**: 如果涉及功能变更，请同步更新 `docs/` 下的相关文档。
- **审查**: 每个 PR 需要至少一名维护者的批准 (Approve) 才能合并。

## 代码审查要求

维护者在审查代码时会关注：
- 代码逻辑是否正确且高效。
- 是否符合项目的技术栈规范。
- 是否有潜在的安全风险（如敏感信息泄漏）。
- 变量命名、注释是否清晰易懂。
