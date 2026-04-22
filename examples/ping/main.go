package main

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/simpossible/courier/rpc"
	"github.com/simpossible/courier/transport"
)

type pingReq struct {
	ClientID string `json:"client_id"`
	Message  string `json:"message"`
}

type pingResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data string `json:"data"`
}

func main() {
	broker := "tcp://115.190.2.139:1883"
	username := "financial"
	password := "financial"

	// --- Server ---
	serverTp := transport.NewMQTTTransport(
		transport.WithBrokers(broker),
		transport.WithClientID("courier-server-001"),
		transport.WithUsername(username),
		transport.WithPassword(password),
	)

	srv := rpc.NewServer(
		rpc.WithServerTransport(serverTp),
		rpc.WithServiceName("PingService"),
		rpc.WithSharedSubscribe(false),
	)

	srv.Register(rpc.ServiceInfo{
		ServiceName: "PingService",
		Methods: []rpc.MethodInfo{
			{
				Cmd:  1,
				Name: "Ping",
				Handle: func(ctx *rpc.Context, raw []byte) ([]byte, error) {
					var req pingReq
					if err := json.Unmarshal(raw, &req); err != nil {
						return nil, err
					}
					log.Printf("[server] ping from %s: %s", req.ClientID, req.Message)
					ctx.ClientID = req.ClientID
					resp := pingResp{Code: 0, Msg: "OK", Data: "pong at " + time.Now().Format(time.RFC3339)}
					return json.Marshal(resp)
				},
			},
		},
	})

	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("[server] start failed: %v", err)
		}
		log.Println("[server] started, waiting for requests...")
	}()

	time.Sleep(5 * time.Second)

	// --- Client ---
	clientTp := transport.NewMQTTTransport(
		transport.WithBrokers(broker),
		transport.WithClientID("courier-client-001"),
		transport.WithUsername(username),
		transport.WithPassword(password),
	)

	client := rpc.NewClient(
		rpc.WithClientTransport(clientTp),
		rpc.WithClientID("courier-client-001"),
		rpc.WithTimeout(10*time.Second),
		rpc.WithRetry(2, 2*time.Second, 1.5),
	)

	if err := client.Connect(); err != nil {
		log.Fatalf("[client] connect failed: %v", err)
	}
	defer client.Close()

	log.Println("[client] connected")

	for i := 1; i <= 3; i++ {
		req := pingReq{ClientID: "courier-client-001", Message: fmt.Sprintf("ping #%d", i)}
		payload, _ := json.Marshal(req)

		respBytes, err := client.Call(context.Background(), "PingService", 1, payload)
		if err != nil {
			log.Printf("[client] ping #%d failed: %v", i, err)
			continue
		}

		var resp pingResp
		json.Unmarshal(respBytes, &resp)
		fmt.Printf("[client] ping #%d → code=%d msg=%s data=%s\n", i, resp.Code, resp.Msg, resp.Data)
		time.Sleep(1 * time.Second)
	}

	fmt.Println("\n=== Done ===")
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
}

func init() {
	_ = rand.Read
}
