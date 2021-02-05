package client

import (
	"fmt"
	"github.com/Gyjnine/nas-grpc-file/common"
	"github.com/Gyjnine/nas-grpc-file/data"
	"github.com/Gyjnine/nas-grpc-file/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

var MaxFileSize = 1024 * 1024 * 256
var InitialWindowSize = int32(1024 * 1024 * 5)
var BufferSize = 1024 * 1024 * 5

//type FileHandler interface {
//	InitConnection(add string) error
//	DescribeFile(callerCode string, remoteFullPath string, timeout time.Duration) (ret data.RetDescribeFile, err error)
//	CreateFile(callerCode string, mountPath string, xType string, fileData []byte, fileName string, filePath string, replace bool, timeout time.Duration) (ret data.RetCreateFile, err error)
//	ModifyFile(filePath string, fileName string, callerCode string, isReplace bool, timeout time.Duration) (ret data.Ret, err error)
//	CopyFile(newFilePath string, originalFilePath string, callerCode string, timeout time.Duration) (ret data.Ret, err error)
//	MoveFile(newFilePath string, originalFilePath string, callerCode string, timeout time.Duration) (ret data.Ret, err error)
//}

type FileClient struct {
	UserCli proto.FileWorkerClient
	address string
	channel *grpc.ClientConn
}

func (f *FileClient) InitConnection(addr string) error {
	address := fmt.Sprintf(addr)
	conn, err := grpc.Dial(address, grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(MaxFileSize),
			grpc.MaxCallSendMsgSize(MaxFileSize)),
		grpc.WithInitialWindowSize(InitialWindowSize),
		grpc.WithInitialConnWindowSize(InitialWindowSize),
		grpc.WithWriteBufferSize(BufferSize),
		grpc.WithReadBufferSize(BufferSize))
	if err != nil {
		return err
	}
	// 存根
	f.UserCli = proto.NewFileWorkerClient(conn)
	f.address = address
	f.channel = conn
	return nil
}

func (f *FileClient) DescribeFile(callerCode string, remoteFullPath string, timeout time.Duration) (ret data.RetDescribeFile, err error) {
	if callerCode == "" || remoteFullPath == "" {
		ret.Code = 128502
		ret.Message = "缺少请求必传参数"
		return ret, nil
	}
	code, message, fileStream, err := common.CallDescribeFile(f.UserCli, callerCode, remoteFullPath, timeout)
	if code == 1 || code == 128509 || code == 128512 {
		if f.channel != nil {
			_ = f.channel.Close()
		}
		conn, err := grpc.Dial(f.address, grpc.WithInsecure(),
			grpc.WithDefaultCallOptions(
				grpc.MaxCallRecvMsgSize(MaxFileSize),
				grpc.MaxCallSendMsgSize(MaxFileSize)),
			grpc.WithInitialWindowSize(InitialWindowSize),
			grpc.WithInitialConnWindowSize(InitialWindowSize),
			grpc.WithWriteBufferSize(BufferSize),
			grpc.WithReadBufferSize(BufferSize))
		if err != nil {
			return ret, err
		}
		// 存根
		UserCli := proto.NewFileWorkerClient(conn)
		code, message, fileStream, err = common.CallDescribeFile(UserCli, callerCode, remoteFullPath, timeout)
		if err != nil {
			return ret, err
		}
		f.UserCli = UserCli
	}
	ret.Code = code
	ret.Message = message
	ret.FileStream = fileStream
	return ret, err
}

func (f *FileClient) CreateFile(callerCode string, mountPath string, xType string, fileData []byte, fileName string, filePath string, replace bool, timeout time.Duration) (ret data.RetCreateFile, err error) {
	if callerCode == "" || filePath == "" || xType == "" || fileName == "" {
		ret.Code = 128502
		ret.Message = "缺少请求必传参数"
		return ret, nil
	}
	code, message, fileMountPath, err := common.CallCreateFile(f.UserCli, callerCode, fileName, fileData, replace, xType, mountPath, filePath, timeout)
	// 重新建立连接
	if code == 1 || code == 128509 || code == 128512 {
		if f.channel != nil {
			_ = f.channel.Close()
		}
		conn, err := grpc.Dial(f.address, grpc.WithInsecure(),
			grpc.WithDefaultCallOptions(
				grpc.MaxCallRecvMsgSize(MaxFileSize),
				grpc.MaxCallSendMsgSize(MaxFileSize)),
			grpc.WithInitialWindowSize(InitialWindowSize),
			grpc.WithInitialConnWindowSize(InitialWindowSize),
			grpc.WithWriteBufferSize(BufferSize),
			grpc.WithReadBufferSize(BufferSize))
		if err != nil {
			return ret, err
		}
		// 存根
		UserCli := proto.NewFileWorkerClient(conn)
		code, message, fileMountPath, err = common.CallCreateFile(UserCli, callerCode, fileName, fileData, replace, xType, mountPath, filePath, timeout)
		if err != nil {
			return ret, err
		}
		f.UserCli = UserCli
	}
	ret.Code = code
	ret.Message = message
	ret.FileMountPath = fileMountPath
	return ret, err
}

func (f *FileClient) ModifyFile(filePath string, fileName string, callerCode string, isReplace bool, timeout time.Duration) (ret data.Ret, err error) {
	if callerCode == "" || filePath == "" || fileName == "" {
		ret.Code = 128502
		ret.Message = "缺少请求必传参数"
		return ret, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()
	dataInfo, err := f.UserCli.ModifyFile(ctx, &proto.ModifyRequest{FilePath: filePath, FileName: fileName, FCode: callerCode, IsReplace: isReplace})
	if err != nil {
		errStatus, _ := status.FromError(err)
		if errStatus.Code() == codes.DeadlineExceeded {
			ret.Code = 128512
			ret.Message = "连接超时"
		} else if errStatus.Code() == codes.Unavailable {
			// 连接暂时不可用 重新尝试
			dataInfo, err = f.UserCli.ModifyFile(ctx, &proto.ModifyRequest{FilePath: filePath, FileName: fileName, FCode: callerCode, IsReplace: isReplace})
			if err != nil {
				ret.Code = 128509
				ret.Message = "读取超时"
				return ret, err
			}
			ret.Code = int(dataInfo.Code)
			ret.Message = dataInfo.Err
		} else if errStatus.Code() == codes.InvalidArgument {
			ret.Code = 128501
			ret.Message = "参数异常"
		}
	} else {
		ret.Code = int(dataInfo.Code)
		ret.Message = dataInfo.Err
	}
	return ret, err
}

func (f *FileClient) CopyFile(newFilePath string, originalFilePath string, callerCode string, timeout time.Duration) (ret data.Ret, err error) {
	if newFilePath == "" || originalFilePath == "" || callerCode == "" {
		ret.Code = 128502
		ret.Message = "缺少请求必传参数"
		return ret, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()
	dataInfo, err := f.UserCli.CopyFile(ctx, &proto.CopyRequest{NewFilePath: newFilePath, OriginalFilePath: originalFilePath, FCode: callerCode})
	if err != nil {
		errStatus, _ := status.FromError(err)
		if errStatus.Code() == codes.DeadlineExceeded {
			ret.Code = 128512
			ret.Message = "连接超时"
		} else if errStatus.Code() == codes.Unavailable {
			// 连接暂时不可用 重新尝试
			dataInfo, err = f.UserCli.CopyFile(ctx, &proto.CopyRequest{NewFilePath: newFilePath, OriginalFilePath: originalFilePath, FCode: callerCode})
			if err != nil {
				ret.Code = 128509
				ret.Message = "读取超时"
				return ret, err
			}
			ret.Code = int(dataInfo.Code)
			ret.Message = dataInfo.Err
		} else if errStatus.Code() == codes.InvalidArgument {
			ret.Code = 128501
			ret.Message = "参数异常"
		}
	} else {
		ret.Code = int(dataInfo.Code)
		ret.Message = dataInfo.Err
	}
	return ret, err
}

func (f *FileClient) MoveFile(newFilePath string, originalFilePath string, callerCode string, timeout time.Duration) (ret data.Ret, err error) {
	if newFilePath == "" || originalFilePath == "" || callerCode == "" {
		ret.Code = 128502
		ret.Message = "缺少请求必传参数"
		return ret, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()
	dataInfo, err := f.UserCli.MoveFile(ctx, &proto.MoveRequest{NewFilePath: newFilePath, OriginalFilePath: originalFilePath, FCode: callerCode})
	if err != nil {
		errStatus, _ := status.FromError(err)
		if errStatus.Code() == codes.DeadlineExceeded {
			ret.Code = 128512
			ret.Message = "连接超时"
		} else if errStatus.Code() == codes.Unavailable {
			// 连接暂时不可用 重新尝试
			dataInfo, err = f.UserCli.MoveFile(ctx, &proto.MoveRequest{NewFilePath: newFilePath, OriginalFilePath: originalFilePath, FCode: callerCode})
			if err != nil {
				ret.Code = 128509
				ret.Message = "读取超时"
				return ret, err
			}
			ret.Code = int(dataInfo.Code)
			ret.Message = dataInfo.Err
		} else if errStatus.Code() == codes.InvalidArgument {
			ret.Code = 128501
			ret.Message = "参数异常"
		}
	} else {
		ret.Code = int(dataInfo.Code)
		ret.Message = dataInfo.Err
	}
	return ret, err
}
