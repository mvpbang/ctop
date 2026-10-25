# 容器搜索功能实现计划

> **说明：** 采用 Native 模式（executing-plans）逐步执行本计划。各步骤使用勾选框 (`- [ ]`) 跟踪进度。

**目标：** 在 ctop 中通过快捷键 `/`（以及兼容的 `f`）随时唤起容器搜索功能，支持不区分大小写的多字段（容器名、容器 ID、镜像名）匹配，并完善终端退格与特殊字符输入支持。

**架构设计：**
1. 扩展 `container.Containers.Filter()`，实现基于容器名称（Name）、短 ID / 完整 ID（ID）、镜像名称（Image）的不区分大小写子串匹配，防止特殊字符导致正则异常。
2. 升级 `widgets.Input`，处理 macOS/Linux 终端发送的 `<backspace>`、`<delete>`、`C-8`、`C-h` 等退格键，支持空格及 `:`、`/` 等字符输入，增大搜索长度上限。
3. 在 `keys.go` 与 `grid.go` 中注册 `/` 键映射，绕过 termui 的 `cleanPath` 路径合并陷阱，使其精准触发 `FilterMenu`，并在帮助面板补充快捷键说明。

**技术栈：** Go 1.16+, `github.com/gizak/termui`, `github.com/nsf/termbox-go`

**设计规范（Spec）：** 2026-10-12 确认方案：支持按 `/` 随时唤起搜索、多属性（名称/ID/镜像）检索、忽略大小写、兼容终端退格键。

## 全局约束

- 保持与现有 `-f` 命令行参数和 `f` 快捷键的向后兼容。
- 全量单元测试 `go test ./...` 必须无报错通过。
- 搜索框中输入任何特殊符号（如 `[`、`*`、`(`、`\`）不能引发 panic 或程序崩溃。
- 不能直接使用 `ui.Handle` 注册 `/sys/kbd//`，因为 `cleanPath` 会将其折叠为 `/sys/kbd`，导致拦截所有终端输入。

## 重点审查项（Review Focus）

1. 包含正则元字符（如 `*`、`(`、`[`）的输入不能导致 `regexp` panic。
2. 终端退格键（`<backspace>`、`<delete>`、`C-8`、`C-h`）必须正常删除最后一个字符，而非被忽略或追加乱码。
3. 按键 `/` 只能触发搜索框，不能影响其他单字符按键的正常响应。
4. 搜索匹配必须完全忽略大小写（如输入 `WEB` 能匹配 `web-server`）。
5. 包含 `/` 或 `:` 的镜像名（如 `nginx:alpine`、`library/redis`）能够在搜索框中正常输入。

---

### 任务 1：容器多属性不区分大小写过滤逻辑

**文件清单：**
- 新建：`container/sort_test.go`
- 修改：`container/sort.go:113-128`

**接口契约：**
- 输入：`config.GetVal("filterStr")`、`c.GetMeta("name")`、`c.GetMeta("id")`、`c.Id`、`c.GetMeta("image")`、`c.GetMeta("state")`
- 输出：`Containers.Filter()` 设置各容器的 `c.Display = true/false`

- [x] **步骤 1：编写失败的单元测试**

新建 `container/sort_test.go`：
```go
package container

import (
	"testing"

	"github.com/bcicen/ctop/config"
	"github.com/bcicen/ctop/models"
)

func TestContainersFilter(t *testing.T) {
	config.Init()
	config.Update("allContainers", "true")

	c1 := &Container{Id: "abcdef123456789", Meta: models.NewMeta("id", "abcdef123456")}
	c1.SetMeta("name", "web-server")
	c1.SetMeta("image", "nginx:alpine")
	c1.SetMeta("state", "running")

	c2 := &Container{Id: "987654321fedcba", Meta: models.NewMeta("id", "987654321fed")}
	c2.SetMeta("name", "db-redis")
	c2.SetMeta("image", "redis:7.0")
	c2.SetMeta("state", "running")

	list := Containers{c1, c2}

	// 测试 1: 按名称忽略大小写匹配
	config.Update("filterStr", "WEB")
	list.Filter()
	if !c1.Display || c2.Display {
		t.Errorf("expected c1 true, c2 false for 'WEB'; got c1=%v, c2=%v", c1.Display, c2.Display)
	}

	// 测试 2: 按镜像名匹配
	config.Update("filterStr", "redis")
	list.Filter()
	if c1.Display || !c2.Display {
		t.Errorf("expected c1 false, c2 true for 'redis'; got c1=%v, c2=%v", c1.Display, c2.Display)
	}

	// 测试 3: 按短 ID 匹配
	config.Update("filterStr", "abcdef")
	list.Filter()
	if !c1.Display || c2.Display {
		t.Errorf("expected c1 true, c2 false for 'abcdef'; got c1=%v, c2=%v", c1.Display, c2.Display)
	}

	// 测试 4: 特殊正则字符不应引发崩溃
	config.Update("filterStr", "nginx[")
	list.Filter()
	if c1.Display || c2.Display {
		t.Errorf("expected neither to match 'nginx['; got c1=%v, c2=%v", c1.Display, c2.Display)
	}
}
```

- [x] **步骤 2：运行测试验证其失败**

运行：`go test -v ./container -run TestContainersFilter`
预期：失败（大小写敏感、缺少镜像/ID匹配、或正则遇到 `nginx[` 发生 panic）。

- [x] **步骤 3：在 `container/sort.go` 中实现过滤逻辑**

使用 `strings.Contains` 和 `strings.ToLower` 实现不区分大小写的包含匹配：
```go
func (a Containers) Filter() {
	filter := strings.ToLower(config.GetVal("filterStr"))

	for _, c := range a {
		c.Display = true
		if filter != "" {
			name := strings.ToLower(c.GetMeta("name"))
			image := strings.ToLower(c.GetMeta("image"))
			id := strings.ToLower(c.Id)
			shortId := strings.ToLower(c.GetMeta("id"))

			if !strings.Contains(name, filter) &&
				!strings.Contains(image, filter) &&
				!strings.Contains(id, filter) &&
				!strings.Contains(shortId, filter) {
				c.Display = false
			}
		}

		// 状态过滤保持不变
		if !config.GetSwitchVal("allContainers") && c.GetMeta("state") != "running" {
			c.Display = false
		}
	}
}
```

- [x] **步骤 4：运行测试验证其通过**

运行：`go test -v ./container -run TestContainersFilter`
预期：PASS

- [x] **步骤 5：提交代码**

```bash
git add container/sort.go container/sort_test.go
git commit -m "feat(container): add case-insensitive multi-attribute filtering"
```

---

### 任务 2：完善输入框退格键与搜索符号支持

**文件清单：**
- 新建：`widgets/input_test.go`
- 修改：`widgets/input.go:9-85`

**接口契约：**
- 输入：`ui.Event`，按键路径如 `/sys/kbd/<backspace>`、`/sys/kbd/C-8`、`/sys/kbd/<delete>`、`/sys/kbd/<space>` 等
- 输出：更新 `i.Data` 并向 `i.Stream()` 推送变动

- [x] **步骤 1：编写失败的单元测试**

新建 `widgets/input_test.go`：
```go
package widgets

import (
	"testing"

	ui "github.com/gizak/termui"
)

func TestInputKeyPress(t *testing.T) {
	input := NewInput()
	_ = input.Stream()

	// 输入常规字母
	input.KeyPress(ui.Event{Path: "/sys/kbd/a"})
	input.KeyPress(ui.Event{Path: "/sys/kbd/b"})
	if input.Data != "ab" {
		t.Fatalf("expected 'ab', got '%s'", input.Data)
	}

	// 通过 <backspace> 退格
	input.KeyPress(ui.Event{Path: "/sys/kbd/<backspace>"})
	if input.Data != "a" {
		t.Fatalf("expected 'a' after <backspace>, got '%s'", input.Data)
	}

	// 通过 <delete> 退格
	input.KeyPress(ui.Event{Path: "/sys/kbd/<delete>"})
	if input.Data != "" {
		t.Fatalf("expected '' after <delete>, got '%s'", input.Data)
	}

	// 特殊符号及空格输入
	input.KeyPress(ui.Event{Path: "/sys/kbd/c"})
	input.KeyPress(ui.Event{Path: "/sys/kbd/:"})
	input.KeyPress(ui.Event{Path: "/sys/kbd//"})
	input.KeyPress(ui.Event{Path: "/sys/kbd/<space>"})
	input.KeyPress(ui.Event{Path: "/sys/kbd/d"})
	if input.Data != "c:/ d" {
		t.Fatalf("expected 'c:/ d', got '%s'", input.Data)
	}
}
```

- [x] **步骤 2：运行测试验证其失败**

运行：`go test -v ./widgets -run TestInputKeyPress`
预期：失败（无法识别 `<backspace>`、符号未在可输入列表中等）。

- [x] **步骤 3：修改 `widgets/input.go`**

1. 扩展 `input_chars` 包含常用符号与空格：
```go
var (
	input_chars = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ-_.:/ "
)
```
2. 在 `KeyPress` 中兼容 `<backspace>`、`<delete>`、`C-8`、`C-h` 与 `<space>`，并增加长度容量至 50：
```go
func (i *Input) KeyPress(e ui.Event) {
	ch := strings.Replace(e.Path, "/sys/kbd/", "", -1)
	if ch == "<space>" {
		ch = " "
	}
	if ch == "C-8" || ch == "<backspace>" || ch == "<delete>" || ch == "C-h" {
		idx := len(i.Data) - 1
		if idx > -1 {
			i.Data = i.Data[0:idx]
			if i.stream != nil {
				i.stream <- i.Data
			}
		}
		ui.Render(i)
		return
	}
	if len(i.Data) >= i.MaxLen {
		return
	}
	if strings.Contains(input_chars, ch) {
		i.Data += ch
		if i.stream != nil {
			i.stream <- i.Data
		}
		ui.Render(i)
	}
}
```

- [x] **步骤 4：运行测试验证其通过**

运行：`go test -v ./widgets -run TestInputKeyPress`
预期：PASS

- [x] **步骤 5：提交代码**

```bash
git add widgets/input.go widgets/input_test.go
git commit -m "feat(widgets): improve input widget backspace and symbol support"
```

---

### 任务 3：注册 `/` 快捷键与主界面检索入口集成

**文件清单：**
- 新建：`grid_test.go`
- 修改：`keys.go:36-42`
- 修改：`grid.go:182-186`
- 修改：`menus.go:19-35`

**接口契约：**
- 输入：终端按键 `/`（路径 `/sys/kbd//`）与 `f`（路径 `/sys/kbd/f`）
- 输出：`HandleKey(path string, f func(ui.Event))`，主界面监听并唤起 `FilterMenu()`，帮助列表显示说明

- [x] **步骤 1：编写失败的单元测试**

新建 `grid_test.go`：
```go
package main

import (
	"testing"

	ui "github.com/gizak/termui"
)

func TestSlashKeyBinding(t *testing.T) {
	ui.DefaultEvtStream.ResetHandlers()

	called := false
	HandleKey("/sys/kbd//", func(e ui.Event) {
		called = true
	})

	if _, ok := ui.DefaultEvtStream.Handlers["/sys/kbd//"]; !ok {
		t.Fatalf("expected /sys/kbd// handler to exist")
	}

	if _, ok := ui.DefaultEvtStream.Handlers["/sys/kbd"]; ok {
		t.Fatalf("/sys/kbd must not be registered directly")
	}

	fn := ui.DefaultEvtStream.Handlers["/sys/kbd//"]
	fn(ui.Event{Path: "/sys/kbd//"})
	if !called {
		t.Fatalf("expected handler to be invoked")
	}
}
```

- [x] **步骤 2：运行测试验证其失败**

运行：`go test -v . -run TestSlashKeyBinding`
预期：失败（`undefined: HandleKey`）。

- [x] **步骤 3：实现 `HandleKey` 并集成快捷键**

1. 在 `keys.go` 中添加 `HandleKey`：
```go
// 直接注册按键处理器，避免被 cleanPath 折叠
func HandleKey(path string, f func(ui.Event)) {
	ui.DefaultEvtStream.Lock()
	defer ui.DefaultEvtStream.Unlock()
	ui.DefaultEvtStream.Handlers[path] = f
}
```

2. 在 `grid.go` 中绑定 `/` 与 `f`：
```go
	openFilter := func(ui.Event) {
		menu = FilterMenu
		ui.StopLoop()
	}
	ui.Handle("/sys/kbd/f", openFilter)
	HandleKey("/sys/kbd//", openFilter)
```

3. 在 `menus.go` 的 `helpDialog` 中添加说明：
```go
	{"[/] - filter/search containers", ""},
	{"[f] - filter/search containers", ""},
```

- [x] **步骤 4：运行测试验证其通过**

运行：`go test -v . -run TestSlashKeyBinding`
预期：PASS

- [x] **步骤 5：运行全量测试套件**

运行：`go test -v ./...`
预期：全部通过

- [x] **步骤 6：提交代码**

```bash
git add keys.go grid.go menus.go grid_test.go
git commit -m "feat(grid): bind slash key to container search"
```
