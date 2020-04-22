# evaluate_serde_protocol

Benchmark JSON vs Protobuf vs GOB
```
go test -bench=. -benchmem
```

Benchmark TCP RPC vs JSON TCP RPC vs HTTP RPC vs GRPC VS HTTP vs HTTPNoKeepAlive
```
pushd protocol
go test -bench=. -benchmem
popd
```
