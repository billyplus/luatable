# luatable
根据lua格式修改的配置表，

## 格式说明

格式如下：

```
Config1 = {
    {
        id=1,
        name="name1",
    },
    {
        id=2,
        name="name2",
    }
},
Config2 = {
    name1 = {
        id=1,
        name="name1",
    },
    name2 = {
        id=2,
        name="name2",
    }
}
```

解析成map为：

```
map[Config1:[map[id:1 name:name1] map[id:2 name:name2]] Config2:map[name1:map[id:1 name:name1] name2:map[id:2 name:name2]]]
```

再将map转成json：

```
{"Config1":[{"id":1,"name":"name1"},{"id":2,"name":"name2"}],"Config2":{"name1":{"id":1,"name":"name1"},"name2":{"id":2,"name":"name2"}}}
```

## 使用方式

和json包里面的Unmarshal一样

```
    var result interface{}
    err := luatable.Unmarshal(tc.data, &result)

    if assert.NoError(err, tc.name) {
        
    }
```


## TODO

- [x] Unmarshal：将字符串转成对象
- [ ] Marshal：将对象转成字符串
