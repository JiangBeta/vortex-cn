# Vortex 中文化（vortex-cn）实现计划

> 方案 2：i18n 抽象层。以「英文原文为 key」集中管理翻译，默认中文、可回退英文。
> 本文件是后续每一步实现与验证的唯一依据，术语表与清单随实现持续更新。

---

## 0. 方案核心设计

### 0.1 i18n 层（新建 `internal/i18n/`）

```go
// internal/i18n/i18n.go
package i18n

import "fmt"

var lang = "zh" // vortex-cn 分支默认中文

func SetLang(l string) { if l == "en" || l == "zh" { lang = l } }
func Lang() string     { return lang }

// T 返回 key 的翻译；未命中时返回 key 本身（即英文原文）。
func T(key string) string {
    if lang == "zh" {
        if v, ok := zh[key]; ok { return v }
    }
    return key
}

// Tf 带参数格式化（key 为英文 format 串，如 "CPU usage is critical: %.2f%%"）。
func Tf(key string, args ...any) string { return fmt.Sprintf(T(key), args...) }
```

```go
// internal/i18n/zh.go —— 唯一术语来源，按类别分组
package i18n

var zh = map[string]string{
    // 模块名
    "Servers": "服务器",
    // ... 见 §2 术语表
}
```

**为什么用英文原文做 key（而非语义 key）**：
- 改造成本最低：`"Dashboard"` → `i18n.T("Dashboard")`，无需为每处新起 key。
- 天然渐进：未翻译的 key 自动显示英文原文，不会出现空白或遗漏，可分批提交。
- 天然可回退：`SetLang("en")` 即回到英文（key 本身）。

**语言来源**：`config.yaml` 新增 `appearance.language`（`"zh"` 默认 / `"en"`），`main.go` 启动时 `i18n.SetLang(config.CurrentConfig.Appearance.Language)`。

### 0.2 改造范式

- 静态文案：`"Dashboard"` → `i18n.T("Dashboard")`
- 带参文案：`fmt.Sprintf("CPU usage: %.2f%%", v)` → `i18n.Tf("CPU usage: %.2f%%", v)`
- 拼接文案：`"Go to " + name` → `i18n.Tf("Go to %s", i18n.T(name))`（内部状态值经 `i18n.T` 后再拼）

---

## 1. 工作清单（译 / 不译）

### 规则速查

| 类别 | 处理 |
|---|---|
| UI 标签/标题/按钮/菜单/placeholder/提示/状态栏 | **译** |
| 本地生成的告警/通知/错误提示（面向用户） | **译** |
| 远程命令字符串（`systemctl restart docker` 等） | **不译** |
| 远程主机返回的输出（`docker ps`/`journalctl`/`cron`/SSH 报错等） | **不译** |
| 内部状态值（`"failed"`/`"running"`/`"active"`，用于 `==` 判断） | **不译**（渲染时经 `i18n.T` 映射成中文） |
| 技术标识符：容器 ID/服务名/路径/yaml·json key/keybind key/theme 名/webhook type | **不译** |
| 专有名词：Docker/systemd/nginx/UFW/Fail2Ban/SSL/SSH/cron/systemctl/journalctl/certbot/PM2 | **不译** |
| 代码注释 | **不译**（本次范围外，避免 diff 噪声） |
| emoji 图标 | **不译** |

### 文件级清单

**新建（3 个）**
- `internal/i18n/i18n.go` — 核心 API（见 §0.1）
- `internal/i18n/zh.go` — 中文翻译表（唯一术语来源）
- `internal/i18n/i18n_test.go` — 单测（`T`/`Tf`/未命中回退/`SetLang`）

**修改：主程序**
- `cmd/vps-manager/main.go`
  - 译：sidebar 模块名（`p.Title()`）、状态栏（`NOT CONNECTED`、`NORMAL`/`ABNORMAL`、`SSH <10ms`、帮助 toast）、palette `Name`/`Description`、logo 下方 `MAIN MENU`/`[ ] to navigate`、空连接提示
  - 不译：`systemctl restart docker/nginx` 命令、keybind key 名、theme 名、`switchTabMsg` 索引

**修改：页面（23 个 `internal/pages/*/*.go`）**
- 译：`Title()` 返回值、`components.Title(...)` 横幅、按钮、placeholder、表单标签、状态提示、本地生成的错误消息
- 不译：远程命令字符串、远程输出、内部状态判断值、`IsInputActive` 逻辑
- 逐页范围见 §4 分步计划

**修改：组件（`internal/components/*.go`）**
- `globe.go`：译 `INFRASTRUCTURE TOPOLOGY`/`SYSTEM`/`RESOURCES`/`NETWORK`/`SERVICES`/`ACTIVITY FEED`/`SERVER TELEMETRY` 等标题
- `palette.go`/`toast.go`/`startup.go`/`card.go`/`sparkline.go`/`progress.go`：仅译用户可见文案；`"error"`/`"success"`/`"Plain"` 等内部类型值不译

**修改：配置（`internal/config/*.go`）**
- `config.go`：新增 `Appearance.Language` 字段 + 默认 `"zh"`；模板注释/模板服务器名可译，yaml key 不译
- `registry.go`：译 `Name`/`Description`/`Options`（显示层）；`ID`、keybind key、`Value` 类型不译

**修改：引擎（`internal/engine/*/*.go`，18 个）**
- **原则上不译**（几乎全是远程命令 + 输出解析）。
- 例外：本地生成、面向用户的告警/错误文案需译，逐点如下：
  - `alerts.go`：`CPU usage is critical/warning`、`RAM/ Disk usage ...`、`Delivered successfully`、`Never fired` 等
  - `uptime.go`：`Alert: Target %s is now %s`、`"up"`/`"down"` 显示值
  - `deploy.go`：`Deployment failed health check`、`Webhook received` 等审计日志文案（审计日志若面向用户则译）
  - `security.go`：`VULNERABLE (Passwords Allowed)` 等报告标签
  - 其余引擎的错误字符串：先保持英文（远程报错为主），在页面层展示时经 `i18n.T` 处理

**不译（保持原样）**
- `internal/agent/payload.go`、`internal/stats/`、`internal/network/`、`internal/docker/`、`internal/services/`（agent 端数据采集，运行在远程主机，JSON 字段与内部值不译）
- `internal/ssh/client.go`：`go build`/`uname`/`powershell` 命令不译；错误消息（`unable to read private key` 等）作为 toast 会显示给用户 → **在页面/主程序展示处翻译，client.go 内部不改**
- `internal/theme/theme.go`：主题名不译
- `go.mod`/`flake.nix`/`.github/`/`README.md`：本次不动

---

## 2. 中英文对应表（核心术语）

> 完整条目实现时逐页补入 `zh.go`；此处为初版，含模块名与高频词。

### 模块名（`Title()` / sidebar）

| 英文原文 | 中文 |
|---|---|
| Servers | 服务器 |
| Mission Control | 任务控制 |
| Processes | 进程 |
| Docker | Docker |
| Services | 服务 |
| Files | 文件 |
| Logs | 日志 |
| Security | 安全 |
| Backups | 备份 |
| SSH | 终端 |
| Settings | 设置 |
| Cron | 定时任务 |
| Certs | 证书 |
| Users & Keys | 用户与密钥 |
| Alerts | 告警 |
| Audit Log | 审计日志 |
| Databases | 数据库 |
| Proxy | 反向代理 |
| Secrets | 密钥 |
| Deploy | 部署 |
| Snapshots | 快照 |
| Uptime Monitor | 可用性监控 |
| Network | 网络 |

### 高频 UI 词

| 英文 | 中文 |
|---|---|
| Connect / Connected | 连接 / 已连接 |
| Disconnect | 断开 |
| Connecting | 连接中 |
| Add / Add New | 添加 / 新增 |
| Save / Saved | 保存 / 已保存 |
| Delete / Remove | 删除 / 移除 |
| Edit | 编辑 |
| Cancel | 取消 |
| Confirm | 确认 |
| Submit | 提交 |
| Start / Stop / Restart | 启动 / 停止 / 重启 |
| Enable / Disable | 启用 / 禁用 |
| Refresh | 刷新 |
| Search | 搜索 |
| Back | 返回 |
| Exit / Quit | 退出 |
| Loading | 加载中 |
| Running | 运行中 |
| Stopped | 已停止 |
| Failed | 失败 |
| Success | 成功 |
| Warning | 警告 |
| Error | 错误 |
| Normal | 正常 |
| Abnormal | 异常 |
| Up / Down | 正常 / 故障（uptime 语境） |
| Not Connected | 未连接 |
| Yes / No | 是 / 否 |
| None | 无 |
| Unknown | 未知 |
| Never | 从不 |
| Delivered successfully | 已送达 |
| Never fired | 从未触发 |
| MAIN MENU | 主菜单 |
| to navigate | 用于导航 |
| Command Palette | 命令面板 |
| Split View | 分屏视图 |

### 短语模板（`Tf` 用，保留 `%s`/`%d`/`%.2f`）

| 英文 format | 中文 |
|---|---|
| Go to %s | 前往%s |
| Switch to %s | 切换到%s |
| Switch to the %s view | 切换到%s视图 |
| Theme: %s | 主题：%s |
| Switch theme to %s | 切换主题为%s |
| CPU usage is critical: %.2f%% | CPU 使用率严重：%.2f%% |
| CPU usage is warning: %.2f%% | CPU 使用率警告：%.2f%% |
| RAM usage is critical/warning: %.2f%% | 内存使用率严重/警告：%.2f%% |
| Disk usage is critical/warning: %.2f%% | 磁盘使用率严重/警告：%.2f%% |
| Alert: Target %s is now %s | 告警：目标 %s 当前为 %s |

> 约定：带 `%%` 的 format 串在 zh.go 里原样保留 `%.2f%%`，`Tf` 用 `fmt.Sprintf` 展开。

---

## 3. 其他规范

### 3.1 宽度对齐（中文每字占 2 列）

1. **所有对齐必须用 lipgloss**（`lipgloss.Width` / `Padding` / `JoinVertical(Left)`），禁止手写 `fmt.Sprintf("%-Ns", ...)` 对齐中文——Go 的 `%-Ns` 按 rune 计宽，中文会错位。
2. 手写 `strings.Repeat(" ", n)` 时，`n` 必须用 `lipgloss.Width(...)` 的**差值**计算，禁止用 `len()`。
3. 已知需改的固定宽度：
   - `main.go` sidebar `Width(26)`、`selectedStyle/normalStyle Width(22)` → 中文模块名需加宽（建议 28/24，实现时按最长中文名实测）。
   - `servers.go:376` `fmt.Sprintf("%-20s %s", sName, status)` → 改 lipgloss。
   - `services.go FormatServices` `%-20s` → 改 lipgloss。
   - `globe.go`/`cron.go`/`apps.go` 中所有 `%-Ns`/手写补空格处。
4. lipgloss 已依赖 `displaywidth`/`uax29`，`lipgloss.Width` 对 CJK 正确，放心使用。

### 3.2 快捷键与 keybind

- `config.Keybinds` 的 **map key 一律不动**（`"Dashboard"`/`"Servers"`/`"Command Palette"`… 是配置标识符）。
- 默认键位（`g`/`f`/`x`/`esc`/`ctrl+p`…）一律不动。
- 只翻译**显示名**：sidebar 标签、帮助 toast、palette 描述里出现的模块名用 `i18n.T`。

### 3.3 内部状态值 vs 显示值（关键）

- 判断逻辑用**原始英文值**，如 `if s.Status == "failed"`、`if t.Type == "discord"`、`if msg.Type == "error"` —— 这些**绝不改**。
- 渲染给用户时才映射：`i18n.T(s.Status)` → 「失败」。
- 反面教材：把 `== "failed"` 改成 `== "失败"` 会导致 systemctl/远程输出永远匹配不上。

### 3.4 提交约定

- 每步一个 commit，commit message 前缀 `i18n:`。
- 禁止同一 commit 混入「翻译」与「宽度修复」之外的逻辑改动。
- 每步先 `go build ./...` 通过再 commit（见 §4 每步验证）。

---

## 4. 分步实现计划

> 每步结束后：`go build ./...` 通过 + 该步验证项通过，才进入下一步。

### Step 0 — i18n 基础设施 + 语言配置

**范围**：新建 `internal/i18n/{i18n.go, zh.go, i18n_test.go}`；`config.go` 加 `Appearance.Language`（默认 `"zh"`）；`main.go` 启动时 `i18n.SetLang(...)`。
**边界**：只搭骨架 + 术语表空壳（先放 §2 模块名 23 条），不动任何页面文案。
**验证**：
- `go build ./...`
- `go test ./internal/i18n/`（`T` 命中/未命中、`Tf`、`SetLang("en")` 回退英文、`SetLang("zh")` 命中中文）

### Step 1 — 主程序外壳（sidebar + 状态栏 + palette）

**范围**：仅 `cmd/vps-manager/main.go`。
**边界**：只改 `View()` 的 sidebar 标签、状态栏、空连接提示、帮助 toast、logo 下方导航文字；`palette` 的 `Name`/`Description`。命令字符串、keybind key、theme 名不动。
**验证**：`go build ./...` + `go vet ./...` + `go run ./cmd/vps-manager` 冒烟（进主界面看 sidebar 中文、未连接状态栏）。

### Step 2 — 全局组件文案

**范围**：`internal/components/*.go`。
**边界**：`globe.go` 各区块标题、`palette.go`/`toast.go`/`startup.go` 可见文案；内部类型值（`"error"`/`"success"`/`"Plain"`）不动。
**验证**：`go build ./...` + `go vet ./...`。

### Step 3 — settings 注册表

**范围**：`internal/config/registry.go`。
**边界**：`Name`/`Description`/`Options` 经 `i18n.T`；`ID`、`Value`、keybind key 不动。
**验证**：`go build ./...` + `go test ./...`（确认 `SaveSettings` 的 `switch s.ID` 未受影响）。

### Step 4 — 页面文案（分 4 批，每批一个 commit）

> 每批译一个页面子集，批次内所有 `Title()` + 横幅 + 按钮 + placeholder + 提示 + 本地错误消息。

- **4a 核心**：`servers`、`dashboard`、`apps`、`docker`、`services`、`files`
- **4b 运维**：`logs`、`security`、`backup`、`cron`、`certs`、`terminal`
- **4c 高级**：`users`、`alerts`、`audit`、`db`、`proxy`、`secrets`
- **4d 剩余**：`deploy`、`snapshots`、`uptime`、`network`、`settings`

**边界**：每批只碰该批页面文件；远程命令/输出/内部状态值不动。
**验证**：每批 `go build ./...` + `go vet ./...`；4a/4b 完成后 `go run` 冒烟进入对应模块看中文。

### Step 5 — 引擎告警文案（例外清单）

**范围**：`internal/engine/{alerts,uptime,deploy,security}.go` 的本地生成告警/审计文案。
**边界**：仅 §1 列出的例外点；远程命令与输出解析不动。
**验证**：`go build ./...` + `go vet ./...`。

### Step 6 — 宽度对齐修复 + 全量回归

**范围**：§3.1 列出的固定宽度/手写对齐点（`main.go` sidebar、`servers.go:376`、`services.go`、`globe.go` 等）。
**边界**：只改对齐，不改文案逻辑。
**验证**：`go build ./...` + `go vet ./...` + `go test ./...` + `golangci-lint run`（nix devShell 内）+ `go run` 全模块冒烟看对齐。

### Step 7 — 英文回退验证（收尾）

**范围**：无代码改动（或仅补 `zh.go` 漏译项）。
**边界**：临时 `i18n.SetLang("en")` 确认全部 UI 回退为英文原文、无空串/无 `key` 泄漏。
**验证**：`go build ./...` + `go test ./...` + 手动 `SetLang("en")` 冒烟后还原为 `"zh"`。

---

## 5. 验收标准

- [ ] 23 个模块标题、sidebar、状态栏、palette 全部中文。
- [ ] 所有页面按钮/placeholder/提示/横幅中文；远程命令与输出保持英文。
- [ ] `SetLang("en")` 可完整回退英文，无 `T()` 未命中导致的 key 泄漏。
- [ ] 中文对齐无错位（sidebar、表格、状态栏）。
- [ ] `go build ./...`、`go vet ./...`、`go test ./...`、`golangci-lint run` 全绿。
- [ ] keybind 键位、配置 yaml key、内部状态判断逻辑零改动。
