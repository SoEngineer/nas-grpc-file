package common

import (
	"encoding/json"
	"github.com/soengineer/nas-grpc-file/proto"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"time"
)

func CallDescribeFile(UserCli proto.FileWorkerClient, callerCode string, remoteFullPath string, timeout time.Duration) (int, string, []byte, error) {
	var code int
	var message string
	var fileStream []byte
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
			dataInfo, err = UserCli.DescribeFile(ctx, &proto.DescribeRequest{FCode: callerCode, LocateFile: remoteFullPath})
			if err != nil {
				code = 128509
				message = "读取超时"
				fileStream = nil
			} else {
				code = int(dataInfo.Code)
				message = dataInfo.Err
				fileStream = dataInfo.File
			}
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
	return code, message, fileStream, err
}

func CallCreateFile(UserCli proto.FileWorkerClient, callerCode string, fileName string, fileData []byte, replace bool, xType string, mountPath string, filePath string, timeout time.Duration) (int, string, string, error) {
	var code int
	var message string
	var fileMountPath string
	dict := make(map[string]string, 1)
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
			dataInfo, err = UserCli.CreateFile(ctx, &proto.CreateRequest{FCode: callerCode, FileName: fileName, FileData: fileData, IsReplace: replace, XType: xType, XMountPath: mountPath, XFilePath: filePath})
			if err != nil {
				code = 128509
				message = "读取超时"
				fileMountPath = ""
			} else {
				code = int(dataInfo.Code)
				message = dataInfo.Err
				_ = json.Unmarshal([]byte(dataInfo.Biz), &dict)
				fileMountPath = dict["mountPath"]
			}
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
	return code, message, fileMountPath, err
}
