# Protoc 插件 (protoc-gen-courier)

## 安装

```bash
go install github.com/simpossible/courier/cmd/protoc-gen-courier
```

确保 `$GOPATH/bin` 在你的 `PATH` 中。

## 使用

```bash
protoc \
  --go_out=. \
  --courier_out=. \
  -I/path/to/courier/options \
  user.proto
```

`--courier_out` 会在 proto 文件同目录下生成 `{name}.courier.go`。

## Proto 文件写法

```protobuf
syntax = "proto3";
package user;

// 引入 courier 的 cmd 扩展定义
import "options.proto";

message EmailCodeGetReq {
  string email = 1;
}

message EmailCodeGetResp {
  int32 code = 1;
  string msg = 2;
}

service RegisterService {
  // 每个方法必须标注 cmd 命令号
  rpc EmailGetCode(EmailCodeGetReq) returns (EmailCodeGetResp) {
    option (courier.cmd) = 10001;
  }

  rpc EmailRegister(EmailRegisterReq) returns (EmailRegisterResp) {
    option (courier.cmd) = 10002;
  }
}
```

## 生成代码结构

插件为每个 service 生成：

### 1. Handler 接口

```go
type RegisterServiceHandler interface {
    EmailGetCode(ctx *courier.Context, req *EmailCodeGetReq) (*EmailCodeGetResp, error)
    EmailRegister(ctx *courier.Context, req *EmailRegisterReq) (*EmailRegisterResp, error)
}
```

### 2. Adapter 闭包

```go
func registerServiceEmailGetCodeAdapter(h RegisterServiceHandler) courier.HandlerFunc {
    return func(ctx *courier.Context, raw []byte) ([]byte, error) {
        req := new(EmailCodeGetReq)
        if err := proto.Unmarshal(raw, req); err != nil {
            return nil, courier.ErrProtoError
        }
        resp, err := h.EmailGetCode(ctx, req)
        if err != nil {
            return nil, err
        }
        return proto.Marshal(resp)
    }
}
```

### 3. 注册函数

```go
func RegisterRegisterService(h RegisterServiceHandler) courier.ServiceInfo {
    info := registerServiceServiceInfo
    info.Methods = make([]courier.MethodInfo, len(registerServiceServiceInfo.Methods))
    copy(info.Methods, registerServiceServiceInfo.Methods)
    info.Methods[0].Handle = registerServiceEmailGetCodeAdapter(h)
    info.Methods[1].Handle = registerServiceEmailRegisterAdapter(h)
    return info
}
```

### 4. 使用

```go
handler := &myRegisterServiceImplementation{}
info := user.RegisterRegisterService(handler)
server.Register(info)
```

## 不使用插件

```go
server.Register(courier.ServiceInfo{
    ServiceName: "MyService",
    Methods: []courier.MethodInfo{
        {
            Cmd:  20001,
            Name: "Ping",
            Handle: func(ctx *courier.Context, raw []byte) ([]byte, error) {
                return []byte("pong"), nil
            },
        },
    },
})
```

适合快速原型或非 Protobuf 场景（JSON、MessagePack 等）。
