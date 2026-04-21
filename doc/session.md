# Session 管理

Courier 使用服务端 Session 代替 JWT token，利用 MQTT 长连接的特性，
登录后以 ClientID 为 key 查询 session，无需每个请求都携带 token。

## Broker 要求

Session 的安全性依赖 **ClientID 不可伪造**。要求使用以下之一：

- **EMQX** — 支持 `$share` 共享订阅 + 连接认证，推荐
- **Mosquitto ≥ 2.0** — 支持 `$share`，配合 auth plugin 使用
- **其他支持共享订阅的 MQTT 5.0 Broker**

核心保证：MQTT 协议限制同一 ClientID 只能有一个活跃连接。配合 Broker 认证（username/password 或 TLS），ClientID 不可被冒充。

## 为什么不用 JWT

| | JWT | Session |
|---|---|---|
| 每请求额外带宽 | ~350 字节 | 0（ClientID 已在帧头中） |
| 10w 活跃用户月带宽成本 | ≈ ¥2.6 万 | ≈ ¥0 |
| 100w 活跃用户月带宽成本 | ≈ ¥26 万 | ≈ ¥500（Redis） |
| 服务端状态 | 无 | 需要（内存或 Redis） |

MQTT 连接本身是有状态的，JWT "无状态"的优势不存在。高频场景下 Session 成本是 JWT 的百分之一。

## 使用方式

### 1. 创建 SessionStore

**单实例（内存）：**

```go
store := rpc.NewMemorySessionStore()
stopCleanup := store.StartCleanup(5*time.Minute, 30*time.Minute)
defer stopCleanup()
```

**多实例（Redis）：**

```go
rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
store := rpc.NewRedisSessionStore(rdb,
    rpc.WithRedisMaxAge(30*time.Minute),
)
```

### 2. 登录时创建 Session

```go
func (h *LoginHandler) EmailLogin(ctx *rpc.Context, req *pb.EmailLoginReq) (*pb.EmailLoginResp, error) {
    user, err := h.userService.Verify(req.Email, req.Password)
    if err != nil {
        return nil, rpc.NewError(1000006, "密码错误")
    }

    // 用 ClientID 作为 session key
    h.sessionStore.Set(ctx.ClientID, &rpc.Session{
        UserID: user.ID,
        Data:   map[string]string{"role": user.Role},
    })

    return &pb.EmailLoginResp{Code: 0, Msg: "OK"}, nil
}
```

### 3. 配置 Session 拦截器

```go
srv := rpc.NewServer(
    rpc.WithServerTransport(tp),
    rpc.WithServiceName("UserService"),
    rpc.WithServerInterceptors(
        rpc.SessionInterceptor(store, rpc.PublicCmds(
            10001, // EmailGetCode
            10002, // EmailRegister
            10003, // EmailLogin
        )),
    ),
)
```

白名单中的命令（登录、注册等）跳过 session 检查，其他命令必须已登录。

### 4. Handler 中读取 Session

```go
func (h *ProfileHandler) GetProfile(ctx *rpc.Context, req *pb.GetProfileReq) (*pb.GetProfileResp, error) {
    // ctx.Session 在拦截器中已注入
    userID := ctx.Session.UserID

    profile := h.userService.GetProfile(userID)
    return &pb.GetProfileResp{Code: 0, Nickname: profile.Nickname}, nil
}
```

### 5. 登出时销毁 Session

```go
func (h *UserHandler) Logout(ctx *rpc.Context, req *pb.LogoutReq) (*pb.LogoutResp, error) {
    h.sessionStore.Delete(ctx.ClientID)
    return &pb.LogoutResp{Code: 0, Msg: "OK"}, nil
}
```

## Session 结构

```go
type Session struct {
    UserID     string            // 用户 ID
    Data       map[string]string // 应用扩展数据
    CreatedAt  time.Time         // 创建时间
    LastActive time.Time         // 最后活跃时间
}
```

## Context 中的字段

```go
type Context struct {
    Cmd         uint32    // 当前命令号（用于白名单判断）
    ClientID    string    // 客户端标识（来自帧头，session key）
    Session     *Session  // 登录后由拦截器自动注入
    // ...
}
```

## 多实例部署

使用 `$share` 共享订阅时，同一个 ClientID 的请求可能分发到不同实例。
此时必须用 Redis 作为 session store：

```go
rdb := redis.NewClient(&redis.Options{
    Addr: "redis://redis-cluster:6379",
})
store := rpc.NewRedisSessionStore(rdb, rpc.WithRedisMaxAge(30*time.Minute))
```

Redis 中的 session 自带 TTL，过期自动删除，不需要手动清理。

## 安全模型

```
1. 客户端连接 Broker:
   ClientID = "device-abc"
   Broker 认证: username/password 或 TLS

2. 客户端发送请求:
   帧头中 ClientID = "device-abc"（必须与连接 ClientID 一致）

3. 服务端验证:
   - Broker 保证 ClientID 唯一（一个 ClientID 只有一个连接）
   - Broker 认证保证只有合法设备能连接
   - 所以帧头中的 ClientID 可信
   - 用 ClientID 查 Session，确认已登录
```
