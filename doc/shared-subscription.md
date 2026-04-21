# 共享订阅 ($share)

## 原理

MQTT 5.0 规范（以及 EMQX、Mosquitto 等 MQTT 3.1.1 broker 的扩展）支持共享订阅语法：

```
$share/{group_name}/{real_topic}
```

同一个 `group_name` 下的多个订阅者形成一个消费组。Broker 收到消息后，
只投递给其中一个订阅者，而不是广播给所有人。这就是天然的负载均衡。

## Courier 中的使用

### Topic 命名

```
服务端订阅:  $share/{serviceName}/mrpc/request/{serviceName}
客户端发布:  mrpc/request/{serviceName}
```

例如 `RegisterService`：

```
Server A 订阅: $share/RegisterService/mrpc/request/RegisterService
Server B 订阅: $share/RegisterService/mrpc/request/RegisterService
Server C 订阅: $share/RegisterService/mrpc/request/RegisterService

Client 发布:   mrpc/request/RegisterService
```

### 多实例水平扩展

```go
// 节点 1
srv1 := rpc.NewServer(
    rpc.WithServerTransport(tp1),
    rpc.WithServiceName("RegisterService"),
    rpc.WithSharedSubscribe(true),  // 默认就是 true
)
srv1.Register(user.RegisterRegisterService(&handler{}))
srv1.Start()

// 节点 2（完全相同的代码，只是不同的 transport 实例）
srv2 := rpc.NewServer(
    rpc.WithServerTransport(tp2),
    rpc.WithServiceName("RegisterService"),
    rpc.WithSharedSubscribe(true),
)
srv2.Register(user.RegisterRegisterService(&handler{}))
srv2.Start()
```

新节点上线后 Broker 自动将其加入分发，下线后自动移除。无需服务注册、心跳或配置中心。

### 关闭共享订阅

如果你的 broker 不支持 `$share`，可以关闭：

```go
srv := rpc.NewServer(
    rpc.WithSharedSubscribe(false),
    // ...
)
```

此时订阅普通 topic `mrpc/request/{serviceName}`，每个实例都会收到全量消息。

## Broker 兼容性

| Broker | `$share` 支持 | 说明 |
|--------|---------------|------|
| EMQX | 支持 | 原生支持，推荐 |
| Mosquitto | 支持 (≥2.0) | 2.0 及以上版本支持 `$share` |
| HiveMQ | 支持 | 原生支持 |
| VerneMQ | 支持 | 原生支持 |
| ActiveMQ | 不确定 | 需要验证 |

## 断线重订阅

`$share` 订阅是基于连接的。断线重连后必须重新订阅。
Courier 的 `transport.MQTTTransport` 在 `OnConnectHandler` 中自动重新订阅所有 topic，
包括 `$share` 共享订阅，无需手动处理。

## 负载均衡策略

负载均衡策略由 Broker 实现，Courier 不控制。常见策略：

- **round-robin** — 轮询分发
- **random** — 随机选择
- **sticky** — 基于 clientID 哈希（部分 broker 支持）

如果需要粘性路由（同一 client 总是打到同一 server），需要在 broker 侧配置。
