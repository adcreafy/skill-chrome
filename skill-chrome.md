# Skill Installer - Chrome Extension 产品方案

## 1. 产品目的

### 1.1 解决的问题

当前 AI Agent 生态中，Skill（规则/能力包）的安装方式依赖命令行操作（如 `npx skills install xxxxx`），存在以下痛点：

- 用户需要打开终端、手动执行命令，操作门槛高
- 不同 Agent 应用（Cursor、Claude Code、Windsurf 等）的 Skill 目录和格式各不相同，用户需要自己了解差异
- Skill 发现和安装是割裂的体验：在网页上浏览 Skill 介绍，却要切到终端安装

### 1.2 产品定位

开发一个 Chrome 插件，用户在浏览 Skill 介绍页面时，可以通过 SidePanel 一键将 Skill 安装到本地已有的 Agent 应用中，无需安装任何额外软件。

### 1.3 目标用户

- 使用 AI Agent 工具（Cursor、Claude Code、Windsurf、Continue 等）的开发者
- 使用基于主流 Agent 框架开发的自定义 Agent 的用户

---

## 2. 技术可行性分析

### 2.1 核心技术约束

| 约束 | 说明 |
|------|------|
| 浏览器沙箱 | Chrome 插件无法执行本地系统命令（如 npx） |
| 无 Node.js 运行时 | 浏览器环境中无法运行 Node.js |
| 文件系统受限 | 插件无法主动扫描用户电脑文件系统 |

### 2.2 解决方案

| 能力需求 | 技术方案 |
|---------|---------|
| 执行安装命令、获取 Skill 产物 | 服务端代为执行，返回文件内容 |
| 写入用户本地文件系统 | File System Access API（零安装） |
| 识别用户已有 Agent 应用 | 已知应用列表 + 用户自选 |
| 确定 Skills 安装目录 | 主流应用预设路径 + 自定义应用手动选择 |

### 2.3 关键 API

- **File System Access API**：`showDirectoryPicker()` 允许用户授权目录访问权限，插件获得读写能力
- **IndexedDB**：持久化存储 Directory Handle，用户不需要每次重新选择
- **Chrome SidePanel API**：Manifest V3 支持的侧边栏，承载主要交互界面

### 2.4 兼容性

- File System Access API：Chrome 86+（覆盖绝大多数用户）
- SidePanel API：Chrome 114+
- 最低要求：Chrome 114

---

## 3. 技术架构

### 3.1 整体架构

```
┌─────────────────────────────────────────────────────────────────┐
│  用户浏览器                                                      │
│                                                                   │
│  ┌──────────────┐     ┌──────────────────────────────────────┐  │
│  │ Content Script│     │ SidePanel (Plasmo + React)           │  │
│  │              │────▶│                                      │  │
│  │ 检测 Skill   │     │ - Agent 管理                          │  │
│  │ 页面/命令    │     │ - Skill 预览                          │  │
│  └──────────────┘     │ - 安装操作                            │  │
│                        └──────────┬───────────────────────────┘  │
│                                   │                               │
│                                   │ fetch                         │
└───────────────────────────────────┼───────────────────────────────┘
                                    │
                       ┌────────────┴────────────┐
                       │                         │
                       ▼                         ▼
┌──────────────────────────────┐  ┌──────────────────────────────┐
│  服务端 (NestJS)              │  │  GitHub API (直连备选)       │
│                               │  │                              │
│  ┌─────────────────────────┐ │  │  GET /repos/{owner}/{repo}/  │
│  │ Command Parser          │ │  │      contents/{path}         │
│  │ (解析 npx skills add)   │ │  │                              │
│  └───────────┬─────────────┘ │  └──────────────────────────────┘
│              ↓                │
│  ┌─────────────────────────┐ │
│  │ GitHub Resolver         │ │
│  │ (调用 GitHub API 下载)  │ │
│  └───────────┬─────────────┘ │
│              ↓                │
│  ┌─────────────────────────┐ │
│  │ Format Adapter          │ │
│  │ (仅 Cursor 需要转 .mdc)│ │
│  └─────────────────────────┘ │
└──────────────────────────────┘

写入本地:
  Chrome 插件通过 File System Access API 直接写入
  ~/.claude/skills/xxx/SKILL.md (Claude Code)
  ~/.cursor/rules/xxx.mdc (Cursor)
  等...
```

### 3.2 技术栈

| 模块 | 技术选型 | 说明 |
|------|---------|------|
| Chrome 插件 | **Plasmo** | 支持 React、TypeScript，内置 SidePanel 支持 |
| 服务端 | **NestJS** | 处理命令解析、GitHub API 调用、格式转换、缓存 |
| 插件 UI | React + TailwindCSS | SidePanel 内的交互界面 |
| 持久化存储 | IndexedDB (via idb) | 存储 Directory Handle 和用户配置 |
| GitHub 访问 | Octokit / GitHub REST API | 下载 Skill 文件，无需 Docker |

### 3.3 Chrome 插件结构 (Plasmo)

```
skill-installer/
├── src/
│   ├── sidepanel.tsx            # SidePanel 入口
│   ├── background.ts            # Service Worker
│   ├── contents/
│   │   └── skill-detector.ts    # Content Script: 检测 Skill 页面
│   ├── components/
│   │   ├── AgentSetup.tsx       # Agent 配置引导
│   │   ├── AgentList.tsx        # 已配置 Agent 列表
│   │   ├── SkillPreview.tsx     # Skill 预览卡片
│   │   └── InstallProgress.tsx  # 安装进度
│   ├── services/
│   │   ├── file-system.ts       # File System Access API 封装
│   │   ├── agent-config.ts      # Agent 配置管理
│   │   └── api-client.ts        # 服务端 API 调用
│   └── config/
│       └── known-agents.ts      # 已知 Agent 应用注册表
├── package.json
└── plasmo.config.ts
```

### 3.4 服务端结构 (NestJS)

```
server/
├── src/
│   ├── skill/
│   │   ├── skill.controller.ts       # API 入口
│   │   ├── skill.service.ts          # 核心业务逻辑
│   │   ├── resolvers/                # Skill 来源解析
│   │   │   ├── npm-resolver.ts       # 从 npm registry 获取
│   │   │   └── github-resolver.ts    # 从 GitHub 获取
│   │   └── adapters/                 # 格式适配器
│   │       ├── cursor-adapter.ts     # 转换为 .mdc 格式
│   │       ├── claude-adapter.ts     # 转换为 Claude Code 格式
│   │       └── markdown-adapter.ts   # 通用 Markdown 格式
│   └── app.module.ts
└── Dockerfile
```

---

## 4. 已知 Agent 应用注册表

### 4.1 主流 Agent 应用

| 应用 | 检测路径 | Skills 目录 | 文件格式 | 说明 |
|------|---------|------------|---------|------|
| Cursor | `~/.cursor/` | `~/.cursor/rules/` | `.mdc` | 全局 Rules |
| Claude Code | `~/.claude/` | `~/.claude/commands/` | `.md` | 自定义命令 |
| Windsurf | `~/.codeium/windsurf/` | `~/.codeium/windsurf/memories/` | `.md` | Memories |
| Continue | `~/.continue/` | `~/.continue/rules/` | `.yaml` / `.md` | Rules |
| Gemini CLI | `~/.gemini/` | `~/.gemini/` | `.md` | 配置文件 |
| Aider | `~/.aider/` | `~/.aider/` | `.md` | 约定文件 |

### 4.2 自定义 Agent 配置

对于非主流 Agent，用户需要提供：
- 应用名称
- 基于哪个主流框架（决定文件格式）
- Skills 目录路径（手动选择文件夹）

---

## 5. UX 交互设计

### 5.1 交互总览

插件所有交互均在 **Chrome SidePanel（侧边栏）** 中完成，共包含以下几个页面/状态：

```
首次使用引导 → 主页（Skill 安装） → 设置页（Agent 管理）
```

---

### 5.2 首次使用引导流程

#### 页面：Welcome

```
┌─ SidePanel ──────────────────────────┐
│                                       │
│  🛠️ Skill Installer                  │
│                                       │
│  一键安装 AI Skill 到你的 Agent 应用  │
│                                       │
│  ─────────────────────────────────    │
│                                       │
│  Step 1/2: 授权文件访问               │
│                                       │
│  为了将 Skill 文件写入你的电脑，      │
│  需要你选择用户主目录                  │
│                                       │
│  ┌─────────────────────────────────┐  │
│  │ 📂 选择主目录 (如 /Users/xxx)  │  │
│  └─────────────────────────────────┘  │
│                                       │
│  ℹ️ 仅用于检测已安装的 Agent 应用     │
│  和写入 Skill 文件，不会读取其他数据  │
│                                       │
└───────────────────────────────────────┘
```

**交互说明：**
- 用户点击"选择主目录"→ 弹出系统文件夹选择器
- 用户选择 Home 目录（如 `/Users/username`）
- 选择成功后自动进入 Step 2

#### 页面：Agent 选择

```
┌─ SidePanel ──────────────────────────┐
│                                       │
│  Step 2/2: 选择你的 Agent 应用        │
│                                       │
│  已检测到：                           │
│  ┌─────────────────────────────────┐  │
│  │ ✅ Cursor                       │  │
│  │    ~/.cursor/rules/             │  │
│  └─────────────────────────────────┘  │
│  ┌─────────────────────────────────┐  │
│  │ ✅ Claude Code                  │  │
│  │    ~/.claude/commands/          │  │
│  └─────────────────────────────────┘  │
│                                       │
│  未检测到：                           │
│  ┌─────────────────────────────────┐  │
│  │ ○ Windsurf                      │  │
│  │ ○ Continue                      │  │
│  │ ○ Gemini CLI                    │  │
│  └─────────────────────────────────┘  │
│                                       │
│  ┌─────────────────────────────────┐  │
│  │ ＋ 添加自定义 Agent             │  │
│  └─────────────────────────────────┘  │
│                                       │
│  [完成设置]                           │
│                                       │
└───────────────────────────────────────┘
```

**交互说明：**
- 插件自动用已授权的 Home 目录逐个检查已知 Agent 的检测路径
- 检测到的应用自动勾选并显示 skills 路径
- 未检测到的应用灰显，用户可手动勾选（会弹出文件夹选择器确认路径）
- 点击"添加自定义 Agent"进入自定义配置

#### 弹层：添加自定义 Agent

```
┌─ SidePanel ──────────────────────────┐
│                                       │
│  ← 返回                              │
│                                       │
│  添加自定义 Agent                     │
│                                       │
│  应用名称                             │
│  ┌─────────────────────────────────┐  │
│  │ My Custom Agent                 │  │
│  └─────────────────────────────────┘  │
│                                       │
│  基于框架（决定 Skill 文件格式）      │
│  ┌─────────────────────────────────┐  │
│  │ ▼ Cursor (.mdc)                │  │
│  └─────────────────────────────────┘  │
│  可选项:                              │
│  • Cursor (.mdc)                      │
│  • Claude Code (.md)                  │
│  • 通用 Markdown (.md)               │
│  • YAML (.yaml)                       │
│                                       │
│  Skills 安装目录                      │
│  ┌─────────────────────────────────┐  │
│  │ 📂 选择文件夹                   │  │
│  └─────────────────────────────────┘  │
│                                       │
│  [确认添加]                           │
│                                       │
└───────────────────────────────────────┘
```

---

### 5.3 主页：Skill 安装

#### 状态 A：当前页面检测到 Skill

当用户访问的网页包含 Skill 信息时，Content Script 提取数据并展示在 SidePanel：

```
┌─ SidePanel ──────────────────────────┐
│                                       │
│  Skill Installer        ⚙️            │
│                                       │
│  ─────────────────────────────────    │
│  检测到 Skill:                        │
│                                       │
│  ┌─────────────────────────────────┐  │
│  │ 📦 awesome-code-review          │  │
│  │                                 │  │
│  │ 代码审查助手，帮助发现常见问题  │  │
│  │ 和优化建议                      │  │
│  │                                 │  │
│  │ 来源: npx skills install        │  │
│  │       awesome-code-review       │  │
│  │                                 │  │
│  │ 包含文件:                       │  │
│  │ • code-review.mdc (12KB)        │  │
│  │ • lint-rules.yaml (3KB)         │  │
│  └─────────────────────────────────┘  │
│                                       │
│  安装到:                              │
│  ┌─────────────────────────────────┐  │
│  │ ☑ Cursor      ~/.cursor/rules/ │  │
│  │ ☑ Claude Code ~/.claude/cmds/  │  │
│  │ ☐ My Agent    ~/my-agent/      │  │
│  └─────────────────────────────────┘  │
│                                       │
│  ┌─────────────────────────────────┐  │
│  │        🚀 安装 Skill            │  │
│  └─────────────────────────────────┘  │
│                                       │
└───────────────────────────────────────┘
```

**交互说明：**
- Content Script 检测页面中的安装命令（如 `npx skills install xxx`）
- SidePanel 展示 Skill 的名称、描述、包含的文件
- 用户勾选目标 Agent 应用
- 点击"安装 Skill"触发安装流程

#### 状态 B：安装进行中

```
┌─ SidePanel ──────────────────────────┐
│                                       │
│  Skill Installer        ⚙️            │
│                                       │
│  ─────────────────────────────────    │
│                                       │
│  正在安装 awesome-code-review...      │
│                                       │
│  ┌─────────────────────────────────┐  │
│  │ ✅ 解析 Skill 信息              │  │
│  │ ✅ 获取 Skill 文件              │  │
│  │ ⏳ 写入 Cursor...               │  │
│  │ ○  写入 Claude Code...          │  │
│  └─────────────────────────────────┘  │
│                                       │
│  ━━━━━━━━━━━━━━░░░░░░ 60%            │
│                                       │
└───────────────────────────────────────┘
```

#### 状态 C：安装完成

```
┌─ SidePanel ──────────────────────────┐
│                                       │
│  Skill Installer        ⚙️            │
│                                       │
│  ─────────────────────────────────    │
│                                       │
│  ✅ 安装成功！                        │
│                                       │
│  awesome-code-review 已安装到:        │
│                                       │
│  • Cursor → ~/.cursor/rules/          │
│    code-review.mdc                    │
│  • Claude Code → ~/.claude/commands/  │
│    code-review.md                     │
│                                       │
│  ┌─────────────────────────────────┐  │
│  │         完成                     │  │
│  └─────────────────────────────────┘  │
│                                       │
└───────────────────────────────────────┘
```

#### 状态 D：当前页面未检测到 Skill

```
┌─ SidePanel ──────────────────────────┐
│                                       │
│  Skill Installer        ⚙️            │
│                                       │
│  ─────────────────────────────────    │
│                                       │
│  当前页面未检测到 Skill               │
│                                       │
│  请访问 Skill 介绍页面，插件会自动   │
│  识别可安装的 Skill。                 │
│                                       │
│  ─────────────────────────────────    │
│                                       │
│  最近安装:                            │
│  • awesome-code-review (6/1)         │
│  • test-helper (5/28)                │
│                                       │
│  ─────────────────────────────────    │
│                                       │
│  支持的 Skill 来源:                   │
│  • skills.dev                         │
│  • npmjs.com (skills-* 包)           │
│  • GitHub (含 SKILL.md 的仓库)       │
│                                       │
└───────────────────────────────────────┘
```

---

### 5.4 设置页

```
┌─ SidePanel ──────────────────────────┐
│                                       │
│  ← 返回           设置               │
│                                       │
│  ─────────────────────────────────    │
│  Agent 应用管理                       │
│  ─────────────────────────────────    │
│                                       │
│  Cursor ✅                            │
│  ~/.cursor/rules/         [修改] [删除]│
│                                       │
│  Claude Code ✅                       │
│  ~/.claude/commands/      [修改] [删除]│
│                                       │
│  My Agent                             │
│  ~/my-agent/skills/       [修改] [删除]│
│                                       │
│  [＋ 添加 Agent]                      │
│                                       │
│  ─────────────────────────────────    │
│  文件访问                             │
│  ─────────────────────────────────    │
│                                       │
│  主目录: /Users/username              │
│  [重新选择]                           │
│                                       │
│  ─────────────────────────────────    │
│  关于                                 │
│  ─────────────────────────────────    │
│  版本: 1.0.0                          │
│  已安装 Skill 数量: 12                │
│                                       │
└───────────────────────────────────────┘
```

---

## 6. 完整链路（End-to-End Flow）

### 6.1 首次使用链路

```
用户安装 Chrome 插件
    ↓
首次打开 SidePanel
    ↓
引导: 选择 Home 目录 (File System Access API)
    ↓
自动检测已知 Agent 目录是否存在
    ↓
展示检测结果，用户确认/补充
    ↓
配置保存到 IndexedDB (包含 DirectoryHandle)
    ↓
进入主页，等待用户访问 Skill 页面
```

### 6.2 安装 Skill 链路

```
1. 用户访问 Skill 介绍页面 (如 skills.dev/awesome-review)
       ↓
2. Content Script 检测页面，提取:
   - Skill 名称
   - 安装命令 (如 npx skills install awesome-review)
   - 描述信息
   - 来源类型 (npm / github / direct)
       ↓
3. 通过 chrome.runtime.sendMessage 发送到 Background
       ↓
4. Background 转发到 SidePanel，SidePanel 展示 Skill 卡片
       ↓
5. 用户勾选目标 Agent，点击"安装"
       ↓
6. SidePanel 调用服务端 API:
   POST /api/skills/resolve
   Body: {
     command: "npx skills install awesome-review",
     targets: ["cursor", "claude-code"]
   }
       ↓
7. 服务端处理:
   a. 解析命令，确定 Skill 来源
   b. 在隔离环境中执行安装命令 / 下载 npm 包 / 克隆仓库
   c. 获取 Skill 源文件内容
   d. 根据目标 Agent 格式进行转换:
      - cursor → 生成 .mdc 文件 (含 frontmatter)
      - claude-code → 生成 .md 文件
   e. 返回文件列表:
      {
        files: [
          { agent: "cursor", path: "awesome-review.mdc", content: "..." },
          { agent: "claude-code", path: "awesome-review.md", content: "..." }
        ]
      }
       ↓
8. SidePanel 接收文件，通过 File System Access API 写入:
   - 从 IndexedDB 取出对应 Agent 的 DirectoryHandle
   - 创建/覆写文件
       ↓
9. 写入成功，展示安装结果
```

### 6.3 权限刷新链路

浏览器重启后，DirectoryHandle 仍存在于 IndexedDB，但需要重新请求写入权限：

```
用户打开 SidePanel
    ↓
从 IndexedDB 读取已保存的 DirectoryHandle
    ↓
调用 handle.requestPermission({ mode: 'readwrite' })
    ↓
浏览器弹出权限确认 → 用户点允许
    ↓
恢复正常使用（无需重新选择目录）
```

---

## 7. 服务端 API 设计 (NestJS)

### 7.1 核心接口

#### POST /api/skills/resolve

解析并获取 Skill 文件内容。由于 Skill 本质是 GitHub 仓库中的静态文件，服务端只需通过 GitHub API 下载并做格式转换。

**Request:**
```json
{
  "source": {
    "type": "github",
    "owner": "vercel-labs",
    "repo": "agent-skills",
    "skill": "awesome-review"
  },
  "targets": [
    { "agent": "cursor", "format": "mdc" },
    { "agent": "claude-code", "format": "skill-md" }
  ]
}
```

**Response:**
```json
{
  "skill": {
    "name": "awesome-review",
    "description": "代码审查助手",
    "repoUrl": "https://github.com/vercel-labs/agent-skills",
    "version": "latest"
  },
  "files": [
    {
      "agent": "cursor",
      "relativePath": "awesome-review.mdc",
      "content": "---\ndescription: 代码审查助手\nglobs: []\nalwaysApply: true\n---\n# Code Review...",
      "size": 12480
    },
    {
      "agent": "claude-code",
      "relativePath": "awesome-review/SKILL.md",
      "content": "---\nname: awesome-review\ndescription: 代码审查助手\n---\n# Code Review...",
      "size": 10240
    },
    {
      "agent": "claude-code",
      "relativePath": "awesome-review/references/checklist.md",
      "content": "# Review Checklist...",
      "size": 3200
    }
  ]
}
```

#### GET /api/skills/preview

预览 Skill 信息（通过 GitHub API 获取 SKILL.md 的 frontmatter）。

**Request:**
```
GET /api/skills/preview?owner=vercel-labs&repo=agent-skills&skill=awesome-review
```

**Response:**
```json
{
  "name": "awesome-review",
  "description": "代码审查助手，帮助发现常见问题和优化建议",
  "repoUrl": "https://github.com/vercel-labs/agent-skills",
  "files": [
    { "path": "SKILL.md", "size": 8200 },
    { "path": "references/checklist.md", "size": 3200 }
  ],
  "totalSize": 11400,
  "supportedAgents": ["cursor", "claude-code", "codex", "windsurf", "continue"]
}
```

### 7.2 服务端模块划分

```
SkillModule
├── SkillController            # 处理 HTTP 请求
├── SkillService               # 协调解析和转换
├── Resolvers/
│   ├── GithubResolver         # 通过 GitHub API 下载 skill 文件（核心）
│   ├── CommandParser          # 解析 npx skills add 命令，提取 owner/repo/skill
│   └── RegistryResolver       # 从 skills.sh API 查询 skill 元数据
└── Adapters/
    ├── CursorAdapter          # SKILL.md → .mdc 格式转换
    ├── PassthroughAdapter     # Claude Code / Codex 等直接透传
    └── CopilotAdapter         # 合并到 copilot-instructions.md（可选）
```

### 7.3 安全措施

| 风险 | 对策 |
|------|------|
| GitHub API 限流 | 服务端使用 GitHub App Token（5000次/小时）+ 响应缓存 |
| 大文件攻击 | 限制单个 Skill 最大体积 (如 1MB) |
| API 滥用 | Rate limiting，可选用户认证 |
| 恶意仓库内容 | 内容安全扫描（可选）、文件类型白名单 |
| 私有仓库 | 用户可在插件配置 GitHub Personal Token |

---

## 8. Skill 检测机制（深度分析）

### 8.1 Skills 生态现状

经过调研，当前 Agent Skills 生态的核心事实：

| 事实 | 说明 |
|------|------|
| 统一标准 | 所有主流 Agent 都采用 `SKILL.md` 作为通用 skill 格式 |
| 来源 | Skill 的源头是 **GitHub 仓库**（不是 npm 包） |
| 安装 CLI | `npx skills add owner/repo`（由 Vercel Labs 维护） |
| 本质 | Skill 是**纯静态文件**，不含可执行逻辑 |
| 注册中心 | skills.sh（官方）、smithery.ai、agentskill.sh 等 |

**Skill 文件结构：**
```
skill-name/
├── SKILL.md              # 必须：YAML frontmatter + Markdown 指令
├── scripts/              # 可选：辅助脚本
├── references/           # 可选：参考文档
└── assets/               # 可选：静态资源
```

**SKILL.md 格式：**
```markdown
---
name: awesome-review
description: 代码审查助手，帮助发现常见问题和优化建议
---

# Code Review Skill

具体指令内容...
```

### 8.2 支持的 Skill 来源页面

| 来源平台 | URL 模式 | 页面特征 | 优先级 |
|---------|----------|---------|--------|
| **skills.sh** | `skills.sh/{owner}/{repo}/{skill}` | 官方注册中心，结构化数据最完整 | 最高 |
| **Smithery** | `smithery.ai/skills/{namespace}/{skill}` | 有安装命令和描述 | 高 |
| **GitHub** | `github.com/{owner}/{repo}` 且包含 `SKILL.md` | 仓库文件列表可见 SKILL.md | 高 |
| **agentskill.sh** | `agentskill.sh/{slug}` | 另一个注册中心 | 中 |
| **任意网页** | 页面中包含 `npx skills add` 命令 | 博客、文档等介绍页 | 中 |

### 8.3 Content Script 检测策略

#### 策略 1：URL 匹配 + 结构化提取（注册中心页面）

针对已知注册中心，使用 URL pattern 精确匹配后，直接从页面 DOM 提取结构化数据：

```typescript
const REGISTRY_DETECTORS = [
  {
    name: 'skills.sh',
    urlPattern: /^https:\/\/(www\.)?skills\.sh\//,
    extract: (doc: Document) => {
      // skills.sh 页面有固定结构，提取 skill 名称、描述、安装命令
      return {
        name: extractFromSelector(doc, '.skill-name'),
        command: extractFromSelector(doc, '.install-command code'),
        description: extractFromSelector(doc, '.skill-description'),
        repoUrl: extractFromSelector(doc, 'a[href*="github.com"]'),
      }
    }
  },
  {
    name: 'smithery',
    urlPattern: /^https:\/\/smithery\.ai\/skills\//,
    extract: (doc: Document) => {
      // Smithery 页面结构
      return {
        name: extractFromUrl(location.pathname), // /skills/owner/skill-name
        command: extractFromSelector(doc, 'code[data-install]'),
        description: extractFromSelector(doc, '.skill-description'),
      }
    }
  },
  {
    name: 'github',
    urlPattern: /^https:\/\/github\.com\/[^/]+\/[^/]+/,
    extract: (doc: Document) => {
      // 检查文件列表中是否存在 SKILL.md 或 skills/ 目录
      const hasSkillMd = doc.querySelector('a[title="SKILL.md"]')
      const hasSkillsDir = doc.querySelector('a[title="skills"]')
      if (!hasSkillMd && !hasSkillsDir) return null
      
      const [, owner, repo] = location.pathname.match(/\/([^/]+)\/([^/]+)/) || []
      return {
        name: repo,
        command: `npx skills add ${owner}/${repo}`,
        repoUrl: `https://github.com/${owner}/${repo}`,
      }
    }
  }
]
```

#### 策略 2：通用命令检测（任意页面）

对不在已知注册中心的页面，扫描页面中的代码块，提取安装命令：

```typescript
const COMMAND_PATTERNS = [
  // npx skills add owner/repo
  /npx\s+skills\s+add\s+([\w\-./]+(?:@[\w\-]+)?)/,
  // npx skills add URL --skill name
  /npx\s+skills\s+add\s+(https?:\/\/[^\s]+)(?:\s+--skill\s+([\w\-]+))?/,
  // smithery skill add namespace/name
  /smithery\s+skill\s+add\s+([\w\-]+\/[\w\-]+)/,
  // ags install slug
  /ags\s+install\s+([@\w\-/]+)/,
]

function detectCommandsInPage(doc: Document): DetectedSkill[] {
  const codeElements = doc.querySelectorAll('code, pre, .highlight')
  const results: DetectedSkill[] = []
  
  for (const el of codeElements) {
    const text = el.textContent || ''
    for (const pattern of COMMAND_PATTERNS) {
      const match = text.match(pattern)
      if (match) {
        results.push(parseMatchToSkill(match, pattern))
      }
    }
  }
  return results
}
```

#### 策略 3：页面 Meta 标签（推荐 Skill 作者添加）

鼓励 Skill 介绍页面添加标准化 meta 标签，便于一键识别：

```html
<meta name="agent-skill" content="owner/repo@skill-name" />
<meta name="agent-skill-description" content="Skill 描述" />
<meta name="agent-skill-install" content="npx skills add owner/repo --skill xxx" />
```

### 8.4 检测流程

```
页面加载完成
    ↓
1. URL 匹配检查（是否为已知注册中心？）
    ├─ 是 → 使用对应 Registry Detector 提取结构化数据
    └─ 否 ↓
2. 通用命令扫描（页面中是否包含安装命令？）
    ├─ 是 → 解析命令，提取 skill 信息
    └─ 否 ↓
3. Meta 标签检查
    ├─ 是 → 直接读取 meta 内容
    └─ 否 → 当前页面无 Skill，SidePanel 显示空状态
    ↓
检测成功 → 通过 chrome.runtime.sendMessage 发送到 Background
    ↓
Background 转发到 SidePanel 展示
```

### 8.5 提取的信息结构

```typescript
interface DetectedSkill {
  name: string                 // skill 名称
  source: SkillSource          // 来源类型
  repoUrl: string              // GitHub 仓库 URL（核心字段）
  skillPath?: string           // 仓库内的 skill 路径（如 skills/awesome-review/）
  command: string              // 原始安装命令
  description?: string         // 描述信息
  pageUrl: string              // 检测到 skill 的页面 URL
  detectedAt: number           // 检测时间戳
  confidence: 'high' | 'medium' | 'low'  // 检测置信度
}

type SkillSource = 
  | { type: 'registry'; registry: 'skills.sh' | 'smithery' | 'agentskill.sh' }
  | { type: 'github'; owner: string; repo: string }
  | { type: 'command'; raw: string }
```

### 8.6 检测结果置信度

| 置信度 | 条件 | 用户体验 |
|--------|------|---------|
| high | 从已知注册中心结构化提取 | 直接展示安装按钮 |
| medium | 从 GitHub 页面检测到 SKILL.md | 展示安装按钮 + 提示"检测到可能的 Skill" |
| low | 从任意页面代码块中提取命令 | 展示确认对话框"是否安装此 Skill？" |

---

## 9. 数据持久化 (IndexedDB)

### 9.1 存储结构

```typescript
// agents 表: 用户配置的 Agent 列表
interface AgentConfig {
  id: string
  name: string
  format: 'mdc' | 'md' | 'yaml' | 'json'
  builtin: boolean
  directoryHandle: FileSystemDirectoryHandle
  detected: boolean
  createdAt: number
}

// installations 表: 安装历史
interface InstallRecord {
  id: string
  skillName: string
  agentId: string
  filename: string
  installedAt: number
  sourceUrl: string
}

// settings 表: 全局配置
interface Settings {
  homeDirectoryHandle: FileSystemDirectoryHandle
  setupCompleted: boolean
  lastPermissionCheck: number
}
```

---

## 10. 实现计划

### Phase 1: MVP（核心流程）

- Chrome 插件 SidePanel 基础框架 (Plasmo)
- 首次使用引导（Home 目录授权 + Agent 检测）
- Content Script 检测 `npx skills add` 命令和 skills.sh 页面
- 服务端 GitHub Resolver（通过 GitHub API 下载 skill 文件）
- CursorAdapter（SKILL.md → .mdc 格式转换）
- File System Access API 写入本地 Agent 目录
- 支持 Claude Code + Cursor 两个 Agent

### Phase 2: 扩展来源与格式

- 增加 Smithery、agentskill.sh 页面检测
- 增加 GitHub 仓库自动检测（页面含 SKILL.md）
- 支持自定义 Agent 配置
- 安装历史记录
- 支持含支撑文件的复杂 Skill（scripts/、references/ 目录）

### Phase 3: 去服务端化 + 增值

- 插件直连 GitHub API（无需经服务端，适合简单场景）
- 服务端加缓存层 + 统计分析
- 热门 Skill 推荐
- Skill 搜索功能集成到 SidePanel
- Meta 标签标准推广（鼓励 Skill 作者添加）

---

## 11. 服务端下载可行性深度分析

### 11.1 核心结论

**服务端方案完全可行，且比预想的更简单。**

原因：`npx skills add` 的底层逻辑是从 GitHub 仓库下载静态文件，不涉及任何本地运行时逻辑。服务端完全可以跳过 npx，直接用 GitHub API 获取文件。

### 11.2 服务端实际需要做什么

```
收到请求: { command: "npx skills add vercel-labs/agent-skills --skill awesome-review" }
    ↓
解析命令 → 提取: owner=vercel-labs, repo=agent-skills, skill=awesome-review
    ↓
调用 GitHub API:
  GET https://api.github.com/repos/vercel-labs/agent-skills/contents/skills/awesome-review/
    ↓
递归下载该目录下所有文件 (SKILL.md + 可选支撑文件)
    ↓
返回文件列表给 Chrome 插件
```

**完全不需要执行 npx，不需要 Docker 容器，不需要 Node.js 运行时。**

### 11.3 服务端为什么不需要执行 npx？

| `npx skills add` 做的事情 | 服务端能否替代 | 方式 |
|---------------------------|-------------|------|
| 解析 owner/repo | ✅ | 正则解析命令字符串 |
| 从 GitHub 下载仓库文件 | ✅ | GitHub REST API / git clone |
| 在仓库中查找 SKILL.md | ✅ | 按约定路径查找 |
| 写入本地 agent 目录 | ❌ 不需要 | 这一步由 Chrome 插件完成 |
| 创建 symlink | ❌ 不需要 | 插件直接 copy 文件 |

### 11.4 各 Agent 的 Skill 安装目录和格式

经过调研，所有主流 Agent 都支持 SKILL.md 格式（因为 `npx skills` CLI 自动处理了映射）：

| Agent | 全局 Skills 目录 | 项目级目录 | 是否需要格式转换 |
|-------|-----------------|-----------|----------------|
| **Claude Code** | `~/.claude/skills/<name>/SKILL.md` | `.claude/skills/<name>/SKILL.md` | ❌ 原样放入 |
| **Cursor** | `~/.cursor/rules/<name>.mdc` | `.cursor/rules/<name>.mdc` | ⚠️ 需要转换为 .mdc |
| **Codex** | `~/.codex/skills/<name>/SKILL.md` | `.codex/skills/<name>/SKILL.md` | ❌ 原样放入 |
| **Windsurf** | `~/.codeium/windsurf/skills/<name>/SKILL.md` | 待确认 | ❌ 原样放入 |
| **Continue** | `~/.continue/skills/<name>/SKILL.md` | 待确认 | ❌ 原样放入 |
| **Cline** | `~/.cline/skills/<name>/SKILL.md` | 待确认 | ❌ 原样放入 |
| **GitHub Copilot** | `.github/copilot-instructions.md` | 项目级别 | ⚠️ 需要合并到单文件 |

### 11.5 唯一需要格式转换的情况：Cursor

Cursor 使用 `.mdc` 格式，与 SKILL.md 的 frontmatter 字段不同：

**SKILL.md 格式（源）：**
```markdown
---
name: awesome-review
description: 代码审查助手
---

# Code Review Instructions...
```

**Cursor .mdc 格式（目标）：**
```markdown
---
description: "代码审查助手"
globs: []
alwaysApply: true
---

# Code Review Instructions...
```

**转换逻辑（服务端实现）：**
```typescript
function convertToMdc(skillMd: string): string {
  const { frontmatter, body } = parseFrontmatter(skillMd)
  
  const mdcFrontmatter = {
    description: frontmatter.description || frontmatter.name,
    globs: [],
    alwaysApply: true,
  }
  
  return formatWithFrontmatter(mdcFrontmatter, body)
}
```

### 11.6 可能不支持 / 有风险的场景

| 场景 | 风险等级 | 说明 | 应对方案 |
|------|---------|------|---------|
| **私有仓库** | 中 | GitHub API 需要认证才能访问 | 让用户在插件中配置 GitHub Token |
| **大型仓库** | 低 | 下载整个仓库过慢 | 只下载 skill 目录，用 Contents API |
| **含支撑文件的 Skill** | 低 | 除 SKILL.md 外有 scripts/ references/ | 递归下载整个 skill 目录 |
| **GitHub API 限流** | 中 | 未认证 60 次/小时，认证 5000 次/小时 | 服务端使用 GitHub App 认证 |
| **Skill 引用外部依赖** | 极低 | 极少数 skill 可能依赖外部工具 | 暂不处理，记录为已知限制 |

### 11.7 服务端简化方案（推荐）

根据以上分析，服务端不需要 Docker 容器，架构大幅简化：

```
原方案（已废弃）:
  服务端 → Docker 容器 → 执行 npx skills → 捕获产物 → 返回

新方案（推荐）:
  服务端 → GitHub API → 下载 SKILL.md + 支撑文件 → [可选]格式转换 → 返回
```

**服务端核心逻辑（NestJS）：**

```typescript
@Injectable()
export class SkillResolverService {
  
  async resolve(command: string, targets: AgentTarget[]): Promise<ResolvedSkill> {
    // 1. 解析命令，提取 GitHub 仓库信息
    const { owner, repo, skillName } = this.parseCommand(command)
    
    // 2. 通过 GitHub API 获取 skill 文件列表
    const files = await this.githubService.getSkillFiles(owner, repo, skillName)
    
    // 3. 根据目标 Agent 做格式转换（大部分不需要）
    const result: ResolvedFile[] = []
    for (const target of targets) {
      if (target.agent === 'cursor') {
        result.push(this.cursorAdapter.convert(files))
      } else {
        // Claude Code / Codex / 其他 → 直接返回原文件
        result.push({ ...files, agent: target.agent })
      }
    }
    
    return { skill: { name: skillName, owner, repo }, files: result }
  }
}
```

### 11.8 不需要服务端的极简方案（备选）

由于 GitHub API 是公开的，Chrome 插件其实可以**直接调用 GitHub API**，完全跳过服务端：

```
Chrome 插件 → GitHub API → 下载文件 → 格式转换（前端完成）→ 写入本地
```

| 方案 | 优点 | 缺点 |
|------|------|------|
| **经服务端** | 可缓存、可统计、可处理私有仓库、格式转换集中管理 | 多一跳延迟 |
| **插件直连 GitHub** | 零延迟、无服务端成本 | API 限流严格、无法统计、无法处理复杂场景 |

**建议：MVP 阶段直连 GitHub，后续加服务端做缓存和增值功能。**

---

## 12. 已知限制

| 限制 | 影响 | 缓解措施 |
|------|------|---------|
| 浏览器重启后需重新确认权限 | 用户需点一次"允许" | 提示用户，交互友好 |
| File System Access API 需用户手势触发 | 首次必须在 SidePanel 内操作 | 引导流程设计清晰 |
| 无法写入需要 sudo 的目录 | 部分自定义路径不可写 | 前端提示选择有权限的目录 |
| GitHub API 限流（未认证 60 次/小时） | 高频使用受限 | 服务端统一认证，或用户配置 Token |
| 私有仓库无法直接访问 | 部分 Skill 不可用 | 支持用户配置 GitHub Token |
| Cursor 需要格式转换 | 转换可能丢失信息 | 保守转换策略，保留所有内容 |
| 页面检测依赖 DOM 结构 | 注册中心改版会导致检测失效 | 优先使用 meta 标签，DOM 为降级方案 |
