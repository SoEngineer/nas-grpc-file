package client

import (
	"encoding/json"
	"flag"
	"fmt"
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

type FileHandler interface {
	InitConnection(add string)
	DescribeFile(callerCode string, remoteFullPath string, timeout time.Duration) (map[string]interface{}, error)
	CreateFile(callerCode string, mountPath string, xType string, fileData []byte, fileName string, filePath string, replace bool, timeout time.Duration) (map[string]interface{}, error)
	ModifyFile(filePath string, fileName string, callerCode string, isReplace bool, timeout time.Duration) (map[string]interface{}, error)
	CopyFile(newFilePath string, originalFilePath string, callerCode string, timeout time.Duration) (map[string]interface{}, error)
	MoveFile(newFilePath string, originalFilePath string, callerCode string, timeout time.Duration) (map[string]interface{}, error)
}

var UserCli proto.FileWorkerClient

type FileClient struct {
}

func (FileClient) InitConnection(addr string) {
	address := *flag.String("host", addr, "")
	opt := grpc.WithInsecure()
	conn, err := grpc.Dial(address, opt, grpc.WithDefaultCallOptions(
		grpc.MaxCallRecvMsgSize(MaxFileSize),
		grpc.MaxCallSendMsgSize(MaxFileSize)),
		grpc.WithInitialWindowSize(InitialWindowSize),
		grpc.WithInitialConnWindowSize(InitialWindowSize),
		grpc.WithWriteBufferSize(BufferSize),
		grpc.WithReadBufferSize(BufferSize))
	if err != nil {
		fmt.Println("failed to connect : ", err)
	}
	// 存根
	UserCli = proto.NewFileWorkerClient(conn)
}

func (FileClient) DescribeFile(callerCode string, remoteFullPath string, timeout time.Duration) (map[string]interface{}, error) {
	var code int
	var message string
	var fileStream []byte
	result := make(map[string]interface{})
	if callerCode == "" || remoteFullPath == "" {
		result["code"] = 128502
		result["message"] = "缺少请求必传参数"
		result["fileStream"] = nil
		return result, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()
	dataInfo, err := UserCli.DescribeFile(ctx, &proto.DescribeRequest{FCode: callerCode, LocateFile: remoteFullPath})
	if err != nil {
		errStatus, _ := status.FromError(err)
		if errStatus.Code() == codes.DeadlineExceeded {
			code = 128512
			message = "连接超时"
			fileStream = nil
		} else if errStatus.Code() == codes.Unavailable {
			// 连接暂时不可用 重新尝试
			dataInfo, err := UserCli.DescribeFile(ctx, &proto.DescribeRequest{FCode: callerCode, LocateFile: remoteFullPath})
			if err != nil {
				fmt.Printf("Get connect DescribeFile failed :%v", err)
				result["code"] = 1
				result["message"] = "请求失败"
				result["fileStream"] = nil
				return result, err
			}
			code = int(dataInfo.Code)
			message = dataInfo.Err
			fileStream = dataInfo.File
		} else if errStatus.Code() == codes.InvalidArgument {
			code = 128501
			message = "参数异常"
			fileStream = nil
		}
	} else {
		code = int(dataInfo.Code)
		message = dataInfo.Err
		fileStream = dataInfo.File
	}
	result["code"] = code
	result["message"] = message
	result["fileStream"] = fileStream
	return result, err
}

func (FileClient) CreateFile(callerCode string, mountPath string, xType string, fileData []byte, fileName string, filePath string, replace bool, timeout time.Duration) (map[string]interface{}, error) {
	var code int
	var message string
	var fileMountPath string
	dict := make(map[string]string, 1)
	result := make(map[string]interface{})
	if callerCode == "" || filePath == "" || xType == "" || fileName == "" {
		result["code"] = 128502
		result["message"] = "缺少请求必传参数"
		result["fileMountPath"] = ""
		return result, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()
	dataInfo, err := UserCli.CreateFile(ctx, &proto.CreateRequest{FCode: callerCode, FileName: fileName, FileData: fileData, IsReplace: replace, XType: xType, XMountPath: mountPath, XFilePath: filePath})
	if err != nil {
		errStatus, _ := status.FromError(err)
		if errStatus.Code() == codes.DeadlineExceeded {
			code = 128512
			message = "连接超时"
			fileMountPath = ""
		} else if errStatus.Code() == codes.Unavailable {
			// 连接暂时不可用 重新尝试
			dataInfo, err := UserCli.CreateFile(ctx, &proto.CreateRequest{FCode: callerCode, FileName: fileName, FileData: fileData, IsReplace: replace, XType: xType, XMountPath: mountPath, XFilePath: filePath})
			if err != nil {
				fmt.Printf("Get connect CreateFile failed :%v", err)
				result["code"] = 1
				result["message"] = "请求失败"
				result["fileMountPath"] = ""
				return result, err
			}
			code = int(dataInfo.Code)
			message = dataInfo.Err
			_ = json.Unmarshal([]byte(dataInfo.Biz), &dict)
			fileMountPath = dict["mountPath"]
		} else if errStatus.Code() == codes.InvalidArgument {
			code = 128501
			message = "参数异常"
			fileMountPath = ""
		}
	} else {
		code = int(dataInfo.Code)
		message = dataInfo.Err
		_ = json.Unmarshal([]byte(dataInfo.Biz), &dict)
		fileMountPath = dict["mountPath"]
	}
	result["code"] = code
	result["message"] = message
	result["fileMountPath"] = fileMountPath
	return result, err
}

func (FileClient) ModifyFile(filePath string, fileName string, callerCode string, isReplace bool, timeout time.Duration) (map[string]interface{}, error) {

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
	dataInfo, err := UserCli.ModifyFile(ctx, &proto.ModifyRequest{FilePath: filePath, FileName: fileName, FCode: callerCode, IsReplace: isReplace})
	if err != nil {
		errStatus, _ := status.FromError(err)
		if errStatus.Code() == codes.DeadlineExceeded {
			code = 128512
			message = "连接超时"
		} else if errStatus.Code() == codes.Unavailable {
			// 连接暂时不可用 重新尝试
			dataInfo, err := UserCli.ModifyFile(ctx, &proto.ModifyRequest{FilePath: filePath, FileName: fileName, FCode: callerCode, IsReplace: isReplace})
			if err != nil {
				fmt.Printf("Get connect ModifyFile failed :%v", err)
				result["code"] = 1
				result["message"] = "请求失败"
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

func (FileClient) CopyFile(newFilePath string, originalFilePath string, callerCode string, timeout time.Duration) (map[string]interface{}, error) {
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
	dataInfo, err := UserCli.CopyFile(ctx, &proto.CopyRequest{NewFilePath: newFilePath, OriginalFilePath: originalFilePath, FCode: callerCode})
	if err != nil {
		errStatus, _ := status.FromError(err)
		if errStatus.Code() == codes.DeadlineExceeded {
			code = 128512
			message = "连接超时"
		} else if errStatus.Code() == codes.Unavailable {
			// 连接暂时不可用 重新尝试
			dataInfo, err := UserCli.CopyFile(ctx, &proto.CopyRequest{NewFilePath: newFilePath, OriginalFilePath: originalFilePath, FCode: callerCode})
			if err != nil {
				fmt.Printf("Get connect CopyFile failed :%v", err)
				result["code"] = 1
				result["message"] = "请求失败"
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

func (FileClient) MoveFile(newFilePath string, originalFilePath string, callerCode string, timeout time.Duration) (map[string]interface{}, error) {
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
	dataInfo, err := UserCli.MoveFile(ctx, &proto.MoveRequest{NewFilePath: newFilePath, OriginalFilePath: originalFilePath, FCode: callerCode})
	if err != nil {
		errStatus, _ := status.FromError(err)
		if errStatus.Code() == codes.DeadlineExceeded {
			code = 128512
			message = "连接超时"
		} else if errStatus.Code() == codes.Unavailable {
			// 连接暂时不可用 重新尝试
			dataInfo, err := UserCli.MoveFile(ctx, &proto.MoveRequest{NewFilePath: newFilePath, OriginalFilePath: originalFilePath, FCode: callerCode})
			if err != nil {
				fmt.Printf("Get connect MoveFile failed :%v", err)
				result["code"] = 1
				result["message"] = "请求失败"
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
