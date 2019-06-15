## 缓存工具

#### 使用链表（list）和哈希表（map）实现的LRU缓存组件

#### 默认缓存容量343

#### 用法：

```
    cache := NewCache() // 使用默认容量
    val,err := cache.Get("key")  // 读取缓存
    
    if err != nil{
        fmt.println(val)
    }
    
    cache.Set("key","value")   // 设置缓存
    

```