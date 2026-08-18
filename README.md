# Vortex-CN

> **本仓库是基于 [berkayyytech/vortex](https://github.com/berkayyytech/vortex) 的中文定制版（MIT License），仅增加中文界面/优化，非官方原版。**

Vortex 是一个功能强大、以键盘为中心的终端用户界面（TUI），用于通过 SSH 管理你的 VPS 服务器集群。它完全使用 Go 和 BubbleTea 框架构建，让你无需复杂 Web 面板，直接在终端中获得快速、视觉丰富、高度可操作的运维环境。

本仓库（Vortex-CN）在保留上游全部功能的基础上，将界面完整中文化，并提供「中文 / 英文」可切换能力。

---

## 中文化说明

- **默认中文**：开箱即中文界面，配置文件 `config.yaml` 中 `appearance.language` 默认值为 `zh`。
- **可切回英文**：将 `appearance.language` 改为 `en` 即可完整回退到英文界面。
- **实现方式**：基于自研 i18n 抽象层（`internal/i18n`），以英文原文为 key，未命中的文案自动回退英文，保证不会出现空白或遗漏。
- **术语表**：所有中文术语集中在 `internal/i18n/zh.go`，按模块分组，便于统一维护。
- **中文化规范**：见 `docs/i18n-plan.md`（译/不译规则、术语表、宽度对齐与快捷键约定）。

---

## 功能模块

Vortex 划分为多个专门模块，均可通过全局快捷键或命令面板访问。

### 核心模块

- **任务控制（Mission Control）**：基础设施总览，实时 CPU、内存、磁盘、网络遥测。
- **服务器管理（Servers）**：通过 SSH 密钥或密码管理、配置并连接多台远程服务器。
- **网络核心（Network）**：可视化监控服务器健康与基础设施连通性。

### 运维模块

- **进程管理（Processes）**：类似 htop 的进程查看与管理，可终止异常进程。
- **Docker 管理（Docker）**：监控容器、查看状态，并可快速进入运行中的容器 shell。
- **服务管理（Services）**：管理 systemd 服务（启动/停止/重启/启用/禁用）并查看日志。
- **日志查看器（Logs）**：查看并实时流式输出服务器上的各类日志。
- **文件管理（Files）**：浏览服务器文件系统、读取文件，并以提权方式写入受限文件。

### 安全与操作层

- **安全与防火墙（Security）**：管理 UFW 规则、查看防火墙状态、监控 Fail2Ban jail。
- **定时任务（Cron）**：创建、查看、安全编辑 cron 任务，实时显示 cron 语法的人性化翻译。
- **SSL 证书管理（Certs）**：监控证书、检查到期时间并轻松续期。
- **用户与 SSH 密钥（Users）**：管理系统用户及其 authorized SSH 密钥。

### 高级基础设施层

- **备份管理（Backups）**：创建与恢复备份、配置备份计划、管理保留策略。
- **数据库管理（Databases）**：查看并管理数据库配置。
- **反向代理（Proxy）**：动态配置与编辑反向代理规则。
- **密钥编辑器（Secrets）**：通过专用编辑器安全管理环境变量与密钥。
- **部署（Deployments）**：触发构建与重启命令，直接在终端监控部署日志。
- **可用性监控（Uptime Monitor）**：跟踪外部服务可用性与响应时间，可通过中央配置定制。
- **快照（Snapshots）**：配置文件在被修改前自动备份其旧版本，支持回滚。

---

## 配置

Vortex 通过 `config.yaml` 全面配置，默认位于 `~/.config/vortex/config.yaml`。配置文件允许你定制：

- **服务器（Servers）**：定义服务器集群，包含主机名、端口、用户名与认证方式。
- **外观（Appearance）**：自定义主题、壁纸与动画（含 `language` 语言选项）。
- **快捷键（Keybinds）**：重映射全局导航快捷键。
- **可用性监控目标（Uptime Targets）**：定义可用性模块监控的外部 URL 与端点。

---

## 安装

确保已安装 Go（建议 1.26.5 或更高版本）。

克隆仓库并构建二进制：

```bash
git clone https://github.com/JiangBeta/vortex-cn.git
cd vortex-cn
go build -o bin/vps-manager.exe ./cmd/vps-manager
```

### Nix

若已启用 flakes，可运行：

```bash
nix run github:JiangBeta/vortex-cn
```

或进入带 Go 与配套工具的开发环境：

```bash
nix develop
```

---

## 如何测试

1. **运行程序**：
   ```bash
   ./bin/vps-manager.exe
   ```

2. **连接服务器**：
   - 使用 `[` 或 `]` 在侧边栏中切换到「服务器」页（或按全局快捷键 `Esc`）。
   - 选中一台服务器，按 `Enter` 建立 SSH 连接。
   - 若尚未配置真实服务器，请编辑 `~/.config/vortex/config.yaml`，填入实际 VPS IP、用户名与 SSH 密钥路径。

3. **测试遥测**：
   - 连接后切换到「任务控制」（默认快捷键 `g`），应能看到实时波动的遥测数据（CPU、内存、磁盘）。

4. **测试各模块**：
   - **文件**：按 `f` 打开文件管理器，用方向键浏览目录，尝试编辑文件体验提示与保存行为。
   - **定时任务**：按 `x` 打开定时任务管理，尝试添加 cron 计划（如 `0 2 * * *`），观察输入时的人性化翻译实时出现。
   - **备份**：按 `k` 打开备份管理，尝试配置新的目标路径或计划。

5. **使用命令面板**：
   - 按 `Ctrl+P` 打开命令面板，输入「Uptime」「Secrets」等模糊搜索词，按 `Enter` 快速跳转。

---

## 键盘快捷键

Vortex 专为无鼠标操作设计。

- **[ / ]**：在侧边栏导航中循环切换。
- **Tab**：切换分屏视图，让侧边栏与当前模块并排显示。
- **Ctrl+P**：打开命令面板，模糊查找模块。
- **Esc**：返回服务器列表（当输入框未激活时）。
- **Enter**：确认操作、连接服务器或提交表单。
- **Ctrl+C**：退出程序。

> 注意：当你在输入框或文本编辑器中输入时，全局快捷键会自动挂起。

---

## License

本仓库遵循 [MIT License](LICENSE)。

Copyright (c) 2026 Vortex Contributors

本项目基于 [berkayyytech/vortex](https://github.com/berkayyytech/vortex)（MIT License）派生，仅做中文界面与体验优化，保留原作者的版权声明。上游许可证全文见 [LICENSE](LICENSE) / [LICENSE.md](LICENSE.md)。

---

## 致谢

- 上游项目：[berkayyytech/vortex](https://github.com/berkayyytech/vortex)
- 构建框架：[charmbracelet/bubbletea](https://github.com/charmbracelet/bubbletea)
