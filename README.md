# zinc-client

> 同步日志文件到 [zincsearch](https://github.com/zincsearch/zincsearch)

说明：为了能够快速将普通文本导入zincsearch而写，内置了Grok的解析器，能够处理单行文本，该工具只是实验性产物，临时解决自己的需求，如果有同类需求可以参考本项目进行二次开发。

## 使用方法

```shell
Usage of zinc-client:
  -file string
        log filepath (default "access.log")
  -index string
        zincsearch index name (default "index_name")
  -password string
        zincsearch password (default "admin")
  -thread int
        thread num (default 4)
  -url string
        zincsearch api host (default "http://localhost:4080")
  -username string
        zincsearch username (default "admin")
```

## 示例

```shell
$ zinc-client -file access.log -index index_name -url http://localhost:4080 -username admin -password admin
```