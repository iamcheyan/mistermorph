# CLI Diff 测试工作区

这个目录用来测试 `mistermorph chat` 的代码比对（diff）功能。

## 测试文件

| 文件 | 语言 | 可修改的内容 |
|---|---|---|
| `hello.go` | Go | `greet()` 函数名、输出字符串 |
| `calc.py` | Python | `add()` / `multiply()` 函数逻辑 |
| `utils.js` | JavaScript | `greet()` 模板字符串、`sum()` 实现 |

## 快速测试

### 1. 编译

```bash
cd /home/tetsuya/Development/mistermorph
go build -o ./bin/mistermorph ./cmd/mistermorph
```

### 2. 运行 chat（指定 workspace 为这个目录）

```bash
./bin/mistermorph chat --workspace /home/tetsuya/Development/mistermorph/.local/test-diff
```

### 3. 输入指令让 agent 修改文件

**测试 1：修改已有文件**
```
把 hello.go 里的 greet 函数改成输出 "Hello, Mistermorph!"
```

**测试 2：新增函数**
```
在 calc.py 里加一个 subtract 函数
```

**测试 3：修改多行**
```
把 utils.js 的 greet 函数改成用模板字符串
```

### 4. 观察终端输出

修改成功后，终端应该实时打印 diff：

```
hello.go
───────
  5 │ 5 │ func greet() {
  6 │   │-    fmt.Println("Hello, World!")
    │ 6 │+    fmt.Println("Hello, Mistermorph!")
  7 │ 7 │ }
```

- 灰色行号（左旧右新）
- 红色 `-` 删除行
- 绿色 `+` 新增行
- 远距离未变更行自动折叠为 `···`

### 5. 验证文件确实被修改

```bash
cat /home/tetsuya/Development/mistermorph/.local/test-diff/hello.go
```

### 6. 恢复测试文件（可选）

```bash
cd /home/tetsuya/Development/mistermorph/.local/test-diff
git checkout -- hello.go calc.py utils.js
```

## 注意事项

- 这些文件不会被提交到 upstream（`.local/` 已在 `.gitignore` 中）
- 测试完可以随意修改、删除、恢复
- 如果 diff 没有显示，检查终端是否支持颜色（`NO_COLOR` 环境变量）
