## 功能介绍

##### 这是一个Go语言工具组件，用于开发Web应用；参考SpringMVC建立Controller的方式及编程习惯。
##### 集成内置配置文件读取、LRU缓存、日志组件；经过高并发长时间段、服务端内存观察、压力等测试，高效、简洁、安全。
##### 可用于生产环境。


#### 〇、下载与引用

```
go get -u github.com/GuanglinGao/cynex

或（右上角绿色按钮）打包zip文件，解压至[gopath]/src中
```


#### 一、快速入门

##### 1、创建程序启动文件main.go，在main方法中启动HttpServer
```
package main

import (
	"cynex/react"
)

func main() {
	react.Server.Start()
}

```

##### 2、创建控制器文件controller/user.go；在文件中定义控制器类型User，在init方法中绑定路径处理方法
```
package controller

import (
	"cynex/log"
	"cynex/react"
	"net/http"
)

type User struct {
}

func init() {
	react.BindGet("/index", new(User), "Index")
}

func (u *User) Index(w http.ResponseWriter, r *http.Request) {
	log.Debug("Index 方法已经执行")
}


```

##### 3、使用默认配置，执行main方法。在浏览器可访问绑定路径的地址。




#### 二、绑定请求路径与处理方法，Controller层API

##### 1、静态路径

```
react.BindGet("/index", new(User), "Index")

// react.BindGet("【此处为浏览器请求路径】", 【此处为响应处理组件指针】, "【此处为处理方法名，将使用此方法处理请求】")
```
##### 2、正则，正则表达式使用小括号包裹，完全占满子路径一格（必须位于两个斜杠中间，或处于路径末尾）
```
react.BindGet("/index/([0-9])", new(User), "Index")

// 匹配 /index/1 等
```
##### 3、常用替换，子路径（双斜线中间的路径标识）中，使用中括号包裹，完全占满子路径一格
```
react.BindGet("/index/[*dex]", new(User), "Index")

// 匹配 /index/index，另有如下三种用法

react.BindGet("/index/[*de*]", new(User), "Index")
react.BindGet("/index/[inde*]", new(User), "Index")
react.BindGet("/index/[*]", new(User), "Index")
```
##### 4、路径变量，使用大括号包裹，占满子路径一格；大括号中的字段为变量名称
```
react.BindGet("/index/{name}", new(User), "Index")

// 匹配 /index/John 时，在FormValue中添加名称为name值为John的变量
```
##### 5、处理器组件与处理器方法
```
type User struct {
}

// 处理器组件如上，处理器可任意建立

func (u *User) Index(w http.ResponseWriter, r *http.Request) {
	log.Debug("Index 方法已经执行")
}

// 处理器方法，如上；链接至处理器，方法参数必须为 http.ResponseWriter、*http.Request，方法参数数量及前后顺序均不可修改。

```

#### 三、Server配置

##### 1、导出变量
```
Config       map[string]string // 配置项
DownloadDir  string            // 文件下载路径
HTTPPort     int               // HTTP端口
HTTPSPort    int               // HTTPS端口
HTTPSEnabled bool              // 是否启用HTTPS
TLSKeyDir    string            // HTTPS私钥文件
TLSCertDir   string            // HTTPS证书文件

```

##### 2、启动示例

```
package main

import (
	_ "cynex-test/controller"
	
	"cynex/react"
)

func main() {
	react.Server.Start()
}

```

##### 3、使用.ini 文件作为服务启动配置；默认加载程序运行目录及conf子目录下的所有INI文件
```
// wd := os.Getwd()
// 例如，wd作为工作目录（程序可执行文件所在的文件夹）；服务启动将会加载此目录和 wd/conf 中的全部INI文件作为配置

// 默认配置文件，示例如下；对应的参数名称不可修改

[default]
download_dir = .
log_dir = ./logs
log_threshold = info

[http]
port = 9090

[https]
enable = true
port = 8080
key_dir = conf/server.key
cert_dir = conf/server.crt

```
[INI语法格式参考](./conf/conf.md "INI语法参考")


#### 四、静态文件及文件下载路径绑定
```
react.BindDownload("/download", "2019-04-11.log")

// react.BindDownload("【此处为浏览器的请求地址】", "【此处为文件相对与下载目录的路径，下载目录使用配置文件配置】")

react.BindStatic("/static", "/")

// react.BindStatic("【此处为浏览器的请求路径，该路径下的全部子路径文件为静态文件】", "【此处为服务器中的文件夹路径】")

```

#### 五、前置处理与后置处理

##### 使用管道方式为处理方法设置前置处理方法与后置处理方法
```
react.PipelineBefore(new(User), "Index", new(User), "Before")

// react.PipelineBefore(【目标组件的指针】, "【目标方法】", 【前置组件的指针】, "【前置方法，此方法将在目标方法前执行】")


react.PipelineBefore(new(User), "Before", new(User), "BeforeBefore")

react.PipelineAfter(new(User), "Index", new(User), "After")

// react.PipelineAfter(【目标组件的指针】, "【目标方法】", 【后置组件的指针】, "【后置方法，此方法将在目标方法后执行】")

react.PipelineAfter(new(User), "After", new(User), "AfterAfter")

```


#### 六、其他引用

[1、LRU缓存](./cache/cache.md "其他组件依赖LRU缓存实现")

[2、日志](./log/log.md "日志输出")

[3、配置文件读取](./conf/conf.md "应用启动配置读取")

#### 七、交流群

![交流群](./Cynex框架实战群二维码.png)
     
