## 配置模块

### 支持使用unix配置风格，使用 .ini 格式的配置文件

#### 1、名称
```
    
    .ini 文件是Initialization File的缩写，即初始化文件。
    
```
#### 2、格式
```
    INI文件由节、键、值组成。
    
    节
    
    [section]
    
    键与值
    name=value
    
    注释
    使用符号#（或半角;或//），标识当前行此符号后面的字符为注释文本
```    
#### 3、示例
```
    [default]
    path= c:/go
    version = 1.44
     
    [test]
    num =	666
    something  = wrong  #注释1
    #fdfdfd = fdfdfd    注释整行
    refer= refer       //注释3     
    note = agian       ;注释3        
    
```    
#### 4、解析规则
```
    1、使用『节』标识不同的配置模块；
    2、使用键值对标识配置项，解析结果省略键或值前后两段的空格（或TAB制表符）
    3、每一行为一条配置项，同一条配置项不可使用换行
    4、使用#号或;号标识当前行此符后面的文本为注释，将不与解析到配置项
    5、解析结果完成为map[string]string;其中（map）的key名称为：section.key；即使用『节』
    名称 + 『.』 作为map结果集的键；配置项value作为结果集value。
    例如：  [server]
           host = 127.0.0.1
           
           key: server.host
           value: 127.0.0.1
    6、配置项键与值使用等号（=）连接
   
```


