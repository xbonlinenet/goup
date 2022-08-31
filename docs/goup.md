- [goup](#goup)
  - [功能](#功能)
    - [DB 错误回调（v0.1.28）](#db-错误回调v0128)

# goup

## 功能

### DB 错误回调（v0.1.28）

在初始化的时候，可以通过 `frame.SetDbErrorCallback` 配置处理回调的函数：

```go
frame.BootstrapServer(
    ctx,
    frame.BeforeServerRun(registerRouter),
    frame.Version(version),
    frame.CustomRouter(customRouter),
    frame.ReportApi("http://192.168.0.22:14000/api/doc/report"),
    frame.CustomSqlConf(customSqlConf),
    frame.BeforeServerExit(func() {
        fmt.Println("it will be done before server shutdown")
    }),
    frame.SetDbErrorCallback(func(name, queryType, sql string, err error) {
        fmt.Printf("[DbErrorCallback] data name: %s, query type: %s, sql: %s, err: %s\n", name, queryType, sql, err)
    }),
)
```

测试日志：

```go
[DbErrorCallback] data name: local, query type: row_query, sql:  select count(*) from some_data_set, err: dial tcp 127.0.0.1:3306: connect: connection refused
```

默认没有任何回调处理函数，可以通过配置 `goup.db.default_error_callback` 为 `true` 使用框架默认的处理方法，将会使用 `sentry` 发送告警。
