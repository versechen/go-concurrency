# go-concurrency
learning go concurrency

## [Once](./once.go)
全局只执行一次


## [Context](./context.go)
控制goroutine的生命周期
1. 取消
2. 超时
3. 传递元素
4. 链路跟踪

## 运行方式

每个文件都可以单独运行，直接指定文件名即可：

```bash
# 运行 context.go
go run context.go
# 或者编译后运行
go build -o context context.go
./context

# 运行 once.go
go run once.go
# 或者编译后运行
go build -o once once.go
./once
```

**说明**：每个文件都使用了 `//go:build ignore` 构建标签，这样：
- 每个文件都可以单独运行（`go run <file>`）
- 默认 `go build` 不会编译任何文件，避免包冲突
- 直接构建单个文件时，构建标签会被忽略，可以正常编译

## 调试方式

**重要**：由于使用了 `ignore` 构建标签，调试时需要指定具体的文件：

1. **使用 VS Code 调试**：
   - 打开要调试的文件（如 `context.go`）
   - 按 `F5` 或点击调试按钮
   - 选择 "Debug current file" 配置
   - 或者使用 `.vscode/launch.json` 中预配置的调试选项

2. **命令行调试**：
   ```bash
   # 调试 context.go
   dlv debug context.go

   # 调试 once.go
   dlv debug once.go
   ```

**注意**：如果调试器报错 "build constraints exclude all Go files"，请确保：
- 调试配置中的 `program` 指向具体的文件（如 `${file}` 或 `${workspaceFolder}/context.go`）
- 不要使用 `go build .` 构建整个包，而是构建单个文件
