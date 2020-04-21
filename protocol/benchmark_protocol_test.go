package main

import (
    "crypto/tls"
    "fmt"
    "io/ioutil"
    "net"
    "net/http"
    "net/rpc"
    "net/rpc/jsonrpc"
    "testing"

    "google.golang.org/grpc"
)

var tcpHandler, jsonHandler, httpHandler *TestHandler
var grpcHandler *grpc.Server
var httpServer, httpsServer *http.Server

type TestHandler struct {
}

func (th *TestHandler) Serve(arg *int, reply *string) error {
    *reply = "OK.\n"
    return nil
}

func (th *TestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/plain")
    w.Write([]byte("OK.\n"))
}

func startTCPRPCServer() {
    if tcpHandler != nil {
        return
    }
    tcpHandler = new(TestHandler)
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
    jsonHandler = new(TestHandler)
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

func startHTTPRPCServer() {
    if httpHandler != nil {
        return
    }
    httpHandler = new(TestHandler)
    rpc.Register(httpHandler)
    rpc.HandleHTTP()

    go func() {
        err := http.ListenAndServe(":8083", nil)
        if err != nil {
            fmt.Println(err.Error())
        }
    }()
}

func startGRPCServer() {
    if grpcHandler != nil {
        return
    }
    grpcHandler = grpc.NewServer()
    listener, err := net.Listen("tcp", ":8084")
    if err != nil {
        panic(err)
    }
    go func() {
        err = grpcHandler.Serve(listener)
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
        Handler: &TestHandler{},
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
        Handler: &TestHandler{},
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

func BenchmarkTCPRPC(b *testing.B) {
    startTCPRPCServer()

    client, _ := rpc.Dial("tcp", "127.0.0.1:8081")
    defer client.Close()

    b.ResetTimer()
    var reply string
    for n := 0; n < b.N; n++ {
      err := client.Call("TestHandler.Serve", n, &reply)
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
    var reply string
    for n := 0; n < b.N; n++ {
        err := client.Call("TestHandler.Serve", n, &reply)
        if err != nil {
            panic(err)
        }
    }
}

func BenchmarkHTTPRPC(b *testing.B) {
    startHTTPRPCServer()

    client, _ := rpc.DialHTTP("tcp", "127.0.0.1:8083")
    defer client.Close()

    b.ResetTimer()
    var reply string
    for n := 0; n < b.N; n++ {
        err := client.Call("TestHandler.Serve", n, &reply)
        if err != nil {
            panic(err)
        }
    }
}

func BenchmarkGPRPC(b *testing.B) {
    startGRPCServer()

    client, _ := grpc.Dial(":8084", grpc.WithInsecure())
    defer client.Close()

    // TODO: Pending grpc
    //b.ResetTimer()
    //var reply string
    //for n := 0; n < b.N; n++ {
    //    err := client.Call("TestHandler.Serve", n, &reply)
    //    if err != nil {
    //        panic(err)
    //    }
    //}
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

func BenchmarkHTTPSNoKeepAlive(b *testing.B) {
    startHTTPSServer()

    client := &http.Client{
        Transport: &http.Transport{
            DisableKeepAlives: true,
            TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
        },
    }

    b.ResetTimer()
    for n := 0; n < b.N; n++ {
        sendRequest(client, "https://127.0.0.1:8443/")
    }
}
