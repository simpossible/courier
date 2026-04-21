# 迁移指南：Cipher → Courier

从旧版 `cipher.mikevillage.com` 项目迁移到 Courier 的对照表。

## 类型名对照

| Cipher 旧版 | Courier 新版 | 说明 |
|---|---|---|
| `CipherServiceInfo` | `ServiceInfo` | 去掉 Cipher 前缀 |
| `CipherMethodInfo` | `MethodInfo` | 同上 |
| `CipherMethod` | `HandlerFunc` | 签名变更 |
| `CipherPbParser` | 内联到 adapter | 不再单独存在 |
| `CipherContext` (空接口) | `*Context` (具体结构体) | 携带 DeviceID、RequestID 等 |
| `CipherServer` | `Server` | 实例化，无全局单例 |
| (无) | `Client` | 新增客户端抽象 |

## Handler 签名变更

**旧版：**
```go
func (s *Service) EmailGetCode(ctx rpc.CipherContext, req *EmailCodeGetReq) (*EmailCodeGetResp, error)
```

**新版：**
```go
func (h *Handler) EmailGetCode(ctx *courier.Context, req *EmailCodeGetReq) (*EmailCodeGetResp, error)
```

签名不变，只是 `CipherContext` → `*courier.Context`。

## 服务注册变更

**旧版（全局单例）：**
```go
// 在 plugin 生成的代码中直接调用全局函数
user.CipherRegisterRegisterService(myHandler)
// 内部调 rpc.RegisterService → rpc.mqttServer.RegisterService
```

**新版（显式实例）：**
```go
srv := rpc.NewServer(
    rpc.WithServerTransport(tp),
    rpc.WithServiceName("RegisterService"),
)
info := user.RegisterRegisterService(myHandler)
srv.Register(info)
srv.Start()
```

## Topic 命名变更

| 旧版 | 新版 |
|---|---|
| `/{ServiceName}` | `$share/{ServiceName}/mrpc/request/{ServiceName}` |
| `/resp/{deviceId}` | `mrpc/response/{deviceId}` |

## 全局单例移除

旧版 `rpc/base.go` 中有：

```go
var mqttServer CipherServer = nil
func SetMqttServer(server CipherServer) { ... }
func RegisterService(service interface{}, info CipherServiceInfo) { ... }
```

新版完全移除。每个 `Server` 实例独立管理自己的服务注册。

## 命令号 (cmd)

不变。proto 文件中的 `option (courier.cmd) = 10001` 继续有效，只是扩展定义从 `cipher.mikevillage.com/proto` 移到了 `github.com/simpossible/courier/options`。
