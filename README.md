# nas-grpc-file
grpc客户端

install
go get github.com/Gyjnine/nas-grpc-file

# 错误码
//	0 = "操作成功"<br>
//	1 = "请求失败"<br>
//	128501= "参数异常"<br>
//	128502= "缺少请求必传参数"<br>
//	128503= "文件缺少扩展名"<br>
//	128504= "目录不存在"<br>
//	128505= "文件不存在"<br>
//	128506= "不是文件"<br>
//	128507= "文件已存在"<br>
//	128508= "目录已存在"<br>
//	128509= "读取超时"<br>
//	128510= "保存文件失败"<br>
//	128511= "文件夹创建失败"<br>
//	128512= "连接超时"<br>
//	128513= "读取文件失败"<br>
//	128514= "权限不足"<br>
//	128515= "业务场景不存在"<br>
//  128516= "文件复制失败"<br>
//  128517= "文件改名失败"<br>
//  128518= "文件移动失败"<br>


# 调用示例
import (
	"fmt"
	"github.com/Gyjnine/nas-grpc-file"
)

func main() {

	grpcClient := client.FileClient{}
	err := grpcClient.InitConnection("172.16.20.30:31547")
	describeFile, err := grpcClient.DescribeFile("test", "/picture/CERTIFICATE/test.jpg", 300)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(describeFile.Code, describeFile.Message)
	file, _ := os.OpenFile("D:/", os.O_CREATE|os.O_RDWR|os.O_TRUNC, os.ModePerm)
	defer file.Close()
	w := bufio.NewWriter(file) //创建新的 Writer 对象
	_, err = w.Write(describeFile.FileStream)
	_ = w.Flush()
	f,_ := os.Open("D:/test.jpg")
	fileData, _:= ioutil.ReadAll(f)
	createFile, err := grpcClient.CreateFile("test", "111", "m128205", fileData, "test111.jpg", "/", true,300)
	if err != nil {
		fmt.Sprintln(err)
	}
	fmt.Println(createFile.Code,createFile.Message, createFile.FileMountPath)
	modifyFile, err := grpcClient.ModifyFile("/picture/CERTIFICATE/test111.jpg", "test1111.jpg", "test", true, 300)
	if err != nil {
		fmt.Sprintln(err)
	}
	fmt.Println(modifyFile.Code, err)
	copyFile, err := grpcClient.CopyFile("/picture/CERTIFICATE/test111.jpg", "/picture/CERTIFICATE/test1111.jpg", "test", 300)
	if err != nil {
		fmt.Sprintln(err)
	}
	fmt.Println(copyFile.Code, err)
	moveFile, err := grpcClient.MoveFile("/picture/CERTIFICATE/test/test1.jpg", "/picture/CERTIFICATE/test1111.jpg", "test", 300)
	if err != nil {
		fmt.Sprintln(err)
	}
	fmt.Println(moveFile.Code, err)
    
}