## 控制台日志输出

#### 自定义日志输出格式

```
    提供三种级别的日志输出：
    
    
    Debug            调试
    Info             信息
    Warning          警告
    Error            错误

```

#### 用法

```
    log.Debug("debug")
    log.Info("info")
    log.Warning("warning")
    log.Error("error")

```

#### 设置

```
log.Threshold = "INFO"

// 设置文件方式输出的日志级别，高于和等于此级别的将输出至文件
// 文件输出等级：ERROR > WARNING > INFO > DEBUG


log.Dir = "./logs"

// 设置日志文件的输出文件夹


log.UseSetting()

// （刷新）应用设置

```