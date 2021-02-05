package data

import (
	"github.com/Gyjnine/nas-grpc-file/proto"
	"google.golang.org/grpc"
)

type Ret struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type FileClient struct {
	UserCli proto.FileWorkerClient
	address string
	channel *grpc.ClientConn
}
