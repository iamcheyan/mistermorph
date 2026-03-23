---
date: 2026-03-24
title: CLI Binary 持有 Console SPA 资源方案
status: draft
---

# CLI Binary 持有 Console SPA 资源方案

## 1) 背景

当前桌面端的事实架构已经是两进程：

1. Wails 桌面壳负责窗口与生命周期
2. 桌面壳拉起一个 backend child process
3. child process 运行 `mistermorph console serve`
4. WebView 访问 child process 暴露的本地 HTTP 服务

当前真正的问题不是“两进程”，而是静态资源归属不对。

现在 `console serve` 负责提供 Console HTTP，但它自己不持有前端资源。
它只能依赖：

- `--console-static-dir`
- 或现有配置项 `console.static_dir`

如果运行环境里没有外部 `web/console/dist`，桌面启动会直接失败：

- `failed to start desktop host: cannot find console static assets directory`

这说明现在的发布语义不完整：提供 HTTP 的进程本身并不是一个完整的 Console server。

## 2) 结论

本次方案选择：

1. 由 `mistermorph` CLI/backend binary 自己持有 Console SPA 资源。
2. `mistermorph console serve` 在没有显式静态目录时，自动回退到 embedded assets。
3. 桌面壳继续只负责拉起 backend binary，不再负责前端资源装配。
4. `--desktop-console-serve` 从目标方案中移除，不再保留这条特殊内部模式。

这是比“让 desktop binary 自己持有前端”更干净的职责划分。

## 3) 设计目标

1. `console serve` 成为真正完整的 Console host。
2. 打包后的桌面应用不再依赖包外 `web/console/dist`。
3. CLI、桌面、systemd、未来其他宿主，都围绕同一个 `console serve` 行为工作。
4. 构建链路即使忘了 stage，也不能默默产出一个运行后只会 404 的坏 binary。

## 4) 非目标

这次不做：

1. 不改成单进程桌面架构。
2. 不切到 Wails asset handler 直出 UI。
3. 不移除 `--console-static-dir` 和 `console.static_dir`，它们继续保留为显式覆盖入口。
4. 不保留 `--desktop-console-serve` 作为长期兼容特性。

## 5) 关键判断

### 5.1 `console.static_dir` 不是新配置

`console.static_dir` 是现有配置项，不是这次新加的。

当前仓库里已经有：

- `assets/config/config.example.yaml`
- `docs/feat/feat_20260219_httpd_admin_console.md`

都在描述它的行为。

本次只是把它从“唯一资源来源之一”扩展为“显式覆盖项”，默认路径改成 embedded fallback。

### 5.2 为什么不让 desktop binary 持有资源

如果资源只在 desktop binary 里，会天然引入一个额外耦合：

1. 只有当 child process 是 desktop executable 自己重启自己时，才能直接消费同一份 embedded assets
2. 一旦桌面壳拉起的是外部 `mistermorph` backend binary，资源就断了

这会把：

- launcher 选择
- 资源来源
- fallback 逻辑

绑死在一起。

而如果资源在 CLI/backend binary 里：

1. 谁提供 HTTP，谁持有静态资源
2. 桌面壳不用理解资源怎么来
3. 两进程结构保持稳定

代价只是 CLI 体积变大，但这是明确接受的权衡。

## 6) 总体方案

目标行为：

```text
mistermorph console serve
  -> 优先使用 --console-static-dir / console.static_dir
  -> 若未显式指定，则使用 embedded Console SPA
  -> 若 embedded 资源损坏或缺失，进程启动立刻 hard-fail

desktop host
  -> 只负责找到并拉起 mistermorph console serve
  -> 不再负责查找 web/console/dist
```

这意味着：

1. Console 资源跟着 backend binary 走
2. 桌面端不再需要特殊的前端资源注入路径
3. `--desktop-console-serve` 这条分支可以删除

## 7) 构建与目录方案

### 7.1 `go:embed` 约束

`go:embed` 不能直接从 `cmd/mistermorph/consolecmd` 去引用仓库根的 `web/console/dist`。
所以仍然必须有一个 staging 步骤。

### 7.2 staging 目标目录

不使用 `embeddedconsole/dist/`，直接简化为：

```text
cmd/mistermorph/consolecmd/static/
```

staging 行为是把 `web/console/dist` 的内容直接复制到 `consolecmd/static/` 下。

结果类似：

```text
cmd/mistermorph/consolecmd/static/index.html
cmd/mistermorph/consolecmd/static/assets/...
```

这样路径更短，也更符合它作为“package-private embedded static root”的语义。

### 7.3 staging 脚本

新增：

```text
scripts/stage-console-assets.sh
```

职责：

1. 校验 `web/console/dist/index.html` 存在
2. 清理并重建 `cmd/mistermorph/consolecmd/static/`
3. 复制 dist 内容到 `consolecmd/static/`
4. 输出清晰日志

建议用 `.gitignore` 忽略 `static/` 下生成内容，只保留占位文件。

## 8) 反脆弱性要求

staging 是整个方案里最脆弱的一环，因为：

1. `go:embed` 必须吃包路径下的真实文件
2. 一旦 staging 没跑，`go build` 仍可能成功
3. 这会生成一个缺前端资源的坏 binary

因此本方案要求在 `console_assets.go` 中加入启动期 hard-fail 校验：

1. 通过 `go:embed all:static`
2. 在 `init()` 中检查 embedded FS 里必须存在 `index.html`
3. 如果不存在，直接 `panic`

目标不是优雅降级，而是让损坏的构建在进程一启动就暴露。

这里故意选 hard-fail，而不是默默 404。
因为这种错误本质上是构建链路损坏，不是运行时用户输入错误。

## 9) 代码改造点

### 9.1 `cmd/mistermorph/consolecmd`

新增静态资源双来源抽象：

- `staticDir string`
- `staticFS fs.FS`

行为规则：

1. 显式 `--console-static-dir`
2. 否则读取已有配置项 `console.static_dir`
3. 若都为空，则使用 embedded `staticFS`

需要统一改造：

1. `loadServeConfig`
2. `handleSPA`
3. `serveSPAIndex`
4. 静态文件读取与响应逻辑

### 9.2 `cmd/mistermorph/consolecmd/console_assets.go`

新增这个文件，职责固定：

1. `go:embed all:static`
2. 暴露 `fs.Sub(...)` 后的静态资源 FS
3. 在 `init()` 中校验 `index.html`
4. 缺失时直接 `panic`

### 9.3 `desktop/wails/host.go`

host 需要被简化，而不是继续长出更多静态资源逻辑。

改造后：

1. 不再把“解析 console 静态目录”作为桌面启动前提
2. 不再强制给 child process 传 `--console-static-dir`
3. 继续负责找到 backend binary、启动、等 health、代理 WebView

也就是说，host 的职责重新收缩为：

- 找 backend
- 起 backend
- 代理 backend

### 9.4 删除 `--desktop-console-serve`

这条 flag 在目标方案里不再必要。

原因很简单：

1. 资源已经统一进 `consolecmd`
2. 桌面壳只需要起 `mistermorph console serve`
3. 继续保留一个只给桌面内部模式用的特殊 flag，只会制造概念噪音

因此目标方案应包含：

1. 删除 `desktop/wails/console_mode.go`
2. 删除 `desktopConsoleServeArgV1`
3. 删除相关分支逻辑和测试

## 10) Windows bundle 变更

这部分不是一句“带上 sibling backend binary”就结束了，它是明确的发布流程调整。

### 10.1 现状问题

当前 Windows 发布物是一个裸的 desktop exe。

如果桌面壳继续是两进程，而 `console serve` 的前端资源又改到了 backend binary 里，那么：

- 只发 desktop exe 将不再成立
- 因为桌面壳没有 child process 可拉起

### 10.2 目标发布形态

Windows 发布物需要从“单 exe”切到“bundle 目录或 zip 包”。

建议发布形态：

```text
mistermorph-desktop-windows-amd64.zip
  /MisterMorph.exe
  /mistermorph.exe
  /LICENSE
  /README-desktop.txt
```

其中：

1. `MisterMorph.exe` 是 Wails 桌面壳
2. `mistermorph.exe` 是 backend binary，内嵌 Console SPA

### 10.3 desktop host 行为

Windows 下 host 启动顺序要明确写死为：

1. 优先找 sibling `mistermorph.exe`
2. 找不到才考虑 PATH 或下载 fallback

对正式发布包而言，sibling backend 是一等路径，不是可选优化。

### 10.4 CI / release workflow 变更

Windows workflow 需要明确新增：

1. `pnpm --dir web/console build`
2. `./scripts/stage-console-assets.sh`
3. `go build -o dist/mistermorph.exe ./cmd/mistermorph`
4. `go build -tags ... -o dist/MisterMorph.exe ./desktop/wails`
5. 组装 bundle 目录
6. 打 zip
7. 对 zip 产物计算 checksum

也就是说，发布资产名称和结构都会变：

- 从裸 `mistermorph-desktop-windows-amd64.exe`
- 变成带 backend 的 `.zip`

这不是附带修补，而是正式的分发模型变更。

### 10.5 文档与用户预期

Windows 用户文档也要同步改：

1. 下载的是 zip bundle，不是单 exe
2. 解压后必须保持 `MisterMorph.exe` 和 `mistermorph.exe` 同目录
3. 不再支持“只复制桌面 exe 到别处单独运行”

## 11) 本地构建与 CI

所有构建 `./cmd/mistermorph` 的路径，都必须先做 staging：

1. 本地 `go build ./cmd/mistermorph`
2. `scripts/build-desktop.sh`
3. 所有 release workflow

标准顺序：

```text
pnpm --dir web/console install
pnpm --dir web/console build
./scripts/stage-console-assets.sh
go build ./cmd/mistermorph
go build -tags ... ./desktop/wails
```

任何绕过这个顺序的发布构建，都应被视为损坏构建。

## 12) 测试计划

### 12.1 `consolecmd`

新增 `fstest.MapFS` 场景：

1. 无 `staticDir` 时可以从 `staticFS` 提供普通资源
2. `index.html` 仍会注入 base path
3. API 路径不会被 SPA handler 吞掉
4. 显式目录覆盖 embedded FS

### 12.2 启动期校验

新增测试，确认：

1. embedded FS 缺少 `index.html` 时会 hard-fail
2. 不是运行到首个 HTTP 请求才暴露问题

### 12.3 桌面端

桌面端测试重点变成：

1. host 不再要求 `--console-static-dir`
2. host 优先找 sibling backend binary
3. host 启动 `mistermorph console serve` 后能通过 `/health`

### 12.4 Windows bundle smoke test

需要补一条发布级 smoke test：

1. 组装临时 bundle 目录
2. 目录里放 `MisterMorph.exe` + `mistermorph.exe`
3. 从 bundle 根目录启动桌面壳
4. 确认 child process 正常被拉起

## 13) 结论

这次要修正的不是桌面进程数，而是资源归属和发布语义。

正确落点是：

1. `mistermorph console serve` 自己携带并提供 Console SPA
2. `console.static_dir` 继续保留，但只作为显式覆盖
3. 构建链路通过 `consolecmd/static/` staging + `init()` hard-fail 校验防止坏构建
4. 桌面壳只负责拉起 backend binary
5. Windows 发布改成包含 backend 的 bundle，而不是裸 desktop exe

这样改完以后，系统边界会更清楚，发布语义也会和真实运行形态一致。
