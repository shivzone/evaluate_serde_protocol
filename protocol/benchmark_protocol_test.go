package main

import (
    //"crypto/tls"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net"
    "net/http"
    "net/rpc"
    "net/rpc/jsonrpc"
    "testing"

    pb "github.com/evaluate_serde_protocol/protocol/agent"
    "golang.org/x/net/context"
    "google.golang.org/grpc"
    "google.golang.org/protobuf/proto"
)

var tcpHandler, jsonHandler, httpHandler, grpcHandler *AgentHandler
var httpServer, httpsServer *http.Server

type AgentHandler struct {}

type AgentData struct {
    Hostname    string   `json:"hostname"`
    Status      string   `json:"status"`
    Timestamp   int64    `json:"timestamp"`
    Lsns        []string `json:"lsns"`
}

func generateObject() *AgentData {
    return &AgentData{
        Hostname:   "10.64.6.138",
        Status:     "In Progress",
        Timestamp:  1282368345,
        Lsns:       []string{"16/B374D848", "16/B374D010"},
    }
}

func (th *AgentHandler) Serve(arg *string, reply *AgentData) error {
    temp := generateObject()
    reply.Hostname = temp.Hostname
    reply.Status = temp.Status
    reply.Lsns = temp.Lsns
    reply.Timestamp = temp.Timestamp
    return nil
}

func (th *AgentHandler) ServeAgentProto(ctx context.Context, in *pb.AgentRequest) (*pb.AgentProto, error) {
    obj := generateObject()

    return &pb.AgentProto{
       Hostname:  *proto.String(obj.Hostname),
       Status:    *proto.String(obj.Status),
       Timestamp: *proto.Int64(int64(obj.Timestamp)),
       Lsns:      obj.Lsns,
    }, nil
}

func (th *AgentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    out, _ := json.Marshal(generateObject())
    w.Write(out)
}

func startTCPRPCServer() {
    if tcpHandler != nil {
        return
    }
    tcpHandler = new(AgentHandler)
    rpc.Register(tcpHandler)

    tcpAddr, err := net.ResolveTCPAddr("tcp", ":8081")
    if err != nil {
        panic(err)
    }
    listener, err := net.ListenTCP("tcp", tcpAddr)
    if err != nil {
        panic(err)
    }
    go func() {
        for {
            conn, err := listener.Accept()
            if err != nil {
                continue
            }
            rpc.ServeConn(conn)
        }
    }()
}

func startJSONRPCServer() {
    if jsonHandler != nil {
        return
    }
    jsonHandler = new(AgentHandler)
    rpc.Register(jsonHandler)

    tcpAddr, err := net.ResolveTCPAddr("tcp", ":8082")
    if err != nil {
        panic(err)
    }
    listener, err := net.ListenTCP("tcp", tcpAddr)
    if err != nil {
        panic(err)
    }
    go func() {
        for {
            conn, err := listener.Accept()
            if err != nil {
                continue
            }
            jsonrpc.ServeConn(conn)
        }
    }()
}

func startGRPCServer() {
    if grpcHandler != nil {
        return
    }
    grpcServer := grpc.NewServer()
    grpcHandler = new(AgentHandler)
    listener, err := net.Listen("tcp", ":8084")
    if err != nil {
        panic(err)
    }
    pb.RegisterAgentServer(grpcServer, grpcHandler)
    go func() {
        err = grpcServer.Serve(listener)
        if err != nil {
            fmt.Println(err.Error())
        }
    }()
}

func startHTTPRPCServer() {
    if httpHandler != nil {
        return
    }
    httpHandler = new(AgentHandler)
    rpc.Register(httpHandler)
    rpc.HandleHTTP()

    go func() {
        err := http.ListenAndServe(":8083", nil)
        if err != nil {
            fmt.Println(err.Error())
        }
    }()
}

func startHTTPServer() {
    if httpServer != nil {
        return
    }

    httpServer = &http.Server{
        Handler: &AgentHandler{},
    }

    listener, err := net.Listen("tcp", ":8080")
    if err != nil {
        panic(err)
    }

    go func() {
        err := httpServer.Serve(listener)
        if err != nil {
            panic(err)
        }
    }()
}

func startHTTPSServer() {
    if httpsServer != nil {
        return
    }

    httpsServer = &http.Server{
        Handler: &AgentHandler{},
    }

    listener, err := net.Listen("tcp", ":8443")
    if err != nil {
        panic(err)
    }

    go func() {
        err := httpServer.ServeTLS(listener, "https-server.crt", "https-server.key")
        if err != nil {
            panic(err)
        }
    }()
}

func sendRequest(client *http.Client, addr string) {
    res, err := client.Get(addr)
    if err != nil {
        panic(err)
    }

    if res.StatusCode != 200 {
        panic("request failed")
    }

    _, err = ioutil.ReadAll(res.Body)
    if err != nil {
        panic(err)
    }

    err = res.Body.Close()
    if err != nil {
        panic(err)
    }
}

func BenchmarkHTTPRPC(b *testing.B) {
    startHTTPRPCServer()

    client, _ := rpc.DialHTTP("tcp", "127.0.0.1:8083")
    defer client.Close()

    b.ResetTimer()
    var reply AgentData
    for n := 0; n < b.N; n++ {
        err := client.Call("AgentHandler.Serve", string(n), &reply)
        if err != nil {
            panic(err)
        }
    }
}

func BenchmarkTCPRPC(b *testing.B) {
    startTCPRPCServer()

    client, _ := rpc.Dial("tcp", "127.0.0.1:8081")
    defer client.Close()

    b.ResetTimer()
    var reply AgentData
    for n := 0; n < b.N; n++ {
      err := client.Call("AgentHandler.Serve", string(n), &reply)
      if err != nil {
          panic(err)
      }
    }
}

func BenchmarkJSONRPC(b *testing.B) {
    startJSONRPCServer()

    client, _ := jsonrpc.Dial("tcp", "127.0.0.1:8082")
    defer client.Close()

    b.ResetTimer()
    var reply AgentData
    for n := 0; n < b.N; n++ {
        err := client.Call("AgentHandler.Serve", string(n), &reply)
        if err != nil {
            panic(err)
        }
    }
}

func BenchmarkGPRPC(b *testing.B) {
    startGRPCServer()

    conn, _ := grpc.Dial(":8084", grpc.WithInsecure())
    defer conn.Close()
    client := pb.NewAgentClient(conn)
    b.ResetTimer()
    for n := 0; n < b.N; n++ {
        _, err := client.ServeAgentProto(context.Background(), &pb.AgentRequest{Data: string(n)})
        if err != nil {
            panic(err)
        }
    }
}

func BenchmarkHTTP(b *testing.B) {
    startHTTPServer()

    client := &http.Client{}

    b.ResetTimer()
    for n := 0; n < b.N; n++ {
        sendRequest(client, "http://127.0.0.1:8080/")
    }
}

func BenchmarkHTTPNoKeepAlive(b *testing.B) {
    startHTTPServer()

    client := &http.Client{
        Transport: &http.Transport{
            DisableKeepAlives: true,
        },
    }

    b.ResetTimer()
    for n := 0; n < b.N; n++ {
        sendRequest(client, "http://127.0.0.1:8080/")
    }
}

//func BenchmarkHTTPSNoKeepAlive(b *testing.B) {
//    startHTTPSServer()
//
//    client := &http.Client{
//        Transport: &http.Transport{
//            DisableKeepAlives: true,
//            TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
//        },
//    }
//
//    b.ResetTimer()
//    for n := 0; n < b.N; n++ {
//        sendRequest(client, "https://127.0.0.1:8443/")
//    }
//}
