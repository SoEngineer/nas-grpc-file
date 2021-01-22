# nas-grpc-file
grpc客户端


//	0 = "操作成功"
//	1 = "请求失败"
//	128501= "参数异常"
//	128502= "缺少请求必传参数"
//	128503= "文件缺少扩展名"
//	128504= "目录不存在"
//	128505= "文件不存在"
//	128506= "不是文件"
//	128507= "文件已存在"
//	128508= "目录已存在"
//	128509= "读取超时"
//	128510= "保存文件失败"
//	128511= "文件夹创建失败"
//	128512= "连接超时"
//	128513= "读取文件失败"
//	128514= "权限不足"
//	128515= "业务场景不存在"


func main() {
var client FileHandler = FileClient{}
client.InitConnection("172.16.20.30:31547")
describeFile, err := client.DescribeFile("test", "D://test1.jpg", 3)
fmt.Println(describeFile["code"], err)
f,_ := os.Open("D:/test.jpg")
fileData, _:= ioutil.ReadAll(f)
createFile, err := client.CreateFile("test", "", "m128205", fileData, "test.jpg", "/", true,300)
fmt.Println(createFile["code"], err)
modifyFile, err := client.ModifyFile("D://test01.jpg", "go_test.jpg", "test", true, 3)
fmt.Println(modifyFile["code"], err)
copyFile, err := client.CopyFile("D://go_test_01.jpg", "D://go_test.jpg", "test", 3)
fmt.Println(copyFile["code"], err)
moveFile, err := client.MoveFile("D://grpc_test/go_test.jpg", "D://go_test.jpg", "test", 3)
fmt.Println(moveFile["code"], err)
}