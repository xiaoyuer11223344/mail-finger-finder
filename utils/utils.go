package utils

import (
	"bufio"
	"fmt"
	"os"
)

func SaveToFile(filename, content string) {

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("无法创建或打开文件: %v\n", err)
		return
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	_, err = writer.WriteString(content)
	if err != nil {
		fmt.Printf("写入文件失败: %v\n", err)
		return
	}
	writer.Flush()
}
