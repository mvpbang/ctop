# Changelog

本项目所有显著变更记录于此。格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.1.0/)。

## [Unreleased]

### Added

- 容器操作菜单支持 `kill`(SIGKILL 强杀,快捷键 `k`,带二次确认),仅对 running 状态容器展示
- 日志查看器支持滚动回溯:`up`/`down` 逐行、`pgup`/`pgdown` 翻页,`t` 切换时间戳,`q`/`Esc`/`Enter` 退出
- 容器过滤无匹配时展示空状态提示(含快捷键引导),不再显示空白表格
- 容器列表新增 `/` 快捷键,直接进入过滤输入框
- Makefile 构建管理:新增 `fmt`/`vet`/`test`/`check`/`all` 目标,完整 `.PHONY` 声明,校验和工具自动兼容 macOS(`shasum`)与 Linux(`sha256sum`)
- 新增 `CHANGELOG.md`

### Changed

- 内存用量计算对齐 Docker 原生算法:cgroup v2 优先使用 `inactive_file`/`total_inactive_file` 扣除页缓存,避免用量虚高与负数下溢
- panic 处理器默认不再向终端抛出堆栈,仅在 `CTOP_DEBUG=1` 时保留完整堆栈
- 排序字段未知时回退到 `state` 排序,避免空排序函数 panic

### Fixed

- `TextView` 日志视图的跨 goroutine 数据竞争:滚动偏移、时间戳开关、文本缓冲区与窗口尺寸改由读写锁保护
- `Container` 指标/元数据的并发访问:排序器改用带锁的 `GetMetrics`,动作方法改用带锁的 `GetMeta` 读取状态
- runc 连接器未初始化 `needsRefresh` 通道导致的潜在发送阻塞
- Mock 日志流关闭时正确退出 goroutine(`select` 内 `break` 无法跳出循环)并关闭通道
- 多处边界防护:日志行数少于屏幕高度、终端尺寸为 0、网格宽度过窄、空网格翻页等场景不再 panic
- exec 进入容器的 shell 探测回退、终端挂起与帧处理健壮性
- 输入框退格与符号输入
- 容器过滤支持多属性大小写不敏感匹配

## [0.7.7] - 2021-10-13

Fork 基线,取自上游 [bcicen/ctop](https://github.com/bcicen/ctop) v0.7.7。更早的历史请参见上游 Releases 页面。
