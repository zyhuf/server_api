# server_api

Compile proto file:
```
protoc --gofast_out=plugins=grpc:. *.proto
```
After compiled, edit the go source code to remove unnecessary import packages.