package client

import (
	"flag"
	"fmt"
	"github.com/Gyjnine/nas-grpc-file/common"
	"github.com/Gyjnine/nas-grpc-file/proto"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

var MaxFileSize = 1024 * 1024 * 256
var InitialWindowSize = int32(1024 * 1024 * 5)
var BufferSize = 1024 * 1024 * 5

type FileHandler interface {
	InitConnection(add string)
	DescribeFile(callerCode string, remoteFullPath string, timeout time.Duration) (map[string]interface{}, error)
	CreateFile(callerCode string, mountPath string, xType string, fileData []byte, fileName string, filePath string, replace bool, timeout time.Duration) (map[string]interface{}, error)
	ModifyFile(filePath string, fileName string, callerCode string, isReplace bool, timeout time.Duration) (map[string]interface{}, error)
	CopyFile(newFilePath string, originalFilePath string, callerCode string, timeout time.Duration) (map[string]interface{}, error)
	MoveFile(newFilePath string, originalFilePath string, callerCode string, timeout time.Duration) (map[string]interface{}, error)
}

type FileClient struct {
	UserCli proto.FileWorkerClient
	address string
	channel *grpc.ClientConn
}

func (f FileClient) InitConnection(addr string, logger logrus.Logger) {
	address := *flag.String("host", addr, "")
	conn, err := grpc.Dial(address, grpc.WithInsecure(),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(MaxFileSize),
			grpc.MaxCallSendMsgSize(MaxFileSize)),
		grpc.WithInitialWindowSize(InitialWindowSize),
		grpc.WithInitialConnWindowSize(InitialWindowSize),
		grpc.WithWriteBufferSize(BufferSize),
		grpc.WithReadBufferSize(BufferSize))
	if err != nil {
		logger.Error(fmt.Sprintf("failed to connect Error=%v", err))
	}
	// 存根
	f.UserCli = proto.NewFileWorkerClient(conn)
	f.address = address
	f.channel = conn
}

func (f FileClient) DescribeFile(callerCode string, remoteFullPath string, timeout time.Duration, logger logrus.Logger) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	if callerCode == "" || remoteFullPath == "" {
		result["code"] = 128502
		result["message"] = "缺少请求必传参数"
		result["fileStream"] = nil
		return result, nil
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
			logger.Errorf(fmt.Sprintf("failed to connect Error=%v", err))
		}
		// 存根
		UserCli := proto.NewFileWorkerClient(conn)
		code, message, fileStream, err = common.CallDescribeFile(UserCli, callerCode, remoteFullPath, timeout)
		if err != nil {
			logger.Errorf(fmt.Sprintf("Get connect DescribeFile failed Error=%v", err))
		}
		f.UserCli = UserCli
	}
	result["code"] = code
	result["message"] = message
	result["fileStream"] = fileStream
	return result, err
}

func (f FileClient) CreateFile(callerCode string, mountPath string, xType string, fileData []byte, fileName string, filePath string, replace bool, timeout time.Duration, logger logrus.Logger) (map[string]interface{}, error) {
	var code int
	var message string
	var fileMountPath string
	var err error

	result := make(map[string]interface{})
	if callerCode == "" || filePath == "" || xType == "" || fileName == "" {
		result["code"] = 128502
		result["message"] = "缺少请求必传参数"
		result["fileMountPath"] = ""
		return result, nil
	}
	code, message, fileMountPath, err = common.CallCreateFile(f.UserCli, callerCode, fileName, fileData, replace, xType, mountPath, filePath, timeout)
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
			logger.Errorf(fmt.Sprintf("failed to connect Error=%v", err))
		}
		// 存根
		UserCli := proto.NewFileWorkerClient(conn)
		code, message, fileMountPath, err = common.CallCreateFile(UserCli, callerCode, fileName, fileData, replace, xType, mountPath, filePath, timeout)
		if err != nil {
			logger.Errorf(fmt.Sprintf("Get connect CreateFile failed Error=%v", err))
		}
		f.UserCli = UserCli
	}
	result["code"] = code
	result["message"] = message
	result["fileMountPath"] = fileMountPath
	return result, err
}

func (f FileClient) ModifyFile(filePath string, fileName string, callerCode string, isReplace bool, timeout time.Duration, logger logrus.Logger) (map[string]interface{}, error) {

	var code int
	var message string
	result := make(map[string]interface{})
	if callerCode == "" || filePath == "" || fileName == "" {
		result["code"] = 128502
		result["message"] = "缺少请求必传参数"
		return result, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()
	dataInfo, err := f.UserCli.ModifyFile(ctx, &proto.ModifyRequest{FilePath: filePath, FileName: fileName, FCode: callerCode, IsReplace: isReplace})
	if err != nil {
		errStatus, _ := status.FromError(err)
		if errStatus.Code() == codes.DeadlineExceeded {
			code = 128512
			message = "连接超时"
		} else if errStatus.Code() == codes.Unavailable {
			// 连接暂时不可用 重新尝试
			dataInfo, err = f.UserCli.ModifyFile(ctx, &proto.ModifyRequest{FilePath: filePath, FileName: fileName, FCode: callerCode, IsReplace: isReplace})
			if err != nil {
				logger.Errorf(fmt.Sprintf("Get connect ModifyFile failed :%v", err))
				result["code"] = 128509
				result["message"] = "读取超时"
				return result, err
			}
			code = int(dataInfo.Code)
			message = dataInfo.Err
		} else if errStatus.Code() == codes.InvalidArgument {
			code = 128501
			message = "参数异常"
		}
	} else {
		code = int(dataInfo.Code)
		message = dataInfo.Err
	}
	result["code"] = code
	result["message"] = message
	return result, err
}

func (f FileClient) CopyFile(newFilePath string, originalFilePath string, callerCode string, timeout time.Duration, logger logrus.Logger) (map[string]interface{}, error) {
	var code int
	var message string
	result := make(map[string]interface{})
	if newFilePath == "" || originalFilePath == "" || callerCode == "" {
		result["code"] = 128502
		result["message"] = "缺少请求必传参数"
		return result, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()
	dataInfo, err := f.UserCli.CopyFile(ctx, &proto.CopyRequest{NewFilePath: newFilePath, OriginalFilePath: originalFilePath, FCode: callerCode})
	if err != nil {
		errStatus, _ := status.FromError(err)
		if errStatus.Code() == codes.DeadlineExceeded {
			code = 128512
			message = "连接超时"
		} else if errStatus.Code() == codes.Unavailable {
			// 连接暂时不可用 重新尝试
			dataInfo, err = f.UserCli.CopyFile(ctx, &proto.CopyRequest{NewFilePath: newFilePath, OriginalFilePath: originalFilePath, FCode: callerCode})
			if err != nil {
				logger.Errorf(fmt.Sprintf("Get connect CopyFile failed :%v", err))
				result["code"] = 128509
				result["message"] = "读取超时"
				return result, err
			}
			code = int(dataInfo.Code)
			message = dataInfo.Err
		} else if errStatus.Code() == codes.InvalidArgument {
			code = 128501
			message = "参数异常"
		}
	} else {
		code = int(dataInfo.Code)
		message = dataInfo.Err
	}
	result["code"] = code
	result["message"] = message
	return result, err
}

func (f FileClient) MoveFile(newFilePath string, originalFilePath string, callerCode string, timeout time.Duration, logger logrus.Logger) (map[string]interface{}, error) {
	var code int
	var message string
	result := make(map[string]interface{})
	if newFilePath == "" || originalFilePath == "" || callerCode == "" {
		result["code"] = 128502
		result["message"] = "缺少请求必传参数"
		return result, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()
	dataInfo, err := f.UserCli.MoveFile(ctx, &proto.MoveRequest{NewFilePath: newFilePath, OriginalFilePath: originalFilePath, FCode: callerCode})
	if err != nil {
		errStatus, _ := status.FromError(err)
		if errStatus.Code() == codes.DeadlineExceeded {
			code = 128512
			message = "连接超时"
		} else if errStatus.Code() == codes.Unavailable {
			// 连接暂时不可用 重新尝试
			dataInfo, err = f.UserCli.MoveFile(ctx, &proto.MoveRequest{NewFilePath: newFilePath, OriginalFilePath: originalFilePath, FCode: callerCode})
			if err != nil {
				logger.Errorf(fmt.Sprintf("Get connect MoveFile failed :%v", err))
				result["code"] = 128509
				result["message"] = "读取超时"
				return result, err
			}
			code = int(dataInfo.Code)
			message = dataInfo.Err
		} else if errStatus.Code() == codes.InvalidArgument {
			code = 128501
			message = "参数异常"
		}
	} else {
		code = int(dataInfo.Code)
		message = dataInfo.Err
	}
	result["code"] = code
	result["message"] = message
	return result, err
}
