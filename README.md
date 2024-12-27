how to compile proto:
```
protoc --proto_path=protos protos/*.proto --go_out=$GOPATH/src --go-grpc_out=$GOPATH/src
```