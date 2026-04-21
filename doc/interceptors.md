# 拦截器

拦截器是包裹 `HandlerFunc` 的函数，用于添加横切关注点（日志、指标、认证、恢复等）。
用法与 gRPC interceptor 一致。

## 类型签名

```go
type Interceptor func(next HandlerFunc) HandlerFunc
```

## 使用方式

```go
srv := rpc.NewServer(
    rpc.WithServerInterceptors(
        RecoveryInterceptor(),
        LoggingInterceptor(logger),
        AuthInterceptor(),
    ),
)
```

拦截器按传入顺序执行：第一个最外层，最后一个最接近业务逻辑。

## 内置示例

### Panic 恢复

```go
func RecoveryInterceptor() rpc.Interceptor {
    return func(next rpc.HandlerFunc) rpc.HandlerFunc {
        return func(ctx *rpc.Context, raw []byte) (resp []byte, err error) {
            defer func() {
                if r := recover(); r != nil {
                    err = rpc.NewError(500, "internal server error")
                }
            }()
            return next(ctx, raw)
        }
    }
}
```

### 请求日志

```go
func LoggingInterceptor(logger *slog.Logger) rpc.Interceptor {
    return func(next rpc.HandlerFunc) rpc.HandlerFunc {
        return func(ctx *rpc.Context, raw []byte) ([]byte, error) {
            start := time.Now()
            resp, err := next(ctx, raw)
            logger.Info("rpc call",
                "service", ctx.ServiceName,
                "method", ctx.MethodName,
                "device", ctx.DeviceID,
                "duration", time.Since(start),
                "error", err,
            )
            return resp, err
        }
    }
}
```

### 认证

```go
func AuthInterceptor() rpc.Interceptor {
    return func(next rpc.HandlerFunc) rpc.HandlerFunc {
        return func(ctx *rpc.Context, raw []byte) ([]byte, error) {
            token, ok := ctx.GetMeta("authorization")
            if !ok || !validateToken(token) {
                return nil, rpc.NewError(401, "unauthorized")
            }
            return next(ctx, raw)
        }
    }
}
```

## 手动组合

```go
chain := rpc.ChainInterceptors(
    RecoveryInterceptor(),
    LoggingInterceptor(logger),
)
// chain 是一个 Interceptor，可以传给 WithServerInterceptors
```
