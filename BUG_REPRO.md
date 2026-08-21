# Bug Reproduction

## 包的性质

当前 test_model_fix 保存的是被测模型修复后的结果源码，不是初始含 Bug 源码。要复现原始缺陷，必须检出下面固定的 parent SHA；不要在当前修复结果源码上期待重新出现修复前失败。生成系统使用的可信验证补丁和完整验证日志仅在本地留存，不提交到结果分支。

## 问题现象

货运核对 worker 正在等待外部系统返回时，运维触发了 pool 停止。监控显示停止信号已经发出，新任务也不再接收，但 Stop 一直不返回；只有再取消启动服务时传入的上层 context，在途任务才退出。请先不要修改代码，定位这段停机阻塞的根因，说明 pool、在途任务和上层 context 之间的生命周期归属及影响范围。

## 含 Bug 版本

- 仓库：VanceMichael/go-label-08
- 仓库地址：https://github.com/VanceMichael/go-label-08.git
- parent SHA：70665f59fdcfd3d221ccbf7c29363e5d09d7ce50

## 复现步骤

```bash
git clone -- https://github.com/VanceMichael/go-label-08.git bug-repro
cd bug-repro
git checkout --detach 70665f59fdcfd3d221ccbf7c29363e5d09d7ce50
go test ./internal/workerpool -run ^TestStopCancelsRunningJobsBeforeWaiting$ -count=1
```

## 双架构完整错误信息

### linux/amd64

- 容器内复现预期退出码：1
- 容器内复现实际退出码：1

stdout：

```text
$ go test ./internal/workerpool -run ^TestStopCancelsRunningJobsBeforeWaiting$ -count=1
--- FAIL: TestStopCancelsRunningJobsBeforeWaiting (0.16s)
    shutdown_test.go:56: Stop waited on a job whose context it did not cancel
FAIL
FAIL	github.com/VanceMichael/go-base-airbridge/internal/workerpool	0.186s
FAIL

```

stderr：

```text
(empty)
```

### linux/arm64

- 容器内复现预期退出码：1
- 容器内复现实际退出码：1

stdout：

```text
$ go test ./internal/workerpool -run ^TestStopCancelsRunningJobsBeforeWaiting$ -count=1
--- FAIL: TestStopCancelsRunningJobsBeforeWaiting (0.15s)
    shutdown_test.go:56: Stop waited on a job whose context it did not cancel
FAIL
FAIL	github.com/VanceMichael/go-base-airbridge/internal/workerpool	0.152s
FAIL

```

stderr：

```text
(empty)
```

## 通过条件

根因结论必须串起 Pool.Start 传递的运行 context、已进入 Job 后不再监听 doneCh、Stop 关闭接收信号后立即等待 worker，以及任务只响应外部 parent 取消的完整停机时序；同时说明为何新提交会被拒绝、为何二次取消上层 context 后 Stop 才恢复，以及无在途任务时的边界。以 TestStopCancelsRunningJobsBeforeWaiting 的 -race 红测复核，目标仓库代码、测试和配置保持零改动，不得把“WaitGroup 卡住”当作缺少生命周期因果链的最终结论。
