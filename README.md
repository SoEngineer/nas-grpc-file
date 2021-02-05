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

# 调用示例
import (
	"fmt"
	"github.com/Gyjnine/nas-grpc-file"
)

func main() {

    // 初始化连接
    var grpcClient client.FileHandler = client.FileClient{}
    client.InitConnection("172.16.20.30:31547")
    // 文件下载
    describeFile, err := client.DescribeFile("test", "D://test1.jpg", 3)
    fmt.Println(describeFile["code"], err)
    //文件上传
    f,_ := os.Open("D:/test.jpg")
    fileData, _:= ioutil.ReadAll(f)
    createFile, err := client.CreateFile("test", "", "m128205", fileData, "test.jpg", "/", true,300)
    fmt.Println(createFile["code"], err)
    //文件重命名
    modifyFile, err := client.ModifyFile("D://test01.jpg", "go_test.jpg", "test", true, 3)
    fmt.Println(modifyFile["code"], err)
    // 文件复制
    copyFile, err := client.CopyFile("D://go_test_01.jpg", "D://go_test.jpg", "test", 3)
    fmt.Println(copyFile["code"], err)
    // 文件移动
    moveFile, err := client.MoveFile("D://grpc_test/go_test.jpg", "D://go_test.jpg", "test", 3)
    fmt.Println(moveFile["code"], err)
    
}