package pop3

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"mailfinger/utils"
	"net"
	"os"
	"strings"
	"time"
)

func DoQuery(target string) {
	var err error

	// 格式化目标
	target = fmt.Sprintf("%s:110", target)

	// 创建目标文件夹
	folderName := fmt.Sprintf("./data/%s/pop3", strings.ReplaceAll(strings.Split(target, ":")[0], ".", "_"))
	err = os.MkdirAll(folderName, os.ModePerm)

	// 建立套接字
	timeout := 2 * time.Second // 设置超时时间为2秒
	conn, err := net.DialTimeout("tcp", target, timeout)
	if err != nil {
		fmt.Printf("%s 连接失败: %v\n", target, err)
		return
	}
	defer conn.Close()

	// 读取欢迎消息
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		fmt.Printf("读取欢迎消息失败: %v\n", err)
		return
	}

	welcomeMsg := string(buffer[:n])
	utils.SaveToFile(fmt.Sprintf("%s/welcome_message.txt", folderName), welcomeMsg)

	// 发送 CAPA 命令以检查是否支持 STLS
	_, err = conn.Write([]byte("CAPA\r\n"))
	if err != nil {
		fmt.Printf("发送 CAPA 失败: %v\n", err)
		return
	}

	// 读取 CAPA 响应
	capaResponse, err := readAll(conn)
	if err != nil {
		fmt.Printf("读取 CAPA 响应失败: %v\n", err)
		return
	}
	utils.SaveToFile(fmt.Sprintf("%s/capa_response.txt", folderName), capaResponse)

	if !strings.Contains(capaResponse, "STLS") {
		fmt.Printf("%s 服务器不支持 STLS\n", target)
		return
	}

	// 发送 STLS 命令
	_, err = conn.Write([]byte("STLS\r\n"))
	if err != nil {
		fmt.Printf("发送 STLS 失败: %v\n", err)
		return
	}

	// 读取 STLS 响应
	stlsResponse, err := readAll(conn)
	if err != nil {
		fmt.Printf("读取 STLS 响应失败: %v\n", err)
		return
	}
	// 保存 STLS 响应到文件
	utils.SaveToFile(fmt.Sprintf("%s/stls_response.txt", folderName), stlsResponse)

	// 创建自定义 TLS 配置（跳过证书验证）
	tlsConfig := &tls.Config{
		MinVersion:         tls.VersionTLS10,
		MaxVersion:         tls.VersionTLS13,
		InsecureSkipVerify: true,        // 仅用于测试环境，请勿在生产环境中使用
		ServerName:         "baidu.com", // 应与目标邮件服务器的SSL证书中的CN匹配
	}

	// 升级到 TLS 连接
	tlsConn := tls.Client(conn, tlsConfig)
	err = tlsConn.Handshake()
	if err != nil {
		fmt.Printf("TLS 握手失败: %v\n", err)
		return
	}
	defer tlsConn.Close()

	// 获取证书链并打印证书信息
	certs := tlsConn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		fmt.Println("未获取到证书")
		return
	}

	fmt.Printf("%s 获取到证书数量:%d\n", target, len(certs))

	for i, cert := range certs {
		// 创建或打开文件
		filename := fmt.Sprintf("%s/cert_%d.txt", folderName, i+1)
		file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Printf("无法创建或打开文件: %v\n", err)
			continue
		}
		defer file.Close()

		// 创建 writer 并写入证书详情
		writer := bufio.NewWriter(file)
		writer.WriteString(fmt.Sprintf("========= 证书 %d =========\n", i+1))
		writer.WriteString(fmt.Sprintf("主题: %v\n", cert.Subject))
		writer.WriteString(fmt.Sprintf("颁发者: %v\n", cert.Issuer))
		writer.WriteString(fmt.Sprintf("有效期:\n 开始: %v\n 结束: %v\n", cert.NotBefore.Format("2006-01-02"), cert.NotAfter.Format("2006-01-02")))
		writer.WriteString(fmt.Sprintf("序列号: %v\n", cert.SerialNumber))
		writer.WriteString(fmt.Sprintf("签名算法: %v\n", cert.SignatureAlgorithm))
		writer.WriteString(fmt.Sprintf("版本: %v\n", cert.Version))
		writer.Flush() // 别忘了刷新缓冲区
	}
}

// 读取所有数据
func readAll(conn net.Conn) (string, error) {
	buffer := make([]byte, 1024)
	var response []byte
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			return "", err
		}
		response = append(response, buffer[:n]...)
		if n < len(buffer) { // 如果读取的数据少于缓冲区大小，则认为读取完成
			break
		}
	}
	return string(response), nil
}
