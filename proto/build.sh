#!/usr/bin/env bash
# 编译proto 文件
protoc --go_out=plugins=grpc:. ./*.proto

# protoc-gen-go下载
go get github.com/golang/protobuf/protoc-gen-go

# 安装grpc
go get -u google.golang.org/grpc