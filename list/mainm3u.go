package list

import (
	"fmt"
	"io/ioutil"
    "os"
	"net/http"
)

type Mainm3u struct {

}

func (t *Mainm3u) GetMainM3u(w http.ResponseWriter, filePath string) {
	// 打开文件
    file, err := os.Open(filePath)
    if err != nil {
        fmt.Println("Error opening file:", err)
        return
    }
    defer file.Close()

    // 读取文件内容到字节切片
    contentBytes, err := ioutil.ReadAll(file)
    if err != nil {
        fmt.Println("Error reading file content:", err)
        return
    }

    // 将字节切片转换为字符串
    contentString := string(contentBytes)

	fmt.Fprint(w, contentString)
}
