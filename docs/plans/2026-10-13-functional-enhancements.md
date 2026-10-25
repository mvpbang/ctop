# ctop 功能完善与体验升级实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 针对 ctop 核心交互与运维痛点，增加容器强制终止（Kill）、搜索无匹配空状态提示、日志查看器上下翻页回溯以及 Cgroup v2 内存指标兼容。

**Architecture:** 
1. 扩展 `connector/manager` 与 `container.Container` 接口及 Docker 实现，支持发送 SIGKILL 强杀容器并在操作菜单中暴露；
2. 在 `compact.CompactGrid` / `grid.go` 引入过滤为空时的友好引导提示；
3. 为 `widgets.TextView` 引入行偏移量 `Offset` 与滚动方法（Up/Down/PgUp/PgDown），并更新 `menus.go:LogMenu` 的按键绑定；
4. 优化 `collector/docker.go:ReadMem`，兼容 Cgroup v2 内存算法（回退使用 `inactive_file` 扣除页缓存）。

**Tech Stack:** Go 1.18, `github.com/fsouza/go-dockerclient`, `github.com/gizak/termui`, `github.com/nsf/termbox-go`

**Spec:** 根据审计报告中梳理的运维操作扩展、搜索空状态交互、日志查看瓶颈与 Cgroups v2 适配需求。

## Global Constraints

- 全量单元测试 `go test -v -race ./...` 必须全部通过，保持零 Data Race。
- 保证对旧 Docker 版本与已有快捷键的向下兼容。
- 遵循 TDD 规范：先编写失败测试，验证红灯，再实施最小代码改动，最后验证绿灯。

## Review Focus

- 当日志行数少于屏幕高度时滚动不能越界或 Panic。
- 当搜索匹配为空时不能阻断用户继续按退格或重新激活搜索框。
- Docker `Kill` 操作在容器已退出状态下应安全返回错误提示，不破坏应用状态。

---

### Task 1: 容器操作菜单增加 Kill（SIGKILL 强杀）功能

**Files:**
- Modify: `connector/manager/main.go`
- Modify: `connector/manager/docker.go`
- Modify: `connector/manager/mock.go`
- Modify: `connector/manager/runc.go`
- Modify: `container/main.go`
- Modify: `menus.go`
- Test: `connector/manager/docker_test.go`

- [x] **步骤 1：编写失败测试**
  在 `connector/manager/docker_test.go` 中验证 Manager 接口包含 `Kill()` 方法，且 `Mock` 实现了 `Kill()`。
- [x] **步骤 2：运行测试确认编译失败**
  `go test ./connector/manager`
- [x] **步骤 3：实现 Kill 方法与菜单按键**
  - 在 `Manager` 接口添加 `Kill() error`
  - 在 `docker.go` 实现 `Kill()`（调用 `client.KillContainer(api.KillContainerOptions{ID: dc.id, Signal: api.SIGKILL})`）
  - 在 `mock.go` 和 `runc.go` 实现 `Kill()`
  - 在 `container/main.go` 增加 `c.Kill()`
  - 在 `menus.go:ContainerMenu` 中为 running 状态容器增加 `menu.Item{Val: "kill", Label: "[k] kill"}` 及快捷键 `k`
- [x] **步骤 4：验证测试通过**
  `go test -v ./connector/manager`

---

### Task 2: 容器搜索过滤空状态提示（Empty State Hint）

**Files:**
- Modify: `cwidgets/compact/grid.go`
- Modify: `grid.go`
- Test: `cwidgets/compact/grid_test.go`

- [x] **步骤 1：编写失败测试**
  在 `cwidgets/compact/grid_test.go` 中测试当容器行数为空且设置了空状态提示时，`CompactGrid.Buffer()` 输出正确的提示字符。
- [x] **步骤 2：运行测试确认失败**
  `go test -v ./cwidgets/compact`
- [x] **步骤 3：实现空状态提示逻辑**
  - 在 `CompactGrid` 中添加 `emptyNotice *ui.Par` 与 `SetEmptyNotice(msg string)`
  - 在 `Buffer()` 中当 `len(cg.Rows) == 0 && cg.emptyNotice != nil` 时合并空状态提示
  - 在 `grid.go:RedrawRows` 中判断当 `cursor.Len() == 0 && config.GetVal("filterStr") != ""` 时设置提示信息
- [x] **步骤 4：验证测试通过**
  `go test -v ./cwidgets/compact`

---

### Task 3: 日志查看器增加上下滚动与导航支持

**Files:**
- Modify: `widgets/view.go`
- Modify: `menus.go`
- Test: `widgets/view_test.go`

- [x] **步骤 1：编写失败测试**
  在 `widgets/view_test.go` 中测试 `TextView` 的 `Up()`、`Down()`、`PageUp()`、`PageDown()` 方法及偏移限制。
- [x] **步骤 2：运行测试确认失败**
  `go test -v ./widgets`
- [x] **步骤 3：实现 TextView 滚动与 LogMenu 按键监听**
  - 为 `TextView` 增加 `offset int` 字段及 `Up()`, `Down()`, `PageUp()`, `PageDown()`
  - 调整 `renderLoop`：根据 `t.offset` 截取显示的行范围，不越界
  - 在 `menus.go:LogMenu` 注册 `up`, `down`, `pgup`, `pgdown` 绑定到 `m.Up`, `m.Down`, `m.PageUp`, `m.PageDown`，将退出按键收敛为 `q`, `Esc`, `Enter`
- [x] **步骤 4：验证测试通过**
  `go test -v ./widgets`

---

### Task 4: Cgroups v2 内存缓存扣除适配

**Files:**
- Modify: `connector/collector/docker.go`
- Test: `connector/collector/docker_test.go`

- [x] **步骤 1：编写失败测试**
  在 `connector/collector/docker_test.go` 中测试在 Cgroup v2 结构下（`Stats.Cache == 0` 但存在 `InactiveFile`）能够正确计算 Memory Usage。
- [x] **步骤 2：运行测试确认失败**
  `go test -v ./connector/collector`
- [x] **步骤 3：实现 Cgroup v2 兼容算法**
  在 `collector/docker.go:ReadMem` 中提取缓存值时优先检查 `Stats.Cache`，若为 0 则检查 `Stats.InactiveFile` / `TotalInactiveFile` 进行扣除，防止负数下溢。
- [x] **步骤 4：验证测试通过**
  `go test -v ./connector/collector`

---

### Task 5: 全量回归与竞态检测

- [x] 运行 `go vet ./...` 检查
- [x] 运行 `go test -v -race ./...` 验证全绿
- [x] 运行 `go build -v -o /dev/null .`
