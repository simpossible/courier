# 配置参考

## Transport 配置

Transport 是 MQTT 连接的抽象层，通过 functional options 配置。

```go
tp := transport.NewMQTTTransport(
    transport.WithBrokers("tcp://broker1:1883", "tcp://broker2:1883"),
    transport.WithClientID("my-service-node1"),
    transport.WithAutoReconnect(true),
    transport.WithConnectRetry(true),
    transport.WithKeepAlive(60 * time.Second),
    transport.WithPingTimeout(30 * time.Second),
    transport.WithMaxReconnectInterval(1 * time.Minute),
    transport.WithCleanSession(true),
    transport.WithDefaultQoS(0),
)
```

| Option | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `WithBrokers` | `...string` | (必填) | MQTT broker 地址列表 |
| `WithClientID` | `string` | `courier_{timestamp}` | MQTT 客户端标识 |
| `WithAutoReconnect` | `bool` | `true` | 是否自动重连 |
| `WithConnectRetry` | `bool` | `true` | 初始连接是否重试 |
| `WithKeepAlive` | `time.Duration` | `60s` | MQTT 心跳间隔 |
| `WithPingTimeout` | `time.Duration` | `30s` | 心跳超时 |
| `WithMaxReconnectInterval` | `time.Duration` | `1m` | 最大重连间隔 |
| `WithCleanSession` | `bool` | `true` | 是否清除 session |
| `WithDefaultQoS` | `byte` | `0` | 发布/订阅 QoS 等级 |
| `WithOnConnect` | `func()` | `nil` | 连接成功回调（在内部重订阅之后） |
| `WithOnConnectionLost` | `func(error)` | `nil` | 断线回调 |

## Server 配置

```go
srv := rpc.NewServer(
    rpc.WithServerTransport(tp),
    rpc.WithServiceName("UserService"),
    rpc.WithSharedSubscribe(true),
    rpc.WithServerInterceptors(loggingInterceptor, recoveryInterceptor),
)
```

| Option | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `WithServerTransport` | `Transport` | (必填) | 传输层实例 |
| `WithServiceName` | `string` | (必填) | 服务名，决定 MQTT topic 和 `$share` group |
| `WithSharedSubscribe` | `bool` | `true` | 是否使用 `$share` 共享订阅 |
| `WithServerInterceptors` | `...Interceptor` | `[]` | 服务端拦截器链 |

## Client 配置

```go
client := rpc.NewClient(
    rpc.WithClientTransport(tp),
    rpc.WithClientID("device-abc123"),
    rpc.WithTimeout(10 * time.Second),
    rpc.WithRetry(3, 1*time.Second, 1.5),
)
```

| Option | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `WithClientTransport` | `Transport` | (必填) | 传输层实例 |
| `WithClientID` | `string` | (必填) | 客户端标识，必须与 MQTT ClientID 一致，用于响应路由和 session |
| `WithTimeout` | `time.Duration` | `10s` | 单次请求超时 |
| `WithRetry` | `(int, Duration, float64)` | `0, 1s, 1.5` | 重试次数、初始间隔、退避因子 |
| `WithClientInterceptors` | `...Interceptor` | `[]` | 客户端拦截器链 |
